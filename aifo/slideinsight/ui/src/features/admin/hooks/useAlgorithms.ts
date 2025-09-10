// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import {
  usePaginatedApi,
  type BasePaginatedOptions,
  type PaginationInfo,
} from "@/hooks/usePaginatedApi";
import { Algorithm } from "./useAdminData";

interface AlgorithmsApiResponse {
  algorithms: Algorithm[];
  pagination: PaginationInfo;
}

interface UseAlgorithmsOptions extends BasePaginatedOptions {
  name?: string; // Filter by name
  version?: string; // Filter by version
  executionMode?: "BATCH" | "STREAM"; // Filter by execution mode
}

export function useAlgorithms(options: UseAlgorithmsOptions = {}) {
  const config = useMemo(
    () => ({
      endpoint: "/api/v1/algorithms",
      queryBuilder: (opts: UseAlgorithmsOptions) => {
        const params: Record<string, string> = {};
        if (opts.q) params.q = opts.q;
        if (opts.name) params.name = opts.name;
        if (opts.version) params.version = opts.version;
        if (opts.executionMode) params.executionMode = opts.executionMode;
        if (opts.sort) params.sort = opts.sort;
        if (opts.dir) params.dir = opts.dir;
        return params;
      },
      dataExtractor: (response: AlgorithmsApiResponse) => ({
        items: response.algorithms,
        pagination: response.pagination,
      }),
      errorMessage: "Failed to load algorithms. Please try again later.",
    }),
    []
  );

  const result = usePaginatedApi<
    Algorithm,
    AlgorithmsApiResponse,
    UseAlgorithmsOptions
  >(options, config);

  return {
    algorithms: result.data,
    pagination: result.pagination,
    loading: result.loading,
    error: result.error,
    refetch: result.refetch,
  };
}
