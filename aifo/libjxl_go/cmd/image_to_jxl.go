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
// image_to_jxl demonstrates converting PNG and JPG images to JPEG XL format
package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	libjxl_go "aifo.dev/aifo/libjxl_go"
)

// loadImage loads an image from file, supporting PNG and JPG formats
func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", filename, err)
	}
	defer file.Close()

	// Read the first few bytes to detect format
	header := make([]byte, 8)
	_, err = io.ReadFull(file, header)
	if err != nil {
		return nil, fmt.Errorf("could not read file header: %v", err)
	}

	// Reset file position
	file.Seek(0, 0)

	// Detect format and decode
	if len(header) >= 8 && string(header[:8]) == "\x89PNG\r\n\x1a\n" {
		// PNG format
		img, err := png.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("could not decode PNG: %v", err)
		}
		return img, nil
	} else if len(header) >= 2 && string(header[:2]) == "\xff\xd8" {
		// JPEG format
		img, err := jpeg.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("could not decode JPEG: %v", err)
		}
		return img, nil
	}

	return nil, fmt.Errorf("unsupported image format (only PNG and JPG are supported)")
}

// imageToRGBA converts any image to RGBA format
func imageToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba
}

// rgbaToBytes converts RGBA image to byte slice
func rgbaToBytes(img *image.RGBA) []byte {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert RGBA to bytes (RGBA format)
	pixels := make([]byte, width*height*4)
	idx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert from uint32 (0-65535) to uint8 (0-255)
			pixels[idx] = uint8(r >> 8)
			pixels[idx+1] = uint8(g >> 8)
			pixels[idx+2] = uint8(b >> 8)
			pixels[idx+3] = uint8(a >> 8)
			idx += 4
		}
	}

	return pixels
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

// getOutputFilename generates output filename based on input and quality
func getOutputFilename(inputPath string, quality float32) string {
	dir := filepath.Dir(inputPath)
	name := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	// Add quality suffix if not lossless
	if quality < 100 {
		return filepath.Join(dir, fmt.Sprintf("%s_q%.0f.jxl", name, quality))
	}
	return filepath.Join(dir, fmt.Sprintf("%s_lossless.jxl", name))
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 4 {
		fmt.Fprintf(os.Stderr, `Usage: %s <input> [output] [quality]
Where:
  input   = input image filename (PNG or JPG)
  output  = output JPEG XL filename (optional, auto-generated if not provided)
  quality = compression quality 0-100 (optional, default: 90)
Examples:
  %s image.png
  %s image.jpg output.jxl
  %s image.png output.jxl 95
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	var outputFilename string
	var quality float32 = 90.0

	if len(os.Args) == 2 {
		// Only input provided, auto-generate output filename
		outputFilename = getOutputFilename(inputFilename, quality)
	} else if len(os.Args) == 3 {
		// Input and output provided
		outputFilename = os.Args[2]
	} else {
		// Input, output, and quality provided
		outputFilename = os.Args[2]
		if _, err := fmt.Sscanf(os.Args[3], "%f", &quality); err != nil {
			fmt.Fprintf(os.Stderr, "invalid quality value: %s (must be 0-100)\n", os.Args[3])
			os.Exit(1)
		}
		if quality < 0 || quality > 100 {
			fmt.Fprintf(os.Stderr, "quality must be between 0 and 100\n")
			os.Exit(1)
		}
	}

	// Load and convert image
	fmt.Printf("Loading image: %s\n", inputFilename)
	img, err := loadImage(inputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load image: %v\n", err)
		os.Exit(1)
	}

	// Convert to RGBA
	rgba := imageToRGBA(img)
	width := uint32(rgba.Bounds().Dx())
	height := uint32(rgba.Bounds().Dy())

	fmt.Printf("Image size: %dx%d pixels\n", width, height)
	fmt.Printf("Quality: %.1f%%\n", quality)

	// Convert to bytes
	pixels := rgbaToBytes(rgba)

	// Set up pixel format for RGBA
	format := libjxl_go.PixelFormat{
		NumChannels: 4, // RGBA
		DataType:    libjxl_go.TypeUint8,
		Endianness:  libjxl_go.NativeEndian,
		Align:       0,
	}

	// Encode to JXL
	fmt.Printf("Encoding to JPEG XL...\n")
	compressed, err := libjxl_go.EncodeOneShot(pixels, width, height, format, quality)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not encode to JXL: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	fmt.Printf("Writing output: %s\n", outputFilename)
	err = writeFile(outputFilename, compressed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write output file: %v\n", err)
		os.Exit(1)
	}

	// Calculate compression ratio
	inputSize := len(pixels)
	outputSize := len(compressed)
	compressionRatio := float64(inputSize) / float64(outputSize)

	fmt.Printf("Compression complete!\n")
	fmt.Printf("Input size:  %d bytes\n", inputSize)
	fmt.Printf("Output size: %d bytes\n", outputSize)
	fmt.Printf("Compression ratio: %.2f:1\n", compressionRatio)
}
