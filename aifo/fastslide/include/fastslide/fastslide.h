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

#ifndef AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_FASTSLIDE_H_
#define AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_FASTSLIDE_H_

/**
 * @file fastslide.h
 * @brief Main header for the FastSlide library
 * 
 * This header includes all the core FastSlide functionality for reading
 * digital pathology slides in various formats (SVS, QPTIFF, etc.).
 */

// Core slide reading functionality
#include "fastslide/readers/readers.h"
#include "fastslide/slide_reader.h"

// Image and histogram utilities
#include "fastslide/histogram.h"
#include "fastslide/image.h"

// Caching utilities
#include "fastslide/utilities/cache.h"

// Color utilities
#include "fastslide/utilities/colors.h"

// Utility functions
#include "fastslide/utilities/combine.h"

#endif  // AIFO_FASTSLIDE_INCLUDE_FASTSLIDE_FASTSLIDE_H_
