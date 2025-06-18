// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// Slide represents a slide in the database
type Slide struct {
	ID          int
	SlideID     string
	SlideName   string
	SlideURI    string
	SlideWidth  int
	SlideHeight int
	SlideMpp    float64
}

// NewSlide represents a new slide to be created to the database
type NewSlide struct {
	SlideID     string
	SlideName   string
	SlideURI    string
	SlideWidth  int
	SlideHeight int
	SlideMpp    float64 // TODO: Make SlideMpp
}

// Mask represents a mask in the database
type Mask struct {
	ID         int
	SlideID    string
	MaskID     string
	Name       string
	MaskURI    string
	TilesURL   string
	MaskWidth  int
	MaskHeight int
	MaskMpp    float64
}

// NewMask represents a new mask to be created to the database
type NewMask struct {
	SlideID    string
	MaskID     string
	Name       string
	MaskURI    string
	TilesURL   string
	MaskWidth  int
	MaskHeight int
	MaskMpp    float64
}

type User struct {
	ID       int
	Username string
	Password string
}

type NewUser struct {
	Username string
	Password string
}

// Database defines the interface for interacting with the slide database.
// Using this interface allows changing the implementation without affecting the rest of the code.
type Database interface {
	// LoadAllSlides retrieves all slides from the database.
	LoadAllSlides(ctx context.Context) ([]Slide, error)

	// CreateSlide adds a new slide to the database.
	CreateSlide(ctx context.Context, newSlide NewSlide) error

	// SlideExists checks if a slide with the given ID already exists.
	SlideExists(ctx context.Context, slideID string) (bool, error)

	// GetSlideByID retrieves a specific slide by its slide_id.
	GetSlideByID(ctx context.Context, slideID string) (Slide, error)

	// LoadAllMasks retrieves all masks from the database.
	LoadAllMasks(ctx context.Context) ([]Mask, error)

	// CreateMask adds a new mask to the database.
	CreateMask(ctx context.Context, newMask NewMask) error

	// GetMaskByID retrieves a specific mask by its mask_id.
	GetMaskByID(ctx context.Context, maskID string) (Mask, error)

	// CreateUser adds a new user to the database.
	CreateUser(ctx context.Context, newUser NewUser) error

	// GetUserByUsername retrieves a specific user by its username.
	GetUserByUsername(ctx context.Context, username string) (User, error)

	// UpdateUserPassword updates the password for a user with the specified username.
	UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error

	// CloseConnections closes all open connections to the database.
	CloseConnections()
}

// NewDatabase creates a new Database instance
func NewDatabase(ctx context.Context, databaseURL string) (Database, error) {
	// Handle SQLite database
	if strings.HasPrefix(databaseURL, "sqlite://") || strings.HasSuffix(databaseURL, ".db") || strings.HasSuffix(databaseURL, ".sqlite") {
		db, err := newSqliteDB(databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SQLite database connection: %w", err)
		}
		slog.Info("Using SQLite database", "databaseURL", databaseURL)
		return db, nil
	}

	// Handle PostgreSQL database
	if strings.HasPrefix(databaseURL, "postgres://") {
		slog.Info("Using PostgreSQL database", "databaseURL", databaseURL)
		return newPostgresDB(databaseURL), nil
	}

	return nil, fmt.Errorf("unsupported database URL scheme: %s", databaseURL)
}

// memoryDB implements Database interface for in-memory storage (for testing)
type memoryDB struct {
	slides []Slide
	masks  []Mask
	users  []User
}

func newMemoryDB() *memoryDB {
	return &memoryDB{
		slides: []Slide{},
		masks:  []Mask{},
		users:  []User{},
	}
}

func (m *memoryDB) LoadAllSlides(ctx context.Context) ([]Slide, error) {
	return m.slides, nil
}

func (m *memoryDB) CreateSlide(ctx context.Context, newSlide NewSlide) error {
	m.slides = append(m.slides, Slide{
		ID:          len(m.slides) + 1,
		SlideID:     newSlide.SlideID,
		SlideName:   newSlide.SlideName,
		SlideURI:    newSlide.SlideURI,
		SlideWidth:  newSlide.SlideWidth,
		SlideHeight: newSlide.SlideHeight,
		SlideMpp:    newSlide.SlideMpp,
	})
	return nil
}

