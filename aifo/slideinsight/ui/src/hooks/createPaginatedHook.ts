// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { UseQueryOptions } from "@tanstack/react-query";
import { useApiQuery } from "../utils/apiQueries";
import { buildQueryUrl } from "../utils/url";
import { ApiError } from "../utils/fetchUtils";
import type { PaginatedResult, PaginationInfo } from "./types";

/**
 * Factory function to create paginated hooks for API endpoints
 * @param endpoint - The API endpoint path
 * @returns A paginated hook function
 */
export function createPaginatedHook<
  T,
  TResponse = { items: T[]; pagination: PaginationInfo }
>(endpoint: string) {
  return function usePaginated(
    params: Record<string, any> = {},
    options?: Omit<UseQueryOptions<TResponse, ApiError>, "queryKey" | "queryFn">
  ): PaginatedResult<T> {
    const { page = 1, limit = 20, ...filters } = params;
    const url = buildQueryUrl(endpoint, { page, limit, ...filters });
    const queryKey = [endpoint, params];

    const query = useApiQuery<TResponse>(queryKey, url, {
      staleTime: 1000 * 60 * 2, // 2 minutes default
      ...options,
    });

    // Helper to extract data from response
    const extractData = (response: TResponse | undefined): T[] => {
      if (!response) return [];
      if (
        typeof response === "object" &&
        response !== null &&
        "items" in response
      ) {
        return (response as any).items || [];
      }
      // Handle different response formats - look for common array property names
      if (
        typeof response === "object" &&
        response !== null &&
        "data" in response
      ) {
        return (response as any).data || [];
      }
      if (
        typeof response === "object" &&
        response !== null &&
        "results" in response
      ) {
        return (response as any).results || [];
      }
      return [];
    };

    // Helper to extract pagination from response
    const extractPagination = (
      response: TResponse | undefined
    ): PaginationInfo => {
      if (!response)
        return {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        };
      if (
        typeof response === "object" &&
        response !== null &&
        "pagination" in response
      ) {
        return (response as any).pagination;
      }
      return {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      };
    };

    return {
      data: extractData(query.data),
      pagination: extractPagination(query.data),
      loading: query.isLoading,
      error: query.error?.message ?? null,
      refetch: query.refetch,
      isStale: query.isStale,
      isFetching: query.isFetching,
    };
  };
}
