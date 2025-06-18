// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabase_SqliteDB(t *testing.T) {
	ctx := context.Background()

	// Test with sqlite:// prefix
	db, err := NewDatabase(ctx, "sqlite://test.db")
	assert.Nil(t, err)
	assert.Equal(t, "*database.sqliteDB", reflect.TypeOf(db).String())
	db.CloseConnections()

	// Test with .db suffix
	db, err = NewDatabase(ctx, "test.db")
	assert.Nil(t, err)
	assert.Equal(t, "*database.sqliteDB", reflect.TypeOf(db).String())
	db.CloseConnections()

	// Test with .sqlite suffix
	db, err = NewDatabase(ctx, "test.sqlite")
	assert.Nil(t, err)
	assert.Equal(t, "*database.sqliteDB", reflect.TypeOf(db).String())
	db.CloseConnections()
}

func TestNewDatabase_PostgresDB(t *testing.T) {
	ctx := context.Background()

	// Test with postgres:// prefix
	db, err := NewDatabase(ctx, "postgres://user:password@localhost:5432/slides")
	assert.Nil(t, err)
	assert.Equal(t, "*database.postgresDB", reflect.TypeOf(db).String())
	db.CloseConnections()
}

func TestNewDatabase_InvalidDatabaseConfiguration(t *testing.T) {
	ctx := context.Background()
	_, err := NewDatabase(ctx, "invalid")
	assert.ErrorContains(t, err, "unsupported database")
}

func TestMemoryDB_Basic(t *testing.T) {
	db := newMemoryDB()
	ctx := context.Background()

	// Test slides
	slides, err := db.LoadAllSlides(ctx)
	assert.Nil(t, err)
	assert.Empty(t, slides)

	slide1 := NewSlide{
		SlideID:  "slide1",
		SlideURI: "slide1.svs",
	}
	err = db.CreateSlide(ctx, slide1)
	assert.Nil(t, err)

	exists, err := db.SlideExists(ctx, slide1.SlideID)
	assert.Nil(t, err)
	assert.True(t, exists)

	retrievedSlide, err := db.GetSlideByID(ctx, slide1.SlideID)
	assert.Nil(t, err)
	assert.Equal(t, slide1.SlideID, retrievedSlide.SlideID)

	// Test masks
	masks, err := db.LoadAllMasks(ctx)
	assert.Nil(t, err)
	assert.Empty(t, masks)

	mask1 := NewMask{
		SlideID: slide1.SlideID,
		MaskID:  "mask1",
		MaskURI: "mask1.png",
	}
	err = db.CreateMask(ctx, mask1)
	assert.Nil(t, err)

	retrievedMask, err := db.GetMaskByID(ctx, mask1.MaskID)
	assert.Nil(t, err)
	assert.Equal(t, mask1.MaskID, retrievedMask.MaskID)

	// Test users
	user1 := NewUser{
		Username: "user1",
		Password: "pass1",
	}
	err = db.CreateUser(ctx, user1)
	assert.Nil(t, err)

	retrievedUser, err := db.GetUserByUsername(ctx, user1.Username)
	assert.Nil(t, err)
	assert.Equal(t, user1.Username, retrievedUser.Username)
	assert.Equal(t, user1.Password, retrievedUser.Password)

	// Test password update
	err = db.UpdateUserPassword(ctx, user1.Username, "newpass")
	assert.Nil(t, err)

	updatedUser, err := db.GetUserByUsername(ctx, user1.Username)
	assert.Nil(t, err)
	assert.Equal(t, "newpass", updatedUser.Password)

	// Test error cases
	_, err = db.GetSlideByID(ctx, "nonexistent")
	assert.Error(t, err)

	_, err = db.GetMaskByID(ctx, "nonexistent")
	assert.Error(t, err)

	_, err = db.GetUserByUsername(ctx, "nonexistent")
	assert.Error(t, err)

	err = db.UpdateUserPassword(ctx, "nonexistent", "anything")
	assert.Error(t, err)
}

func TestDatabaseMock(t *testing.T) {
	ctx := context.Background()
	mockDB := new(DatabaseMock)

	// Setup mock expectations for LoadAllSlides
	mockSlides := []Slide{{ID: 1, SlideID: "mock-slide"}}
	mockDB.On("LoadAllSlides", ctx).Return(mockSlides, nil)

	// Call the mocked method
	slides, err := mockDB.LoadAllSlides(ctx)
	assert.NoError(t, err)
	assert.Equal(t, mockSlides, slides)

	// Verify expectations were met
	mockDB.AssertExpectations(t)
}

func assertSlide(t *testing.T, slide Slide, expectedID int, expected NewSlide) {
	assert.Equal(t, expectedID, slide.ID)
	assert.Equal(t, expected.SlideID, slide.SlideID)
	assert.Equal(t, expected.SlideURI, slide.SlideURI)
}
