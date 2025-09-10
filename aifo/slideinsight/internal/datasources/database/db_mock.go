// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) LoadAllSlides(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Slide, error) {
	args := m.Called(ctx, search, pagination)
	return args.Get(0).([]ports.Slide), args.Error(1)
}

func (m *DatabaseMock) GetSlidesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) CreateSlide(ctx context.Context, newSlide ports.NewSlide) error {
	args := m.Called(ctx, newSlide)
	return args.Error(0)
}

func (m *DatabaseMock) SlideExists(ctx context.Context, slideUID string) (bool, error) {
	args := m.Called(ctx, slideUID)
	return args.Bool(0), args.Error(1)
}

func (m *DatabaseMock) GetSlideByUID(ctx context.Context, slideUID string) (ports.Slide, error) {
	args := m.Called(ctx, slideUID)
	return args.Get(0).(ports.Slide), args.Error(1)
}

func (m *DatabaseMock) LoadAllMasks(ctx context.Context) ([]ports.Mask, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Mask), args.Error(1)
}

func (m *DatabaseMock) CreateMask(ctx context.Context, newMask ports.NewMask) error {
	args := m.Called(ctx, newMask)
	return args.Error(0)
}

func (m *DatabaseMock) GetMaskByUID(ctx context.Context, maskUID string) (ports.Mask, error) {
	args := m.Called(ctx, maskUID)
	return args.Get(0).(ports.Mask), args.Error(1)
}

func (m *DatabaseMock) CreateUser(ctx context.Context, newUser ports.NewUser) error {
	args := m.Called(ctx, newUser)
	return args.Error(0)
}

func (m *DatabaseMock) GetUserByUsername(ctx context.Context, username string) (ports.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(ports.User), args.Error(1)
}

func (m *DatabaseMock) GetUserByUID(ctx context.Context, userUID string) (ports.User, error) {
	args := m.Called(ctx, userUID)
	return args.Get(0).(ports.User), args.Error(1)
}

func (m *DatabaseMock) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	args := m.Called(ctx, username, hashedPassword)
	return args.Error(0)
}

func (m *DatabaseMock) UpdateUser(ctx context.Context, username string, updates ports.UserUpdates) error {
	args := m.Called(ctx, username, updates)
	return args.Error(0)
}

func (m *DatabaseMock) UpdateUserByUID(ctx context.Context, userUID string, updates ports.UserUpdates) error {
	args := m.Called(ctx, userUID, updates)
	return args.Error(0)
}

func (m *DatabaseMock) LoadAllTenants(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Tenant, error) {
	args := m.Called(ctx, search, pagination)
	return args.Get(0).([]ports.Tenant), args.Error(1)
}

func (m *DatabaseMock) GetTenantsCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) GetTenantsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) CreateTenant(ctx context.Context, newTenant ports.NewTenant) error {
	args := m.Called(ctx, newTenant)
	return args.Error(0)
}

func (m *DatabaseMock) TenantExists(ctx context.Context, tenantUID string) (bool, error) {
	args := m.Called(ctx, tenantUID)
	return args.Bool(0), args.Error(1)
}

func (m *DatabaseMock) GetTenantByUID(ctx context.Context, tenantUID string) (ports.Tenant, error) {
	args := m.Called(ctx, tenantUID)
	return args.Get(0).(ports.Tenant), args.Error(1)
}

func (m *DatabaseMock) CloseConnections() {
	m.Called()
}

func (m *DatabaseMock) LoadAllCases(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Case, error) {
	args := m.Called(ctx, search, pagination)
	return args.Get(0).([]ports.Case), args.Error(1)
}

func (m *DatabaseMock) GetCasesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) LoadAllUsers(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.User, error) {
	args := m.Called(ctx, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.User), args.Error(1)
}

