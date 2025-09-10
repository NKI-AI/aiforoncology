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

/*
#include <jxl/decode.h>
#include <jxl/resizable_parallel_runner.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// DecoderStatus represents the decoder status
type DecoderStatus int

const (
	DecSuccess              DecoderStatus = C.JXL_DEC_SUCCESS
	DecError                DecoderStatus = C.JXL_DEC_ERROR
	DecNeedMoreInput        DecoderStatus = C.JXL_DEC_NEED_MORE_INPUT
	DecNeedPreviewOutBuffer DecoderStatus = C.JXL_DEC_NEED_PREVIEW_OUT_BUFFER
	DecNeedImageOutBuffer   DecoderStatus = C.JXL_DEC_NEED_IMAGE_OUT_BUFFER
	DecJpegNeedMoreOutput   DecoderStatus = C.JXL_DEC_JPEG_NEED_MORE_OUTPUT
	DecBoxNeedMoreOutput    DecoderStatus = C.JXL_DEC_BOX_NEED_MORE_OUTPUT
	DecBasicInfo            DecoderStatus = C.JXL_DEC_BASIC_INFO
	DecColorEncoding        DecoderStatus = C.JXL_DEC_COLOR_ENCODING
	DecPreviewImage         DecoderStatus = C.JXL_DEC_PREVIEW_IMAGE
	DecFrame                DecoderStatus = C.JXL_DEC_FRAME
	DecFullImage            DecoderStatus = C.JXL_DEC_FULL_IMAGE
	DecJpegReconstruction   DecoderStatus = C.JXL_DEC_JPEG_RECONSTRUCTION
	DecBox                  DecoderStatus = C.JXL_DEC_BOX
	DecFrameProgression     DecoderStatus = C.JXL_DEC_FRAME_PROGRESSION
	DecBoxComplete          DecoderStatus = C.JXL_DEC_BOX_COMPLETE
)

// DecoderEventFlags for subscribing to events
const (
	EventBasicInfo          = C.JXL_DEC_BASIC_INFO
	EventColorEncoding      = C.JXL_DEC_COLOR_ENCODING
	EventPreviewImage       = C.JXL_DEC_PREVIEW_IMAGE
	EventFrame              = C.JXL_DEC_FRAME
	EventFullImage          = C.JXL_DEC_FULL_IMAGE
	EventJpegReconstruction = C.JXL_DEC_JPEG_RECONSTRUCTION
	EventBox                = C.JXL_DEC_BOX
	EventFrameProgression   = C.JXL_DEC_FRAME_PROGRESSION
	EventBoxComplete        = C.JXL_DEC_BOX_COMPLETE
)

// ColorProfileTarget specifies which color profile to get
type ColorProfileTarget int

const (
	ColorProfileTargetOriginal ColorProfileTarget = C.JXL_COLOR_PROFILE_TARGET_ORIGINAL
	ColorProfileTargetData     ColorProfileTarget = C.JXL_COLOR_PROFILE_TARGET_DATA
)

// Decoder represents a JPEG XL decoder
type Decoder struct {
	dec    *C.JxlDecoder
	runner unsafe.Pointer // void*
}

// NewDecoder creates a new JPEG XL decoder
func NewDecoder() (*Decoder, error) {
	dec := C.JxlDecoderCreate(nil)
	if dec == nil {
		return nil, ErrDecoderFailed
	}

	runner := C.JxlResizableParallelRunnerCreate(nil)
	if runner == nil {
		C.JxlDecoderDestroy(dec)
		return nil, ErrOutOfMemory
	}

	if C.JxlDecoderSetParallelRunner(dec, (*[0]byte)(C.JxlResizableParallelRunner), runner) != C.JXL_DEC_SUCCESS {
		C.JxlResizableParallelRunnerDestroy(runner)
		C.JxlDecoderDestroy(dec)
		return nil, ErrDecoderFailed
	}

	d := &Decoder{
		dec:    dec,
		runner: runner,
	}

	setFinalizer(d, (*Decoder).destroy)
	return d, nil
}

// destroy cleans up the decoder resources
func (d *Decoder) destroy() {
	if d.runner != nil {
		C.JxlResizableParallelRunnerDestroy(d.runner)
		d.runner = nil
	}
	if d.dec != nil {
		C.JxlDecoderDestroy(d.dec)
		d.dec = nil
	}
}

// Close manually cleans up the decoder resources
func (d *Decoder) Close() {
	d.destroy()
	runtime.SetFinalizer(d, nil)
}

// Reset resets the decoder to its initial state
func (d *Decoder) Reset() {
	C.JxlDecoderReset(d.dec)
}

// SubscribeEvents subscribes to decoder events
func (d *Decoder) SubscribeEvents(events int) error {
	if C.JxlDecoderSubscribeEvents(d.dec, C.int(events)) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// SetInput sets the input data for the decoder
func (d *Decoder) SetInput(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidInput
	}
	cData := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	if C.JxlDecoderSetInput(d.dec, cData, C.size_t(len(data))) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// CloseInput signals that no more input will be provided
func (d *Decoder) CloseInput() {
	C.JxlDecoderCloseInput(d.dec)
}

// ProcessInput processes the input and returns the decoder status
func (d *Decoder) ProcessInput() DecoderStatus {
	return DecoderStatus(C.JxlDecoderProcessInput(d.dec))
}

// GetBasicInfo gets basic image information
func (d *Decoder) GetBasicInfo() (BasicInfo, error) {
	var cInfo C.JxlBasicInfo
	if C.JxlDecoderGetBasicInfo(d.dec, &cInfo) != C.JXL_DEC_SUCCESS {
		return BasicInfo{}, ErrDecoderFailed
	}
	return basicInfoFromC(cInfo), nil
}

// GetColorEncoding gets the color encoding information
func (d *Decoder) GetColorEncoding(target ColorProfileTarget) (ColorEncoding, error) {
	var cEncoding C.JxlColorEncoding
	if C.JxlDecoderGetColorAsEncodedProfile(d.dec, C.JxlColorProfileTarget(target), &cEncoding) != C.JXL_DEC_SUCCESS {
		return ColorEncoding{}, ErrDecoderFailed
	}

	return ColorEncoding{
		ColorSpace:       int(cEncoding.color_space),
		WhitePoint:       int(cEncoding.white_point),
		Primaries:        int(cEncoding.primaries),
		TransferFunction: int(cEncoding.transfer_function),
		RenderingIntent:  int(cEncoding.rendering_intent),
		Gamma:            float64(cEncoding.gamma),
	}, nil
}

// GetICCProfileSize gets the size of the ICC profile
func (d *Decoder) GetICCProfileSize(target ColorProfileTarget) (int, error) {
	var size C.size_t
	if C.JxlDecoderGetICCProfileSize(d.dec, C.JxlColorProfileTarget(target), &size) != C.JXL_DEC_SUCCESS {
		return 0, ErrDecoderFailed
	}
	return int(size), nil
}

// GetICCProfile gets the ICC profile data
func (d *Decoder) GetICCProfile(target ColorProfileTarget) ([]byte, error) {
	size, err := d.GetICCProfileSize(target)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return []byte{}, nil
	}

	profile := make([]byte, size)
	cProfile := (*C.uint8_t)(unsafe.Pointer(&profile[0]))
	if C.JxlDecoderGetColorAsICCProfile(d.dec, C.JxlColorProfileTarget(target), cProfile, C.size_t(size)) != C.JXL_DEC_SUCCESS {
		return nil, ErrDecoderFailed
	}

	return profile, nil
}

// ImageOutBufferSize calculates the required buffer size for image output
func (d *Decoder) ImageOutBufferSize(format PixelFormat) (int, error) {
	cFormat := format.toCPixelFormat()
	var size C.size_t
	if C.JxlDecoderImageOutBufferSize(d.dec, &cFormat, &size) != C.JXL_DEC_SUCCESS {
		return 0, ErrDecoderFailed
	}
	return int(size), nil
}

// SetImageOutBuffer sets the output buffer for image data
func (d *Decoder) SetImageOutBuffer(format PixelFormat, buffer []byte) error {
	if len(buffer) == 0 {
		return ErrInvalidInput
	}

	cFormat := format.toCPixelFormat()
	cBuffer := unsafe.Pointer(&buffer[0])
	if C.JxlDecoderSetImageOutBuffer(d.dec, &cFormat, cBuffer, C.size_t(len(buffer))) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// GetFrameHeader gets frame header information
func (d *Decoder) GetFrameHeader() (FrameHeader, error) {
	var cHeader C.JxlFrameHeader
	if C.JxlDecoderGetFrameHeader(d.dec, &cHeader) != C.JXL_DEC_SUCCESS {
		return FrameHeader{}, ErrDecoderFailed
	}

	return FrameHeader{
		Duration:   uint32(cHeader.duration),
		Timecode:   uint32(cHeader.timecode),
		NameLength: uint32(cHeader.name_length),
		IsLast:     cHeader.is_last != 0,
	}, nil
}

// FlushImage flushes the current partially decoded image
func (d *Decoder) FlushImage() error {
	if C.JxlDecoderFlushImage(d.dec) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// ReleaseInput releases unconsumed input and returns the number of remaining bytes
func (d *Decoder) ReleaseInput() int {
	return int(C.JxlDecoderReleaseInput(d.dec))
}

// SetThreads sets the number of threads for parallel processing
func (d *Decoder) SetThreads(numThreads int) {
	C.JxlResizableParallelRunnerSetThreads(d.runner, C.size_t(numThreads))
}

// SuggestThreads returns the suggested number of threads for the given image dimensions
func (d *Decoder) SuggestThreads(xsize, ysize uint32) int {
	return int(C.JxlResizableParallelRunnerSuggestThreads(C.uint64_t(xsize), C.uint64_t(ysize)))
}

// SetDecompressBoxes configures whether to decompress boxes
func (d *Decoder) SetDecompressBoxes(decompress bool) error {
	cDecompress := C.int(0)
	if decompress {
		cDecompress = C.int(1)
	}
	if C.JxlDecoderSetDecompressBoxes(d.dec, cDecompress) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// GetBoxType gets the type of the current box
func (d *Decoder) GetBoxType(decompressed bool) ([4]byte, error) {
	var boxType C.JxlBoxType
	cDecompressed := C.int(0)
	if decompressed {
		cDecompressed = C.int(1)
	}
	if C.JxlDecoderGetBoxType(d.dec, (*C.char)(unsafe.Pointer(&boxType[0])), cDecompressed) != C.JXL_DEC_SUCCESS {
		return [4]byte{}, ErrDecoderFailed
	}

	var result [4]byte
	for i := 0; i < 4; i++ {
		result[i] = byte(boxType[i])
	}
	return result, nil
}

// SetBoxBuffer sets the buffer for box data
func (d *Decoder) SetBoxBuffer(buffer []byte) error {
	if len(buffer) == 0 {
		return ErrInvalidInput
	}
	cBuffer := (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
	if C.JxlDecoderSetBoxBuffer(d.dec, cBuffer, C.size_t(len(buffer))) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// ReleaseBoxBuffer releases the box buffer and returns the number of remaining bytes
func (d *Decoder) ReleaseBoxBuffer() int {
	return int(C.JxlDecoderReleaseBoxBuffer(d.dec))
}

// ExtraChannelBufferSize calculates the required buffer size for an extra channel
func (d *Decoder) ExtraChannelBufferSize(format PixelFormat, index uint32) (int, error) {
	cFormat := format.toCPixelFormat()
	var size C.size_t
	if C.JxlDecoderExtraChannelBufferSize(d.dec, &cFormat, &size, C.uint32_t(index)) != C.JXL_DEC_SUCCESS {
		return 0, ErrDecoderFailed
	}
	return int(size), nil
}

// SetExtraChannelBuffer sets the output buffer for extra channel data
func (d *Decoder) SetExtraChannelBuffer(format PixelFormat, buffer []byte, index uint32) error {
	if len(buffer) == 0 {
		return ErrInvalidInput
	}

	cFormat := format.toCPixelFormat()
	cBuffer := unsafe.Pointer(&buffer[0])
	if C.JxlDecoderSetExtraChannelBuffer(d.dec, &cFormat, cBuffer, C.size_t(len(buffer)), C.uint32_t(index)) != C.JXL_DEC_SUCCESS {
		return ErrDecoderFailed
	}
	return nil
}

// DecodeJxlMultiplex decodes multi-channel data exactly like the C++ DecodeJxlMultiplex function
// Returns data in channel-major format to match EncodeJxlMultiplex
func DecodeJxlMultiplex(compressed []byte) (pixels []uint16, width, height, numChannels uint32, err error) {
	decoder, err := NewDecoder()
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer decoder.Close()

	// Subscribe to required events
	err = decoder.SubscribeEvents(EventBasicInfo | EventColorEncoding | EventFullImage)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// Set input data
	err = decoder.SetInput(compressed)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	decoder.CloseInput()

	var info BasicInfo
	var colorPixels []uint16
	var colorBuffer []byte
	var extraChannelBuffers [][]byte

	for {
		status := decoder.ProcessInput()
		switch status {
		case DecError:
			return nil, 0, 0, 0, ErrDecoderFailed
		case DecNeedMoreInput:
			return nil, 0, 0, 0, fmt.Errorf("unexpected need for more input")
		case DecBasicInfo:
			info, err = decoder.GetBasicInfo()
			if err != nil {
				return nil, 0, 0, 0, err
			}

			width = info.XSize
			height = info.YSize
			numChannels = info.NumColorChannels + info.NumExtraChannels

			// Set suggested thread count
			numThreads := decoder.SuggestThreads(info.XSize, info.YSize)
			decoder.SetThreads(numThreads)

		case DecColorEncoding:
			// We can ignore ICC profile for this use case

		case DecNeedImageOutBuffer:
			// Set up buffer for color channels
			format := PixelFormat{
				NumChannels: info.NumColorChannels,
				DataType:    TypeUint16,
				Endianness:  NativeEndian,
				Align:       0,
			}

			bufferSize, err := decoder.ImageOutBufferSize(format)
			if err != nil {
				return nil, 0, 0, 0, err
			}

			colorBuffer = make([]byte, bufferSize)
			err = decoder.SetImageOutBuffer(format, colorBuffer)
			if err != nil {
				return nil, 0, 0, 0, err
			}

			// Set up buffers for extra channels
			extraChannelBuffers = make([][]byte, info.NumExtraChannels)
			extraFormat := PixelFormat{
				NumChannels: 1,
				DataType:    TypeUint16,
				Endianness:  NativeEndian,
				Align:       0,
			}

			for i := uint32(0); i < info.NumExtraChannels; i++ {
				extraSize, err := decoder.ExtraChannelBufferSize(extraFormat, i)
				if err != nil {
					return nil, 0, 0, 0, err
				}

				extraChannelBuffers[i] = make([]byte, extraSize)
				err = decoder.SetExtraChannelBuffer(extraFormat, extraChannelBuffers[i], i)
				if err != nil {
					return nil, 0, 0, 0, err
				}
			}

		case DecFullImage:
			// Image is fully decoded

		case DecSuccess:
			// Convert color buffer to uint16 after decoding is complete
			colorPixels = BytesToUint16(colorBuffer)

			// Debug: Check buffer content after decoding
			if len(colorPixels) > 0 {
				debugSize := 10
				if len(colorPixels) < debugSize {
					debugSize = len(colorPixels)
				}
				fmt.Printf("DEBUG: colorBuffer size: %d bytes, colorPixels size: %d, first few values: %v\n",
					len(colorBuffer), len(colorPixels), colorPixels[:debugSize])
			}

			// All done, organize data in channel-major format
			totalPixels := int(width) * int(height) * int(numChannels)
			pixels = make([]uint16, totalPixels)
			pixelsPerChannel := int(width) * int(height)

			// Copy color channels
			for c := uint32(0); c < info.NumColorChannels; c++ {
				for i := 0; i < pixelsPerChannel; i++ {
					// Source: interleaved format
					srcIdx := i*int(info.NumColorChannels) + int(c)
					// Destination: channel-major format
					dstIdx := int(c)*pixelsPerChannel + i
					if srcIdx < len(colorPixels) && dstIdx < len(pixels) {
						pixels[dstIdx] = colorPixels[srcIdx]
					}
				}
			}

			// Convert extra channel buffers to uint16 and copy
			for c := uint32(0); c < info.NumExtraChannels; c++ {
				channelOffset := int(info.NumColorChannels+c) * pixelsPerChannel
				if int(c) < len(extraChannelBuffers) && extraChannelBuffers[c] != nil {
					extraChannelPixels := BytesToUint16(extraChannelBuffers[c])
					if channelOffset+pixelsPerChannel <= len(pixels) && len(extraChannelPixels) >= pixelsPerChannel {
						copy(pixels[channelOffset:channelOffset+pixelsPerChannel], extraChannelPixels[:pixelsPerChannel])
					}
				}
			}

			return pixels, width, height, numChannels, nil

		default:
			return nil, 0, 0, 0, fmt.Errorf("unknown decoder status: %d", status)
		}
	}
}

// DecodeOneShot is a convenience function for one-shot decoding
func DecodeOneShot(data []byte, format PixelFormat) (pixels []byte, info BasicInfo, iccProfile []byte, err error) {
	decoder, err := NewDecoder()
	if err != nil {
		return nil, BasicInfo{}, nil, err
	}
	defer decoder.Close()

	// Subscribe to required events
	err = decoder.SubscribeEvents(EventBasicInfo | EventColorEncoding | EventFullImage)
	if err != nil {
		return nil, BasicInfo{}, nil, err
	}

	// Set input data
	err = decoder.SetInput(data)
	if err != nil {
		return nil, BasicInfo{}, nil, err
	}
	decoder.CloseInput()

	var pixelBuffer []byte

	for {
		status := decoder.ProcessInput()
		switch status {
		case DecError:
			return nil, BasicInfo{}, nil, ErrDecoderFailed
		case DecNeedMoreInput:
			return nil, BasicInfo{}, nil, errors.New("error, already provided all input")
		case DecBasicInfo:
			info, err = decoder.GetBasicInfo()
			if err != nil {
				return nil, BasicInfo{}, nil, err
			}
			// Set suggested thread count
			numThreads := decoder.SuggestThreads(info.XSize, info.YSize)
			decoder.SetThreads(numThreads)
		case DecColorEncoding:
			iccProfile, err = decoder.GetICCProfile(ColorProfileTargetData)
			if err != nil {
				return nil, BasicInfo{}, nil, err
			}
		case DecNeedImageOutBuffer:
			bufferSize, err := decoder.ImageOutBufferSize(format)
			if err != nil {
				return nil, BasicInfo{}, nil, err
			}
			pixelBuffer = make([]byte, bufferSize)
			err = decoder.SetImageOutBuffer(format, pixelBuffer)
			if err != nil {
				return nil, BasicInfo{}, nil, err
			}
		case DecFullImage:
			// Image is fully decoded
		case DecSuccess:
			return pixelBuffer, info, iccProfile, nil
		default:
			return nil, BasicInfo{}, nil, errors.New("unknown decoder status")
		}
	}
}
