load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

licenses(["notice"])

# Expand version.h from a template
expand_template(
    name = "version_h",
    out = "include/dicom/version.h",
    substitutions = {
        "@DCM_VERSION@": "1.1.0",
        "@DCM_SUFFIXED_VERSION@": "1.1.0",
        "@DCM_VERSION_MAJOR@": "1",
        "@DCM_VERSION_MINOR@": "1",
        "@DCM_VERSION_MICRO@": "0",
        "@DCM_ABI_VERSION_MAJOR@": "1",
        "@DCM_ABI_VERSION_MINOR@": "0",
        "@DCM_ABI_VERSION_PATCH@": "0",
        "@DCM_STATIC@": "0",
    },
    template = "include/dicom/version.h.in",
)

cc_library(
    name = "libdicom_version_h",
    hdrs = ["include/dicom/version.h"],
    includes = [
        "include/dicom",
    ],
)

cc_binary(
    name = "dicom_dict_build",
    srcs = [
        "config.h",
        "include/dicom/dicom.h",
        "src/dicom-dict-build.c",
        "src/dicom-dict-tables.c",
        "src/dicom-dict-tables.h",
        "src/pdicom.h",
    ],
    includes = ["include"],
    deps = [
        ":libdicom_version_h",  # For version.h
        "@uthash",
    ],
)

# Use it to generate both the .c and .h files
genrule(
    name = "generate_dict_lookup",
    outs = [
        "src/dicom-dict-lookup.c",
        "src/dicom-dict-lookup.h",
    ],
    cmd = "$(location :dicom_dict_build) $(OUTS)",
    tools = [":dicom_dict_build"],
)

# Main dicom library
cc_library(
    name = "libdicom",
    srcs = [
        "config.h",
        "src/dicom.c",
        "src/dicom-data.c",
        "src/dicom-dict.c",
        "src/dicom-dict-tables.c",
        "src/dicom-dict-tables.h",
        "src/dicom-file.c",
        "src/dicom-io.c",
        "src/dicom-parse.c",
        "src/getopt.c",
        "src/pdicom.h",
        ":generate_dict_lookup",
    ],
    hdrs = [
        "include/dicom/dicom.h",
        "include/dicom/version.h",
        "src/dicom-dict-lookup.h",  # Generated header
    ],
    includes = [
        "include",
        "src",
    ],
    visibility = ["//visibility:public"],
    deps = [
        ":libdicom_version_h",
        "@uthash",
    ],
)

# Generate config.h
genrule(
    name = "configure",
    outs = ["config.h"],
    cmd = "\n".join([
        "cat <<'EOF' >$@",
        "#pragma once",
        "#define HAS_CONSTRUCTOR",
        "#define HAVE_UNISTD_H 1",
        "EOF",
    ]),
)
