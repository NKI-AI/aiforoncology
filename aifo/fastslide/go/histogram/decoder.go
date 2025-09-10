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

package histogram

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// HistogramData represents a decoded FastSlide histogram
type HistogramData struct {
	Version   uint32    `json:"version"`
	NBins     uint32    `json:"nBins"`
	BinEdges  []float64 `json:"binEdges"`
	BinCounts []int64   `json:"binCounts"`
	MinValue  float64   `json:"minValue"`
	MaxValue  float64   `json:"maxValue"`
	MeanValue float64   `json:"meanValue"`
	StdDev    float64   `json:"stdDev"`
	Sum       float64   `json:"sum"`
	NValues   int64     `json:"nValues"`
	NMissing  int64     `json:"nMissing"`
}

// HistogramError represents an error that occurred during histogram operations
type HistogramError struct {
	Operation string
	Message   string
}

func (e *HistogramError) Error() string {
	return fmt.Sprintf("histogram %s error: %s", e.Operation, e.Message)
}

// DecodeBinaryHistogram decodes a binary histogram from a byte slice
func DecodeBinaryHistogram(data []byte) (*HistogramData, error) {
	if len(data) < 12 {
		return nil, &HistogramError{
			Operation: "decode",
			Message:   fmt.Sprintf("data too short: need at least 12 bytes for header, got %d", len(data)),
		}
	}

	// Check magic header
	magic := string(data[0:4])
	if magic != "HIST" {
		return nil, &HistogramError{
			Operation: "decode",
			Message:   fmt.Sprintf("invalid magic header: expected 'HIST', got '%s'", magic),
		}
	}

	// Read version (little-endian uint32)
	version := binary.LittleEndian.Uint32(data[4:8])
	if version != 1 {
		return nil, &HistogramError{
			Operation: "decode",
			Message:   fmt.Sprintf("unsupported version: %d", version),
		}
	}

	// Read number of bins (little-endian uint32)
	nBins := binary.LittleEndian.Uint32(data[8:12])

	// Calculate expected sizes
	headerSize := 12
	edgesSize := int(nBins+1) * 8 // double = 8 bytes
	countsSize := int(nBins) * 8  // int64 = 8 bytes
	statsSize := 7 * 8            // 7 doubles = 56 bytes
	expectedSize := headerSize + edgesSize + countsSize + statsSize

	if len(data) < expectedSize {
		return nil, &HistogramError{
			Operation: "decode",
			Message:   fmt.Sprintf("data too short: need %d bytes, got %d", expectedSize, len(data)),
		}
	}

	hist := &HistogramData{
		Version:   version,
		NBins:     nBins,
		BinEdges:  make([]float64, nBins+1),
		BinCounts: make([]int64, nBins),
	}

	offset := headerSize

	// Read bin edges (little-endian doubles)
	for i := 0; i < int(nBins+1); i++ {
		bits := binary.LittleEndian.Uint64(data[offset : offset+8])
		hist.BinEdges[i] = math.Float64frombits(bits)
		offset += 8
	}

	// Read bin counts (little-endian int64s)
	for i := 0; i < int(nBins); i++ {
		hist.BinCounts[i] = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Read statistics (little-endian doubles)
	readFloat64 := func() float64 {
		bits := binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8
		return math.Float64frombits(bits)
	}

	hist.MinValue = readFloat64()
	hist.MaxValue = readFloat64()
	hist.MeanValue = readFloat64()
	hist.StdDev = readFloat64()
	hist.Sum = readFloat64()
	hist.NValues = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	hist.NMissing = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))

	return hist, nil
}

// DecodeBinaryHistogramFromReader decodes a binary histogram from an io.Reader
func DecodeBinaryHistogramFromReader(reader io.Reader) (*HistogramData, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, &HistogramError{
			Operation: "read",
			Message:   fmt.Sprintf("failed to read data: %v", err),
		}
	}
	return DecodeBinaryHistogram(data)
}

// GetBinCenter returns the center value of the bin at the given index
func (h *HistogramData) GetBinCenter(binIndex int) (float64, error) {
	if binIndex < 0 || binIndex >= int(h.NBins) {
		return 0, &HistogramError{
			Operation: "getBinCenter",
			Message:   fmt.Sprintf("bin index %d out of range [0, %d)", binIndex, h.NBins),
		}
	}
	return (h.BinEdges[binIndex] + h.BinEdges[binIndex+1]) / 2.0, nil
}

// GetBinWidth returns the width of the bin at the given index
func (h *HistogramData) GetBinWidth(binIndex int) (float64, error) {
	if binIndex < 0 || binIndex >= int(h.NBins) {
		return 0, &HistogramError{
			Operation: "getBinWidth",
			Message:   fmt.Sprintf("bin index %d out of range [0, %d)", binIndex, h.NBins),
		}
	}
	return h.BinEdges[binIndex+1] - h.BinEdges[binIndex], nil
}

// GetNormalizedCount returns the normalized count for the bin at the given index
func (h *HistogramData) GetNormalizedCount(binIndex int) (float64, error) {
	if binIndex < 0 || binIndex >= int(h.NBins) {
		return 0, &HistogramError{
			Operation: "getNormalizedCount",
			Message:   fmt.Sprintf("bin index %d out of range [0, %d)", binIndex, h.NBins),
		}
	}

	totalCount := int64(0)
	for _, count := range h.BinCounts {
		totalCount += count
	}

	if totalCount == 0 {
		return 0, nil
	}

	return float64(h.BinCounts[binIndex]) / float64(totalCount), nil
}

// GetTotalCount returns the sum of all bin counts
func (h *HistogramData) GetTotalCount() int64 {
	total := int64(0)
	for _, count := range h.BinCounts {
		total += count
	}
	return total
}

// String returns a string representation of the histogram
func (h *HistogramData) String() string {
	return fmt.Sprintf("Histogram: Version=%d, Bins=%d, Range=[%.2f, %.2f], Values=%d",
		h.Version, h.NBins, h.MinValue, h.MaxValue, h.NValues)
}

// ToSlideInsightFormat converts the histogram to the format used by SlideInsight services
// This matches the existing domain.SlideHistogram structure
func (h *HistogramData) ToSlideInsightFormat() (map[string]interface{}, error) {
	// Convert int64 counts to int counts for API compatibility
	counts := make([]int, len(h.BinCounts))
	for i, count := range h.BinCounts {
		if count > math.MaxInt32 {
			return nil, &HistogramError{
				Operation: "convert",
				Message:   fmt.Sprintf("count value %d exceeds int32 range", count),
			}
		}
		counts[i] = int(count)
	}

	// Create binary data compatible with existing SlideInsight format
	// This uses the existing pattern of 4-byte little-endian uint32 values
	histogramData := make([]byte, len(counts)*4)
	for i, count := range counts {
		binary.LittleEndian.PutUint32(histogramData[i*4:(i+1)*4], uint32(count))
	}

	return map[string]interface{}{
		"binCount":      int(h.NBins),
		"minValue":      h.MinValue,
		"maxValue":      h.MaxValue,
		"histogramData": histogramData,
		"counts":        counts,
		"metadata": map[string]interface{}{
			"version":   h.Version,
			"meanValue": h.MeanValue,
			"stdDev":    h.StdDev,
			"sum":       h.Sum,
			"nValues":   h.NValues,
			"nMissing":  h.NMissing,
			"binEdges":  h.BinEdges,
		},
	}, nil
}
