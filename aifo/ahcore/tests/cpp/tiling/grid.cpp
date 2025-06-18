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
#include <aifocore/tiling/grid.h>

#include <gtest/gtest.h>
#include <cmath>
#include <map>
#include <stdexcept>
#include <vector>

using namespace aifocore::tiling;

TEST(TilingTest, TestAllZero) {
  TilingMode mode = TilingMode::kSkip;

  std::vector<double> size = {0};
  std::vector<double> tile_size = {0};
  std::vector<double> tile_overlap = {0};

  EXPECT_THROW(
      {
        auto coordinates =
            TilesGridCoordinates<double>(size, tile_size, tile_overlap, mode);
      },
      std::invalid_argument);
}

TEST(TilingTest, TestTileBiggerThanSize) {
  std::vector<TilingMode> modes = {TilingMode::kSkip, TilingMode::kOverflow};
  std::vector<double> tile_overlaps = {0, 1, 2};

  double size_value = 2.0;
  double tile_size_value = 10.0;

  std::map<TilingMode, size_t> expected_lengths = {{TilingMode::kSkip, 0},
                                                   {TilingMode::kOverflow, 1}};

  for (auto mode : modes) {
    for (auto tile_overlap_value : tile_overlaps) {
      std::vector<double> size = {size_value};
      std::vector<double> tile_size = {tile_size_value};
      std::vector<double> tile_overlap = {tile_overlap_value};

      auto coordinates =
          TilesGridCoordinates<double>(size, tile_size, tile_overlap, mode);
      auto& basis = coordinates[0];

      // Check that all elements are >= 0
      bool all_non_negative = std::all_of(basis.begin(), basis.end(),
                                          [](double x) { return x >= 0; });
      EXPECT_TRUE(all_non_negative);

      size_t expected_length = expected_lengths[mode];
      EXPECT_EQ(basis.size(), expected_length);
    }
  }
}

TEST(TilingTest, TestSpannedBasis) {
  std::vector<std::tuple<double, double, double>> parameters = {
      {10, 3, 0}, {3, 1, 2}, {17, 3.2, 2}, {53.2, 12.2, 15}, {1, 2, 3}};

  std::vector<TilingMode> modes = {TilingMode::kSkip, TilingMode::kOverflow};

  for (const auto& param : parameters) {
    double size_value, tile_size_value, tile_overlap_value;
    std::tie(size_value, tile_size_value, tile_overlap_value) = param;

    for (auto mode : modes) {
      std::vector<double> size = {size_value};
      std::vector<double> tile_size = {tile_size_value};
      std::vector<double> tile_overlap = {tile_overlap_value};

      auto coordinates =
          TilesGridCoordinates<double>(size, tile_size, tile_overlap, mode);
      auto& basis = coordinates[0];

      // Check that basis is sorted
      bool is_sorted = std::is_sorted(basis.begin(), basis.end());
      EXPECT_TRUE(is_sorted);

      if (basis.empty()) {
        continue;
      }

      // First coordinate is always zero
      EXPECT_DOUBLE_EQ(basis[0], 0.0);

      // Compute stride as differences between basis elements
      std::vector<double> stride;
      for (size_t i = 1; i < basis.size(); ++i) {
        stride.push_back(basis[i] - basis[i - 1]);
      }

      double tiled_size = basis.back() + tile_size_value;

      // Check that grid is uniform
      if (!stride.empty()) {
        double first_stride = stride[0];
        bool strides_equal =
            std::all_of(stride.begin(), stride.end(), [first_stride](double s) {
              return std::abs(s - first_stride) < 1e-6;
            });
        EXPECT_TRUE(strides_equal);
      }

      if (std::abs(tiled_size - size_value) < 1e-6) {
        continue;
      }

      if (mode == TilingMode::kSkip) {
        EXPECT_LT(tiled_size, size_value);
      }

      if (mode == TilingMode::kOverflow) {
        EXPECT_GT(tiled_size, size_value);
      }
    }
  }
}

TEST(TilingTest, TestTilingWithOverlap) {
  std::vector<std::tuple<double, double, double>> parameters = {
      {10, 4, 1}, {15, 5, 2}, {20, 7, 3}};

  std::vector<TilingMode> modes = {TilingMode::kSkip, TilingMode::kOverflow};

  for (const auto& param : parameters) {
    double size_value, tile_size_value, tile_overlap_value;
    std::tie(size_value, tile_size_value, tile_overlap_value) = param;

    for (auto mode : modes) {
      std::vector<double> size = {size_value};
      std::vector<double> tile_size = {tile_size_value};
      std::vector<double> tile_overlap = {tile_overlap_value};

      auto coordinates =
          TilesGridCoordinates<double>(size, tile_size, tile_overlap, mode);
      auto& basis = coordinates[0];

      // Check that basis is sorted
      bool is_sorted = std::is_sorted(basis.begin(), basis.end());
      EXPECT_TRUE(is_sorted);

      if (basis.empty()) {
        continue;
      }

      // Compute stride as tile_size - tile_overlap
      double stride = tile_size_value - tile_overlap_value;

      // Verify the positions of the tiles
      for (size_t i = 1; i < basis.size(); ++i) {
        EXPECT_NEAR(basis[i] - basis[i - 1], stride, 1e-6);
      }

      // Check that the last tile's end does not
      // exceed size unless in overflow mode
      double last_tile_end = basis.back() + tile_size_value;

      if (mode == TilingMode::kSkip) {
        EXPECT_LE(last_tile_end, size_value);
      } else if (mode == TilingMode::kOverflow) {
        EXPECT_GE(last_tile_end, size_value);
      }
    }
  }
}

