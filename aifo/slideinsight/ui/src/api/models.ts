// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Centralized API model definitions for type-safe API calls
 * These models mirror the backend API responses exactly
 */

// ===== COMMON TYPES =====

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface ApiError {
  error: string;
  details?: string;
  timestamp?: string;
}

// ===== USER MODELS =====

export interface User {
  userUid: string;
  email: string;
  firstName: string;
  lastName: string;
  isActive: boolean;
  mustResetPassword: boolean;
  tenantUid: string;
  createdAt: string;
  updatedAt: string;
  // Optional field that may be added in the future
  emailVerified?: boolean;
}

export interface UsersResponse {
  users: User[];
  pagination: PaginationInfo;
}

export interface CreateUserRequest {
  email: string;
  firstName?: string;
  lastName?: string;
  password: string;
  tenantUid: string;
  mustResetPassword?: boolean;
}

export interface UpdateUserRequest {
  email?: string;
  firstName?: string;
  lastName?: string;
  isActive?: boolean;
  mustResetPassword?: boolean;
}

// ===== TENANT MODELS =====

export interface Tenant {
  tenantUid: string;
  name: string;
  description: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface TenantsResponse {
  tenants: Tenant[];
  pagination: PaginationInfo;
}

export interface CreateTenantRequest {
  name: string;
  description?: string;
}

export interface UpdateTenantRequest {
  name?: string;
  description?: string;
  status?: string;
}

// ===== STUDY MODELS =====

export interface Study {
  tenantUid: string;
  studyUid: string;
  creatorUid: string;
  name: string;
  description: string;
  metadata: string;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface StudyWithCasesAndSlides extends Study {
  caseCount?: number;
  slideCount?: number;
}

export interface StudiesResponse {
  studies: Study[];
  pagination: PaginationInfo;
}

export interface CreateStudyRequest {
  name: string;
  description?: string;
  tenantUid: string;
  isPublished?: boolean;
  metadata?: string;
}

export interface UpdateStudyRequest {
  name?: string;
  description?: string;
  isPublished?: boolean;
  metadata?: string;
}

// ===== CASE MODELS =====

export interface Case {
  tenantUid: string;
  caseUid: string;
  creatorUid: string;
  name: string;
  metadata: string;
  createdAt: string;
  updatedAt: string;
}

export interface CaseWithSlides extends Case {
  slides?: Slide[];
  slideCount?: number;
}

export interface CasesResponse {
  cases: Case[];
  pagination: PaginationInfo;
}

export interface CreateCaseRequest {
  name: string;
  tenantUid: string;
  studyUid?: string;
  metadata?: string;
}

export interface UpdateCaseRequest {
  name?: string;
  metadata?: string;
}

// ===== SLIDE MODELS =====

export interface Slide {
  slideUid: string;
  slideName?: string;
  slideUri: string;
  slideWidth?: number;
  slideHeight?: number;
  slideMpp?: number;
  creatorUid?: string;
  createdAt: string;
  updatedAt: string;
  metadata?: Record<string, any>;
}

export interface SlideWithCount extends Slide {
  maskCount?: number;
}

export interface SlidesResponse {
  slides: Slide[];
  pagination: PaginationInfo;
}

export interface CreateSlideRequest {
  slideName?: string;
  slideUri: string;
  caseUid: string;
  slideWidth?: number;
  slideHeight?: number;
  slideMpp?: number;
  metadata?: Record<string, any>;
}

export interface UpdateSlideRequest {
  slideName?: string;
  slideWidth?: number;
  slideHeight?: number;
  slideMpp?: number;
  metadata?: Record<string, any>;
}

// ===== ALGORITHM MODELS =====

export interface Algorithm {
  id: string;
  tenantId: number;
  tenantName?: string;
  name: string;
  description?: string;
  version: string;
  endpointUrl: string;
  httpMethod: string;
  executionMode: "BATCH" | "STREAM";
  progressTransport: "WEBSOCKET" | "SSE";
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface AlgorithmsResponse {
  algorithms: Algorithm[];
  pagination: PaginationInfo;
}

export interface CreateAlgorithmRequest {
  name: string;
  description?: string;
  version: string;
  endpointUrl: string;
  httpMethod: string;
  executionMode: "BATCH" | "STREAM";
  progressTransport: "WEBSOCKET" | "SSE";
  tenantId: number;
  metadata?: Record<string, any>;
}

export interface UpdateAlgorithmRequest {
  name?: string;
  description?: string;
  version?: string;
  endpointUrl?: string;
  httpMethod?: string;
  executionMode?: "BATCH" | "STREAM";
  progressTransport?: "WEBSOCKET" | "SSE";
  metadata?: Record<string, any>;
}

// ===== EMAIL TEMPLATE MODELS =====

export interface EmailTemplate {
  id: number;
  name: string;
  subject: string;
  templateType: string;
  htmlContent: string;
  textContent?: string;
  isActive: boolean;
  isSystem: boolean;
  tenantId: number;
  tenantName?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EmailTemplatesResponse {
  templates: EmailTemplate[];
  pagination: PaginationInfo;
}

export interface CreateEmailTemplateRequest {
  name: string;
  subject: string;
  templateType: string;
  htmlContent: string;
  textContent?: string;
  isActive?: boolean;
  tenantId: number;
}

export interface UpdateEmailTemplateRequest {
  name?: string;
  subject?: string;
  templateType?: string;
  htmlContent?: string;
  textContent?: string;
  isActive?: boolean;
}

// ===== ROLE AND PERMISSION MODELS =====

export interface Role {
  name: string;
  short_uid: string;
  displayName?: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface RolesResponse {
  roles: Role[];
  pagination?: PaginationInfo;
}

export interface Permission {
  name: string;
  displayName?: string;
  description?: string;
  category?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PermissionsResponse {
  permissions: Permission[];
  pagination?: PaginationInfo;
}

export interface Group {
  name: string;
  short_uid: string;
  displayName?: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface GroupsResponse {
  groups: Group[];
  pagination?: PaginationInfo;
}

// ===== COUNT RESPONSES =====

export interface CountResponse {
  count: number;
}

// ===== QUERY FILTER TYPES =====

export interface BasePaginatedQuery {
  page?: number;
  limit?: number;
  q?: string;
  sort?: string;
  dir?: "asc" | "desc";
}

export interface UserQuery extends BasePaginatedQuery {
  email?: string;
  firstName?: string;
  lastName?: string;
  isActive?: boolean;
  mustResetPassword?: boolean;
  tenantUid?: string;
}

export interface TenantQuery extends BasePaginatedQuery {
  name?: string;
  description?: string;
  status?: string;
}

export interface StudyQuery extends BasePaginatedQuery {
  name?: string;
  status?: string;
  isPublished?: boolean;
  tenantUid?: string;
  creatorUid?: string;
  filterAccessibleStudies?: boolean;
}

export interface CaseQuery extends BasePaginatedQuery {
  name?: string;
  tenantUid?: string;
  creatorUid?: string;
  studyUid?: string;
}

export interface SlideQuery extends BasePaginatedQuery {
  name?: string;
  slideName?: string;
  caseUid?: string;
  withMaskCounts?: boolean;
}

export interface AlgorithmQuery extends BasePaginatedQuery {
  name?: string;
  version?: string;
  executionMode?: "BATCH" | "STREAM";
  tenantId?: number;
}

export interface EmailTemplateQuery extends BasePaginatedQuery {
  name?: string;
  templateType?: string;
  isActive?: boolean;
  isSystem?: boolean;
  tenantId?: number;
}

export interface RoleQuery extends BasePaginatedQuery {
  name?: string;
  displayName?: string;
}

export interface PermissionQuery extends BasePaginatedQuery {
  name?: string;
  category?: string;
}

export interface GroupQuery extends BasePaginatedQuery {
  name?: string;
  displayName?: string;
}

// ===== SETTINGS MODELS =====

export interface Setting {
  id: number;
  tenantId: number;
  key: string;
  valueType: "boolean" | "number" | "string" | "json";
  value: string;
  createdAt: string;
  updatedAt: string;
}

export interface SettingsResponse {
  settings: Setting[];
  pagination: PaginationInfo;
}

export interface CreateSettingRequest {
  tenantId: number;
  key: string;
  valueType: "boolean" | "number" | "string" | "json";
  value: string;
}

export interface UpdateSettingRequest {
  valueType?: "boolean" | "number" | "string" | "json";
  value?: string;
}

export interface SettingQuery extends BasePaginatedQuery {
  key?: string;
  valueType?: "boolean" | "number" | "string" | "json";
  tenantId?: number;
}
