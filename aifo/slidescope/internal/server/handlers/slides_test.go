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
	"net/http"
	"net/http/httptest"
	"testing"

	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var slidesRoute = "/api/v1/slides"

// SlidesServiceMock is a mock implementation of the SlidesService interface for testing
type SlidesServiceMock struct {
	mock.Mock
}

func (m *SlidesServiceMock) GetSlides(ctx context.Context) ([]domain.Slide, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Slide), args.Error(1)
}

func (m *SlidesServiceMock) GetSlideByID(ctx context.Context, slideID string) (domain.Slide, error) {
	args := m.Called(ctx, slideID)
	if args.Get(0) == nil {
		return domain.Slide{}, args.Error(1)
	}
	return args.Get(0).(domain.Slide), args.Error(1)
}

func (m *SlidesServiceMock) GetSlideMetadata(ctx context.Context, slideID string) (domain.SlideMetadata, error) {
	args := m.Called(ctx, slideID)
	if args.Get(0) == nil {
		return domain.SlideMetadata{}, args.Error(1)
	}
	return args.Get(0).(domain.SlideMetadata), args.Error(1)
}

func (m *SlidesServiceMock) GetSlideTile(ctx context.Context, slideID string, z, x, y int, format string) (domain.SlideTile, error) {
	args := m.Called(ctx, slideID, z, x, y, format)
	if args.Get(0) == nil {
		return domain.SlideTile{}, args.Error(1)
	}
	return args.Get(0).(domain.SlideTile), args.Error(1)
}

func (m *SlidesServiceMock) SaveSlide(ctx context.Context, newSlide domain.Slide) (domain.Slide, error) {
	args := m.Called(ctx, newSlide)
	if args.Get(0) == nil {
		return domain.Slide{}, args.Error(1)
	}
	return args.Get(0).(domain.Slide), args.Error(1)
}

func (m *SlidesServiceMock) Close() {
	m.Called()
}

func setupMockService(t *testing.T) *SlidesServiceMock {
	mockService := new(SlidesServiceMock)
	mockService.On("Close").Return()
	return mockService
}

func TestGetSlides(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlides", mock.Anything).Return([]domain.Slide{{SlideID: "Slide", SlideURI: "slide.svs"}}, nil)

	app := fiber.New()
	app.Get(slidesRoute, GetSlides(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute, nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body := bodyFromResponse[domain.SlidesResponse](t, resp)
	assert.Len(t, body.Slides, 1)
	assert.Equal(t, "Slide", body.Slides[0].SlideID)
	assert.Equal(t, "slide.svs", body.Slides[0].SlideURI)
}

func TestGetSlides_ServiceFails(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlides", mock.Anything).Return(nil, assert.AnError)

	app := fiber.New()
	app.Get(slidesRoute, GetSlides(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute, nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "internal error", body.Error)
}

func TestGetSlideByID(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideByID", mock.Anything, "slide1").Return(domain.Slide{SlideID: "slide1", SlideURI: "slide1.svs"}, nil)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id", GetSlideByID(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body := bodyFromResponse[domain.Slide](t, resp)
	assert.Equal(t, "slide1", body.SlideID)
	assert.Equal(t, "slide1.svs", body.SlideURI)
}

func TestGetSlideByID_EmptyID(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id", GetSlideByID(mockService))

	// This test is a bit artificial because Fiber will not route to a handler with an empty param
	// But we'll test the handler's behavior for completeness
	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for missing route params
}

func TestGetSlideByID_NotFound(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideByID", mock.Anything, "nonexistent").Return(domain.Slide{},
		fmt.Errorf("failed to get slide: slide with ID '%s' not found", "nonexistent"))

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id", GetSlideByID(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/nonexistent", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "slide not found", body.Error)
}

func TestGetSlideByID_ServiceFails(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideByID", mock.Anything, "slide1").Return(domain.Slide{}, assert.AnError)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id", GetSlideByID(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1", nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "internal error", body.Error)
}

func TestGetSlideMetadata(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideMetadata", mock.Anything, "slide1").Return(domain.SlideMetadata{
		SlideID:   "slide1",
		SlideName: "Slide 1",
		MinLevel:  0,
		MaxLevel:  10,
		TileSize:  512,
	}, nil)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/metadata", GetSlideMetadata(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/metadata", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body := bodyFromResponse[domain.SlideMetadata](t, resp)
	assert.Equal(t, "slide1", body.SlideID)
	assert.Equal(t, "Slide 1", body.SlideName)
	assert.Equal(t, 0, body.MinLevel)
	assert.Equal(t, 10, body.MaxLevel)
	assert.Equal(t, 512, body.TileSize)
}

func TestGetSlideMetadata_EmptyID(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/metadata", GetSlideMetadata(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for missing route params
}

func TestGetSlideMetadata_NotFound(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideMetadata", mock.Anything, "nonexistent").Return(domain.SlideMetadata{},
		fmt.Errorf("failed to get slide: slide with ID '%s' not found", "nonexistent"))

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/metadata", GetSlideMetadata(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/nonexistent/metadata", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "slide not found", body.Error)
}

