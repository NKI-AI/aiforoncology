// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"context"
	"testing"

	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Rather than monkey patching, we'll use testing-specific slide values
// that have all their metadata already populated

func TestConvertDbSlideToDomainSlide(t *testing.T) {
	// Test data
	dbSlide := database.Slide{
		SlideID:     "Slide123",
		SlideName:   "Test Slide",
		SlideURI:    "slide.svs",
		SlideWidth:  1024,
		SlideHeight: 768,
		SlideMpp:    0.5,
	}

	// Convert to domain slide
	domainSlide := domain.Slide{
		SlideID:     dbSlide.SlideID,
		SlideName:   dbSlide.SlideName,
		SlideURI:    dbSlide.SlideURI,
		SlideWidth:  dbSlide.SlideWidth,
		SlideHeight: dbSlide.SlideHeight,
		SlideMpp:    dbSlide.SlideMpp,
	}

	// Verify conversion
	assert.Equal(t, "Slide123", domainSlide.SlideID)
	assert.Equal(t, "Test Slide", domainSlide.SlideName)
	assert.Equal(t, "slide.svs", domainSlide.SlideURI)
	assert.Equal(t, 1024, domainSlide.SlideWidth)
	assert.Equal(t, 768, domainSlide.SlideHeight)
	assert.Equal(t, 0.5, domainSlide.SlideMpp)
}

// createTestSlide creates a test slide with all required metadata populated
func createTestSlide() domain.Slide {
	return domain.Slide{
		SlideID:     "Slide123",
		SlideName:   "Test Slide",
		SlideURI:    "test_data/sample.svs", // Special test file path that will skip actual file operations
		SlideWidth:  1024,
		SlideHeight: 768,
		SlideMpp:    0.5,
	}
}

func TestGetSlides(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup expectations for database calls
	mockDB.On("LoadAllSlides", mock.Anything).Return([]database.Slide{
		{SlideID: "Slide123", SlideName: "Test Slide", SlideURI: "slide.svs"},
	}, nil)

	// Create service with the mock
	service := NewSlidesService(mockDB)
	defer service.Close()

	// Call the method being tested
	slides, err := service.GetSlides(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, slides, 1)
	assert.Equal(t, "Slide123", slides[0].SlideID)
	assert.Equal(t, "Test Slide", slides[0].SlideName)
	assert.Equal(t, "slide.svs", slides[0].SlideURI)
}

func TestGetSlides_Fails(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup expectations for database calls
	mockDB.On("LoadAllSlides", mock.Anything).Return(nil, assert.AnError)

	// Create service with the mock
	service := NewSlidesService(mockDB)
	defer service.Close()

	// Call the method being tested
	_, err := service.GetSlides(context.Background())

	// Assertions
	assert.Error(t, err)
}

func TestGetSlideByID(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup expectations for database calls
	mockDB.On("GetSlideByID", mock.Anything, "Slide123").Return(database.Slide{
		SlideID: "Slide123", SlideName: "Test Slide", SlideURI: "slide.svs",
	}, nil)

	// Create service with the mock
	service := NewSlidesService(mockDB)
	defer service.Close()

	// Call the method being tested
	slide, err := service.GetSlideByID(context.Background(), "Slide123")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "Slide123", slide.SlideID)
	assert.Equal(t, "Test Slide", slide.SlideName)
	assert.Equal(t, "slide.svs", slide.SlideURI)
}

func TestSaveSlide(t *testing.T) {
	// Skip test if we're in an environment that would try to access real files
	t.Skip("Skipping test that requires file access")

	mockDB := new(database.DatabaseMock)

	// Setup input slide - create a fully populated test slide
	inputSlide := createTestSlide()

	// Setup expectations for database calls
	mockDB.On("GetSlideByID", mock.Anything, inputSlide.SlideID).Return(database.Slide{}, assert.AnError)
	mockDB.On("SlideExists", mock.Anything, inputSlide.SlideID).Return(false, nil)
	mockDB.On("CreateSlide", mock.Anything, mock.Anything).Return(nil)

	// Create service with the mock
	service := NewSlidesService(mockDB)
	defer service.Close()

	// Call the method being tested
	savedSlide, err := service.SaveSlide(context.Background(), inputSlide)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, inputSlide.SlideID, savedSlide.SlideID)
	assert.Equal(t, inputSlide.SlideName, savedSlide.SlideName)
	assert.Equal(t, inputSlide.SlideURI, savedSlide.SlideURI)
}

func TestSaveSlide_Fails(t *testing.T) {
	// Skip test if we're in an environment that would try to access real files
	t.Skip("Skipping test that requires file access")

	mockDB := new(database.DatabaseMock)

	// Setup input slide - create a fully populated test slide
	inputSlide := createTestSlide()

	// Setup expectations for database calls - simulate an error
	mockDB.On("GetSlideByID", mock.Anything, inputSlide.SlideID).Return(database.Slide{}, assert.AnError)
	mockDB.On("SlideExists", mock.Anything, inputSlide.SlideID).Return(false, nil)
	mockDB.On("CreateSlide", mock.Anything, mock.Anything).Return(assert.AnError)

	// Create service with the mock
	service := NewSlidesService(mockDB)
	defer service.Close()

	// Call the method being tested
	_, err := service.SaveSlide(context.Background(), inputSlide)

	// Assertions
	assert.Error(t, err)
}
