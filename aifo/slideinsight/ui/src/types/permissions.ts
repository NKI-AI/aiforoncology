// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface Permission {
  id: number;
  tenant_id: number; // 0 = system tenant, >0 = regular tenant
  short_uid: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePermissionRequest {
  name: string;
  description: string;
}

// Permission filter parameters for API requests
export interface PermissionFilterParams {
  filterAccessibleStudies?: boolean;
  filterAccessibleCases?: boolean;
  filterAccessibleSlides?: boolean;
}

export interface PermissionsResponse {
  permissions?: Permission[];
}

interface CreatePermissionResponse {
  permission: Permission;
}
