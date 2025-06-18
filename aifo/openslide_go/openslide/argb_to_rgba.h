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
#include <stddef.h>
#include <stdint.h>

typedef unsigned char u8;

// Converts an array of ARGB pixels (premultiplied alpha) to RGBA format in-place.
void ArgbToRgba(uint32_t* buf, size_t len) {
  for (size_t i = 0; i < len; ++i) {
    uint32_t pixel = buf[i];

    u8 a = (pixel >> 24) & 0xff;
    u8 r = (pixel >> 16) & 0xff;
    u8 g = (pixel >> 8) & 0xff;
    u8 b = (pixel >> 0) & 0xff;

    if (a != 0 && a != 255) {
      // Un-premultiply if alpha is not opaque or transparent
      r = (u8)(255 * r / a);
      g = (u8)(255 * g / a);
      b = (u8)(255 * b / a);
    }

    // Repack as RGBA
#if defined(__BYTE_ORDER__) && __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
    buf[i] = (r << 24) | (g << 16) | (b << 8) | a;
#else
    buf[i] = (a << 24) | (b << 16) | (g << 8) | r;
#endif
  }
}
