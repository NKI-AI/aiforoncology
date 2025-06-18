cc_library(
    name = "libintl",
    srcs = [
        "libintl.c",
    ],
    hdrs = [
        "libintl.h",
    ],
    includes = ["."],
    linkstatic = True,
    local_defines = [
        "STUB_ONLY",
        "G_INTL_STATIC_COMPILATION",
    ],
    visibility = ["//visibility:public"],
)
