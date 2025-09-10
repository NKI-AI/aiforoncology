load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

VERSION_MAJOR = "3"

VERSION_MINOR = "4"

VERSION_MICRO = "7"

CMD_COMMON = [
    "#define EH_FRAME_FLAGS \"a\"",
    "#define HAVE_ALLOCA 1",
    "#define HAVE_ALLOCA_H 1",
    "#define HAVE_AS_CFI_PSEUDO_OP 1",
    "#define HAVE_DLFCN_H 1",
    "#define HAVE_HIDDEN_VISIBILITY_ATTRIBUTE 1",
    "#define HAVE_INTTYPES_H 1",
    "#define HAVE_MEMCPY 1",
    "#define HAVE_MKOSTEMP 1",
    "#define HAVE_MKSTEMP 1",
    "#define HAVE_MMAP 1",
    "#define HAVE_MMAP_ANON 1",
    "#define HAVE_MMAP_FILE 1",
    "#define HAVE_RO_EH_FRAME 1",
    "#define HAVE_STDINT_H 1",
    "#define HAVE_STDIO_H 1",
    "#define HAVE_STDLIB_H 1",
    "#define HAVE_STRINGS_H 1",
    "#define HAVE_STRING_H 1",
    "#define HAVE_SYS_MMAN_H 1",
    "#define HAVE_SYS_STAT_H 1",
    "#define HAVE_SYS_TYPES_H 1",
    "#define HAVE_UNISTD_H 1",
    "#define LT_OBJDIR \".libs/\"",
    "#define PACKAGE_BUGREPORT \"http://github.com/libffi/libffi/issues\"",
    "#define PACKAGE_NAME \"libffi\"",
    "#define PACKAGE_STRING \"libffi {}.{}.{}\"".format(VERSION_MAJOR, VERSION_MINOR, VERSION_MICRO),
    "#define PACKAGE_TARNAME \"libffi\"",
    "#define PACKAGE_URL \"\"",
    "#define PACKAGE_VERSION \"{}.{}.{}\"".format(VERSION_MAJOR, VERSION_MINOR, VERSION_MICRO),
    "#define SIZEOF_DOUBLE 8",
    "#define SIZEOF_SIZE_T 8",
    "#define STDC_HEADERS 1",
    "#define VERSION \"{}.{}.{}\"".format(VERSION_MAJOR, VERSION_MINOR, VERSION_MICRO),
    "#if defined AC_APPLE_UNIVERSAL_BUILD",
    "# if defined __BIG_ENDIAN__",
    "#  define WORDS_BIGENDIAN 1",
    "# endif",
    "#else",
    "# ifndef WORDS_BIGENDIAN",
    "# endif",
    "#endif",
    "#ifdef HAVE_HIDDEN_VISIBILITY_ATTRIBUTE",
    "#ifdef LIBFFI_ASM",
    "#ifdef __APPLE__",
    "#define FFI_HIDDEN(name) .private_extern name",
    "#else",
    "#define FFI_HIDDEN(name) .hidden name",
    "#endif",
    "#else",
    "#define FFI_HIDDEN __attribute__ ((visibility (\"hidden\")))",
    "#endif",
    "#else",
    "#ifdef LIBFFI_ASM",
    "#define FFI_HIDDEN(name)",
    "#else",
    "#define FFI_HIDDEN",
    "#endif",
    "#endif",
]

genrule(
    name = "generate_fficonfig_h",
    outs = ["_generated/fficonfig.h"],
    cmd = select({
        "@platforms//os:macos": "\n".join([
            "cat <<'EOF' >$@",
            "#define FFI_EXEC_TRAMPOLINE_TABLE 1",
            "#define HAVE_RO_EH_FRAME 1",
        ] + CMD_COMMON + ["EOF"]),
        "//conditions:default": "\n".join([
            "cat <<'EOF' >$@",
        ] + CMD_COMMON + ["EOF"]),
    }),
)


config_setting(
    name = "macos_arm64",
    constraint_values = [
        "@platforms//os:macos",
        "@platforms//cpu:arm64",
    ],
)

config_setting(
    name = "linux_arm64",
    constraint_values = [
        "@platforms//os:linux",
        "@platforms//cpu:arm64",
    ],
)

config_setting(
    name = "macos_x86_64",
    constraint_values = [
        "@platforms//os:macos",
        "@platforms//cpu:x86_64",
    ],
)

