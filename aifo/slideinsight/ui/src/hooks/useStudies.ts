// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import { useApiQuery, queryKeys } from "../utils/apiQueries";
import type { PaginationInfo } from "./usePaginatedApi";

export interface Study {
  tenantUid: string;
  studyUid: string;
  creatorUid: string;
  name: string;
  description: string;
  metadata: string;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface StudyWithCasesAndSlides extends Study {
  caseCount: number;
  slideCount: number;
}

interface StudyMetadata {
  [key: string]: any;
}

interface UseStudiesOptions {
  page?: number;
  limit?: number;
  q?: string;
  name?: string;
  status?: string;
  sort?: string;
  dir?: string;
  filterAccessibleStudies?: boolean;
}

interface StudiesApiResponse {
  studies: StudyWithCasesAndSlides[];
  pagination: PaginationInfo;
}

export function useStudies(options: UseStudiesOptions = {}) {
  const {
    page = 1,
    limit = 100,
    q = "",
    name = "",
    status = "",
    sort = "",
    dir = "",
    filterAccessibleStudies = false,
  } = options;

  // Create a stable query key based on all options
  const queryKey = useMemo(
    () =>
      queryKeys.studies.list({
        page,
        limit,
        q,
        name,
        status,
        sort,
        dir,
        filterAccessibleStudies,
      }),
    [page, limit, q, name, status, sort, dir, filterAccessibleStudies]
  );

  // Build the URL with query parameters
  const url = useMemo(() => {
    const params = new URLSearchParams();
    params.append("page", page.toString());
    params.append("limit", limit.toString());

    if (q) params.append("q", q);
    if (name) params.append("name", name);
    if (status) params.append("status", status);
    if (sort) params.append("sort", sort);
    if (dir) params.append("dir", dir);
    if (filterAccessibleStudies)
      params.append("filterAccessibleStudies", "true");

    return `/api/v1/studies?${params.toString()}`;
  }, [page, limit, q, name, status, sort, dir, filterAccessibleStudies]);

  const queryResult = useApiQuery<StudiesApiResponse>(queryKey, url, {
    staleTime: 1000 * 60 * 2, // 2 minutes - shorter for frequently changing data
    placeholderData: {
      studies: [],
      pagination: {
        page: 1,
        limit: 100,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      },
    },
  });

  return {
    studies: queryResult.data?.studies || [],
    pagination: queryResult.data?.pagination || {
      page: 1,
      limit: 100,
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
  };
}

// Hook to fetch a single study by UID
export function useStudyByUID(studyUid: string | null) {
  const queryKey = useMemo(
    () => (studyUid ? queryKeys.studies.detail(studyUid) : null),
    [studyUid]
  );

  const url = useMemo(
    () => (studyUid ? `/api/v1/studies/${studyUid}` : null),
    [studyUid]
  );

  const queryResult = useApiQuery<StudyWithCasesAndSlides>(
    queryKey || ["studies", "detail", "disabled"],
    url || "",
    {
      enabled: !!studyUid,
      staleTime: 1000 * 60 * 5, // 5 minutes
    }
  );

  return {
    data: queryResult.data || null,
    isLoading: queryResult.isLoading,
    error: queryResult.error,
    refetch: queryResult.refetch,
  };
}
