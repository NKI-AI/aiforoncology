// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo, useState } from "react";
import {
  useGroups,
  useDeleteGroup,
  type Group,
  type GroupQuery,
} from "../../../../api";
import { GroupForm } from "../forms/GroupForm";
import { UsersIcon } from "../../../../components/icons";
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
import GroupUsersManager from "./GroupUsersManager";
import { FilterField } from "../../../../types/search";

interface GroupFilters extends NameSearchFilters {}

function GroupsAdminComponent() {
  const [selectedGroupForUsers, setSelectedGroupForUsers] =
    useState<Group | null>(null);
  const [isUsersModalOpen, setIsUsersModalOpen] = useState(false);

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Groups",
      entityNameSingular: "Group",
      deleteEndpoint: (group: Group) => `/api/v1/groups/${group.short_uid}`,
      getEntityDisplayName: (group: Group) => group.name || "Unnamed Group",
      getEntityId: (group: Group) => group.short_uid || "",
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
  const groupQuery: GroupQuery = useMemo(() => {
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

  // Fetch groups data using the new typed API
  const {
    data: groups,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useGroups(groupQuery);

  // Delete mutation
  const deleteGroup = useDeleteGroup({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete group:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  const handleViewGroup = useCallback((group: Group) => {
    // For now, just copy the group name to clipboard
    navigator.clipboard.writeText(group.name);
    toast.success("Group name copied to clipboard");
  }, []);

  const handleManageUsers = useCallback((group: Group) => {
    setSelectedGroupForUsers(group);
    setIsUsersModalOpen(true);
  }, []);

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteGroup = useCallback(
    (group: Group) => {
      deleteGroup.mutate(group.short_uid);
    },
    [deleteGroup]
  );

  // Create groups-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<Group>({
        entityName: "Group",
        titleConfig: {
          accessor: "name",
          header: "Group",
          getTitle: (group) =>
            group.displayName || group.name || "Unnamed Group",
          getDescription: (group) => group.description || null,
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
              const group = row.original;
              return (
                <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded">
                  {group.name}
                </span>
              );
            },
            enableSorting: true,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewGroup,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteGroup, // Use the enhanced handler
          customActions: [
            {
              label: "Manage Users",
              onClick: handleManageUsers,
              icon: UsersIcon,
            },
            {
              label: "Copy group name",
              onClick: (group) =>
                navigator.clipboard.writeText(group.name || ""),
            },
            {
              label: "Copy group UID",
              onClick: (group) =>
                navigator.clipboard.writeText(group.short_uid || ""),
            },
          ],
        },
      }),
    [
      handleViewGroup,
      pageState.adminState.handleEditEntity,
      handleDeleteGroup,
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
        placeholder: "Filter by group name...",
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
      onSuccess: (group: Group) => void;
      onCancel: () => void;
    }) => (
      <GroupForm group={undefined} onSubmit={onSuccess} onCancel={onCancel} />
    ),
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (group: Group) => void;
      onCancel: () => void;
    }) => (
      <GroupForm
        group={pageState.adminState.selectedEntity || undefined}
        onSubmit={onSuccess}
        onCancel={onCancel}
        isEditing={true}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<Group, GroupFilters> = {
    title: "Groups",
    description: "Manage user groups and team organization",
    searchPlaceholder: "Search groups by name or description...",
    emptyMessage: "No groups found.",

    entities: groups,
    loading: loading || deleteGroup.isPending,
    error: error || (deleteGroup.error?.message ?? null),
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
    deleteDescription: (group) =>
      `Are you sure you want to delete the group "${group.name}"? This action cannot be undone and will remove all users from this group.`,

    enablePagination: true,
  };

  return (
    <>
      <AdminEntityPage config={pageConfig} state={pageState} />

      {/* Group Users Manager Modal */}
      {selectedGroupForUsers && (
        <GroupUsersManager
          isOpen={isUsersModalOpen}
          onClose={() => {
            setIsUsersModalOpen(false);
            setSelectedGroupForUsers(null);
          }}
          group={selectedGroupForUsers}
        />
      )}
    </>
  );
}

export default GroupsAdminComponent;
