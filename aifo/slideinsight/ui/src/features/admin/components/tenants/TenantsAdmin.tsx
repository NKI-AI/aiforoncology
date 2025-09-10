// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useCallback, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  useTenants,
  useDeleteTenant,
  type Tenant,
  type TenantQuery,
} from "../../../../api";
import TenantDomainsModal from "./TenantDomainsModal";
import TenantDomainsCell from "./TenantDomainsCell";
import { TrashIcon } from "../../../../components/icons";
import AdminPageLayout from "../AdminPageLayout";
import DataListPage from "../../../../components/DataListPage";
import AdminModalManager from "../../../../components/AdminModalManager";
import ErrorStateAlert from "../../../../components/ErrorStateAlert";
import { useServerSideFilters } from "../useServerSideFilters";
import { useAdminState } from "../../../../hooks/useAdminState";
import { usePaginationState } from "../../../../hooks/usePaginationState";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
  createIdColumn,
  createActionsColumn,
} from "../../../../components/table-helpers/columnHelpers";
import { TenantStatusCell } from "../../../../components/ui/TenantStatusCell";
import { Button } from "../../../../components/ui/button";
import { PlusIcon } from "../../../../components/icons";
import { TenantForm } from "../forms";
import {
  AdminEntityPage,
  type AdminEntityPageConfig,
} from "../AdminEntityPage";
import {
  useAdminEntityPage,
  type AdminEntityFilters,
} from "../../hooks/useAdminEntityPage";
import { createTenantEntityColumns } from "../../utils/adminColumnUtils";
import {
  type NameSearchFilters,
  type StatusSearchFilters,
  createInitialFilters,
  createTenantFilterFields,
} from "../../utils/adminFilters";
import { FilterField } from "../../../../types/search";

interface TenantFilters extends NameSearchFilters, StatusSearchFilters {
  q?: string;
  name?: string;
  status?: string;
}

