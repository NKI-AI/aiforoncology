// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useCallback, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useStudiesTableState } from "@/hooks/useStudiesTableState";
import { Study } from "@/hooks/useStudies";
import { useServerSideFilters } from "@/features/admin/components/useServerSideFilters";
import DataListPage from "@/components/DataListPage";
import ContentBadges from "@/components/ContentBadges";
import StudyCell from "@/components/StudyCell";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/AlertDialog";
import { Card, CardContent } from "@/components/ui/card";
import {
  createSelectionColumn,
  createDateColumn,
  createIdColumn,
  createActionsColumn,
  createUserColumn,
} from "@/components/table-helpers/columnHelpers";
import { Badge } from "@/components/ui/badge";
import { StudyPermissionBadge } from "@/components/StudyPermissionBadge";
import { Button } from "@/components/ui/button";
import { ArrowUpDown } from "lucide-react";

// Define filters interface - use the same as useStudiesTableState expects
interface StudyFilters {
  searchQuery: string;
  searchName: string;
  searchStatus: string;
  filterAccessibleStudies: boolean;
}

const StudiesListNew: React.FC = () => {
  const navigate = useNavigate();

  // Error dialog state
  const [isErrorDialogOpen, setIsErrorDialogOpen] = useState(false);

  // Initialize filters
  const initialFilters: StudyFilters = {
    searchQuery: "",
    searchName: "",
    searchStatus: "",
    filterAccessibleStudies: false,
  };

  // Use the consolidated hook for studies table state
  const {
    studies,
    pagination,
    loading,
    error,
    currentPage,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
    handleFiltersChange,
    refetch,
  } = useStudiesTableState({
    initialPageSize: 20,
    documentTitle: "SlideInsight - Studies",
  });

  // Set up filters
  const { filters, updateFilter, clearFilters, hasActiveFilters } =
    useServerSideFilters(handleFiltersChange, initialFilters);

  // Show error dialog when error exists
  React.useEffect(() => {
    if (error) {
      setIsErrorDialogOpen(true);
    }
  }, [error]);

  const handleErrorRetry = useCallback(() => {
    if (refetch) {
      refetch();
    }
    setIsErrorDialogOpen(false);
  }, [refetch]);

  const handleErrorDismiss = useCallback(() => {
    setIsErrorDialogOpen(false);
  }, []);

  const handleViewStudy = useCallback(
    (study: Study) => {
      navigate({ to: `/studies/${study.studyUid}` });
    },
    [navigate]
  );

  // Filter the studies based on current filters
  const filteredStudies = useMemo(() => {
    return studies.filter((study) => {
      const matchesSearch =
        !filters.searchQuery ||
        study.name?.toLowerCase().includes(filters.searchQuery.toLowerCase()) ||
        study.description
          ?.toLowerCase()
          .includes(filters.searchQuery.toLowerCase());

      const matchesName =
        !filters.searchName ||
        study.name?.toLowerCase().includes(filters.searchName.toLowerCase());

      const matchesStatus =
        !filters.searchStatus ||
        (filters.searchStatus === "published" && study.isPublished) ||
        (filters.searchStatus === "unpublished" && !study.isPublished);

      return matchesSearch && matchesName && matchesStatus;
    });
  }, [studies, filters]);

  // Table columns configuration using the new helpers
  const columns = useMemo(
    () => [
      createSelectionColumn<Study>(),

      // Custom Study column with improved cell rendering
      {
        id: "name",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 font-medium hover:bg-transparent"
          >
            Study
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => {
          const study = row.original;
          return (
            <StudyCell
              title={study.name || "Unnamed Study"}
              description={study.description || null}
            />
          );
        },
        enableSorting: true,
        sortingFn: (rowA, rowB) => {
          const nameA = rowA.original.name || "";
          const nameB = rowB.original.name || "";
          return nameA.localeCompare(nameB);
        },
      },

      createUserColumn<Study>({
        accessor: "creatorUid",
        header: "Creator",
        sortable: false,
        showIcon: true,
      }),

      {
        accessorKey: "isPublished",
        header: "Status",
        cell: ({ getValue }) => {
          const isPublished = getValue() as boolean;
          return (
            <Badge variant={isPublished ? "default" : "secondary"}>
              {isPublished ? "Published" : "Unpublished"}
            </Badge>
          );
        },
        enableSorting: false,
      },

      {
        id: "content",
        header: "Content",
        cell: ({ row }) => {
          const study = row.original;
          const caseCount = (study as any).caseCount || 0;
          const slideCount = (study as any).slideCount || 0;
          return (
            <ContentBadges
              caseCount={caseCount}
              slideCount={slideCount}
              showCases={true}
              showSlides={true}
              loading={loading}
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

      createDateColumn<Study>({
        accessor: "createdAt",
        header: "Created",
        sortable: true,
      }),

      createIdColumn<Study>({
        accessor: "studyUid",
        header: "ID",
        maxLength: 8,
      }),

      createActionsColumn<Study>({
        onView: handleViewStudy,
        entityName: "Study",
        customActions: [
          {
            label: "Copy study ID",
            onClick: (study) =>
              navigator.clipboard.writeText(study.studyUid || ""),
          },
        ],
      }),
    ],
    [handleViewStudy]
  );

  // Filter fields configuration
  const filterFields = useMemo(
    () => [
      {
        type: "text" as const,
        key: "searchName",
        label: "Study Name",
        placeholder: "Filter by name...",
        value: filters.searchName,
        onChange: (value: string | boolean) =>
          updateFilter("searchName", value as string),
      },
      {
        type: "select" as const,
        key: "searchStatus",
        label: "Status",
        placeholder: "All statuses",
        value: filters.searchStatus,
        onChange: (value: string | boolean) =>
          updateFilter("searchStatus", value as string),
        options: [
          { label: "Published", value: "published" },
          { label: "Unpublished", value: "unpublished" },
        ],
      },
      {
        type: "checkbox" as const,
        key: "filterAccessibleStudies",
        label: "Filter cases",
        description: "Only show studies that you have access to",
        value: filters.filterAccessibleStudies,
        onChange: (value: string | boolean) =>
          updateFilter("filterAccessibleStudies", value as boolean),
      },
    ],
    [filters, updateFilter]
  );

  const getRowId = useCallback(
    (study: Study, index: number) => study.studyUid || `study-${index}`,
    []
  );

  return (
    <div className="container mx-auto max-w-7xl px-3 sm:px-4 py-4 sm:py-6">
      <Card className="bg-card text-card-foreground border">
        <CardContent className="p-4 sm:p-6">
          <DataListPage<Study>
            title="Research Studies"
            subtitle="Explore available research studies and their data"
            data={filteredStudies}
            loading={loading}
            error={null} // Handled via dialog
            columns={columns}
            getRowId={getRowId}
            onRowClick={handleViewStudy}
            searchQuery={filters.searchQuery}
            onSearchQueryChange={(query) => updateFilter("searchQuery", query)}
            searchPlaceholder="Search studies by name or description..."
            filterFields={filterFields}
            hasActiveFilters={hasActiveFilters}
            onClearFilters={clearFilters}
            emptyMessage="No studies found."
            pagination={true}
            pageSize={pageSize}
            currentPage={currentPage}
            totalPages={pagination?.totalPages || 0}
            totalItems={pagination?.total || 0}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
            useContainer={false}
          />
        </CardContent>
      </Card>

      {/* Error Dialog */}
      <AlertDialog open={isErrorDialogOpen} onOpenChange={setIsErrorDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Failed to Load Studies</AlertDialogTitle>
            <AlertDialogDescription>
              {error && typeof error === "object" && "message" in error
                ? (error as Error).message
                : "An unexpected error occurred while loading studies. Please try again."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleErrorDismiss}>
              Dismiss
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleErrorRetry}>
              Retry
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default StudiesListNew;
