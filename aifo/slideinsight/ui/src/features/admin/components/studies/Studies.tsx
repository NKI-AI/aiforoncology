// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  useStudies,
  useDeleteStudy,
  type StudyWithCasesAndSlides,
  type StudyQuery,
} from "../../../../api";
import { StudyForm } from "../forms";
import { TrashIcon, SecurityIcon } from "../../../../components/icons";
import { StatCard } from "../../../../components/ui/stat-card";
import { StudyStatusCell } from "../../../../components/ui/StudyStatusCell";
import { StudyPermissionBadge } from "../../../../components/StudyPermissionBadge";
import ContentBadges from "../../../../components/ContentBadges";
import { FileText, CheckCircle, Clock, FolderOpen, Image } from "lucide-react";
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
  type CommonEntityFilters,
  initialCommonEntityFilters,
  createStudyFilterFields,
} from "../../utils/adminFilters";
import { copyToClipboard } from "../../../../utils/clipboardUtils";
import { FilterField } from "../../../../types/search";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
  createIdColumn,
  createActionsColumn,
  createUserColumn,
} from "../../../../components/table-helpers/columnHelpers";

function StudiesAdminRefactoredComponent() {
  const navigate = useNavigate();

  // Optimistic status updates for immediate UI feedback
  const [optimisticStatuses, setOptimisticStatuses] = useState<
    Record<string, boolean>
  >({});

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Studies",
      entityNameSingular: "Study",
      deleteEndpoint: (study: StudyWithCasesAndSlides) =>
        `/api/v1/studies/${study.studyUid}`,
      getEntityDisplayName: (study: StudyWithCasesAndSlides) =>
        study.name || "Unnamed Study",
      getEntityId: (study: StudyWithCasesAndSlides) => study.studyUid || "",
    }),
    []
  );

  // Set up the admin entity page state
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: initialCommonEntityFilters,
      enableOptimisticUpdates: true,
      optimisticUpdateFields: ["isPublished"],
    },
    () => {} // Refetch will be handled by the new API hooks
  );

  // Prepare query for the new API hooks
  const studyQuery: StudyQuery = useMemo(() => {
    const { searchQuery, filterAccessibleStudies, ...otherFilters } =
      pageState.filters;
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: searchQuery,
      filterAccessibleStudies,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch studies data using the new typed API
  const {
    data: studies,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useStudies(studyQuery);

  // Delete mutation
  const deleteStudy = useDeleteStudy({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete study:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Apply optimistic status updates to studies data
  const optimisticStudies = useMemo(() => {
    return studies.map((study) => ({
      ...study,
      isPublished:
        optimisticStatuses[study.studyUid] !== undefined
          ? optimisticStatuses[study.studyUid]
          : study.isPublished,
    }));
  }, [studies, optimisticStatuses]);

  // Calculate statistics for the stats cards
  const statistics = useMemo(() => {
    const publishedStudies = optimisticStudies.filter(
      (study) => study?.isPublished === true
    ).length;
    const draftStudies = optimisticStudies.filter(
      (study) => study?.isPublished === false
    ).length;
    const totalCases = optimisticStudies.reduce(
      (sum, study) => sum + (study.caseCount || 0),
      0
    );

    return {
      totalStudies: optimisticStudies.length,
      publishedStudies,
      draftStudies,
      totalCases,
    };
  }, [optimisticStudies]);

  const handleViewStudy = useCallback(
    (study: StudyWithCasesAndSlides) => {
      navigate({ to: `/admin/studies/${study.studyUid}` });
    },
    [navigate]
  );

  const handleCustomAction = useCallback(
    (action: string, study: StudyWithCasesAndSlides) => {
      if (action === "permissions") {
        navigate({ to: `/admin/studies/${study.studyUid}/permissions` });
      }
    },
    [navigate]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteStudy = useCallback(
    (study: StudyWithCasesAndSlides) => {
      deleteStudy.mutate(study.studyUid);
    },
    [deleteStudy]
  );

  // Handle optimistic status updates from StudyStatusCell
  const handleOptimisticStatusUpdate = useCallback(
    (studyUid: string, newStatus: boolean) => {
      setOptimisticStatuses((prev) => ({
        ...prev,
        [studyUid]: newStatus,
      }));
      // Also update the page state optimistic updates
      pageState.updateOptimistic(studyUid, "isPublished", newStatus);
    },
    [pageState]
  );

  // Create studies-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<StudyWithCasesAndSlides>({
        entityName: "Study",
        titleConfig: {
          accessor: "name",
          header: "Study",
          getTitle: (study) => study.name || "Unnamed Study",
          getDescription: (study) => study.description || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        customColumns: [
          createUserColumn<StudyWithCasesAndSlides>({
            accessor: "creatorUid",
            header: "Creator",
            sortable: false,
            showIcon: true,
          }),
          {
            id: "status",
            header: "Status",
            cell: ({ row }) => {
              const study = row.original;
              return (
                <StudyStatusCell
                  study={study}
                  onStatusUpdate={handleOptimisticStatusUpdate}
                />
              );
            },
            enableSorting: false,
          },
          {
            id: "content",
            header: "Content",
            cell: ({ row }) => {
              const study = row.original;
              const caseCount = study.caseCount || 0;
              const slideCount = study.slideCount || 0;
              return (
                <ContentBadges
                  caseCount={caseCount}
                  slideCount={slideCount}
                  showCases={true}
                  showSlides={true}
                  loading={loading || deleteStudy.isPending}
                />
              );
            },
            enableSorting: false,
          },
          {
            id: "access",
            header: "Access",
            cell: ({ row }) => {
              const study = row.original;
              return (
                <StudyPermissionBadge
                  studyUid={study.studyUid}
                  permission="studies.view"
                  size="sm"
                />
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewStudy,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteStudy, // Use the enhanced handler
          customActions: [
            {
              label: "Permissions",
              onClick: (study) => handleCustomAction("permissions", study),
              variant: "default",
              icon: SecurityIcon,
            },
            {
              label: "Copy study ID",
              onClick: (study) => copyToClipboard(study.studyUid || ""),
            },
          ],
        },
      }),
    [
      handleViewStudy,
      pageState.adminState.handleEditEntity,
      handleCustomAction,
      handleDeleteStudy,
      handleOptimisticStatusUpdate,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createStudyFilterFields({
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
      onSuccess: (study: any) => void;
      onCancel: () => void;
    }) => <StudyForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />,
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (study: any) => void;
      onCancel: () => void;
    }) => (
      <StudyForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Statistics cards component
  const statisticsComponent = useMemo(
    () => (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {loading ? (
          <>
            {[...Array(4)].map((_, i) => (
              <StatCard
                key={i}
                title="Loading..."
                value="--"
                className="animate-pulse"
              />
            ))}
          </>
        ) : (
          <>
            <StatCard
              title="Total Studies"
              value={statistics.totalStudies}
              subtitle={`${statistics.publishedStudies} published, ${statistics.draftStudies} draft`}
              icon={FileText}
            />

            <StatCard
              title="Published Studies"
              value={statistics.publishedStudies}
              subtitle="Active and visible studies"
              icon={CheckCircle}
            />

            <StatCard
              title="Draft Studies"
              value={statistics.draftStudies}
              subtitle="Unpublished studies"
              icon={Clock}
            />

            <StatCard
              title="Total Cases"
              value={statistics.totalCases}
              subtitle="Cases across all studies"
              icon={FolderOpen}
            />
          </>
        )}
      </div>
    ),
    [loading, statistics]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<
    StudyWithCasesAndSlides,
    CommonEntityFilters
  > = {
    title: "Studies",
    description: "Manage research studies and their data",
    searchPlaceholder: "Search studies by name or description...",
    emptyMessage: "No studies found.",

    entities: optimisticStudies,
    loading: loading || deleteStudy.isPending,
    error: error || (deleteStudy.error?.message ?? null),
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
    deleteDescription: (study) =>
      `Are you sure you want to delete the study "${study.name}"? This action cannot be undone and will also delete all associated cases, slides, and annotations.`,

    onRowClick: handleViewStudy,
    enablePagination: true,
    statistics: statisticsComponent,
  };

  return <AdminEntityPage config={pageConfig} state={pageState} />;
}

export default StudiesAdminRefactoredComponent;
