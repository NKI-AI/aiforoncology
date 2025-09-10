// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface Role {
  id: number;
  short_uid: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  name: string;
  description: string;
}

interface RolesResponse {
  roles?: Role[];
}

interface CreateRoleResponse {
  role: Role;
}

interface UserRole {
  id: number;
  user_id: number;
  role_id: number;
  tenant_id?: number;
  created_at: string;
  updated_at: string;
}

export interface RolePermissionAssignment {
  permission_ids: number[];
}

export interface UserRoleAssignment {
  user_id: number;
  tenant_id?: number;
}
