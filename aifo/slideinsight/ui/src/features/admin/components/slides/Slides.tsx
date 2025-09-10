// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  useSlides,
  useDeleteSlide,
  type SlideWithCount,
} from "../../../../api";
import SlideForm from "./SlideForm";
import { TrashIcon } from "../../../../components/icons";
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
  createSlideFilterFields,
} from "../../utils/adminFilters";
import { copyToClipboard } from "../../../../utils/clipboardUtils";
import { FilterField } from "../../../../types/search";
import { createUserColumn } from "../../../../components/table-helpers/columnHelpers";

function SlidesAdminRefactoredComponent() {
  const navigate = useNavigate();

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Slides",
      entityNameSingular: "Slide",
      deleteEndpoint: (slide: SlideWithCount) =>
        `/api/v1/slides/${slide.slideUid}`,
      getEntityDisplayName: (slide: SlideWithCount) =>
        slide.slideName || "Unnamed Slide",
      getEntityId: (slide: SlideWithCount) => slide.slideUid || "",
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
  const slideQuery = useMemo(() => {
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: pageState.filters.searchQuery,
      withMaskCounts: true, // Enable mask counts for the admin table
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch slides data using the new typed API
  const {
    data: slides,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useSlides(slideQuery);

  // Delete mutation
  const deleteSlide = useDeleteSlide({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete slide:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  const handleViewSlide = useCallback(
    (slide: SlideWithCount) => {
      // Navigate to slide detail page
      navigate({
        to: "/admin/slides/$slideUid",
        params: { slideUid: slide.slideUid },
      });
    },
    [navigate]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteSlide = useCallback(
    (slide: SlideWithCount) => {
      deleteSlide.mutate(slide.slideUid);
    },
    [deleteSlide]
  );

  // Create slides-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<SlideWithCount>({
        entityName: "Slide",
        titleConfig: {
          accessor: "slideName",
          header: "Slide",
          getTitle: (slide) => slide.slideName || "Unnamed Slide",
          getDescription: (slide) =>
            `${slide.slideWidth} × ${slide.slideHeight} pixels`,
        },
        includeId: true,
        idConfig: {
          accessor: "slideUid",
          header: "ID",
          maxLength: 8,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        customColumns: [
          createUserColumn<SlideWithCount>({
            accessor: "creatorUid",
            header: "Creator",
            sortable: false,
            showIcon: true,
          }),
          {
            id: "resolution",
            header: "Resolution",
            cell: ({ row }) => {
              const slide = row.original;
              return (
                <div className="text-sm">
                  <div className="font-medium">
                    {slide.slideMpp?.toFixed(2)} μm/pixel
                  </div>
                  <div className="text-muted-500 text-xs">
                    {slide.slideWidth} × {slide.slideHeight}
                  </div>
                </div>
              );
            },
            enableSorting: false,
          },
          {
            id: "masks",
            header: "Masks",
            cell: ({ row }) => {
              const slide = row.original;
              const maskCount = slide.maskCount || 0;
              return maskCount > 0 ? (
                <Badge
                  variant="outline"
                  className="text-purple-700 bg-purple-50 border-purple-200"
                >
                  {maskCount} {maskCount === 1 ? "mask" : "masks"}
                </Badge>
              ) : (
                <span className="text-muted-400 text-sm">No masks</span>
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewSlide,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteSlide, // Use the enhanced handler
          customActions: [
            {
              label: "Copy slide ID",
              onClick: (slide) => copyToClipboard(slide.slideUid || ""),
            },
          ],
        },
      }),
    [handleViewSlide, pageState.adminState.handleEditEntity, handleDeleteSlide]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createSlideFilterFields({
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
      onSuccess: (slide: SlideWithCount) => void;
      onCancel: () => void;
    }) => <SlideForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />,
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (slide: SlideWithCount) => void;
      onCancel: () => void;
    }) => (
      <SlideForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<SlideWithCount, NameSearchFilters> = {
    title: "Slides",
    description: "Manage slides and slide data",
    searchPlaceholder: "Search slides by name...",
    emptyMessage: "No slides found.",

    entities: slides,
    loading: loading || deleteSlide.isPending,
    error: error || (deleteSlide.error?.message ?? null),
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
    deleteDescription: (slide) =>
      `Are you sure you want to delete the slide "${slide.slideName}"? This action cannot be undone and will also delete all associated annotations and mask data.`,

    onRowClick: handleViewSlide,
    enablePagination: true,
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default SlidesAdminRefactoredComponent;
