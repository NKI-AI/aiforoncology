// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useEffect, useCallback } from "react";
import { apiFetch } from "../utils/fetchUtils";
import type { PaginationInfo, BaseApiOptions } from "./types";

// Re-export types for external use
export type { PaginationInfo };

export interface BasePaginatedOptions extends BaseApiOptions {}

interface UsePaginatedApiConfig<
  TData,
  TResponse,
  TOptions extends BasePaginatedOptions
> {
  endpoint: string;
  queryBuilder: (options: TOptions) => Record<string, string>;
  queryKeyFactory?: (options: TOptions) => readonly unknown[];
  dataExtractor: (response: TResponse) => {
    items: TData[];
    pagination: PaginationInfo;
  };
  postProcessor?: (items: TData[]) => Promise<TData[]>;
  errorMessage?: string;
}

interface UsePaginatedApiResult<TData> {
  data: TData[];
  pagination: PaginationInfo;
  loading: boolean;
  error: string | null;
  refetch: (newPage?: number, newLimit?: number) => Promise<void>;
}

export function usePaginatedApi<
  TData,
  TResponse,
  TOptions extends BasePaginatedOptions
>(
  options: TOptions,
  config: UsePaginatedApiConfig<TData, TResponse, TOptions>
): UsePaginatedApiResult<TData> {
  // Destructure options to get stable primitive values for dependencies
  const {
    page = 1,
    limit = 100,
    q = "",
    sort = "",
    dir = "",
    ...otherOptions
  } = options;

  const [data, setData] = useState<TData[]>([]);
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    limit: 100,
    total: 0,
    totalPages: 0,
    hasNext: false,
    hasPrev: false,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(
    async (currentPage: number = page, currentLimit: number = limit) => {
      try {
        setLoading(true);
        setError(null);

        // Build query parameters
        const params = new URLSearchParams();
        params.append("page", currentPage.toString());
        params.append("limit", currentLimit.toString());

        // Add custom query parameters
        const queryParams = config.queryBuilder({
          ...options,
          page: currentPage,
          limit: currentLimit,
        });
        Object.entries(queryParams).forEach(([key, value]) => {
          if (value && key !== "page" && key !== "limit") {
            params.append(key, value);
          }
        });

        const response = await apiFetch<TResponse>(
          `${config.endpoint}?${params.toString()}`
        );
        const { items, pagination: responsePagination } =
          config.dataExtractor(response);

        // Apply post-processing if provided
        const processedItems = config.postProcessor
          ? await config.postProcessor(items)
          : items;

        setData(processedItems);
        setPagination(responsePagination);
      } catch (err) {
        console.error(`Failed to fetch data from ${config.endpoint}:`, err);
        setError(
          config.errorMessage || "Failed to load data. Please try again later."
        );
      } finally {
        setLoading(false);
      }
    },
    [
      // Use individual primitive values instead of the options object
      page,
      limit,
      q,
      sort,
      dir,
      // Include otherOptions serialized to catch any additional properties
      JSON.stringify(otherOptions),
      // Config should be stable due to useMemo in calling hooks
      config.endpoint,
      config.queryBuilder,
      config.dataExtractor,
      config.postProcessor,
      config.errorMessage,
    ]
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const refetch = useCallback(
    (newPage?: number, newLimit?: number) => {
      return fetchData(newPage, newLimit);
    },
    [fetchData]
  );

  return { data, pagination, loading, error, refetch };
}