TEST(TilingTest, TestSpannedBasisMultipleDims) {
  // 1D case
  std::vector<double> size1D = {10};
  std::vector<double> tile_size1D = {3};
  std::vector<double> tile_overlap1D = {1.2};
  auto coordinates1D = TilesGridCoordinates<double>(
      size1D, tile_size1D, tile_overlap1D, TilingMode::kSkip);
  auto& basis = coordinates1D[0];

  // 2D case
  std::vector<double> size2D = {10, 5};
  std::vector<double> tile_size2D = {3, 2};
  std::vector<double> tile_overlap2D = {1.2, 1};
  auto coordinates2D = TilesGridCoordinates<double>(
      size2D, tile_size2D, tile_overlap2D, TilingMode::kSkip);
  auto& dbasis = coordinates2D[0];

  // Check that basis == dbasis
  ASSERT_EQ(basis.size(), dbasis.size());
  for (size_t i = 0; i < basis.size(); ++i) {
    EXPECT_DOUBLE_EQ(basis[i], dbasis[i]);
  }
}

TEST(TilingTest, TestGrid) {
  std::vector<GridOrder> orders = {GridOrder::kF, GridOrder::kC};

  for (auto order : orders) {
    // Initialize grid
    std::vector<std::vector<double>> coordinates = {{0, 1}, {2, 3, 4}};

    Grid<double> grid(coordinates, order);

    // Test grid size
    auto size = grid.Size();
    ASSERT_EQ(size.size(), 2);
    EXPECT_EQ(size[0], 2);
    EXPECT_EQ(size[1], 3);

    // Test grid length
    EXPECT_EQ(grid.Length(), 6);

    // First row, first column
    auto point0 = grid[0];
    ASSERT_EQ(point0.size(), 2);
    EXPECT_DOUBLE_EQ(point0[0], 0);
    EXPECT_DOUBLE_EQ(point0[1], 2);

    if (order == GridOrder::kF) {
      // First row, second column
      auto point1 = grid[1];
      EXPECT_DOUBLE_EQ(point1[0], 0);
      EXPECT_DOUBLE_EQ(point1[1], 3);

      // Get grid[0:2]
      std::vector<std::vector<double>> expected_points = {{0, 2}, {0, 3}};
      for (size_t i = 0; i < 2; ++i) {
        auto point = grid[i];
        EXPECT_EQ(point.size(), 2);
        EXPECT_EQ(point, expected_points[i]);
      }
    } else {
      // In C order we need to look at the third element
      auto point2 = grid[2];
      EXPECT_DOUBLE_EQ(point2[0], 0);
      EXPECT_DOUBLE_EQ(point2[1], 3);

      // Get grid[0:3]
      std::vector<std::vector<double>> expected_points = {
          {0, 2}, {1, 2}, {0, 3}};
      for (size_t i = 0; i < 3; ++i) {
        auto point = grid[i];
        EXPECT_EQ(point.size(), 2);
        EXPECT_EQ(point, expected_points[i]);
      }
    }

    // Test grid.order()
    EXPECT_EQ(grid.Order(), order);
  }
}

TEST(TilingTest, TestExceptions) {
  std::vector<std::pair<double, double>> tile_sizes = {{1, 2}, {-1, 0}};

  for (const auto& tile_size_pair : tile_sizes) {
    std::vector<double> tile_size = {tile_size_pair.first,
                                     tile_size_pair.second};

    if (tile_size[0] < 0 || tile_size[1] < 0) {
      EXPECT_THROW(
          {
            auto grid =
                Grid<double>::FromTiling({0, 0}, {2, 1}, tile_size, {3, 2},
                                         TilingMode::kSkip, GridOrder::kC);
          },
          std::invalid_argument);
    } else {
      EXPECT_THROW(
          {
            auto grid =
                Grid<double>::FromTiling({0, 0}, {2}, tile_size, {3, 2},
                                         TilingMode::kSkip, GridOrder::kC);
          },
          std::invalid_argument);
    }
  }
}

int main(int argc, char** argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
