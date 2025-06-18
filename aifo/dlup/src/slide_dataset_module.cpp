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
#include <pybind11/detail/type_caster_base.h>
#include <pybind11/functional.h>
#include <pybind11/pybind11.h>
#include <pybind11/stl.h>
#include <pybind11/stl_bind.h>

#include <memory>
#include <tuple>

#include "aifocore/tiling/python/grid_wrapper.h"
#include "dlup/slide_dataset.h"
#include "dlup/slide_image.h"

namespace py = pybind11;

// TODO(jonasteuwen): A std::pair maps to a tuple!
// This is a good example of how to map a C++ pair to a Python tuple.
PYBIND11_MODULE(_slide_dataset, m) {
  py::class_<dlup::DatasetSample>(m, "DatasetSample")
      .def_readonly("tile", &dlup::DatasetSample::tile, "Tile image data")
      .def_readonly("coordinates", &dlup::DatasetSample::coordinates,
                    "Tile's top-left coordinates")
      .def_readonly("identifier", &dlup::DatasetSample::identifier,
                    "Image identifier");
  py::class_<dlup::SlideDataset, std::shared_ptr<dlup::SlideDataset>>(
      m, "SlideDataset")
      .def(
          py::init([](std::shared_ptr<dlup::SlideImage> slide,
                      const aifocore::tiling::python::GridWrapper& grid_wrapper,
                      double mpp, std::tuple<int, int> tile_size, bool crop) {
            // Check SlideImage validity
            if (!slide) {
              throw std::invalid_argument(
                  "Invalid SlideImage object provided.");
            }

            // Extract Grid<int> from GridWrapper
            auto grid = std::get_if<Grid<int>>(&grid_wrapper.GetGrid());
            if (!grid) {
              throw std::invalid_argument(
                  "SlideDataset constructor expects a Grid<int>, but received "
                  "a Grid<double> or invalid type.");
            }

            // Validate mpp
            if (mpp <= 0) {
              throw std::invalid_argument(
                  "Microns per pixel (mpp) must be positive.");
            }

            // Validate tile_size
            auto [width, height] = tile_size;
            if (width <= 0 || height <= 0) {
              throw std::invalid_argument(
                  "Tile size must be positive integers.");
            }

            // Construct SlideDataset
            return std::make_shared<dlup::SlideDataset>(
                slide, *grid, mpp, aifocore::Size<int, 2>{width, height}, crop);
          }),
          py::arg("slide"), py::arg("grid"), py::arg("mpp"),
          py::arg("tile_size"), py::arg("crop") = true,
          "Initialize a SlideDataset.")
      .def("begin", &dlup::SlideDataset::begin,
           "Returns an iterator to the beginning of the dataset.")
      .def("end", &dlup::SlideDataset::end,
           "Returns an iterator to the end of the dataset.")
      .def("__len__", &dlup::SlideDataset::Length,
           "Returns the total number of tiles in the dataset.")
      .def("__getitem__", &dlup::SlideDataset::GetTile, py::arg("index"),
           "Retrieve the tile at a specified index.")
      .def("get_grid", &dlup::SlideDataset::GetGrid,
           py::return_value_policy::reference,
           "Returns the grid associated with the dataset.")
      .def("get_mpp", &dlup::SlideDataset::GetMpp,
           "Returns the microns per pixel (MPP) scaling.")
      .def("get_tile_size", &dlup::SlideDataset::GetTileSize,
           py::return_value_policy::reference,
           "Returns the tile size dimensions.")
      .def("get_scaling", &dlup::SlideDataset::GetScaling,
           "Returns the scaling factor for the slide image.")
      .def("get_slide", &dlup::SlideDataset::GetSlide,
           py::return_value_policy::reference,
           "Returns the slide image associated with the dataset.")
      .def("get_crop", &dlup::SlideDataset::GetCrop,
           "Returns whether tiles are cropped to fit exact dimensions.")
      .def("get_image_bounds", &dlup::SlideDataset::GetImageBounds,
           py::return_value_policy::reference,
           "Returns the bounds of the slide image.")
      .def_property_readonly(
          "tiles",
          [](dlup::SlideDataset& self) {
            return py::make_iterator(self.begin(), self.end());
          },
          "Provides an iterable for the tiles in the dataset.");
}
