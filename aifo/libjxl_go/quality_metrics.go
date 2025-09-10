// Copyright 2025 Jonas Teuwen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package libjxl

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// QualityMetrics contains comprehensive quality metrics for multi-channel data
type QualityMetrics struct {
	PSNRPerChannel []float64
	SSIMPerChannel []float64
	PSNRMean       float64
	PSNRStd        float64
	SSIMMean       float64
	SSIMStd        float64
}

// CalculatePSNR calculates PSNR between two single-channel 16-bit images
func CalculatePSNR(original, decoded []uint16, xsize, ysize uint32) float64 {
	totalPixels := int(xsize) * int(ysize)
	if len(original) < totalPixels || len(decoded) < totalPixels {
		return 0.0
	}

	mse := 0.0
	for i := 0; i < totalPixels; i++ {
		diff := float64(original[i]) - float64(decoded[i])
		mse += diff * diff
	}
	mse /= float64(totalPixels)

	if mse == 0.0 {
		return math.Inf(1) // Perfect match
	}

	const maxVal = 65535.0 // 16-bit max value
	return 20.0 * math.Log10(maxVal/math.Sqrt(mse))
}

// CalculateSSIM calculates SSIM between two single-channel 16-bit images using a sliding window
func CalculateSSIM(original, decoded []uint16, xsize, ysize uint32) float64 {
	const windowSize = 11
	const halfWindow = windowSize / 2
	const k1, k2 = 0.01, 0.03
	const maxVal = 65535.0 // 16-bit max value
	const c1 = (k1 * maxVal) * (k1 * maxVal)
	const c2 = (k2 * maxVal) * (k2 * maxVal)

	ssimSum := 0.0
	validWindows := 0

	for y := halfWindow; y < int(ysize)-halfWindow; y++ {
		for x := halfWindow; x < int(xsize)-halfWindow; x++ {
			sum1, sum2 := 0.0, 0.0
			sum1Sq, sum2Sq, sum12 := 0.0, 0.0, 0.0

			// Calculate means and variances in the window
			for wy := -halfWindow; wy <= halfWindow; wy++ {
				for wx := -halfWindow; wx <= halfWindow; wx++ {
					idx := (y+wy)*int(xsize) + (x + wx)
					val1 := float64(original[idx])
					val2 := float64(decoded[idx])

					sum1 += val1
					sum2 += val2
					sum1Sq += val1 * val1
					sum2Sq += val2 * val2
					sum12 += val1 * val2
				}
			}

			const windowPixels = windowSize * windowSize
			mu1 := sum1 / windowPixels
			mu2 := sum2 / windowPixels
			mu1Sq := mu1 * mu1
			mu2Sq := mu2 * mu2
			mu12 := mu1 * mu2

			sigma1Sq := (sum1Sq / windowPixels) - mu1Sq
			sigma2Sq := (sum2Sq / windowPixels) - mu2Sq
			sigma12 := (sum12 / windowPixels) - mu12

			numerator := (2*mu12 + c1) * (2*sigma12 + c2)
			denominator := (mu1Sq + mu2Sq + c1) * (sigma1Sq + sigma2Sq + c2)

			if denominator != 0.0 {
				ssimSum += numerator / denominator
				validWindows++
			}
		}
	}

	if validWindows > 0 {
		return ssimSum / float64(validWindows)
	}
	return 0.0
}

