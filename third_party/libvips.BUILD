load("@//bazel:system.bzl", "cc_system_headers")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

package(default_visibility = ["//visibility:public"])

config_setting(
    name = "clang_compiler",
    flag_values = {"@bazel_tools//tools/cpp:compiler": "clang"},
)

VIPS_MAJOR_VERSION = "8"

VIPS_MINOR_VERSION = "16"

VIPS_MICRO_VERSION = "0"

VIPS_LIBRARY_CURRENT = "60"

VIPS_LIBRARY_REVISION = "0"

VIPS_LIBRARY_AGE = "18"

LIBVIPS_FOREIGN_SOURCES = [
    "libvips/foreign/analyze2vips.c",
    "libvips/foreign/analyzeload.c",
    "libvips/foreign/archive.c",
    "libvips/foreign/cairo.c",
    # "libvips/foreign/cgifsave.c",
    "libvips/foreign/csvload.c",
    "libvips/foreign/csvsave.c",
    "libvips/foreign/dzsave.c",
    "libvips/foreign/exif.c",
    # "libvips/foreign/fits.c",
    # "libvips/foreign/fitsload.c",
    # "libvips/foreign/fitssave.c",
    "libvips/foreign/foreign.c",
    # "libvips/foreign/heifload.c",
    # "libvips/foreign/heifsave.c",
    "libvips/foreign/jp2kload.c",
    "libvips/foreign/jp2ksave.c",
    "libvips/foreign/jpeg2vips.c",
    "libvips/foreign/jpegload.c",
    "libvips/foreign/jpegsave.c",
    # "libvips/foreign/jxlload.c",
    # "libvips/foreign/jxlsave.c",
    # "libvips/foreign/magick.c",
    # "libvips/foreign/magick2vips.c",
    # "libvips/foreign/magick6load.c",
    # "libvips/foreign/magick7load.c",
    # "libvips/foreign/magickload.c",
    # "libvips/foreign/magicksave.c",
    # "libvips/foreign/matlab.c",
    # "libvips/foreign/matload.c",
    "libvips/foreign/matrixload.c",
    "libvips/foreign/matrixsave.c",
    "libvips/foreign/niftiload.c",
    "libvips/foreign/niftisave.c",
    "libvips/foreign/nsgifload.c",
    # "libvips/foreign/openexr2vips.c",
    # "libvips/foreign/openexrload.c",
    "libvips/foreign/openslideload.c",
    # "libvips/foreign/pdf.c",
    # "libvips/foreign/pdfiumload.c",
    "libvips/foreign/pngload.c",
    "libvips/foreign/pngsave.c",
    # "libvips/foreign/popplerload.c",
    "libvips/foreign/ppmload.c",
    "libvips/foreign/ppmsave.c",
    "libvips/foreign/quantise.c",
    "libvips/foreign/radiance.c",
    "libvips/foreign/radload.c",
    "libvips/foreign/radsave.c",
    "libvips/foreign/rawload.c",
    "libvips/foreign/rawsave.c",
    "libvips/foreign/spngload.c",
    "libvips/foreign/spngsave.c",
    # "libvips/foreign/svgload.c",
    "libvips/foreign/tiff.c",
    "libvips/foreign/tiff2vips.c",
    "libvips/foreign/tiffload.c",
    "libvips/foreign/tiffsave.c",
    "libvips/foreign/vips2jpeg.c",
    # "libvips/foreign/vips2magick.c",
    "libvips/foreign/vips2tiff.c",
    "libvips/foreign/vipsload.c",
    "libvips/foreign/vipspng.c",
    "libvips/foreign/vipssave.c",
    "libvips/foreign/webp2vips.c",
    "libvips/foreign/webpload.c",
    "libvips/foreign/webpsave.c",
]

