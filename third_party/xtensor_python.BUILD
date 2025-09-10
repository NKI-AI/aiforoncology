cc_library(
    name = "numpy_headers",
    hdrs = [
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/__multiarray_api.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/__ufunc_api.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/_neighborhood_iterator_imp.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/_numpyconfig.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/_public_dtype_api_table.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/arrayobject.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/arrayscalars.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/dtype_api.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/ndarrayobject.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/ndarraytypes.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_1_7_deprecated_api.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_2_compat.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_common.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_cpu.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_endian.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/npy_math.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/numpyconfig.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/ufuncobject.h",
        "@@rules_python~~pip~pip_311_numpy//:site-packages/numpy/_core/include/numpy/utils.h",
    ],
    strip_include_prefix = "/site-packages/numpy/_core/include",
    visibility = ["//visibility:public"],
)

cc_library(
    name = "xtensor_python",
    hdrs = glob(["include/xtensor-python/*.hpp"]),
    includes = ["include"],
    visibility = ["//visibility:public"],
    deps = [
        "@xtensor",
        "@pybind11",
        ":numpy_headers",
    ],
)
