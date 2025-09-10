// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useMemo, useCallback } from "react";
import { useNavigate } from "@tanstack/react-router";
import DataListPage from "@/components/DataListPage";
import { Badge } from "@/components/ui/badge";
import ContentBadges from "@/components/ContentBadges";
import { apiFetch } from "@/utils/fetchUtils";
import { CaseWithSlides } from "@/hooks/useCases";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
  createIdColumn,
  createActionsColumn,
  createUserColumn,
} from "@/components/table-helpers/columnHelpers";

interface CaseFilters {
  searchQuery: string;
  searchName: string;
  hasVectorAnnotations: boolean;
  hasRasterAnnotations: boolean;
}

interface PaginationInfo {
  totalPages: number;
  total: number;
}

interface StudyCasesTableProps {
  studyUid: string;
  cases: CaseWithSlides[];
  loading: boolean;
  pagination: PaginationInfo | null;
  filters: CaseFilters;
  hasActiveFilters: boolean;
  currentPage: number;
  pageSize: number;
  onFilterChange: (key: keyof CaseFilters, value: any) => void;
  onClearFilters: () => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  getTitle: () => string;
  getSubtitle: () => string;
}

const StudyCasesTable: React.FC<StudyCasesTableProps> = ({
  studyUid,
  cases,
  loading,
  pagination,
  filters,
  hasActiveFilters,
  currentPage,
  pageSize,
  onFilterChange,
  onClearFilters,
  onPageChange,
  onPageSizeChange,
  getTitle,
  getSubtitle,
}) => {
  const navigate = useNavigate();

  const handleViewCase = useCallback(
    (caseItem: CaseWithSlides) => {
      // Navigate to the first slide of the case
      if (caseItem.slides && caseItem.slides.length > 0) {
        navigate({
          to: `/studies/${studyUid}/v/${caseItem.caseUid}/i/${caseItem.slides[0].slideUid}`,
        });
      } else {
        // If no slides available, try to fetch them first
        apiFetch(`/api/v1/cases/${caseItem.caseUid}/slides`)
          .then((slidesData: any) => {
            if (slidesData.slides && slidesData.slides.length > 0) {
              navigate({
                to: `/studies/${studyUid}/v/${caseItem.caseUid}/i/${slidesData.slides[0].slideUid}`,
              });
            } else {
              console.warn("No slides found for case:", caseItem.caseUid);
            }
          })
          .catch((error) => {
            console.error("Failed to fetch slides for case:", error);
          });
      }
    },
    [navigate, studyUid]
  );

  // Table columns configuration
  const columns = useMemo(
    () => [
      createSelectionColumn<CaseWithSlides>(),

      createTitleDescriptionColumn<CaseWithSlides>({
        accessor: "name",
        header: "Case",
        getTitle: (caseItem: CaseWithSlides) => caseItem.name || "Unnamed Case",
        getDescription: (caseItem: CaseWithSlides) => caseItem.metadata || null,
        sortable: true,
      }),

      createUserColumn<CaseWithSlides>({
        accessor: "creatorUid",
        header: "Creator",
        sortable: false,
        showIcon: true,
      }),

      {
        id: "annotations",
        header: "Annotations",
        cell: ({ row }: { row: { original: CaseWithSlides } }) => {
          const caseItem = row.original;
          // Use annotationCount for total annotations or fall back to legacy individual counts
          const hasAnnotations = (caseItem.annotationCount ?? 0) > 0;
          const vectorCount = (caseItem as any).vectorAnnotationCount ?? 0;
          const rasterCount = (caseItem as any).rasterAnnotationCount ?? 0;

          return (
            <div className="flex gap-2">
              {vectorCount > 0 && (
                <Badge
                  variant="outline"
                  className="text-blue-700 bg-blue-50 border-blue-200 dark:text-blue-300 dark:bg-blue-950 dark:border-blue-800"
                >
                  Vector ({vectorCount})
                </Badge>
              )}
              {rasterCount > 0 && (
                <Badge
                  variant="outline"
                  className="text-green-700 bg-green-50 border-green-200 dark:text-green-300 dark:bg-green-950 dark:border-green-800"
                >
                  Raster ({rasterCount})
                </Badge>
              )}
              {hasAnnotations && !vectorCount && !rasterCount && (
                <Badge
                  variant="outline"
                  className="text-green-700 bg-green-50 border-green-200 dark:text-green-300 dark:bg-green-950 dark:border-green-800"
                >
                  {caseItem.annotationCount} annotations
                </Badge>
              )}
              {!hasAnnotations && !vectorCount && !rasterCount && (
                <Badge variant="secondary">No annotations</Badge>
              )}
            </div>
          );
        },
        enableSorting: false,
      },

      {
        id: "content",
        header: "Content",
        cell: ({ row }: { row: { original: CaseWithSlides } }) => {
          const caseItem = row.original;
          const slideCount = caseItem.slideCount || 0;

          return (
            <ContentBadges
              slideCount={slideCount}
              showCases={false}
              showSlides={true}
              loading={loading}
            />
          );
        },
        enableSorting: false,
      },

      createDateColumn<CaseWithSlides>({
        accessor: "createdAt",
        header: "Created",
        sortable: true,
      }),

      createIdColumn<CaseWithSlides>({
        accessor: "caseUid",
        header: "Case ID",
        maxLength: 8,
      }),

      createActionsColumn<CaseWithSlides>({
        onView: handleViewCase,
        entityName: "Case",
        customActions: [
          {
            label: "Copy case ID",
            onClick: (caseItem: CaseWithSlides) =>
              navigator.clipboard.writeText(caseItem.caseUid),
          },
        ],
      }),
    ],
    [handleViewCase, loading]
  );

  // Filter fields configuration
  const filterFields = useMemo(
    () => [
      {
        type: "text" as const,
        key: "searchName",
        label: "Case Name",
        placeholder: "Filter by name...",
        value: filters.searchName,
        onChange: (value: string | boolean) =>
          onFilterChange("searchName", value as string),
      },
    ],
    [filters, onFilterChange]
  );

  const getRowId = useCallback(
    (caseItem: CaseWithSlides, index: number) =>
      caseItem.caseUid || `case-${index}`,
    []
  );

  return (
    <DataListPage
      title={getTitle()}
      subtitle={getSubtitle()}
      data={cases}
      loading={loading}
      error={null}
      columns={columns}
      getRowId={getRowId}
      onRowClick={handleViewCase}
      searchQuery={filters.searchQuery}
      onSearchQueryChange={(query) => onFilterChange("searchQuery", query)}
      searchPlaceholder="Search by name, metadata..."
      filterFields={filterFields}
      hasActiveFilters={hasActiveFilters}
      onClearFilters={onClearFilters}
      emptyMessage="No cases found."
      pagination={true}
      pageSize={pageSize}
      currentPage={currentPage}
      totalPages={pagination?.totalPages || 0}
      totalItems={pagination?.total || 0}
      onPageChange={onPageChange}
      onPageSizeChange={onPageSizeChange}
      maxWidth=""
      className=""
    />
  );
};

export default StudyCasesTable;
