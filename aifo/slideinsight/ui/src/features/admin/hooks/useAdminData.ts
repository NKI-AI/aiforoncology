// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useQueries } from "@tanstack/react-query";
import { queryKeys, useApiQuery } from "../../../utils/apiQueries";
import { apiFetch } from "../../../utils/fetchUtils";

export interface User {
  userUid: string;
  email: string;
  firstName: string;
  lastName: string;
  isActive: boolean;
  mustResetPassword: boolean;
  tenantUid: string;
  createdAt: string;
  updatedAt: string;
}

export interface Tenant {
  tenantUid: string;
  name: string;
  description: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface Algorithm {
  id: string;
  tenantId: number;
  tenantName?: string;
  name: string;
  description?: string;
  version: string;
  endpointUrl: string;
  httpMethod: string;
  executionMode: "BATCH" | "STREAM";
  progressTransport: "WEBSOCKET" | "SSE";
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface AlgorithmRun {
  id: string;
  algorithmId: string;
  caseId?: string;
  imageIds?: string[];
  regions?: Record<string, any>;
  parameters?: Record<string, any>;
  executionMode: "BATCH" | "STREAM";
  status: "QUEUED" | "RUNNING" | "SUCCEEDED" | "FAILED";
  progress: number;
  resultUri?: string;
  errorInfo?: Record<string, any>;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  updatedAt: string;
}

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

interface UsersResponse {
  users: User[];
  pagination: any;
}

interface TenantsResponse {
  tenants: Tenant[];
  pagination: any;
}

interface StudiesResponse {
  studies: Study[];
  pagination: any;
}

interface CountResponse {
  count: number;
}

interface AdminData {
  users: User[];
  tenants: Tenant[];
  studies: Study[];
  slidesCount: number;
  casesCount: number;
  studiesCount: number;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useAdminData(): AdminData {
  // Use parallel queries for better performance with standardized retry logic
  const queries = useQueries({
    queries: [
      {
        queryKey: queryKeys.users.list(),
        queryFn: () => apiFetch<UsersResponse>("/api/v1/users"),
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
      },
      {
        queryKey: queryKeys.tenants.list(),
        queryFn: () => apiFetch<TenantsResponse>("/api/v1/tenants"),
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
      },
      {
        queryKey: queryKeys.studies.list(),
        queryFn: () => apiFetch<StudiesResponse>("/api/v1/studies"),
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
      },
      {
        queryKey: [...queryKeys.studies.all, "count"],
        queryFn: () => apiFetch<CountResponse>("/api/v1/studies/count"),
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
        staleTime: 1000 * 60 * 10,
      },
      {
        queryKey: [...queryKeys.slides.all, "count"],
        queryFn: () => apiFetch<CountResponse>("/api/v1/slides/count"),
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
        staleTime: 1000 * 60 * 10, // 10 minutes - counts change less frequently
      },
      {
        queryKey: [...queryKeys.cases.all, "count"],
        queryFn: () => apiFetch<CountResponse>("/api/v1/cases/count"),
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
        staleTime: 1000 * 60 * 10, // 10 minutes - counts change less frequently
      },
    ],
  });

  const [
    usersQuery,
    tenantsQuery,
    studiesQuery,
    studiesCountQuery,
    slidesCountQuery,
    casesCountQuery,
  ] = queries;
  // Calculate aggregate loading state
  const loading = queries.some((query) => query.isLoading);

  // Calculate aggregate error state
  const error = queries.find((query) => query.error)?.error?.message || null;

  // Refetch function for all queries
  const refetch = async () => {
    await Promise.all(queries.map((query) => query.refetch()));
  };

  return {
    users: usersQuery.data?.users || [],
    tenants: tenantsQuery.data?.tenants || [],
    studies: studiesQuery.data?.studies || [],
    slidesCount: slidesCountQuery.data?.count || 0,
    casesCount: casesCountQuery.data?.count || 0,
    studiesCount: studiesCountQuery.data?.count || 0,
    loading,
    error,
    refetch,
  };
}