func (m *DatabaseMock) GetUserCount(ctx context.Context, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) DeactivateUser(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *DatabaseMock) ActivateUser(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

// LoadAllStudies mock implementation
func (m *DatabaseMock) LoadAllStudies(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	args := m.Called(ctx, search, pagination)
	return args.Get(0).([]ports.Study), args.Error(1)
}

func (m *DatabaseMock) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) CreateStudy(ctx context.Context, newStudy ports.NewStudy) error {
	args := m.Called(ctx, newStudy)
	return args.Error(0)
}

func (m *DatabaseMock) GetStudyByUID(ctx context.Context, studyUID string) (ports.Study, error) {
	args := m.Called(ctx, studyUID)
	return args.Get(0).(ports.Study), args.Error(1)
}

func (m *DatabaseMock) UpdateStudy(ctx context.Context, studyUID string, updates ports.StudyUpdates) error {
	args := m.Called(ctx, studyUID, updates)
	return args.Error(0)
}

func (m *DatabaseMock) GetStudyIDByShortUID(ctx context.Context, studyUID string) (int, error) {
	args := m.Called(ctx, studyUID)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) GetStudyCaseCounts(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *DatabaseMock) SoftDeleteStudy(ctx context.Context, studyUID string, deletedBy int) error {
	args := m.Called(ctx, studyUID, deletedBy)
	return args.Error(0)
}

func (m *DatabaseMock) GetDeletedStudies(ctx context.Context) ([]ports.Study, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Study), args.Error(1)
}

func (m *DatabaseMock) RestoreStudy(ctx context.Context, studyUID string) error {
	args := m.Called(ctx, studyUID)
	return args.Error(0)
}

// Additional slides methods that might be missing
func (m *DatabaseMock) GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]ports.Slide, error) {
	args := m.Called(ctx, caseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Slide), args.Error(1)
}

func (m *DatabaseMock) SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error {
	args := m.Called(ctx, slideUID, deletedBy)
	return args.Error(0)
}

func (m *DatabaseMock) GetDeletedSlides(ctx context.Context) ([]ports.Slide, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Slide), args.Error(1)
}

func (m *DatabaseMock) RestoreSlide(ctx context.Context, slideUID string) error {
	args := m.Called(ctx, slideUID)
	return args.Error(0)
}

// Cases methods
func (m *DatabaseMock) CreateCase(ctx context.Context, newCase ports.NewCase) error {
	args := m.Called(ctx, newCase)
	return args.Error(0)
}

func (m *DatabaseMock) GetCaseByUID(ctx context.Context, caseUID string) (ports.Case, error) {
	args := m.Called(ctx, caseUID)
	return args.Get(0).(ports.Case), args.Error(1)
}

func (m *DatabaseMock) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]ports.Case, error) {
	args := m.Called(ctx, studyUID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Case), args.Error(1)
}

func (m *DatabaseMock) GetCasesByStudyUIDCount(ctx context.Context, studyUID string, search utils.SearchParams) (int, error) {
	args := m.Called(ctx, studyUID, search)
	return args.Int(0), args.Error(1)
}

func (m *DatabaseMock) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	args := m.Called(ctx, studyUID, caseUID)
	return args.Error(0)
}

func (m *DatabaseMock) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	args := m.Called(ctx, studyUID, caseUID)
	return args.Error(0)
}

func (m *DatabaseMock) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	args := m.Called(ctx, caseUID, deletedBy)
	return args.Error(0)
}

func (m *DatabaseMock) GetDeletedCases(ctx context.Context) ([]ports.Case, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Case), args.Error(1)
}

func (m *DatabaseMock) RestoreCase(ctx context.Context, caseUID string) error {
	args := m.Called(ctx, caseUID)
	return args.Error(0)
}

// Additional mask methods that are missing
func (m *DatabaseMock) SoftDeleteMask(ctx context.Context, maskUID string, deletedBy int) error {
	args := m.Called(ctx, maskUID, deletedBy)
	return args.Error(0)
}

func (m *DatabaseMock) GetDeletedMasks(ctx context.Context) ([]ports.Mask, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.Mask), args.Error(1)
}

func (m *DatabaseMock) RestoreMask(ctx context.Context, maskUID string) error {
	args := m.Called(ctx, maskUID)
	return args.Error(0)
}
