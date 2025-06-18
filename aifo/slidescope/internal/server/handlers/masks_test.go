// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var masksRoute = "/api/v1/raster"

// MasksServiceMock is a mock implementation of the MasksService interface for testing
type MasksServiceMock struct {
	mock.Mock
}

func (m *MasksServiceMock) GetMasks(ctx context.Context) ([]domain.Mask, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Mask), args.Error(1)
}

func (m *MasksServiceMock) SaveMask(ctx context.Context, mask domain.Mask) (domain.Mask, error) {
	args := m.Called(ctx, mask)
	if args.Get(0) == nil {
		return domain.Mask{}, args.Error(1)
	}
	return args.Get(0).(domain.Mask), args.Error(1)
}

func (m *MasksServiceMock) GetMaskTile(ctx context.Context, slideID, maskID string, z, x, y int) (domain.MaskTile, error) {
	args := m.Called(ctx, slideID, maskID, z, x, y)
	if args.Get(0) == nil {
		return domain.MaskTile{}, args.Error(1)
	}
	return args.Get(0).(domain.MaskTile), args.Error(1)
}

func (m *MasksServiceMock) GetMasksForSlide(ctx context.Context, slideID string) ([]domain.Mask, error) {
	args := m.Called(ctx, slideID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Mask), args.Error(1)
}

func (m *MasksServiceMock) Close() {
	// We don't need to mock this for handlers tests as the handlers don't call Close()
}

