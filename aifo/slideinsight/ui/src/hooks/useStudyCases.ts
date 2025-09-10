// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useCallback, useMemo } from "react";
import { useCasesByStudy } from "@/hooks/useCases";

interface CaseFilters {
  searchQuery: string;
  searchName: string;
  hasVectorAnnotations: boolean;
  hasRasterAnnotations: boolean;
}

interface UseStudyCasesOptions {
  studyUid: string | undefined;
  initialPageSize?: number;
}

export function useStudyCases({
  studyUid,
  initialPageSize = 20,
}: UseStudyCasesOptions) {
  // Pagination state
  const [currentPage, setCurrentPage] = useState(0);
  const [pageSize, setPageSize] = useState(initialPageSize);

  // Filter state
  const [filters, setFilters] = useState<CaseFilters>({
    searchQuery: "",
    searchName: "",
    hasVectorAnnotations: false,
    hasRasterAnnotations: false,
  });

  // Fetch cases data using the existing hook
  const { cases, pagination, loading, error, refetch } = useCasesByStudy(
    studyUid,
    {
      page: currentPage + 1, // API expects 1-based indexing
      limit: pageSize,
      q: filters.searchQuery,
      name: filters.searchName,
      has_vector_annotations: filters.hasVectorAnnotations ? "true" : "",
      has_raster_annotations: filters.hasRasterAnnotations ? "true" : "",
      sort: "createdAt",
      dir: "desc",
      withAnnotations: true, // Include annotation data
    }
  );

  const updateFilter = useCallback((key: keyof CaseFilters, value: any) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setCurrentPage(0); // Reset to first page when filters change
  }, []);

  const clearFilters = useCallback(() => {
    setFilters({
      searchQuery: "",
      searchName: "",
      hasVectorAnnotations: false,
      hasRasterAnnotations: false,
    });
    setCurrentPage(0);
  }, []);

  const hasActiveFilters = useMemo(() => {
    return !!(
      filters.searchQuery ||
      filters.searchName ||
      filters.hasVectorAnnotations ||
      filters.hasRasterAnnotations
    );
  }, [filters]);

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  const handlePageSizeChange = useCallback((newPageSize: number) => {
    setPageSize(newPageSize);
    setCurrentPage(0); // Reset to first page when page size changes
  }, []);

  return {
    // Data
    cases,
    pagination,
    loading,
    error,
    refetch,

    // Filters
    filters,
    updateFilter,
    clearFilters,
    hasActiveFilters,

    // Pagination
    currentPage,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
  };
}
