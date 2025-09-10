// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Enhanced API hooks with centralized error handling and type safety
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
  UseQueryOptions,
  UseMutationOptions,
} from "@tanstack/react-query";
import { apiFetch, ApiError } from "../utils/fetchUtils";
import { queryKeys } from "./queryKeys";
import type {
  User,
  Tenant,
  Study,
  StudyWithCasesAndSlides,
  Case,
  CaseWithSlides,
  Slide,
  SlideWithCount,
  Algorithm,
  EmailTemplate,
  Role,
  Permission,
  Group,
  UsersResponse,
  TenantsResponse,
  StudiesResponse,
  CasesResponse,
  SlidesResponse,
  AlgorithmsResponse,
  EmailTemplatesResponse,
  RolesResponse,
  PermissionsResponse,
  GroupsResponse,
  UserQuery,
  TenantQuery,
  StudyQuery,
  CaseQuery,
  SlideQuery,
  AlgorithmQuery,
  EmailTemplateQuery,
  RoleQuery,
  PermissionQuery,
  GroupQuery,
  CreateUserRequest,
  UpdateUserRequest,
  CreateTenantRequest,
  UpdateTenantRequest,
  CreateStudyRequest,
  UpdateStudyRequest,
  CreateCaseRequest,
  UpdateCaseRequest,
  CreateSlideRequest,
  UpdateSlideRequest,
  CreateAlgorithmRequest,
  UpdateAlgorithmRequest,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
  PaginationInfo,
  Setting,
  SettingQuery,
  SettingsResponse,
  CreateSettingRequest,
  UpdateSettingRequest,
  CountResponse,
} from "./models";

// ===== CENTRALIZED ERROR MESSAGES =====

const API_ERROR_MESSAGES = {
  users: {
    fetch: "Failed to load users. Please try again later.",
    create: "Failed to create user. Please check your input and try again.",
    update: "Failed to update user. Please try again.",
    delete: "Failed to delete user. Please try again.",
  },
  tenants: {
    fetch: "Failed to load tenants. Please try again later.",
    create: "Failed to create tenant. Please check your input and try again.",
    update: "Failed to update tenant. Please try again.",
    delete: "Failed to delete tenant. Please try again.",
  },
  studies: {
    fetch: "Failed to load studies. Please try again later.",
    create: "Failed to create study. Please check your input and try again.",
    update: "Failed to update study. Please try again.",
    delete: "Failed to delete study. Please try again.",
  },
  cases: {
    fetch: "Failed to load cases. Please try again later.",
    create: "Failed to create case. Please check your input and try again.",
    update: "Failed to update case. Please try again.",
    delete: "Failed to delete case. Please try again.",
  },
  slides: {
    fetch: "Failed to load slides. Please try again later.",
    create: "Failed to create slide. Please check your input and try again.",
    update: "Failed to update slide. Please try again.",
    delete: "Failed to delete slide. Please try again.",
  },
  algorithms: {
    fetch: "Failed to load algorithms. Please try again later.",
    create:
      "Failed to create algorithm. Please check your input and try again.",
    update: "Failed to update algorithm. Please try again.",
    delete: "Failed to delete algorithm. Please try again.",
  },
  emailTemplates: {
    fetch: "Failed to load email templates. Please try again later.",
    create:
      "Failed to create email template. Please check your input and try again.",
    update: "Failed to update email template. Please try again.",
    delete: "Failed to delete email template. Please try again.",
  },
  roles: {
    fetch: "Failed to load roles. Please try again later.",
    create: "Failed to create role. Please check your input and try again.",
    update: "Failed to update role. Please try again.",
    delete: "Failed to delete role. Please try again.",
  },
  permissions: {
    fetch: "Failed to load permissions. Please try again later.",
    create:
      "Failed to create permission. Please check your input and try again.",
    update: "Failed to update permission. Please try again.",
    delete: "Failed to delete permission. Please try again.",
  },
  groups: {
    fetch: "Failed to load groups. Please try again later.",
    create: "Failed to create group. Please check your input and try again.",
    update: "Failed to update group. Please try again.",
    delete: "Failed to delete group. Please try again.",
  },
  general: {
    network: "Network error. Please check your connection and try again.",
    unauthorized: "You are not authorized to perform this action.",
    forbidden: "Access denied. You don't have permission for this action.",
    notFound: "The requested resource was not found.",
    serverError: "Server error. Please try again later.",
  },
} as const;

// ===== ENHANCED API QUERY HOOK =====

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

// ===== ENHANCED API MUTATION HOOK =====

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

// ===== UTILITY FUNCTIONS =====

/**
 * Build URL with query parameters
 */
function buildQueryUrl(baseUrl: string, params: Record<string, any>): string {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      searchParams.append(key, String(value));
    }
  });

  return searchParams.toString()
    ? `${baseUrl}?${searchParams.toString()}`
    : baseUrl;
}

