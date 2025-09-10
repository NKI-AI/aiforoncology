"""API for declaring a Ruff lint aspect that visits py_library rules.

Forked from https://github.com/aspect-build/rules_lint/blob/main/lint/ruff.bzl
Modified to be able to output machine readable JSON and fix paths.
Original license:

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
```
"""

load("@aspect_rules_lint//lint:ruff.bzl", "ruff_fix")
load("@aspect_rules_lint//lint/private:lint_aspect.bzl", "LintOptionsInfo", "filter_srcs", "noop_lint_action", "output_files", "patch_and_output_files", "should_visit")

_MNEMONIC = "AspectRulesLintRuff"

def ruff_action(ctx, executable, srcs, config, stdout, exit_code = None, env = {}, machine_output = False):
    """Run ruff as an action under Bazel.

    Ruff will select the configuration file to use for each source file, as documented here:
    https://docs.astral.sh/ruff/configuration/#config-file-discovery

    Note: all config files are passed to the action.
    This means that a change to any config file invalidates the action cache entries for ALL
    ruff actions.

    However this is needed because:

    1. ruff has an `extend` field, so it may need to read more than one config file
    2. ruff's logic for selecting the appropriate config needs to read the file content to detect
      a `[tool.ruff]` section.

    Args:
        ctx: Bazel Rule or Aspect evaluation context
        executable: label of the the ruff program
        srcs: python files to be linted
        config: labels of ruff config files (pyproject.toml, ruff.toml, or .ruff.toml)
        stdout: output file of linter results to generate
        exit_code: output file to write the exit code.
            If None, then fail the build when ruff exits non-zero.
            See https://github.com/astral-sh/ruff/blob/dfe4291c0b7249ae892f5f1d513e6f1404436c13/docs/linter.md#exit-codes
        env: environment variaables for ruff
        machine_output: if True, generate machine-readable output (JSON format) and process paths
    """
    inputs = srcs + config
    outputs = [stdout]

    # Wire command-line options, see
    # `ruff help check` to see available options
    args = ctx.actions.args()
    args.add("check")

    # Honor exclusions in pyproject.toml even though we pass explicit list of files
    args.add("--force-exclude")
    args.add_all(srcs)

    if machine_output:
        args.add_all(["--output-format=rdjson"])

        # Create a temporary file for the raw ruff output
        raw_output = ctx.actions.declare_file(stdout.basename + ".raw")

        if exit_code:
            command = "{ruff} $@ >{raw_output}; echo $? >" + exit_code.path
            outputs.append(exit_code)
        else:
            # Create empty file on success, as Bazel expects one
            command = "{ruff} $@ >{raw_output} || touch {raw_output}"

        # Run the initial ruff action that generates the raw output
        ctx.actions.run_shell(
            inputs = inputs,
            outputs = [raw_output] + ([exit_code] if exit_code else []),
            command = command.format(ruff = executable.path, raw_output = raw_output.path),
            arguments = [args],
            mnemonic = _MNEMONIC,
            env = env,
            progress_message = "Linting %{label} with Ruff",
            tools = [executable],
        )

        # Postprocess the output to fix paths using jq
        ctx.actions.run_shell(
            inputs = [raw_output],
            outputs = [stdout],
            command = "if [ -s {raw_output} ]; then jq '.diagnostics[].location.path |= sub(\"^.*?aifo/\"; \"aifo/\")' {raw_output} > {stdout}; else touch {stdout}; fi".format(
                raw_output = raw_output.path,
                stdout = stdout.path,
            ),
            mnemonic = _MNEMONIC + "PathFix",
            progress_message = "Fixing paths in Ruff JSON output for %{label}",
        )
    else:
        if exit_code:
            command = "{ruff} $@ >{stdout}; echo $? >" + exit_code.path
            outputs.append(exit_code)
        else:
            # Create empty file on success, as Bazel expects one
            command = "{ruff} $@ && touch {stdout}"

        ctx.actions.run_shell(
            inputs = inputs,
            outputs = outputs,
            command = command.format(ruff = executable.path, stdout = stdout.path),
            arguments = [args],
            mnemonic = _MNEMONIC,
            env = env,
            progress_message = "Linting %{label} with Ruff",
            tools = [executable],
        )

# buildifier: disable=function-docstring
def _ruff_aspect_impl(target, ctx):
    if not should_visit(ctx.rule, ctx.attr._rule_kinds):
        return []

    files_to_lint = filter_srcs(ctx.rule)
    if ctx.attr._options[LintOptionsInfo].fix:
        outputs, info = patch_and_output_files(_MNEMONIC, target, ctx)
    else:
        outputs, info = output_files(_MNEMONIC, target, ctx)

    if len(files_to_lint) == 0:
        noop_lint_action(ctx, outputs)
        return [info]

    color_env = {"FORCE_COLOR": "1"} if ctx.attr._options[LintOptionsInfo].color else {}

    # Ruff can produce a patch at the same time as reporting the unpatched violations
    if hasattr(outputs, "patch"):
        ruff_fix(ctx, ctx.executable, files_to_lint, ctx.files._config_files, outputs.patch, outputs.human.out, outputs.human.exit_code, env = color_env)
    else:
        ruff_action(ctx, ctx.executable._ruff, files_to_lint, ctx.files._config_files, outputs.human.out, outputs.human.exit_code, env = color_env)

    # Process machine-readable output with path normalization
    ruff_action(ctx, ctx.executable._ruff, files_to_lint, ctx.files._config_files, outputs.machine.out, outputs.machine.exit_code, machine_output = True)

    return [info]

def lint_ruff_aspect_aifo(binary, configs, rule_kinds = ["py_binary", "py_library", "py_test"]):
    """A factory function to create a linter aspect.

    Attrs:
        binary: a ruff executable
        configs: ruff config file(s) (`pyproject.toml`, `ruff.toml`, or `.ruff.toml`)
        rule_kinds: which [kinds](https://bazel.build/query/language#kind) of rules should be visited by the aspect
    """

    # syntax-sugar: allow a single config file in addition to a list
    if type(configs) == "string":
        configs = [configs]

    return aspect(
        implementation = _ruff_aspect_impl,
        attrs = {
            "_options": attr.label(
                default = "//tools/lint:options",
                providers = [LintOptionsInfo],
            ),
            "_ruff": attr.label(
                default = binary,
                allow_files = True,
                executable = True,
                cfg = "exec",
            ),
            "_patcher": attr.label(
                default = "@aspect_rules_lint//lint/private:patcher",
                executable = True,
                cfg = "exec",
            ),
            "_config_files": attr.label_list(
                default = configs,
                allow_files = True,
            ),
            "_rule_kinds": attr.string_list(
                default = rule_kinds,
            ),
        },
    )
