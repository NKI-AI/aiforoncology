// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) LoadAllSlides(ctx context.Context) ([]Slide, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Slide), args.Error(1)
}

func (m *DatabaseMock) CreateSlide(ctx context.Context, newSlide NewSlide) error {
	args := m.Called(ctx, newSlide)
	return args.Error(0)
}

func (m *DatabaseMock) SlideExists(ctx context.Context, slideID string) (bool, error) {
	args := m.Called(ctx, slideID)
	return args.Bool(0), args.Error(1)
}

func (m *DatabaseMock) GetSlideByID(ctx context.Context, slideID string) (Slide, error) {
	args := m.Called(ctx, slideID)
	return args.Get(0).(Slide), args.Error(1)
}

func (m *DatabaseMock) LoadAllMasks(ctx context.Context) ([]Mask, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Mask), args.Error(1)
}

func (m *DatabaseMock) CreateMask(ctx context.Context, newMask NewMask) error {
	args := m.Called(ctx, newMask)
	return args.Error(0)
}

func (m *DatabaseMock) GetMaskByID(ctx context.Context, maskID string) (Mask, error) {
	args := m.Called(ctx, maskID)
	return args.Get(0).(Mask), args.Error(1)
}

func (m *DatabaseMock) CreateUser(ctx context.Context, newUser NewUser) error {
	args := m.Called(ctx, newUser)
	return args.Error(0)
}

func (m *DatabaseMock) GetUserByUsername(ctx context.Context, username string) (User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(User), args.Error(1)
}

func (m *DatabaseMock) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	args := m.Called(ctx, username, hashedPassword)
	return args.Error(0)
}

func (m *DatabaseMock) CloseConnections() {
	m.Called()
}
