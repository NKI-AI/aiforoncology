#include <emscripten.h>
#include <emscripten/bind.h>

#include <jxl/codestream_header.h>
#include <jxl/color_encoding.h>
#include <jxl/decode.h>
#include <jxl/decode_cxx.h>
#include <jxl/thread_parallel_runner.h>
#include <jxl/thread_parallel_runner_cxx.h>
#include <jxl/types.h>

#include <cmath>
#include <cstdint>
#include <cstring>
#include <limits>
#include <numeric>
#include <vector>

extern "C" {

// Structure to hold decode result (matching JavaScript expectations)
struct DecodeResult {
  uint16_t* data;
  uint32_t xsize;
  uint32_t ysize;
  uint32_t channels;
  uint32_t success;
};

// Global storage for results (simple approach for WASM)
static std::vector<uint8_t> g_compressed_data;
static std::vector<uint16_t> g_decoded_data;
static std::vector<uint8_t> g_decoded_rgba_data;
static uint32_t g_rgba_bits_per_sample = 8;

/**
 * Simple JPEG XL multiplex decoding function - matches demo.html interface
 */
EMSCRIPTEN_KEEPALIVE
uint32_t* decode_jxl_multiplex(const uint8_t* compressed,
                               size_t compressed_size) {
  static uint32_t result[5] = {
      0, 0, 0, 0, 0};  // [data_ptr, xsize, ysize, channels, success]
  {
    auto dec = JxlDecoderMake(nullptr);

    if (JXL_DEC_SUCCESS !=
        JxlDecoderSubscribeEvents(dec.get(), JXL_DEC_BASIC_INFO |
                                                 JXL_DEC_COLOR_ENCODING |
                                                 JXL_DEC_FULL_IMAGE)) {
      return result;
    }

    JxlDecoderSetInput(dec.get(), compressed, compressed_size);

    if (JXL_DEC_BASIC_INFO != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    JxlBasicInfo info;
    if (JXL_DEC_SUCCESS != JxlDecoderGetBasicInfo(dec.get(), &info)) {
      return result;
    }

    uint32_t total_channels = info.num_color_channels + info.num_extra_channels;

    if (JXL_DEC_COLOR_ENCODING != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    JxlPixelFormat format = {info.num_color_channels, JXL_TYPE_UINT16,
                             JXL_NATIVE_ENDIAN, 0};

    g_decoded_data.resize(info.xsize * info.ysize * total_channels);

    if (JXL_DEC_NEED_IMAGE_OUT_BUFFER != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    std::vector<uint16_t> color_pixels(info.xsize * info.ysize *
                                       info.num_color_channels);
    if (JXL_DEC_SUCCESS !=
        JxlDecoderSetImageOutBuffer(dec.get(), &format, color_pixels.data(),
                                    color_pixels.size() * sizeof(uint16_t))) {
      return result;
    }

    std::vector<std::vector<uint16_t>> extra_channel_data(
        info.num_extra_channels);
    JxlPixelFormat extra_format = {1, JXL_TYPE_UINT16, JXL_NATIVE_ENDIAN, 0};

    for (uint32_t i = 0; i < info.num_extra_channels; i++) {
      size_t extra_size;
      if (JXL_DEC_SUCCESS != JxlDecoderExtraChannelBufferSize(
                                 dec.get(), &extra_format, &extra_size, i)) {
        return result;
      }

      extra_channel_data[i].resize(extra_size / sizeof(uint16_t));

      if (JXL_DEC_SUCCESS != JxlDecoderSetExtraChannelBuffer(
                                 dec.get(), &extra_format,
                                 extra_channel_data[i].data(), extra_size, i)) {
        return result;
      }
    }

    if (JXL_DEC_FULL_IMAGE != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    if (JXL_DEC_SUCCESS != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    // Copy all channels to output in pixel-major order (interleaved format)
    // This format is: C0P0, C1P0, C2P0, ..., CnP0, C0P1, C1P1, C2P1, ..., CnP1, etc.
    for (size_t pixel = 0; pixel < info.xsize * info.ysize; pixel++) {
      // Copy color channels for this pixel
      for (uint32_t c = 0; c < info.num_color_channels; c++) {
        g_decoded_data[pixel * total_channels + c] =
            color_pixels[pixel * info.num_color_channels + c];
      }

      // Copy extra channels for this pixel
      for (uint32_t c = 0; c < info.num_extra_channels; c++) {
        g_decoded_data[pixel * total_channels + info.num_color_channels + c] =
            extra_channel_data[c][pixel];
      }
    }

    // Set result data - matching demo.html expectations
    result[0] = reinterpret_cast<uint32_t>(g_decoded_data.data());
    result[1] = info.xsize;
    result[2] = info.ysize;
    result[3] = total_channels;
    result[4] = 1;  // Success
  }

  return result;
}

/**
 * Simple JPEG XL RGBA decoding function for regular images
 */
EMSCRIPTEN_KEEPALIVE
uint32_t* decode_jxl_rgba(const uint8_t* compressed, size_t compressed_size) {
  static uint32_t result[5] = {
      0, 0, 0, 0, 0};  // [data_ptr, xsize, ysize, channels, success]
  {
    auto dec = JxlDecoderMake(nullptr);

    if (JXL_DEC_SUCCESS !=
        JxlDecoderSubscribeEvents(dec.get(), JXL_DEC_BASIC_INFO |
                                                 JXL_DEC_COLOR_ENCODING |
                                                 JXL_DEC_FULL_IMAGE)) {
      return result;
    }

    JxlDecoderSetInput(dec.get(), compressed, compressed_size);

    if (JXL_DEC_BASIC_INFO != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    JxlBasicInfo info;
    if (JXL_DEC_SUCCESS != JxlDecoderGetBasicInfo(dec.get(), &info)) {
      return result;
    }

    // For RGBA, we only expect 3-4 channels (RGB or RGBA)
    uint32_t total_channels = info.num_color_channels + info.num_extra_channels;
    if (total_channels < 3 || total_channels > 4) {
      return result;
    }

    if (JXL_DEC_COLOR_ENCODING != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    // Store bits per sample for later retrieval
    g_rgba_bits_per_sample = info.bits_per_sample;

    // Choose format based on bits per sample
    JxlDataType data_type =
        (info.bits_per_sample <= 8) ? JXL_TYPE_UINT8 : JXL_TYPE_UINT16;
    JxlPixelFormat format = {total_channels, data_type, JXL_NATIVE_ENDIAN, 0};

    if (JXL_DEC_NEED_IMAGE_OUT_BUFFER != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    // Calculate buffer size
    size_t buffer_size;
    if (JXL_DEC_SUCCESS !=
        JxlDecoderImageOutBufferSize(dec.get(), &format, &buffer_size)) {
      return result;
    }

    // Resize buffer to accommodate the data
    g_decoded_rgba_data.resize(buffer_size);

    if (JXL_DEC_SUCCESS !=
        JxlDecoderSetImageOutBuffer(dec.get(), &format,
                                    g_decoded_rgba_data.data(), buffer_size)) {
      return result;
    }

    if (JXL_DEC_FULL_IMAGE != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    if (JXL_DEC_SUCCESS != JxlDecoderProcessInput(dec.get())) {
      return result;
    }

    // Set result data - data is already in pixel-major format (RGBARGBARGBA...)
    result[0] = reinterpret_cast<uint32_t>(g_decoded_rgba_data.data());
    result[1] = info.xsize;
    result[2] = info.ysize;
    result[3] = total_channels;
    result[4] = 1;  // Success
  }

  return result;
}

/**
 * Get compressed data (for use with JavaScript)
 */
EMSCRIPTEN_KEEPALIVE
uint8_t* get_compressed_data() {
  return g_compressed_data.data();
}

/**
 * Get compressed data size
 */
EMSCRIPTEN_KEEPALIVE
size_t get_compressed_size() {
  return g_compressed_data.size();
}

/**
 * Get decoded data (for use with JavaScript)
 */
EMSCRIPTEN_KEEPALIVE
uint16_t* get_decoded_data() {
  return g_decoded_data.data();
}

/**
 * Get total decoded data size in elements
 */
EMSCRIPTEN_KEEPALIVE
size_t get_decoded_size() {
  return g_decoded_data.size();
}

/**
 * Get RGBA decoded data (for use with JavaScript)
 */
EMSCRIPTEN_KEEPALIVE
uint8_t* get_decoded_rgba_data() {
  return g_decoded_rgba_data.data();
}

/**
 * Get RGBA decoded data size in bytes
 */
EMSCRIPTEN_KEEPALIVE
size_t get_decoded_rgba_size() {
  return g_decoded_rgba_data.size();
}

/**
 * Get RGBA bits per sample (8 or 16)
 */
EMSCRIPTEN_KEEPALIVE
uint32_t get_rgba_bits_per_sample() {
  return g_rgba_bits_per_sample;
}

}  // extern "C"