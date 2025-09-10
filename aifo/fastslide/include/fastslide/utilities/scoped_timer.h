// Copyright 2025 Jonas Teuwen. All Rights Reserved.
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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_SCOPED_TIMER_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_SCOPED_TIMER_H_

#include <chrono>
#include <string>
#include <string_view>

#include "absl/log/log.h"

namespace fastslide {

#ifdef FASTSLIDE_ENABLE_TIMERS

/// \brief RAII scoped timer that logs elapsed time at scope exit.
/// \details Enable by defining FASTSLIDE_ENABLE_TIMERS at build time.
class ScopedTimer {
 public:
  explicit ScopedTimer(std::string_view label)
      : label_(label), start_(std::chrono::steady_clock::now()) {}

  ScopedTimer(const ScopedTimer&) = delete;
  ScopedTimer& operator=(const ScopedTimer&) = delete;

  ScopedTimer(ScopedTimer&&) = delete;
  ScopedTimer& operator=(ScopedTimer&&) = delete;

  ~ScopedTimer() {
    const auto end = std::chrono::steady_clock::now();
    const auto us =
        std::chrono::duration_cast<std::chrono::microseconds>(end - start_)
            .count();
    // Log in microseconds for precision; readers can aggregate externally.
    LOG(INFO) << "[Perf] " << label_ << ": " << us << " us";
  }

 private:
  std::string label_;
  std::chrono::steady_clock::time_point start_;
};

#else  // FASTSLIDE_ENABLE_TIMERS

// No-op timer when timers are disabled.
class ScopedTimer {
 public:
  explicit ScopedTimer(std::string_view /*label*/) {}
};

#endif  // FASTSLIDE_ENABLE_TIMERS

}  // namespace fastslide

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_UTILITIES_SCOPED_TIMER_H_
