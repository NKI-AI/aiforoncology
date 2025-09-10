// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Centralized query key factory for consistent cache management
 * Based on TanStack Query best practices
 */

import type {
  UserQuery,
  TenantQuery,
  StudyQuery,
  CaseQuery,
  SlideQuery,
  AlgorithmQuery,
  EmailTemplateQuery,
  RoleQuery,
  PermissionQuery,
  GroupQuery,
  SettingQuery,
} from "./models";

/**
 * Helper to create consistent query keys with sorted parameters
 */
function createQueryKey<T extends Record<string, any>>(
  base: readonly string[],
  params?: T
): readonly unknown[] {
  if (!params) return base;

  // Sort parameters for consistent cache keys
  const sortedParams = Object.keys(params)
    .sort()
    .reduce((acc, key) => {
      const value = params[key];
      if (value !== undefined && value !== null && value !== "") {
        acc[key] = value;
      }
      return acc;
    }, {} as Record<string, any>);

  return Object.keys(sortedParams).length > 0
    ? ([...base, sortedParams] as const)
    : base;
}

/**
 * Hierarchical query key factory following TanStack Query conventions
 * Each entity has all/lists/list/details/detail patterns
 */
export const queryKeys = {
  // ===== USERS =====
  users: {
    all: ["users"] as const,
    lists: () => [...queryKeys.users.all, "list"] as const,
    list: (filters?: UserQuery) =>
      createQueryKey(queryKeys.users.lists(), filters),
    details: () => [...queryKeys.users.all, "detail"] as const,
    detail: (userUid: string) =>
      [...queryKeys.users.details(), userUid] as const,
  },

  // ===== TENANTS =====
  tenants: {
    all: ["tenants"] as const,
    lists: () => [...queryKeys.tenants.all, "list"] as const,
    list: (filters?: TenantQuery) =>
      createQueryKey(queryKeys.tenants.lists(), filters),
    details: () => [...queryKeys.tenants.all, "detail"] as const,
    detail: (tenantUid: string) =>
      [...queryKeys.tenants.details(), tenantUid] as const,
    domains: (tenantUid: string) =>
      [...queryKeys.tenants.detail(tenantUid), "domains"] as const,
  },

  // ===== STUDIES =====
  studies: {
    all: ["studies"] as const,
    lists: () => [...queryKeys.studies.all, "list"] as const,
    list: (filters?: StudyQuery) =>
      createQueryKey(queryKeys.studies.lists(), filters),
    details: () => [...queryKeys.studies.all, "detail"] as const,
    detail: (studyUid: string) =>
      [...queryKeys.studies.details(), studyUid] as const,
    navigation: (studyUid: string) =>
      [...queryKeys.studies.detail(studyUid), "navigation"] as const,
    metadata: (studyUid: string) =>
      [...queryKeys.studies.detail(studyUid), "metadata-field"] as const,
    count: () => [...queryKeys.studies.all, "count"] as const,
  },

  // ===== CASES =====
  cases: {
    all: ["cases"] as const,
    lists: () => [...queryKeys.cases.all, "list"] as const,
    list: (filters?: CaseQuery) =>
      createQueryKey(queryKeys.cases.lists(), filters),
    details: () => [...queryKeys.cases.all, "detail"] as const,
    detail: (caseUid: string) =>
      [...queryKeys.cases.details(), caseUid] as const,
    slides: (caseUid: string) =>
      [...queryKeys.cases.detail(caseUid), "slides"] as const,
    count: () => [...queryKeys.cases.all, "count"] as const,
  },

  // ===== SLIDES =====
  slides: {
    all: ["slides"] as const,
    lists: () => [...queryKeys.slides.all, "list"] as const,
    list: (filters?: SlideQuery) =>
      createQueryKey(queryKeys.slides.lists(), filters),
    details: () => [...queryKeys.slides.all, "detail"] as const,
    detail: (slideUid: string) =>
      [...queryKeys.slides.details(), slideUid] as const,
    metadata: (slideUid: string) =>
      [...queryKeys.slides.detail(slideUid), "metadata"] as const,
    annotations: (slideUid: string) =>
      [...queryKeys.slides.detail(slideUid), "annotations"] as const,
    count: () => [...queryKeys.slides.all, "count"] as const,
  },

  // ===== ALGORITHMS =====
  algorithms: {
    all: ["algorithms"] as const,
    lists: () => [...queryKeys.algorithms.all, "list"] as const,
    list: (filters?: AlgorithmQuery) =>
      createQueryKey(queryKeys.algorithms.lists(), filters),
    details: () => [...queryKeys.algorithms.all, "detail"] as const,
    detail: (algorithmId: string) =>
      [...queryKeys.algorithms.details(), algorithmId] as const,
    runs: (algorithmId: string) =>
      [...queryKeys.algorithms.detail(algorithmId), "runs"] as const,
  },

  // ===== ALGORITHM RUNS =====
  runs: {
    all: ["runs"] as const,
    lists: () => [...queryKeys.runs.all, "list"] as const,
    list: (filters?: Record<string, any>) =>
      createQueryKey(queryKeys.runs.lists(), filters),
    details: () => [...queryKeys.runs.all, "detail"] as const,
    detail: (runId: string) => [...queryKeys.runs.details(), runId] as const,
    outputs: (runId: string) =>
      [...queryKeys.runs.detail(runId), "outputs"] as const,
  },

  // ===== EMAIL TEMPLATES =====
  emailTemplates: {
    all: ["emailTemplates"] as const,
    lists: () => [...queryKeys.emailTemplates.all, "list"] as const,
    list: (filters?: EmailTemplateQuery) =>
      createQueryKey(queryKeys.emailTemplates.lists(), filters),
    details: () => [...queryKeys.emailTemplates.all, "detail"] as const,
    detail: (templateId: string) =>
      [...queryKeys.emailTemplates.details(), templateId] as const,
  },

  // ===== ROLES & PERMISSIONS =====
  roles: {
    all: ["roles"] as const,
    lists: () => [...queryKeys.roles.all, "list"] as const,
    list: (filters?: RoleQuery) =>
      createQueryKey(queryKeys.roles.lists(), filters),
    details: () => [...queryKeys.roles.all, "detail"] as const,
    detail: (roleName: string) =>
      [...queryKeys.roles.details(), roleName] as const,
    users: (roleName: string) =>
      [...queryKeys.roles.detail(roleName), "users"] as const,
    permissions: (roleName: string) =>
      [...queryKeys.roles.detail(roleName), "permissions"] as const,
  },

  permissions: {
    all: ["permissions"] as const,
    lists: () => [...queryKeys.permissions.all, "list"] as const,
    list: (filters?: PermissionQuery) =>
      createQueryKey(queryKeys.permissions.lists(), filters),
    details: () => [...queryKeys.permissions.all, "detail"] as const,
    detail: (permissionName: string) =>
      [...queryKeys.permissions.details(), permissionName] as const,
  },

  // ===== GROUPS =====
  groups: {
    all: ["groups"] as const,
    lists: () => [...queryKeys.groups.all, "list"] as const,
    list: (filters?: GroupQuery) =>
      createQueryKey(queryKeys.groups.lists(), filters),
    details: () => [...queryKeys.groups.all, "detail"] as const,
    detail: (groupUid: string) =>
      [...queryKeys.groups.details(), groupUid] as const,
    users: (groupUid: string) =>
      [...queryKeys.groups.detail(groupUid), "users"] as const,
  },

  // ===== SETTINGS =====
  settings: {
    all: ["settings"] as const,
    lists: () => [...queryKeys.settings.all, "list"] as const,
    list: (filters?: SettingQuery) =>
      createQueryKey(queryKeys.settings.lists(), filters),
    details: () => [...queryKeys.settings.all, "detail"] as const,
    detail: (tenantId: number, key: string) =>
      [...queryKeys.settings.details(), tenantId, key] as const,
    count: () => [...queryKeys.settings.all, "count"] as const,
  },

  // ===== REGIONS =====
  regions: {
    all: ["regions"] as const,
    lists: () => [...queryKeys.regions.all, "list"] as const,
    list: (slideUid: string) =>
      [...queryKeys.regions.lists(), slideUid] as const,
    details: () => [...queryKeys.regions.all, "detail"] as const,
    detail: (regionUid: string) =>
      [...queryKeys.regions.details(), regionUid] as const,
  },
} as const;

