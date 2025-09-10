package libjxl_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	libjxl "aifo.dev/aifo/libjxl_go"
)

// Advanced test cases that mirror specific tests from jxl_test.cc

// TestAdvancedEncoding tests various encoding parameters and settings
func TestAdvancedEncoding(t *testing.T) {
	testImage := NewTestImage(128, 128, 3)
	testImage.RandomFill()

	tests := []struct {
		name          string
		quality       float32
		expectMaxSize int
		expectMinPSNR float64
	}{
		{"high_quality", 95, 50000, 24.0},
		{"medium_quality", 80, 30000, 13.0},
		{"low_quality", 60, 20000, 10.0},
		{"very_low_quality", 30, 15000, 9.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressedSize, decodedPixels, _, err := Roundtrip(testImage, tt.quality)
			if err != nil {
				t.Fatalf("Roundtrip failed: %v", err)
			}

			if compressedSize > tt.expectMaxSize {
				t.Errorf("Compressed size %d exceeds maximum %d for quality %.0f",
					compressedSize, tt.expectMaxSize, tt.quality)
			}

			// Compute PSNR
			decodedFloats := bytesToFloat32(decodedPixels)
			psnr := ComputePSNR(testImage.pixels, decodedFloats)

			if !math.IsInf(psnr, 1) && psnr < tt.expectMinPSNR {
				t.Errorf("PSNR %.2f below minimum %.2f for quality %.0f",
					psnr, tt.expectMinPSNR, tt.quality)
			}

			t.Logf("Quality %.0f: %d bytes, PSNR %.2f dB", tt.quality, compressedSize, psnr)
		})
	}
}

// TestFrameSettings tests various frame setting options
func TestFrameSettings(t *testing.T) {
	testImage := NewTestImage(64, 64, 3)
	testImage.RandomFill()

	encoder, err := libjxl.NewEncoder()
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	basicInfo := testImage.BasicInfo()
	// Required by libjxl when enabling lossless
	basicInfo.UsesOriginalProfile = true
	err = encoder.SetBasicInfo(basicInfo)
	if err != nil {
		t.Fatalf("Failed to set basic info: %v", err)
	}

	var colorEncoding libjxl.ColorEncoding
	colorEncoding.SetToSRGB(false)
	err = encoder.SetColorEncoding(colorEncoding)
	if err != nil {
		t.Fatalf("Failed to set color encoding: %v", err)
	}

	frameSettings := encoder.FrameSettingsCreate()

	// Test different effort levels
	efforts := []int64{1, 3, 5, 7}
	for _, effort := range efforts {
		err = frameSettings.SetOption(libjxl.FrameSettingEffort, effort)
		if err != nil {
			t.Errorf("Failed to set effort %d: %v", effort, err)
		} else {
			t.Logf("Successfully set effort level %d", effort)
		}
	}

	// Test distance setting
	err = frameSettings.SetDistance(1.0)
	if err != nil {
		t.Errorf("Failed to set distance: %v", err)
	}

	// Test lossless setting
	err = frameSettings.SetLossless(true)
	if err != nil {
		t.Errorf("Failed to set lossless: %v", err)
	}

	// Test various frame settings
	settingsTests := []struct {
		setting libjxl.FrameSettingId
		value   int64
		name    string
	}{
		{libjxl.FrameSettingDecodingSpeed, 1, "decoding_speed"},
		{libjxl.FrameSettingResampling, 1, "resampling"},
		{libjxl.FrameSettingEPF, 2, "epf"},
		{libjxl.FrameSettingGaborish, 1, "gaborish"},
		{libjxl.FrameSettingModular, 0, "modular"},
		{libjxl.FrameSettingResponsive, 1, "responsive"},
		{libjxl.FrameSettingProgressiveAC, 0, "progressive_ac"},
		{libjxl.FrameSettingProgressiveDC, 0, "progressive_dc"},
	}

	for _, test := range settingsTests {
		err = frameSettings.SetOption(test.setting, test.value)
		if err != nil {
			t.Errorf("Failed to set %s to %d: %v", test.name, test.value, err)
		} else {
			t.Logf("Successfully set %s to %d", test.name, test.value)
		}
	}
}

