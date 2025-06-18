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
#ifndef AIFO_DLUP_INCLUDE_DLUP_GEOMETRY_LAZY_ARRAY_H_
#define AIFO_DLUP_INCLUDE_DLUP_GEOMETRY_LAZY_ARRAY_H_

#include <functional>
#include <utility>
#include <vector>

#include <xtensor/containers/xarray.hpp>
#include <xtensor/containers/xtensor.hpp>
#include <xtensor/io/xio.hpp>
#include <xtensor/views/xview.hpp>

template <typename T>
class LazyArray {
 public:
  using ComputeFunction =
      std::function<xt::xtensor<T, 2>()>;  // Function returning xtensor

  LazyArray(ComputeFunction compute_func, std::vector<std::size_t> shape)
      : compute_func_(std::move(compute_func)),
        computed_(false),
        shape_(std::move(shape)) {}

  // Return the computed xtensor
  xt::xtensor<T, 2> xtensor() const {
    if (!computed_) {
      data_ = compute_func_();  // Compute lazily
      computed_ = true;
    }
    return data_;  // Return the stored xtensor
  }

  // Return the shape of the xtensor
  std::vector<std::size_t> shape() const { return shape_; }

  LazyArray<T> reshape(const std::vector<std::size_t>& new_shape) const {
    return LazyArray<T>(
        [this, new_shape]() {
          auto arr = this->xtensor();
          return xt::reshape_view(arr, new_shape);
        },
        new_shape);
  }

  LazyArray<T> operator+(const LazyArray<T>& other) const {
    return LazyArray<T>(
        [this, &other]() { return this->xtensor() + other.xtensor(); }, shape_);
  }

  LazyArray<T> operator-(const LazyArray<T>& other) const {
    return LazyArray<T>(
        [this, &other]() { return this->xtensor() - other.xtensor(); }, shape_);
  }

 private:
  ComputeFunction compute_func_;    // Function to compute the xtensor
  mutable xt::xtensor<T, 2> data_;  // Store the computed xtensor
  mutable bool computed_;           // Flag to indicate if computed
  std::vector<std::size_t> shape_;  // Shape of the xtensor
};

#endif  // AIFO_DLUP_INCLUDE_DLUP_GEOMETRY_LAZY_ARRAY_H_
