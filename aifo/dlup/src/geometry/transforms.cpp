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
#include "dlup/transforms.h"

#include <algorithm>
#include <array>
#include <cmath>
#include <limits>
#include <memory>
#include <stdexcept>
#include <string>
#include <tuple>
#include <utility>
#include <variant>
#include <vector>

#include <xtensor/containers/xadapt.hpp>
#include <xtensor/containers/xtensor.hpp>
#include <xtensor/core/xmath.hpp>
#include <xtensor/generators/xbuilder.hpp>

#include "absl/status/status.h"
#include "aifocore/status/status_macros.h"
#include "dlup/geometry/box.h"
#include "dlup/geometry/polygon.h"

namespace dlup {

/**
 * @brief Edge structure for the Active Edge Table (AET) used in polygon filling.
 */
struct Edge {
  /// Scanline where the edge is no longer active.
  int y_max;
  /// Current x-coordinate of the intersection.
  double x;
  /// Change in x per scanline (dx/dy).
  double inv_slope;
};

/**
 * @brief Insertion sort for the Active Edge Table (AET).
 * @param AET The Active Edge Table to be sorted by x-coordinate.
 */
static void InsertationSortAET(std::vector<Edge>& AET) {
  for (std::size_t i = 1; i < AET.size(); ++i) {
    Edge key = AET[i];
    int j = static_cast<int>(i) - 1;
    while (j >= 0 && AET[j].x > key.x) {
      AET[j + 1] = AET[j];
      j--;
    }
    AET[j + 1] = key;
  }
}

/**
 * @brief Fills the given mask inside the area defined by the polygon with the specified fill value.
 *
 * Uses a scanline algorithm with an Active Edge Table (AET) to efficiently fill the polygon.
 * 
 * @tparam Mask The xtensor type of the mask to be filled.
 * @param mask The mask to fill (an xtensor of any integral type).
 * @param polygon The polygon defining the area to be filled. The polygon must provide:
 *    - GetExterior() const, returning std::vector<std::pair<double, double>>
 *    - GetBoundingBox() const, returning a Box (from which we get the coordinates and size)
 * @param value The value to fill the polygon with.
 * @return absl::Status Status indicating success or failure.
 */
template <typename Mask>
absl::Status FillPoly(Mask& mask, const geometry::Polygon& polygon,
                      typename Mask::value_type value) {
  using T = typename Mask::value_type;
  int height = static_cast<int>(mask.shape()[0]);
  int width = static_cast<int>(mask.shape()[1]);

  // Validate mask dimensions to avoid overflow
  if (height <= 0 || width <= 0 ||
      static_cast<uint64_t>(height) * static_cast<uint64_t>(width) >
          static_cast<uint64_t>(std::numeric_limits<int>::max())) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Mask dimensions are invalid or too large");
  }

  // Retrieve the exterior vertices.
  auto vertices = polygon.GetExterior();
  if (vertices.empty())
    return absl::OkStatus();

  // Ensure the polygon is closed.
  if (vertices.front() != vertices.back())
    vertices.push_back(vertices.front());

  // Get the polygon's bounding box.
  auto bounding_box = polygon.GetBoundingBox();  // Box GetBoundingBox() const;
  auto bbox_start = bounding_box.GetCoordinates();  // std::array<double,2>
  auto bbox_size = bounding_box.GetSize();          // std::array<double,2>

  int bbox_x_min = std::max(0, static_cast<int>(std::round(bbox_start[0])));
  int bbox_y_min = std::max(0, static_cast<int>(std::round(bbox_start[1])));
  int bbox_x_max = std::min(
      width - 1, bbox_x_min + static_cast<int>(std::round(bbox_size[0])) - 1);
  int bbox_y_max = std::min(
      height - 1, bbox_y_min + static_cast<int>(std::round(bbox_size[1])) - 1);

  // Build the Edge Table (ET): vector of edges for each scanline
  std::vector<std::vector<Edge>> ET(height);  // Initialize with image height
  int n = static_cast<int>(vertices.size()) - 1;
  for (int i = 0; i < n; ++i) {
    int x1 = static_cast<int>(std::round(vertices[i].first));
    int y1 = static_cast<int>(std::round(vertices[i].second));
    int x2 = static_cast<int>(std::round(vertices[i + 1].first));
    int y2 = static_cast<int>(std::round(vertices[i + 1].second));

    // Skip horizontal edges.
    if (y1 == y2)
      continue;

    int y_min, y_max, x_at_ymin;
    double inv_slope;
    if (y1 < y2) {
      y_min = y1;
      y_max = y2;
      x_at_ymin = x1;
      inv_slope = static_cast<double>(x2 - x1) / (y2 - y1);
    } else {
      y_min = y2;
      y_max = y1;
      x_at_ymin = x2;
      inv_slope = static_cast<double>(x1 - x2) / (y1 - y2);
    }
    // Skip edges completely outside the mask.
    if (y_max < 0 || y_min >= height)
      continue;
    // If the edge starts above the image, adjust to y=0.
    if (y_min < 0) {
      x_at_ymin += inv_slope * (0 - y_min);
      y_min = 0;
    }
    Edge edge{.y_max = y_max,
              .x = static_cast<double>(x_at_ymin),
              .inv_slope = inv_slope};
    ET[y_min].push_back(edge);
  }

