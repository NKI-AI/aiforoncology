// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo } from "react";
import {
  useAlgorithms,
  useDeleteAlgorithm,
  type Algorithm,
  type AlgorithmQuery,
} from "../../../../api";
import { useNavigate } from "@tanstack/react-router";
import { Badge } from "../../../../components/ui/badge";
import { TrashIcon } from "../../../../components/icons";
import { AlgorithmForm } from "../forms";
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
import { FilterField } from "../../../../types/search";

interface AlgorithmFilters extends NameSearchFilters {
  searchVersion: string;
  executionMode: string;
}

function AlgorithmsAdmin() {
  const navigate = useNavigate();

  // Initialize filters
  const initialFilters: AlgorithmFilters = {
    ...initialNameSearchFilters,
    searchVersion: "",
    executionMode: "",
  };

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Algorithms",
      entityNameSingular: "Algorithm",
      deleteEndpoint: (algorithm: Algorithm) =>
        `/api/v1/algorithms/${algorithm.id}`,
      getEntityDisplayName: (algorithm: Algorithm) =>
        algorithm.name || "Unnamed Algorithm",
      getEntityId: (algorithm: Algorithm) => algorithm.id || "",
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
  const algorithmQuery: AlgorithmQuery = useMemo(() => {
    const { searchQuery, searchVersion, executionMode, ...otherFilters } =
      pageState.filters;
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: searchQuery,
      version: searchVersion || undefined,
      executionMode: executionMode as "BATCH" | "STREAM" | undefined,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch algorithms data using the new typed API
  const {
    data: algorithms,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useAlgorithms(algorithmQuery);

  // Delete mutation
  const deleteAlgorithm = useDeleteAlgorithm({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete algorithm:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  const handleViewAlgorithm = useCallback(
    (algorithm: Algorithm) => {
      navigate({ to: `/admin/algorithms/${algorithm.id}` });
    },
    [navigate]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteAlgorithm = useCallback(
    (algorithm: Algorithm) => {
      deleteAlgorithm.mutate(algorithm.id);
    },
    [deleteAlgorithm]
  );

  // Create algorithm-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<Algorithm>({
        entityName: "Algorithm",
        titleConfig: {
          accessor: "name",
          header: "Algorithm",
          getTitle: (algorithm) => algorithm.name || "Unnamed Algorithm",
          getDescription: (algorithm) => algorithm.description || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        includeId: true,
        idConfig: {
          accessor: "id",
          header: "ID",
          maxLength: 8,
        },
        customColumns: [
          {
            id: "version",
            header: "Version",
            cell: ({ row }) => {
              const algorithm = row.original;
              return (
                <Badge variant="outline" className="font-mono">
                  {algorithm.version}
                </Badge>
              );
            },
            enableSorting: true,
          },
          {
            id: "tenant",
            header: "Tenant",
            cell: ({ row }) => {
              const algorithm = row.original;
              return (
                <div className="text-sm">
                  <div className="text-muted-900">
                    {algorithm.tenantName || "Unknown Tenant"}
                  </div>
                  <div className="text-muted-500 text-xs">
                    ID: {algorithm.tenantId}
                  </div>
                </div>
              );
            },
            enableSorting: false,
          },
          {
            id: "executionMode",
            header: "Execution Mode",
            cell: ({ row }) => {
              const algorithm = row.original;
              return (
                <Badge
                  variant={
                    algorithm.executionMode === "STREAM"
                      ? "default"
                      : "secondary"
                  }
                  className={
                    algorithm.executionMode === "STREAM"
                      ? "bg-blue-100 text-blue-800"
                      : "bg-gray-100 text-muted-800"
                  }
                >
                  {algorithm.executionMode}
                </Badge>
              );
            },
            enableSorting: true,
          },
          {
            id: "progressTransport",
            header: "Progress Transport",
            cell: ({ row }) => {
              const algorithm = row.original;
              return (
                <Badge variant="outline">{algorithm.progressTransport}</Badge>
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewAlgorithm,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteAlgorithm, // Use the enhanced handler
          customActions: [
            {
              label: "View Runs",
              onClick: (algorithm) =>
                navigate({ to: `/admin/algorithms/${algorithm.id}/runs` }),
            },
            {
              label: "Copy algorithm ID",
              onClick: (algorithm) =>
                navigator.clipboard.writeText(algorithm.id || ""),
            },
          ],
        },
      }),
    [
      handleViewAlgorithm,
      pageState.adminState.handleEditEntity,
      handleDeleteAlgorithm,
      navigate,
    ]
  );

  // Filter fields configuration
  const filterFields: FilterField[] = useMemo(
    () => [
      {
        type: "text",
        key: "searchName",
        label: "Name",
        placeholder: "Filter by algorithm name...",
        value: pageState.filters.searchName,
        onChange: (value) => pageState.updateFilter("searchName", value),
      },
      {
        type: "text",
        key: "searchVersion",
        label: "Version",
        placeholder: "Filter by version...",
        value: pageState.filters.searchVersion,
        onChange: (value) => pageState.updateFilter("searchVersion", value),
      },
      {
        type: "select",
        key: "executionMode",
        label: "Execution Mode",
        placeholder: "All execution modes",
        value: pageState.filters.executionMode,
        onChange: (value) => pageState.updateFilter("executionMode", value),
        options: [
          { label: "All", value: "" },
          { label: "Batch", value: "BATCH" },
          { label: "Stream", value: "STREAM" },
        ],
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
      onSuccess: (algorithm: Algorithm) => void;
      onCancel: () => void;
    }) => (
      <AlgorithmForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />
    ),
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (algorithm: Algorithm) => void;
      onCancel: () => void;
    }) => (
      <AlgorithmForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<Algorithm, AlgorithmFilters> = {
    title: "Algorithms",
    description: "Manage machine learning algorithms and their configurations",
    searchPlaceholder: "Search algorithms by name or description...",
    emptyMessage: "No algorithms found.",

    entities: algorithms,
    loading: loading || deleteAlgorithm.isPending,
    error: error || (deleteAlgorithm.error?.message ?? null),
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
    deleteDescription: (algorithm) =>
      `Are you sure you want to delete the algorithm "${algorithm.name}"? This action cannot be undone and will also delete all associated runs and configurations.`,

    onRowClick: handleViewAlgorithm,
    enablePagination: true,
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default AlgorithmsAdmin;
