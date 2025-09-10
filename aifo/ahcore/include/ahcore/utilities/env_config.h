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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_ENV_CONFIG_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_ENV_CONFIG_H_

#include <algorithm>
#include <cstdlib>
#include <string>
#include <string_view>

namespace aifo::utilities {

/**
 * @brief Class that centralizes environment variable access for AhCore
 * 
 * This class provides a unified way to access environment variables
 * used by the AhCore application. It allows for default values and
 * typed access to environment variables.
 */
class EnvConfig {
 public:
  /**
   * @brief Get a singleton instance of the configuration
   * @return Reference to the singleton instance
   */
  static EnvConfig& GetInstance() {
    static EnvConfig instance;
    return instance;
  }

  /**
   * @brief Check if an environment variable is set to a "true" value
   * 
   * Considers "1", "true", "yes", "on" (case-insensitive) as true values
   * 
   * @param name The environment variable name to check
   * @param default_value Value to return if the environment variable is not set
   * @return bool Whether the environment variable is set to a "true" value
   */
  bool GetBool(std::string_view name, bool default_value = false) const {
    const char* env_value = std::getenv(std::string(name).c_str());
    if (env_value == nullptr) {
      return default_value;
    }

    std::string value = env_value;
    // Convert to lowercase for case-insensitive comparison
    std::transform(value.begin(), value.end(), value.begin(),
                   [](unsigned char chara) { return std::tolower(chara); });

    return (value == "1" || value == "true" || value == "yes" || value == "on");
  }

  /**
   * @brief Get a string environment variable
   * 
   * @param name The environment variable name
   * @param default_value Value to return if the environment variable is not set
   * @return std::string The environment variable value or default_value
   */
  [[nodiscard]] std::string GetString(std::string_view name,
                                      std::string default_value = "") const {
    const char* env_value = std::getenv(std::string(name).c_str());
    if (env_value == nullptr) {
      return default_value;
    }
    return env_value;
  }

  /**
   * @brief Get an integer environment variable
   * 
   * @param name The environment variable name
   * @param default_value Value to return if the environment variable is not set
   * @return int The environment variable value as an integer or default_value
   */
  [[nodiscard]] static int GetInt(std::string_view name,
                                  int default_value = 0) {
    const char* env_value = std::getenv(std::string(name).c_str());
    if (env_value == nullptr) {
      return default_value;
    }

    try {
      return std::stoi(env_value);
    } catch (const std::exception&) {
      return default_value;
    }
  }

  // AhCore specific environment variables
  static constexpr std::string_view kKeepTemporaryFiles =
      "AHCORE_KEEP_TEMPORARY_FILES";

 private:
  // Private constructor to enforce singleton pattern
  EnvConfig() = default;
};

}  // namespace aifo::utilities

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_ENV_CONFIG_H_