// TestDifferentImageSizes tests encoding/decoding with various image dimensions
func TestDifferentImageSizes(t *testing.T) {
	sizes := []struct {
		width, height uint32
		name          string
	}{
		{1, 1, "single_pixel"},
		{8, 8, "tiny"},
		{32, 32, "small"},
		{64, 128, "rectangular"},
		{128, 64, "wide"},
		{256, 256, "medium"},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			testImage := NewTestImage(size.width, size.height, 3)
			testImage.RandomFill()

			compressedSize, decodedPixels, decodedInfo, err := Roundtrip(testImage, 90)
			if err != nil {
				t.Fatalf("Roundtrip failed for %dx%d: %v", size.width, size.height, err)
			}

			if decodedInfo.XSize != size.width || decodedInfo.YSize != size.height {
				t.Errorf("Size mismatch: got %dx%d, want %dx%d",
					decodedInfo.XSize, decodedInfo.YSize, size.width, size.height)
			}

			// decodedPixels are uint8 bytes, not float32s. Expect width*height*channels bytes.
			expectedBytes := int(size.width * size.height * 3)
			if len(decodedPixels) != expectedBytes {
				t.Errorf("Pixel count mismatch: got %d bytes, want %d bytes",
					len(decodedPixels), expectedBytes)
			}

			t.Logf("%dx%d (%s): %d bytes compressed", size.width, size.height, size.name, compressedSize)
		})
	}
}

// TestChannelConfigurations tests different channel configurations
func TestChannelConfigurations(t *testing.T) {
	configs := []struct {
		channels uint32
		name     string
		hasAlpha bool
	}{
		{1, "grayscale", false},
		{2, "grayscale_alpha", true},
		{3, "rgb", false},
		{4, "rgba", true},
	}

	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			testImage := NewTestImage(32, 32, config.channels)
			testImage.RandomFill()

			// Set alpha channel to full opacity if present
			if config.hasAlpha {
				alphaChannel := config.channels - 1
				for y := uint32(0); y < 32; y++ {
					for x := uint32(0); x < 32; x++ {
						testImage.SetPixel(x, y, alphaChannel, 1.0)
					}
				}
			}

			compressedSize, decodedPixels, decodedInfo, err := Roundtrip(testImage, 95)
			if err != nil {
				t.Fatalf("Roundtrip failed for %d channels: %v", config.channels, err)
			}

			// Verify channel information
			// For GA we expect 1 color channel (gray) and extra alpha
			expectedColorChannels := min(config.channels, 3)
			if config.channels == 2 {
				expectedColorChannels = 1
			}
			if decodedInfo.NumColorChannels != expectedColorChannels {
				t.Errorf("Color channel mismatch: got %d, want %d",
					decodedInfo.NumColorChannels, expectedColorChannels)
			}

			if config.hasAlpha && decodedInfo.AlphaBits == 0 {
				t.Error("Expected alpha channel but none found")
			}
			if !config.hasAlpha && decodedInfo.AlphaBits != 0 {
				t.Error("Unexpected alpha channel found")
			}

			// Verify pixel data size
			expectedBytes := int(32 * 32 * config.channels)
			if len(decodedPixels) != expectedBytes {
				t.Errorf("Pixel data size mismatch: got %d bytes, want %d bytes",
					len(decodedPixels), expectedBytes)
			}

			t.Logf("%d channels (%s): %d bytes compressed", config.channels, config.name, compressedSize)
		})
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	t.Run("empty_data_encode", func(t *testing.T) {
		format := libjxl.PixelFormat{NumChannels: 3, DataType: libjxl.TypeFloat}
		_, err := libjxl.EncodeOneShot([]byte{}, 10, 10, format, 90)
		if err == nil {
			t.Error("Expected error for empty pixel data")
		}
	})

	t.Run("empty_data_decode", func(t *testing.T) {
		format := libjxl.PixelFormat{NumChannels: 3, DataType: libjxl.TypeFloat}
		_, _, _, err := libjxl.DecodeOneShot([]byte{}, format)
		if err == nil {
			t.Error("Expected error for empty compressed data")
		}
	})

	t.Run("invalid_dimensions", func(t *testing.T) {
		format := libjxl.PixelFormat{NumChannels: 3, DataType: libjxl.TypeFloat}
		smallData := make([]byte, 100) // Too small for claimed dimensions
		_, err := libjxl.EncodeOneShot(smallData, 1000, 1000, format, 90)
		if err == nil {
			t.Error("Expected error for insufficient pixel data")
		}
	})

	t.Run("decoder_without_events", func(t *testing.T) {
		decoder, err := libjxl.NewDecoder()
		if err != nil {
			t.Fatalf("Failed to create decoder: %v", err)
		}
		defer decoder.Close()

		// Try to process without subscribing to events or setting input
		status := decoder.ProcessInput()
		// Some libjxl builds return DecNeedMoreInput here instead of DecError
		if status != libjxl.DecError && status != libjxl.DecNeedMoreInput {
			t.Errorf("Expected libjxl.DecError or libjxl.DecNeedMoreInput, got %v", status)
		}
	})
}