config_setting(
    name = "linux_x86_64",
    constraint_values = [
        "@platforms//os:linux",
        "@platforms//cpu:x86_64",
    ],
)

VERSION = "{}.{}.{}".format(VERSION_MAJOR, VERSION_MINOR, VERSION_MICRO)

expand_template(
    name = "generate_ffi_h",
    out = "_generated/include/ffi.h",
    substitutions = select({
        ":macos_arm64": {
            "@VERSION@": VERSION,
            "@TARGET@": "AARCH64",
            "@HAVE_LONG_DOUBLE@": "0",
            "@FFI_EXEC_TRAMPOLINE_TABLE@": "1",
        },
        ":linux_arm64": {
            "@VERSION@": VERSION,
            "@TARGET@": "AARCH64",
            "@HAVE_LONG_DOUBLE@": "0",
            "@FFI_EXEC_TRAMPOLINE_TABLE@": "0",
        },
        ":macos_x86_64": {
            "@VERSION@": VERSION,
            "@TARGET@": "X86_64",
            "@HAVE_LONG_DOUBLE@": "1",
            "@FFI_EXEC_TRAMPOLINE_TABLE@": "1",
        },
        ":linux_x86_64": {
            "@VERSION@": VERSION,
            "@TARGET@": "X86_64",
            "@HAVE_LONG_DOUBLE@": "1",
            "@FFI_EXEC_TRAMPOLINE_TABLE@": "0",
        },
        "//conditions:default": {
            "@VERSION@": VERSION,
            "@TARGET@": "",
            "@HAVE_LONG_DOUBLE@": "0",
            "@FFI_EXEC_TRAMPOLINE_TABLE@": "0",
        },
    }),
    template = "include/ffi.h.in",
)


cc_library(
    name = "ffi_h",
    hdrs = [
        "_generated/include/ffi.h",
    ],
    includes = [
        "_generated/include",
    ],
)

cc_library(
    name = "x86_ffitarget_h",
    hdrs = [
        "src/x86/ffitarget.h",
    ],
    strip_include_prefix = "src/x86",
)

cc_library(
    name = "arm64_ffitarget_h",
    hdrs = [
        "src/aarch64/ffitarget.h",
    ],
    strip_include_prefix = "src/aarch64",
)

cc_library(
    name = "common_h",
    hdrs = [
        "include/ffi_cfi.h",
        "include/ffi_common.h",
        "include/tramp.h",
    ],
    strip_include_prefix = "include",
)

cc_library(
    name = "fficonfig_h",
    hdrs = [
        "_generated/fficonfig.h",
    ],
    strip_include_prefix = "_generated",
)

cc_library(
    name = "libffi",
    srcs = [
        "src/closures.c",
        "src/debug.c",
        "src/prep_cif.c",
        "src/raw_api.c",
        "src/tramp.c",
        "src/types.c",
    ] + select({
        "@platforms//cpu:x86_64": [
            "src/x86/asmnames.h",
            "src/x86/ffi64.c",
            "src/x86/ffiw64.c",
            "src/x86/internal.h",
            "src/x86/internal64.h",
            "src/x86/sysv.S",
            "src/x86/unix64.S",
            "src/x86/win64.S",
        ],
        "@platforms//cpu:arm64": [
            "src/aarch64/ffi.c",
            "src/aarch64/internal.h",
            "src/aarch64/sysv.S",
        ],
    }),
    copts = [
        "-Wno-implicit-function-declaration",
    ],
    local_defines = select({
        "@platforms//cpu:x86_64": [
            "HAVE_AS_X86_64_UNWIND_SECTION_TYPE=1",
            "HAVE_AS_X86_PCREL=1",
            "HAVE_LONG_DOUBLE=1",
            "HAVE_MEMFD_CREATE=1",
            "HAVE_MMAP_DEV_ZERO=1",
            "SIZEOF_LONG_DOUBLE=16",
            "FFI_EXEC_STATIC_TRAMP=1",
        ],
        "@platforms//cpu:arm64": [
            "SIZEOF_LONG_DOUBLE=16",
        ],
        "//conditions:default": [],
    }),
    textual_hdrs = ["src/dlmalloc.c"],
    visibility = ["//visibility:public"],
    deps = [
        ":common_h",
        ":ffi_h",
        ":fficonfig_h",
    ] + select({
        "@platforms//cpu:x86_64": [
            ":x86_ffitarget_h",
        ],
        "@platforms//cpu:arm64": [
            ":arm64_ffitarget_h",
        ],
    }),
)
