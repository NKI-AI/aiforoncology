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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_TRANSFORMS_PRE_WRITER_TRANSFORMS_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_TRANSFORMS_PRE_WRITER_TRANSFORMS_H_

#include <torch/script.h>

#include <memory>
#include <string>

#include "ahcore/pathology/transforms/base_transform.h"

namespace aifo::pathology::transforms {

class PreWriterTransform : public BaseTransform {
 public:
  virtual ~PreWriterTransform() = default;

  /**
   * @brief Process the model output tensor before writing
   * @param input The input tensor from model output
   * @return The processed tensor ready for writing
   */
  virtual torch::Tensor Forward(const torch::Tensor& input) const = 0;
};

class CropPreWriterTransform : public PreWriterTransform {
 public:
  torch::Tensor Forward(const torch::Tensor& input) const override {
    // For crop, we want argmax'd uint8 output
    return input.softmax(1).argmax(1).to(torch::kUInt8).unsqueeze(1);
  }

  std::string GetName() const override { return "crop"; }
};

class AveragePreWriterTransform : public PreWriterTransform {
 public:
  torch::Tensor Forward(const torch::Tensor& input) const override {
    // For average, keep probabilities as float
    return input.softmax(1);
  }

  std::string GetName() const override { return "average"; }
};

class MaximumPreWriterTransform : public PreWriterTransform {
 public:
  torch::Tensor Forward(const torch::Tensor& input) const override {
    // For maximum, keep probabilities as float
    // return input.softmax(1);
    throw std::runtime_error("Maximum pre-writer transform not implemented");
  }

  std::string GetName() const override { return "maximum"; }
};

// Factory function to create pre-writer transforms
inline std::shared_ptr<PreWriterTransform> CreatePreWriterTransform(
    const std::string& method) {
  if (method == "crop") {
    return std::make_shared<CropPreWriterTransform>();
  } else if (method == "average") {
    return std::make_shared<AveragePreWriterTransform>();
  } else if (method == "maximum") {
    return std::make_shared<MaximumPreWriterTransform>();
  }
  throw std::runtime_error("Unsupported pre-writer transform method: " +
                           method);
}

}  // namespace aifo::pathology::transforms

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_PATHOLOGY_TRANSFORMS_PRE_WRITER_TRANSFORMS_H_
