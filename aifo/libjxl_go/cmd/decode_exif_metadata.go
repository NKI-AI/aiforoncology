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
// decode_exif_metadata demonstrates extracting EXIF metadata from JPEG XL files
package main

import (
	"fmt"
	"io"
	"os"

	libjxl_go "aifo.dev/aifo/libjxl_go"
)

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

// decodeJpegXlExif extracts EXIF data from a JPEG XL file
func decodeJpegXlExif(jxlData []byte) ([]byte, error) {
	decoder, err := libjxl_go.NewDecoder()
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	// We're only interested in boxes for EXIF extraction
	err = decoder.SubscribeEvents(libjxl_go.EventBox | libjxl_go.EventBoxComplete)
	if err != nil {
		return nil, err
	}

	// Try to enable box decompression (may not be supported)
	supportDecompression := true
	if err := decoder.SetDecompressBoxes(true); err != nil {
		fmt.Printf("NOTE: decompressing brob boxes not supported with the currently used jxl library.\n")
		supportDecompression = false
	}

	// Set input data
	err = decoder.SetInput(jxlData)
	if err != nil {
		return nil, err
	}
	decoder.CloseInput()

	const chunkSize = 65536
	var exifData []byte
	var outputPos int

	for {
		status := decoder.ProcessInput()

		switch status {
		case libjxl_go.DecError:
			return nil, fmt.Errorf("decoder error")

		case libjxl_go.DecNeedMoreInput:
			return nil, fmt.Errorf("error, already provided all input")

		case libjxl_go.DecBox:
			// If we already found EXIF data, release the buffer and return it
			if len(exifData) > 0 {
				remaining := decoder.ReleaseBoxBuffer()
				if remaining > 0 {
					// Trim the unused part of the buffer
					exifData = exifData[:len(exifData)-remaining]
				}
				return exifData, nil
			}

			// Get box type
			boxType, err := decoder.GetBoxType(supportDecompression)
			if err != nil {
				return nil, fmt.Errorf("failed to get box type: %v", err)
			}

			// Check if this is an Exif box
			if string(boxType[:]) == "Exif" {
				// Set up buffer for EXIF data
				exifData = make([]byte, chunkSize)
				err = decoder.SetBoxBuffer(exifData)
				if err != nil {
					return nil, fmt.Errorf("failed to set box buffer: %v", err)
				}
				outputPos = 0
			}

		case libjxl_go.DecBoxNeedMoreOutput:
			if len(exifData) > 0 {
				// Release current buffer and expand it
				remaining := decoder.ReleaseBoxBuffer()
				outputPos += chunkSize - remaining

				// Expand buffer
				newSize := len(exifData) + chunkSize
				newExifData := make([]byte, newSize)
				copy(newExifData, exifData)
				exifData = newExifData

				// Set new buffer starting from current position
				err = decoder.SetBoxBuffer(exifData[outputPos:])
				if err != nil {
					return nil, fmt.Errorf("failed to set expanded box buffer: %v", err)
				}
			}

		case libjxl_go.DecBoxComplete:
			if len(exifData) > 0 {
				// Box is complete, trim to actual size
				remaining := decoder.ReleaseBoxBuffer()
				if remaining > 0 {
					exifData = exifData[:len(exifData)-remaining]
				}
				return exifData, nil
			}
			// Continue to look for more boxes

		case libjxl_go.DecSuccess:
			// Decoding completed, no EXIF found
			return []byte{}, nil

		default:
			return nil, fmt.Errorf("unknown decoder status: %v", status)
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, `Usage: %s <jxl> <exif>
Where:
  jxl = input JPEG XL image filename
  exif = output exif filename
Output files will be overwritten.
`, os.Args[0])
		os.Exit(1)
	}

	jxlFilename := os.Args[1]
	exifFilename := os.Args[2]

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

	// Extract EXIF data
	exifData, err := decodeJpegXlExif(jxlData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while decoding the jxl file: %v\n", err)
		os.Exit(1)
	}

	if len(exifData) == 0 {
		fmt.Printf("No exif data present in this image\n")
	} else {
		// TODO: the exif box data contains the 4-byte TIFF header at the
		// beginning, check whether this is desired to be part of the output, or
		// should be removed.
		err = writeFile(exifFilename, exifData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while writing the exif file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully wrote %s\n", exifFilename)
		fmt.Printf("EXIF data size: %d bytes\n", len(exifData))
	}
}
