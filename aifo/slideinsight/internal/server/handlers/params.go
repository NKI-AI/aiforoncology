// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import "encoding/json"

// Common parameter validation structs used across multiple handlers.
// This file consolidates all shared parameter structs to avoid duplication.

// =============================================================================
// PATH PARAMETER STRUCTS
// =============================================================================

// SlideUIDParams represents path parameters for slide UID requests
type SlideUIDParams struct {
	SlideUID string `params:"slideUID" validate:"required"`
}

// CaseUIDParams represents path parameters for case UID requests
type CaseUIDParams struct {
	CaseUID string `params:"caseUID" validate:"required"`
}

// StudyUIDParams represents path parameters for study UID requests
type StudyUIDParams struct {
	StudyUID string `params:"studyUID" validate:"required"`
}

// TenantUIDParams represents path parameters for tenant UID requests
type TenantUIDParams struct {
	TenantUID string `params:"tenantUID" validate:"required"`
}

// UserUIDParams represents path parameters for user UID requests
type UserUIDParams struct {
	UserUID string `params:"userUID" validate:"required"`
}

// DomainUIDParams represents path parameters for domain UID requests
type DomainUIDParams struct {
	DomainUID int `params:"domainUID" validate:"required"`
}

// TenantUIDDomainUIDParams represents path parameters for tenant UID and domain UID requests
type TenantUIDDomainUIDParams struct {
	TenantUID string `params:"tenantUID" validate:"required"`
	DomainUID int    `params:"domainUID" validate:"required"`
}

// StudyUIDCaseUIDParams represents path parameters for study ID and case ID requests
type StudyUIDCaseUIDParams struct {
	StudyUID string `params:"studyUID" validate:"required"`
	CaseUID  string `params:"caseUID" validate:"required"`
}

// VectorAnnotationFileParams represents path parameters for vector annotation file requests
type VectorAnnotationFileParams struct {
	SlideUID  string `params:"slideUID" validate:"required"`
	VectorUID string `params:"vectorUID" validate:"required"`
}

// MaskTileParams represents path parameters for mask tile requests
type MaskTileParams struct {
	SlideUID string `params:"slideUID" validate:"required"`
	MaskUID  string `params:"maskUID" validate:"required"`
	Z        int    `params:"z" validate:"required"`
	X        int    `params:"x" validate:"required"`
	Y        int    `params:"y" validate:"required"`
	Format   string `params:"format" validate:"required,oneof=png"`
}

// SlideTileParams represents path parameters for slide tile requests
type SlideTileParams struct {
	SlideUID string `params:"slideUID" validate:"required"`
	Z        int    `params:"z" validate:"required,min=0"`
	X        int    `params:"x" validate:"required,min=0"`
	Y        int    `params:"y" validate:"required,min=0"`
	Format   string `params:"format" validate:"required,oneof=jpg png jxl"`
	Quality  int    `query:"q" validate:"omitempty,min=1,max=100"`
}

// =============================================================================
// REQUEST BODY INPUT STRUCTS
// =============================================================================

// -----------------------------------------------------------------------------
// Authentication Input Structs
// -----------------------------------------------------------------------------

// LoginInput represents the login request payload
type LoginInput struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required"`
}

// -----------------------------------------------------------------------------
// User Input Structs
// -----------------------------------------------------------------------------

// UserInput represents the input for creating a user
type UserInput struct {
	Email     string `json:"email" example:"user@example.com" validate:"required,email"`
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Password  string `json:"password" example:"securepassword123" validate:"required"`
	TenantUID string `json:"tenantUid" example:"tenant-123" validate:"required"`
}

// UserUpdateInput represents the input for updating user information
type UserUpdateInput struct {
	Email             *string `json:"email,omitempty" example:"user@example.com" validate:"omitempty,email"`
	FirstName         *string `json:"firstName,omitempty" example:"John"`
	LastName          *string `json:"lastName,omitempty" example:"Doe"`
	MustResetPassword *bool   `json:"mustResetPassword,omitempty" example:"true"`
	IsActive          *bool   `json:"isActive,omitempty" example:"true"`
}

// RegisterUserInput represents the input for registering a new user
type RegisterUserInput struct {
	Email     string `json:"email" example:"user@example.com" validate:"required,email"`
	FirstName string `json:"firstName" example:"John" validate:"required"`
	LastName  string `json:"lastName" example:"Doe" validate:"required"`
	Password  string `json:"password" example:"securepassword123" validate:"required,min=8"`
}

// ChangePasswordInput represents the input for changing password
type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" example:"oldpassword123" validate:"required"`
	NewPassword     string `json:"newPassword" example:"newpassword123" validate:"required,min=8"`
}

// ForcedChangePasswordInput represents the input for forced password change
type ForcedChangePasswordInput struct {
	Email           string `json:"email" example:"user@example.com" validate:"required,email"`
	CurrentPassword string `json:"currentPassword" example:"oldpassword123" validate:"required"`
	NewPassword     string `json:"newPassword" example:"newpassword123" validate:"required,min=8"`
}

// ResetPasswordInput represents the input for requesting password reset
type ResetPasswordInput struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
}

// ResetPasswordConfirmInput represents the input for confirming password reset
type ResetPasswordConfirmInput struct {
	Token       string `json:"token" example:"reset-token-123" validate:"required"`
	NewPassword string `json:"newPassword" example:"newpassword123" validate:"required,min=8"`
}

// VerifyEmailInput represents the input for email verification
type VerifyEmailInput struct {
	Token string `json:"token" example:"verification-token-123" validate:"required"`
}

