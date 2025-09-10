package(default_visibility = ["//visibility:public"])

cc_library(
    name = "lcms2",
    srcs = [
        # Source files
        "src/cmsalpha.c",
        "src/cmscam02.c",
        "src/cmscgats.c",
        "src/cmscnvrt.c",
        "src/cmserr.c",
        "src/cmsgamma.c",
        "src/cmsgmt.c",
        "src/cmshalf.c",
        "src/cmsintrp.c",
        "src/cmsio0.c",
        "src/cmsio1.c",
        "src/cmslut.c",
        "src/cmsmd5.c",
        "src/cmsmtrx.c",
        "src/cmsnamed.c",
        "src/cmsopt.c",
        "src/cmspack.c",
        "src/cmspcs.c",
        "src/cmsplugin.c",
        "src/cmssamp.c",
        "src/cmssm.c",
        "src/cmstypes.c",
        "src/cmsvirt.c",
        "src/cmswtpnt.c",
        "src/cmsxform.c",
    ],
    hdrs = [
        "include/lcms2.h",
        "include/lcms2_plugin.h",
        "src/lcms2_internal.h",
    ],
    copts = [
        "-DHAVE_FUNC_ATTRIBUTE_VISIBILITY=1",
        "-DWORDS_BIGENDIAN=0",  # Adjust based on the host machine's endianness
        "-DHAVE_GMTIME_R=1",  # Adjust based on your platform
        "-DHAVE_TIMESPEC_GET=1",
        "-DCMS_DONT_USE_SSE2=1",  # Adjust based on the host machine's SSE2 support
    ],
    includes = ["include"],
    visibility = ["//visibility:public"],
    deps = [
        "@zlib",  # Add zlib dependency if required
        # "@libjpeg",   # Link against libjpeg if JPEG support is needed
        "@libtiff",  # Link against libtiff if TIFF support is needed
    ],
)
