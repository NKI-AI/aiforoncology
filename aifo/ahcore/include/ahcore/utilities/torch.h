// Copyright 2024 Jonas Teuwen. All Rights Reserved.
// Copyright 2025 Jonas Teuwen & Joren Brunekreef. All Rights Reserved.
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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_

#include <torch/script.h>
#include <torch/torch.h>
#include <vips/vips8>

#include <stdexcept>
#include <string>
#include <unordered_map>
#include <variant>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"

namespace aifo::utilities {

extern const std::unordered_map<VipsBandFormat, torch::ScalarType>
    VIPS_TO_TORCH_FORMAT;

extern const std::unordered_map<torch::ScalarType, VipsBandFormat>
    TORCH_TO_VIPS_FORMAT;

// Helper for std::visit
template <class... Ts>
struct overloaded : Ts... {
  using Ts::operator()...;
};

template <class... Ts>
overloaded(Ts...) -> overloaded<Ts...>;

class DeviceManager {
 public:
  static torch::Device GetAvailableDevice(bool force_cpu = false);
  static std::string DeviceName(const torch::Device& device);
};

class TorchModel {
 public:
  // Custom exceptions for better error handling
  class ModelException : public std::runtime_error {
    using std::runtime_error::runtime_error;
  };

  struct ModelOutput {
    using TensorList = std::vector<torch::Tensor>;
    using TensorDict = std::unordered_map<std::string, torch::Tensor>;
    using Variant = std::variant<torch::Tensor, TensorList, TensorDict>;

    Variant data;

    bool IsSingleTensor() const;
    bool IsTensorList() const;
    bool IsTensorDict() const;

    const torch::Tensor& AsTensor() const;
    const TensorList& AsList() const;
    const TensorDict& AsDict() const;

    TensorList ToList() const;
    TensorDict ToDict() const;
    torch::Tensor ToTensor() const;
  };

  // Constructor with model path
  static TorchModel FromFile(const std::string& model_path,
                             torch::Device device = torch::kCPU);

  // Primary constructor with model data
  explicit TorchModel(const std::string& model_data,
                      torch::Device device = torch::kCPU);

  // Move constructor and assignment operator
  TorchModel(TorchModel&&) noexcept;
  TorchModel& operator=(TorchModel&&) noexcept;

  // Delete copy operations as torch::jit::Module is not copyable
  TorchModel(const TorchModel&) = delete;
  TorchModel& operator=(const TorchModel&) = delete;

  // Inference method with input validation
  c10::IValue Infer(const torch::Tensor& input) const;

  // Process output
  std::vector<torch::Tensor> ProcessVectorOutput(
      const c10::IValue& output) const;

  // Process dictionary output
  std::unordered_map<std::string, torch::Tensor> ProcessDictionaryOutput(
      const c10::IValue& output) const;

  ModelOutput ProcessOutput(const c10::IValue& output) const;

  // Getter for device
  torch::Device GetDevice() const noexcept;

 private:
  torch::Device device_;
  mutable torch::jit::script::Module model_;

  // Helper methods
  static std::string LoadModelFromFile(const std::string& path);
  void ValidateInput(const torch::Tensor& input) const;
  void ProcessTupleOutput(const c10::intrusive_ptr<c10::ivalue::Tuple>& tuple,
                          std::vector<torch::Tensor>& tensors) const;
  void ProcessListOutput(const c10::IValue& list_value,
                         std::vector<torch::Tensor>& tensors) const;
  void FlattenDictionary(
      const c10::impl::GenericDict& dict, const std::string& parent_key,
      std::unordered_map<std::string, torch::Tensor>& flat_dict) const;
  void ProcessListInDictionary(
      const c10::List<c10::IValue>& list, const std::string& key,
      std::unordered_map<std::string, torch::Tensor>& flat_dict) const;
};

/**
 * Convert a VIPS image to a libtorch tensor.
 * @param image Input VIPS image
 * @param device Target device for the tensor
 * @return StatusOr containing a libtorch tensor in [C, H, W] format
 */
absl::StatusOr<torch::Tensor> VImageToTensor(
    const vips::VImage& image, torch::Device device = torch::kCPU);

/**
 * Convert a libtorch tensor to a VIPS image.
 * @param tensor Input tensor in [C, H, W] format
 * @return StatusOr containing a VIPS image
 */
absl::StatusOr<vips::VImage> TensorToVImage(const torch::Tensor& tensor);

/**
 * Convert a batch of VIPS images to a torch tensor.
 * @param images Vector of VIPS images
 * @return StatusOr containing a tensor in [N, C, H, W] format
**/
absl::StatusOr<torch::Tensor> ConvertImageBatch(
    const std::vector<vips::VImage>& images,
    torch::Device device = torch::kCPU);

/**
 * Normalize a tensor using the provided mean and standard deviation.
 * @param input_tensor Input tensor
 * @param mean Mean values for each channel
 * @param std Standard deviation values for each channel
 * @return Normalized tensor
 */
torch::Tensor NormalizeTensor(const torch::Tensor& input_tensor,
                              const std::vector<float>& mean,
                              const std::vector<float>& std);

}  // namespace aifo::utilities

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_