LIBVIPS_HISTOGRAM_SOURCES = [
    "libvips/histogram/hist_plot.c",
    "libvips/histogram/stdif.c",
    "libvips/histogram/histogram.c",
    "libvips/histogram/hist_match.c",
    "libvips/histogram/maplut.c",
    "libvips/histogram/hist_unary.c",
    "libvips/histogram/case.c",
    "libvips/histogram/hist_local.c",
    "libvips/histogram/hist_ismonotonic.c",
    "libvips/histogram/hist_cum.c",
    "libvips/histogram/hist_equal.c",
    "libvips/histogram/hist_entropy.c",
    "libvips/histogram/hist_norm.c",
    "libvips/histogram/percent.c",
]

LIBVIPS_RESAMPLE_SOURCES = glob(["libvips/resample/*.c"]) + glob(["libvips/resample/*.cpp"])

LIBVIPS_CONVOLUTION_SOURCES = glob(["libvips/convolution/*.c"]) + glob(["libvips/convolution/*.cpp"])

LIBVIPS_CONVERSION_SOURCES = glob(["libvips/conversion/*.c"]) + ["libvips/conversion/composite.cpp"]

LIBVIPS_MODULE_SOURCES = glob(["libvips/module/*.c"])

LIBVIPS_MOSAICING_SOURCES = glob(["libvips/mosaicing/*.c"])

LIBVIPS_FREQFILT_SOURCES = glob(["libvips/freqfilt/*.c"])

LIBVIPS_DRAW_SOURCES = glob(["libvips/draw/*.c"])

LIBVIPS_COLOUR_SOURCES = glob(["libvips/colour/*.c"])

LIBVIPS_ARITHMETIC_SOURCES = glob(["libvips/arithmetic/*.c"])

LIBVIPS_MORPHOLOGY_SOURCES = glob(["libvips/morphology/*.c"]) + ["libvips/morphology/morph_hwy.cpp"]

LIBVIPS_CREATE_SOURCES = [
    "libvips/create/black.c",
    "libvips/create/buildlut.c",
    "libvips/create/create.c",
    "libvips/create/eye.c",
    "libvips/create/fractsurf.c",
    "libvips/create/gaussmat.c",
    "libvips/create/gaussnoise.c",
    "libvips/create/grey.c",
    "libvips/create/identity.c",
    "libvips/create/invertlut.c",
    "libvips/create/logmat.c",
    "libvips/create/mask.c",
    "libvips/create/mask_butterworth.c",
    "libvips/create/mask_butterworth_band.c",
    "libvips/create/mask_butterworth_ring.c",
    "libvips/create/mask_fractal.c",
    "libvips/create/mask_gaussian.c",
    "libvips/create/mask_gaussian_band.c",
    "libvips/create/mask_gaussian_ring.c",
    "libvips/create/mask_ideal.c",
    "libvips/create/mask_ideal_band.c",
    "libvips/create/mask_ideal_ring.c",
    "libvips/create/perlin.c",
    "libvips/create/point.c",
    "libvips/create/sdf.c",
    "libvips/create/sines.c",
    "libvips/create/text.c",
    "libvips/create/tonelut.c",
    "libvips/create/worley.c",
    "libvips/create/xyz.c",
    "libvips/create/zone.c",
]

LIBVIPS_IOFUNCS_SOURCES = glob([
    "libvips/iofuncs/*.c",
]) + ["libvips/iofuncs/vector.cpp"]

