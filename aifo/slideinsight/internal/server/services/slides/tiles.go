// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"time"

	"aifo.dev/aifo/fastslide_go/fastslide"
	"aifo.dev/aifo/libjxl_go"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/slides"
	"github.com/gofiber/fiber/v2/log"
)

// GetSlideTile retrieves a specific tile from a slide
func (s *slidesService) GetSlideTile(ctx context.Context, slideUID string, z, x, y int, format string, quality int) (domain.SlideTile, error) {
	// Fetch slide data directly from database
	slide, err := s.GetSlideByUID(ctx, slideUID)
	if err != nil {
		log.Error("Failed to get slide for tile", "slideUID", slideUID, "z", z, "x", x, "y", y, "error", err)
		return domain.SlideTile{}, err
	}

	uri := slide.SlideURI

	// Validate format and quality
	validFormat := false
	contentType := ""
	switch format {
	case "jpg":
		validFormat = true
		contentType = "image/jpeg"
		// Quality parameter is supported for JPEG
	case "png":
		validFormat = true
		contentType = "image/png"
		// PNG doesn't support quality compression - give error if quality is set and not 100
		if quality > 0 && quality != 100 {
			return domain.SlideTile{}, errors.WithDetails(errors.ErrInvalidFormat, "PNG format does not support quality compression (quality must be 100 or omitted)")
		}
	case "jxl":
		validFormat = true
		contentType = "image/jxl"
		// Quality parameter is supported for JPEG XL: 1-99=lossy compression, 100=lossless, default=75 if not specified
	}

	if !validFormat {
		log.Warn("Invalid tile format requested", "slideUID", slideUID, "format", format, "z", z, "x", x, "y", y)
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInvalidFormat, "format: %s", format)
	}

	// Get tile image by opening file fresh each time
	// getTileStart := time.Now()
	img, err := s.getTileFresh(slideUID, uri, z, x, y, true)
	if err != nil {
		// Enhanced error logging with all context
		log.Error("Fresh tile generation failed",
			"slideUID", slideUID,
			"uri", uri,
			"z", z,
			"x", x,
			"y", y,
			"format", format,
			"error", err,
			"errorType", fmt.Sprintf("%T", err))

		// Special logging for z=0, x=0, y=0 case
		if z == 0 && x == 0 && y == 0 {
			log.Error("CRITICAL: Failed to get tile 0/0/0",
				"slideUID", slideUID,
				"uri", uri,
				"error", err)
		}

		// Check for out of bounds error using proper error types
		if errors.IsOutOfBounds(err) {
			log.Debug("Tile request out of bounds", "slideUID", slideUID, "z", z, "x", x, "y", y)
			return domain.SlideTile{}, errors.ErrTileOutOfBounds
		}
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to generate tile: %v", err)
	}
	// getTileDuration := time.Since(getTileStart)
	// fmt.Printf("JXL GetTileFresh took: %v\n", getTileDuration)
	// Encode the image to the requested format
	buf := &bytes.Buffer{}
	switch format {
	case "png":
		image, err := img.ToGoImage()
		if err != nil {
			return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to convert tile to Go image: %v", err)
		}
		err = png.Encode(buf, image)
	case "jxl":
		if img.Format() == fastslide.FormatRGB || img.Format() == fastslide.FormatRGBA {
			err = encodeRGBAasJXL(buf, img, quality)
		} else {
			// tempInterleaved, err := img.ToInterleaved()
			// if err != nil {
			// 	return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to convert to interleaved: %w", err)
			// }

			err = encodeImageAsJXL(buf, img, quality)
		}
	default: // jpeg
		image, err := img.ToGoImage()
		if err != nil {
			return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to convert tile to Go image: %v", err)
		}
		// Use provided quality or default to 75 if not specified
		jpegQuality := 75
		if quality > 0 {
			jpegQuality = quality
		}
		err = jpeg.Encode(buf, image, &jpeg.Options{Quality: jpegQuality})
	}

	if err != nil {
		log.Error("Failed to encode tile",
			"slideUID", slideUID,
			"z", z,
			"x", x,
			"y", y,
			"format", format,
			"error", err)
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to encode tile: %v", err)
	}

	return domain.SlideTile{
		Image:       buf.Bytes(),
		Format:      format,
		ContentType: contentType,
	}, nil
}

