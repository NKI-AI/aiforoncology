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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_

#include <torch/script.h>
#include <torch/torch.h>
#include <vips/vips8>

#include <stdexcept>
#include <string>
#include <unordered_map>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "aifocore/status/status_macros.h"
#include "aifocore/utilities/fmt.h"

namespace aifo::utilities {

static const std::unordered_map<VipsBandFormat, torch::ScalarType>
    VIPS_TO_TORCH_FORMAT = {{VIPS_FORMAT_UCHAR, torch::kUInt8},
                            {VIPS_FORMAT_FLOAT, torch::kFloat32},
                            {VIPS_FORMAT_DOUBLE, torch::kFloat64},
                            {VIPS_FORMAT_INT, torch::kInt32},
                            {VIPS_FORMAT_SHORT, torch::kInt16}};

static const std::unordered_map<torch::ScalarType, VipsBandFormat>
    TORCH_TO_VIPS_FORMAT = {{torch::kUInt8, VIPS_FORMAT_UCHAR},
                            {torch::kFloat32, VIPS_FORMAT_FLOAT},
                            {torch::kFloat64, VIPS_FORMAT_DOUBLE},
                            {torch::kInt32, VIPS_FORMAT_INT},
                            {torch::kInt16, VIPS_FORMAT_SHORT}};

// Helper for std::visit
template <class... Ts>
struct overloaded : Ts... {
  using Ts::operator()...;
};

template <class... Ts>
overloaded(Ts...) -> overloaded<Ts...>;

class DeviceManager {
 public:
  static torch::Device GetAvailableDevice(bool force_cpu = false) {
    if (force_cpu) {
      return torch::kCPU;
    }

    if (torch::cuda::is_available()) {
      return torch::kCUDA;
    }
    if (torch::mps::is_available()) {
      return torch::kMPS;
    }
    return torch::kCPU;
  }

  static std::string DeviceName(const torch::Device& device) {
    if (device.type() == torch::kCUDA) {
      return "CUDA";
    }
    if (device.type() == torch::kMPS) {
      return "MPS";
    }
    return "CPU";
  }
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

    /**
     * @brief Checks if the output contains a single tensor.
     *
     * @return true if the output is a single tensor, false otherwise.
     */
    bool IsSingleTensor() const {
      return std::holds_alternative<torch::Tensor>(data);
    }

    /**
     * @brief Checks if the output is a list of tensors.
     *
     * @return true if the output is a list of tensors, false otherwise.
     */
    bool IsTensorList() const {
      return std::holds_alternative<TensorList>(data);
    }

    /**
     * @brief Checks if the output is a dictionary of tensors.
     *
     * @return true if the output is a dictionary of tensors, false otherwise.
     */
    bool IsTensorDict() const {
      return std::holds_alternative<TensorDict>(data);
    }

    /**
     * @brief Access the output as a single tensor.
     *
     * @return A constant reference to the tensor.
     * @throws std::bad_variant_access if the output is not a single tensor.
     */
    const torch::Tensor& AsTensor() const {
      return std::get<torch::Tensor>(data);
    }

    /**
     * @brief Access the output as a list of tensors.
     *
     * @return A constant reference to the list of tensors.
     * @throws std::bad_variant_access if the output is not a list of tensors.
     */
    const TensorList& AsList() const { return std::get<TensorList>(data); }

    /**
     * @brief Access the output as a dictionary of tensors.
     *
     * @return A constant reference to the dictionary of tensors.
     * @throws std::bad_variant_access if the output is not a dictionary of tensors.
     */
    const TensorDict& AsDict() const { return std::get<TensorDict>(data); }

    /**
     * @brief Converts the output to a list of tensors.
     *
     * If the output is already a list of tensors, it returns the list directly.
     * If the output is a single tensor, it wraps the tensor in a list.
     * If the output is a dictionary, it extracts the tensor values into a list.
     *
     * @return A list of tensors.
     */
    TensorList ToList() const {
      return std::visit(
          overloaded{[](const torch::Tensor& t) -> TensorList { return {t}; },
                     [](const TensorList& l) -> TensorList { return l; },
                     [](const TensorDict& d) -> TensorList {
                       TensorList result;
                       for (const auto& [_, tensor] : d) {
                         result.push_back(tensor);
                       }
                       return result;
                     }},
          data);
    }

