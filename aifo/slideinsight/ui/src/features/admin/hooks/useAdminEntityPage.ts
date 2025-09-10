// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useMemo, useCallback } from "react";
import { useServerSideFilters } from "../components/useServerSideFilters";
import { useAdminState } from "../../../hooks/useAdminState";
import { usePaginationState } from "../../../hooks/usePaginationState";

export interface AdminEntityConfig<T> {
  entityName: string;
  entityNameSingular: string;
  deleteEndpoint: (entity: T) => string;
  getEntityDisplayName: (entity: T) => string;
  getEntityId: (entity: T) => string;
}

export interface AdminEntityFilters {
  searchQuery: string;
  [key: string]: any;
}

interface AdminEntityPageConfig<T, F extends AdminEntityFilters> {
  crudConfig: AdminEntityConfig<T>;
  initialFilters: F;
  initialPageSize?: number;
  enableOptimisticUpdates?: boolean;
  optimisticUpdateFields?: string[];
}

export interface AdminEntityPageState<T, F extends AdminEntityFilters> {
  // Filter state
  filters: F;
  updateFilter: (key: keyof F, value: any) => void;
  clearFilters: () => void;
  hasActiveFilters: boolean;

  // Pagination state
  pagination: {
    currentPage: number;
    pageSize: number;
    handlePageChange: (page: number) => void;
    handlePageSizeChange: (size: number) => void;
  };

  // Admin state (modals, delete management)
  adminState: {
    isCreateModalOpen: boolean;
    setIsCreateModalOpen: (open: boolean) => void;
    isEditModalOpen: boolean;
    setIsEditModalOpen: (open: boolean) => void;
    isDeleteDialogOpen: boolean;
    setIsDeleteDialogOpen: (open: boolean) => void;
    selectedEntity: T | null;
    entityToDelete: T | null;
    isDeleting: boolean;
    handleAddEntity: () => void;
    handleEditEntity: (entity: T) => void;
    handleDeleteEntity: (entity: T) => void;
    handleFormSuccess: (entity: T) => void;
    handleFormCancel: () => void;
    confirmDelete: () => Promise<void>;
    cancelDelete: () => void;
  };

  // Optimistic updates
  optimisticUpdates: Record<string, any>;
  updateOptimistic: (entityId: string, field: string, value: any) => void;
  clearOptimistic: (entityId: string) => void;

  // Utilities
  applyOptimisticUpdates: (entities: T[]) => T[];
  filterEntities: (entities: T[]) => T[];
  getRowId: (entity: T, index: number) => string;
}

export function useAdminEntityPage<T, F extends AdminEntityFilters>(
  config: AdminEntityPageConfig<T, F>,
  refetch: () => void
): AdminEntityPageState<T, F> {
  const {
    crudConfig,
    initialFilters,
    initialPageSize = 20,
    enableOptimisticUpdates = false,
    optimisticUpdateFields = [],
  } = config;

  // Filter state
  const { filters, updateFilter, clearFilters, hasActiveFilters } =
    useServerSideFilters(
      () => {}, // Local filtering for now
      initialFilters
    );

  // Pagination state
  const pagination = usePaginationState(1, initialPageSize);

  // Admin state (modals, delete management)
  const adminState = useAdminState(crudConfig, refetch);

  // Optimistic updates state
  const [optimisticUpdates, setOptimisticUpdates] = useState<
    Record<string, any>
  >({});

  // Update optimistic state
  const updateOptimistic = useCallback(
    (entityId: string, field: string, value: any) => {
      if (!enableOptimisticUpdates || !optimisticUpdateFields.includes(field)) {
        return;
      }

      setOptimisticUpdates((prev) => ({
        ...prev,
        [entityId]: {
          ...prev[entityId],
          [field]: value,
        },
      }));
    },
    [enableOptimisticUpdates, optimisticUpdateFields]
  );

  // Clear optimistic updates for an entity
  const clearOptimistic = useCallback((entityId: string) => {
    setOptimisticUpdates((prev) => {
      const newState = { ...prev };
      delete newState[entityId];
      return newState;
    });
  }, []);

  // Apply optimistic updates to entities
  const applyOptimisticUpdates = useCallback(
    (entities: T[]) => {
      if (!enableOptimisticUpdates) {
        return entities;
      }

      return entities.map((entity) => {
        const entityId = crudConfig.getEntityId(entity);
        const updates = optimisticUpdates[entityId];
        return updates ? { ...entity, ...updates } : entity;
      });
    },
    [enableOptimisticUpdates, optimisticUpdates, crudConfig]
  );

  // Filter entities based on current filters
  const filterEntities = useCallback(
    (entities: T[]) => {
      return entities.filter((entity) => {
        // Always check searchQuery if it exists
        if (filters.searchQuery) {
          const searchableFields = getSearchableFields(entity);
          const matchesSearch = searchableFields.some((field) =>
            field?.toLowerCase().includes(filters.searchQuery.toLowerCase())
          );
          if (!matchesSearch) return false;
        }

        // Check other filters
        return Object.entries(filters).every(([key, value]) => {
          if (key === "searchQuery" || !value) return true;
          return matchesEntityFilter(entity, key, value);
        });
      });
    },
    [filters]
  );

  // Generate row ID for table
  const getRowId = useCallback(
    (entity: T, index: number) => {
      const id = crudConfig.getEntityId(entity);
      return id || `${crudConfig.entityName.toLowerCase()}-${index}`;
    },
    [crudConfig]
  );

  return {
    filters,
    updateFilter,
    clearFilters,
    hasActiveFilters,
    pagination,
    adminState,
    optimisticUpdates,
    updateOptimistic,
    clearOptimistic,
    applyOptimisticUpdates,
    filterEntities,
    getRowId,
  };
}

// Helper function to extract searchable fields from an entity
function getSearchableFields(entity: any): string[] {
  const fields: string[] = [];

  // Common searchable fields
  if (entity.name) fields.push(entity.name);
  if (entity.email) fields.push(entity.email);
  if (entity.description) fields.push(entity.description);
  if (entity.firstName && entity.lastName) {
    fields.push(`${entity.firstName} ${entity.lastName}`);
  }
  if (entity.slideName) fields.push(entity.slideName);
  if (entity.metadata) fields.push(entity.metadata);

  return fields;
}

// Helper function to match entity against filter
function matchesEntityFilter(
  entity: any,
  filterKey: string,
  filterValue: any
): boolean {
  switch (filterKey) {
    case "searchName":
      return (
        !filterValue ||
        entity.name?.toLowerCase().includes(filterValue.toLowerCase())
      );
    case "searchEmail":
      return (
        !filterValue ||
        entity.email?.toLowerCase().includes(filterValue.toLowerCase())
      );
    case "searchDescription":
      return (
        !filterValue ||
        entity.description?.toLowerCase().includes(filterValue.toLowerCase())
      );
    case "searchStatus":
      // Handle different status patterns
      if (entity.status) {
        return entity.status === filterValue;
      }
      if (entity.isActive !== undefined) {
        return (
          (filterValue === "active" && entity.isActive) ||
          (filterValue === "inactive" && !entity.isActive)
        );
      }
      if (entity.isPublished !== undefined) {
        return (
          (filterValue === "published" && entity.isPublished) ||
          (filterValue === "draft" && !entity.isPublished)
        );
      }
      return true;
    default:
      return true;
  }
}
