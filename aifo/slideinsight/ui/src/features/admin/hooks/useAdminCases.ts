// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import { useApiQuery, queryKeys } from "../../../utils/apiQueries";
import type { CaseWithSlides, UseCasesOptions } from "../../../hooks/useCases";
import type { PaginationInfo } from "../../../hooks/usePaginatedApi";

interface AdminCasesApiResponse {
  cases: CaseWithSlides[];
  pagination: PaginationInfo;
}

interface UseAdminCasesOptions extends UseCasesOptions {
  // Admin-specific options can be added here
}

/**
 * Admin-optimized hook for loading cases with enhanced performance
 * Uses TanStack Query for caching and background updates
 */
export function useAdminCases(options: UseAdminCasesOptions = {}) {
  const {
    page = 1,
    limit = 20, // Smaller default for admin screens
    q = "",
    name = "",
    status = "",
    has_vector_annotations = "",
    has_raster_annotations = "",
    sort = "createdAt",
    dir = "desc",
    withAnnotations = true, // Default to true for admin screens to get annotation counts
  } = options;

  // Create a stable query key for admin cases
  const queryKey = useMemo(
    () => [
      ...queryKeys.cases.list(),
      "admin",
      {
        page,
        limit,
        q,
        name,
        status,
        has_vector_annotations,
        has_raster_annotations,
        sort,
        dir,
        withAnnotations,
      },
    ],
    [
      page,
      limit,
      q,
      name,
      status,
      has_vector_annotations,
      has_raster_annotations,
      sort,
      dir,
      withAnnotations,
    ]
  );

  // Build the URL with query parameters
  const url = useMemo(() => {
    const params = new URLSearchParams();
    params.append("page", page.toString());
    params.append("limit", limit.toString());

    if (q) params.append("q", q);
    if (name) params.append("name", name);
    if (status) params.append("status", status);
    if (has_vector_annotations)
      params.append("has_vector_annotations", has_vector_annotations);
    if (has_raster_annotations)
      params.append("has_raster_annotations", has_raster_annotations);
    if (sort) params.append("sort", sort);
    if (dir) params.append("dir", dir);

    return `/api/v1/cases?${params.toString()}`;
  }, [
    page,
    limit,
    q,
    name,
    status,
    has_vector_annotations,
    has_raster_annotations,
    sort,
    dir,
  ]);

  const queryResult = useApiQuery<AdminCasesApiResponse>(queryKey, url, {
    staleTime: 1000 * 60 * 1, // 1 minute - admin needs fresher data
    placeholderData: {
      cases: [],
      pagination: {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      },
    },
    select: (data) => {
      // Transform data to include annotation counts if needed
      const processedCases: CaseWithSlides[] = data.cases.map((caseItem) => ({
        ...caseItem,
        slideCount: (caseItem as any).slideCount || 0,
        slides: (caseItem as any).slides || [],
        annotationCount: withAnnotations
          ? (caseItem as any).annotationCount || 0
          : undefined,
        slidesWithAnnotations: withAnnotations
          ? (caseItem as any).slidesWithAnnotations || 0
          : undefined,
      }));

      return {
        cases: processedCases,
        pagination: data.pagination,
      };
    },
  });

  return {
    cases: queryResult.data?.cases || [],
    pagination: queryResult.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: queryResult.isLoading,
    error: queryResult.error?.message || null,
    refetch: queryResult.refetch,
    isStale: queryResult.isStale,
    isFetching: queryResult.isFetching,
    // Additional admin-specific return values
    isRefetching: queryResult.isRefetching,
    dataUpdatedAt: queryResult.dataUpdatedAt,
  };
}