// TestConsistency verifies that encoding the same data multiple times produces identical results
func TestConsistency(t *testing.T) {
	testImage := NewTestImage(64, 64, 3)
	testImage.RandomFill()

	// Encode the same image multiple times
	var results [][]byte
	for i := 0; i < 3; i++ {
		compressed, err := libjxl.EncodeOneShot(testImage.ToBytes(), testImage.width, testImage.height, testImage.format, 100)
		if err != nil {
			t.Fatalf("Encoding attempt %d failed: %v", i+1, err)
		}
		results = append(results, compressed)
	}

	// All results should be identical for lossless encoding
	for i := 1; i < len(results); i++ {
		if !bytes.Equal(results[0], results[i]) {
			t.Errorf("Encoding attempt %d produced different result than attempt 1", i+1)
		}
	}

	t.Logf("Consistency test passed: %d identical results of %d bytes each", len(results), len(results[0]))
}

// TestLinearSRGB tests linear sRGB color encoding
func TestLinearSRGB(t *testing.T) {
	testImage := NewTestImage(32, 32, 3)
	testImage.RandomFill()

	encoder, err := libjxl.NewEncoder()
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	basicInfo := testImage.BasicInfo()
	// Required by libjxl when enabling lossless
	basicInfo.UsesOriginalProfile = true
	err = encoder.SetBasicInfo(basicInfo)
	if err != nil {
		t.Fatalf("Failed to set basic info: %v", err)
	}

	// Test linear sRGB color encoding
	var colorEncoding libjxl.ColorEncoding
	colorEncoding.SetToLinearSRGB(false)
	err = encoder.SetColorEncoding(colorEncoding)
	if err != nil {
		t.Fatalf("Failed to set linear sRGB color encoding: %v", err)
	}

	frameSettings := encoder.FrameSettingsCreate()
	err = frameSettings.SetLossless(true)
	if err != nil {
		t.Fatalf("Failed to set lossless: %v", err)
	}

	err = frameSettings.AddImageFrame(testImage.format, testImage.ToBytes())
	if err != nil {
		t.Fatalf("Failed to add image frame: %v", err)
	}

	encoder.CloseInput()

	// Process output
	var compressed []byte
	buffer := make([]byte, 64*1024)

	for {
		bytesWritten, status, err := encoder.ProcessOutput(buffer)
		if err != nil {
			t.Fatalf("Failed to process output: %v", err)
		}

		compressed = append(compressed, buffer[:bytesWritten]...)

		switch status {
		case libjxl.EncSuccess:
			t.Logf("Linear sRGB encoding successful: %d bytes", len(compressed))
			return
		case libjxl.EncNeedMoreOutput:
			continue
		case libjxl.EncError:
			t.Fatal("Encoder error")
		}
	}
}

