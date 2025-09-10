package dicomreader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/frame"
)

const (
	dicomMagicWord  = "DICM"
	dicomHeaderSize = 132
)

func ReadDICOMMetadata(filePath string) (*dicom.Dataset, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Get the file size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Check if the file is too small to be a valid DICOM file
	if info.Size() < dicomHeaderSize {
		return nil, fmt.Errorf("file too small")
	}

	// Create a buffer to read the first 132 bytes (for checking the DICOM magic word)
	header := make([]byte, dicomHeaderSize)
	_, err = io.ReadFull(file, header)
	if err != nil {
		log.Printf("ERROR: Failed to read file header: %v", err)
		return nil, err
	}

	// Check if bytes 128-132 contain the DICOM magic word
	if !bytes.Equal(header[128:], []byte(dicomMagicWord)) {
		log.Printf("ERROR: File %s is not a valid DICOM file", filePath)
		return nil, fmt.Errorf("not a valid DICOM file")
	}

	// Now combine the header and the rest of the file into a single reader
	remainingFile := io.MultiReader(bytes.NewReader(header), file)

	// Now we can proceed to parse the DICOM file with streaming
	frameChan := make(chan *frame.Frame, 100) // Buffer size for frame streaming

	// Start a goroutine to process frames as they are streamed
	go func() {
		// for fr := range frameChan {
		// 	fmt.Printf("Received frame: %+v\n", fr)
		// }
	}()

	// Parse the file using streaming, passing the correct file size and skipping pixel data
	dataset, err := dicom.Parse(remainingFile, info.Size(), frameChan, dicom.SkipPixelData())
	if err != nil {
		log.Printf("ERROR: Error parsing DICOM file: %v", err)
		return nil, err
	}

	return &dataset, nil
}