LIBVIPS_DEPRECATED_SOURCES = [
    "libvips/deprecated/arith_dispatch.c",
    "libvips/deprecated/cimg_dispatch.c",
    "libvips/deprecated/colour_dispatch.c",
    "libvips/deprecated/conver_dispatch.c",
    "libvips/deprecated/convol_dispatch.c",
    "libvips/deprecated/cooc_funcs.c",
    "libvips/deprecated/deprecated_dispatch.c",
    "libvips/deprecated/dispatch_types.c",
    "libvips/deprecated/fits.c",
    "libvips/deprecated/format.c",
    "libvips/deprecated/format_dispatch.c",
    "libvips/deprecated/freq_dispatch.c",
    "libvips/deprecated/glds_funcs.c",
    "libvips/deprecated/hist_dispatch.c",
    "libvips/deprecated/im_align_bands.c",
    "libvips/deprecated/im_analyze2vips.c",
    "libvips/deprecated/im_benchmark.c",
    "libvips/deprecated/im_bernd.c",
    "libvips/deprecated/im_clamp.c",
    "libvips/deprecated/im_cmulnorm.c",
    "libvips/deprecated/im_convsub.c",
    "libvips/deprecated/im_csv2vips.c",
    "libvips/deprecated/im_debugim.c",
    "libvips/deprecated/im_dif_std.c",
    "libvips/deprecated/im_exr2vips.c",
    "libvips/deprecated/im_fav4.c",
    # "libvips/deprecated/im_freq_mask.c",
    "libvips/deprecated/im_freq_mask.c",
    "libvips/deprecated/im_gadd.c",
    "libvips/deprecated/im_gaddim.c",
    "libvips/deprecated/im_gfadd.c",
    "libvips/deprecated/im_gradcor.c",
    "libvips/deprecated/im_jpeg2vips.c",
    "libvips/deprecated/im_lab_morph.c",
    "libvips/deprecated/im_line.c",
    "libvips/deprecated/im_linreg.c",
    "libvips/deprecated/im_litecor.c",
    "libvips/deprecated/im_magick2vips.c",
    "libvips/deprecated/im_mask2vips.c",
    "libvips/deprecated/im_matcat.c",
    "libvips/deprecated/im_matinv.c",
    "libvips/deprecated/im_matmul.c",
    "libvips/deprecated/im_mattrn.c",
    "libvips/deprecated/im_maxpos_avg.c",
    "libvips/deprecated/im_maxpos_subpel.c",
    "libvips/deprecated/im_measure.c",
    "libvips/deprecated/im_nifti2vips.c",
    "libvips/deprecated/im_openslide2vips.c",
    "libvips/deprecated/im_png2vips.c",
    "libvips/deprecated/im_point_bilinear.c",
    "libvips/deprecated/im_ppm2vips.c",
    "libvips/deprecated/im_print.c",
    "libvips/deprecated/im_printlines.c",
    "libvips/deprecated/im_resize_linear.c",
    "libvips/deprecated/im_setbox.c",
    "libvips/deprecated/im_simcontr.c",
    "libvips/deprecated/im_slice.c",
    "libvips/deprecated/im_spatres.c",
    "libvips/deprecated/im_stretch3.c",
    "libvips/deprecated/im_thresh.c",
    "libvips/deprecated/im_tiff2vips.c",
    "libvips/deprecated/im_video_test.c",
    "libvips/deprecated/im_vips2csv.c",
    "libvips/deprecated/im_vips2dz.c",
    "libvips/deprecated/im_vips2jpeg.c",
    "libvips/deprecated/im_vips2mask.c",
    "libvips/deprecated/im_vips2png.c",
    "libvips/deprecated/im_vips2ppm.c",
    "libvips/deprecated/im_vips2tiff.c",
    "libvips/deprecated/im_vips2webp.c",
    "libvips/deprecated/im_webp2vips.c",
    "libvips/deprecated/im_zerox.c",
    "libvips/deprecated/inplace_dispatch.c",
    "libvips/deprecated/lazy.c",
    "libvips/deprecated/mask_dispatch.c",
    "libvips/deprecated/matalloc.c",
    "libvips/deprecated/matlab.c",
    "libvips/deprecated/morph_dispatch.c",
    "libvips/deprecated/mosaicing_dispatch.c",
    "libvips/deprecated/other_dispatch.c",
    "libvips/deprecated/package.c",
    "libvips/deprecated/radiance.c",
    "libvips/deprecated/raw.c",
    "libvips/deprecated/rename.c",
    "libvips/deprecated/resample_dispatch.c",
    "libvips/deprecated/rotmask.c",
    "libvips/deprecated/rw_mask.c",
    "libvips/deprecated/tone.c",
    "libvips/deprecated/video_dispatch.c",
    "libvips/deprecated/vips7compat.c",
]

