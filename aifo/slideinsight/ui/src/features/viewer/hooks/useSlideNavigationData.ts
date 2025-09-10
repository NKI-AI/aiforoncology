// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useEffect, useMemo } from "react";
import { useSearch, useLocation } from "@tanstack/react-router";
import { apiFetch } from "@/utils/fetchUtils";
import type { SlideNavigationItem } from "@/features/viewer/contexts/SlideNavigationContext";

interface NavigationSource {
  type: "study" | "case-list" | "slide-list" | "all-slides" | "filtered-subset";
  id?: string;
  name?: string;
  filters?: Record<string, any>;
}

interface UseSlideNavigationDataOptions {
  slideUid: string | null;
  caseUid?: string;
  searchQuery?: string;
  searchName?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  limit?: number;
  isSubsetMode?: boolean;
}

interface NavigationResponse {
  slides: SlideNavigationItem[];
  source: NavigationSource;
}

export function useSlideNavigationData({
  slideUid,
  caseUid,
  searchQuery = "",
  searchName = "",
  sortBy = "",
  sortDir = "asc",
  page = 1,
  limit = 1000,
  isSubsetMode = false,
}: UseSlideNavigationDataOptions) {
  const [navigationData, setNavigationData] =
    useState<NavigationResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const location = useLocation();

  // Check if we have active filters that would justify subset mode
  const hasActiveFilters = searchQuery || searchName || sortBy;
  const shouldUseSubsetMode = isSubsetMode && hasActiveFilters;

  // Extract context from URL search params or referrer
  const sourceContext = useMemo(() => {
    if (!slideUid) return null;

    const searchParams = new URLSearchParams(location.search);
    const source = searchParams.get("source");
    const sourceId = searchParams.get("sourceId");
    const sourceName = searchParams.get("sourceName");

    // Parse filters from URL
    const filters: Record<string, any> = {};
    for (const [key, value] of searchParams.entries()) {
      if (key.startsWith("filter_")) {
        filters[key.replace("filter_", "")] = value;
      }
    }

    if (
      source &&
      ["study", "case-list", "slide-list", "all-slides"].includes(source)
    ) {
      return {
        type: source as NavigationSource["type"],
        id: sourceId || undefined,
        name: sourceName || undefined,
        filters: Object.keys(filters).length > 0 ? filters : undefined,
      };
    }

    return null;
  }, [location.search, slideUid]);

  useEffect(() => {
    if (!slideUid) {
      setNavigationData(null);
      return;
    }

    const fetchNavigationData = async () => {
      try {
        setLoading(true);
        setError(null);

        let response: NavigationResponse;

        // Priority 1: If we have a caseUid, fetch slides for this specific case
        if (caseUid) {
          try {
            const slidesResponse = (await apiFetch(
              `/api/v1/cases/${caseUid}/slides`
            )) as any;
            const slides = slidesResponse.slides || slidesResponse || [];

            // Also get case info for context
            const caseInfo = (await apiFetch(
              `/api/v1/cases/${caseUid}`
            )) as any;

            const navigationSlides = slides.map((slide: any) => ({
              slideUid: slide.slideUid,
              slideName: slide.slideName,
              caseUid: caseUid,
              caseName: caseInfo.name,
              studyUid: caseInfo.studyUid,
              studyName: caseInfo.studyName,
            }));

            response = {
              slides: navigationSlides,
              source: {
                type: "study",
                id: caseInfo.studyUid,
                name: caseInfo.studyName,
              },
            };
          } catch (caseError) {
            console.error("Failed to fetch slides for case:", caseError);
            throw caseError;
          }
        }
        // Priority 2: If subset mode is enabled with filters, force filtered navigation
        else if (shouldUseSubsetMode) {
          console.log("Using subset mode with filters:", {
            searchQuery,
            searchName,
            sortBy,
            sortDir,
          });
          const queryParams = buildSlidesQueryParams({
            searchQuery,
            searchName,
            sortBy,
            sortDir,
            page,
            limit,
          });
          const slidesUrl =
            "/api/v1/slides" + (queryParams ? "?" + queryParams : "");
          console.log(
            "Fetching filtered slides for subset navigation:",
            slidesUrl
          );

          const allSlides = (await apiFetch(slidesUrl)) as any;
          const slidesData = allSlides.slides || allSlides.data || allSlides;
          console.log(
            "Subset mode: fetched",
            slidesData.length,
            "filtered slides"
          );

          const slides = slidesData.map((slide: any) => ({
            slideUid: slide.slideUid,
            slideName: slide.slideName,
            caseUid: slide.caseUid,
            caseName: slide.caseName,
            studyUid: slide.studyUid,
            studyName: slide.studyName,
          }));

          response = {
            slides,
            source: {
              type: "filtered-subset",
              filters: { searchQuery, searchName, sortBy, sortDir },
            },
          };
        } else if (sourceContext) {
          // Fetch slides based on source context
          response = await fetchSlidesForContext(sourceContext, {
            searchQuery,
            searchName,
            sortBy,
            sortDir,
            page,
            limit,
          });
        } else {
          // Fallback: try to determine context from slide metadata
          response = await fetchSlideContextAndNavigation(slideUid, {
            searchQuery,
            searchName,
            sortBy,
            sortDir,
            page,
            limit,
          });
        }

        setNavigationData(response);
      } catch (err) {
        console.error("Failed to fetch navigation data:", err);
        setError("Failed to load navigation data");

        // Fallback: provide minimal navigation with just the current slide
        if (slideUid) {
          setNavigationData({
            slides: [
              {
                slideUid,
                slideName: slideUid, // Will be updated when slide metadata loads
              },
            ],
            source: { type: "all-slides" },
          });
        }
      } finally {
        setLoading(false);
      }
    };

    fetchNavigationData();
  }, [
    slideUid,
    caseUid,
    sourceContext,
    searchQuery,
    searchName,
    sortBy,
    sortDir,
    page,
    limit,
    shouldUseSubsetMode,
  ]);

  return {
    navigationData,
    loading,
    error,
  };
}

