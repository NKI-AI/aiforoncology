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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_LOGGING_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_LOGGING_H_
#include <spdlog/sinks/stdout_color_sinks.h>  // For colored console output
#include <spdlog/spdlog.h>
#include <cstdlib>  // For getenv
#include <memory>
#include <string>
#include <unordered_map>

// TODO(jonasteuwen): Use env_config.h to set the log level.
namespace aifo::logging {

/**
 * @brief Get or create the global logger instance with the name "aifo_logger".
 */
inline std::shared_ptr<spdlog::logger> GetGlobalLogger() {
  auto logger = spdlog::get("aifo_logger");
  if (!logger) {
    logger = spdlog::stdout_color_mt("aifo_logger");  // Colored console logger
    spdlog::set_default_logger(logger);  // Optionally set it as default
  }
  return logger;
}

/**
 * @brief Sets the global log level from the environment variable AIFO_LOG_LEVEL.
 */
inline void SetGlobalLogLevelFromEnv() {
  auto logger = GetGlobalLogger();
  const char* log_level_env = std::getenv("AIFO_LOG_LEVEL");

  if (log_level_env) {
    std::string level_str(log_level_env);
    std::unordered_map<std::string, spdlog::level::level_enum> log_levels = {
        {"trace", spdlog::level::trace}, {"debug", spdlog::level::debug},
        {"info", spdlog::level::info},   {"warn", spdlog::level::warn},
        {"error", spdlog::level::err},   {"critical", spdlog::level::critical},
        {"off", spdlog::level::off}};

    auto level_iter = log_levels.find(level_str);
    if (level_iter != log_levels.end()) {
      logger->set_level(level_iter->second);
      logger->info("Global log level set to: {}", level_str);
    } else {
      logger->warn("Invalid AIFO_LOG_LEVEL value '{}'. Falling back to 'info'.",
                   level_str);
      logger->set_level(spdlog::level::info);
    }
  } else {
    logger->info("AIFO_LOG_LEVEL not set. Defaulting to 'info' level.");
    logger->set_level(spdlog::level::info);
  }
}

}  // namespace aifo::logging

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_LOGGING_H_
