load("@rules_cc//cc:defs.bzl", "cc_library")

package(default_visibility = ["//visibility:public"])

cc_library(
    name = "libspng",
    srcs = [
        "spng/spng.c",
    ],
    hdrs = [
        "spng/spng.h",
    ],
    copts = [
        "-DSPNG_STATIC",  # Define SPNG_STATIC for the static library
        "-std=c99",
    ],
    includes = ["spng"],
    deps = [
        "@zlib",
    ],
)