// TestProgressiveDecoding tests progressive decoding capabilities
func TestProgressiveDecoding(t *testing.T) {
	// Create a test image
	testImage := NewTestImage(128, 128, 3)
	testImage.RandomFill()

	// First encode with progressive settings
	encoder, err := libjxl.NewEncoder()
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	basicInfo := testImage.BasicInfo()
	// Required by libjxl when enabling lossless
	basicInfo.UsesOriginalProfile = true
	err = encoder.SetBasicInfo(basicInfo)
	if err != nil {
		t.Fatalf("Failed to set basic info: %v", err)
	}

	var colorEncoding libjxl.ColorEncoding
	colorEncoding.SetToSRGB(false)
	err = encoder.SetColorEncoding(colorEncoding)
	if err != nil {
		t.Fatalf("Failed to set color encoding: %v", err)
	}

	frameSettings := encoder.FrameSettingsCreate()
	err = frameSettings.SetDistance(1.0)
	if err != nil {
		t.Fatalf("Failed to set distance: %v", err)
	}

	// Enable progressive encoding
	err = frameSettings.SetOption(libjxl.FrameSettingProgressiveAC, 1)
	if err != nil {
		t.Fatalf("Failed to enable progressive AC: %v", err)
	}

	err = frameSettings.SetOption(libjxl.FrameSettingProgressiveDC, 1)
	if err != nil {
		t.Fatalf("Failed to enable progressive DC: %v", err)
	}

	err = frameSettings.AddImageFrame(testImage.format, testImage.ToBytes())
	if err != nil {
		t.Fatalf("Failed to add image frame: %v", err)
	}

	encoder.CloseInput()

	// Get compressed data
	var compressed []byte
	buffer := make([]byte, 64*1024)

	for {
		bytesWritten, status, err := encoder.ProcessOutput(buffer)
		if err != nil {
			t.Fatalf("Failed to process output: %v", err)
		}

		compressed = append(compressed, buffer[:bytesWritten]...)

		switch status {
		case libjxl.EncSuccess:
			goto decode
		case libjxl.EncNeedMoreOutput:
			continue
		case libjxl.EncError:
			t.Fatal("Encoder error")
		}
	}

decode:
	// Now decode progressively
	decoder, err := libjxl.NewDecoder()
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// Subscribe to frame progression events
	err = decoder.SubscribeEvents(libjxl.EventBasicInfo | libjxl.EventFullImage | libjxl.EventFrameProgression)
	if err != nil {
		t.Fatalf("Failed to subscribe to events: %v", err)
	}

	err = decoder.SetInput(compressed)
	if err != nil {
		t.Fatalf("Failed to set input: %v", err)
	}
	decoder.CloseInput()

	progressionCount := 0
	var finalPixels []byte

	for {
		status := decoder.ProcessInput()
		switch status {
		case libjxl.DecError:
			t.Fatal("Decoder error")
		case libjxl.DecBasicInfo:
			info, err := decoder.GetBasicInfo()
			if err != nil {
				t.Fatalf("Failed to get basic info: %v", err)
			}
			if info.XSize != testImage.width || info.YSize != testImage.height {
				t.Errorf("Size mismatch: got %dx%d, want %dx%d",
					info.XSize, info.YSize, testImage.width, testImage.height)
			}
		case libjxl.DecFrameProgression:
			progressionCount++
			t.Logf("Progressive frame update #%d", progressionCount)
		case libjxl.DecNeedImageOutBuffer:
			bufferSize, err := decoder.ImageOutBufferSize(testImage.format)
			if err != nil {
				t.Fatalf("Failed to get buffer size: %v", err)
			}
			finalPixels = make([]byte, bufferSize)
			err = decoder.SetImageOutBuffer(testImage.format, finalPixels)
			if err != nil {
				t.Fatalf("Failed to set output buffer: %v", err)
			}
		case libjxl.DecFullImage:
			// Image is fully decoded
		case libjxl.DecSuccess:
			t.Logf("Progressive decoding completed with %d progression updates", progressionCount)
			if len(finalPixels) == 0 {
				t.Error("No pixel data received")
			}
			return
		}
	}
}

// TestQualityVsSize tests the relationship between quality and compressed size
func TestQualityVsSize(t *testing.T) {
	testImage := NewTestImage(128, 128, 3)
	testImage.RandomFill()

	type result struct {
		quality float32
		size    int
		psnr    float64
	}

	var results []result

	qualities := []float32{100, 95, 90, 85, 80, 75, 70, 65, 60, 55, 50}
	for _, quality := range qualities {
		compressedSize, decodedPixels, _, err := Roundtrip(testImage, quality)
		if err != nil {
			t.Fatalf("Roundtrip failed for quality %.0f: %v", quality, err)
		}

		decodedFloats := bytesToFloat32(decodedPixels)
		psnr := ComputePSNR(testImage.pixels, decodedFloats)

		results = append(results, result{
			quality: quality,
			size:    compressedSize,
			psnr:    psnr,
		})
	}

	// Print results table
	t.Log("Quality vs Size vs PSNR:")
	t.Log("Quality | Size (bytes) | PSNR (dB)")
	t.Log("--------|--------------|----------")
	for _, r := range results {
		psnrStr := "∞"
		if !math.IsInf(r.psnr, 1) {
			psnrStr = fmt.Sprintf("%.1f", r.psnr)
		}
		t.Logf("  %3.0f   | %10d   | %8s", r.quality, r.size, psnrStr)
	}

	// Verify general trends
	for i := 1; i < len(results); i++ {
		// Generally, higher quality should result in larger files
		// Allow some tolerance due to encoding complexity
		if results[i].quality > results[i-1].quality && results[i].size < results[i-1].size/2 {
			t.Logf("Warning: Quality %.0f has unexpectedly small size compared to quality %.0f",
				results[i].quality, results[i-1].quality)
		}
	}
}
