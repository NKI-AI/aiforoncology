package(
    default_visibility = ["//visibility:public"],
)

# List of source files for Minizip
SOURCE_FILES = [
    "mz_crypt.c",
    "mz_crypt_openssl.c",
    # "mz_crypt_winvista.c",
    # "mz_crypt_winxp.c",
    "mz_os.c",
    "mz_os_posix.c",
    # "mz_os_win32.c",
    "mz_strm.c",
    "mz_strm_buf.c",
    "mz_strm_bzip.c",
    "mz_strm_lzma.c",
    "mz_strm_mem.c",
    "mz_strm_os_posix.c",
    # "mz_strm_os_win32.c",
    # "mz_strm_pkcrypt.c",
    "mz_strm_split.c",
    # "mz_strm_wzaes.c",
    "mz_strm_zlib.c", 
    "mz_strm_zstd.c",
    "mz_zip.c",
    # "mz_zip_rw.c",
    # Files from compat
    "compat/ioapi.c",
    "compat/unzip.c",
    "compat/zip.c",
] + select({
    "@platforms//os:macos": [
        "mz_crypt_apple.c",
        "mz_strm_libcomp.c",
    ],
    "//conditions:default": [],
})

# List of header files for Minizip
HEADER_FILES = [
    "compat/crypt.h",
    "compat/ioapi.h",
    "compat/unzip.h",
    "compat/zip.h",
    "mz_crypt.h",
    "mz.h",
    "mz_os.h",
    "mz_strm_buf.h",
    "mz_strm_bzip.h",
    "mz_strm.h",
    "mz_strm_lzma.h",
    "mz_strm_mem.h",
    "mz_strm_os.h",
    # "mz_strm_pkcrypt.h",
    "mz_strm_split.h",
    # "mz_strm_wzaes.h",
    "mz_strm_zlib.h",
    "mz_strm_zstd.h",
    "mz_zip.h",
    "mz_zip_rw.h",
] + select({
    "@platforms//os:macos": [
        "mz_strm_libcomp.h",
    ],
    "//conditions:default": [],
})

cc_library(
    name = "minizip",
    srcs = SOURCE_FILES,
    hdrs = HEADER_FILES,
    copts = [
        "-DMZ_ZLIB",
        "-DZLIB_COMPAT",
        "-DMZ_COMPAT",
        "-DMZ_BZIP2",
        "-DMZ_LZMA",
        "-DMZ_ZSTD",
        # "-DMZ_CRYPTO",
        # "-DMZ_OPENSSL",
        "-D_FILE_OFFSET_BITS=64",  # Enable large file support
    ],
    linkopts = ["-lz", "-pthread", "-ldl"],
    includes = [
        ".",
        "compat",  # Include current directory for headers
    ],
    deps = [
        "@bzip2//:bzip2",
        "@xz//:lzma",
        "@openssl//:crypto",
        "@openssl//:ssl",
        "@openssl//:openssl",
        "@zlib",
        "@zstd",
    ],
)

cc_test(
    name = "minizip_tests",
    srcs = [
        "test/test_compat.cc",
        # "test/test_crypt.cc",
        "test/test_encoding.cc",
        "test/test_file.cc",
        "test/test_main.cc",
        "test/test_path.cc",
        "test/test_stream.cc",
        "test/test_stream_compress.cc",
        # "test/test_stream_crypt.cc",
    ],
    linkopts = ["-lz", "-pthread", "-ldl"],
    copts = [
        "-DMZ_ZLIB",
        "-DMZ_COMPAT",
        "-DMZ_BZIP2",
        "-DMZ_LZMA",
        "-DMZ_ZSTD",
        # "-DMZ_CRYPTO",
        # "-DMZ_OPENSSL",
        "-D_FILE_OFFSET_BITS=64",
    ],
    deps = [
        ":minizip",
        "@bzip2//:bzip2",
        "@xz//:lzma",
        "@openssl//:crypto",
        "@openssl//:ssl",
        "@openssl//:openssl",
        "@zlib",
        "@zstd",
        "@googletest//:gtest",
        "@googletest//:gtest_main",
    ]
)

cc_binary(
    name = "minizip_bin",
    srcs = ["minizip.c"],
    deps = [":minizip"],
)

cc_binary(
    name = "minigzip_bin",
    srcs = ["minigzip.c"],
    deps = [":minizip"],
)