/**
 * Type-safe query key invalidation patterns
 */
export const invalidationPatterns = {
  users: {
    all: () => queryKeys.users.all,
    byTenant: (tenantUid: string) => queryKeys.users.list({ tenantUid }),
  },
  tenants: {
    all: () => queryKeys.tenants.all,
  },
  studies: {
    all: () => queryKeys.studies.all,
    byTenant: (tenantUid: string) => queryKeys.studies.list({ tenantUid }),
    byCreator: (creatorUid: string) => queryKeys.studies.list({ creatorUid }),
  },
  cases: {
    all: () => queryKeys.cases.all,
    byStudy: (studyUid: string) => queryKeys.cases.list({ studyUid }),
    byTenant: (tenantUid: string) => queryKeys.cases.list({ tenantUid }),
  },
  slides: {
    all: () => queryKeys.slides.all,
    byCase: (caseUid: string) => queryKeys.slides.list({ caseUid }),
  },
  algorithms: {
    all: () => queryKeys.algorithms.all,
    byTenant: (tenantId: number) => queryKeys.algorithms.list({ tenantId }),
  },
  emailTemplates: {
    all: () => queryKeys.emailTemplates.all,
    byTenant: (tenantId: number) => queryKeys.emailTemplates.list({ tenantId }),
    byType: (templateType: string) =>
      queryKeys.emailTemplates.list({ templateType }),
  },
  roles: {
    all: () => queryKeys.roles.all,
  },
  permissions: {
    all: () => queryKeys.permissions.all,
    byCategory: (category: string) => queryKeys.permissions.list({ category }),
  },
  groups: {
    all: () => queryKeys.groups.all,
  },
  settings: {
    all: () => queryKeys.settings.all,
    byTenant: (tenantId: number) => queryKeys.settings.list({ tenantId }),
    byType: (valueType: "boolean" | "number" | "string" | "json") =>
      queryKeys.settings.list({ valueType }),
  },
} as const;