    /**
     * @brief Converts the output to a dictionary of tensors.
     *
     * If the output is already a dictionary of tensors, it returns the dictionary directly.
     * If the output is a single tensor, it wraps the tensor in a dictionary with the key "output".
     * If the output is a list, it converts the list to a dictionary with keys "output_0", "output_1", etc.
     *
     * @return A dictionary of tensors.
     */
    TensorDict ToDict() const {
      return std::visit(overloaded{[](const torch::Tensor& t) -> TensorDict {
                                     return {{"output", t}};
                                   },
                                   [](const TensorList& l) -> TensorDict {
                                     TensorDict result;
                                     for (size_t i = 0; i < l.size(); ++i) {
                                       result["output_" + std::to_string(i)] =
                                           l[i];
                                     }
                                     return result;
                                   },
                                   [](const TensorDict& d) -> TensorDict {
                                     return d;
                                   }},
                        data);
    }

    /**
      * @brief Converts the output to a single tensor.
      *
      * If the output is a single tensor, it returns the tensor directly.
      * If the output is a list with a single tensor, it extracts and returns that tensor.
      * If the output is a dictionary with a single key-value pair, it returns the tensor value.
      *
      * @return A single tensor.
      * @throws ModelException if the list or dictionary contains multiple tensors.
      */
    torch::Tensor ToTensor() const {
      return std::visit(
          overloaded{[](const torch::Tensor& t) -> torch::Tensor { return t; },
                     [](const TensorList& l) -> torch::Tensor {
                       if (l.size() != 1) {
                         throw ModelException(
                             "Cannot convert a list with multiple tensors to a "
                             "single tensor");
                       }
                       return l[0];
                     },
                     [](const TensorDict& d) -> torch::Tensor {
                       if (d.size() != 1) {
                         throw ModelException(
                             "Cannot convert a dictionary with multiple "
                             "entries to a single tensor");
                       }
                       return d.begin()->second;
                     }},
          data);
    }
  };

  // Constructor with model path
  static TorchModel FromFile(const std::string& model_path,
                             torch::Device device = torch::kCPU) {
    return TorchModel(LoadModelFromFile(model_path), device);
  }

  // Primary constructor with model data
  explicit TorchModel(const std::string& model_data,
                      torch::Device device = torch::kCPU) try
      : device_(device) {
    std::istringstream model_stream(model_data);
    model_ = torch::jit::load(model_stream);
    model_.to(device_);
    model_.eval();

  } catch (const torch::Error& e) {

    throw ModelException("Failed to initialize model: " +
                         std::string(e.what()));
  }

  // Move constructor and assignment operator
  TorchModel(TorchModel&&) noexcept = default;
  TorchModel& operator=(TorchModel&&) noexcept = default;

  // Delete copy operations as torch::jit::Module is not copyable
  TorchModel(const TorchModel&) = delete;
  TorchModel& operator=(const TorchModel&) = delete;

  // Inference method with input validation
  c10::IValue Infer(const torch::Tensor& input) const {
    ValidateInput(input);
    torch::NoGradGuard no_grad;
    try {
      return model_.forward({input.to(device_)});
    } catch (const torch::Error& e) {
      throw ModelException("Inference failed: " + std::string(e.what()));
    }
  }

  // Process output
  std::vector<torch::Tensor> ProcessVectorOutput(
      const c10::IValue& output) const {
    try {
      std::vector<torch::Tensor> tensors;
      if (output.isTensor()) {
        tensors.push_back(output.toTensor());
      } else if (output.isTuple()) {
        ProcessTupleOutput(output.toTuple(), tensors);
      } else if (output.isList()) {
        ProcessListOutput(output, tensors);
      } else {
        throw ModelException("Unsupported model output type");
      }
      return tensors;
    } catch (const c10::Error& e) {
      throw ModelException("Failed to process output: " +
                           std::string(e.what()));
    }
  }

  // Process dictionary output
  std::unordered_map<std::string, torch::Tensor> ProcessDictionaryOutput(
      const c10::IValue& output) const {
    if (!output.isGenericDict()) {
      throw ModelException("Output is not a dictionary");
    }

    std::unordered_map<std::string, torch::Tensor> flat_dict;
    try {
      FlattenDictionary(output.toGenericDict(), "", flat_dict);
    } catch (const c10::Error& e) {
      throw ModelException("Failed to process dictionary: " +
                           std::string(e.what()));
    }
    return flat_dict;
  }

