"Define linter aspects"

load("@aspect_rules_lint//lint:buf.bzl", "lint_buf_aspect")
load("@aspect_rules_lint//lint:clang_tidy.bzl", "lint_clang_tidy_aspect")
load("@aspect_rules_lint//lint:keep_sorted.bzl", "lint_keep_sorted_aspect")
load("@aspect_rules_lint//lint:lint_test.bzl", "lint_test")
load("@aspect_rules_lint//lint:shellcheck.bzl", "lint_shellcheck_aspect")
load("@aspect_rules_lint//lint:spotbugs.bzl", "lint_spotbugs_aspect")
load("//tools/lint:ruff.bzl", "lint_ruff_aspect_aifo")

buf = lint_buf_aspect(
    config = Label("@//:buf.yaml"),
)

ruff = lint_ruff_aspect_aifo(
    binary = "@multitool//tools/ruff",
    configs = [
        Label("@//:.ruff.toml"),
    ],
)

ruff_test = lint_test(aspect = ruff)

shellcheck = lint_shellcheck_aspect(
    binary = "@multitool//tools/shellcheck",
    config = Label("@//:.shellcheckrc"),
)

shellcheck_test = lint_test(aspect = shellcheck)

clang_tidy = lint_clang_tidy_aspect(
    binary = "@@//tools/lint:clang_tidy",
    global_config = "@@//:.clang-tidy",
    lint_target_headers = True,
    angle_includes_are_system = False,
    verbose = False,
)
clang_tidy_test = lint_test(aspect = clang_tidy)

spotbugs = lint_spotbugs_aspect(
    binary = Label("@//tools/lint:spotbugs"),
    exclude_filter = Label("@//:spotbugs-exclude.xml"),
)

spotbugs_test = lint_test(aspect = spotbugs)

keep_sorted = lint_keep_sorted_aspect(
    binary = Label("@com_github_google_keep_sorted//:keep-sorted"),
)

keep_sorted_test = lint_test(aspect = keep_sorted)
