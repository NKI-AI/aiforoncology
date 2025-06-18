// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.

// Place sample TIFFs in ./testdata/:
//   - testdata/sample8.tif      (8-bit, single-channel, tiled TIFF)
//   - testdata/sample16.tif     (16-bit, single-channel, tiled TIFF)
//   - testdata/pyramid.tif      (pyramidal, tiled TIFF)
package tiff

import (
	"os"
	"path/filepath"
	"testing"
)

func dataFile(name string) string {
	return filepath.Join("testdata", name)
}

func skipIfMissing(t *testing.T, name string) {
	path := dataFile(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("skipping: %s not found", path)
	}
}

func TestOpenClose(t *testing.T) {
	skipIfMissing(t, "sample8.tif")
	tif, err := Open(dataFile("sample8.tif"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	tif.Close()
}

func TestLevelCount(t *testing.T) {
	skipIfMissing(t, "pyramid.tif")
	tif, err := Open(dataFile("pyramid.tif"))
	if err != nil {
		t.Fatalf("Open pyramid failed: %v", err)
	}
	defer tif.Close()
	levels := tif.LevelCount()
	if levels < 1 {
		t.Errorf("expected at least 1 level, got %d", levels)
	}
}

func TestLevelSizeAndInfo(t *testing.T) {
	skipIfMissing(t, "pyramid.tif")
	tif, err := Open(dataFile("pyramid.tif"))
	if err != nil {
		t.Fatalf("Open pyramid failed: %v", err)
	}
	defer tif.Close()

	for lvl := 0; lvl < tif.LevelCount(); lvl++ {
		w, h, err := tif.LevelSize(lvl)
		if err != nil {
			t.Errorf("LevelSize(%d) error: %v", lvl, err)
			continue
		}
		if w == 0 || h == 0 {
			t.Errorf("LevelSize(%d) returned zero dimensions: %dx%d", lvl, w, h)
		}
		info, err := tif.LevelInfo(lvl)
		if err != nil {
			t.Errorf("LevelInfo(%d) error: %v", lvl, err)
		}
		if got := info; len(got) == 0 {
			t.Errorf("LevelInfo(%d) returned empty string", lvl)
		}
	}
}

func TestBaseResolution(t *testing.T) {
	skipIfMissing(t, "sample8.tif")
	tif, err := Open(dataFile("sample8.tif"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer tif.Close()

	xµm, yµm, err := tif.BaseResolution()
	if err != nil {
		t.Logf("BaseResolution not available: %v", err)
	} else {
		if xµm <= 0 || yµm <= 0 {
			t.Errorf("BaseResolution returned non-positive values: %.3f×%.3f", xµm, yµm)
		}
	}
}

func TestReadRegion8(t *testing.T) {
	skipIfMissing(t, "sample8.tif")
	tif, err := Open(dataFile("sample8.tif"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer tif.Close()

	// read a small 10×10 region at (0,0)
	data, err := tif.ReadRegion(0, 0, 0, 10, 10)
	if err != nil {
		t.Fatalf("ReadRegion failed: %v", err)
	}
	buf, ok := data.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", data)
	}
	if len(buf) != 100 {
		t.Errorf("expected buffer length 100, got %d", len(buf))
	}
}

func TestReadRegion16(t *testing.T) {
	skipIfMissing(t, "sample16.tif")
	tif, err := Open(dataFile("sample16.tif"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer tif.Close()

	data, err := tif.ReadRegion(0, 0, 0, 8, 8)
	if err != nil {
		t.Fatalf("ReadRegion failed: %v", err)
	}
	buf, ok := data.([]uint16)
	if !ok {
		t.Fatalf("expected []uint16, got %T", data)
	}
	if len(buf) != 64 {
		t.Errorf("expected buffer length 64, got %d", len(buf))
	}
}

func TestReadRegionInvalidLevel(t *testing.T) {
	skipIfMissing(t, "sample8.tif")
	tif, err := Open(dataFile("sample8.tif"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer tif.Close()

	_, err = tif.ReadRegion(999, 0, 0, 1, 1)
	if err == nil {
		t.Error("expected error for invalid level, got nil")
	}
}