func TestGetMasks_All(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup mock response
	masks := []domain.Mask{
		{
			MaskID:   "mask1",
			MaskName: "Mask 1",
			MaskURI:  "file:///raster/mask1.tiff",
			SlideID:  "slide1",
		},
		{
			MaskID:   "mask2",
			MaskName: "Mask 2",
			MaskURI:  "file:///raster/mask2.tiff",
			SlideID:  "slide2",
		},
	}

	// Set expectation
	mockService.On("GetMasks", mock.Anything).Return(masks, nil)

	// Setup app with handler
	app := fiber.New()
	app.Get(masksRoute, GetMasks(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", masksRoute, nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response body
	var response domain.MaskList
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Len(t, response.Masks, 2)
	assert.Equal(t, "mask1", response.Masks[0].MaskID)
	assert.Equal(t, "mask2", response.Masks[1].MaskID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetMasks_ForSlide(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup mock response with masks for multiple slides
	masks := []domain.Mask{
		{
			MaskID:   "mask1",
			MaskName: "Mask 1",
			MaskURI:  "file:///raster/mask1.tiff",
			SlideID:  "slide1",
		},
		{
			MaskID:   "mask2",
			MaskName: "Mask 2",
			MaskURI:  "file:///raster/mask2.tiff",
			SlideID:  "slide2",
		},
		{
			MaskID:   "mask3",
			MaskName: "Mask 3",
			MaskURI:  "file:///raster/mask3.tiff",
			SlideID:  "slide1",
		},
	}

	// Set expectation
	mockService.On("GetMasks", mock.Anything).Return(masks, nil)

	// Setup app with handler
	app := fiber.New()
	app.Get(masksRoute+"/slide/:slide_id", GetMasks(mockService))

	// Test for slide1
	resp, err := app.Test(httptest.NewRequest("GET", masksRoute+"/slide/slide1", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response body - should only include masks for slide1
	var response domain.MaskList
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Len(t, response.Masks, 2)
	assert.Equal(t, "mask1", response.Masks[0].MaskID)
	assert.Equal(t, "mask3", response.Masks[1].MaskID)
	assert.Equal(t, "slide1", response.Masks[0].SlideID)
	assert.Equal(t, "slide1", response.Masks[1].SlideID)
	assert.Equal(t, "slide1", response.SlideID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetMasks_ServiceError(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Set expectation - service returns error
	mockService.On("GetMasks", mock.Anything).Return(nil, assert.AnError)

	// Setup app with handler
	app := fiber.New()
	app.Get(masksRoute, GetMasks(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", masksRoute, nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "internal error", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestAddMask_Success(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup input mask
	inputMask := domain.Mask{
		MaskURI: "file:///raster/newmask.tiff",
		SlideID: "slide1",
	}

	// Setup mock response
	createdMask := domain.Mask{
		MaskID:   "mask123",
		MaskName: "newmask Mask",
		MaskURI:  "file:///raster/newmask.tiff",
		SlideID:  "slide1",
	}

	// Set expectation
	mockService.On("SaveMask", mock.Anything, inputMask).Return(createdMask, nil)

	// Setup app with handler
	app := fiber.New()
	app.Post(masksRoute, AddMask(mockService))

	// Create request
	requestBody := `{"maskUri":"file:///raster/newmask.tiff","slideId":"slide1"}`
	req := httptest.NewRequest("POST", masksRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	// Check response body
	var response domain.Mask
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "mask123", response.MaskID)
	assert.Equal(t, "newmask Mask", response.MaskName)
	assert.Equal(t, "file:///raster/newmask.tiff", response.MaskURI)
	assert.Equal(t, "slide1", response.SlideID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestAddMask_MissingMaskURI(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(masksRoute, AddMask(mockService))

	// Create request with missing maskUri
	requestBody := `{"slideId":"slide1"}`
	req := httptest.NewRequest("POST", masksRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "maskUri is required", errorResp.Error)
}

func TestAddMask_MissingSlideID(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(masksRoute, AddMask(mockService))

	// Create request with missing slideId
	requestBody := `{"maskUri":"file:///raster/newmask.tiff"}`
	req := httptest.NewRequest("POST", masksRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "slideId is required", errorResp.Error)
}

func TestAddMask_SlideDoesNotExist(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup input mask
	inputMask := domain.Mask{
		MaskURI: "file:///raster/newmask.tiff",
		SlideID: "nonexistent",
	}

	// Set expectation - slide not found error
	mockService.On("SaveMask", mock.Anything, inputMask).Return(domain.Mask{},
		fmt.Errorf("slide with ID '%s' does not exist", "nonexistent"))

	// Setup app with handler
	app := fiber.New()
	app.Post(masksRoute, AddMask(mockService))

	// Create request
	requestBody := `{"maskUri":"file:///raster/newmask.tiff","slideId":"nonexistent"}`
	req := httptest.NewRequest("POST", masksRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "slide not found", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestAddMask_InvalidJSON(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(masksRoute, AddMask(mockService))

	// Create request with invalid JSON
	requestBody := `{"maskUri":"file:///raster/newmask.tiff",slideId:"slide1"}`
	req := httptest.NewRequest("POST", masksRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "invalid request body", errorResp.Error)
}

func TestGetMaskTile_Success(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup mock tile response
	mockTile := domain.MaskTile{
		Image:       []byte("fake mask tile data"),
		Format:      "png",
		ContentType: "image/png",
	}

	// Set expectation
	mockService.On("GetMaskTile", mock.Anything, "slide1", "mask1", 0, 0, 0).Return(mockTile, nil)

	// Setup app with handler
	app := fiber.New()
	app.Get("/api/v1/slides/:slide_id/raster/:mask_id/tile/:z/:x/:y.:format", GetMaskTile(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/slides/slide1/raster/mask1/tile/0/0/0.png", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, "public, max-age=86400", resp.Header.Get("Cache-Control"))

	// Check response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, mockTile.Image, body)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetMaskTile_MaskNotFound(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Set expectation - mask not found error
	mockService.On("GetMaskTile", mock.Anything, "slide1", "nonexistent", 0, 0, 0).Return(domain.MaskTile{},
		fmt.Errorf("no masks found for slide slide1"))

	// Setup app with handler
	app := fiber.New()
	app.Get("/api/v1/slides/:slide_id/raster/:mask_id/tile/:z/:x/:y.:format", GetMaskTile(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/slides/slide1/raster/nonexistent/tile/0/0/0.png", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "mask or slide not found", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetMaskTile_OutOfBounds(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Set expectation - tile out of bounds error
	mockService.On("GetMaskTile", mock.Anything, "slide1", "mask1", 0, 999, 999).Return(domain.MaskTile{},
		fmt.Errorf("tile coordinates out of bounds"))

	// Setup app with handler
	app := fiber.New()
	app.Get("/api/v1/slides/:slide_id/raster/:mask_id/tile/:z/:x/:y.:format", GetMaskTile(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/slides/slide1/raster/mask1/tile/0/999/999.png", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "tile not found", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetMaskTile_InvalidFormat(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Get("/api/v1/slides/:slide_id/raster/:mask_id/tile/:z/:x/:y.:format", GetMaskTile(mockService))

	// Test with invalid format (not png)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/slides/slide1/raster/mask1/tile/0/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "only png format is supported for masks", errorResp.Error)
}

func TestGetMaskTile_ServiceError(t *testing.T) {
	mockService := new(MasksServiceMock)

	// Set expectation - general service error
	mockService.On("GetMaskTile", mock.Anything, "slide1", "mask1", 0, 0, 0).Return(domain.MaskTile{}, assert.AnError)

	// Setup app with handler
	app := fiber.New()
	app.Get("/api/v1/slides/:slide_id/raster/:mask_id/tile/:z/:x/:y.:format", GetMaskTile(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/slides/slide1/raster/mask1/tile/0/0/0.png", nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "internal error", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}
