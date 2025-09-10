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
// Package libjxl provides Go bindings for the libjxl library
package libjxl

/*
#include <jxl/decode.h>
#include <jxl/encode.h>
#include <jxl/resizable_parallel_runner.h>
#include <jxl/thread_parallel_runner.h>
#include <jxl/types.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// DataType represents the sample data type for pixels
type DataType int

const (
	TypeFloat   DataType = C.JXL_TYPE_FLOAT
	TypeUint8   DataType = C.JXL_TYPE_UINT8
	TypeUint16  DataType = C.JXL_TYPE_UINT16
	TypeFloat16 DataType = C.JXL_TYPE_FLOAT16
)

// Endianness represents byte order
type Endianness int

const (
	NativeEndian Endianness = C.JXL_NATIVE_ENDIAN
	LittleEndian Endianness = C.JXL_LITTLE_ENDIAN
	BigEndian    Endianness = C.JXL_BIG_ENDIAN
)

// PixelFormat describes the format of pixel data
type PixelFormat struct {
	NumChannels uint32
	DataType    DataType
	Endianness  Endianness
	Align       int
}

// BasicInfo contains basic image information
type BasicInfo struct {
	XSize                 uint32
	YSize                 uint32
	BitsPerSample         uint32
	ExponentBitsPerSample uint32
	IntensityTarget       float32
	MinNits               float32
	RelativeToMaxDisplay  bool
	LinearBelow           float32
	UsesOriginalProfile   bool
	HaveContainer         bool
	AlphaBits             uint32
	AlphaPremultiplied    bool
	NumExtraChannels      uint32
	NumColorChannels      uint32
	Orientation           uint32
	IntrinsicXSize        uint32
	IntrinsicYSize        uint32
	HavePreview           bool
	HaveAnimation         bool
}

// ColorEncoding represents color space information
type ColorEncoding struct {
	ColorSpace       int
	WhitePoint       int
	Primaries        int
	TransferFunction int
	RenderingIntent  int
	Gamma            float64
}

// FrameHeader contains frame-specific information
type FrameHeader struct {
	Duration   uint32
	Timecode   uint32
	NameLength uint32
	IsLast     bool
	LayerInfo  interface{} // Simplified for now
}

// ExtraChannelType represents the type of an extra channel
type ExtraChannelType int

const (
	ChannelAlpha         ExtraChannelType = C.JXL_CHANNEL_ALPHA
	ChannelDepth         ExtraChannelType = C.JXL_CHANNEL_DEPTH
	ChannelSpotColor     ExtraChannelType = C.JXL_CHANNEL_SPOT_COLOR
	ChannelSelectionMask ExtraChannelType = C.JXL_CHANNEL_SELECTION_MASK
	ChannelBlack         ExtraChannelType = C.JXL_CHANNEL_BLACK
	ChannelCFA           ExtraChannelType = C.JXL_CHANNEL_CFA
	ChannelThermal       ExtraChannelType = C.JXL_CHANNEL_THERMAL
	ChannelUnknown       ExtraChannelType = C.JXL_CHANNEL_UNKNOWN
	ChannelOptional      ExtraChannelType = C.JXL_CHANNEL_OPTIONAL
	ChannelGeneric       ExtraChannelType = C.JXL_CHANNEL_UNKNOWN // Generic is same as Unknown in JXL
)

// ExtraChannelInfo contains information for a single extra channel
type ExtraChannelInfo struct {
	Type                  ExtraChannelType
	BitsPerSample         uint32
	ExponentBitsPerSample uint32
	DimShift              uint32
	NameLength            uint32
	AlphaPremultiplied    bool
	SpotColor             [4]float32
	CFAChannel            uint32
}

// Error types
var (
	ErrDecoderFailed = errors.New("decoder operation failed")
	ErrEncoderFailed = errors.New("encoder operation failed")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrOutOfMemory   = errors.New("out of memory")
)

// Helper functions to convert between Go and C types
func (pf *PixelFormat) toCPixelFormat() C.JxlPixelFormat {
	return C.JxlPixelFormat{
		num_channels: C.uint32_t(pf.NumChannels),
		data_type:    C.JxlDataType(pf.DataType),
		endianness:   C.JxlEndianness(pf.Endianness),
		align:        C.size_t(pf.Align),
	}
}

func basicInfoFromC(cInfo C.JxlBasicInfo) BasicInfo {
	return BasicInfo{
		XSize:                 uint32(cInfo.xsize),
		YSize:                 uint32(cInfo.ysize),
		BitsPerSample:         uint32(cInfo.bits_per_sample),
		ExponentBitsPerSample: uint32(cInfo.exponent_bits_per_sample),
		IntensityTarget:       float32(cInfo.intensity_target),
		MinNits:               float32(cInfo.min_nits),
		RelativeToMaxDisplay:  cInfo.relative_to_max_display != 0,
		LinearBelow:           float32(cInfo.linear_below),
		UsesOriginalProfile:   cInfo.uses_original_profile != 0,
		HaveContainer:         cInfo.have_container != 0,
		AlphaBits:             uint32(cInfo.alpha_bits),
		AlphaPremultiplied:    cInfo.alpha_premultiplied != 0,
		NumExtraChannels:      uint32(cInfo.num_extra_channels),
		NumColorChannels:      uint32(cInfo.num_color_channels),
		Orientation:           uint32(cInfo.orientation),
		IntrinsicXSize:        uint32(cInfo.intrinsic_xsize),
		IntrinsicYSize:        uint32(cInfo.intrinsic_ysize),
		HavePreview:           cInfo.have_preview != 0,
		HaveAnimation:         cInfo.have_animation != 0,
	}
}

func (bi *BasicInfo) toCBasicInfo() C.JxlBasicInfo {
	var cInfo C.JxlBasicInfo
	C.JxlEncoderInitBasicInfo(&cInfo)

	// Only override fields that are explicitly set (non-zero) or always required
	cInfo.xsize = C.uint32_t(bi.XSize)
	cInfo.ysize = C.uint32_t(bi.YSize)
	cInfo.bits_per_sample = C.uint32_t(bi.BitsPerSample)
	cInfo.exponent_bits_per_sample = C.uint32_t(bi.ExponentBitsPerSample)
	cInfo.uses_original_profile = C.int(boolToInt(bi.UsesOriginalProfile))

	// Only override color channel fields if they're explicitly set
	if bi.NumColorChannels != 0 {
		cInfo.num_color_channels = C.uint32_t(bi.NumColorChannels)
	}
	if bi.NumExtraChannels != 0 {
		cInfo.num_extra_channels = C.uint32_t(bi.NumExtraChannels)
	}
	if bi.AlphaBits != 0 {
		cInfo.alpha_bits = C.uint32_t(bi.AlphaBits)
	}

	// Override other fields only if they're non-zero
	if bi.IntensityTarget != 0 {
		cInfo.intensity_target = C.float(bi.IntensityTarget)
	}
	if bi.MinNits != 0 {
		cInfo.min_nits = C.float(bi.MinNits)
	}
	if bi.LinearBelow != 0 {
		cInfo.linear_below = C.float(bi.LinearBelow)
	}
	if bi.Orientation != 0 {
		cInfo.orientation = C.JxlOrientation(bi.Orientation)
	}
	if bi.IntrinsicXSize != 0 {
		cInfo.intrinsic_xsize = C.uint32_t(bi.IntrinsicXSize)
	}
	if bi.IntrinsicYSize != 0 {
		cInfo.intrinsic_ysize = C.uint32_t(bi.IntrinsicYSize)
	}

	// Boolean fields can override defaults
	cInfo.relative_to_max_display = C.int(boolToInt(bi.RelativeToMaxDisplay))
	cInfo.have_container = C.int(boolToInt(bi.HaveContainer))
	cInfo.alpha_premultiplied = C.int(boolToInt(bi.AlphaPremultiplied))
	cInfo.have_preview = C.int(boolToInt(bi.HavePreview))
	cInfo.have_animation = C.int(boolToInt(bi.HaveAnimation))

	return cInfo
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Version returns the version of the libjxl library
func Version() uint32 {
	return uint32(C.JxlDecoderVersion())
}

// CheckSignature checks if the input data contains a valid JPEG XL signature
func CheckSignature(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, ErrInvalidInput
	}

	cData := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	sig := C.JxlSignatureCheck(cData, C.size_t(len(data)))

	switch sig {
	case C.JXL_SIG_CODESTREAM, C.JXL_SIG_CONTAINER:
		return true, nil
	case C.JXL_SIG_INVALID:
		return false, nil
	case C.JXL_SIG_NOT_ENOUGH_BYTES:
		return false, fmt.Errorf("not enough bytes to determine signature")
	default:
		return false, fmt.Errorf("unknown signature status: %d", int(sig))
	}
}

// Helper function to calculate buffer size for pixel data
func CalculateBufferSize(width, height uint32, format PixelFormat) int {
	bytesPerSample := 0
	switch format.DataType {
	case TypeUint8:
		bytesPerSample = 1
	case TypeUint16, TypeFloat16:
		bytesPerSample = 2
	case TypeFloat:
		bytesPerSample = 4
	}
	return int(width) * int(height) * int(format.NumChannels) * bytesPerSample
}

// SetToSRGB initializes a ColorEncoding struct for sRGB color space
func (ce *ColorEncoding) SetToSRGB(isGray bool) {
	var cEncoding C.JxlColorEncoding
	C.JxlColorEncodingSetToSRGB(&cEncoding, C.int(boolToInt(isGray)))
	ce.ColorSpace = int(cEncoding.color_space)
	ce.WhitePoint = int(cEncoding.white_point)
	ce.Primaries = int(cEncoding.primaries)
	ce.TransferFunction = int(cEncoding.transfer_function)
	ce.RenderingIntent = int(cEncoding.rendering_intent)
	ce.Gamma = float64(cEncoding.gamma)
}

// SetToLinearSRGB initializes a ColorEncoding struct for linear sRGB color space
func (ce *ColorEncoding) SetToLinearSRGB(isGray bool) {
	var cEncoding C.JxlColorEncoding
	C.JxlColorEncodingSetToLinearSRGB(&cEncoding, C.int(boolToInt(isGray)))
	ce.ColorSpace = int(cEncoding.color_space)
	ce.WhitePoint = int(cEncoding.white_point)
	ce.Primaries = int(cEncoding.primaries)
	ce.TransferFunction = int(cEncoding.transfer_function)
	ce.RenderingIntent = int(cEncoding.rendering_intent)
	ce.Gamma = float64(cEncoding.gamma)
}

func (ce *ColorEncoding) toCColorEncoding() C.JxlColorEncoding {
	return C.JxlColorEncoding{
		color_space:       C.JxlColorSpace(ce.ColorSpace),
		white_point:       C.JxlWhitePoint(ce.WhitePoint),
		primaries:         C.JxlPrimaries(ce.Primaries),
		transfer_function: C.JxlTransferFunction(ce.TransferFunction),
		rendering_intent:  C.JxlRenderingIntent(ce.RenderingIntent),
		gamma:             C.double(ce.Gamma),
	}
}

// Runtime finalizer cleanup helper
func setFinalizer(obj interface{}, finalizer interface{}) {
	runtime.SetFinalizer(obj, finalizer)
}

// Helper functions for ExtraChannelInfo conversion
func (eci *ExtraChannelInfo) toCExtraChannelInfo() C.JxlExtraChannelInfo {
	var cInfo C.JxlExtraChannelInfo
	C.JxlEncoderInitExtraChannelInfo(C.JxlExtraChannelType(eci.Type), &cInfo)

	cInfo._type = C.JxlExtraChannelType(eci.Type)
	cInfo.bits_per_sample = C.uint32_t(eci.BitsPerSample)
	cInfo.exponent_bits_per_sample = C.uint32_t(eci.ExponentBitsPerSample)
	cInfo.dim_shift = C.uint32_t(eci.DimShift)
	cInfo.name_length = C.uint32_t(eci.NameLength)
	cInfo.alpha_premultiplied = C.int(boolToInt(eci.AlphaPremultiplied))
	cInfo.cfa_channel = C.uint32_t(eci.CFAChannel)

	// Copy spot color array
	for i := 0; i < 4; i++ {
		cInfo.spot_color[i] = C.float(eci.SpotColor[i])
	}

	return cInfo
}

func extraChannelInfoFromC(cInfo C.JxlExtraChannelInfo) ExtraChannelInfo {
	var spotColor [4]float32
	for i := 0; i < 4; i++ {
		spotColor[i] = float32(cInfo.spot_color[i])
	}

	return ExtraChannelInfo{
		Type:                  ExtraChannelType(cInfo._type),
		BitsPerSample:         uint32(cInfo.bits_per_sample),
		ExponentBitsPerSample: uint32(cInfo.exponent_bits_per_sample),
		DimShift:              uint32(cInfo.dim_shift),
		NameLength:            uint32(cInfo.name_length),
		AlphaPremultiplied:    cInfo.alpha_premultiplied != 0,
		SpotColor:             spotColor,
		CFAChannel:            uint32(cInfo.cfa_channel),
	}
}
