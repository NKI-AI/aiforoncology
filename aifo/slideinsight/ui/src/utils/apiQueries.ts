// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import {
  useQuery,
  useMutation,
  useQueryClient,
  UseQueryOptions,
  UseMutationOptions,
} from "@tanstack/react-query";
import { apiFetch, ApiError } from "./fetchUtils";

// Generic API query hook
export function useApiQuery<TData = unknown, TError = ApiError>(
  queryKey: readonly unknown[],
  url: string,
  options?: Omit<UseQueryOptions<TData, TError>, "queryKey" | "queryFn">
) {
  return useQuery<TData, TError>({
    queryKey,
    queryFn: () => apiFetch<TData>(url),
    retry: (failureCount, error: any) => {
      // Don't retry on 401, 403, or 404 errors
      if (
        error?.status === 401 ||
        error?.status === 403 ||
        error?.status === 404
      ) {
        return false;
      }
      return failureCount < 3;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    ...options,
  });
}

// Generic API mutation hook
export function useApiMutation<
  TData = unknown,
  TVariables = unknown,
  TError = ApiError
>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options?: UseMutationOptions<TData, TError, TVariables>
) {
  return useMutation<TData, TError, TVariables>({
    mutationFn,
    retry: false,
    ...options,
  });
}

// Helper to create standard API mutation functions
export const createApiMutation = {
  post:
    <TData = unknown, TBody = unknown>(url: string) =>
    (body: TBody): Promise<TData> =>
      apiFetch<TData>(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),

  put:
    <TData = unknown, TBody = unknown>(url: string) =>
    (body: TBody): Promise<TData> =>
      apiFetch<TData>(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),

  patch:
    <TData = unknown, TBody = unknown>(url: string) =>
    (body: TBody): Promise<TData> =>
      apiFetch<TData>(url, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),

  delete:
    <TData = unknown>(url: string) =>
    (): Promise<TData> =>
      apiFetch<TData>(url, { method: "DELETE" }),
};

// Helper to create query keys in a consistent way
const createQueryKey = (base: string, params?: Record<string, any>) => {
  const key: (string | Record<string, any>)[] = [base];
  if (params) {
    const sortedParams = Object.keys(params)
      .sort()
      .reduce((acc, k) => {
        if (params[k] !== undefined && params[k] !== null && params[k] !== "") {
          acc[k] = params[k];
        }
        return acc;
      }, {} as Record<string, any>);

    if (Object.keys(sortedParams).length > 0) {
      key.push(sortedParams);
    }
  }
  return key;
};

// Helper hook to invalidate queries by pattern
// export function useInvalidateQueries() {
//     const queryClient = useQueryClient();

//     return (queryKey: readonly unknown[]) => {
//         queryClient.invalidateQueries({ queryKey });
//     };
// }

// Helper hook to remove queries by pattern
// export function useRemoveQueries() {
//     const queryClient = useQueryClient();

//     return (queryKey: readonly unknown[]) => {
//         queryClient.removeQueries({ queryKey });
//     };
// }

// Common query key factories for API endpoints
export const queryKeys = {
  studies: {
    all: ["studies"] as const,
    lists: () => [...queryKeys.studies.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.studies.lists(), params] as const,
    details: () => [...queryKeys.studies.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.studies.details(), id] as const,
    navigation: (studyUid: string) =>
      [...queryKeys.studies.detail(studyUid), "navigation"] as const,
  },
  cases: {
    all: ["cases"] as const,
    lists: () => [...queryKeys.cases.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.cases.lists(), params] as const,
    details: () => [...queryKeys.cases.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.cases.details(), id] as const,
    slides: (caseUid: string) =>
      [...queryKeys.cases.detail(caseUid), "slides"] as const,
  },
  slides: {
    all: ["slides"] as const,
    lists: () => [...queryKeys.slides.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.slides.lists(), params] as const,
    details: () => [...queryKeys.slides.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.slides.details(), id] as const,
    metadata: (slideUid: string) =>
      [...queryKeys.slides.detail(slideUid), "metadata"] as const,
    annotations: (slideUid: string) =>
      [...queryKeys.slides.detail(slideUid), "annotations"] as const,
  },
  tenants: {
    all: ["tenants"] as const,
    lists: () => [...queryKeys.tenants.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.tenants.lists(), params] as const,
    details: () => [...queryKeys.tenants.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.tenants.details(), id] as const,
    domains: (tenantUid: string) =>
      [...queryKeys.tenants.detail(tenantUid), "domains"] as const,
  },
  users: {
    all: ["users"] as const,
    lists: () => [...queryKeys.users.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.users.lists(), params] as const,
    details: () => [...queryKeys.users.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.users.details(), id] as const,
  },
  algorithms: {
    all: ["algorithms"] as const,
    lists: () => [...queryKeys.algorithms.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.algorithms.lists(), params] as const,
    details: () => [...queryKeys.algorithms.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.algorithms.details(), id] as const,
    runs: (algorithmId: string) =>
      [...queryKeys.algorithms.detail(algorithmId), "runs"] as const,
  },
  runs: {
    all: ["runs"] as const,
    lists: () => [...queryKeys.runs.all, "list"] as const,
    list: (params?: Record<string, any>) =>
      [...queryKeys.runs.lists(), params] as const,
    details: () => [...queryKeys.runs.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.runs.details(), id] as const,
    outputs: (runId: string) =>
      [...queryKeys.runs.detail(runId), "outputs"] as const,
  },
} as const;