// ResendVerificationInput represents the input for resending verification email
type ResendVerificationInput struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
}

// SendUserEmailInput represents the input for sending an email to a user
type SendUserEmailInput struct {
	Template string `json:"template" example:"password_reset" validate:"required,oneof=password_reset email_verification welcome"`
	Subject  string `json:"subject,omitempty" example:"Password Reset Request"`
}

// -----------------------------------------------------------------------------
// Case Input Structs
// -----------------------------------------------------------------------------

// CasesInput represents the case creation request payload
type CasesInput struct {
	Name     string `json:"name" example:"Sample Case" validate:"required"`
	Metadata string `json:"metadata" example:"{\"key\": \"value\"}"`
}

// UpdateCaseInput represents the case update request payload
type UpdateCaseInput struct {
	Name     *string          `json:"name,omitempty" example:"Updated Case Name"`
	Metadata *json.RawMessage `json:"metadata,omitempty" example:"{\"key\": \"updated_value\"}"`
}

// -----------------------------------------------------------------------------
// Study Input Structs
// -----------------------------------------------------------------------------

// StudiesInput represents the study creation request payload
type StudiesInput struct {
	TenantUID   string `json:"tenantUid" example:"XxxyXz1" validate:"required"`
	StudyUID    string `json:"studyUid" example:"XxxyXz1" validate:"required"`
	Name        string `json:"name" example:"DCIS study" validate:"required"`
	Description string `json:"description" example:"this is a new study" validate:"required"`
}

// StudyUpdateInput represents the study update request payload
type StudyUpdateInput struct {
	Name        *string          `json:"name,omitempty" example:"Updated DCIS study"`
	Description *string          `json:"description,omitempty" example:"this is an updated study"`
	Metadata    *json.RawMessage `json:"metadata,omitempty" example:"{}"`
	IsPublished *bool            `json:"isPublished,omitempty" example:"true"`
}

// AddCaseToStudyInput represents the request payload for adding a case to a study
type AddCaseToStudyInput struct {
	CaseUID string `json:"caseUid" example:"Abcd1234" validate:"required"`
}

// -----------------------------------------------------------------------------
// Tenant Input Structs
// -----------------------------------------------------------------------------

// TenantsInput represents the tenant creation request payload
type TenantsInput struct {
	Name        string `json:"name" example:"acme-corp" validate:"required"`
	Description string `json:"description" example:"ACME Corporation medical imaging department"`
}

// TenantUpdateInput represents the tenant update request payload
type TenantUpdateInput struct {
	Name        *string `json:"name,omitempty" example:"Updated Tenant Name"`
	Description *string `json:"description,omitempty" example:"Updated tenant description"`
	Status      *string `json:"status,omitempty" example:"inactive"`
}

// DomainInput represents the domain creation request payload
type DomainInput struct {
	Domain string `json:"domain" example:"organization.com" validate:"required"`
}

// AddDomainInput represents the request payload for adding a domain to a tenant
type AddDomainInput struct {
	Domain    string `json:"domain" example:"acme.com" validate:"required"`
	IsPrimary bool   `json:"isPrimary" example:"false"`
}

// UpdateDomainInput represents the request payload for updating a domain
type UpdateDomainInput struct {
	IsVerified *bool `json:"isVerified,omitempty" example:"true"`
	IsPrimary  *bool `json:"isPrimary,omitempty" example:"false"`
}

// -----------------------------------------------------------------------------
// Slide Input Structs
// -----------------------------------------------------------------------------

// SlideCreationInput represents the slide creation request payload
type SlideCreationInput struct {
	CaseUID     string `json:"caseUid" example:"Abcd1234"`
	SlideUID    string `json:"slideUid" example:"slide123"`
	SlideName   string `json:"slideName" example:"Sample Slide"`
	SlideURI    string `json:"slideUri" example:"file:///path/to/slide.svs" validate:"required"`
	ImageTypeId string `json:"imageTypeId" example:"img_type_bf_he"`
}

// UpdateSlideInput represents the slide update request payload
type UpdateSlideInput struct {
	SlideName   *string `json:"slideName,omitempty" example:"Updated Slide Name"`
	SlideURI    *string `json:"slideUri,omitempty" example:"file:///path/to/updated-slide.svs"`
	ImageTypeId *string `json:"imageTypeId,omitempty" example:"img_type_bf_he"`
}

// -----------------------------------------------------------------------------
// Permission Input Structs
// -----------------------------------------------------------------------------

// PermissionInput represents the input for creating a permission
type PermissionInput struct {
	Name        string `json:"name" example:"studies.view" validate:"required"`
	Description string `json:"description" example:"View a study" validate:"required"`
}

// -----------------------------------------------------------------------------
// Settings Input Structs
// -----------------------------------------------------------------------------

// SettingInput represents the input for creating a setting
type SettingInput struct {
	TenantID  int    `json:"tenantId" example:"1" validate:"required"`
	Key       string `json:"key" example:"enable_registration" validate:"required"`
	ValueType string `json:"valueType" example:"boolean" validate:"required,oneof=boolean number string json"`
	Value     string `json:"value" example:"true" validate:"required"`
}

// SettingUpdateInput represents the input for updating setting information
type SettingUpdateInput struct {
	ValueType *string `json:"valueType,omitempty" example:"boolean" validate:"omitempty,oneof=boolean number string json"`
	Value     *string `json:"value,omitempty" example:"false"`
}