function TenantsAdmin() {
  const navigate = useNavigate();

  // Separate state for tenant-specific modals
  const [isDomainsModalOpen, setIsDomainsModalOpen] = useState(false);
  const [selectedTenantForDomains, setSelectedTenantForDomains] =
    useState<Tenant | null>(null);

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Tenants",
      entityNameSingular: "Tenant",
      deleteEndpoint: (tenant: Tenant) => `/api/v1/tenants/${tenant.tenantUid}`,
      getEntityDisplayName: (tenant: Tenant) => tenant.name || "Unnamed Tenant",
      getEntityId: (tenant: Tenant) => tenant.tenantUid || "",
    }),
    []
  );

  // Set up the admin entity page state with optimistic updates for status
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: createInitialFilters<TenantFilters>({
        q: "",
        name: "",
        status: "",
      }),
      enableOptimisticUpdates: true,
      optimisticUpdateFields: ["status"],
    },
    () => {} // Refetch will be handled by the new API hooks
  );

  // Prepare query for the new API hooks
  const tenantQuery: TenantQuery = useMemo(() => {
    const { q, name, status, ...otherFilters } = pageState.filters;
    return {
      page: 1,
      limit: 20,
      q: q || "",
      name: name || undefined,
      status: status || undefined,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters]);

  // Fetch tenants data using the new typed API
  const {
    data: tenants,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useTenants(tenantQuery);

  // Delete mutation
  const deleteTenant = useDeleteTenant({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete tenant:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  const handleManageDomains = useCallback((tenant: Tenant) => {
    setSelectedTenantForDomains(tenant);
    setIsDomainsModalOpen(true);
  }, []);

  const handleDomainsChanged = useCallback(() => {
    setIsDomainsModalOpen(false);
    // Refresh domains cache if needed
  }, []);

  const handleViewTenant = useCallback(
    (tenant: Tenant) => {
      navigate({ to: `/admin/tenants/${tenant.tenantUid}` });
    },
    [navigate]
  );

  // Handle optimistic status updates from TenantStatusCell
  const handleOptimisticStatusUpdate = useCallback(
    (tenantUid: string, newStatus: string) => {
      pageState.updateOptimistic(tenantUid, "status", newStatus);
    },
    [pageState.updateOptimistic]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteTenant = useCallback(
    (tenant: Tenant) => {
      deleteTenant.mutate(tenant.tenantUid);
    },
    [deleteTenant]
  );

  // Create custom columns for status and domains
  const statusColumn = useMemo(
    () => ({
      id: "status",
      header: "Status",
      cell: ({ row }: { row: { original: Tenant } }) => {
        const tenant = row.original;
        return (
          <TenantStatusCell
            tenant={tenant}
            onStatusUpdate={handleOptimisticStatusUpdate}
          />
        );
      },
      enableSorting: false,
    }),
    [handleOptimisticStatusUpdate]
  );

  const domainsColumn = useMemo(
    () => ({
      id: "domains",
      header: "Domains",
      cell: ({ row }: { row: { original: Tenant } }) => {
        const tenant = row.original;
        return <TenantDomainsCell tenantUid={tenant.tenantUid} />;
      },
      enableSorting: false,
    }),
    []
  );

  // Create columns using the utility function
  const columns = useMemo(
    () =>
      createTenantEntityColumns<Tenant>({
        onView: handleViewTenant,
        onEdit: pageState.adminState.handleEditEntity,
        onDelete: handleDeleteTenant, // Use the enhanced handler
        statusColumn,
        domainsColumn,
        customActions: [
          {
            label: "Manage Domains",
            onClick: handleManageDomains,
          },
          {
            label: "Copy tenant ID",
            onClick: (tenant) =>
              navigator.clipboard.writeText(tenant.tenantUid || ""),
          },
        ],
      }),
    [
      handleViewTenant,
      pageState.adminState.handleEditEntity,
      handleDeleteTenant,
      statusColumn,
      domainsColumn,
      handleManageDomains,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createTenantFilterFields({
        filters: pageState.filters,
        updateFilter: pageState.updateFilter,
      }),
    [pageState.filters, pageState.updateFilter]
  );

  // Form components for modals
  const CreateFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (tenant: Tenant) => void;
      onCancel: () => void;
    }) => (
      <TenantForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />
    ),
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (tenant: Tenant) => void;
      onCancel: () => void;
    }) => (
      <TenantForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<Tenant, TenantFilters> = {
    title: "Tenants",
    description: "Manage system tenants and organizations",
    searchPlaceholder: "Search tenants by name or description...",
    emptyMessage: "No tenants found.",

    entities: pageState.applyOptimisticUpdates(tenants),
    loading: loading || deleteTenant.isPending,
    error: error || (deleteTenant.error?.message ?? null),
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

    modalMaxWidth: "md",
    deleteDescription: (tenant) => (
      <>
        Are you sure you want to permanently delete "{tenant.name}"? This action
        cannot be undone and will also delete all associated data including:
        <ul className="mt-2 ml-4 list-disc text-sm">
          <li>All domains associated with this tenant</li>
          <li>All users in this tenant</li>
          <li>All studies, cases, and slides</li>
          <li>All annotations and associated files</li>
        </ul>
        <span className="font-semibold text-red-600 mt-2 block">
          This is a destructive action that should only be used as a last
          resort.
        </span>
      </>
    ),

    onRowClick: handleViewTenant,
    enablePagination: true,
  };

  return (
    <>
      <AdminEntityPage config={pageConfig} state={pageState} />

      {/* Tenant Domains Modal - kept separate as it's tenant-specific */}
      <TenantDomainsModal
        isOpen={isDomainsModalOpen}
        tenant={selectedTenantForDomains}
        onClose={() => setIsDomainsModalOpen(false)}
        onSuccess={handleDomainsChanged}
      />
    </>
  );
}

export default TenantsAdmin;
