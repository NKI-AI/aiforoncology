# Library for low-level pixel manipulation. https://github.com/freedesktop/pixman
# TODO: Why doesn't neon work?
licenses(["notice"])

package(default_visibility = ["//visibility:public"])

config_setting(
    name = "clang_compiler",
    flag_values = {"@bazel_tools//tools/cpp:compiler": "clang"},
)

PIXMAN_CORE_SOURCES = [
    "pixman/pixman.c",
    "pixman/pixman-access.c",
    "pixman/pixman-edge.c",
    "pixman/pixman-access-accessors.c",
    "pixman/pixman-edge-accessors.c",
    "pixman/pixman-bits-image.c",
    "pixman/pixman-combine32.c",
    "pixman/pixman-combine-float.c",
    "pixman/pixman-conical-gradient.c",
    "pixman/pixman-filter.c",
    "pixman/pixman-fast-path.c",
    "pixman/pixman-general.c",
    "pixman/pixman-x86.c",
    "pixman/pixman-mips.c",
    "pixman/pixman-arm.c",
    "pixman/pixman-ppc.c",
    "pixman/pixman-glyph.c",
    "pixman/pixman-gradient-walker.c",
    "pixman/pixman-image.c",
    "pixman/pixman-implementation.c",
    "pixman/pixman-linear-gradient.c",
    "pixman/pixman-matrix.c",
    "pixman/pixman-noop.c",
    "pixman/pixman-radial-gradient.c",
    "pixman/pixman-region16.c",
    "pixman/pixman-region32.c",
    "pixman/pixman-solid-fill.c",
    "pixman/pixman-timer.c",
    "pixman/pixman-trap.c",
    "pixman/pixman-utils.c",
]

PIXMAN_TEXTUAL_HEADERS = [
    "pixman/pixman-access.c",
    "pixman/pixman-edge.c",
    "pixman/pixman-region.c",
]

PIXMAN_X86_SOURCES = [
    "pixman/pixman-mmx.c",
    "pixman/pixman-sse2.c",
    "pixman/pixman-ssse3.c",
]

PIXMAN_ARM_SOURCES = [
    "pixman/pixman-arm-neon.c",
    "pixman/pixman-arm-simd.c",
]

# These are linux specific, gcc
PIXMAN_ARM_ASM_SOURCES = [
    # "pixman/pixman-arm-neon-asm.S",
    # "pixman/pixman-arm-neon-asm-bilinear.S",
]

PIXMAN_SOURCES = PIXMAN_CORE_SOURCES + select({
    "@platforms//cpu:x86_64": PIXMAN_X86_SOURCES,
    "@platforms//cpu:arm64": PIXMAN_ARM_SOURCES + PIXMAN_ARM_ASM_SOURCES,
    "//conditions:default": [],
})

PIXMAN_HEADERS = [
    "pixman/pixman.h",
    "pixman/pixman-version.h",
    "pixman/pixman-compiler.h",
    "pixman/pixman-private.h",
    "pixman/pixman-accessor.h",
    "pixman/pixman-combine32.h",
    "pixman/pixman-inlines.h",
    "pixman/pixman-edge-imp.h",
    "pixman/dither/blue-noise-64x64.h",
] + select({
    "@platforms//cpu:arm64": [
        # "pixman/pixman-arm-asm.h",
        "pixman/pixman-arm-common.h",
        # "pixman/pixman-arm-neon-asm.h",
        # "pixman/pixman-arm-simd-asm.h", # armv6
    ],
    "//conditions:default": [],
})

cc_library(
    name = "pixman",
    srcs = PIXMAN_SOURCES + PIXMAN_HEADERS,
    hdrs = ["pixman/pixman.h"],
    copts = [
        # Common warnings disabled
        "-Wno-unused-const-variable",
    ] + select({
        "@platforms//cpu:x86_64": [
            # If we assume SSE2, SSSE3, etc. on x86_64:
            "-mmmx",
            "-msse2",
            "-mssse3",
            # Potentially "-msse4.1", etc. if you want to get fancy
            "-Winline",
        ],
        "@platforms//cpu:arm64": [
            # NEON is guaranteed on ARMv8, so we assume it's safe
            "-march=armv8-a+simd+crypto",
            "-fPIC",
            "-DHAVE_AS_ASM_CONVERSION=1",
            "-DHAVE_AS_NEON=1",
            "-D__ARM_NEON__=1",
        ],
        "//conditions:default": [],
    }) + select({
        ":clang_compiler": [
            "-Wno-unused-local-typedef",
            "-Wno-expansion-to-defined",
            "-Wno-incompatible-function-pointer-types",
            "-Wno-incompatible-pointer-types",
            "-Wno-unknown-attributes",
            "-Wno-uninitialized",
            "-Wno-unused-function",
            "-Wno-deprecated-declarations",
            "-Wno-ignored-qualifiers",
            "-Wno-incompatible-pointer-types-discards-qualifiers",
        ],
        "//conditions:default": [
            "-Wno-uninitialized",
            "-Wno-unused-local-typedefs",
        ],
    }),
    includes = [
        "pixman",
        "pixman/dither",
    ],
    linkopts = select({
        "@platforms//cpu:arm64": [
            # Some Apple arm64 builds might need this (optional):
            "-framework CoreGraphics",
        ],
        "//conditions:default": [],
    }),
    local_defines = [
        "PACKAGE=\"pixman\"",
        "TOOLCHAIN_SUPPORTS_ATTRIBUTE_CONSTRUCTOR=1",
        "TLS=__thread",
        "HAVE_BUILTIN_CLZ=1",
        # "HAVE_PTHREAD_SETSPECIFIC",
        # "HAVE_POSIX_MEMALIGN",
        # You could define more detection flags here (like 'HAVE_GCC_VECTOR_EXTENSIONS')
        # if you want to replicate all the Meson checks shipped with pixman
    ] + select({
        "@platforms//cpu:x86_64": [
            # Enable x86 macros
            "USE_X86_MMX=1",
            "USE_SSE2=1",
            "USE_SSSE3=1",
            "SIZEOF_LONG=8",
        ],
        "@platforms//cpu:arm64": [
            # "USE_ARM_NEON=1",
            # "HAVE_ARM_NEON_INTRINSICS=1",
            # "DISABLE_ARM_NEON_ASM=1",
            # "USE_ARM_SIMD=1", # armv6
            "SIZEOF_LONG=8",
        ],
        "//conditions:default": [],
    }),
    textual_hdrs = PIXMAN_TEXTUAL_HEADERS,
    visibility = ["//visibility:public"],
)
