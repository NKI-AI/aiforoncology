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
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func newSqliteDB(databaseURL string) (Database, error) {
	// Strip the sqlite:// prefix if present
	dbPath := databaseURL
	if strings.HasPrefix(dbPath, "sqlite://") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite://")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open SQLite database: %w", err)
	}

	// Initialize the database schema if needed
	if err := initializeSQLiteSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize SQLite schema: %w", err)
	}

	return &sqliteDB{
		db: db,
	}, nil
}

func initializeSQLiteSchema(db *sql.DB) error {
	// Enable foreign key constraints
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS slides (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slide_id TEXT NOT NULL UNIQUE,
			slide_name TEXT,
			slide_uri TEXT NOT NULL,
			slide_width INTEGER,
			slide_height INTEGER,
			slide_mpp REAL
		);
		CREATE TABLE IF NOT EXISTS masks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slide_id TEXT NOT NULL,
			mask_id TEXT NOT NULL, 
			name TEXT,
			mask_uri TEXT NOT NULL,
			mask_width INTEGER,
			mask_height INTEGER,
			mask_mpp REAL,
			FOREIGN KEY (slide_id) REFERENCES slides(slide_id)
		);
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
	`)
	return err
}

type sqliteDB struct {
	db *sql.DB
}

func (db *sqliteDB) LoadAllSlides(_ context.Context) ([]Slide, error) {
	rows, err := db.db.Query("SELECT id, slide_id, slide_name, slide_uri, slide_width, slide_height, slide_mpp FROM slides")
	if err != nil {
		return nil, fmt.Errorf("failed to query slides table: %w", err)
	}
	defer rows.Close()

	var slides []Slide
	for rows.Next() {
		var slide Slide
		if err := rows.Scan(&slide.ID, &slide.SlideID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp); err != nil {
			return nil, fmt.Errorf("failed to scan slide row: %w", err)
		}
		slides = append(slides, slide)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over slide rows: %w", err)
	}

	return slides, nil
}

func (db *sqliteDB) CreateSlide(_ context.Context, newSlide NewSlide) error {
	_, err := db.db.Exec("INSERT INTO slides (slide_id, slide_name, slide_uri, slide_width, slide_height, slide_mpp) VALUES (?, ?, ?, ?, ?, ?)",
		newSlide.SlideID, newSlide.SlideName, newSlide.SlideURI, newSlide.SlideWidth, newSlide.SlideHeight, newSlide.SlideMpp)
	if err != nil {
		return fmt.Errorf("failed to insert slide: %w", err)
	}
	return nil
}

// SlideExists checks if a slide with the given ID already exists
func (db *sqliteDB) SlideExists(_ context.Context, slideID string) (bool, error) {
	var exists bool
	err := db.db.QueryRow("SELECT EXISTS(SELECT 1 FROM slides WHERE slide_id = ?)", slideID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if slide exists: %w", err)
	}
	return exists, nil
}

// GetSlideByID retrieves a specific slide by its slide_id
func (db *sqliteDB) GetSlideByID(_ context.Context, slideID string) (Slide, error) {
	var slide Slide
	err := db.db.QueryRow("SELECT id, slide_id, slide_name, slide_uri, slide_width, slide_height, slide_mpp FROM slides WHERE slide_id = ?", slideID).Scan(
		&slide.ID, &slide.SlideID, &slide.SlideName, &slide.SlideURI, &slide.SlideWidth, &slide.SlideHeight, &slide.SlideMpp)
	if err != nil {
		if err == sql.ErrNoRows {
			return Slide{}, fmt.Errorf("slide with ID '%s' not found", slideID)
		}
		return Slide{}, fmt.Errorf("failed to get slide: %w", err)
	}
	return slide, nil
}

func (db *sqliteDB) LoadAllMasks(_ context.Context) ([]Mask, error) {
	rows, err := db.db.Query("SELECT id, slide_id, mask_id, name, mask_uri, mask_width, mask_height, mask_mpp FROM masks")
	if err != nil {
		return nil, fmt.Errorf("failed to query masks table: %w", err)
	}
	defer rows.Close()

	var masks []Mask
	for rows.Next() {
		var mask Mask
		if err := rows.Scan(&mask.ID, &mask.SlideID, &mask.MaskID, &mask.Name, &mask.MaskURI, &mask.MaskWidth, &mask.MaskHeight, &mask.MaskMpp); err != nil {
			return nil, fmt.Errorf("failed to scan mask row: %w", err)
		}
		masks = append(masks, mask)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over mask rows: %w", err)
	}

	return masks, nil
}

func (db *sqliteDB) CreateMask(ctx context.Context, newMask NewMask) error {
	// First check if the slide exists
	exists, err := db.SlideExists(ctx, newMask.SlideID)
	if err != nil {
		return fmt.Errorf("failed to check if slide exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("cannot create mask: slide with ID '%s' does not exist", newMask.SlideID)
	}

	_, err = db.db.Exec("INSERT INTO masks (slide_id, mask_id, name, mask_uri, mask_width, mask_height, mask_mpp) VALUES (?, ?, ?, ?, ?, ?, ?)",
		newMask.SlideID, newMask.MaskID, newMask.Name, newMask.MaskURI, newMask.MaskWidth, newMask.MaskHeight, newMask.MaskMpp)
	if err != nil {
		return fmt.Errorf("failed to insert mask: %w", err)
	}
	return nil
}

// GetMaskByID retrieves a specific mask by its mask_id
func (db *sqliteDB) GetMaskByID(_ context.Context, maskID string) (Mask, error) {
	var mask Mask
	err := db.db.QueryRow(`
		SELECT id, slide_id, mask_id, name, mask_uri, mask_width, mask_height, mask_mpp 
		FROM masks 
		WHERE mask_id = ?`, maskID).Scan(
		&mask.ID, &mask.SlideID, &mask.MaskID, &mask.Name, &mask.MaskURI,
		&mask.MaskWidth, &mask.MaskHeight, &mask.MaskMpp)
	if err != nil {
		if err == sql.ErrNoRows {
			return Mask{}, fmt.Errorf("mask with ID '%s' not found", maskID)
		}
		return Mask{}, fmt.Errorf("failed to get mask: %w", err)
	}

	return mask, nil
}

func (db *sqliteDB) CreateUser(ctx context.Context, newUser NewUser) error {
	_, err := db.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", newUser.Username, newUser.Password)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (db *sqliteDB) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := db.db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user with username '%s' not found", username)
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (db *sqliteDB) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	result, err := db.db.Exec("UPDATE users SET password = ? WHERE username = ?", hashedPassword, username)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with username '%s' not found", username)
	}

	return nil
}

func (db *sqliteDB) CloseConnections() {
	if db.db != nil {
		db.db.Close()
	}
}