func (m *memoryDB) SlideExists(ctx context.Context, slideID string) (bool, error) {
	for _, slide := range m.slides {
		if slide.SlideID == slideID {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryDB) GetSlideByID(ctx context.Context, slideID string) (Slide, error) {
	for _, slide := range m.slides {
		if slide.SlideID == slideID {
			return slide, nil
		}
	}
	return Slide{}, fmt.Errorf("slide not found")
}

func (m *memoryDB) LoadAllMasks(ctx context.Context) ([]Mask, error) {
	return m.masks, nil
}

func (m *memoryDB) CreateMask(ctx context.Context, newMask NewMask) error {
	m.masks = append(m.masks, Mask{
		ID:         len(m.masks) + 1,
		SlideID:    newMask.SlideID,
		MaskID:     newMask.MaskID,
		Name:       newMask.Name,
		MaskURI:    newMask.MaskURI,
		TilesURL:   newMask.TilesURL,
		MaskWidth:  newMask.MaskWidth,
		MaskHeight: newMask.MaskHeight,
		MaskMpp:    newMask.MaskMpp,
	})
	return nil
}

func (m *memoryDB) GetMaskByID(ctx context.Context, maskID string) (Mask, error) {
	for _, mask := range m.masks {
		if mask.MaskID == maskID {
			return mask, nil
		}
	}
	return Mask{}, fmt.Errorf("mask not found")
}

func (m *memoryDB) CreateUser(ctx context.Context, newUser NewUser) error {
	m.users = append(m.users, User{
		ID:       len(m.users) + 1,
		Username: newUser.Username,
		Password: newUser.Password,
	})
	return nil
}

func (m *memoryDB) GetUserByUsername(ctx context.Context, username string) (User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return User{}, fmt.Errorf("user not found")
}

func (m *memoryDB) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	for i, user := range m.users {
		if user.Username == username {
			m.users[i].Password = hashedPassword
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (m *memoryDB) CloseConnections() {
	// No connections to close in memory DB
}

// postgresDB implements Database interface for PostgreSQL storage
type postgresDB struct {
	connString string
}

func newPostgresDB(connString string) *postgresDB {
	slog.Warn("PostgreSQL support is a stub implementation and not fully functional yet")
	return &postgresDB{
		connString: connString,
	}
}

func (p *postgresDB) LoadAllSlides(ctx context.Context) ([]Slide, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "LoadAllSlides")
	return []Slide{}, nil
}

func (p *postgresDB) CreateSlide(ctx context.Context, newSlide NewSlide) error {
	slog.Warn("PostgreSQL method not implemented", "method", "CreateSlide")
	return nil
}

func (p *postgresDB) SlideExists(ctx context.Context, slideID string) (bool, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "SlideExists")
	return false, nil
}

func (p *postgresDB) GetSlideByID(ctx context.Context, slideID string) (Slide, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "GetSlideByID")
	return Slide{}, nil
}

func (p *postgresDB) LoadAllMasks(ctx context.Context) ([]Mask, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "LoadAllMasks")
	return []Mask{}, nil
}

func (p *postgresDB) CreateMask(ctx context.Context, newMask NewMask) error {
	slog.Warn("PostgreSQL method not implemented", "method", "CreateMask")
	return nil
}

func (p *postgresDB) GetMaskByID(ctx context.Context, maskID string) (Mask, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "GetMaskByID")
	return Mask{}, nil
}

func (p *postgresDB) CreateUser(ctx context.Context, newUser NewUser) error {
	slog.Warn("PostgreSQL method not implemented", "method", "CreateUser")
	return nil
}

func (p *postgresDB) GetUserByUsername(ctx context.Context, username string) (User, error) {
	slog.Warn("PostgreSQL method not implemented", "method", "GetUserByUsername")
	return User{}, nil
}

func (p *postgresDB) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	slog.Warn("PostgreSQL method not implemented", "method", "UpdateUserPassword")
	return nil
}

func (p *postgresDB) CloseConnections() {
	slog.Warn("PostgreSQL method not implemented", "method", "CloseConnections")
	// Nothing to close in this stub implementation
}
