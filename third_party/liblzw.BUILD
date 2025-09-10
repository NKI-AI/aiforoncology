load("@//bazel:system.bzl", "cc_system_headers")

package(default_visibility = ["//visibility:public"])

config_setting(
    name = "clang_compiler",
    flag_values = {"@bazel_tools//tools/cpp:compiler": "clang"},
)

genrule(
    name = "generate_config_h",
    srcs = [],
    outs = ["config.h"],
    cmd = "\n".join([
        "cat <<'EOF' > $@",
        "#ifndef __CONFIG_H__",
        "#define __CONFIG_H__",
        "#define HAVE_ASSERT_H 1",
        "#define HAVE_CTYPE_H 1",
        "#define HAVE_DLFCN_H 1",
        "#define HAVE_ERRNO_H 1",
        "#define HAVE_FCNTL_H 1",
        "#undef HAVE_FEATURES_H",
        "#define HAVE_INTTYPES_H 1",
        "#define HAVE_MEMORY_H 1",
        "#define HAVE_STDARG_H 1",
        "#define HAVE_STDINT_H 1",
        "#define HAVE_STDIO_H 1",
        "#define HAVE_STDLIB_H 1",
        "#define HAVE_STRINGS_H 1",
        "#define HAVE_STRING_H 1",
        "#define HAVE_SYS_CDEFS_H 1",
        "#define HAVE_SYS_STAT_H 1",
        "#define HAVE_SYS_TYPES_H 1",
        "#define HAVE_TIME_H 1",
        "#define HAVE_UNISTD_H 1",
        '#define LT_OBJDIR ".libs/"',
        '#define PACKAGE "liblzw"',
        '#define PACKAGE_BUGREPORT "https://github.com/vapier/liblzw/issues"',
        '#define PACKAGE_NAME "liblzw"',
        '#define PACKAGE_STRING "liblzw 0.3"',
        '#define PACKAGE_TARNAME "liblzw"',
        '#define PACKAGE_URL "https://github.com/vapier/liblzw"',
        '#define PACKAGE_VERSION "0.3"',
        "#define STDC_HEADERS 1",
        '#define VERSION "0.3"',
        "#endif /* __CONFIG_H__ */",
        "EOF",
    ]),
    visibility = ["//visibility:public"],
)

cc_system_headers(
    name = "liblzw_headers_config",
    hdrs = ["config.h"],
)

LIBLZW_HEADERS = [
    "headers.h",
    "helpers.h",
    "lzw.h",
    "lzw_internal.h",
]

cc_system_headers(
    name = "liblzw_headers",
    hdrs = LIBLZW_HEADERS,
)

cc_library(
    name = "liblzw",
    srcs = ["lzw.c"],
    hdrs = LIBLZW_HEADERS + ["config.h"],
    linkopts = [],
    visibility = ["//visibility:public"],
    deps = [
        ":liblzw_headers",
        ":liblzw_headers_config",
    ],
)
