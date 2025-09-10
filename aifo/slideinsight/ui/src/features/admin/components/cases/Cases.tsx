// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useCases, useDeleteCase, type CaseWithSlides } from "../../../../api";
import { apiFetch } from "../../../../utils/fetchUtils";
import { formatDate } from "../../../../utils/adminTableUtils";
import { copyToClipboard } from "../../../../utils/clipboardUtils";
import { TrashIcon } from "../../../../components/icons";
import { Button } from "../../../../components/ui/button";
import { Badge } from "../../../../components/ui/badge";
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
  createCaseFilterFields,
} from "../../utils/adminFilters";
import { FilterField } from "../../../../types/search";
import { createUserColumn } from "../../../../components/table-helpers/columnHelpers";

// Placeholder form component for cases
const CaseForm = ({
  entity,
  onSuccess,
  onCancel,
}: {
  entity: CaseWithSlides | null;
  onSuccess: (entity: CaseWithSlides) => void;
  onCancel: () => void;
}) => {
  // This is a placeholder implementation
  // TODO: Implement proper case form when needed
  return (
    <div className="p-4">
      <p className="text-muted-600 mb-4">
        Case creation and editing is not yet implemented.
      </p>
      <div className="flex justify-end space-x-2">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </div>
  );
};

function CasesAdminRefactoredComponent() {
  const navigate = useNavigate();

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Cases",
      entityNameSingular: "Case",
      deleteEndpoint: (caseItem: CaseWithSlides) =>
        `/api/v1/cases/${caseItem.caseUid}`,
      getEntityDisplayName: (caseItem: CaseWithSlides) =>
        caseItem.name || "Unnamed Case",
      getEntityId: (caseItem: CaseWithSlides) => caseItem.caseUid || "",
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
  const caseQuery = useMemo(() => {
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: pageState.filters.searchQuery,
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch cases data using the new typed API
  const {
    data: cases,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useCases(caseQuery);

  // Delete mutation
  const deleteCase = useDeleteCase({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete case:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  const handleViewCase = useCallback(
    async (caseItem: CaseWithSlides) => {
      try {
        // Fetch the first image of the case
        const caseData = (await apiFetch(
          `/api/v1/cases/${caseItem.caseUid}/slides`
        )) as any;
        const slides = caseData.slides || [];

        if (slides.length > 0 && slides[0].slideUid) {
          // Navigate to study-agnostic route (admin view)
          navigate({
            to: "/v/$caseUid/i/$slideUid",
            params: {
              caseUid: caseItem.caseUid,
              slideUid: slides[0].slideUid,
            },
          });
        }
      } catch (error) {
        console.error("Failed to fetch case images:", error);
      }
    },
    [navigate]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteCase = useCallback(
    (caseItem: CaseWithSlides) => {
      deleteCase.mutate(caseItem.caseUid);
    },
    [deleteCase]
  );

  // Create cases-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<CaseWithSlides>({
        entityName: "Case",
        titleConfig: {
          accessor: "name",
          header: "Case",
          getTitle: (caseItem) => caseItem.name || "Unnamed Case",
          getDescription: (caseItem) => caseItem.metadata || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        includeId: true,
        idConfig: {
          accessor: "caseUid",
          header: "ID",
          maxLength: 8,
        },
        customColumns: [
          createUserColumn<CaseWithSlides>({
            accessor: "creatorUid",
            header: "Creator",
            sortable: false,
            showIcon: true,
          }),
          {
            id: "slides",
            header: "Slides",
            cell: ({ row }) => {
              const caseItem = row.original;
              const slideCount =
                caseItem.slides?.length || caseItem.slideCount || 0;
              return (
                <Badge
                  variant="outline"
                  className="text-purple-700 bg-purple-50 border-purple-200"
                >
                  {slideCount} {slideCount === 1 ? "slide" : "slides"}
                </Badge>
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewCase,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteCase, // Use the enhanced handler
          customActions: [
            {
              label: "Copy case ID",
              onClick: (caseItem) => copyToClipboard(caseItem.caseUid || ""),
            },
            {
              label: "View details",
              onClick: (caseItem) =>
                navigate({ to: `/admin/cases/${caseItem.caseUid}` }),
            },
          ],
        },
      }),
    [
      handleViewCase,
      navigate,
      pageState.adminState.handleEditEntity,
      handleDeleteCase,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createCaseFilterFields({
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
      onSuccess: (caseItem: CaseWithSlides) => void;
      onCancel: () => void;
    }) => <CaseForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />,
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (caseItem: CaseWithSlides) => void;
      onCancel: () => void;
    }) => (
      <CaseForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<CaseWithSlides, NameSearchFilters> = {
    title: "Cases",
    description: "Manage cases and case data",
    searchPlaceholder: "Search cases by name or metadata...",
    emptyMessage: "No cases found.",

    entities: cases,
    loading: loading || deleteCase.isPending,
    error: error || (deleteCase.error?.message ?? null),
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
    deleteDescription: (caseItem) =>
      `Are you sure you want to delete the case "${caseItem.name}"? This action cannot be undone and will also delete all associated slides and annotations.`,

    onRowClick: handleViewCase,
    enablePagination: true,
    disableCreate: true, // Disable create since form is not implemented
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default CasesAdminRefactoredComponent;