// Helper to build query parameters for API calls
function buildSlidesQueryParams(filters: {
  searchQuery?: string;
  searchName?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  limit?: number;
}): string {
  const params = new URLSearchParams();

  if (filters.page) params.append("page", String(filters.page));
  if (filters.limit) params.append("limit", String(filters.limit));
  if (filters.searchQuery) params.append("q", filters.searchQuery);
  if (filters.searchName) params.append("name", filters.searchName);
  if (filters.sortBy) params.append("sort", filters.sortBy);
  if (filters.sortDir) params.append("dir", filters.sortDir);

  return params.toString();
}

async function fetchSlidesForContext(
  source: NavigationSource,
  filters: {
    searchQuery?: string;
    searchName?: string;
    sortBy?: string;
    sortDir?: string;
    page?: number;
    limit?: number;
  }
): Promise<NavigationResponse> {
  let slides: SlideNavigationItem[] = [];

  switch (source.type) {
    case "study":
      if (source.id) {
        const cases = (await apiFetch(
          `/api/v1/studies/${source.id}/cases`
        )) as any;
        slides = flattenCasesToSlides(cases.cases || cases, {
          studyUid: source.id,
          studyName: source.name,
        });
      }
      break;

    case "case-list":
      // Fetch slides from all cases (admin view) - respect filters
      const casesUrl =
        "/api/v1/cases" +
        (source.filters
          ? "?" + new URLSearchParams(source.filters).toString()
          : "");
      const allCases = (await apiFetch(casesUrl)) as any;
      slides = flattenCasesToSlides(
        allCases.cases || allCases.data || allCases
      );
      break;

    case "slide-list":
    case "all-slides":
      // Fetch all slides directly with filters
      const queryParams = buildSlidesQueryParams(filters);
      const slidesUrl =
        "/api/v1/slides" + (queryParams ? "?" + queryParams : "");
      console.log("Fetching slides with URL:", slidesUrl); // Debug log
      const allSlides = (await apiFetch(slidesUrl)) as any;
      const slidesData = allSlides.slides || allSlides.data || allSlides;
      console.log("Fetched slides:", slidesData.length); // Debug log
      slides = slidesData.map((slide: any) => ({
        slideUid: slide.slideUid,
        slideName: slide.slideName,
        caseUid: slide.caseUid,
        caseName: slide.caseName,
        studyUid: slide.studyUid,
        studyName: slide.studyName,
      }));
      break;
  }

  return {
    slides,
    source,
  };
}