// CalculateQualityMetrics calculates comprehensive quality metrics for all channels
func CalculateQualityMetrics(original, decoded []uint16, xsize, ysize, numChannels uint32) QualityMetrics {
	metrics := QualityMetrics{}
	metrics.PSNRPerChannel = make([]float64, numChannels)
	metrics.SSIMPerChannel = make([]float64, numChannels)

	pixelsPerChannel := int(xsize) * int(ysize)

	// Calculate per-channel metrics
	for c := uint32(0); c < numChannels; c++ {
		startIdx := int(c) * pixelsPerChannel
		endIdx := startIdx + pixelsPerChannel

		if endIdx <= len(original) && endIdx <= len(decoded) {
			origChannel := original[startIdx:endIdx]
			decChannel := decoded[startIdx:endIdx]

			psnr := CalculatePSNR(origChannel, decChannel, xsize, ysize)
			ssim := CalculateSSIM(origChannel, decChannel, xsize, ysize)

			metrics.PSNRPerChannel[c] = psnr
			metrics.SSIMPerChannel[c] = ssim
		}
	}

	// Calculate PSNR statistics, handling infinite values
	psnrSum := 0.0
	finitePSNRCount := 0

	for c := uint32(0); c < numChannels; c++ {
		if !math.IsInf(metrics.PSNRPerChannel[c], 1) {
			psnrSum += metrics.PSNRPerChannel[c]
			finitePSNRCount++
		}
	}

	// Calculate PSNR statistics only for finite values
	if finitePSNRCount > 0 {
		metrics.PSNRMean = psnrSum / float64(finitePSNRCount)

		psnrVar := 0.0
		for c := uint32(0); c < numChannels; c++ {
			if !math.IsInf(metrics.PSNRPerChannel[c], 1) {
				diff := metrics.PSNRPerChannel[c] - metrics.PSNRMean
				psnrVar += diff * diff
			}
		}
		metrics.PSNRStd = math.Sqrt(psnrVar / float64(finitePSNRCount))
	} else {
		// All values are infinite (perfect reconstruction)
		metrics.PSNRMean = math.Inf(1)
		metrics.PSNRStd = 0.0
	}

	// SSIM values should always be finite
	ssimSum := 0.0
	for c := uint32(0); c < numChannels; c++ {
		ssimSum += metrics.SSIMPerChannel[c]
	}
	metrics.SSIMMean = ssimSum / float64(numChannels)

	ssimVar := 0.0
	for c := uint32(0); c < numChannels; c++ {
		ssimDiff := metrics.SSIMPerChannel[c] - metrics.SSIMMean
		ssimVar += ssimDiff * ssimDiff
	}
	metrics.SSIMStd = math.Sqrt(ssimVar / float64(numChannels))

	return metrics
}

// ComparePixels compares two pixel arrays and reports differences
func ComparePixels(original, decoded []uint16, xsize, ysize, numChannels uint32) (bool, uint32, uint32) {
	if len(original) != len(decoded) {
		return false, 0, 0
	}

	totalPixels := int(xsize) * int(ysize) * int(numChannels)
	differences := uint32(0)
	maxDiff := uint32(0)

	for i := 0; i < totalPixels && i < len(original); i++ {
		if original[i] != decoded[i] {
			differences++
			diff := uint32(math.Abs(float64(original[i]) - float64(decoded[i])))
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	return differences == 0, differences, maxDiff
}

// BytesToUint16 converts byte slice to uint16 slice (little-endian)
func BytesToUint16(bytes []byte) []uint16 {
	result := make([]uint16, len(bytes)/2)
	for i := 0; i < len(result); i++ {
		result[i] = uint16(bytes[i*2]) | uint16(bytes[i*2+1])<<8
	}
	return result
}

// Uint16ToBytes converts uint16 slice to byte slice (little-endian)
func Uint16ToBytes(data []uint16) []byte {
	result := make([]byte, len(data)*2)
	for i, val := range data {
		result[i*2] = byte(val & 0xFF)
		result[i*2+1] = byte(val >> 8)
	}
	return result
}

// LoadTestData loads real test data from binary file and converts from numpy format to channel-major format
// The numpy data is expected to be in shape (height, width, channels) with uint16 values
func LoadTestData(filename string, xsize, ysize, numChannels uint32) ([]uint16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	// Calculate expected file size
	expectedSize := int(xsize) * int(ysize) * int(numChannels)

	// Read the entire file as uint16 values (little-endian)
	tempData := make([]uint16, expectedSize)
	err = binary.Read(file, binary.LittleEndian, tempData)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary data: %w", err)
	}

	// Convert from numpy format (height, width, channels) to channel-major format
	// Numpy shape (512, 512, 32) means data is laid out as:
	// [pixel(0,0,0), pixel(0,0,1), ..., pixel(0,0,31), pixel(0,1,0), ...]
	// We need channel-major: [all_channel_0, all_channel_1, ..., all_channel_31]

	pixels := make([]uint16, expectedSize)

	for c := uint32(0); c < numChannels; c++ {
		for y := uint32(0); y < ysize; y++ {
			for x := uint32(0); x < xsize; x++ {
				// Source: (y, x, c) in numpy format
				srcIdx := int(y)*int(xsize)*int(numChannels) + int(x)*int(numChannels) + int(c)
				// Destination: channel-major format
				dstIdx := int(c)*int(xsize)*int(ysize) + int(y)*int(xsize) + int(x)
				pixels[dstIdx] = tempData[srcIdx]
			}
		}
	}

	return pixels, nil
}
