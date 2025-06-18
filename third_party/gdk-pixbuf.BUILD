load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

licenses(["restricted"])  # GPL v2.1 license

exports_files(["COPYING"])

GDKPIXBUF_FEATURES_SUBSTITUTIONS = {
    "@GDK_PIXBUF_MAJOR@": "2",
    "@GDK_PIXBUF_MINOR@": "42",
    "@GDK_PIXBUF_MICRO@": "9",
    "@GDK_PIXBUF_VERSION@": "2.42.9",
}

expand_template(
    name = "gdk_pixbuf_features_h",
    out = "gdk-pixbuf/gdk-pixbuf-features.h",
    substitutions = GDKPIXBUF_FEATURES_SUBSTITUTIONS,
    template = "gdk-pixbuf/gdk-pixbuf-features.h.in",
)

genrule(
    name = "config_h",
    outs = ["config.h"],
    cmd = "\n".join([
        "cat <<\"EOF\" >$@",
        "#ifndef __CONFIG_H_",
        "#define __CONFIG_H_",
        "#define GETTEXT_PACKAGE \"gdk-pixbuf\"",
        "#define HAVE_BIND_TEXTDOMAIN_CODESET 1",
        "#define HAVE_LRINT 1",
        "#define HAVE_ROUND 1",
        "#define HAVE_SETRLIMIT 1",
        "#define HAVE_SIGSETJMP",
        "#define HAVE_SYS_RESOURCE_H 1",
        "#define HAVE_SYS_TIME_H 1",
        "#define HAVE_UNISTD_H 1",
        "#define OS_LINUX 1",
        "#define USE_GMODULE",
        "#define _GDK_PIXBUF_EXTERN __attribute__((visibility(\"default\"))) extern",
        "#endif  // __CONFIG_H",
        "EOF",
    ]),
)

GDKPIXBUF_FEATURES = [
    "gdk-pixbuf/gdk-pixbuf-features.h",
]

GDKPIXBUF_HEADERS = [
    "gdk-pixbuf/gdk-pixbuf.h",
    "gdk-pixbuf/gdk-pixbuf-animation.h",
    "gdk-pixbuf/gdk-pixbuf-autocleanups.h",
    "gdk-pixbuf/gdk-pixbuf-core.h",
    "gdk-pixbuf/gdk-pixbuf-io.h",
    "gdk-pixbuf/gdk-pixbuf-loader.h",
    "gdk-pixbuf/gdk-pixbuf-macros.h",
    "gdk-pixbuf/gdk-pixbuf-simple-anim.h",
    "gdk-pixbuf/gdk-pixbuf-transform.h",
]

GDKPIXBUF_SOURCES = [
    "gdk-pixbuf/gdk-pixbuf.c",
    "gdk-pixbuf/gdk-pixbuf-animation.c",
    "gdk-pixbuf/gdk-pixbuf-data.c",
    "gdk-pixbuf/gdk-pixbuf-io.c",
    "gdk-pixbuf/gdk-pixbuf-loader.c",
    "gdk-pixbuf/gdk-pixbuf-scale.c",
    "gdk-pixbuf/gdk-pixbuf-simple-anim.c",
    "gdk-pixbuf/gdk-pixbuf-scaled-anim.c",
    "gdk-pixbuf/gdk-pixbuf-util.c",
]

GDKPIXBUF_HEADERS_PRIVATE = [
    "gdk-pixbuf/gdk-pixbuf-private.h",
    "gdk-pixbuf/gdk-pixdata.h",
    "gdk-pixbuf/gdk-pixbuf-scaled-anim.h",
]

GDKPIXDATA_SOURCES = [
    "gdk-pixbuf/gdk-pixdata.c",
]

cc_library(
    name = "pixops",
    srcs = [
        "config.h",
        "gdk-pixbuf/pixops/pixops.c",
    ],
    hdrs = ["gdk-pixbuf/pixops/pixops.h"],
    strip_include_prefix = "gdk-pixbuf",
    textual_hdrs = ["gdk-pixbuf/fallback-c89.c"],
    deps = [
        "@glib2//:glib",
    ],
)

