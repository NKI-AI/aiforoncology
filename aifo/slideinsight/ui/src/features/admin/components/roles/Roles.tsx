// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo, useState } from "react";
import { useRoles, useDeleteRole, type RoleQuery } from "../../../../api";
import type { Role as NewRole } from "../../../../api";
import { RoleForm } from "../forms/RoleForm";
import { SecurityIcon, UsersIcon } from "../../../../components/icons";
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
import RolePermissionsManager from "./RolePermissionsManager";
import RoleUsersManager from "./RoleUsersManager";
// Import the old Role type with a different name for the form components
import type { Role as OldRole } from "../../../../types/roles";
import { Button } from "../../../../components/ui/button";
import { FilterField } from "../../../../types/search";

interface RoleFilters extends NameSearchFilters {}

function RolesAdminComponent() {
  const [selectedRoleForPermissions, setSelectedRoleForPermissions] =
    useState<NewRole | null>(null);
  const [isPermissionsModalOpen, setIsPermissionsModalOpen] = useState(false);
  const [selectedRoleForUsers, setSelectedRoleForUsers] =
    useState<NewRole | null>(null);
  const [isUsersModalOpen, setIsUsersModalOpen] = useState(false);

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Roles",
      entityNameSingular: "Role",
      deleteEndpoint: (role: NewRole) => `/api/v1/roles/${role.name}`,
      getEntityDisplayName: (role: NewRole) => role.name || "Unnamed Role",
      getEntityId: (role: NewRole) => role.short_uid || "",
    }),
    []
  );

  // Set up the admin entity page state
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: initialNameSearchFilters,
      enableOptimisticUpdates: false,
    },
    () => refetch()
  );

  // Prepare query for the new API hooks
  const roleQuery: RoleQuery = useMemo(() => {
    const { searchQuery, searchName, ...otherFilters } = pageState.filters;
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: searchQuery,
      name: searchName || undefined,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch roles data using the new typed API
  const {
    data: roles,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useRoles(roleQuery);

  // Delete mutation
  const deleteRole = useDeleteRole({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete role:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Convert new Role to old format for form compatibility
  const convertToOldFormat = useCallback((role: NewRole): OldRole => {
    return {
      id: 0, // Default value as not available in new format
      short_uid: role.short_uid,
      name: role.name,
      description: role.description || "",
      created_at: role.createdAt,
      updated_at: role.updatedAt,
    };
  }, []);

  // Convert old Role to new format
  const convertToNewFormat = useCallback((role: OldRole): NewRole => {
    return {
      name: role.name,
      short_uid: role.short_uid,
      displayName: role.name, // Use name as displayName
      description: role.description,
      createdAt: role.created_at,
      updatedAt: role.updated_at,
    };
  }, []);

  const handleViewRole = useCallback((role: NewRole) => {
    // For now, just copy the role name to clipboard
    navigator.clipboard.writeText(role.name);
    toast.success("Role name copied to clipboard");
  }, []);

  const handleManagePermissions = useCallback((role: NewRole) => {
    setSelectedRoleForPermissions(role);
    setIsPermissionsModalOpen(true);
  }, []);

  const handleManageUsers = useCallback((role: NewRole) => {
    setSelectedRoleForUsers(role);
    setIsUsersModalOpen(true);
  }, []);

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteRole = useCallback(
    (role: NewRole) => {
      deleteRole.mutate(role.name);
    },
    [deleteRole]
  );

  // Create roles-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<NewRole>({
        entityName: "Role",
        titleConfig: {
          accessor: "name",
          header: "Role",
          getTitle: (role) => role.displayName || role.name || "Unnamed Role",
          getDescription: (role) => role.description || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        includeId: true,
        idConfig: {
          accessor: "short_uid",
          header: "ID",
          maxLength: 8,
        },
        customColumns: [
          {
            id: "name",
            header: "System Name",
            cell: ({ row }) => {
              const role = row.original;
              return (
                <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded">
                  {role.name}
                </span>
              );
            },
            enableSorting: true,
          },
          {
            id: "permissions",
            header: "Permissions",
            cell: ({ row }) => {
              const role = row.original;
              return (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleManagePermissions(role)}
                  className="h-8 px-2 py-1"
                  title="Manage role permissions"
                >
                  <SecurityIcon className="h-4 w-4" />
                </Button>
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewRole,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteRole, // Use the enhanced handler
          customActions: [
            {
              label: "Manage Users",
              onClick: handleManageUsers,
              icon: UsersIcon,
            },
            {
              label: "Copy role name",
              onClick: (role) => navigator.clipboard.writeText(role.name),
            },
          ],
        },
      }),
    [
      handleViewRole,
      pageState.adminState.handleEditEntity,
      handleDeleteRole,
      handleManagePermissions,
      handleManageUsers,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () => [
      {
        type: "text",
        key: "searchName",
        label: "Name",
        placeholder: "Filter by role name...",
        value: pageState.filters.searchName,
        onChange: (value) => pageState.updateFilter("searchName", value),
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
      onSuccess: (role: NewRole) => void;
      onCancel: () => void;
    }) => (
      <RoleForm
        role={undefined}
        onSubmit={(oldRole: OldRole) => {
          onSuccess(convertToNewFormat(oldRole));
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
      onSuccess: (role: NewRole) => void;
      onCancel: () => void;
    }) => (
      <RoleForm
        role={
          pageState.adminState.selectedEntity
            ? convertToOldFormat(pageState.adminState.selectedEntity)
            : undefined
        }
        onSubmit={(oldRole: OldRole) => {
          onSuccess(convertToNewFormat(oldRole));
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
  const pageConfig: AdminEntityPageConfig<NewRole, RoleFilters> = {
    title: "Roles",
    description: "Manage user roles and permissions",
    searchPlaceholder: "Search roles by name or description...",
    emptyMessage: "No roles found.",

    entities: roles,
    loading: loading || deleteRole.isPending,
    error: error || (deleteRole.error?.message ?? null),
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
    deleteDescription: (role) =>
      `Are you sure you want to delete the role "${role.name}"? This action cannot be undone and will remove this role from all users.`,

    enablePagination: true,
  };

  return (
    <>
      <AdminEntityPage config={pageConfig} state={pageState} />

      {/* Role Permissions Manager Modal */}
      {selectedRoleForPermissions && (
        <RolePermissionsManager
          isOpen={isPermissionsModalOpen}
          onClose={() => {
            setIsPermissionsModalOpen(false);
            setSelectedRoleForPermissions(null);
          }}
          role={convertToOldFormat(selectedRoleForPermissions)}
        />
      )}

      {/* Role Users Manager Modal */}
      {selectedRoleForUsers && (
        <RoleUsersManager
          isOpen={isUsersModalOpen}
          onClose={() => {
            setIsUsersModalOpen(false);
            setSelectedRoleForUsers(null);
          }}
          role={selectedRoleForUsers}
        />
      )}
    </>
  );
}

export default RolesAdminComponent;
