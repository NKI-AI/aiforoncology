// Copyright 2024 Jonas Teuwen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
#include <vips/vips8>

#include <pybind11/cast.h>
#include <pybind11/functional.h>
#include <pybind11/pybind11.h>
#include <pybind11/stl.h>

#include <memory>

#include "aifocore/concepts/numeric.h"
#include "dlup/slide_image.h"

namespace py = pybind11;

PYBIND11_MODULE(_slide_image, m) {
  py::enum_<dlup::Resampling>(m, "Resampling")
      .value("NEAREST", dlup::Resampling::kNearest)
      .value("LANCZOS", dlup::Resampling::kLanczos);

  // Define SlideGeometry struct binding
  py::class_<dlup::SlideGeometry>(m, "SlideGeometry")
      .def(py::init<>())
      .def_readwrite("size", &dlup::SlideGeometry::size)
      .def_readwrite("offset", &dlup::SlideGeometry::offset)
      .def_readwrite("bounds", &dlup::SlideGeometry::bounds)
      .def("scaled", &dlup::SlideGeometry::Scaled, py::arg("scaling"));

  py::class_<dlup::SlideImage, std::shared_ptr<dlup::SlideImage>>(m,
                                                                  "SlideImage")
      .def(py::init<std::shared_ptr<dlup::backends::AbstractSlideBackend>,
                    const std::optional<std::string_view>&, dlup::Resampling,
                    const std::optional<aifocore::Size<double, 2>>&>(),
           py::arg("wsi"), py::arg("identifier") = std::nullopt,
           py::arg("interpolator") = dlup::Resampling::kLanczos,
           py::arg("overwrite_mpp") = std::nullopt)
      .def("close", &dlup::SlideImage::Close)
      .def("read_region", &dlup::SlideImage::ReadRegion, py::arg("location"),
           py::arg("scaling"), py::arg("size"),
           "Read a region of the slide and return a pyvips.Image.")
      .def("get_scaled_size", &dlup::SlideImage::GetScaledSize,
           py::arg("scaling"), py::arg("limit_bounds") = false)
      .def("get_scaling", &dlup::SlideImage::GetScaling)
      .def("get_closest_native_level", &dlup::SlideImage::GetClosestNativeLevel)
      .def("get_closest_native_mpp", &dlup::SlideImage::GetClosestNativeMpp)
      .def_property_readonly(
          "mpp", py::overload_cast<>(&dlup::SlideImage::GetMpp, py::const_))
      .def_property_readonly("slide_bounds", &dlup::SlideImage::GetSlideBounds)
      .def_property_readonly("identifier", &dlup::SlideImage::GetIdentifier)
      .def_property_readonly("properties", &dlup::SlideImage::GetProperties)
      .def_property_readonly("vendor", &dlup::SlideImage::GetVendor)
      .def_property_readonly("interpolator", &dlup::SlideImage::GetInterpolator)
      .def_property_readonly("dimensions", &dlup::SlideImage::GetDimensions)
      .def_property_readonly("size", &dlup::SlideImage::GetSize)
      .def_property_readonly("geometry", &dlup::SlideImage::GetGeometry)
      .def_property_readonly("aspect_ratio", &dlup::SlideImage::GetAspectRatio)
      .def("get_scaled_slide_bounds", &dlup::SlideImage::GetScaledSlideBounds,
           py::arg("scaling"))
      .def("get_scaled", &dlup::SlideImage::GetScaled, py::arg("scaling"),
           py::arg("limit_bounds") = false, "Get the scaled size of the slide.")
      .def("get_scaled_geometry", &dlup::SlideImage::GetScaledGeometry,
           py::arg("scaling"), "Get the scaled geometry of the slide.")
      .def("get_thumbnail", &dlup::SlideImage::GetThumbnail,
           py::arg("size") = aifocore::Size<int, 2>{512, 512},
           "Get a thumbnail of the slide as a pyvips.Image.");
}
