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
// decode_oneshot demonstrates basic JPEG XL decoding
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unsafe"

	libjxl_go "aifo.dev/aifo/libjxl_go"
)

// writePFM writes pixels to a Portable FloatMap file
func writePFM(filename string, pixels []float32, width, height uint32) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", filename, err)
	}
	defer file.Close()

	// Write PFM header
	_, err = fmt.Fprintf(file, "PF\n%d %d\n-1.0\n", width, height)
	if err != nil {
		return err
	}

	// Write pixel data (RGB only, bottom to top)
	for y := int(height) - 1; y >= 0; y-- {
		for x := uint32(0); x < width; x++ {
			for c := 0; c < 3; c++ {
				pixelIdx := (uint32(y)*width+x)*4 + uint32(c) // RGBA format
				err = binary.Write(file, binary.LittleEndian, pixels[pixelIdx])
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// loadFile loads a file into memory
func loadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// writeFile writes data to a file
func writeFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", filename, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, `Usage: %s <jxl> <pfm> <icc>
Where:
  jxl = input JPEG XL image filename
  pfm = output Portable FloatMap image filename
  icc = output ICC color profile filename
Output files will be overwritten.
`, os.Args[0])
		os.Exit(1)
	}

	jxlFilename := os.Args[1]
	pfmFilename := os.Args[2]
	iccFilename := os.Args[3]

	// Load JXL file
	jxlData, err := loadFile(jxlFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't load %s: %v\n", jxlFilename, err)
		os.Exit(1)
	}

	// Check if it's a valid JXL file
	isValid, err := libjxl_go.CheckSignature(jxlData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error checking signature: %v\n", err)
		os.Exit(1)
	}
	if !isValid {
		fmt.Fprintf(os.Stderr, "not a valid JPEG XL file\n")
		os.Exit(1)
	}

	// Set up pixel format for RGBA float32
	format := libjxl_go.PixelFormat{
		NumChannels: 4,
		DataType:    libjxl_go.TypeFloat,
		Endianness:  libjxl_go.NativeEndian,
		Align:       0,
	}

	// Decode the image
	pixelBytes, info, iccProfile, err := libjxl_go.DecodeOneShot(jxlData, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding JXL file: %v\n", err)
		os.Exit(1)
	}

	// Convert byte slice to float32 slice
	numPixels := info.XSize * info.YSize * 4
	pixels := make([]float32, numPixels)
	for i := 0; i < int(numPixels); i++ {
		// Convert 4 bytes to float32
		bits := binary.LittleEndian.Uint32(pixelBytes[i*4 : (i+1)*4])
		pixels[i] = *(*float32)(unsafe.Pointer(&bits))
	}

	// Write PFM file
	err = writePFM(pfmFilename, pixels, info.XSize, info.YSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing PFM file: %v\n", err)
		os.Exit(1)
	}

	// Write ICC profile
	err = writeFile(iccFilename, iccProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing ICC profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote %s and %s\n", pfmFilename, iccFilename)
	fmt.Printf("Image dimensions: %dx%d\n", info.XSize, info.YSize)
	fmt.Printf("Bits per sample: %d\n", info.BitsPerSample)
	fmt.Printf("ICC profile size: %d bytes\n", len(iccProfile))
}