LIBVIPS_MOSAICING_HEADERS = [
    "libvips/mosaicing/global_balance.h",
    "libvips/mosaicing/pmosaicing.h",
]

LIBVIPS_FREQFILT_HEADERS = [
    "libvips/freqfilt/pfreqfilt.h",
]

LIBVIPS_RESAMPLE_HEADERS = [
    "libvips/resample/presample.h",
    "libvips/resample/templates.h",
]

LIBVIPS_INCLUDE_HEADERS = glob(["libvips/include/vips/*.h"])

LIBVIPS_CONVERSION_HEADERS = [
    "libvips/conversion/pconversion.h",
    "libvips/conversion/bandary.h",
]

LIBVIPS_ARITHMETIC_HEADERS = glob(["libvips/arithmetic/*.h"])

LIBVIPS_MORPHOLOGY_HEADERS = [
    "libvips/morphology/pmorphology.h",
]

LIBVIPS_CREATE_HEADERS = [
    "libvips/create/point.h",
    "libvips/create/pmask.h",
    "libvips/create/pcreate.h",
]

LIBVIPS_IOFUNCS_HEADERS = [
    "libvips/iofuncs/sink.h",
]

LIBVIPS_FOREIGN_HEADERS = [
    "libvips/foreign/dbh.h",
    "libvips/foreign/tiff.h",
    # "libvips/foreign/magick.h",
    "libvips/foreign/jpeg.h",
    "libvips/foreign/pforeign.h",
    "libvips/foreign/quantise.h",
    "libvips/foreign/libnsgif/lzw.h",
    "libvips/foreign/libnsgif/nsgif.h",
]

LIBVIPS_CONVOLUTION_HEADERS = [
    "libvips/convolution/pconvolution.h",
    "libvips/convolution/correlation.h",
]

LIBVIPS_DRAW_HEADERS = [
    "libvips/draw/pdraw.h",
    "libvips/draw/drawink.h",
]

LIBVIPS_COLOUR_HEADERS = [
    "libvips/colour/profiles.h",
    "libvips/colour/pcolour.h",
]

LIBVIPS_HISTOGRAM_HEADERS = [
    "libvips/histogram/phistogram.h",
    "libvips/histogram/hist_unary.h",
]

LIBVIPS_SOURCES = (
    LIBVIPS_FOREIGN_SOURCES +
    LIBVIPS_HISTOGRAM_SOURCES +
    LIBVIPS_RESAMPLE_SOURCES +
    LIBVIPS_CONVOLUTION_SOURCES +
    LIBVIPS_CONVERSION_SOURCES +
    LIBVIPS_MODULE_SOURCES +
    LIBVIPS_MOSAICING_SOURCES +
    LIBVIPS_FREQFILT_SOURCES +
    LIBVIPS_DRAW_SOURCES +
    LIBVIPS_COLOUR_SOURCES +
    LIBVIPS_ARITHMETIC_SOURCES +
    LIBVIPS_MORPHOLOGY_SOURCES +
    LIBVIPS_CREATE_SOURCES +
    LIBVIPS_IOFUNCS_SOURCES
    # LIBVIPS_DEPRECATED_SOURCES
)

LIBVIPS_HEADERS = (
    LIBVIPS_MOSAICING_HEADERS +
    LIBVIPS_FREQFILT_HEADERS +
    LIBVIPS_RESAMPLE_HEADERS +
    LIBVIPS_INCLUDE_HEADERS +
    LIBVIPS_CONVERSION_HEADERS +
    LIBVIPS_ARITHMETIC_HEADERS +
    LIBVIPS_MORPHOLOGY_HEADERS +
    LIBVIPS_CREATE_HEADERS +
    LIBVIPS_IOFUNCS_HEADERS +
    LIBVIPS_FOREIGN_HEADERS +
    LIBVIPS_CONVOLUTION_HEADERS +
    LIBVIPS_DRAW_HEADERS +
    LIBVIPS_COLOUR_HEADERS +
    LIBVIPS_HISTOGRAM_HEADERS
)

