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
#include "ahcore/utilities/torch.h"

#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

namespace aifo::utilities {

const std::unordered_map<VipsBandFormat, torch::ScalarType>
    VIPS_TO_TORCH_FORMAT = {{VIPS_FORMAT_UCHAR, torch::kUInt8},
                            {VIPS_FORMAT_FLOAT, torch::kFloat32},
                            {VIPS_FORMAT_DOUBLE, torch::kFloat64},
                            {VIPS_FORMAT_INT, torch::kInt32},
                            {VIPS_FORMAT_SHORT, torch::kInt16}};

const std::unordered_map<torch::ScalarType, VipsBandFormat>
    TORCH_TO_VIPS_FORMAT = {{torch::kUInt8, VIPS_FORMAT_UCHAR},
                            {torch::kFloat32, VIPS_FORMAT_FLOAT},
                            {torch::kFloat64, VIPS_FORMAT_DOUBLE},
                            {torch::kInt32, VIPS_FORMAT_INT},
                            {torch::kInt16, VIPS_FORMAT_SHORT}};

torch::Device DeviceManager::GetAvailableDevice(bool force_cpu) {
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

std::string DeviceManager::DeviceName(const torch::Device& device) {
  if (device.type() == torch::kCUDA) {
    return "CUDA";
  }
  if (device.type() == torch::kMPS) {
    return "MPS";
  }
  return "CPU";
}

// TorchModel implementations

// ModelOutput inline helpers
bool TorchModel::ModelOutput::IsSingleTensor() const {
  return std::holds_alternative<torch::Tensor>(data);
}

bool TorchModel::ModelOutput::IsTensorList() const {
  return std::holds_alternative<TensorList>(data);
}

bool TorchModel::ModelOutput::IsTensorDict() const {
  return std::holds_alternative<TensorDict>(data);
}

const torch::Tensor& TorchModel::ModelOutput::AsTensor() const {
  return std::get<torch::Tensor>(data);
}

const TorchModel::ModelOutput::TensorList& TorchModel::ModelOutput::AsList()
    const {
  return std::get<TensorList>(data);
}

const TorchModel::ModelOutput::TensorDict& TorchModel::ModelOutput::AsDict()
    const {
  return std::get<TensorDict>(data);
}

TorchModel::ModelOutput::TensorList TorchModel::ModelOutput::ToList() const {
  return std::visit(
      overloaded{[](const torch::Tensor& t) -> TensorList { return {t}; },
                 [](const TensorList& l) -> TensorList { return l; },
                 [](const TensorDict& d) -> TensorList {
                   TensorList result;
                   for (const auto& kv : d) {
                     result.push_back(kv.second);
                   }
                   return result;
                 }},
      data);
}

TorchModel::ModelOutput::TensorDict TorchModel::ModelOutput::ToDict() const {
  return std::visit(overloaded{[](const torch::Tensor& t) -> TensorDict {
                                 return {{"output", t}};
                               },
                               [](const TensorList& l) -> TensorDict {
                                 TensorDict result;
                                 for (size_t i = 0; i < l.size(); ++i) {
                                   result["output_" + std::to_string(i)] = l[i];
                                 }
                                 return result;
                               },
                               [](const TensorDict& d) -> TensorDict {
                                 return d;
                               }},
                    data);
}

torch::Tensor TorchModel::ModelOutput::ToTensor() const {
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
                         "Cannot convert a dictionary with multiple entries to "
                         "a single tensor");
                   }
                   return d.begin()->second;
                 }},
      data);
}

TorchModel TorchModel::FromFile(const std::string& model_path,
                                torch::Device device) {
  return TorchModel(LoadModelFromFile(model_path), device);
}

TorchModel::TorchModel(const std::string& model_data, torch::Device device) try
    : device_(device) {
  std::istringstream model_stream(model_data);
  model_ = torch::jit::load(model_stream);
  model_.to(device_);
  model_.eval();

} catch (const torch::Error& e) {

  throw ModelException("Failed to initialize model: " + std::string(e.what()));
}

TorchModel::TorchModel(TorchModel&&) noexcept = default;
TorchModel& TorchModel::operator=(TorchModel&&) noexcept = default;

c10::IValue TorchModel::Infer(const torch::Tensor& input) const {
  ValidateInput(input);
  torch::NoGradGuard no_grad;
  try {
    return model_.forward({input.to(device_)});
  } catch (const torch::Error& e) {
    throw ModelException("Inference failed: " + std::string(e.what()));
  }
}

std::vector<torch::Tensor> TorchModel::ProcessVectorOutput(
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
    throw ModelException("Failed to process output: " + std::string(e.what()));
  }
}

