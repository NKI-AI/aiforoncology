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
#ifndef AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_BLOSC2_H_
#define AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_BLOSC2_H_

#include <blosc2.h>

/**
 * @class Blosc2Context
 * @brief Manages the Blosc2 library lifecycle using RAII.
 */
class Blosc2Context {
 public:
  /**
   * @brief Constructor that initializes the Blosc2 library.
   */
  Blosc2Context() {
    blosc2_init();  // Initialize Blosc2 library
  }

  /**
   * @brief Destructor that cleans up the Blosc2 library.
   */
  ~Blosc2Context() {
    blosc2_destroy();  // Clean up Blosc2 library
  }

  // Disable copy and move semantics to prevent multiple initializations
  Blosc2Context(const Blosc2Context&) = delete;
  Blosc2Context& operator=(const Blosc2Context&) = delete;
  Blosc2Context(Blosc2Context&&) = delete;
  Blosc2Context& operator=(Blosc2Context&&) = delete;
};

#endif  // AIFO_AHCORE_INCLUDE_AHCORE_UTILITIES_BLOSC2_H_
