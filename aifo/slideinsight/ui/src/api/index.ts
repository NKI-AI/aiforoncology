// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Centralized API exports for type-safe API operations
 */

// Export all types
export * from "./models";

// Export query keys
export { queryKeys, invalidationPatterns } from "./queryKeys";

// Export API hooks
export {
  useApiQuery,
  useApiMutation,
  type PaginatedApiResult,
  // User hooks
  useUsers,
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  // Tenant hooks
  useTenants,
  useCreateTenant,
  useUpdateTenant,
  useDeleteTenant,
  // Study hooks
  useStudies,
  useCreateStudy,
  useUpdateStudy,
  useDeleteStudy,
  // Case hooks
  useCases,
  useCreateCase,
  useUpdateCase,
  useDeleteCase,
  // Slide hooks
  useSlides,
  useCreateSlide,
  useUpdateSlide,
  useDeleteSlide,
  // Algorithm hooks
  useAlgorithms,
  useCreateAlgorithm,
  useUpdateAlgorithm,
  useDeleteAlgorithm,
  // Email Template hooks
  useEmailTemplates,
  useCreateEmailTemplate,
  useUpdateEmailTemplate,
  useDeleteEmailTemplate,
  // Settings hooks
  useSettings,
  useSetting,
  useCreateSetting,
  useUpdateSetting,
  useDeleteSetting,
  useSettingsCount,
  // Role hooks
  useRoles,
  useCreateRole,
  useUpdateRole,
  useDeleteRole,
  // Permission hooks
  usePermissions,
  useCreatePermission,
  useUpdatePermission,
  useDeletePermission,
  // Group hooks
  useGroups,
  useCreateGroup,
  useUpdateGroup,
  useDeleteGroup,
} from "./hooks";
