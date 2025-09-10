load("@//bazel:system.bzl", "cc_system_headers")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

package(default_visibility = ["//visibility:public"])

# Generate config.h from template
expand_template(
    name = "config_h",
    out = "src/config.h",
    substitutions = {
        "@PACKAGE_NAME@": "Fontconfig",
        "@PACKAGE_VERSION@": "2.16.0",
        "@ENABLE_LIBXML2@": "1",
        "@FLEXIBLE_ARRAY_MEMBER@": "[]",  # For C99 and later
    },
    template = "config.h.in",
)

# Generate fcstdint.h from template
expand_template(
    name = "fcstdint_h",
    out = "src/fcstdint.h",
    substitutions = {},
    template = "src/fcstdint.h.in",
)

py_binary(
    name = "makealias_tool",
    srcs = ["src/makealias.py"],
    main = "src/makealias.py",
    visibility = ["//visibility:public"],
)

filegroup(
    name = "src_files",
    srcs = glob(["src/**"]),
)

filegroup(
    name = "fontconfig_files",
    srcs = glob(["fontconfig/**"]),
)

genrule(
    name = "fcalias_headers",
    srcs = [
        ":src_files",
        ":fontconfig_files",
    ],
    outs = [
        "fcalias.h",
        "fcaliastail.h",
    ],
    cmd = """
        mkdir -p src fontconfig
        for src in $(locations :src_files); do
            cp $$src src/
        done
        for fc in $(locations :fontconfig_files); do
            cp $$fc fontconfig/
        done
        $(location :makealias_tool) src fcalias.h.tmp fcaliastail.h.tmp fontconfig/fontconfig.h src/fcdeprecate.h fontconfig/fcprivate.h
        mv fcalias.h.tmp $(location fcalias.h)
        mv fcaliastail.h.tmp $(location fcaliastail.h)    
        """,
    tools = [":makealias_tool"],
)

genrule(
    name = "ft_alias_headers",
    srcs = [
        ":src_files",
        ":fontconfig_files",
    ],
    outs = [
        "src/fcftalias.h",
        "src/fcftaliastail.h",
    ],
    cmd = """
        mkdir -p src fontconfig
        for src in $(locations :src_files); do
            cp $$src src/
        done
        for fc in $(locations :fontconfig_files); do
            cp $$fc fontconfig/
        done
        $(location :makealias_tool) src fcftalias.h.tmp fcftaliastail.h.tmp fontconfig/fcfreetype.h
        mkdir -p $(RULEDIR)/src
        mv fcftalias.h.tmp $(location src/fcftalias.h)
        mv fcftaliastail.h.tmp $(location src/fcftaliastail.h)
    """,
    tools = [":makealias_tool"],
)

cc_library(
    name = "fontconfig",
    srcs = glob([
        "src/*.c",
        "fc-lang/*.c",
        "fc-case/*.c",
    ]),
    hdrs = glob([
        "fontconfig/*.h",
        "src/*.h",
    ]) + [
        "config-fixups.h",
        "fcalias.h",
        "fcaliastail.h",
        "src/config.h",
        "src/fcftalias.h",
        "src/fcftaliastail.h",
        "src/fcstdint.h",
    ],
    copts = [
        "-DHAVE_CONFIG_H",
        "-I$(execpath :config_h)",
        "-I$(BINDIR)/@fontconfig",
    ],
    data = [
        ":config_h",
        ":fcalias_headers",
        ":fcstdint_h",
        ":ft_alias_headers",
    ],
    includes = [
        ".",
        "src",
    ],
    deps = [
        "@freetype",
        "@libxml2",
        "@zlib",
    ],
)
