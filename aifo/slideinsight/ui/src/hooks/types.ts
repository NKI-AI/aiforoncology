// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Common hook return types for consistent interfaces across the application
 */

// ===== PAGINATION TYPES =====

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface PaginatedResult<T> {
  data: T[];
  pagination: PaginationInfo;
  loading: boolean;
  error: string | null;
  refetch: () => void;
  isStale?: boolean;
  isFetching?: boolean;
}

export interface PaginationState {
  currentPage: number;
  pageSize: number;
  handlePageChange: (page: number) => void;
  handlePageSizeChange: (newPageSize: number) => void;
  setCurrentPage: (page: number) => void;
  setPageSize: (size: number) => void;
}

// ===== MODAL STATE TYPES =====

export interface ModalState<T> {
  // Modal states
  isCreateModalOpen: boolean;
  isEditModalOpen: boolean;
  selectedEntity: T | null;

  // Handlers
  handleAddEntity: () => void;
  handleEditEntity: (entity: T) => void;
  handleFormSuccess: () => void;
  handleFormCancel: () => void;

  // Setters for manual control
  setIsCreateModalOpen: (open: boolean) => void;
  setIsEditModalOpen: (open: boolean) => void;
  setSelectedEntity: (entity: T | null) => void;
}

// ===== API STATE TYPES =====

export interface ApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export interface ApiListState<T> {
  data: T[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

// ===== FORM STATE TYPES =====

export interface FormState<T> {
  data: T;
  loading: boolean;
  error: string | null;
  fieldErrors: Record<string, string>;
  success: boolean;
  updateField: (field: keyof T) => (value: any) => void;
  setData: (data: T) => void;
  setFieldError: (field: keyof T, error: string) => void;
  clearFieldError: (field: keyof T) => void;
  clearAllErrors: () => void;
  reset: () => void;
}

// ===== ASYNC OPERATION TYPES =====

export interface AsyncOperationState {
  loading: boolean;
  error: string | null;
  success: boolean;
}

export interface AsyncOperationResult<T = any> extends AsyncOperationState {
  data: T | null;
  execute: (...args: any[]) => Promise<void>;
  reset: () => void;
}

// ===== SEARCH/FILTER TYPES =====

export interface SearchState {
  query: string;
  filters: Record<string, any>;
  setQuery: (query: string) => void;
  setFilter: (key: string, value: any) => void;
  removeFilter: (key: string) => void;
  clearFilters: () => void;
  clearAll: () => void;
}

// ===== SELECTION STATE TYPES =====

export interface SelectionState<T> {
  selectedItems: T[];
  isSelected: (item: T) => boolean;
  selectItem: (item: T) => void;
  deselectItem: (item: T) => void;
  toggleItem: (item: T) => void;
  selectAll: (items: T[]) => void;
  deselectAll: () => void;
  selectedCount: number;
}

// ===== TABLE STATE TYPES =====

export interface TableState {
  sortField: string | null;
  sortDirection: "asc" | "desc";
  setSorting: (field: string, direction?: "asc" | "desc") => void;
  clearSorting: () => void;
}

export interface TableStateWithPagination extends TableState, PaginationState {}

// ===== WEBSOCKET STATE TYPES =====

export interface WebSocketState {
  connected: boolean;
  connecting: boolean;
  error: string | null;
  lastMessage: any;
  send: (message: any) => void;
  connect: () => void;
  disconnect: () => void;
}

// ===== UI STATE TYPES =====

export interface ToggleState {
  isOpen: boolean;
  open: () => void;
  close: () => void;
  toggle: () => void;
}

export interface LoadingState {
  loading: boolean;
  setLoading: (loading: boolean) => void;
}

export interface ErrorState {
  error: string | null;
  setError: (error: string | null) => void;
  clearError: () => void;
}

// ===== KEYBOARD SHORTCUT TYPES =====

export interface KeyboardShortcutState {
  shortcuts: Record<string, () => void>;
  addShortcut: (key: string, action: () => void) => void;
  removeShortcut: (key: string) => void;
  clearShortcuts: () => void;
}

// ===== RESPONSIVE DESIGN TYPES =====

export interface ResponsiveState {
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isInitialized: boolean;
  breakpoint: "mobile" | "tablet" | "desktop";
}

export interface UseMobileResult {
  /** Whether the current viewport is considered mobile (< 768px) */
  isMobile: boolean;
  /** Whether the hook has initialized and determined the viewport size */
  isInitialized: boolean;
}

// ===== COMMON BASE TYPES =====

export interface BaseApiOptions {
  page?: number;
  limit?: number;
  q?: string;
  sort?: string;
  dir?: "asc" | "desc";
}

export interface RefetchableState {
  refetch: () => void;
  isRefetching?: boolean;
  lastRefetched?: Date;
}
