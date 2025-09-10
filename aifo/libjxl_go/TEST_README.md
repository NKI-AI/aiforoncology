# JPEG XL Go Library Test Suite

This directory contains a comprehensive test suite for the `libjxl_go` library, providing Go bindings for the JPEG XL image format. The test suite is designed to mirror and extend the functionality tested in the original C++ `jxl_test.cc` file.

## Test Files Overview

### `libjxl_test.go` - Core Functionality Tests

Contains fundamental tests covering:

- **Basic Operations**: Version checks, signature validation
- **Roundtrip Testing**: Encode/decode cycles with quality verification
- **Image Formats**: Single pixel, small images, different channel configurations
- **Quality Settings**: Lossless and lossy compression with various quality levels
- **PSNR Validation**: Peak Signal-to-Noise Ratio calculations for quality assessment
- **Manual Encoder/Decoder**: Direct API usage without convenience functions
- **Benchmarks**: Performance testing for encoding, decoding, and roundtrip operations

### `advanced_test.go` - Advanced Feature Tests

Tests more sophisticated functionality:

- **Frame Settings**: Various encoding options (effort, resampling, progressive, etc.)
- **Multiple Image Sizes**: Testing different dimensions and aspect ratios
- **Channel Configurations**: RGB, RGBA, Grayscale, Grayscale+Alpha
- **Error Handling**: Validation of error conditions and edge cases
- **Consistency**: Verifying deterministic encoding results
- **Color Encoding**: sRGB and Linear sRGB color spaces
- **Progressive Decoding**: Frame progression events and incremental decoding
- **Quality Analysis**: Detailed quality vs size vs PSNR relationship analysis

### `integration_test.go` - Integration and Stress Tests

Covers complex scenarios and system-level testing:

- **Concurrent Operations**: Multi-threaded encoding/decoding
- **Memory Management**: Resource cleanup and finalizer testing
- **Encoder/Decoder Reuse**: Multiple images with same encoder/decoder instance
- **Large Images**: Performance and memory testing with bigger images (512x512)
- **Edge Cases**: Boundary conditions, extreme aspect ratios, special patterns
- **Robustness**: Handling of invalid data and error conditions
- **Performance Regression**: Basic performance benchmarking

## Running the Tests

### Prerequisites

- Go 1.21+ with modern type support
- Properly built libjxl library with Go bindings
- CGO enabled (`CGO_ENABLED=1`)

### Basic Test Execution

```bash
# Run all tests
go test -v

# Run specific test files
go test -v libjxl_test.go
go test -v advanced_test.go
go test -v integration_test.go

# Run with short mode (skips slow tests)
go test -short

# Run with race detection
go test -race

# Run benchmarks
go test -bench=.
```

### Test Categories

#### Quick Tests (< 1 second each)

- Basic functionality tests
- Small image roundtrips
- API validation
- Error handling

#### Standard Tests (1-10 seconds each)

- Medium-sized image processing
- Quality analysis
- Frame settings validation
- Concurrent operations

#### Slow Tests (> 10 seconds, skipped with -short)

- Large image processing (512x512+)
- Performance regression tests
- Memory stress tests

### Performance Testing

```bash
# Run only benchmarks
go test -run=^$ -bench=.

# Run benchmarks with memory profiling
go test -run=^$ -bench=. -benchmem

# Run benchmarks multiple times for stability
go test -run=^$ -bench=. -count=5
```

## Test Structure and Helpers

### TestImage Helper Class

The test suite includes a `TestImage` helper that provides:

- Configurable dimensions, channels, and bit depth
- Random and pattern-based pixel generation
- Pixel manipulation utilities
- Format conversion helpers

### Key Test Functions

- `Roundtrip()`: Encode and decode an image, returning metrics
- `ComputePSNR()`: Calculate Peak Signal-to-Noise Ratio
- `bytesToFloat32()`: Convert byte data to float32 pixels
- Various image generators for specific test patterns

### Test Validation

Tests validate:

- **Correctness**: Proper encoding/decoding with expected results
- **Performance**: Reasonable compression ratios and processing times
- **Compatibility**: Consistent behavior across different settings
- **Robustness**: Graceful handling of edge cases and errors

## Expected Test Results

### Compression Performance

- **Lossless**: Perfect reconstruction (infinite PSNR)
- **High Quality (90+)**: PSNR > 40 dB, moderate compression
- **Medium Quality (70-90)**: PSNR > 30 dB, good compression
- **Low Quality (<70)**: PSNR > 25 dB, high compression

### Processing Speed (128x128 RGB image)

- **Encoding**: < 1 second per image
- **Decoding**: < 500ms per image
- **Roundtrip**: < 1.5 seconds per image

### Memory Usage

- Encoders/decoders should be properly cleaned up
- No significant memory leaks during repeated operations
- Concurrent operations should not cause excessive memory growth

## Test Data Generation

Since the test suite doesn't rely on external test files (unlike the C++ version), it generates test data programmatically:

- **Random Images**: Using seeded random number generation
- **Pattern Images**: Checkerboards, gradients, solid colors
- **Edge Cases**: Single pixels, extreme aspect ratios
- **Synthetic Data**: Controlled patterns for specific test scenarios

## Common Test Patterns

### Roundtrip Testing

```go
testImage := NewTestImage(width, height, channels)
testImage.RandomFill()
compressedSize, decodedPixels, decodedInfo, err := Roundtrip(testImage, quality)
// Validate size, quality, and correctness
```

### Quality Assessment

```go
decodedFloats := bytesToFloat32(decodedPixels)
psnr := ComputePSNR(testImage.pixels, decodedFloats)
// Verify PSNR meets expectations for given quality
```

### Error Handling

```go
_, err := EncodeOneShot(invalidData, width, height, format, quality)
if err == nil {
    t.Error("Expected error for invalid input")
}
```

## Debugging Failed Tests

### Common Issues

1. **CGO Problems**: Ensure libjxl library is properly linked
2. **Memory Issues**: Check for resource leaks or double-free errors
3. **Quality Issues**: PSNR lower than expected may indicate algorithm changes
4. **Performance Issues**: Slow tests may indicate system load or configuration problems

### Debugging Techniques

- Run with `-v` flag for detailed output
- Use `-race` flag to detect race conditions
- Add debug logging in test helpers
- Check system resources during long-running tests

## Contributing to Tests

When adding new tests:

1. Follow Go testing conventions
2. Use descriptive test names
3. Include both positive and negative test cases
4. Add appropriate benchmarks for performance-critical code
5. Ensure tests are deterministic and reproducible
6. Document any special requirements or setup

### Test Naming Convention

- `Test*`: Standard functionality tests
- `TestAdvanced*`: Complex feature tests
- `TestIntegration*`: System-level tests
- `Benchmark*`: Performance tests

The test suite aims to provide comprehensive coverage of the libjxl_go library while maintaining good performance and reliability.