std::unordered_map<std::string, torch::Tensor>
TorchModel::ProcessDictionaryOutput(const c10::IValue& output) const {
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

TorchModel::ModelOutput TorchModel::ProcessOutput(
    const c10::IValue& output) const {
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

torch::Device TorchModel::GetDevice() const noexcept {
  return device_;
}

std::string TorchModel::LoadModelFromFile(const std::string& path) {
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

void TorchModel::ValidateInput(const torch::Tensor& input) const {
  if (!input.defined()) {
    throw ModelException("Input tensor is undefined");
  }
}

void TorchModel::ProcessTupleOutput(
    const c10::intrusive_ptr<c10::ivalue::Tuple>& tuple,
    std::vector<torch::Tensor>& tensors) const {
  for (const auto& element : tuple->elements()) {
    if (element.isTensor()) {
      tensors.push_back(element.toTensor());
    } else {
      throw ModelException("Unsupported element type in tuple output");
    }
  }
}

void TorchModel::ProcessListOutput(const c10::IValue& list_value,
                                   std::vector<torch::Tensor>& tensors) const {
  auto list = list_value.toTensorVector();
  tensors.insert(tensors.end(), list.begin(), list.end());
}

void TorchModel::FlattenDictionary(
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

void TorchModel::ProcessListInDictionary(
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

absl::StatusOr<torch::Tensor> VImageToTensor(const vips::VImage& image,
                                             torch::Device device) {
  if (image.is_null()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Input VImage is not valid");
  }

  const int width = image.width();
  const int height = image.height();
  const int channels = image.bands();
  const auto format = static_cast<VipsBandFormat>(image.format());

  auto format_it = VIPS_TO_TORCH_FORMAT.find(format);
  if (format_it == VIPS_TO_TORCH_FORMAT.end()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Unsupported image format");
  }

  auto options =
      torch::TensorOptions().dtype(format_it->second).device(torch::kCPU);

  size_t buffer_size = 0;
  void* image_data = image.write_to_memory(&buffer_size);
  if (image_data == nullptr) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       "Failed to write image data to memory");
  }

  size_t expected_buffer_size =
      width * height * channels * torch::elementSize(format_it->second);

  if (buffer_size != expected_buffer_size) {
    g_free(image_data);
    return MAKE_STATUS(
        absl::StatusCode::kInternal,
        aifocore::fmt::format(
            "Buffer size mismatch: got {} bytes, expected {} bytes",
            buffer_size, expected_buffer_size));
  }

  auto tensor = torch::from_blob(
      image_data, {height, width, channels}, [](void* ptr) { g_free(ptr); },
      options);

  return tensor.to(device).permute({2, 0, 1}).contiguous();
}

absl::StatusOr<vips::VImage> TensorToVImage(const torch::Tensor& tensor) {
  if (!tensor.defined()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Input tensor is not defined");
  }

  if (tensor.dim() != 3) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Tensor must have 3 dimensions [C, H, W], got " +
                           std::to_string(tensor.dim()) + " dimensions");
  }

  const int64_t channels = tensor.size(0);
  const int64_t height = tensor.size(1);
  const int64_t width = tensor.size(2);

  auto format_it = TORCH_TO_VIPS_FORMAT.find(tensor.scalar_type());
  if (format_it == TORCH_TO_VIPS_FORMAT.end()) {
    return MAKE_STATUS(absl::StatusCode::kInvalidArgument,
                       "Unsupported tensor data type");
  }

  auto permuted_tensor = tensor.permute({1, 2, 0}).contiguous().to(torch::kCPU);

  const size_t element_size = permuted_tensor.element_size();
  const size_t buffer_size = width * height * channels * element_size;

  try {
    vips::VImage image = vips::VImage::new_from_memory_copy(
        permuted_tensor.data_ptr(), buffer_size, width, height, channels,
        format_it->second);
    return image;
  } catch (const std::exception& e) {
    return MAKE_STATUS(absl::StatusCode::kInternal,
                       std::string("Failed to create VImage: ") + e.what());
  }
}

absl::StatusOr<torch::Tensor> ConvertImageBatch(
    const std::vector<vips::VImage>& images, torch::Device device) {
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

torch::Tensor NormalizeTensor(const torch::Tensor& input_tensor,
                              const std::vector<float>& mean,
                              const std::vector<float>& std) {
  auto device = input_tensor.device();

  torch::Tensor mean_tensor = torch::tensor(mean, torch::kFloat).to(device);
  torch::Tensor std_tensor = torch::tensor(std, torch::kFloat).to(device);

  mean_tensor = mean_tensor.view({-1, 1, 1});
  std_tensor = std_tensor.view({-1, 1, 1});

  return (input_tensor - mean_tensor) / std_tensor;
}

}  // namespace aifo::utilities