  std::vector<Edge> active_edges;
  // Process only the scanlines within the bounding box.
  for (int y = bbox_y_min; y <= bbox_y_max; ++y) {
    // Remove edges that are no longer active.
    std::vector<Edge> new_active;
    for (auto& edge : active_edges) {
      if (y < edge.y_max)
        new_active.push_back(edge);
    }
    active_edges = std::move(new_active);

    // Add new edges that start at this scanline.
    if (!ET[y].empty()) {
      active_edges.insert(active_edges.end(), ET[y].begin(), ET[y].end());
    }
    if (active_edges.empty())
      continue;

    InsertationSortAET(active_edges);

    // Process pairs of intersections and fill the pixels between them.
    for (std::size_t i = 0; i < active_edges.size(); i += 2) {
      if (i + 1 >= active_edges.size())
        break;  // Safety check.

      int x_start = static_cast<int>(std::ceil(active_edges[i].x));
      int x_end = static_cast<int>(std::floor(active_edges[i + 1].x));

      // Clip the x coordinates to the mask and bounding box.
      if (x_end < 0 || x_start >= width)
        continue;
      x_start = std::max(x_start, bbox_x_min);
      x_end = std::min(x_end, bbox_x_max);

      if (x_start <= x_end) {
        for (int x = x_start; x <= x_end; ++x) {
          mask(y, x) = value;
        }
      }
    }
    // Update the x coordinate for each active edge.
    for (auto& edge : active_edges)
      edge.x += edge.inv_slope;
  }
  return absl::OkStatus();
}

/**
 * @brief Generates a mask from a vector of polygon annotations.
 *
 * Creates a 2D mask where each polygon is filled with its corresponding index value.
 * Supports polygons with interior holes.
 *
 * @param annotations Vector of polygon annotations to be rendered into the mask.
 * @param mask_size Tuple containing the width and height of the mask.
 * @param default_value Default value for areas of the mask not covered by any polygon.
 * @return std::shared_ptr<xt::xtensor<int, 2>> Generated mask as a shared pointer.
 * @throws std::runtime_error If mask dimensions are invalid, too large, or if annotation 
 *         processing fails.
 */
std::shared_ptr<xt::xtensor<int, 2>> GenerateMaskFromAnnotations(
    const std::vector<std::shared_ptr<geometry::Polygon>>& annotations,
    const std::tuple<int, int>& mask_size, int default_value) {
  int width = std::get<0>(mask_size);
  int height = std::get<1>(mask_size);

  // Check if mask dimensions are valid and not too large
  if (width <= 0 || height <= 0 ||
      static_cast<uint64_t>(width) * static_cast<uint64_t>(height) >
          static_cast<uint64_t>(std::numeric_limits<std::size_t>::max() /
                                sizeof(int))) {
    throw std::runtime_error(
        "Invalid mask dimensions or mask too large for available memory");
  }

  // Create the mask as a 2D xtensor initialized with default_value.
  std::array<std::size_t, 2> shape = {static_cast<std::size_t>(height),
                                      static_cast<std::size_t>(width)};
  auto mask = std::make_shared<xt::xtensor<int, 2>>(shape, default_value);

  // Process each annotation.
  for (const auto& annotation : annotations) {
    // Retrieve the fill value from the "index" field.
    auto index_value_field = annotation->GetField("index");
    const int* index_value_ptr = std::get_if<int>(&(*index_value_field));
    if (!index_value_ptr) {
      throw std::runtime_error("Index field is not an integer.");
    }
    int index_value = *index_value_ptr;

    // Process the exterior using the original annotation approach
    if (!annotation->GetInteriors().empty()) {
      // Create a holes mask (of type uint8_t)
      // with the same dimensions, initialized to 0.
      xt::xtensor<uint8_t, 2> holes_mask = xt::zeros<uint8_t>(
          {static_cast<std::size_t>(height), static_cast<std::size_t>(width)});
      // For each interior ring,
      // construct a temporary polygon and fill it with 1.
      for (const auto& interior : annotation->GetInteriors()) {
        geometry::Polygon hole;
        hole.SetExterior(interior);
        auto status = FillPoly(holes_mask, hole, static_cast<uint8_t>(1));
        if (!status.ok()) {
          throw std::runtime_error(std::string(status.message()));
        }
      }
      // Make a copy of the original values
      xt::xtensor<int, 2> original_values = *mask;
      // Fill the exterior polygon with the annotation's fill value.
      auto status = FillPoly(*mask, *annotation, index_value);
      if (!status.ok()) {
        throw std::runtime_error(std::string(status.message()));
      }
      // Restore original mask values where holes were filled.
      *mask = xt::where(xt::equal(holes_mask, static_cast<uint8_t>(1)),
                        original_values, *mask);
    } else {
      // No interiors—fill the polygon directly.
      auto status = FillPoly(*mask, *annotation, index_value);
      if (!status.ok()) {
        throw std::runtime_error(std::string(status.message()));
      }
    }
  }
  return mask;
}

}  // namespace dlup
