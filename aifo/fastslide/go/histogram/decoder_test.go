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
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestDecodeBinaryHistogram(t *testing.T) {
	// Create test data that matches the C++ binary format
	nBins := uint32(5)
	binEdges := []float64{0.0, 1.0, 2.0, 3.0, 4.0, 5.0}
	binCounts := []int64{10, 20, 30, 25, 15}

	var buf bytes.Buffer

	// Write magic header
	buf.WriteString("HIST")

	// Write version
	binary.Write(&buf, binary.LittleEndian, uint32(1))

	// Write number of bins
	binary.Write(&buf, binary.LittleEndian, nBins)

	// Write bin edges
	for _, edge := range binEdges {
		binary.Write(&buf, binary.LittleEndian, math.Float64bits(edge))
	}

	// Write bin counts
	for _, count := range binCounts {
		binary.Write(&buf, binary.LittleEndian, uint64(count))
	}

	// Write statistics (5 doubles + 2 int64s)
	binary.Write(&buf, binary.LittleEndian, math.Float64bits(0.0))   // min
	binary.Write(&buf, binary.LittleEndian, math.Float64bits(5.0))   // max
	binary.Write(&buf, binary.LittleEndian, math.Float64bits(2.5))   // mean
	binary.Write(&buf, binary.LittleEndian, math.Float64bits(1.5))   // stddev
	binary.Write(&buf, binary.LittleEndian, math.Float64bits(250.0)) // sum
	binary.Write(&buf, binary.LittleEndian, uint64(100))             // nValues (int64)
	binary.Write(&buf, binary.LittleEndian, uint64(0))               // nMissing (int64)

	// Test decoding
	hist, err := DecodeBinaryHistogram(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to decode histogram: %v", err)
	}

	// Verify header
	if hist.Version != 1 {
		t.Errorf("Expected version 1, got %d", hist.Version)
	}

	if hist.NBins != nBins {
		t.Errorf("Expected %d bins, got %d", nBins, hist.NBins)
	}

	// Verify bin edges
	for i, expected := range binEdges {
		if hist.BinEdges[i] != expected {
			t.Errorf("Bin edge %d: expected %f, got %f", i, expected, hist.BinEdges[i])
		}
	}

	// Verify bin counts
	for i, expected := range binCounts {
		if hist.BinCounts[i] != expected {
			t.Errorf("Bin count %d: expected %d, got %d", i, expected, hist.BinCounts[i])
		}
	}

	// Verify statistics
	if hist.MinValue != 0.0 {
		t.Errorf("Expected min value 0.0, got %f", hist.MinValue)
	}

	if hist.MaxValue != 5.0 {
		t.Errorf("Expected max value 5.0, got %f", hist.MaxValue)
	}

	if hist.NValues != 100 {
		t.Errorf("Expected nValues 100, got %d", hist.NValues)
	}
}

func TestGetBinCenter(t *testing.T) {
	hist := &HistogramData{
		NBins:    3,
		BinEdges: []float64{0.0, 1.0, 2.0, 3.0},
	}

	center, err := hist.GetBinCenter(1)
	if err != nil {
		t.Fatalf("Failed to get bin center: %v", err)
	}

	expected := 1.5
	if center != expected {
		t.Errorf("Expected bin center %f, got %f", expected, center)
	}
}

func TestGetBinWidth(t *testing.T) {
	hist := &HistogramData{
		NBins:    3,
		BinEdges: []float64{0.0, 1.0, 3.0, 6.0},
	}

	width, err := hist.GetBinWidth(1)
	if err != nil {
		t.Fatalf("Failed to get bin width: %v", err)
	}

	expected := 2.0
	if width != expected {
		t.Errorf("Expected bin width %f, got %f", expected, width)
	}
}

func TestGetNormalizedCount(t *testing.T) {
	hist := &HistogramData{
		NBins:     3,
		BinCounts: []int64{10, 20, 30},
	}

	normalized, err := hist.GetNormalizedCount(1)
	if err != nil {
		t.Fatalf("Failed to get normalized count: %v", err)
	}

	expected := 20.0 / 60.0 // 20 out of total 60
	if math.Abs(normalized-expected) > 1e-10 {
		t.Errorf("Expected normalized count %f, got %f", expected, normalized)
	}
}