async function fetchSlideContextAndNavigation(
  slideUid: string,
  filters: {
    searchQuery?: string;
    searchName?: string;
    sortBy?: string;
    sortDir?: string;
    page?: number;
    limit?: number;
  }
): Promise<NavigationResponse> {
  try {
    // First, get slide metadata to understand its context
    console.log("Fetching slide metadata for:", slideUid);
    const slideInfo = (await apiFetch(`/api/v1/slides/${slideUid}`)) as any;
    console.log("Slide info:", slideInfo);

    if (slideInfo.caseUid) {
      console.log("Slide has caseUid:", slideInfo.caseUid);
      try {
        // If we have a case short UID, get all slides from that case
        const caseInfo = (await apiFetch(
          `/api/v1/cases/${slideInfo.caseUid}`
        )) as any;
        console.log("Case info:", caseInfo);

        const slides =
          caseInfo.slides?.map((slide: any) => ({
            slideUid: slide.slideUid,
            slideName: slide.slideName,
            caseUid: slideInfo.caseUid,
            caseName: caseInfo.name,
            studyUid: caseInfo.studyUid,
            studyName: caseInfo.studyName,
          })) || [];

        console.log("Successfully fetched case slides:", slides.length);

        if (slides.length > 0) {
          return {
            slides,
            source: {
              type: "study",
              id: caseInfo.studyUid,
              name: caseInfo.studyName,
            },
          };
        }
      } catch (caseError) {
        console.warn("Failed to fetch case info:", caseError);
      }
    } else {
      console.log(
        "No caseUid found in slide info, trying alternative case detection"
      );

      // Alternative: try to find case context from the slide data itself
      if (slideInfo.caseUid) {
        console.log("Slide has caseUid:", slideInfo.caseUid);
        try {
          const caseInfo = (await apiFetch(
            `/api/v1/cases/${slideInfo.caseUid}`
          )) as any;
          const slides =
            caseInfo.slides?.map((slide: any) => ({
              slideUid: slide.slideUid,
              slideName: slide.slideName,
              caseUid: slideInfo.caseUid,
              caseName: caseInfo.name,
              studyUid: caseInfo.studyUid,
              studyName: caseInfo.studyName,
            })) || [];

          if (slides.length > 0) {
            return {
              slides,
              source: {
                type: "study",
                id: caseInfo.studyUid,
                name: caseInfo.studyName,
              },
            };
          }
        } catch (caseError) {
          console.warn("Failed to fetch case info with caseUid:", caseError);
        }
      }
    }

    // Fallback 1: Try to get all slides with filters applied
    console.log("Trying fallback: fetch all slides with filters");
    try {
      const queryParams = buildSlidesQueryParams(filters);
      const slidesUrl =
        "/api/v1/slides" + (queryParams ? "?" + queryParams : "");
      console.log("Fetching slides with URL:", slidesUrl); // Debug log
      const allSlidesResponse = (await apiFetch(slidesUrl)) as any;
      const allSlides =
        allSlidesResponse.slides || allSlidesResponse.data || allSlidesResponse;

      console.log("All slides response:", allSlidesResponse);
      console.log("Filtered slides count:", allSlides.length); // Debug log

      if (Array.isArray(allSlides) && allSlides.length > 0) {
        // Convert to navigation items - ensure case context is preserved
        const slides = allSlides.map((slide: any) => ({
          slideUid: slide.slideUid,
          slideName: slide.slideName,
          caseUid: slide.caseUid,
          caseName: slide.caseName,
          studyUid: slide.studyUid,
          studyName: slide.studyName,
        }));

        console.log(
          "Fallback slides with case context:",
          slides.filter((s) => s.caseUid).length,
          "of",
          slides.length
        );

        return {
          slides,
          source: { type: "all-slides" },
        };
      }
    } catch (allSlidesError) {
      console.warn("Failed to fetch filtered slides:", allSlidesError);
    }

    // Fallback 2: just return this slide
    console.log("Using final fallback: single slide");
    return {
      slides: [
        {
          slideUid,
          slideName: slideInfo.slideName,
          caseUid: slideInfo.caseUid,
          caseName: slideInfo.caseName,
          studyUid: slideInfo.studyUid,
          studyName: slideInfo.studyName,
        },
      ],
      source: { type: "all-slides" },
    };
  } catch (err) {
    console.error("Error in fetchSlideContextAndNavigation:", err);
    // If we can't get context, just return the current slide
    return {
      slides: [
        {
          slideUid,
          slideName: slideUid,
        },
      ],
      source: { type: "all-slides" },
    };
  }
}

function flattenCasesToSlides(
  cases: any[],
  context?: { studyUid?: string; studyName?: string }
): SlideNavigationItem[] {
  const slides: SlideNavigationItem[] = [];

  cases.forEach((caseItem) => {
    if (caseItem.slides && Array.isArray(caseItem.slides)) {
      caseItem.slides.forEach((slide: any) => {
        slides.push({
          slideUid: slide.slideUid,
          slideName: slide.slideName,
          caseUid: caseItem.caseUid || caseItem.id,
          caseName: caseItem.name,
          studyUid: context?.studyUid || caseItem.studyUid,
          studyName: context?.studyName || caseItem.studyName,
        });
      });
    }
  });

  return slides;
}
