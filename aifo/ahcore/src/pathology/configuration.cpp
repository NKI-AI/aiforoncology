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
#include <memory>
#include <string>

#include "ahcore/pathology/configuration.h"

// Helper to get text content of an XML node
std::string GetNodeContent(const xmlNode* node) {
  if (node && node->children && node->children->content) {
    return reinterpret_cast<const char*>(node->children->content);
  }
  return {};
}

// Core parsing function
std::unique_ptr<AifoModelConfiguration> ParseConfiguration(xmlDoc* doc) {
  if (!doc) {
    throw std::runtime_error("Failed to parse XML.");
  }

  auto root = xmlDocGetRootElement(doc);
  if (!root || xmlStrcmp(root->name, reinterpret_cast<const xmlChar*>(
                                         "AifoModelConfiguration"))) {
    xmlFreeDoc(doc);
    throw std::runtime_error(
        "Invalid root element. Expected 'AifoModelConfiguration'.");
  }

  auto config = std::make_unique<AifoModelConfiguration>();

  for (xmlNode* node = root->children; node; node = node->next) {
    if (node->type != XML_ELEMENT_NODE) {
      continue;
    }

    std::string node_name = reinterpret_cast<const char*>(node->name);

    if (node_name == "ModelName") {
      config->model_name = GetNodeContent(node);
    } else if (node_name == "Version") {
      config->version = GetNodeContent(node);
    } else if (node_name == "Task") {
      for (xmlNode* task_node = node->children; task_node;
           task_node = task_node->next) {
        if (task_node->type != XML_ELEMENT_NODE) {
          continue;
        }
        std::string task_node_name =
            reinterpret_cast<const char*>(task_node->name);
        if (task_node_name == "Type") {
          config->task_type = GetNodeContent(task_node);
        } else if (task_node_name == "MergeMethod") {
          config->merge_method = GetNodeContent(task_node);
        }
      }
    } else if (node_name == "Mpp") {
      config->mpp = std::stod(GetNodeContent(node));
    } else if (node_name == "TileSize") {
      int width = 0, height = 0;
      for (xmlNode* size_node = node->children; size_node;
           size_node = size_node->next) {
        if (size_node->type != XML_ELEMENT_NODE) {
          continue;
        }
        std::string size_node_name =
            reinterpret_cast<const char*>(size_node->name);
        if (size_node_name == "Width") {
          width = std::stoi(GetNodeContent(size_node));
        } else if (size_node_name == "Height") {
          height = std::stoi(GetNodeContent(size_node));
        }
      }
      config->tile_size = {width, height};
    } else if (node_name == "TileOverlap") {
      int width = 0, height = 0;
      for (xmlNode* overlap_node = node->children; overlap_node;
           overlap_node = overlap_node->next) {
        if (overlap_node->type != XML_ELEMENT_NODE) {
          continue;
        }
        std::string overlap_node_name =
            reinterpret_cast<const char*>(overlap_node->name);
        if (overlap_node_name == "Width") {
          width = std::stoi(GetNodeContent(overlap_node));
        } else if (overlap_node_name == "Height") {
          height = std::stoi(GetNodeContent(overlap_node));
        }
      }
      config->tile_overlap = {width, height};
    } else if (node_name == "Labels") {
      for (xmlNode* label_node = node->children; label_node;
           label_node = label_node->next) {
        if (label_node->type != XML_ELEMENT_NODE) {
          continue;
        }
        if (std::string(reinterpret_cast<const char*>(label_node->name)) ==
            "Label") {
          Label label;
          for (xmlNode* label_attr = label_node->children; label_attr;
               label_attr = label_attr->next) {
            if (label_attr->type != XML_ELEMENT_NODE) {
              continue;
            }
            std::string label_attr_name =
                reinterpret_cast<const char*>(label_attr->name);
            if (label_attr_name == "Name") {
              label.name = GetNodeContent(label_attr);
            } else if (label_attr_name == "HexColor") {
              label.hex_color = GetNodeContent(label_attr);
            } else if (label_attr_name == "Index") {
              label.index = std::stoi(GetNodeContent(label_attr));
            }
          }
          config->labels.push_back(label);
        }
      }
    } else if (node_name == "Normalization") {
      for (xmlNode* norm_node = node->children; norm_node;
           norm_node = norm_node->next) {
        if (norm_node->type != XML_ELEMENT_NODE) {
          continue;
        }
        std::string norm_node_name =
            reinterpret_cast<const char*>(norm_node->name);
        if (norm_node_name == "Mean") {
          for (xmlNode* mean_node = norm_node->children; mean_node;
               mean_node = mean_node->next) {
            if (mean_node->type != XML_ELEMENT_NODE) {
              continue;
            }
            config->normalization.mean.push_back(
                std::stof(GetNodeContent(mean_node)));
          }
        } else if (norm_node_name == "Std") {
          for (xmlNode* std_node = norm_node->children; std_node;
               std_node = std_node->next) {
            if (std_node->type != XML_ELEMENT_NODE) {
              continue;
            }
            config->normalization.std.push_back(
                std::stof(GetNodeContent(std_node)));
          }
        }
      }
    }
  }

  xmlFreeDoc(doc);
  return config;
}

// Parsing from file
std::unique_ptr<AifoModelConfiguration> ParseConfigurationFromFile(
    const std::string& filename) {
  xmlDoc* doc = xmlReadFile(filename.c_str(), nullptr, 0);
  return ParseConfiguration(doc);
}

// Parsing from string
std::unique_ptr<AifoModelConfiguration> ParseConfigurationFromString(
    const std::string& xml_content) {
  xmlDoc* doc = xmlReadMemory(xml_content.c_str(), xml_content.size(), nullptr,
                              nullptr, 0);
  return ParseConfiguration(doc);
}