func TestGetTotalCount(t *testing.T) {
	hist := &HistogramData{
		NBins:     3,
		BinCounts: []int64{10, 20, 30},
	}

	total := hist.GetTotalCount()
	expected := int64(60)
	if total != expected {
		t.Errorf("Expected total count %d, got %d", expected, total)
	}
}

func TestToSlideInsightFormat(t *testing.T) {
	hist := &HistogramData{
		NBins:     3,
		BinCounts: []int64{10, 20, 30},
		BinEdges:  []float64{0.0, 1.0, 2.0, 3.0},
		MinValue:  0.0,
		MaxValue:  3.0,
		MeanValue: 1.5,
		StdDev:    0.5,
		Sum:       100.0,
		NValues:   60,
		NMissing:  0,
		Version:   1,
	}

	format, err := hist.ToSlideInsightFormat()
	if err != nil {
		t.Fatalf("Failed to convert to SlideInsight format: %v", err)
	}

	// Check basic fields
	if format["binCount"] != 3 {
		t.Errorf("Expected binCount 3, got %v", format["binCount"])
	}

	if format["minValue"] != 0.0 {
		t.Errorf("Expected minValue 0.0, got %v", format["minValue"])
	}

	if format["maxValue"] != 3.0 {
		t.Errorf("Expected maxValue 3.0, got %v", format["maxValue"])
	}

	// Check counts slice
	counts, ok := format["counts"].([]int)
	if !ok {
		t.Fatalf("Counts is not []int")
	}
	if len(counts) != 3 {
		t.Errorf("Expected 3 counts, got %d", len(counts))
	}
	if counts[1] != 20 {
		t.Errorf("Expected count[1] = 20, got %d", counts[1])
	}

	// Check metadata
	metadata, ok := format["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("Metadata is not map[string]interface{}")
	}
	if metadata["version"] != uint32(1) {
		t.Errorf("Expected metadata version 1, got %v", metadata["version"])
	}
}

func TestInvalidMagicHeader(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("ABCD") // Wrong magic

	_, err := DecodeBinaryHistogram(buf.Bytes())
	if err == nil {
		t.Error("Expected error for invalid magic header")
	}

	histErr, ok := err.(*HistogramError)
	if !ok {
		t.Errorf("Expected HistogramError, got %T", err)
	}

	if histErr.Operation != "decode" {
		t.Errorf("Expected operation 'decode', got '%s'", histErr.Operation)
	}
}

func TestInvalidVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("HIST")
	binary.Write(&buf, binary.LittleEndian, uint32(99)) // Unsupported version

	_, err := DecodeBinaryHistogram(buf.Bytes())
	if err == nil {
		t.Error("Expected error for unsupported version")
	}
}

func TestDataTooShort(t *testing.T) {
	data := []byte("HIST") // Only magic header

	_, err := DecodeBinaryHistogram(data)
	if err == nil {
		t.Error("Expected error for data too short")
	}
}

func TestBoundaryChecks(t *testing.T) {
	hist := &HistogramData{
		NBins:     3,
		BinEdges:  []float64{0.0, 1.0, 2.0, 3.0},
		BinCounts: []int64{10, 20, 30},
	}

	// Test out of bounds bin access
	_, err := hist.GetBinCenter(-1)
	if err == nil {
		t.Error("Expected error for negative bin index")
	}

	_, err = hist.GetBinCenter(3)
	if err == nil {
		t.Error("Expected error for bin index >= nBins")
	}

	_, err = hist.GetBinWidth(-1)
	if err == nil {
		t.Error("Expected error for negative bin index")
	}

	_, err = hist.GetNormalizedCount(5)
	if err == nil {
		t.Error("Expected error for bin index >= nBins")
	}
}

func TestString(t *testing.T) {
	hist := &HistogramData{
		Version:  1,
		NBins:    100,
		MinValue: 0.5,
		MaxValue: 255.7,
		NValues:  12345,
	}

	s := hist.String()
	expected := "Histogram: Version=1, Bins=100, Range=[0.50, 255.70], Values=12345"
	if s != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, s)
	}
}
