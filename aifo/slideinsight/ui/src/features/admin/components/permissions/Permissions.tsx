// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo } from "react";
import {
  usePermissions,
  useDeletePermission,
  type PermissionQuery,
} from "../../../../api";
import type { Permission as NewPermission } from "../../../../api";
import { PermissionForm } from "../forms/PermissionForm";
import { SecurityIcon } from "../../../../components/icons";
import { Badge } from "../../../../components/ui/badge";
import { toast } from "sonner";
import {
  AdminEntityPage,
  type AdminEntityPageConfig,
} from "../AdminEntityPage";
import {
  useAdminEntityPage,
  type AdminEntityFilters,
} from "../../hooks/useAdminEntityPage";
import { createStandardAdminColumns } from "../../utils/adminColumnUtils";
import {
  type NameSearchFilters,
  initialNameSearchFilters,
} from "../../utils/adminFilters";
// Import the old Permission type with a different name for the form components
import type { Permission as OldPermission } from "../../../../types/permissions";
import { FilterField } from "../../../../types/search";

interface PermissionFilters extends NameSearchFilters {
  searchCategory: string;
}

function PermissionsAdminComponent() {
  // Initialize filters
  const initialFilters: PermissionFilters = {
    ...initialNameSearchFilters,
    searchCategory: "",
  };

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Permissions",
      entityNameSingular: "Permission",
      deleteEndpoint: (permission: NewPermission) =>
        `/api/v1/permissions/${permission.name}`,
      getEntityDisplayName: (permission: NewPermission) =>
        permission.name || "Unnamed Permission",
      getEntityId: (permission: NewPermission) => permission.name || "",
    }),
    []
  );

  // Set up the admin entity page state
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters,
      enableOptimisticUpdates: false,
    },
    () => refetch()
  );

  // Prepare query for the new API hooks
  const permissionQuery: PermissionQuery = useMemo(() => {
    const { searchQuery, searchName, searchCategory, ...otherFilters } =
      pageState.filters;
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: searchQuery,
      name: searchName || undefined,
      category: searchCategory || undefined,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch permissions data using the new typed API
  const {
    data: permissions,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = usePermissions(permissionQuery);

  // Delete mutation
  const deletePermission = useDeletePermission({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete permission:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Convert new Permission to old format for form compatibility
  const convertToOldFormat = useCallback(
    (permission: NewPermission): OldPermission => {
      return {
        id: 0, // Default value as not available in new format
        short_uid: permission.name, // Use name as short_uid since it's unique
        name: permission.name,
        description: permission.description || "",
        tenant_id: 0, // Default to system tenant since not available in new format
        created_at: permission.createdAt,
        updated_at: permission.updatedAt,
      };
    },
    []
  );

  // Convert old Permission to new format
  const convertToNewFormat = useCallback(
    (permission: OldPermission): NewPermission => {
      return {
        name: permission.name,
        displayName: permission.name, // Use name as displayName
        description: permission.description,
        category: undefined, // Default as not available in old format
        createdAt: permission.created_at,
        updatedAt: permission.updated_at,
      };
    },
    []
  );

  const handleViewPermission = useCallback((permission: NewPermission) => {
    // For now, just copy the permission name to clipboard
    navigator.clipboard.writeText(permission.name);
    toast.success("Permission name copied to clipboard");
  }, []);

  // Enhanced delete handler that uses the typed mutation
  const handleDeletePermission = useCallback(
    (permission: NewPermission) => {
      deletePermission.mutate(permission.name);
    },
    [deletePermission]
  );

  // Create permissions-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<NewPermission>({
        entityName: "Permission",
        titleConfig: {
          accessor: "name",
          header: "Permission",
          getTitle: (permission) =>
            permission.displayName || permission.name || "Unnamed Permission",
          getDescription: (permission) => permission.description || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        customColumns: [
          {
            id: "name",
            header: "System Name",
            cell: ({ row }) => {
              const permission = row.original;
              return (
                <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded">
                  {permission.name}
                </span>
              );
            },
            enableSorting: true,
          },
          {
            id: "category",
            header: "Category",
            cell: ({ row }) => {
              const permission = row.original;
              return permission.category ? (
                <Badge
                  variant="outline"
                  className="text-blue-700 bg-blue-50 border-blue-200"
                >
                  {permission.category}
                </Badge>
              ) : (
                <span className="text-muted-400 text-sm">No category</span>
              );
            },
            enableSorting: true,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewPermission,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeletePermission, // Use the enhanced handler
          customActions: [
            {
              label: "Copy permission name",
              onClick: (permission) =>
                navigator.clipboard.writeText(permission.name),
            },
          ],
        },
      }),
    [
      handleViewPermission,
      pageState.adminState.handleEditEntity,
      handleDeletePermission,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () => [
      {
        type: "text",
        key: "searchName",
        label: "Name",
        placeholder: "Filter by permission name...",
        value: pageState.filters.searchName,
        onChange: (value) => pageState.updateFilter("searchName", value),
      },
      {
        type: "text",
        key: "searchCategory",
        label: "Category",
        placeholder: "Filter by category...",
        value: pageState.filters.searchCategory,
        onChange: (value) => pageState.updateFilter("searchCategory", value),
      },
    ],
    [pageState.filters, pageState.updateFilter]
  );

  // Form components for modals
  const CreateFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (permission: NewPermission) => void;
      onCancel: () => void;
    }) => (
      <PermissionForm
        permission={undefined}
        onSubmit={(oldPermission: OldPermission) => {
          onSuccess(convertToNewFormat(oldPermission));
        }}
        onCancel={onCancel}
      />
    ),
    [convertToNewFormat]
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (permission: NewPermission) => void;
      onCancel: () => void;
    }) => (
      <PermissionForm
        permission={
          pageState.adminState.selectedEntity
            ? convertToOldFormat(pageState.adminState.selectedEntity)
            : undefined
        }
        onSubmit={(oldPermission: OldPermission) => {
          onSuccess(convertToNewFormat(oldPermission));
        }}
        onCancel={onCancel}
      />
    ),
    [
      pageState.adminState.selectedEntity,
      convertToOldFormat,
      convertToNewFormat,
    ]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<NewPermission, PermissionFilters> = {
    title: "Permissions",
    description: "Manage system permissions and access controls",
    searchPlaceholder: "Search permissions by name or description...",
    emptyMessage: "No permissions found.",

    entities: permissions,
    loading: loading || deletePermission.isPending,
    error: error || (deletePermission.error?.message ?? null),
    refetch,
    entityConfig: crudConfig,

    pagination: serverPagination
      ? {
          totalPages: serverPagination.totalPages,
          totalItems: serverPagination.total,
        }
      : undefined,

    columns,
    filterFields,

    CreateFormComponent,
    EditFormComponent,

    modalMaxWidth: "2xl",
    deleteDescription: (permission) =>
      `Are you sure you want to delete the permission "${permission.name}"? This action cannot be undone and will remove this permission from all roles.`,

    enablePagination: true,
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default PermissionsAdminComponent;