expand_template(
    name = "generate_version_h",
    out = "libvips/include/vips/version.h",
    substitutions = {
        "@VIPS_VERSION@": "{}.{}.{}".format(VIPS_MAJOR_VERSION, VIPS_MINOR_VERSION, VIPS_MICRO_VERSION),
        "@VIPS_VERSION_STRING@": "{}.{}.{}".format(VIPS_MAJOR_VERSION, VIPS_MINOR_VERSION, VIPS_MICRO_VERSION),
        "@VIPS_MAJOR_VERSION@": "{}".format(VIPS_MAJOR_VERSION),
        "@VIPS_MINOR_VERSION@": "{}".format(VIPS_MINOR_VERSION),
        "@VIPS_MICRO_VERSION@": "{}".format(VIPS_MICRO_VERSION),
        "@LIBRARY_CURRENT@": "{}".format(VIPS_LIBRARY_CURRENT),
        "@LIBRARY_REVISION@": "{}".format(VIPS_LIBRARY_REVISION),
        "@LIBRARY_AGE@": "{}".format(VIPS_LIBRARY_AGE),
        "@VIPS_CONFIG@": "built with bazel",
        "@VIPS_ENABLE_DEPRECATED@": "0",
    },
    template = "libvips/include/vips/version.h.in",
)

genrule(
    name = "generate_config_h",
    outs = ["config.h"],
    cmd = "\n".join([
        "cat <<'EOF' > $@",
        "#pragma once",
        "#undef ENABLE_DEPRECATED",
        "#undef ENABLE_MAGICKLOAD",
        "#undef ENABLE_MAGICKSAVE",
        "#define ENABLE_MODULES 1",
        '#define GETTEXT_PACKAGE "vips{}.{}"'.format(VIPS_MAJOR_VERSION, VIPS_MINOR_VERSION),
        '#define G_LOG_DOMAIN "VIPS"',
        "#define HAVE_ANALYZE 1",
        "#undef HAVE_BIND_TEXTDOMAIN_CODESET",
        "#undef HAVE_CFITSIO",
        "#undef HAVE_CGIF",
        # "#define HAVE_CGIF_ATTR_NO_LOOP 0",
        # "#define HAVE_CGIF_FRAME_ATTR_INTERLACED 0",
        "#define HAVE_EXIF 1",
        "#define HAVE_EXIF_0_6_22 1",
        "#define HAVE_EXIF_0_6_23 1",
        "#undef HAVE_FFTW",
        "#undef HAVE_FONTCONFIG",
        "#undef HAVE_HEIF",  # All HEIF things were 1
        # "#define HAVE_HEIF_AVIF 0",
        # "#define HAVE_HEIF_COLOR_PROFILE 0",
        # "#define HAVE_HEIF_DECODING_OPTIONS_CONVERT_HDR_TO_8BIT 0",
        # "#define HAVE_HEIF_ENCODER_PARAMETER_GET_VALID_INTEGER_VALUES 0",
        # "#define HAVE_HEIF_ENCODING_OPTIONS_OUTPUT_NCLX_PROFILE 0",
        # "#define HAVE_HEIF_ERROR_SUCCESS 0",
        # "#define HAVE_HEIF_INIT 0",
        # "#define HAVE_HEIF_SET_MAX_IMAGE_SIZE_LIMIT 0",
        "#define HAVE_HWY 1",
        "#define HAVE_HWY_1_1_0 1",
        "#undef HAVE_IMAGEQUANT",  # Need to fix this
        "#define HAVE_IMAGESTOBLOB 1",
        "#define HAVE_IMPORTIMAGEPIXELS 1",
        "#define HAVE_JPEG 1",
        "#define HAVE_LCMS2 1",
        "#undef HAVE_LIBJXL",
        "#define HAVE_LIBJXL_0_7 0",
        "#define HAVE_LIBJXL_0_9 0",
        "#define HAVE_LIBOPENJP2 1",
        "#define HAVE_LIBWEBP 1",
        "#undef HAVE_MAGICK7",
        "#undef HAVE_MATIO",
        "#define HAVE_NSGIF 1",
        "#undef HAVE_OPENEXR",
        "#define HAVE_OPENSLIDE 1",
        "#define HAVE_OPENSLIDE_3_4 1",  # This means the version is 3.4 or higher
        "#define HAVE_OPENSLIDE_ICC 1",  # This is true for openslide 4.0.0
        "#undef HAVE_PANGOCAIRO",
        "#undef HAVE_POPPLER",
        "#define HAVE_POSIX_MEMALIGN 1",
        "#undef HAVE_PPM",
        "#undef HAVE_RADIANCE",
        "#undef HAVE_RSVG",
        "#define HAVE_SPNG 1",
        "#define HAVE_SYS_FILE_H 1",
        "#define HAVE_SYS_MMAN_H 1",
        "#define HAVE_SYS_PARAM_H 1",
        "#define HAVE_TIFF 1",
        "#define HAVE_TIFF_COMPRESSION_WEBP 1",
        "#define HAVE_UNISTD_H 1",
        "#define HAVE_VECTOR_ARITH 1",
        "#define HAVE_ZLIB 1",
        "#define HEIF_MODULE 1",
        "#define LIBJXL_MODULE 1",
        "#define MAGICK_MODULE 1",
        "#undef OPENSLIDE_MODULE",
        "#define POPPLER_MODULE 1",
        '#define VIPS_ICC_DIR "/Library/ColorSync/Profiles"',
        '#define VIPS_LIBDIR "$(BINDIR)"',
        '#define VIPS_PREFIX "$(BINDIR)"',
        '#define _VIPS_PUBLIC __attribute__((visibility("default")))',
        "EOF",
    ]),
)

