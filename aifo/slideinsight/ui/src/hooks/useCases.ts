// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useEffect, useMemo } from "react";
import { apiFetch } from "@/utils/fetchUtils";
import { useApiQuery, queryKeys } from "@/utils/apiQueries";
import type { PaginationInfo } from "@/hooks/usePaginatedApi";

interface Case {
  tenantUid: string;
  caseUid: string;
  slideCount: number;
  creatorUid: string;
  name: string;
  metadata: string;
  createdAt: string;
  updatedAt: string;
}

interface CaseSlide {
  caseUid: string;
  slideUid: string;
  slideName: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
}

export interface CasesApiResponse {
  cases: Case[];
  pagination: PaginationInfo;
}

interface CaseSlidesApiResponse {
  slides: CaseSlide[];
}

export interface CaseWithSlides extends Case {
  slideCount: number;
  slides: CaseSlide[];
  annotationCount?: number; // Total annotations across all slides
  slidesWithAnnotations?: number; // Number of slides that have annotations
}

export interface UseCasesOptions {
  page?: number;
  limit?: number;
  q?: string;
  name?: string;
  status?: string;
  has_vector_annotations?: string;
  has_raster_annotations?: string;
  sort?: string;
  dir?: string;
  withAnnotations?: boolean;
}

export function useCases(options: UseCasesOptions = {}) {
  const {
    page = 1,
    limit = 100,
    q = "",
    name = "",
    status = "",
    has_vector_annotations = "",
    has_raster_annotations = "",
    sort = "",
    dir = "",
    withAnnotations = false,
  } = options;

  // Create a stable query key based on all options
  const queryKey = useMemo(
    () =>
      queryKeys.cases.list({
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
      }),
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

  const queryResult = useApiQuery<CasesApiResponse>(queryKey, url, {
    staleTime: 1000 * 60 * 2, // 2 minutes
    placeholderData: {
      cases: [],
      pagination: {
        page: 1,
        limit: 100,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      },
    },
    select: (data) => {
      // Transform the data to include slide counts if we're not fetching annotations
      if (withAnnotations) {
        // When withAnnotations is true, the API should include annotation data
        const casesWithSlides: CaseWithSlides[] = data.cases.map(
          (caseItem) => ({
            ...caseItem,
            slideCount:
              caseItem.slideCount || (caseItem as any).slideCount || 0,
            slides: [], // We don't need individual slides when just showing counts
            annotationCount: (caseItem as any).annotationCount || 0,
            slidesWithAnnotations: (caseItem as any).slidesWithAnnotations || 0,
          })
        );

        return {
          cases: casesWithSlides,
          pagination: data.pagination,
        };
      } else {
        // For non-annotation mode, we'll need to fetch slides separately
        // This is handled in a separate useEffect below for now
        return data;
      }
    },
  });

  // For cases where we need slide counts but not annotations,
  // we'll still need to fetch slides. This is a performance optimization opportunity.
  const [casesWithSlides, setCasesWithSlides] = useState<CaseWithSlides[]>([]);

  useEffect(() => {
    const fetchSlidesForCases = async () => {
      if (!queryResult.data?.cases || withAnnotations) {
        setCasesWithSlides((queryResult.data?.cases as CaseWithSlides[]) || []);
        return;
      }

      try {
        // Fetch slides for each case in parallel
        const casesWithSlidesData = await Promise.all(
          queryResult.data.cases.map(async (caseItem) => {
            try {
              const slidesData = await apiFetch<CaseSlidesApiResponse>(
                `/api/v1/cases/${caseItem.caseUid}/slides`
              );

              return {
                ...caseItem,
                slideCount: slidesData.slides.length,
                slides: slidesData.slides,
              } as CaseWithSlides;
            } catch (err) {
              console.error(
                `Failed to fetch slides for case ${caseItem.caseUid}:`,
                err
              );
              return {
                ...caseItem,
                slideCount: 0,
                slides: [],
              } as CaseWithSlides;
            }
          })
        );

        setCasesWithSlides(casesWithSlidesData);
      } catch (error) {
        console.error("Failed to fetch slides for cases:", error);
        setCasesWithSlides((queryResult.data.cases as CaseWithSlides[]) || []);
      }
    };

    fetchSlidesForCases();
  }, [queryResult.data?.cases, withAnnotations]);

  return {
    cases: withAnnotations
      ? (queryResult.data?.cases as CaseWithSlides[]) || []
      : casesWithSlides,
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

interface UseCasesByStudyOptions {
  page?: number;
  limit?: number;
  q?: string; // General search query
  name?: string; // Filter by name
  status?: string; // Filter by status
  has_vector_annotations?: string; // Filter by vector annotations (true/false)
  has_raster_annotations?: string; // Filter by raster annotations (true/false)
  sort?: string; // Sort field
  dir?: string; // Sort direction
  withAnnotations?: boolean; // Whether to fetch annotation counts
}

export function useCasesByStudy(
  studyUid: string | undefined,
  options: UseCasesByStudyOptions = {}
) {
  const {
    page = 1,
    limit = 100,
    q = "",
    name = "",
    status = "",
    has_vector_annotations = "",
    has_raster_annotations = "",
    sort = "",
    dir = "",
    withAnnotations = false,
  } = options;

  // Create a stable query key for study-specific cases
  const queryKey = useMemo(
    () =>
      studyUid
        ? [
            ...queryKeys.cases.list(),
            "study",
            studyUid,
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
          ]
        : null,
    [
      studyUid,
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
    if (!studyUid) return null;

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

    return `/api/v1/studies/${studyUid}/cases?${params.toString()}`;
  }, [
    studyUid,
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

  const queryResult = useApiQuery<CasesApiResponse>(
    queryKey || ["cases", "study", "disabled"],
    url || "",
    {
      enabled: !!studyUid,
      staleTime: 1000 * 60 * 2, // 2 minutes
      placeholderData: {
        cases: [],
        pagination: {
          page: 1,
          limit: 100,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
    }
  );

  // Handle slide fetching for non-annotation mode
  const [casesWithSlides, setCasesWithSlides] = useState<CaseWithSlides[]>([]);

  useEffect(() => {
    const fetchSlidesForStudyCases = async () => {
      if (!queryResult.data?.cases || withAnnotations) {
        if (withAnnotations) {
          // Convert cases to include annotation data
          const casesWithAnnotationData: CaseWithSlides[] =
            queryResult.data?.cases.map((caseItem) => ({
              ...caseItem,
              slideCount:
                caseItem.slideCount || (caseItem as any).slideCount || 0,
              slides: [],
              annotationCount: (caseItem as any).annotationCount || 0,
              slidesWithAnnotations:
                (caseItem as any).slidesWithAnnotations || 0,
            })) || [];
          setCasesWithSlides(casesWithAnnotationData);
        } else {
          setCasesWithSlides(
            (queryResult.data?.cases as CaseWithSlides[]) || []
          );
        }
        return;
      }

      try {
        // Fetch slides for each case in parallel
        const casesWithSlidesData = await Promise.all(
          queryResult.data.cases.map(async (caseItem) => {
            try {
              const slidesData = await apiFetch<CaseSlidesApiResponse>(
                `/api/v1/cases/${caseItem.caseUid}/slides`
              );

              return {
                ...caseItem,
                slideCount: slidesData.slides.length,
                slides: slidesData.slides,
              } as CaseWithSlides;
            } catch (err) {
              console.error(
                `Failed to fetch slides for case ${caseItem.caseUid}:`,
                err
              );
              return {
                ...caseItem,
                slideCount: 0,
                slides: [],
              } as CaseWithSlides;
            }
          })
        );

        setCasesWithSlides(casesWithSlidesData);
      } catch (error) {
        console.error("Failed to fetch slides for study cases:", error);
        setCasesWithSlides((queryResult.data.cases as CaseWithSlides[]) || []);
      }
    };

    fetchSlidesForStudyCases();
  }, [queryResult.data?.cases, withAnnotations]);

  return {
    cases: casesWithSlides,
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

interface CaseNeighbor {
  caseUid: string;
  name: string;
}

interface CaseNeighborsData {
  prev: CaseNeighbor | null;
  next: CaseNeighbor | null;
  number: number;
  count: number;
}

export interface CaseNeighborsOptions {
  q?: string; // General search query
  name?: string; // Filter by name
  status?: string; // Filter by status
  has_vector_annotations?: string; // Filter by vector annotations (true/false)
  has_raster_annotations?: string; // Filter by raster annotations (true/false)
  sort?: string; // Sort field
  dir?: string; // Sort direction
}

export function useCaseNeighbors(
  studyUid: string | undefined,
  caseUid: string | undefined,
  options: CaseNeighborsOptions = {}
) {
  const {
    q = "",
    name = "",
    status = "",
    has_vector_annotations = "",
    has_raster_annotations = "",
    sort = "",
    dir = "",
  } = options;

  // Create query key for case neighbors
  const queryKey = useMemo(
    () =>
      studyUid && caseUid
        ? [
            ...queryKeys.cases.detail(caseUid),
            "neighbors",
            studyUid,
            {
              q,
              name,
              status,
              has_vector_annotations,
              has_raster_annotations,
              sort,
              dir,
            },
          ]
        : null,
    [
      studyUid,
      caseUid,
      q,
      name,
      status,
      has_vector_annotations,
      has_raster_annotations,
      sort,
      dir,
    ]
  );

  // Build the URL with query parameters
  const url = useMemo(() => {
    if (!studyUid || !caseUid) return null;

    const params = new URLSearchParams();
    if (q) params.append("q", q);
    if (name) params.append("name", name);
    if (status) params.append("status", status);
    if (has_vector_annotations)
      params.append("has_vector_annotations", has_vector_annotations);
    if (has_raster_annotations)
      params.append("has_raster_annotations", has_raster_annotations);
    if (sort) params.append("sort", sort);
    if (dir) params.append("dir", dir);

    const queryString = params.toString();
    return `/api/v1/studies/${studyUid}/cases/${caseUid}/neighbors${
      queryString ? `?${queryString}` : ""
    }`;
  }, [
    studyUid,
    caseUid,
    q,
    name,
    status,
    has_vector_annotations,
    has_raster_annotations,
    sort,
    dir,
  ]);

  const queryResult = useApiQuery<CaseNeighborsData>(
    queryKey || ["case-neighbors", "disabled"],
    url || "",
    {
      enabled: !!(studyUid && caseUid),
      staleTime: 1000 * 60 * 5, // 5 minutes - neighbors don't change often
    }
  );

  return {
    neighbors: queryResult.data || null,
    loading: queryResult.isLoading,
    error: queryResult.error?.message || null,
    refetch: queryResult.refetch,
  };
}
