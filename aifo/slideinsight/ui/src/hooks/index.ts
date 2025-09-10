// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Centralized exports for all custom hooks
 */

// ===== CORE HOOK TYPES =====
export type * from "./types";

// ===== UTILITY HOOKS =====
export { useMobile } from "./useMobile";
export { useToggle } from "./useToggle";
export { useAsyncOperation } from "./useAsyncOperation";
export { useKeyboardShortcuts } from "./useKeyboardShortcuts";
export { usePanels } from "./usePanels";

// ===== STATE MANAGEMENT HOOKS =====
export { useModalState } from "./useModalState";
export { usePaginationState } from "./usePaginationState";
export { useAdminState } from "./useAdminState";

// ===== API & DATA HOOKS =====
export { usePaginatedApi } from "./usePaginatedApi";
export { createPaginatedHook } from "./createPaginatedHook";
export { useDeleteEntity } from "./useDeleteEntity";

// ===== DOMAIN-SPECIFIC HOOKS =====
export { useStudies } from "./useStudies";
export { useStudiesTableState } from "./useStudiesTableState";
export { useSlides } from "./useSlides";
export { useCases } from "./useCases";

// ===== COMMUNICATION HOOKS =====
export { useNotifications } from "./useNotifications";
export { useSystemMonitor } from "./useSystemMonitor";
export { useWebSocketManager } from "./useWebSocketManager";