cc_system_headers(
    name = "libvips_headers_config",
    hdrs = ["config.h"],
)

ENUMTYPES_HEADER = [
    "libvips/include/vips/resample.h",
    "libvips/include/vips/memory.h",
    "libvips/include/vips/create.h",
    "libvips/include/vips/foreign.h",
    "libvips/include/vips/arithmetic.h",
    "libvips/include/vips/conversion.h",
    "libvips/include/vips/util.h",
    "libvips/include/vips/image.h",
    "libvips/include/vips/colour.h",
    "libvips/include/vips/operation.h",
    "libvips/include/vips/convolution.h",
    "libvips/include/vips/morphology.h",
    "libvips/include/vips/draw.h",
    "libvips/include/vips/basic.h",
    "libvips/include/vips/object.h",
    "libvips/include/vips/region.h",
    # "libvips/include/vips/almostdeprecated.h",
]

genrule(
    name = "generate_libvips_enumtypes_h",
    srcs = ["libvips/include/vips/enumtypes.h.in"] + ENUMTYPES_HEADER,
    outs = ["libvips/include/vips/enumtypes.h"],
    cmd = "$(location @glib2//:gobject_glib_mkenums_tool) --template $(SRCS) > $@",
    tools = ["@glib2//:gobject_glib_mkenums_tool"],
)

genrule(
    name = "generate_libvips_enumtypes_c",
    srcs = ["libvips/include/vips/enumtypes.c.in"] + ENUMTYPES_HEADER,
    outs = ["libvips/include/vips/enumtypes.c"],
    cmd = "$(location @glib2//:gobject_glib_mkenums_tool) --template $(SRCS) > $@",
    tools = ["@glib2//:gobject_glib_mkenums_tool"],
)

genrule(
    name = "generate_libvips_marshal_h",
    srcs = [
        "libvips/iofuncs/vipsmarshal.list",
    ],
    outs = ["libvips/iofuncs/vipsmarshal.h"],
    cmd = "$(location @glib2//:gobject_glib_genmarshal_tool) --header --prefix vips $(SRCS) --output $@",
    tools = ["@glib2//:gobject_glib_genmarshal_tool"],
)

