// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback, useEffect } from "react";
import { useStudies, type Study } from "./useStudies";

interface StudiesFilters {
  searchQuery: string;
  searchName: string;
  searchStatus: string;
  filterAccessibleStudies?: boolean;
}

interface UseStudiesTableStateOptions {
  initialPageSize?: number;
  sort?: string;
  dir?: "asc" | "desc";
  documentTitle?: string;
}

interface UseStudiesTableStateReturn {
  // Studies data
  studies: Study[];
  pagination: any;
  loading: boolean;
  error: string | null;
  refetch: () => void;

  // Pagination state
  currentPage: number;
  pageSize: number;
  handlePageChange: (page: number) => void;
  handlePageSizeChange: (newPageSize: number) => void;

  // Filter state
  serverFilters: StudiesFilters;
  handleFiltersChange: (newFilters: StudiesFilters) => void;
}

export function useStudiesTableState({
  initialPageSize = 20,
  sort = "createdAt",
  dir = "desc",
  documentTitle,
}: UseStudiesTableStateOptions = {}): UseStudiesTableStateReturn {
  // Pagination state
  const [currentPage, setCurrentPage] = useState(0);
  const [pageSize, setPageSize] = useState(initialPageSize);

  // Server-side filter state
  const [serverFilters, setServerFilters] = useState<StudiesFilters>({
    searchQuery: "",
    searchName: "",
    searchStatus: "",
    filterAccessibleStudies: false,
  });

  // Fetch studies data
  const { studies, pagination, loading, error, refetch } = useStudies({
    page: currentPage + 1, // API expects 1-based indexing
    limit: pageSize,
    q: serverFilters.searchQuery,
    name: serverFilters.searchName,
    status: serverFilters.searchStatus,
    sort,
    dir,
    filterAccessibleStudies: serverFilters.filterAccessibleStudies,
  });

  // Handle filter changes from the table component (debounced)
  const handleFiltersChange = useCallback((newFilters: StudiesFilters) => {
    setServerFilters(newFilters);
    setCurrentPage(0); // Reset to first page when filters change
  }, []);

  // Handle pagination changes
  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  const handlePageSizeChange = useCallback((newPageSize: number) => {
    setPageSize(newPageSize);
    setCurrentPage(0); // Reset to first page when page size changes
  }, []);

  // Set document title
  useEffect(() => {
    if (documentTitle) {
      const originalTitle = document.title;
      document.title = documentTitle;
      return () => {
        document.title = originalTitle;
      };
    }
  }, [documentTitle]);

  return {
    // Studies data
    studies: Array.isArray(studies) ? studies : [],
    pagination,
    loading,
    error,
    refetch,

    // Pagination
    currentPage,
    pageSize,
    handlePageChange,
    handlePageSizeChange,

    // Filters
    serverFilters,
    handleFiltersChange,
  };
}