func TestGetSlideMetadata_ServiceFails(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideMetadata", mock.Anything, "slide1").Return(domain.SlideMetadata{}, assert.AnError)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/metadata", GetSlideMetadata(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/metadata", nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "internal error", body.Error)
}

func TestGetSlideTile(t *testing.T) {
	mockService := setupMockService(t)
	mockTile := domain.SlideTile{
		Image:       []byte("fake image data"),
		Format:      "jpeg",
		ContentType: "image/jpeg",
	}
	mockService.On("GetSlideTile", mock.Anything, "slide1", 0, 0, 0, "jpeg").Return(mockTile, nil)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))
	assert.Equal(t, "public, max-age=86400", resp.Header.Get("Cache-Control"))

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, mockTile.Image, body)
}

func TestGetSlideTile_EmptyID(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/tile/0/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode) // Missing slide_id parameter
}

func TestGetSlideTile_InvalidParams(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	// Invalid Z coordinate
	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/invalid/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "Invalid zoom level", body.Error)

	// Invalid X coordinate
	resp, err = app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/invalid/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	body = bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "Invalid x coordinate", body.Error)

	// Invalid Y coordinate
	resp, err = app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/0/invalid.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	body = bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "Invalid y coordinate", body.Error)
}

func TestGetSlideTile_NotFound(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideTile", mock.Anything, "nonexistent", 0, 0, 0, "jpeg").Return(domain.SlideTile{},
		fmt.Errorf("failed to get slide: slide with ID '%s' not found", "nonexistent"))

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/nonexistent/tile/0/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "slide not found", body.Error)
}

func TestGetSlideTile_UnsupportedFormat(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideTile", mock.Anything, "slide1", 0, 0, 0, "gif").Return(domain.SlideTile{},
		fmt.Errorf("unsupported image format: gif"))

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/0/0.gif", nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "unsupported image format: gif", body.Error)
}

func TestGetSlideTile_OutOfBounds(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideTile", mock.Anything, "slide1", 0, 999, 999, "jpeg").Return(domain.SlideTile{},
		fmt.Errorf("tile coordinates out of bounds"))

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/999/999.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "tile not found", body.Error)
}

func TestGetSlideTile_ServiceFails(t *testing.T) {
	mockService := setupMockService(t)
	mockService.On("GetSlideTile", mock.Anything, "slide1", 0, 0, 0, "jpeg").Return(domain.SlideTile{}, assert.AnError)

	app := fiber.New()
	app.Get(slidesRoute+"/:slide_id/tile/:z/:x/:y.:format", GetSlideTile(mockService))

	resp, err := app.Test(httptest.NewRequest("GET", slidesRoute+"/slide1/tile/0/0/0.jpeg", nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "internal error", body.Error)
}

func TestAddSlide(t *testing.T) {
	mockService := setupMockService(t)
	inputSlide := domain.Slide{SlideID: "slide1", SlideURI: "slide1.svs"}
	returnedSlide := domain.Slide{SlideID: "slide1", SlideURI: "slide1.svs"}
	mockService.On("SaveSlide", mock.Anything, inputSlide).Return(returnedSlide, nil)

	app := fiber.New()
	app.Post(slidesRoute, AddSlide(mockService))

	resp, err := app.Test(postRequest(slidesRoute, `{"slideId":"slide1","slideUri":"slide1.svs"}`))
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	body := bodyFromResponse[domain.Slide](t, resp)
	assert.Equal(t, "slide1", body.SlideID)
	assert.Equal(t, "slide1.svs", body.SlideURI)
}

func TestAddSlide_InvalidRequest(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Post(slidesRoute, AddSlide(mockService))

	resp, err := app.Test(httptest.NewRequest("POST", slidesRoute, nil))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "invalid request", body.Error)
}

func TestAddSlide_MissingURI(t *testing.T) {
	mockService := setupMockService(t)

	app := fiber.New()
	app.Post(slidesRoute, AddSlide(mockService))

	resp, err := app.Test(postRequest(slidesRoute, `{"slideId":"slide1"}`))
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "slideUri is required", body.Error)
}

func TestAddSlide_ServiceFails(t *testing.T) {
	mockService := setupMockService(t)
	inputSlide := domain.Slide{SlideID: "slide1", SlideURI: "slide1.svs"}
	mockService.On("SaveSlide", mock.Anything, inputSlide).Return(domain.Slide{}, assert.AnError)

	app := fiber.New()
	app.Post(slidesRoute, AddSlide(mockService))

	resp, err := app.Test(postRequest(slidesRoute, `{"slideId":"slide1","slideUri":"slide1.svs"}`))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := bodyFromResponse[domain.ErrorResponse](t, resp)
	assert.Equal(t, "internal error", body.Error)
}

func postRequest(url string, body string) *http.Request {
	req := httptest.NewRequest("POST", url, bytes.NewBufferString(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return req
}

func bodyFromResponse[T any](t *testing.T, resp *http.Response) T {
	defer resp.Body.Close()
	var body T
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.Nil(t, err)
	return body
}