/**
 * Handle API errors with user-friendly messages
 */
function getErrorMessage(error: any, fallbackMessage: string): string {
  if (error?.status === 401) return API_ERROR_MESSAGES.general.unauthorized;
  if (error?.status === 403) return API_ERROR_MESSAGES.general.forbidden;
  if (error?.status === 404) return API_ERROR_MESSAGES.general.notFound;
  if (error?.status >= 500) return API_ERROR_MESSAGES.general.serverError;

  return error?.message || fallbackMessage;
}

// ===== PAGINATED API RESULT TYPE =====

export interface PaginatedApiResult<TData> {
  data: TData[];
  pagination: PaginationInfo;
  loading: boolean;
  error: string | null;
  refetch: () => void;
  isStale: boolean;
  isFetching: boolean;
}

// ===== USER API HOOKS =====

export function useUsers(
  query: UserQuery = {},
  options?: Omit<
    UseQueryOptions<UsersResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<User> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/users", { page, limit, ...filters });

  const result = useApiQuery<UsersResponse>(queryKeys.users.list(query), url, {
    placeholderData: {
      users: [],
      pagination: {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      },
    },
    ...options,
  });

  return {
    data: result.data?.users || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.users.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateUser(
  options?: UseMutationOptions<User, ApiError, CreateUserRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<User, CreateUserRequest>(
    (userData) =>
      apiFetch<User>("/api/v1/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(userData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch users list
        queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
        options?.onSuccess?.(data, {} as CreateUserRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.users.create);
        console.error("Failed to create user:", message);
        options?.onError?.(error, {} as CreateUserRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateUser(
  userUid: string,
  options?: UseMutationOptions<User, ApiError, UpdateUserRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<User, UpdateUserRequest>(
    (userData) =>
      apiFetch<User>(`/api/v1/users/${userUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(userData),
      }),
    {
      onSuccess: (data) => {
        // Update specific user cache and invalidate lists
        queryClient.setQueryData(queryKeys.users.detail(userUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
        options?.onSuccess?.(data, {} as UpdateUserRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.users.update);
        console.error("Failed to update user:", message);
        options?.onError?.(error, {} as UpdateUserRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteUser(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (userUid) =>
      apiFetch<void>(`/api/v1/users/${userUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, userUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.users.detail(userUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
        options?.onSuccess?.(undefined, userUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.users.delete);
        console.error("Failed to delete user:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

export function useUserByUID(
  userUid: string | undefined,
  options?: Omit<UseQueryOptions<User, ApiError>, "queryKey" | "queryFn">
) {
  return useApiQuery<User>(
    queryKeys.users.detail(userUid || ""),
    `/api/v1/users/${userUid}`,
    {
      enabled: !!userUid,
      staleTime: 5 * 60 * 1000, // 5 minutes
      ...options,
    }
  );
}

// ===== TENANT API HOOKS =====

export function useTenants(
  query: TenantQuery = {},
  options?: Omit<
    UseQueryOptions<TenantsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Tenant> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/tenants", { page, limit, ...filters });

  const result = useApiQuery<TenantsResponse>(
    queryKeys.tenants.list(query),
    url,
    {
      placeholderData: {
        tenants: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data: result.data?.tenants || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.tenants.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateTenant(
  options?: UseMutationOptions<Tenant, ApiError, CreateTenantRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Tenant, CreateTenantRequest>(
    (tenantData) =>
      apiFetch<Tenant>("/api/v1/tenants", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tenantData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch tenants list
        queryClient.invalidateQueries({ queryKey: queryKeys.tenants.all });
        options?.onSuccess?.(data, {} as CreateTenantRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.tenants.create
        );
        console.error("Failed to create tenant:", message);
        options?.onError?.(error, {} as CreateTenantRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateTenant(
  tenantUid: string,
  options?: UseMutationOptions<Tenant, ApiError, UpdateTenantRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Tenant, UpdateTenantRequest>(
    (tenantData) =>
      apiFetch<Tenant>(`/api/v1/tenants/${tenantUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tenantData),
      }),
    {
      onSuccess: (data) => {
        // Update specific tenant cache and invalidate lists
        queryClient.setQueryData(queryKeys.tenants.detail(tenantUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.tenants.lists() });
        options?.onSuccess?.(data, {} as UpdateTenantRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.tenants.update
        );
        console.error("Failed to update tenant:", message);
        options?.onError?.(error, {} as UpdateTenantRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteTenant(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (tenantUid) =>
      apiFetch<void>(`/api/v1/tenants/${tenantUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, tenantUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.tenants.detail(tenantUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.tenants.lists() });
        options?.onSuccess?.(undefined, tenantUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.tenants.delete
        );
        console.error("Failed to delete tenant:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== STUDY API HOOKS =====

export function useStudies(
  query: StudyQuery = {},
  options?: Omit<
    UseQueryOptions<StudiesResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<StudyWithCasesAndSlides> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/studies", { page, limit, ...filters });

  const result = useApiQuery<StudiesResponse>(
    queryKeys.studies.list(query),
    url,
    {
      placeholderData: {
        studies: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data:
      result.data?.studies.map(
        (study) =>
          ({
            ...study,
            // Add additional fields that might be included in StudyWithCasesAndSlides
          } as StudyWithCasesAndSlides)
      ) || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.studies.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateStudy(
  options?: UseMutationOptions<Study, ApiError, CreateStudyRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Study, CreateStudyRequest>(
    (studyData) =>
      apiFetch<Study>("/api/v1/studies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(studyData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch studies list
        queryClient.invalidateQueries({ queryKey: queryKeys.studies.all });
        options?.onSuccess?.(data, {} as CreateStudyRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.studies.create
        );
        console.error("Failed to create study:", message);
        options?.onError?.(error, {} as CreateStudyRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateStudy(
  studyUid: string,
  options?: UseMutationOptions<Study, ApiError, UpdateStudyRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Study, UpdateStudyRequest>(
    (studyData) =>
      apiFetch<Study>(`/api/v1/studies/${studyUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(studyData),
      }),
    {
      onSuccess: (data) => {
        // Update specific study cache and invalidate lists
        queryClient.setQueryData(queryKeys.studies.detail(studyUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.studies.lists() });
        options?.onSuccess?.(data, {} as UpdateStudyRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.studies.update
        );
        console.error("Failed to update study:", message);
        options?.onError?.(error, {} as UpdateStudyRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteStudy(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (studyUid) =>
      apiFetch<void>(`/api/v1/studies/${studyUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, studyUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.studies.detail(studyUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.studies.lists() });
        options?.onSuccess?.(undefined, studyUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.studies.delete
        );
        console.error("Failed to delete study:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== CASE API HOOKS =====

export function useCases(
  query: CaseQuery = {},
  options?: Omit<
    UseQueryOptions<CasesResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<CaseWithSlides> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/cases", { page, limit, ...filters });

  const result = useApiQuery<CasesResponse>(queryKeys.cases.list(query), url, {
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
    ...options,
  });

  return {
    data:
      result.data?.cases.map(
        (caseItem) =>
          ({
            ...caseItem,
            // Add additional fields that might be included in CaseWithSlides
          } as CaseWithSlides)
      ) || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.cases.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateCase(
  options?: UseMutationOptions<Case, ApiError, CreateCaseRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Case, CreateCaseRequest>(
    (caseData) =>
      apiFetch<Case>("/api/v1/cases", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(caseData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch cases list
        queryClient.invalidateQueries({ queryKey: queryKeys.cases.all });
        options?.onSuccess?.(data, {} as CreateCaseRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.cases.create);
        console.error("Failed to create case:", message);
        options?.onError?.(error, {} as CreateCaseRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateCase(
  caseUid: string,
  options?: UseMutationOptions<Case, ApiError, UpdateCaseRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Case, UpdateCaseRequest>(
    (caseData) =>
      apiFetch<Case>(`/api/v1/cases/${caseUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(caseData),
      }),
    {
      onSuccess: (data) => {
        // Update specific case cache and invalidate lists
        queryClient.setQueryData(queryKeys.cases.detail(caseUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.cases.lists() });
        options?.onSuccess?.(data, {} as UpdateCaseRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.cases.update);
        console.error("Failed to update case:", message);
        options?.onError?.(error, {} as UpdateCaseRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteCase(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (caseUid) =>
      apiFetch<void>(`/api/v1/cases/${caseUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, caseUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.cases.detail(caseUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.cases.lists() });
        options?.onSuccess?.(undefined, caseUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.cases.delete);
        console.error("Failed to delete case:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== SLIDE API HOOKS =====

export function useSlides(
  query: SlideQuery = {},
  options?: Omit<
    UseQueryOptions<SlidesResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<SlideWithCount> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/slides", { page, limit, ...filters });

  const result = useApiQuery<SlidesResponse>(
    queryKeys.slides.list(query),
    url,
    {
      placeholderData: {
        slides: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data:
      result.data?.slides.map(
        (slide) =>
          ({
            ...slide,
            // Add additional fields that might be included in SlideWithCount
          } as SlideWithCount)
      ) || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.slides.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateSlide(
  options?: UseMutationOptions<Slide, ApiError, CreateSlideRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Slide, CreateSlideRequest>(
    (slideData) =>
      apiFetch<Slide>("/api/v1/slides", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(slideData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch slides list
        queryClient.invalidateQueries({ queryKey: queryKeys.slides.all });
        options?.onSuccess?.(data, {} as CreateSlideRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.slides.create
        );
        console.error("Failed to create slide:", message);
        options?.onError?.(error, {} as CreateSlideRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateSlide(
  slideUid: string,
  options?: UseMutationOptions<Slide, ApiError, UpdateSlideRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Slide, UpdateSlideRequest>(
    (slideData) =>
      apiFetch<Slide>(`/api/v1/slides/${slideUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(slideData),
      }),
    {
      onSuccess: (data) => {
        // Update specific slide cache and invalidate lists
        queryClient.setQueryData(queryKeys.slides.detail(slideUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.slides.lists() });
        options?.onSuccess?.(data, {} as UpdateSlideRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.slides.update
        );
        console.error("Failed to update slide:", message);
        options?.onError?.(error, {} as UpdateSlideRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteSlide(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (slideUid) =>
      apiFetch<void>(`/api/v1/slides/${slideUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, slideUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.slides.detail(slideUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.slides.lists() });
        options?.onSuccess?.(undefined, slideUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.slides.delete
        );
        console.error("Failed to delete slide:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== ALGORITHM API HOOKS =====

export function useAlgorithms(
  query: AlgorithmQuery = {},
  options?: Omit<
    UseQueryOptions<AlgorithmsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Algorithm> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/algorithms", { page, limit, ...filters });

  const result = useApiQuery<AlgorithmsResponse>(
    queryKeys.algorithms.list(query),
    url,
    {
      placeholderData: {
        algorithms: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data: result.data?.algorithms || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.algorithms.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateAlgorithm(
  options?: UseMutationOptions<Algorithm, ApiError, CreateAlgorithmRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Algorithm, CreateAlgorithmRequest>(
    (algorithmData) =>
      apiFetch<Algorithm>("/api/v1/algorithms", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(algorithmData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch algorithms list
        queryClient.invalidateQueries({ queryKey: queryKeys.algorithms.all });
        options?.onSuccess?.(data, {} as CreateAlgorithmRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.algorithms.create
        );
        console.error("Failed to create algorithm:", message);
        options?.onError?.(error, {} as CreateAlgorithmRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateAlgorithm(
  algorithmId: string,
  options?: UseMutationOptions<Algorithm, ApiError, UpdateAlgorithmRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Algorithm, UpdateAlgorithmRequest>(
    (algorithmData) =>
      apiFetch<Algorithm>(`/api/v1/algorithms/${algorithmId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(algorithmData),
      }),
    {
      onSuccess: (data) => {
        // Update specific algorithm cache and invalidate lists
        queryClient.setQueryData(
          queryKeys.algorithms.detail(algorithmId),
          data
        );
        queryClient.invalidateQueries({
          queryKey: queryKeys.algorithms.lists(),
        });
        options?.onSuccess?.(data, {} as UpdateAlgorithmRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.algorithms.update
        );
        console.error("Failed to update algorithm:", message);
        options?.onError?.(error, {} as UpdateAlgorithmRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteAlgorithm(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (algorithmId) =>
      apiFetch<void>(`/api/v1/algorithms/${algorithmId}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, algorithmId) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.algorithms.detail(algorithmId),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.algorithms.lists(),
        });
        options?.onSuccess?.(undefined, algorithmId, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.algorithms.delete
        );
        console.error("Failed to delete algorithm:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== EMAIL TEMPLATE API HOOKS =====

export function useEmailTemplates(
  query: EmailTemplateQuery = {},
  options?: Omit<
    UseQueryOptions<{ status: string; data: EmailTemplatesResponse }, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<EmailTemplate> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/admin/system/email-templates", {
    page,
    limit,
    ...filters,
  });

  const result = useApiQuery<{ status: string; data: EmailTemplatesResponse }>(
    queryKeys.emailTemplates.list(query),
    url,
    {
      placeholderData: {
        status: "success",
        data: {
          templates: [],
          pagination: {
            page: 1,
            limit: 20,
            total: 0,
            totalPages: 0,
            hasNext: false,
            hasPrev: false,
          },
        },
      },
      ...options,
    }
  );

  return {
    data: result.data?.data?.templates || [],
    pagination: result.data?.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.emailTemplates.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateEmailTemplate(
  options?: UseMutationOptions<
    EmailTemplate,
    ApiError,
    CreateEmailTemplateRequest
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<EmailTemplate, CreateEmailTemplateRequest>(
    async (templateData) => {
      const response = await apiFetch<{ status: string; data: EmailTemplate }>(
        "/api/v1/admin/system/email-templates",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(templateData),
        }
      );
      // Return the unwrapped template data
      return response.data;
    },
    {
      onSuccess: (data) => {
        // Invalidate and refetch email templates list
        queryClient.invalidateQueries({
          queryKey: queryKeys.emailTemplates.all,
        });
        options?.onSuccess?.(data, {} as CreateEmailTemplateRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.emailTemplates.create
        );
        console.error("Failed to create email template:", message);
        options?.onError?.(error, {} as CreateEmailTemplateRequest, undefined);
      },
      ...options,
    }
  );
}

export function useUpdateEmailTemplate(
  templateId: string,
  options?: UseMutationOptions<
    EmailTemplate,
    ApiError,
    UpdateEmailTemplateRequest
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<EmailTemplate, UpdateEmailTemplateRequest>(
    async (templateData) => {
      const response = await apiFetch<{ status: string; data: EmailTemplate }>(
        `/api/v1/admin/system/email-templates/${templateId}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(templateData),
        }
      );
      // Return the unwrapped template data
      return response.data;
    },
    {
      onSuccess: (data) => {
        // Update specific template cache and invalidate lists
        queryClient.setQueryData(
          queryKeys.emailTemplates.detail(templateId),
          data
        );
        queryClient.invalidateQueries({
          queryKey: queryKeys.emailTemplates.lists(),
        });
        options?.onSuccess?.(data, {} as UpdateEmailTemplateRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.emailTemplates.update
        );
        console.error("Failed to update email template:", message);
        options?.onError?.(error, {} as UpdateEmailTemplateRequest, undefined);
      },
      ...options,
    }
  );
}

export function useDeleteEmailTemplate(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (templateId) =>
      apiFetch<void>(`/api/v1/admin/system/email-templates/${templateId}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, templateId) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.emailTemplates.detail(templateId),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.emailTemplates.lists(),
        });
        options?.onSuccess?.(undefined, templateId, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.emailTemplates.delete
        );
        console.error("Failed to delete email template:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== ROLES API HOOKS =====

export function useRoles(
  query: RoleQuery = {},
  options?: Omit<
    UseQueryOptions<RolesResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Role> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/roles", { page, limit, ...filters });

  const result = useApiQuery<RolesResponse>(queryKeys.roles.list(query), url, {
    placeholderData: {
      roles: [],
      pagination: {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0,
        hasNext: false,
        hasPrev: false,
      },
    },
    ...options,
  });

  return {
    data: result.data?.roles || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.roles.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateRole(
  options?: UseMutationOptions<
    Role,
    ApiError,
    { name: string; displayName?: string; description?: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<
    Role,
    { name: string; displayName?: string; description?: string }
  >(
    (roleData) =>
      apiFetch<Role>("/api/v1/roles", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(roleData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch roles list
        queryClient.invalidateQueries({ queryKey: queryKeys.roles.all });
        options?.onSuccess?.(
          data,
          {} as { name: string; displayName?: string; description?: string },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.roles.create);
        console.error("Failed to create role:", message);
        options?.onError?.(
          error,
          {} as { name: string; displayName?: string; description?: string },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useUpdateRole(
  roleName: string,
  options?: UseMutationOptions<
    Role,
    ApiError,
    { displayName?: string; description?: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<Role, { displayName?: string; description?: string }>(
    (roleData) =>
      apiFetch<Role>(`/api/v1/roles/${roleName}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(roleData),
      }),
    {
      onSuccess: (data) => {
        // Update specific role cache and invalidate lists
        queryClient.setQueryData(queryKeys.roles.detail(roleName), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
        options?.onSuccess?.(
          data,
          {} as { displayName?: string; description?: string },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.roles.update);
        console.error("Failed to update role:", message);
        options?.onError?.(
          error,
          {} as { displayName?: string; description?: string },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useDeleteRole(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (roleName) =>
      apiFetch<void>(`/api/v1/roles/${roleName}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, roleName) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.roles.detail(roleName),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
        options?.onSuccess?.(undefined, roleName, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, API_ERROR_MESSAGES.roles.delete);
        console.error("Failed to delete role:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== PERMISSIONS API HOOKS =====

export function usePermissions(
  query: PermissionQuery = {},
  options?: Omit<
    UseQueryOptions<PermissionsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Permission> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/permissions", { page, limit, ...filters });

  const result = useApiQuery<PermissionsResponse>(
    queryKeys.permissions.list(query),
    url,
    {
      placeholderData: {
        permissions: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data: result.data?.permissions || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.permissions.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreatePermission(
  options?: UseMutationOptions<
    Permission,
    ApiError,
    {
      name: string;
      displayName?: string;
      description?: string;
      category?: string;
    }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<
    Permission,
    {
      name: string;
      displayName?: string;
      description?: string;
      category?: string;
    }
  >(
    (permissionData) =>
      apiFetch<Permission>("/api/v1/permissions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(permissionData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch permissions list
        queryClient.invalidateQueries({ queryKey: queryKeys.permissions.all });
        options?.onSuccess?.(
          data,
          {} as {
            name: string;
            displayName?: string;
            description?: string;
            category?: string;
          },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.permissions.create
        );
        console.error("Failed to create permission:", message);
        options?.onError?.(
          error,
          {} as {
            name: string;
            displayName?: string;
            description?: string;
            category?: string;
          },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useUpdatePermission(
  permissionName: string,
  options?: UseMutationOptions<
    Permission,
    ApiError,
    { displayName?: string; description?: string; category?: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<
    Permission,
    { displayName?: string; description?: string; category?: string }
  >(
    (permissionData) =>
      apiFetch<Permission>(`/api/v1/permissions/${permissionName}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(permissionData),
      }),
    {
      onSuccess: (data) => {
        // Update specific permission cache and invalidate lists
        queryClient.setQueryData(
          queryKeys.permissions.detail(permissionName),
          data
        );
        queryClient.invalidateQueries({
          queryKey: queryKeys.permissions.lists(),
        });
        options?.onSuccess?.(
          data,
          {} as {
            displayName?: string;
            description?: string;
            category?: string;
          },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.permissions.update
        );
        console.error("Failed to update permission:", message);
        options?.onError?.(
          error,
          {} as {
            displayName?: string;
            description?: string;
            category?: string;
          },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useDeletePermission(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (permissionName) =>
      apiFetch<void>(`/api/v1/permissions/${permissionName}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, permissionName) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.permissions.detail(permissionName),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.permissions.lists(),
        });
        options?.onSuccess?.(undefined, permissionName, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.permissions.delete
        );
        console.error("Failed to delete permission:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== GROUPS API HOOKS =====

export function useGroups(
  query: GroupQuery = {},
  options?: Omit<
    UseQueryOptions<GroupsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Group> {
  const { page = 1, limit = 20, ...filters } = query;
  const url = buildQueryUrl("/api/v1/groups", { page, limit, ...filters });

  const result = useApiQuery<GroupsResponse>(
    queryKeys.groups.list(query),
    url,
    {
      placeholderData: {
        groups: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      },
      ...options,
    }
  );

  return {
    data: result.data?.groups || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, API_ERROR_MESSAGES.groups.fetch)
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

export function useCreateGroup(
  options?: UseMutationOptions<
    Group,
    ApiError,
    { name: string; displayName?: string; description?: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<
    Group,
    { name: string; displayName?: string; description?: string }
  >(
    (groupData) =>
      apiFetch<Group>("/api/v1/groups", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(groupData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate and refetch groups list
        queryClient.invalidateQueries({ queryKey: queryKeys.groups.all });
        options?.onSuccess?.(
          data,
          {} as { name: string; displayName?: string; description?: string },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.groups.create
        );
        console.error("Failed to create group:", message);
        options?.onError?.(
          error,
          {} as { name: string; displayName?: string; description?: string },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useUpdateGroup(
  groupUid: string,
  options?: UseMutationOptions<
    Group,
    ApiError,
    { displayName?: string; description?: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<Group, { displayName?: string; description?: string }>(
    (groupData) =>
      apiFetch<Group>(`/api/v1/groups/${groupUid}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(groupData),
      }),
    {
      onSuccess: (data) => {
        // Update specific group cache and invalidate lists
        queryClient.setQueryData(queryKeys.groups.detail(groupUid), data);
        queryClient.invalidateQueries({ queryKey: queryKeys.groups.lists() });
        options?.onSuccess?.(
          data,
          {} as { displayName?: string; description?: string },
          undefined
        );
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.groups.update
        );
        console.error("Failed to update group:", message);
        options?.onError?.(
          error,
          {} as { displayName?: string; description?: string },
          undefined
        );
      },
      ...options,
    }
  );
}

export function useDeleteGroup(
  options?: UseMutationOptions<void, ApiError, string>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, string>(
    (groupUid) =>
      apiFetch<void>(`/api/v1/groups/${groupUid}`, {
        method: "DELETE",
      }),
    {
      onSuccess: (_, groupUid) => {
        // Remove from cache and invalidate lists
        queryClient.removeQueries({
          queryKey: queryKeys.groups.detail(groupUid),
        });
        queryClient.invalidateQueries({ queryKey: queryKeys.groups.lists() });
        options?.onSuccess?.(undefined, groupUid, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          API_ERROR_MESSAGES.groups.delete
        );
        console.error("Failed to delete group:", message);
        options?.onError?.(error, "", undefined);
      },
      ...options,
    }
  );
}

// ===== ADDITIONAL HOOKS FOR CENTRALIZED API INTERACTIONS =====

/**
 * Hook to fetch a single study by UID with error handling
 */
export function useStudy(
  studyUid: string | undefined,
  options?: Omit<UseQueryOptions<Study, ApiError>, "queryKey" | "queryFn">
) {
  return useApiQuery<Study>(
    studyUid ? queryKeys.studies.detail(studyUid) : ["studies", "disabled"],
    studyUid ? `/api/v1/studies/${studyUid}` : "",
    {
      enabled: !!studyUid,
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: (failureCount, error: any) => {
        // Don't retry on authorization errors
        if (error?.status === 401 || error?.status === 403) {
          return false;
        }
        return failureCount < 2;
      },
      ...options,
    }
  );
}

/**
 * Tenant domain types
 */
export interface TenantDomain {
  id: number;
  domain: string;
  isVerified: boolean;
  isPrimary: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TenantDomainsResponse {
  domains: TenantDomain[];
}

export interface CreateTenantDomainRequest {
  domain: string;
  isPrimary: boolean;
}

export interface UpdateTenantDomainRequest {
  isVerified?: boolean;
  isPrimary?: boolean;
}

/**
 * Hook to fetch tenant domains
 */
export function useTenantDomains(
  tenantUid: string | undefined,
  options?: Omit<
    UseQueryOptions<TenantDomainsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
) {
  const result = useApiQuery<TenantDomainsResponse>(
    tenantUid
      ? queryKeys.tenants.domains(tenantUid)
      : ["tenants", "domains", "disabled"],
    tenantUid ? `/api/v1/tenants/${tenantUid}/domains` : "",
    {
      enabled: !!tenantUid,
      staleTime: 1000 * 30, // 30 seconds
      ...options,
    }
  );

  return {
    data: result.data?.domains || [],
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

/**
 * Hook to create a new tenant domain
 */
export function useCreateTenantDomain(
  tenantUid: string,
  options?: UseMutationOptions<
    TenantDomain,
    ApiError,
    CreateTenantDomainRequest
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<TenantDomain, CreateTenantDomainRequest>(
    (domainData) =>
      apiFetch<TenantDomain>(`/api/v1/tenants/${tenantUid}/domains`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(domainData),
      }),
    {
      onSuccess: (data) => {
        // Invalidate tenant domains cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });
        options?.onSuccess?.(data, {} as CreateTenantDomainRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, "Failed to add domain");
        console.error("Failed to create tenant domain:", message);
        options?.onError?.(error, {} as CreateTenantDomainRequest, undefined);
      },
      ...options,
    }
  );
}

/**
 * Hook to update a tenant domain
 */
export function useUpdateTenantDomain(
  tenantUid: string,
  domainId: number,
  options?: UseMutationOptions<
    TenantDomain,
    ApiError,
    UpdateTenantDomainRequest
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<TenantDomain, UpdateTenantDomainRequest>(
    (domainData) =>
      apiFetch<TenantDomain>(
        `/api/v1/tenants/${tenantUid}/domains/${domainId}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(domainData),
        }
      ),
    {
      onSuccess: (data) => {
        // Invalidate tenant domains cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });
        options?.onSuccess?.(data, {} as UpdateTenantDomainRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, "Failed to update domain");
        console.error("Failed to update tenant domain:", message);
        options?.onError?.(error, {} as UpdateTenantDomainRequest, undefined);
      },
      ...options,
    }
  );
}

/**
 * Hook to delete a tenant domain
 */
export function useDeleteTenantDomain(
  tenantUid: string,
  options?: UseMutationOptions<void, ApiError, number>
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, number>(
    (domainId) =>
      apiFetch<void>(`/api/v1/tenants/${tenantUid}/domains/${domainId}`, {
        method: "DELETE",
      }),
    {
      onSuccess: () => {
        // Invalidate tenant domains cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });
        options?.onSuccess?.(undefined, 0, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, "Failed to delete domain");
        console.error("Failed to delete tenant domain:", message);
        options?.onError?.(error, 0, undefined);
      },
      ...options,
    }
  );
}

// ===== SETTINGS HOOKS =====

/**
 * Hook to fetch settings with pagination and search
 */
export function useSettings(
  query: SettingQuery = {},
  options?: Omit<
    UseQueryOptions<SettingsResponse, ApiError>,
    "queryKey" | "queryFn"
  >
): PaginatedApiResult<Setting> {
  const url = buildQueryUrl("/api/v1/settings", query);

  const result = useApiQuery<SettingsResponse>(
    queryKeys.settings.list(query),
    url,
    {
      staleTime: 1000 * 30, // 30 seconds
      ...options,
    }
  );

  return {
    data: result.data?.settings || [],
    pagination: result.data?.pagination || {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    },
    loading: result.isLoading,
    error: result.error
      ? getErrorMessage(result.error, "Failed to fetch settings")
      : null,
    refetch: result.refetch,
    isStale: result.isStale,
    isFetching: result.isFetching,
  };
}

/**
 * Hook to fetch a single setting by tenant ID and key
 */
export function useSetting(
  tenantId: number | undefined,
  key: string | undefined,
  options?: Omit<UseQueryOptions<Setting, ApiError>, "queryKey" | "queryFn">
) {
  return useApiQuery<Setting>(
    tenantId && key
      ? queryKeys.settings.detail(tenantId, key)
      : ["settings", "disabled"],
    tenantId && key
      ? `/api/v1/settings/${tenantId}/${encodeURIComponent(key)}`
      : "",
    {
      enabled: !!(tenantId && key),
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: (failureCount, error: any) => {
        // Don't retry on authorization errors
        if (error?.status === 401 || error?.status === 403) {
          return false;
        }
        return failureCount < 2;
      },
      ...options,
    }
  );
}

/**
 * Hook to create a new setting
 */
export function useCreateSetting(
  options?: UseMutationOptions<Setting, ApiError, CreateSettingRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Setting, CreateSettingRequest>(
    (settingData) =>
      apiFetch<Setting>("/api/v1/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(settingData),
      }),
    {
      onSuccess: (data, variables) => {
        // Invalidate settings cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.settings.all,
        });

        // Update specific setting cache
        queryClient.setQueryData(
          queryKeys.settings.detail(variables.tenantId, variables.key),
          data
        );

        options?.onSuccess?.(data, variables, undefined);
      },
      onError: (error, variables) => {
        const message = getErrorMessage(error, "Failed to create setting");
        console.error("Failed to create setting:", message);
        options?.onError?.(error, variables, undefined);
      },
      ...options,
    }
  );
}

/**
 * Hook to update an existing setting
 */
export function useUpdateSetting(
  tenantId: number,
  key: string,
  options?: UseMutationOptions<Setting, ApiError, UpdateSettingRequest>
) {
  const queryClient = useQueryClient();

  return useApiMutation<Setting, UpdateSettingRequest>(
    (settingData) =>
      apiFetch<Setting>(
        `/api/v1/settings/${tenantId}/${encodeURIComponent(key)}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(settingData),
        }
      ),
    {
      onSuccess: (data) => {
        // Invalidate settings cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.settings.all,
        });

        // Update specific setting cache
        queryClient.setQueryData(
          queryKeys.settings.detail(tenantId, key),
          data
        );

        options?.onSuccess?.(data, {} as UpdateSettingRequest, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(error, "Failed to update setting");
        console.error("Failed to update setting:", message);
        options?.onError?.(error, {} as UpdateSettingRequest, undefined);
      },
      ...options,
    }
  );
}

/**
 * Hook to delete a setting
 */
export function useDeleteSetting(
  options?: UseMutationOptions<
    void,
    ApiError,
    { tenantId: number; key: string }
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<void, { tenantId: number; key: string }>(
    ({ tenantId, key }) =>
      apiFetch<void>(
        `/api/v1/settings/${tenantId}/${encodeURIComponent(key)}`,
        {
          method: "DELETE",
        }
      ),
    {
      onSuccess: (_, variables) => {
        // Invalidate settings cache
        queryClient.invalidateQueries({
          queryKey: queryKeys.settings.all,
        });

        // Remove specific setting from cache
        queryClient.removeQueries({
          queryKey: queryKeys.settings.detail(
            variables.tenantId,
            variables.key
          ),
        });

        options?.onSuccess?.(undefined, variables, undefined);
      },
      onError: (error, variables) => {
        const message = getErrorMessage(error, "Failed to delete setting");
        console.error("Failed to delete setting:", message);
        options?.onError?.(error, variables, undefined);
      },
      ...options,
    }
  );
}

/**
 * Hook to get settings count
 */
export function useSettingsCount(
  options?: Omit<
    UseQueryOptions<CountResponse, ApiError>,
    "queryKey" | "queryFn"
  >
) {
  return useApiQuery<CountResponse>(
    queryKeys.settings.count(),
    "/api/v1/settings/count",
    {
      staleTime: 1000 * 60, // 1 minute
      ...options,
    }
  );
}

// Study metadata field hooks
export function useStudyMetadataField(studyUid: string) {
  return useApiQuery<{ metadata: Record<string, any> }>(
    queryKeys.studies.metadata(studyUid),
    `/api/v1/studies/${studyUid}/metadata-field`,
    {
      enabled: !!studyUid,
      placeholderData: { metadata: {} },
    }
  );
}

export function useUpdateStudyMetadataField(
  studyUid: string,
  options?: UseMutationOptions<
    { message: string; metadata: Record<string, any> },
    ApiError,
    Record<string, any>
  >
) {
  const queryClient = useQueryClient();

  return useApiMutation<
    { message: string; metadata: Record<string, any> },
    Record<string, any>
  >(
    (metadata) =>
      apiFetch(`/api/v1/studies/${studyUid}/metadata-field`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(metadata),
      }),
    {
      onSuccess: (data) => {
        // Invalidate the metadata query for this study
        queryClient.invalidateQueries({
          queryKey: queryKeys.studies.metadata(studyUid),
        });
        // Also invalidate the study details since metadata is part of it
        queryClient.invalidateQueries({
          queryKey: queryKeys.studies.detail(studyUid),
        });
        options?.onSuccess?.(data, {} as Record<string, any>, undefined);
      },
      onError: (error) => {
        const message = getErrorMessage(
          error,
          "Failed to update study metadata"
        );
        console.error("Failed to update study metadata:", message);
        options?.onError?.(error, {} as Record<string, any>, undefined);
      },
      ...options,
    }
  );
}
