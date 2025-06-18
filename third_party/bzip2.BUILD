package(
    default_visibility = ["//visibility:public"],
)

# Source files for bzip2 library
BZIP2_SRCS = [
    "blocksort.c",
    "huffman.c",
    "crctable.c",
    "randtable.c",
    "compress.c",
    "decompress.c",
    "bzlib.c",
]

BZIP2_HDRS = [
    "bzlib.h",
    "bzlib_private.h",
]

cc_library(
    name = "bzip2",
    srcs = BZIP2_SRCS,
    hdrs = BZIP2_HDRS,
    copts = [
        "-Wall",
        "-Winline",
        "-O2",
        "-g",
        "-D_FILE_OFFSET_BITS=64",  # Enable large file support
    ],
    includes = ["."],
    visibility = ["//visibility:public"],
)

cc_binary(
    name = "bzip2_bin",
    srcs = ["bzip2.c"],
    deps = [":bzip2"],
)

cc_binary(
    name = "bzip2recover",
    srcs = ["bzip2recover.c"],
)
