// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteDB_Integration(t *testing.T) {
	// Skip this test if we're not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-sqlite-*.db")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new SQLite database
	db, err := newSqliteDB(tempFile.Name())
	require.NoError(t, err)
	defer db.CloseConnections()

	// Test the database operations
	testDatabaseOperations(t, db)
}

func testDatabaseOperations(t *testing.T, db Database) {
	ctx := context.Background()

	// Initially, the database should be empty
	slides, err := db.LoadAllSlides(ctx)
	require.NoError(t, err)
	assert.Empty(t, slides)

	// Create a slide
	newSlide := NewSlide{SlideID: "Test Slide", SlideURI: "test.svs"}
	err = db.CreateSlide(ctx, newSlide)
	require.NoError(t, err)

	// Verify the slide was created
	slides, err = db.LoadAllSlides(ctx)
	require.NoError(t, err)
	assert.Len(t, slides, 1)
	assert.Equal(t, newSlide.SlideID, slides[0].SlideID)

	// Create another slide
	anotherSlide := NewSlide{SlideID: "Another Test Slide", SlideURI: "another.svs"}
	err = db.CreateSlide(ctx, anotherSlide)
	require.NoError(t, err)

	// Verify both slides are present
	slides, err = db.LoadAllSlides(ctx)
	require.NoError(t, err)
	assert.Len(t, slides, 2)
	// Slides should be ordered by ID, which should match insertion order
	assert.Equal(t, newSlide.SlideID, slides[0].SlideID)
	assert.Equal(t, anotherSlide.SlideID, slides[1].SlideID)

	// Test slide existence check
	exists, err := db.SlideExists(ctx, newSlide.SlideID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test get slide by ID
	slide, err := db.GetSlideByID(ctx, newSlide.SlideID)
	require.NoError(t, err)
	assert.Equal(t, newSlide.SlideID, slide.SlideID)

	// Test mask operations
	// First, create a mask associated with the first slide
	newMask := NewMask{
		SlideID:    newSlide.SlideID,
		MaskID:     "Test Mask",
		Name:       "Annotation Mask",
		MaskURI:    "mask.png",
		TilesURL:   "/tiles/mask",
		MaskWidth:  1024,
		MaskHeight: 768,
		MaskMpp:    0.5,
	}
	err = db.CreateMask(ctx, newMask)
	require.NoError(t, err)

	// Verify the mask was created
	masks, err := db.LoadAllMasks(ctx)
	require.NoError(t, err)
	assert.Len(t, masks, 1)
	assert.Equal(t, newMask.MaskID, masks[0].MaskID)
	assert.Equal(t, newMask.SlideID, masks[0].SlideID)

	// Test retrieving a mask by ID
	mask, err := db.GetMaskByID(ctx, newMask.MaskID)
	require.NoError(t, err)
	assert.Equal(t, newMask.MaskID, mask.MaskID)
	assert.Equal(t, newMask.Name, mask.Name)

	// Test user operations
	newUser := NewUser{
		Username: "testuser",
		Password: "hashedpassword123",
	}
	err = db.CreateUser(ctx, newUser)
	require.NoError(t, err)

	// Verify the user was created
	user, err := db.GetUserByUsername(ctx, newUser.Username)
	require.NoError(t, err)
	assert.Equal(t, newUser.Username, user.Username)
	assert.Equal(t, newUser.Password, user.Password)

	// Test updating a user's password
	newPassword := "newhashed456"
	err = db.UpdateUserPassword(ctx, newUser.Username, newPassword)
	require.NoError(t, err)

	// Verify the password was updated
	updatedUser, err := db.GetUserByUsername(ctx, newUser.Username)
	require.NoError(t, err)
	assert.Equal(t, newPassword, updatedUser.Password)
}

// TestSqliteDB_UnitTests tests the SQLite implementation using a mocked database
func TestSqliteDB_UnitTests(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize the schema
	err = initializeSQLiteSchema(db)
	require.NoError(t, err)

	// Create the SQLite DB instance
	sqliteDB := &sqliteDB{db: db}

	// Test LoadAllSlides with empty database
	slides, err := sqliteDB.LoadAllSlides(context.Background())
	require.NoError(t, err)
	assert.Empty(t, slides)

	// Test CreateSlide
	err = sqliteDB.CreateSlide(context.Background(), NewSlide{SlideID: "Test Slide", SlideURI: "test.svs"})
	require.NoError(t, err)

	// Verify the slide was created
	slides, err = sqliteDB.LoadAllSlides(context.Background())
	require.NoError(t, err)
	assert.Len(t, slides, 1)
	assert.Equal(t, "Test Slide", slides[0].SlideID)

	// Test SlideExists
	exists, err := sqliteDB.SlideExists(context.Background(), "Test Slide")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = sqliteDB.SlideExists(context.Background(), "Non-Existent")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test creating and retrieving masks
	ctx := context.Background()
	maskData := NewMask{
		SlideID: "Test Slide",
		MaskID:  "Test Mask",
		MaskURI: "mask.png",
	}
	err = sqliteDB.CreateMask(ctx, maskData)
	require.NoError(t, err)

	masks, err := sqliteDB.LoadAllMasks(ctx)
	require.NoError(t, err)
	assert.Len(t, masks, 1)

	// Test creating and retrieving users
	userData := NewUser{
		Username: "testuser",
		Password: "password123",
	}
	err = sqliteDB.CreateUser(ctx, userData)
	require.NoError(t, err)

	user, err := sqliteDB.GetUserByUsername(ctx, userData.Username)
	require.NoError(t, err)
	assert.Equal(t, userData.Username, user.Username)

	// Test password update
	err = sqliteDB.UpdateUserPassword(ctx, userData.Username, "newpassword")
	require.NoError(t, err)

	updatedUser, err := sqliteDB.GetUserByUsername(ctx, userData.Username)
	require.NoError(t, err)
	assert.Equal(t, "newpassword", updatedUser.Password)

	// Test error handling for invalid database operations
	// Close the database to simulate an error
	db.Close()

	// Operations should now fail
	_, err = sqliteDB.LoadAllSlides(context.Background())
	assert.Error(t, err)

	err = sqliteDB.CreateSlide(context.Background(), NewSlide{SlideID: "Will Fail"})
	assert.Error(t, err)

	_, err = sqliteDB.GetMaskByID(context.Background(), "Any ID")
	assert.Error(t, err)

	_, err = sqliteDB.GetUserByUsername(context.Background(), "Any User")
	assert.Error(t, err)
}

func TestNewSqliteDB(t *testing.T) {
	// Test with invalid database path - use a non-existent directory path
	_, err := newSqliteDB("/non-existent-directory/test.db")
	assert.Error(t, err)

	// Test with valid sqlite URL format
	tempFile, err := os.CreateTemp("", "test-sqlite-*.db")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	db, err := newSqliteDB("sqlite://" + tempFile.Name())
	assert.NoError(t, err)
	db.CloseConnections()
}

func TestSqliteDBConstraints(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping constraint test in short mode")
	}

	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = initializeSQLiteSchema(db)
	require.NoError(t, err)

	sqliteDB := &sqliteDB{db: db}
	ctx := context.Background()

	// Test unique constraint on slide_id
	slide1 := NewSlide{SlideID: "duplicate", SlideURI: "slide1.svs"}
	err = sqliteDB.CreateSlide(ctx, slide1)
	require.NoError(t, err)

	// Attempt to create another slide with the same ID
	slide2 := NewSlide{SlideID: "duplicate", SlideURI: "slide2.svs"}
	err = sqliteDB.CreateSlide(ctx, slide2)
	assert.Error(t, err, "Should error on duplicate slide_id")

	// Test foreign key constraint on masks
	// Attempt to create a mask for a non-existent slide
	invalidMask := NewMask{
		SlideID: "non-existent-slide",
		MaskID:  "invalid-mask",
		MaskURI: "invalid.png",
	}
	err = sqliteDB.CreateMask(ctx, invalidMask)
	assert.Error(t, err, "Should error on non-existent slide_id foreign key")
}
