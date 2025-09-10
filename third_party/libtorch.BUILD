package(default_visibility = ["//visibility:public"])

cc_library(
    name = "libtorch",
    srcs = select({
        "@platforms//os:macos": glob(["lib/lib*.dylib*"]),
        "//conditions:default": glob(["lib/lib*.so*"]),
    }),
    hdrs = glob([
        "include/**/*.h",
        "include/**/*.cuh",
    ]),
    includes = [
        "include",
        "include/torch/csrc/api/include",
    ],
    deps = [
        "@rules_python//python/cc:current_py_cc_headers",
        "@rules_python//python/cc:current_py_cc_libs",
    ],
)