  ModelOutput ProcessOutput(const c10::IValue& output) const {
    if (output.isTensor()) {
      return ModelOutput{output.toTensor()};
    } else if (output.isList()) {
      std::vector<torch::Tensor> tensors;
      ProcessListOutput(output, tensors);
      return ModelOutput{tensors};
    } else if (output.isTuple()) {
      std::vector<torch::Tensor> tensors;
      ProcessTupleOutput(output.toTuple(), tensors);
      return ModelOutput{tensors};
    } else if (output.isGenericDict()) {
      return ModelOutput{ProcessDictionaryOutput(output)};
    }
    throw ModelException("Unsupported model output type");
  }

  // Getter for device
  torch::Device GetDevice() const noexcept { return device_; }

 private:
  torch::Device device_;
  mutable torch::jit::script::Module model_;

  // Helper methods
  static std::string LoadModelFromFile(const std::string& path) {
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file) {
      throw ModelException("Failed to open model file: " + path);
    }

    auto size = file.tellg();
    std::string buffer(size, '\0');
    file.seekg(0);
    if (!file.read(&buffer[0], size)) {
      throw ModelException("Failed to read model file: " + path);
    }
    return buffer;
  }

  void ValidateInput(const torch::Tensor& input) const {
    if (!input.defined()) {
      throw ModelException("Input tensor is undefined");
    }
  }

  void ProcessTupleOutput(const c10::intrusive_ptr<c10::ivalue::Tuple>& tuple,
                          std::vector<torch::Tensor>& tensors) const {
    for (const auto& element : tuple->elements()) {
      if (element.isTensor()) {
        tensors.push_back(element.toTensor());
      } else {
        throw ModelException("Unsupported element type in tuple output");
      }
    }
  }

  void ProcessListOutput(const c10::IValue& list_value,
                         std::vector<torch::Tensor>& tensors) const {
    auto list = list_value.toTensorVector();
    tensors.insert(tensors.end(), list.begin(), list.end());
  }

  void FlattenDictionary(
      const c10::impl::GenericDict& dict, const std::string& parent_key,
      std::unordered_map<std::string, torch::Tensor>& flat_dict) const {
    for (const auto& item : dict) {
      const auto& key = item.key().toStringRef();
      const auto new_key = parent_key.empty() ? key : parent_key + "." + key;
      const auto& value = item.value();

      try {
        if (value.isTensor()) {
          flat_dict[new_key] = value.toTensor();
        } else if (value.isGenericDict()) {
          FlattenDictionary(value.toGenericDict(), new_key, flat_dict);
        } else if (value.isList()) {
          ProcessListInDictionary(value.toList(), new_key, flat_dict);
        } else {
          throw ModelException("Unsupported dictionary value type");
        }
      } catch (const c10::Error& e) {
        throw ModelException("Failed to process dictionary entry '" + new_key +
                             "': " + std::string(e.what()));
      }
    }
  }

  void ProcessListInDictionary(
      const c10::List<c10::IValue>& list, const std::string& key,
      std::unordered_map<std::string, torch::Tensor>& flat_dict) const {
    for (size_t i = 0; i < list.size(); ++i) {
      const auto indexed_key = key + "." + std::to_string(i);
      const auto& item = list.get(i);

      if (item.isTensor()) {
        flat_dict[indexed_key] = item.toTensor();
      } else if (item.isGenericDict()) {
        FlattenDictionary(item.toGenericDict(), indexed_key, flat_dict);
      } else {
        throw ModelException("Unsupported nested dictionary type");
      }
    }
  }
};

/**
 * Convert a VIPS image to a libtorch tensor.
 * @param image Input VIPS image
 * @param device Target device for the tensor
 * @return StatusOr containing a libtorch tensor in [C, H, W] format
 */
