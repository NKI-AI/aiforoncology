// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"testing"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase for testing permission consistency - implements PermissionDB interface
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) UserHasRolePermission(ctx context.Context, userID int, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabase) HasObjectGrant(ctx context.Context, userID int, permission, resourceType string, resourceID int) (bool, error) {
	args := m.Called(ctx, userID, permission, resourceType, resourceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabase) GetCaseStudyRelations(ctx context.Context, caseID int) ([]int, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockDatabase) GetStudyByInternalID(ctx context.Context, studyID int) (ports.Study, error) {
	args := m.Called(ctx, studyID)
	return args.Get(0).(ports.Study), args.Error(1)
}

// TestSharedPermissionChecker_StudyAndCaseConsistency tests that the permission
// checker ensures consistency between study and case access permissions
func TestSharedPermissionChecker_StudyAndCaseConsistency(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDatabase{}
	checker := NewSharedPermissionChecker(mockDB)

	userID := 123
	studyID := 456
	caseID := 789

	// Test case: User has studies.view permission on a study via object grant
	// They should be able to:
	// 1. View the study
	// 2. View cases in that study (this was the bug we fixed)

	study := ports.Study{
		ID:       studyID,
		StudyUID: "test-study-123",
		Name:     "Test Study",
	}

	case_ := ports.Case{
		ID:      caseID,
		CaseUID: "test-case-456",
		Name:    "Test Case",
	}

	t.Run("User with studies.view object grant can access study", func(t *testing.T) {
		// Mock: User has no role permissions
		mockDB.On("UserHasRolePermission", ctx, userID, "studies.view").Return(false, nil)

		// Mock: User has studies.view object grant on the study
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(true, nil)

		// Test that user can view the study
		canView, err := checker.CanViewStudy(ctx, userID, study)
		assert.NoError(t, err)
		assert.True(t, canView, "User should be able to view study with studies.view object grant")
	})

	t.Run("User with studies.view object grant can access cases in that study", func(t *testing.T) {
		// Mock: User has no role permissions for cases
		mockDB.On("UserHasRolePermission", ctx, userID, "cases.view").Return(false, nil)

		// Mock: User has no direct case permission
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "case", caseID).Return(false, nil)

		// Mock: Case belongs to the study
		mockDB.On("GetCaseStudyRelations", ctx, caseID).Return([]int{studyID}, nil)

		// Mock: User has no cases.view permission on parent study
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "study", studyID).Return(false, nil)

		// Mock: User has studies.view permission on parent study (the key fix!)
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(true, nil)

		// Test that user can view the case through study access
		canView, err := checker.CanViewCase(ctx, userID, case_)
		assert.NoError(t, err)
		assert.True(t, canView, "User should be able to view case when they have studies.view on parent study")
	})

	t.Run("Permission consistency: both study and case access work with same grant", func(t *testing.T) {
		// This test demonstrates the fix: a single studies.view grant should give access to both
		// the study AND the cases within it

		// Reset mocks
		mockDB.ExpectedCalls = nil

		// For study access:
		mockDB.On("UserHasRolePermission", ctx, userID, "studies.view").Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(true, nil)

		// For case access (same user, same permissions):
		mockDB.On("UserHasRolePermission", ctx, userID, "cases.view").Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "case", caseID).Return(false, nil)
		mockDB.On("GetCaseStudyRelations", ctx, caseID).Return([]int{studyID}, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "study", studyID).Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(true, nil)

		// Test study access
		canViewStudy, err := checker.CanViewStudy(ctx, userID, study)
		assert.NoError(t, err)
		assert.True(t, canViewStudy)

		// Test case access
		canViewCase, err := checker.CanViewCase(ctx, userID, case_)
		assert.NoError(t, err)
		assert.True(t, canViewCase)

		// Both should work with the same studies.view object grant
		assert.Equal(t, canViewStudy, canViewCase,
			"Study and case access should be consistent when user has studies.view on the study")
	})

	t.Run("User without any permissions cannot access study or cases", func(t *testing.T) {
		// Reset mocks
		mockDB.ExpectedCalls = nil

		// For study access:
		mockDB.On("UserHasRolePermission", ctx, userID, "studies.view").Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(false, nil)

		// For case access:
		mockDB.On("UserHasRolePermission", ctx, userID, "cases.view").Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "case", caseID).Return(false, nil)
		mockDB.On("GetCaseStudyRelations", ctx, caseID).Return([]int{studyID}, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "study", studyID).Return(false, nil)
		mockDB.On("HasObjectGrant", ctx, userID, "studies.view", "study", studyID).Return(false, nil)

		// Test study access
		canViewStudy, err := checker.CanViewStudy(ctx, userID, study)
		assert.NoError(t, err)
		assert.False(t, canViewStudy)

		// Test case access
		canViewCase, err := checker.CanViewCase(ctx, userID, case_)
		assert.NoError(t, err)
		assert.False(t, canViewCase)
	})

	mockDB.AssertExpectations(t)
}

// TestSharedPermissionChecker_CaseSpecificPermissions tests that case-specific
// permissions still work correctly
func TestSharedPermissionChecker_CaseSpecificPermissions(t *testing.T) {
	ctx := context.Background()
	mockDB := &MockDatabase{}
	checker := NewSharedPermissionChecker(mockDB)

	userID := 123
	caseID := 789

	case_ := ports.Case{
		ID:      caseID,
		CaseUID: "test-case-456",
		Name:    "Test Case",
	}

	t.Run("User with direct cases.view role permission can access any case", func(t *testing.T) {
		// Mock: User has role permission for cases
		mockDB.On("UserHasRolePermission", ctx, userID, "cases.view").Return(true, nil)

		// Test that user can view the case
		canView, err := checker.CanViewCase(ctx, userID, case_)
		assert.NoError(t, err)
		assert.True(t, canView, "User should be able to view case with cases.view role permission")
	})

	t.Run("User with direct cases.view object grant on case can access it", func(t *testing.T) {
		// Reset mocks
		mockDB.ExpectedCalls = nil

		// Mock: User has no role permissions
		mockDB.On("UserHasRolePermission", ctx, userID, "cases.view").Return(false, nil)

		// Mock: User has direct case permission
		mockDB.On("HasObjectGrant", ctx, userID, "cases.view", "case", caseID).Return(true, nil)

		// Test that user can view the case
		canView, err := checker.CanViewCase(ctx, userID, case_)
		assert.NoError(t, err)
		assert.True(t, canView, "User should be able to view case with direct cases.view object grant")
	})

	mockDB.AssertExpectations(t)
}
