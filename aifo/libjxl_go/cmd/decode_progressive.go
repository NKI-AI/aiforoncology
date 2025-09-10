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
// decode_progressive demonstrates progressive JPEG XL decoding
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	libjxl_go "aifo.dev/aifo/libjxl_go"
)

// writePAM writes RGBA pixels to a Portable Arbitrary Map file
func writePAM(filename string, pixels []byte, width, height uint32) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", filename, err)
	}
	defer file.Close()

	// Write PAM header
	_, err = fmt.Fprintf(file, "P7\nWIDTH %d\nHEIGHT %d\nDEPTH 4\nMAXVAL 255\nTUPLTYPE RGB_ALPHA\nENDHDR\n", width, height)
	if err != nil {
		return err
	}

	// Write pixel data
	_, err = file.Write(pixels)
	return err
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

// decodeJpegXlProgressive decodes JPEG XL progressively
func decodeJpegXlProgressive(jxlData []byte, baseFilename string, chunkSize int) error {
	decoder, err := libjxl_go.NewDecoder()
	if err != nil {
		return err
	}
	defer decoder.Close()

	// Subscribe to required events
	err = decoder.SubscribeEvents(libjxl_go.EventBasicInfo | libjxl_go.EventColorEncoding | libjxl_go.EventFullImage)
	if err != nil {
		return err
	}

	// Set up pixel format for RGBA uint8
	format := libjxl_go.PixelFormat{
		NumChannels: 4,
		DataType:    libjxl_go.TypeUint8,
		Endianness:  libjxl_go.NativeEndian,
		Align:       0,
	}

	var info libjxl_go.BasicInfo
	var pixelBuffer []byte
	var xsize, ysize uint32

	seen := 0
	totalSize := len(jxlData)
	remaining := chunkSize
	if remaining > totalSize {
		remaining = totalSize
	}

	// Set initial input
	err = decoder.SetInput(jxlData[:remaining])
	if err != nil {
		return err
	}

	for {
		status := decoder.ProcessInput()

		switch status {
		case libjxl_go.DecError:
			return fmt.Errorf("decoder error")

		case libjxl_go.DecNeedMoreInput, libjxl_go.DecSuccess, libjxl_go.DecFullImage:
			// Release consumed input and get remaining bytes
			unconsumed := decoder.ReleaseInput()
			seen += remaining - unconsumed
			fmt.Printf("Flushing after %d bytes\n", seen)

			// Try to flush the image (might not work if no preview yet)
			if status == libjxl_go.DecNeedMoreInput {
				if err := decoder.FlushImage(); err != nil {
					fmt.Printf("flush error (no preview yet)\n")
				} else if len(pixelBuffer) > 0 {
					// Write intermediate result
					filename := fmt.Sprintf("%s-%d.pam", baseFilename, seen)
					if err := writePAM(filename, pixelBuffer, xsize, ysize); err != nil {
						fmt.Printf("Error writing progressive output: %v\n", err)
					}
				}
			} else if len(pixelBuffer) > 0 {
				// Write final or intermediate result
				filename := fmt.Sprintf("%s-%d.pam", baseFilename, seen)
				if err := writePAM(filename, pixelBuffer, xsize, ysize); err != nil {
					fmt.Printf("Error writing progressive output: %v\n", err)
				}
			}

			// Calculate remaining data
			remaining = totalSize - seen
			if remaining > chunkSize {
				remaining = chunkSize
			}

			if remaining == 0 {
				if status == libjxl_go.DecNeedMoreInput {
					return fmt.Errorf("error, already provided all input")
				} else {
					return nil // Done
				}
			}

			// Set next chunk of input
			err = decoder.SetInput(jxlData[seen : seen+remaining])
			if err != nil {
				return err
			}

		case libjxl_go.DecBasicInfo:
			info, err = decoder.GetBasicInfo()
			if err != nil {
				return fmt.Errorf("failed to get basic info: %v", err)
			}
			xsize = info.XSize
			ysize = info.YSize

			// Set suggested thread count
			numThreads := decoder.SuggestThreads(info.XSize, info.YSize)
			decoder.SetThreads(numThreads)

		case libjxl_go.DecColorEncoding:
			// We could get the ICC profile here if needed
			// _, err = decoder.GetICCProfile(libjxl_go.ColorProfileTargetOriginal)

		case libjxl_go.DecNeedImageOutBuffer:
			bufferSize, err := decoder.ImageOutBufferSize(format)
			if err != nil {
				return fmt.Errorf("failed to get buffer size: %v", err)
			}

			expectedSize := int(xsize * ysize * 4) // RGBA
			if bufferSize != expectedSize {
				return fmt.Errorf("invalid buffer size %d != %d", bufferSize, expectedSize)
			}

			pixelBuffer = make([]byte, bufferSize)
			err = decoder.SetImageOutBuffer(format, pixelBuffer)
			if err != nil {
				return fmt.Errorf("failed to set image out buffer: %v", err)
			}

		default:
			return fmt.Errorf("unknown decoder status: %v", status)
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, `Usage: %s <jxl> <basename> [chunksize]
Where:
  jxl = input JPEG XL image filename
  basename = prefix of output filenames
  chunksize = loads chunksize bytes at a time and writes
              intermediate results to basename-[bytes loaded].pam
Output files will be overwritten.
`, os.Args[0])
		os.Exit(1)
	}

	jxlFilename := os.Args[1]
	baseFilename := os.Args[2]

	// Load JXL file
	jxlData, err := loadFile(jxlFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't load %s: %v\n", jxlFilename, err)
		os.Exit(1)
	}

	// Set chunk size
	chunkSize := len(jxlData) // Default to full file
	if len(os.Args) > 3 {
		if cs, err := strconv.Atoi(os.Args[3]); err == nil {
			if cs < 100 {
				fmt.Fprintf(os.Stderr, "Chunk size is too low, try at least 100 bytes\n")
				os.Exit(1)
			}
			chunkSize = cs
		} else {
			fmt.Fprintf(os.Stderr, "Invalid chunk size: %v\n", err)
			os.Exit(1)
		}
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

	// Decode progressively
	err = decodeJpegXlProgressive(jxlData, baseFilename, chunkSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while decoding JXL file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully completed progressive decoding\n")
}
