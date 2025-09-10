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
#include <jxl/encode.h>
#include <jxl/thread_parallel_runner.h>
#include <jxl/codestream_header.h>
#include <stdlib.h>
#include <string.h>

// C wrapper to avoid cgo pointer-to-pointer violations
static size_t process_output_wrapper(JxlEncoder* enc,
                                     uint8_t* buf, size_t buf_size,
                                     JxlEncoderStatus* status_out) {
    uint8_t* next_out = buf;
    size_t avail = buf_size;
    *status_out = JxlEncoderProcessOutput(enc, &next_out, &avail);
    return buf_size - avail;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// EncoderStatus represents the encoder status
type EncoderStatus int

const (
	EncSuccess        EncoderStatus = C.JXL_ENC_SUCCESS
	EncError          EncoderStatus = C.JXL_ENC_ERROR
	EncNeedMoreOutput EncoderStatus = C.JXL_ENC_NEED_MORE_OUTPUT
)

// EncoderError represents encoder error types
type EncoderError int

const (
	EncErrOk           EncoderError = C.JXL_ENC_ERR_OK
	EncErrGeneric      EncoderError = C.JXL_ENC_ERR_GENERIC
	EncErrOOM          EncoderError = C.JXL_ENC_ERR_OOM
	EncErrJBRD         EncoderError = C.JXL_ENC_ERR_JBRD
	EncErrBadInput     EncoderError = C.JXL_ENC_ERR_BAD_INPUT
	EncErrNotSupported EncoderError = C.JXL_ENC_ERR_NOT_SUPPORTED
	EncErrAPIUsage     EncoderError = C.JXL_ENC_ERR_API_USAGE
)

// FrameSettingId represents frame setting options
type FrameSettingId int

const (
	FrameSettingEffort                 FrameSettingId = C.JXL_ENC_FRAME_SETTING_EFFORT
	FrameSettingDecodingSpeed          FrameSettingId = C.JXL_ENC_FRAME_SETTING_DECODING_SPEED
	FrameSettingResampling             FrameSettingId = C.JXL_ENC_FRAME_SETTING_RESAMPLING
	FrameSettingExtraChannelResampling FrameSettingId = C.JXL_ENC_FRAME_SETTING_EXTRA_CHANNEL_RESAMPLING
	FrameSettingAlreadyDownsampled     FrameSettingId = C.JXL_ENC_FRAME_SETTING_ALREADY_DOWNSAMPLED
	FrameSettingPhotonNoise            FrameSettingId = C.JXL_ENC_FRAME_SETTING_PHOTON_NOISE
	FrameSettingNoise                  FrameSettingId = C.JXL_ENC_FRAME_SETTING_NOISE
	FrameSettingDots                   FrameSettingId = C.JXL_ENC_FRAME_SETTING_DOTS
	FrameSettingPatches                FrameSettingId = C.JXL_ENC_FRAME_SETTING_PATCHES
	FrameSettingEPF                    FrameSettingId = C.JXL_ENC_FRAME_SETTING_EPF
	FrameSettingGaborish               FrameSettingId = C.JXL_ENC_FRAME_SETTING_GABORISH
	FrameSettingModular                FrameSettingId = C.JXL_ENC_FRAME_SETTING_MODULAR
	FrameSettingKeepInvisible          FrameSettingId = C.JXL_ENC_FRAME_SETTING_KEEP_INVISIBLE
	FrameSettingGroupOrder             FrameSettingId = C.JXL_ENC_FRAME_SETTING_GROUP_ORDER
	FrameSettingGroupOrderCenterX      FrameSettingId = C.JXL_ENC_FRAME_SETTING_GROUP_ORDER_CENTER_X
	FrameSettingGroupOrderCenterY      FrameSettingId = C.JXL_ENC_FRAME_SETTING_GROUP_ORDER_CENTER_Y
	FrameSettingResponsive             FrameSettingId = C.JXL_ENC_FRAME_SETTING_RESPONSIVE
	FrameSettingProgressiveAC          FrameSettingId = C.JXL_ENC_FRAME_SETTING_PROGRESSIVE_AC
	FrameSettingQProgressiveAC         FrameSettingId = C.JXL_ENC_FRAME_SETTING_QPROGRESSIVE_AC
	FrameSettingProgressiveDC          FrameSettingId = C.JXL_ENC_FRAME_SETTING_PROGRESSIVE_DC
)

// Encoder represents a JPEG XL encoder
type Encoder struct {
	enc    *C.JxlEncoder
	runner unsafe.Pointer // void*
}

// FrameSettings represents frame-specific encoder settings
type FrameSettings struct {
	settings *C.JxlEncoderFrameSettings
	enc      *Encoder // Keep reference to encoder
}

// NewEncoder creates a new JPEG XL encoder
func NewEncoder() (*Encoder, error) {
	enc := C.JxlEncoderCreate(nil)
	if enc == nil {
		return nil, ErrEncoderFailed
	}

	runner := C.JxlThreadParallelRunnerCreate(nil, C.JxlThreadParallelRunnerDefaultNumWorkerThreads())
	if runner == nil {
		C.JxlEncoderDestroy(enc)
		return nil, ErrOutOfMemory
	}

	if C.JxlEncoderSetParallelRunner(enc, (*[0]byte)(C.JxlThreadParallelRunner), runner) != C.JXL_ENC_SUCCESS {
		C.JxlThreadParallelRunnerDestroy(runner)
		C.JxlEncoderDestroy(enc)
		return nil, ErrEncoderFailed
	}

	e := &Encoder{
		enc:    enc,
		runner: runner,
	}

	setFinalizer(e, (*Encoder).destroy)
	return e, nil
}

// destroy cleans up the encoder resources
func (e *Encoder) destroy() {
	if e.runner != nil {
		C.JxlThreadParallelRunnerDestroy(e.runner)
		e.runner = nil
	}
	if e.enc != nil {
		C.JxlEncoderDestroy(e.enc)
		e.enc = nil
	}
}

// Close manually cleans up the encoder resources
func (e *Encoder) Close() {
	e.destroy()
	runtime.SetFinalizer(e, nil)
}

// Reset resets the encoder to its initial state
func (e *Encoder) Reset() {
	C.JxlEncoderReset(e.enc)
}

// SetBasicInfo sets the basic image information
func (e *Encoder) SetBasicInfo(info BasicInfo) error {
	cInfo := info.toCBasicInfo()
	if C.JxlEncoderSetBasicInfo(e.enc, &cInfo) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetColorEncoding sets the color encoding
func (e *Encoder) SetColorEncoding(encoding ColorEncoding) error {
	cEncoding := encoding.toCColorEncoding()
	if C.JxlEncoderSetColorEncoding(e.enc, &cEncoding) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetICCProfile sets the ICC profile
func (e *Encoder) SetICCProfile(profile []byte) error {
	if len(profile) == 0 {
		return ErrInvalidInput
	}
	cProfile := (*C.uint8_t)(unsafe.Pointer(&profile[0]))
	if C.JxlEncoderSetICCProfile(e.enc, cProfile, C.size_t(len(profile))) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// UseContainer sets whether to use the container format
func (e *Encoder) UseContainer(useContainer bool) error {
	cUseContainer := C.int(0)
	if useContainer {
		cUseContainer = C.int(1)
	}
	if C.JxlEncoderUseContainer(e.enc, cUseContainer) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// StoreJPEGMetadata configures whether to store JPEG metadata
func (e *Encoder) StoreJPEGMetadata(store bool) error {
	cStore := C.int(0)
	if store {
		cStore = C.int(1)
	}
	if C.JxlEncoderStoreJPEGMetadata(e.enc, cStore) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetCodestreamLevel sets the codestream level
func (e *Encoder) SetCodestreamLevel(level int) error {
	if C.JxlEncoderSetCodestreamLevel(e.enc, C.int(level)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// GetRequiredCodestreamLevel gets the required codestream level
func (e *Encoder) GetRequiredCodestreamLevel() int {
	return int(C.JxlEncoderGetRequiredCodestreamLevel(e.enc))
}

// FrameSettingsCreate creates new frame settings
func (e *Encoder) FrameSettingsCreate() *FrameSettings {
	settings := C.JxlEncoderFrameSettingsCreate(e.enc, nil)
	return &FrameSettings{
		settings: settings,
		enc:      e,
	}
}

// SetOption sets an integer frame setting option
func (fs *FrameSettings) SetOption(option FrameSettingId, value int64) error {
	if C.JxlEncoderFrameSettingsSetOption(fs.settings, C.JxlEncoderFrameSettingId(option), C.int64_t(value)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetFloatOption sets a float frame setting option
func (fs *FrameSettings) SetFloatOption(option FrameSettingId, value float32) error {
	if C.JxlEncoderFrameSettingsSetFloatOption(fs.settings, C.JxlEncoderFrameSettingId(option), C.float(value)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetLossless sets whether to use lossless encoding
func (fs *FrameSettings) SetLossless(lossless bool) error {
	cLossless := C.int(0)
	if lossless {
		cLossless = C.int(1)
	}
	if C.JxlEncoderSetFrameLossless(fs.settings, cLossless) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetDistance sets the lossy compression distance
func (fs *FrameSettings) SetDistance(distance float32) error {
	if C.JxlEncoderSetFrameDistance(fs.settings, C.float(distance)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetExtraChannelDistance sets the compression distance for a specific extra channel
func (fs *FrameSettings) SetExtraChannelDistance(index uint32, distance float32) error {
	if C.JxlEncoderSetExtraChannelDistance(fs.settings, C.size_t(index), C.float(distance)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// SetExtraChannelBuffer sets the buffer for a specific extra channel
func (fs *FrameSettings) SetExtraChannelBuffer(format PixelFormat, buffer []byte, index uint32) error {
	if len(buffer) == 0 {
		return ErrInvalidInput
	}

	cFormat := format.toCPixelFormat()
	cBuffer := unsafe.Pointer(&buffer[0])
	if C.JxlEncoderSetExtraChannelBuffer(fs.settings, &cFormat, cBuffer, C.size_t(len(buffer)), C.uint32_t(index)) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// AddImageFrame adds an image frame to the encoder
func (fs *FrameSettings) AddImageFrame(format PixelFormat, pixels []byte) error {
	if len(pixels) == 0 {
		return ErrInvalidInput
	}

	cFormat := format.toCPixelFormat()
	cPixels := unsafe.Pointer(&pixels[0])
	if C.JxlEncoderAddImageFrame(fs.settings, &cFormat, cPixels, C.size_t(len(pixels))) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// AddJPEGFrame adds a JPEG frame to the encoder
func (fs *FrameSettings) AddJPEGFrame(jpegData []byte) error {
	if len(jpegData) == 0 {
		return ErrInvalidInput
	}

	cData := (*C.uint8_t)(unsafe.Pointer(&jpegData[0]))
	if C.JxlEncoderAddJPEGFrame(fs.settings, cData, C.size_t(len(jpegData))) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}
	return nil
}

// CloseInput signals that no more input will be provided
func (e *Encoder) CloseInput() {
	C.JxlEncoderCloseInput(e.enc)
}

// ProcessOutput processes the encoder output
func (e *Encoder) ProcessOutput(buf []byte) (int, EncoderStatus, error) {
	if len(buf) == 0 {
		return 0, EncError, ErrInvalidInput
	}

	// Use C wrapper to avoid CGO pointer-to-pointer violations
	var status C.JxlEncoderStatus
	written := C.process_output_wrapper(
		e.enc,
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&status,
	)

	return int(written), EncoderStatus(status), nil
}

// DistanceFromQuality converts quality to distance
func DistanceFromQuality(quality float32) float32 {
	return float32(C.JxlEncoderDistanceFromQuality(C.float(quality)))
}

// SetExtraChannelInfo sets information for an extra channel
func (e *Encoder) SetExtraChannelInfo(index uint32, info ExtraChannelInfo) error {
	cInfo := info.toCExtraChannelInfo()

	if C.JxlEncoderSetExtraChannelInfo(e.enc, C.size_t(index), &cInfo) != C.JXL_ENC_SUCCESS {
		return ErrEncoderFailed
	}

	return nil
}

// InitExtraChannelInfo initializes an ExtraChannelInfo struct with default values
func InitExtraChannelInfo(channelType ExtraChannelType, bitsPerSample uint32) ExtraChannelInfo {
	return ExtraChannelInfo{
		Type:                  channelType,
		BitsPerSample:         bitsPerSample,
		ExponentBitsPerSample: 0,
		DimShift:              0,
		NameLength:            0,
		AlphaPremultiplied:    false,
		SpotColor:             [4]float32{0, 0, 0, 0},
		CFAChannel:            0,
	}
}

// EncodeJxlMultiplex encodes multi-channel data exactly like the C++ EncodeJxlMultiplex function
// Uses 1 color channel + (numChannels-1) extra channels approach for consistency with C++
func EncodeJxlMultiplex(pixels []uint16, xsize, ysize, numChannels uint32, quality float32) ([]byte, error) {
	encoder, err := NewEncoder()
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	// Set basic info with 1 color channel + (numChannels-1) extra channels
	basicInfo := BasicInfo{
		XSize:                 xsize,
		YSize:                 ysize,
		BitsPerSample:         16,
		ExponentBitsPerSample: 0,
		NumColorChannels:      1,               // Always 1 color channel for multiplex
		NumExtraChannels:      numChannels - 1, // Rest are extra channels
		UsesOriginalProfile:   true,
	}

	err = encoder.SetBasicInfo(basicInfo)
	if err != nil {
		return nil, err
	}

	// Set color encoding for the single color channel
	var colorEncoding ColorEncoding
	colorEncoding.SetToSRGB(true) // Grayscale for single color channel
	err = encoder.SetColorEncoding(colorEncoding)
	if err != nil {
		return nil, err
	}

	// Configure extra channels
	for idx := uint32(0); idx < numChannels-1; idx++ {
		extraChannelInfo := InitExtraChannelInfo(ChannelOptional, 16)
		extraChannelInfo.ExponentBitsPerSample = 0

		err = encoder.SetExtraChannelInfo(idx, extraChannelInfo)
		if err != nil {
			return nil, err
		}
	}

	// Set codestream level and use container
	req := encoder.GetRequiredCodestreamLevel()
	err = encoder.SetCodestreamLevel(req)
	if err != nil {
		return nil, err
	}

	err = encoder.UseContainer(true)
	if err != nil {
		return nil, err
	}

	// Create frame settings
	frameSettings := encoder.FrameSettingsCreate()
	err = frameSettings.SetOption(FrameSettingModular, 1)
	if err != nil {
		return nil, err
	}

	err = frameSettings.SetOption(FrameSettingEffort, 1)
	if err != nil {
		return nil, err
	}

	// Calculate distance from quality
	var distance float32
	if quality >= 100 {
		err = frameSettings.SetLossless(true)
		distance = 0.0
	} else {
		distance = DistanceFromQuality(quality)
		// Use VarDCT mode for true lossy compression (modular=0)
		// TODO: It gives weird results?

		err = frameSettings.SetOption(FrameSettingModular, 0)
		if err != nil {
			return nil, err
		}

		err = frameSettings.SetDistance(distance)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	// Set distance for each extra channel
	for i := uint32(0); i < numChannels-1; i++ {
		err = frameSettings.SetExtraChannelDistance(i, distance)
		if err != nil {
			return nil, err
		}
	}

	// Add the frame with proper channel data extraction
	format := PixelFormat{
		NumChannels: 1, // Only the first channel for base image
		DataType:    TypeUint16,
		Endianness:  NativeEndian,
		Align:       0,
	}

	// Extract first channel for base image
	pixelsPerChannel := int(xsize) * int(ysize)
	firstChannelData := pixels[0:pixelsPerChannel] // First channel from channel-major format
	firstChannelBytes := Uint16ToBytes(firstChannelData)

	err = frameSettings.AddImageFrame(format, firstChannelBytes)
	if err != nil {
		return nil, err
	}

	// Set extra channel buffers (channels 1 through numChannels-1)
	extraFormat := PixelFormat{
		NumChannels: 1,
		DataType:    TypeUint16,
		Endianness:  NativeEndian,
		Align:       0,
	}

	for i := uint32(0); i < numChannels-1; i++ {
		// Extract channel data (channel i+1 since channel 0 is the base image)
		channelStartIdx := int(i+1) * pixelsPerChannel
		channelEndIdx := channelStartIdx + pixelsPerChannel

		if channelEndIdx <= len(pixels) {
			channelData := pixels[channelStartIdx:channelEndIdx]
			channelBytes := Uint16ToBytes(channelData)

			err = frameSettings.SetExtraChannelBuffer(extraFormat, channelBytes, i)
			if err != nil {
				return nil, err
			}
		}
	}

	// Close input
	encoder.CloseInput()

	// Process output
	// Pre-allocate output with capacity estimate to avoid reallocations
	capHint := int(xsize) * int(ysize) * int(numChannels) * 2 // 16-bit data estimate
	output := make([]byte, 0, capHint)
	buffer := make([]byte, 64*1024)

	for {
		bytesWritten, status, err := encoder.ProcessOutput(buffer)
		if err != nil {
			return nil, err
		}

		output = append(output, buffer[:bytesWritten]...)

		switch status {
		case EncSuccess:
			return output, nil
		case EncNeedMoreOutput:
			continue
		case EncError:
			return nil, ErrEncoderFailed
		default:
			return nil, fmt.Errorf("unknown encoder status")
		}
	}
}

// EncodeOneShot is a convenience function for one-shot encoding
func EncodeOneShot(pixels []byte, xsize, ysize uint32, format PixelFormat, quality float32) ([]byte, error) {
	encoder, err := NewEncoder()
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	// Set basic info based on the format
	var bitsPerSample, exponentBits uint32
	switch format.DataType {
	case TypeUint8:
		bitsPerSample = 8
		exponentBits = 0
	case TypeUint16:
		bitsPerSample = 16
		exponentBits = 0
	case TypeFloat16:
		bitsPerSample = 16
		exponentBits = 5
	case TypeFloat:
		bitsPerSample = 32
		exponentBits = 8
	default:
		bitsPerSample = 8
		exponentBits = 0
	}

	// Split logic: handle GA/RGBA specially, otherwise 1-3 color channels as-is
	var numColorChannels, numExtraChannels uint32
	var alphaBits uint32

	switch format.NumChannels {
	case 1:
		numColorChannels = 1
		numExtraChannels = 0
		alphaBits = 0
	case 2: // Grayscale + Alpha
		numColorChannels = 1
		numExtraChannels = 1
		alphaBits = bitsPerSample
	case 3:
		numColorChannels = 3
		numExtraChannels = 0
		alphaBits = 0
	case 4: // RGB + Alpha
		numColorChannels = 3
		numExtraChannels = 1
		alphaBits = bitsPerSample
	default:
		// 5+ channels: treat all as extra channels (advanced multiplex use-case)
		numColorChannels = 0
		numExtraChannels = format.NumChannels
		alphaBits = 0
	}

	basicInfo := BasicInfo{
		XSize:                 xsize,
		YSize:                 ysize,
		BitsPerSample:         bitsPerSample,
		ExponentBitsPerSample: exponentBits,
		NumColorChannels:      numColorChannels,
		NumExtraChannels:      numExtraChannels,
		AlphaBits:             alphaBits,
		UsesOriginalProfile:   false,
	}

	// For lossless encoding the encoder expects UsesOriginalProfile=true
	if quality >= 100 {
		basicInfo.UsesOriginalProfile = true
	}

	err = encoder.SetBasicInfo(basicInfo)
	if err != nil {
		return nil, err
	}

	// Only set color encoding when we actually have color channels
	if numColorChannels > 0 {
		var colorEncoding ColorEncoding
		isGray := (numColorChannels == 1)
		colorEncoding.SetToSRGB(isGray)
		err = encoder.SetColorEncoding(colorEncoding)
		if err != nil {
			return nil, err
		}
	}

	// Configure extra channels BEFORE getting required codestream level
	// The encoder needs to know the complete extra channel layout first
	if numExtraChannels > 0 {
		for idx := uint32(0); idx < numExtraChannels; idx++ {
			var chanType ExtraChannelType
			// For GA/RGBA, first extra channel is alpha; otherwise generic
			if (format.NumChannels == 2 || format.NumChannels == 4) && idx == 0 {
				chanType = ChannelAlpha
			} else {
				chanType = ChannelGeneric
			}

			extraChannelInfo := InitExtraChannelInfo(chanType, bitsPerSample)
			// For alpha channel we must set AlphaBits in BasicInfo; already set above
			extraChannelInfo.AlphaPremultiplied = false

			if C.JxlEncoderSetExtraChannelInfo == nil {
				// Fallback to container mode-only workflow if function is not available
				// (should not happen with our headers)
			}

			err = encoder.SetExtraChannelInfo(idx, extraChannelInfo)
			if err != nil {
				return nil, err
			}
		}
	}

	// Now that encoder knows the complete channel layout, set codestream level and container
	if numExtraChannels > 0 {
		// 1) Get required codestream level based on the extra channels we just configured
		req := encoder.GetRequiredCodestreamLevel()
		err = encoder.SetCodestreamLevel(req)
		if err != nil {
			return nil, err
		}

		// 2) Wrap in the JXL container so extra-channel metadata is preserved
		err = encoder.UseContainer(true)
		if err != nil {
			return nil, err
		}
	}

	// Create frame settings
	frameSettings := encoder.FrameSettingsCreate()

	// Enable modular mode for >3 channels (required for multi-channel images)
	// JPEG path can handle up to 3 color channels natively
	// if format.NumChannels > 3 {
	// 	err = frameSettings.SetOption(FrameSettingModular, 1)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// Set quality/distance
	if quality >= 100 {
		err = frameSettings.SetLossless(true)
	} else {
		distance := DistanceFromQuality(quality)
		err = frameSettings.SetDistance(distance)
	}
	if err != nil {
		return nil, err
	}

	// Build base image format and pixels (exclude alpha if present)
	baseFormat := format
	baseFormat.NumChannels = numColorChannels

	basePixels := pixels
	// Determine bytes per sample for slicing
	bytesPerSample := 1
	switch format.DataType {
	case TypeUint8:
		bytesPerSample = 1
	case TypeUint16, TypeFloat16:
		bytesPerSample = 2
	case TypeFloat:
		bytesPerSample = 4
	}

	// Optionally prepared alpha buffer for GA/RGBA; must be set AFTER AddImageFrame
	var alphaBuffer []byte

	// Handle GA/RGBA: split alpha out to extra channel buffer
	if (format.NumChannels == 2 || format.NumChannels == 4) && numColorChannels > 0 {
		// Prepare base (color) pixel buffer without alpha
		numPixels := int(xsize) * int(ysize)
		baseStride := int(format.NumChannels) * bytesPerSample
		colorStride := int(numColorChannels) * bytesPerSample
		basePixels = make([]byte, numPixels*colorStride)

		// Extract color channels, skipping alpha
		for i := 0; i < numPixels; i++ {
			// copy first numColorChannels from each pixel
			copy(
				basePixels[i*colorStride:(i+1)*colorStride],
				pixels[i*baseStride:i*baseStride+colorStride],
			)
		}

		// Build alpha extra channel buffer
		alphaBuffer = make([]byte, numPixels*bytesPerSample)
		// alpha is the last channel
		alphaOffset := (int(format.NumChannels) - 1) * bytesPerSample
		for i := 0; i < numPixels; i++ {
			copy(
				alphaBuffer[i*bytesPerSample:(i+1)*bytesPerSample],
				pixels[i*baseStride+alphaOffset:i*baseStride+alphaOffset+bytesPerSample],
			)
		}
	}

	// Add image frame for base (color) channels
	err = frameSettings.AddImageFrame(baseFormat, basePixels)
	if err != nil {
		return nil, err
	}

	// Now set extra channel buffers (must come after AddImageFrame)
	if alphaBuffer != nil {
		extraFormat := PixelFormat{
			NumChannels: 1,
			DataType:    format.DataType,
			Endianness:  format.Endianness,
			Align:       0,
		}
		if err := frameSettings.SetExtraChannelBuffer(extraFormat, alphaBuffer, 0); err != nil {
			return nil, err
		}
	}

	// Close input
	encoder.CloseInput()

	// Process output
	// Pre-allocate output with capacity estimate to avoid reallocations
	bytesPerPixel := int(bitsPerSample) / 8 * int(format.NumChannels)
	capHint := int(xsize) * int(ysize) * bytesPerPixel
	output := make([]byte, 0, capHint)
	buffer := make([]byte, 64*1024) // Reuse this buffer for all chunks

	for {
		bytesWritten, status, err := encoder.ProcessOutput(buffer)
		if err != nil {
			return nil, err
		}

		output = append(output, buffer[:bytesWritten]...)

		switch status {
		case EncSuccess:
			return output, nil
		case EncNeedMoreOutput:
			// Continue with the same buffer
			continue
		case EncError:
			return nil, ErrEncoderFailed
		default:
			return nil, errors.New("unknown encoder status")
		}
	}
}