absl::StatusOr<torch::Tensor> VImageToTensor(
    const vips::VImage& image, torch::Device device = torch::kCPU) {
  // Validate input image
  if (image.is_null()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Input VImage is not valid");
  }

  // Extract dimensions
  const int width = image.width();
  const int height = image.height();
  const int channels = image.bands();
  const auto format = static_cast<VipsBandFormat>(image.format());

  // Check supported format
  auto format_it = VIPS_TO_TORCH_FORMAT.find(format);
  if (format_it == VIPS_TO_TORCH_FORMAT.end()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Unsupported image format");
  }

  // Tensor options
  auto options =
      torch::TensorOptions().dtype(format_it->second).device(torch::kCPU);

  // Write image data to memory
  size_t buffer_size = 0;
  void* image_data = image.write_to_memory(&buffer_size);
  if (image_data == nullptr) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "Failed to write image data to memory");
  }

  size_t expected_buffer_size =
      width * height * channels * torch::elementSize(format_it->second);

  // Validate buffer size
  if (buffer_size != expected_buffer_size) {
    g_free(image_data);
    return MAKE_STATUS(
        absl::StatusCode::kInternal,
        aifocore::fmt::format(
            "Buffer size mismatch: got {} bytes, expected {} bytes",
            buffer_size, expected_buffer_size));
  }

  // Create tensor from memory
  auto tensor = torch::from_blob(
      image_data, {height, width, channels}, [](void* ptr) { g_free(ptr); },
      options);

  // Convert to desired device and format (CHW)
  return tensor.to(device).permute({2, 0, 1}).contiguous();
}

/**
 * Convert a libtorch tensor to a VIPS image.
 * @param tensor Input tensor in [C, H, W] format
 * @return StatusOr containing a VIPS image
 */
absl::StatusOr<vips::VImage> TensorToVImage(const torch::Tensor& tensor) {
  // Validate tensor
  if (!tensor.defined()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Input tensor is not defined");
  }

  if (tensor.dim() != 3) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Tensor must have 3 dimensions [C, H, W], got " +
                           std::to_string(tensor.dim()) + " dimensions");
  }

  // Get tensor properties
  const int64_t channels = tensor.size(0);
  const int64_t height = tensor.size(1);
  const int64_t width = tensor.size(2);

  // Validate tensor type
  auto format_it = TORCH_TO_VIPS_FORMAT.find(tensor.scalar_type());
  if (format_it == TORCH_TO_VIPS_FORMAT.end()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Unsupported tensor data type");
  }

  // Permute to [H, W, C] and make contiguous
  auto permuted_tensor = tensor.permute({1, 2, 0}).contiguous().to(torch::kCPU);

  // Calculate buffer size
  const size_t element_size = permuted_tensor.element_size();
  const size_t buffer_size = width * height * channels * element_size;

  // Create VImage
  try {
    vips::VImage image = vips::VImage::
        new_from_memory_copy(  // TODO(jonasteuwen): Verify if copy is needed
            permuted_tensor.data_ptr(), buffer_size, width, height, channels,
            format_it->second);
    return image;
  } catch (const std::exception& e) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       std::string("Failed to create VImage: ") + e.what());
  }
}

/**
 * Convert a batch of VIPS images to a torch tensor.
 * @param images Vector of VIPS images
 * @return StatusOr containing a tensor in [N, C, H, W] format
**/
absl::StatusOr<torch::Tensor> ConvertImageBatch(
    const std::vector<vips::VImage>& images,
    torch::Device device = torch::kCPU) {
  std::vector<torch::Tensor> tensors;
  tensors.reserve(images.size());

  for (const auto& image : images) {
    auto tensor_result = VImageToTensor(image, device);
    if (!tensor_result.ok()) {
      return tensor_result.status();
    }
    tensors.push_back(tensor_result.value());
  }

  return torch::stack(tensors);
}

/**
 * Normalize a tensor using the provided mean and standard deviation.
 * @param input_tensor Input tensor
 * @param mean Mean values for each channel
 * @param std Standard deviation values for each channel
 * @return Normalized tensor
 */
torch::Tensor NormalizeTensor(const torch::Tensor& input_tensor,
                              const std::vector<float>& mean,
                              const std::vector<float>& std) {
  // Ensure the input tensor is on the same device as the mean and std
  auto device = input_tensor.device();

  // Create tensors for mean and std, matching the number of channels
  torch::Tensor mean_tensor = torch::tensor(mean, torch::kFloat).to(device);
  torch::Tensor std_tensor = torch::tensor(std, torch::kFloat).to(device);

  // Reshape mean and std to match the input tensor's channel dimension
  mean_tensor = mean_tensor.view({-1, 1, 1});
  std_tensor = std_tensor.view({-1, 1, 1});

  // Perform normalization
  return (input_tensor - mean_tensor) / std_tensor;
}

}  // namespace aifo::utilities

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_TORCH_H_