genrule(
    name = "gdkpixbuf_enums_h",
    srcs = ["gdk-pixbuf/gdk-pixbuf-enum-types.h.template"] + GDKPIXBUF_HEADERS,
    outs = ["gdk-pixbuf/gdk-pixbuf-enum-types.h"],
    cmd = "$(location @glib2//:gobject_glib_mkenums_tool) --template $(SRCS) > $@",
    tools = ["@glib2//:gobject_glib_mkenums_tool"],
)

genrule(
    name = "gdkpixbuf_enums_c",
    srcs = ["gdk-pixbuf/gdk-pixbuf-enum-types.c.template"] + GDKPIXBUF_HEADERS,
    outs = ["gdk-pixbuf/gdk-pixbuf-enum-types.c"],
    cmd = "$(location @glib2//:gobject_glib_mkenums_tool) --template $(SRCS) > $@",
    tools = ["@glib2//:gobject_glib_mkenums_tool"],
)

cc_library(
    name = "gdkpixbuf_enums",
    srcs = [
        "config.h",
        "gdk-pixbuf/gdk-pixbuf-enum-types.c",
    ],
    hdrs = ["gdk-pixbuf/gdk-pixbuf-enum-types.h"],
    strip_include_prefix = "gdk-pixbuf",
    deps = [
        ":gdk_pixbuf_sub_headers",
        "@glib2//:gio",
        "@glib2//:glib",
        "@glib2//:gobject",
    ],
)

genrule(
    name = "gdkpixbuf_marshals_h",
    srcs = ["gdk-pixbuf/gdk-pixbuf-marshal.list"],
    outs = ["gdk-pixbuf/gdk-pixbuf-marshal.h"],
    cmd = "$(location @glib2//:gobject_glib_genmarshal_tool) " +
          "--prefix _gdk_pixbuf_marshal --pragma-once --header $(SRCS) " +
          "--output $@",
    tools = ["@glib2//:gobject_glib_genmarshal_tool"],
)

genrule(
    name = "gdkpixbuf_marshals_c",
    srcs = [
        "gdk-pixbuf/gdk-pixbuf-marshal.list",
        "gdk-pixbuf/gdk-pixbuf-marshal.h",
    ],
    outs = ["gdk-pixbuf/gdk-pixbuf-marshal.c"],
    cmd = "$(location @glib2//:gobject_glib_genmarshal_tool) " +
          "--prefix _gdk_pixbuf_marshal " +
          "--body $(location gdk-pixbuf/gdk-pixbuf-marshal.list) " +
          "--include-header gdk-pixbuf/gdk-pixbuf-marshal.h " +
          "--output $@",
    tools = ["@glib2//:gobject_glib_genmarshal_tool"],
)

cc_library(
    name = "gdk_pixbuf_headers",
    hdrs = [
               "gdk-pixbuf/gdk-pixbuf-marshal.h",
           ] + GDKPIXBUF_HEADERS_PRIVATE + GDKPIXBUF_HEADERS +
           GDKPIXBUF_FEATURES,
    include_prefix = ".",
    strip_include_prefix = "gdk-pixbuf",
)

cc_library(
    name = "gdk_pixbuf_sub_headers",
    hdrs = GDKPIXBUF_HEADERS + GDKPIXBUF_FEATURES,
    includes = ["."],
)

cc_library(
    name = "gdk-pixbuf",
    srcs = [
        "config.h",
        "gdk-pixbuf/gdk-pixbuf-marshal.c",
        "gdk-pixbuf/io-bmp.c",
    ] + GDKPIXBUF_SOURCES + GDKPIXDATA_SOURCES,
    copts = [
        "-Wno-uninitialized",
    ],
    local_defines = [
        "GDK_PIXBUF_BINARY_VERSION='\"2.42.9\"'",
        "GDK_PIXBUF_COMPILATION=1",
        "GDK_PIXBUF_LOCALEDIR='\"/nonexistent\"'",
        "GDK_PIXBUF_ENABLE_BACKEND",
        "GDK_PIXBUF_LIBDIR='\"/nonexistent\"'",
        "PIXBUF_LIBDIR='\"/nonexistent\"'",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":gdk_pixbuf_headers",
        ":gdk_pixbuf_sub_headers",
        ":gdkpixbuf_enums",
        ":pixops",
        "@glib2//:gio",
        "@glib2//:glib",
        "@glib2//:gobject",
    ],
)
