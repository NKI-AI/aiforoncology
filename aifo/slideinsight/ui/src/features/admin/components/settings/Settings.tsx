// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo, useState } from "react";
import {
  useSettings,
  useDeleteSetting,
  type Setting,
  type SettingQuery,
} from "../../../../api";
import { TrashIcon } from "../../../../components/icons";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
} from "../../../../components/table-helpers/columnHelpers";
import { CreateSettingForm } from "./CreateSettingForm";
import { EditSettingForm } from "./EditSettingForm";
import {
  AdminEntityPage,
  type AdminEntityPageConfig,
} from "../AdminEntityPage";
import {
  useAdminEntityPage,
  type AdminEntityFilters,
} from "../../hooks/useAdminEntityPage";
import {
  type CommonSettingsFilters,
  initialCommonSettingsFilters,
  createSettingsFilterFields,
} from "../../utils/adminFilters";
import { Button } from "../../../../components/ui/button";
import { Badge } from "../../../../components/ui/badge";
import { FilterField } from "../../../../types/search";

// Enhanced filters interface that extends the base filters with API types
interface EnhancedSettingsFilters extends CommonSettingsFilters {
  q?: string;
  key?: string;
  valueType?: "boolean" | "number" | "string" | "json";
  tenantId?: number;
}

function SettingsAdmin() {
  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Settings",
      entityNameSingular: "Setting",
      deleteEndpoint: (setting: Setting) =>
        `/api/v1/settings/${setting.tenantId}/${encodeURIComponent(
          setting.key
        )}`,
      getEntityDisplayName: (setting: Setting) =>
        setting.key || "Unknown Setting",
      getEntityId: (setting: Setting) => `${setting.tenantId}-${setting.key}`,
    }),
    []
  );

  // Set up the admin entity page state with optimistic updates disabled for settings
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: {
        ...initialCommonSettingsFilters,
        q: "",
        key: "",
        valueType: undefined,
        tenantId: undefined,
      } as EnhancedSettingsFilters,
      enableOptimisticUpdates: false, // Settings don't have boolean toggles like users
    },
    () => {} // Refetch will be handled by the new API hooks
  );

  // Prepare query for the new API hooks
  const settingQuery: SettingQuery = useMemo(() => {
    const {
      q,
      key,
      valueType,
      tenantId,
      searchKey,
      searchValueType,
      searchTenantId,
      ...otherFilters
    } = pageState.filters;

    return {
      page: 1,
      limit: 20,
      q: q || "",
      key: key || searchKey || undefined,
      valueType: valueType || (searchValueType as any) || undefined,
      tenantId:
        tenantId || (searchTenantId ? parseInt(searchTenantId) : undefined),
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters]);

  // Fetch settings data using the new typed API
  const {
    data: settings,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useSettings(settingQuery);

  // Delete mutation
  const deleteSetting = useDeleteSetting({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete setting:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteSetting = useCallback(
    (setting: Setting) => {
      deleteSetting.mutate({ tenantId: setting.tenantId, key: setting.key });
    },
    [deleteSetting]
  );

  // Helper function to format value based on type
  const formatSettingValue = (setting: Setting): string => {
    const { value, valueType } = setting;
    if (!value) return "—";

    switch (valueType) {
      case "boolean":
        return value === "true" ? "True" : "False";
      case "number":
        return value;
      case "string":
        return value.length > 50 ? `${value.substring(0, 50)}...` : value;
      case "json":
        try {
          const parsed = JSON.parse(value);
          const stringified = JSON.stringify(parsed, null, 0);
          return stringified.length > 50
            ? `${stringified.substring(0, 50)}...`
            : stringified;
        } catch {
          return value.length > 50 ? `${value.substring(0, 50)}...` : value;
        }
      default:
        return value;
    }
  };

  // Helper function to get value type badge variant
  const getValueTypeBadgeVariant = (valueType: string) => {
    switch (valueType) {
      case "boolean":
        return "default";
      case "number":
        return "secondary";
      case "string":
        return "outline";
      case "json":
        return "destructive";
      default:
        return "outline";
    }
  };

  // Table columns configuration using the new helpers
  const columns = useMemo(
    () => [
      createSelectionColumn<Setting>(),

      createTitleDescriptionColumn<Setting>({
        accessor: "key",
        header: "Setting",
        getTitle: (setting) => setting.key || "Unknown Key",
        getDescription: (setting) =>
          `Tenant: ${setting.tenantId} | Value: ${formatSettingValue(setting)}`,
        sortable: true,
      }),

      {
        id: "valueType",
        header: "Type",
        cell: ({ row }) => {
          const setting = row.original;
          return (
            <Badge variant={getValueTypeBadgeVariant(setting.valueType) as any}>
              {setting.valueType}
            </Badge>
          );
        },
        enableSorting: false,
      },

      {
        id: "tenantId",
        header: "Tenant ID",
        cell: ({ row }) => {
          const setting = row.original;
          return (
            <div className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
              {setting.tenantId}
            </div>
          );
        },
        enableSorting: true,
      },

      createDateColumn<Setting>({
        accessor: "createdAt",
        header: "Created",
        sortable: true,
      }),

      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
          const setting = row.original;
          return (
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={(event) => {
                  event.stopPropagation();
                  pageState.adminState.handleEditEntity(setting);
                }}
                className="h-8 px-3 py-1"
                title="Edit setting"
              >
                Edit
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={(event) => {
                  event.stopPropagation();
                  pageState.adminState.handleDeleteEntity(setting);
                }}
                className="h-8 px-3 py-1"
                title="Delete setting"
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            </div>
          );
        },
        enableSorting: false,
      },
    ],
    [pageState.adminState, formatSettingValue]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createSettingsFilterFields({
        filters: pageState.filters,
        updateFilter: pageState.updateFilter,
      }),
    [pageState.filters, pageState.updateFilter]
  );

  // Form components for modals - creating wrappers to bridge interface differences
  const CreateFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (entity: Setting) => void;
      onCancel: () => void;
    }) => (
      <CreateSettingForm
        onSuccess={() => {
          // Since we don't have the created setting from the form,
          // we'll refetch and call onSuccess with a dummy setting
          refetch();
          // AdminEntityPage expects an entity parameter, but our form doesn't provide one
          // We'll pass a placeholder that won't be used since we're refetching
          onSuccess({} as Setting);
        }}
        onCancel={onCancel}
      />
    ),
    [refetch]
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (entity: Setting) => void;
      onCancel: () => void;
    }) => (
      <EditSettingForm
        setting={pageState.adminState.selectedEntity as Setting}
        onSuccess={() => {
          // Since we don't have the updated setting from the form,
          // we'll refetch and call onSuccess with the current setting
          refetch();
          onSuccess(pageState.adminState.selectedEntity as Setting);
        }}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity, refetch]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<Setting, EnhancedSettingsFilters> = {
    title: "Settings",
    description: "Manage system and tenant-specific configuration settings",
    searchPlaceholder: "Search settings by key or value...",
    emptyMessage: "No settings found.",

    entities: settings, // Use the typed data directly
    loading: loading || deleteSetting.isPending,
    error: error || (deleteSetting.error?.message ?? null),
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

    modalMaxWidth: "lg",
    deleteDescription: (setting) => (
      <div>
        Are you sure you want to permanently delete the setting "{setting.key}"
        for tenant {setting.tenantId}?
        <br />
        <br />
        <div className="font-semibold text-red-600">
          This action is permanent and cannot be undone.
        </div>
        <br />
        <br />
        <div className="text-sm text-muted-foreground">
          Current value: {formatSettingValue(setting)}
        </div>
      </div>
    ),

    enablePagination: true,
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default SettingsAdmin;
