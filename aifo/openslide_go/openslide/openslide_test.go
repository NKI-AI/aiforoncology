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

package openslide

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// getTestDataPath returns the path to the test file, handling both Go native and Bazel test environments
func getTestDataPath(t *testing.T, filename string) string {
	t.Helper() // Mark this as a helper function

	// Try the standard Go testdata directory first
	if _, err := os.Stat(filepath.Join("testdata", filename)); err == nil {
		path := filepath.Join("testdata", filename)
		t.Logf("Using test file from Go testdata: %s", path)
		return path
	}

	// Check if we're running in Bazel
	if dir := os.Getenv("TEST_SRCDIR"); dir != "" {
		// Bazel sets TEST_WORKSPACE to the name of the workspace
		workspace := os.Getenv("TEST_WORKSPACE")
		// Look for the file in the runfiles tree
		path := filepath.Join(dir, workspace, "aifo/openslide_go/openslide/testdata", filename)
		if _, err := os.Stat(path); err == nil {
			t.Logf("Using test file from Bazel runfiles: %s", path)
			return path
		}
	}

	// If we can't find it, log a warning and return the regular path
	t.Logf("Warning: Could not locate test data file %s", filename)
	t.Logf("Current working directory: %s", mustGetwd(t))
	t.Skip(fmt.Sprintf("Test data file %s not found - skipping test", filename))
	return "" // We'll never get here because t.Skip() stops test execution
}

// mustGetwd gets the current working directory and fails the test if it errors
func mustGetwd(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return dir
}

// skipIfOpenSlideNotAvailable skips the test if OpenSlide is not available
func skipIfOpenSlideNotAvailable(t *testing.T) {
	t.Helper()
	// Try to get the version as a quick check if OpenSlide is available
	ver := Version()
	if ver == "" {
		t.Skip("OpenSlide library not available - skipping test")
	}
	t.Logf("Using OpenSlide version: %s", ver)
}

const testFilename = "CMU-1.tiff"

func TestDetectVendor(t *testing.T) {
	skipIfOpenSlideNotAvailable(t)

	testPath := getTestDataPath(t, testFilename)
	vendor, err := DetectVendor(testPath)
	if err != nil {
		t.Fatalf("Failed to detect vendor: %v", err)
	}

	if vendor == "" {
		t.Fatal("Vendor is empty but no error was returned")
	}

	t.Logf("Detected vendor: %s", vendor)
}

func TestOpen(t *testing.T) {
	skipIfOpenSlideNotAvailable(t)

	testPath := getTestDataPath(t, testFilename)
	slide, err := Open(testPath)
	if err != nil {
		t.Fatalf("Failed to open slide: %v", err)
	}
	defer slide.Close()

	t.Log("Successfully opened slide")
}

func TestLevels(t *testing.T) {
	skipIfOpenSlideNotAvailable(t)

	testPath := getTestDataPath(t, testFilename)
	slide, err := Open(testPath)
	if err != nil {
		t.Fatalf("Failed to open slide: %v", err)
	}
	defer slide.Close()

	levels, err := slide.LevelCount()
	if err != nil {
		t.Fatalf("Failed to get level count: %v", err)
	}
	if levels <= 0 {
		t.Fatalf("Invalid level count: %d", levels)
	}

	largestDimensions, err := slide.LargestLevelDimensions()
	if err != nil {
		t.Fatalf("Failed to get largest level dimensions: %v", err)
	}
	if largestDimensions[0] <= 0 || largestDimensions[1] <= 0 {
		t.Fatalf("Invalid dimensions for largest level: %v", largestDimensions)
	}

	downsample0, err := slide.LevelDownsample(0)
	if err != nil {
		t.Fatalf("Failed to get level 0 downsample: %v", err)
	}
	t.Logf("Base level 0 (%d, %d): %f",
		largestDimensions[0], largestDimensions[1], downsample0)

	for i := 1; i < levels; i++ {
		levelDimensions, err := slide.LevelDimensions(i)
		if err != nil {
			t.Fatalf("Failed to get dimensions for level %d: %v", i, err)
		}

		downsample, err := slide.LevelDownsample(i)
		if err != nil {
			t.Fatalf("Failed to get downsample for level %d: %v", i, err)
		}

		if levelDimensions[0] <= 0 || levelDimensions[1] <= 0 {
			t.Errorf("Invalid dimensions for level %d: %v", i, levelDimensions)
		}

		if downsample <= 0 {
			t.Errorf("Invalid downsample value for level %d: %f", i, downsample)
		}

		t.Logf("Level %d (%d, %d): %f",
			i, levelDimensions[0], levelDimensions[1], downsample)
	}
}

func TestReadRegion(t *testing.T) {
	skipIfOpenSlideNotAvailable(t)

	testPath := getTestDataPath(t, testFilename)
	slide, err := Open(testPath)
	if err != nil {
		t.Fatalf("Failed to open slide: %v", err)
	}
	defer slide.Close()

	// Read a region from level 6 (should be a reasonable level in most slides)
	x, y := 10, 10
	level := 6
	width, height := 400, 400

	region, err := slide.ReadRegion(x, y, level, width, height)
	if err != nil {
		t.Fatalf("Failed to read region: %v", err)
	}

	// Verify the region dimensions
	bounds := region.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Fatalf("Region dimensions don't match requested size: got %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), width, height)
	}

	// Save the region to a file (only if test is not skipped)
	outputDir := filepath.Join("testdata")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	const testRawFilename = "region.png"
	outputPath := filepath.Join(outputDir, testRawFilename)

	f, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, region); err != nil {
		t.Fatalf("Failed to encode region as PNG: %v", err)
	}

	t.Logf("Saved region image to %s", outputPath)
}

func TestProperties(t *testing.T) {
	skipIfOpenSlideNotAvailable(t)

	testPath := getTestDataPath(t, testFilename)
	slide, err := Open(testPath)
	if err != nil {
		t.Fatalf("Failed to open slide: %v", err)
	}
	defer slide.Close()

	props, err := slide.PropertyNames()
	if err != nil {
		t.Fatalf("Failed to get property names: %v", err)
	}
	if len(props) == 0 {
		t.Error("No properties found in slide")
	}

	t.Logf("Found %d properties", len(props))

	// Check a few common properties
	for _, propName := range []string{PropMPPX, PropMPPY, PropObjectivePower} {
		value, err := slide.PropertyValue(propName)
		if err != nil {
			t.Fatalf("Failed to get property value for %s: %v", propName, err)
		}
		t.Logf("%s = %s", propName, value)
	}

	// Test the Properties() map function
	propMap, err := slide.Properties()
	if err != nil {
		t.Fatalf("Failed to get properties map: %v", err)
	}
	if len(propMap) == 0 {
		t.Error("Properties map is empty")
	}

	if len(propMap) != len(props) {
		t.Errorf("Properties map length doesn't match property names length: %d vs %d",
			len(propMap), len(props))
	}
}

func TestVersion(t *testing.T) {
	ver := Version()
	if ver == "" {
		t.Fatal("Version returned empty string, OpenSlide may not be properly linked")
	}
	t.Logf("OpenSlide version: %s", ver)
}