genrule(
    name = "generate_libvips_marshal_c",
    srcs = [
        "libvips/iofuncs/vipsmarshal.list",
    ],
    outs = ["libvips/iofuncs/vipsmarshal.c"],
    cmd = "$(location @glib2//:gobject_glib_genmarshal_tool) --body --prefix vips $(SRCS) --output $@",
    tools = ["@glib2//:gobject_glib_genmarshal_tool"],
)

genrule(
    name = "libvips_include_dir",
    srcs = LIBVIPS_HEADERS,
    outs = ["vips_include"],
    cmd = "mkdir -p $@ && cp -r $(SRCS) $@/",
    visibility = ["//visibility:public"],
)

cc_library(
    name = "libvips",
    srcs = LIBVIPS_SOURCES + [
        "libvips/include/vips/enumtypes.c",
        "libvips/iofuncs/vipsmarshal.c",
    ],
    hdrs = LIBVIPS_HEADERS + [
        "config.h",
        "libvips/include/vips/enumtypes.h",
        ":generate_libvips_enumtypes_h",
        ":generate_libvips_marshal_h",
        ":generate_version_h",
    ],
    copts = [
        "-DHAVE_CONFIG_H=1",
    ],
    includes = [
        "libvips",
        "libvips/arithmetic",
        "libvips/colour",
        "libvips/conversion",
        "libvips/convolution",
        "libvips/create",
        "libvips/draw",
        "libvips/foreign",
        "libvips/foreign/libnsgif",
        "libvips/foreign/libnsgif/include",
        "libvips/freqfilt",
        "libvips/histogram",
        "libvips/include",
        "libvips/iofuncs",
        "libvips/memory",
        "libvips/module",
        "libvips/morphology",
        "libvips/mosaicing",
        "libvips/resample",
    ],
    linkopts = [
        "-lexpat",
        "-lz",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":generate_libvips_enumtypes_c",
        ":libvips_headers_config",
        "@glib2//:gio",
        "@glib2//:glib",
        "@glib2//:gobject",
        "@hwy",
        "@knusperli//:quantize",
        "@lcms2",
        "@libexif",
        "@libjpeg_turbo//:jpeg",
        "@libnsgif",
        "@libpng//:png",
        "@libspng",
        "@libtiff",
        "@libwebp",
        "@openslide",
        "@zlib",
    ],
)

filegroup(
    name = "libvips-headers",
    srcs = glob([
        "libvips/include/**/*.h",
    ]),
    visibility = ["//visibility:public"],
)

cc_binary(
    name = "libvips-shared-impl",
    srcs = LIBVIPS_SOURCES + LIBVIPS_HEADERS,
    linkshared = True,
    linkstatic = True,
    visibility = ["//visibility:public"],
    deps = [
        ":libvips",
    ],
)

genrule(
    name = "libvips-shared-macos",
    srcs = [":libvips-shared-impl"],
    outs = ["libvips.42.dylib"],
    cmd = "cp $(location :libvips-shared-impl) $@",
    visibility = ["//visibility:public"],
)

genrule(
    name = "libvips-shared-linux",
    srcs = [":libvips-shared-impl"],
    outs = ["libvips.so.42"],
    cmd = "cp $(location :libvips-shared-impl) $@",
    visibility = ["//visibility:public"],
)

cc_library(
    name = "libvips-cpp",
    srcs = [
        "cplusplus/VConnection.cpp",
        "cplusplus/VError.cpp",
        "cplusplus/VImage.cpp",
        "cplusplus/VInterpolate.cpp",
        "cplusplus/VRegion.cpp",
    ],
    hdrs = glob(["cplusplus/include/vips/*.h"]) + [
        "cplusplus/include/vips/vips8",
        "cplusplus/vips-operators.cpp",
    ],
    copts = [
        "-std=c++20",
    ],
    includes = [
        "cplusplus",
        "cplusplus/include",
        "cplusplus/include/vips",
        "libvips",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":libvips",
    ],
)
