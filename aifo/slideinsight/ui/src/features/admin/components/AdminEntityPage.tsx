// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useMemo } from "react";
import { ColumnDef } from "@tanstack/react-table";
import AdminPageLayout from "./AdminPageLayout";
import DataListPage from "../../../components/DataListPage";
import AdminModalManager from "../../../components/AdminModalManager";
import ErrorStateAlert from "../../../components/ErrorStateAlert";
import { Button } from "../../../components/ui/button";
import { PlusIcon } from "../../../components/icons";
import { FilterField } from "../../../types/search";
import {
  AdminEntityPageState,
  AdminEntityFilters,
  AdminEntityConfig,
} from "../hooks/useAdminEntityPage";

export interface AdminEntityPageConfig<T, F extends AdminEntityFilters> {
  // Page metadata
  title: string;
  description: string;
  searchPlaceholder: string;
  emptyMessage: string;

  // Data
  entities: T[];
  loading: boolean;
  error: any;
  refetch: () => void;

  // Entity configuration (for modals and CRUD operations)
  entityConfig: AdminEntityConfig<T>;

  // Pagination (optional)
  pagination?: {
    totalPages: number;
    totalItems: number;
  };

  // Columns and filtering
  columns: ColumnDef<T>[];
  filterFields: FilterField[];

  // Form components
  CreateFormComponent: React.ComponentType<{
    onSuccess: (entity: T) => void;
    onCancel: () => void;
  }>;
  EditFormComponent: React.ComponentType<{
    onSuccess: (entity: T) => void;
    onCancel: () => void;
  }>;

  // Modal configuration
  modalMaxWidth?: "sm" | "md" | "lg" | "xl" | "2xl";
  deleteDescription?: (entity: T) => React.ReactNode;

  // Customization
  customActions?: React.ReactNode;
  onRowClick?: (entity: T) => void;
  disableCreate?: boolean;
  enablePagination?: boolean;

  // Statistics (optional)
  statistics?: React.ReactNode;
}

interface AdminEntityPageProps<T, F extends AdminEntityFilters> {
  config: AdminEntityPageConfig<T, F>;
  state: AdminEntityPageState<T, F>;
}

export function AdminEntityPage<T, F extends AdminEntityFilters>({
  config,
  state,
}: AdminEntityPageProps<T, F>) {
  const {
    title,
    description,
    searchPlaceholder,
    emptyMessage,
    entities,
    loading,
    error,
    refetch,
    entityConfig,
    pagination: serverPagination,
    columns,
    filterFields,
    CreateFormComponent,
    EditFormComponent,
    modalMaxWidth = "lg",
    deleteDescription,
    customActions,
    onRowClick,
    disableCreate = false,
    enablePagination = false,
    statistics,
  } = config;

  const {
    filters,
    updateFilter,
    clearFilters,
    hasActiveFilters,
    pagination,
    adminState,
    applyOptimisticUpdates,
    filterEntities,
    getRowId,
  } = state;

  // Apply optimistic updates and filtering
  const processedEntities = useMemo(() => {
    const withOptimisticUpdates = applyOptimisticUpdates(entities);
    return filterEntities(withOptimisticUpdates);
  }, [entities, applyOptimisticUpdates, filterEntities]);

  // Create wrapper functions for AdminModalManager
  const getEntityDisplayNameWrapper = useMemo(
    () => (entity: T) => {
      return entityConfig.getEntityDisplayName(entity);
    },
    [entityConfig]
  );

  const deleteDescriptionWrapper = useMemo(() => {
    return (
      deleteDescription ||
      ((entity: T) =>
        `Are you sure you want to delete this ${entityConfig.entityNameSingular.toLowerCase()}? This action cannot be undone.`)
    );
  }, [deleteDescription, entityConfig]);

  // Early return for error state
  if (error) {
    return (
      <AdminPageLayout title={title} description={description}>
        <ErrorStateAlert
          error={error}
          title={`Error loading ${title.toLowerCase()}`}
          onRetry={refetch}
          variant="detailed"
        />
      </AdminPageLayout>
    );
  }

  return (
    <AdminPageLayout
      title={title}
      description={description}
      actions={
        <div className="flex gap-2">
          {customActions}
          {!disableCreate && (
            <Button onClick={adminState.handleAddEntity}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add {entityConfig.entityNameSingular}
            </Button>
          )}
        </div>
      }
    >
      <div className="space-y-6">
        {/* Statistics */}
        {statistics}

        <DataListPage<T>
          title=""
          subtitle=""
          data={processedEntities}
          loading={loading}
          error={null} // Handled above
          columns={columns}
          getRowId={getRowId}
          onRowClick={onRowClick}
          searchQuery={filters.searchQuery}
          onSearchQueryChange={(query) => updateFilter("searchQuery", query)}
          searchPlaceholder={searchPlaceholder}
          filterFields={filterFields}
          hasActiveFilters={hasActiveFilters}
          onClearFilters={clearFilters}
          emptyMessage={emptyMessage}
          pagination={enablePagination}
          pageSize={enablePagination ? pagination.pageSize : undefined}
          currentPage={enablePagination ? pagination.currentPage : undefined}
          totalPages={
            enablePagination ? serverPagination?.totalPages || 0 : undefined
          }
          totalItems={
            enablePagination ? serverPagination?.totalItems || 0 : undefined
          }
          onPageChange={
            enablePagination ? pagination.handlePageChange : undefined
          }
          onPageSizeChange={
            enablePagination ? pagination.handlePageSizeChange : undefined
          }
        />
      </div>

      {/* Common Admin Modals */}
      <AdminModalManager<T>
        // Create Modal
        isCreateModalOpen={adminState.isCreateModalOpen}
        onCreateModalClose={() => adminState.setIsCreateModalOpen(false)}
        CreateFormComponent={CreateFormComponent}
        // Edit Modal
        isEditModalOpen={adminState.isEditModalOpen}
        onEditModalClose={() => adminState.setIsEditModalOpen(false)}
        selectedEntity={adminState.selectedEntity}
        EditFormComponent={EditFormComponent}
        // Delete Dialog
        isDeleteDialogOpen={adminState.isDeleteDialogOpen}
        onDeleteDialogOpenChange={adminState.setIsDeleteDialogOpen}
        entityToDelete={adminState.entityToDelete}
        isDeleting={adminState.isDeleting}
        onConfirmDelete={adminState.confirmDelete}
        onCancelDelete={adminState.cancelDelete}
        // Form handlers - create wrappers that match expected signatures
        onFormSuccess={() => {
          // The forms handle calling refetch internally, so we just need to close modals
          adminState.setIsCreateModalOpen(false);
          adminState.setIsEditModalOpen(false);
        }}
        onFormCancel={() => {
          // Just close the modals
          adminState.setIsCreateModalOpen(false);
          adminState.setIsEditModalOpen(false);
        }}
        // Configuration from entity config
        entityName={entityConfig.entityName}
        entityNameSingular={entityConfig.entityNameSingular}
        getEntityDisplayName={getEntityDisplayNameWrapper}
        deleteDescription={deleteDescriptionWrapper}
        modalMaxWidth={modalMaxWidth}
      />
    </AdminPageLayout>
  );
}