// getTileFresh opens the file fresh and generates a tile without any caching
func (s *slidesService) getTileFresh(slideUID string, uri string, z, x, y int, isSlide bool) (*fastslide.Image, error) {
	// Open slide fresh for each request
	var slideObj slides.Slide
	var err error
	// This taks about 10-20ms.
	if isSlide {
		slideObj, err = slides.OpenSlide(uri, "")
	} else {
		slideObj, err = slides.NewTiffAdapter(uri)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open slide %s: %w", uri, err)
	}
	defer slideObj.Close() // Always close after use

	// Create tile generator fresh
	bgColor := "#FFFFFF" // White background for slides
	if !isSlide {
		bgColor = "transparent" // Transparent for masks
	}

	// This is very fast and takes less than 1ms.
	generator, err := slides.NewXYZTileGenerator(slideObj, bgColor, !isSlide) // precise=true for masks
	if err != nil {
		return nil, fmt.Errorf("failed to create tile generator: %w", err)
	}
	// Generate the tile
	// This takes the bulk of th time.
	tile, err := generator.GetTile(z, x, y)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}

	return tile, nil
}

func encodeImageAsJXL(buf *bytes.Buffer, img *fastslide.Image, quality int) error {
	// 1. Gather image info
	info, err := img.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get image info: %w", err)
	}
	width, height, channels := uint32(info.Width), uint32(info.Height), uint32(info.Channels)

	// 2.1 Get the raw data as bytes
	// Takes about 5ms
	rawData, err := img.GetRawData()
	if err != nil {
		return fmt.Errorf("failed to get raw image data: %w", err)
	}

	// 3. Convert bytes to uint16 using proper byte order conversion
	// Takes about 2ms
	result := libjxl.BytesToUint16(rawData)

	// 4. Set quality - use default of lossless if not specified (matching Go example)
	jxlQuality := float32(100)
	if quality > 0 {
		jxlQuality = float32(quality)
	}

	// encodeStart := time.Now()
	compressed, err := libjxl.EncodeJxlMultiplex(result, width, height, channels, jxlQuality)
	// encodeDuration := time.Since(encodeStart)
	// fmt.Printf("JXL EncodeJxlMultiplex took: %v\n", encodeDuration)
	if err != nil {
		return fmt.Errorf("failed to compress image with JPEG XL: %w", err)
	}

	// 6. Write compressed data to buffer
	// Takes less than 1ms.
	buf.Write(compressed)

	return nil
}

// encodeRGBAasJXL encodes RGB or RGBA images to JPEG XL format
func encodeRGBAasJXL(buf *bytes.Buffer, img *fastslide.Image, quality int) error {
	// 1. Gather image info
	info, err := img.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get image info: %w", err)
	}
	width, height, channels := uint32(info.Width), uint32(info.Height), uint32(info.Channels)

	// 2. Get raw image bytes
	rawData, err := img.GetRawData()
	if err != nil {
		return fmt.Errorf("failed to get raw image data: %w", err)
	}

	// 3. Convert bytes to uint8 (RGB/RGBA are already uint8)
	result := rawData

	// 4. Set quality - use default of 75 if not specified
	jxlQuality := float32(75)
	if quality > 0 {
		jxlQuality = float32(quality)
	}

	// 5. Set up pixel format for RGB/RGBA
	format := libjxl.PixelFormat{
		NumChannels: channels, // RGB=3, RGBA=4
		DataType:    libjxl.TypeUint8,
		Endianness:  libjxl.NativeEndian,
		Align:       0,
	}

	// 6. Encode to JXL
	encodeStart := time.Now()
	compressed, err := libjxl.EncodeOneShot(result, width, height, format, jxlQuality)
	encodeDuration := time.Since(encodeStart)
	fmt.Printf("JXL EncodeOneShot took: %v\n", encodeDuration)
	if err != nil {
		return fmt.Errorf("failed to compress image with JPEG XL: %w", err)
	}

	// 7. Write compressed data to buffer
	buf.Write(compressed)

	return nil
}
