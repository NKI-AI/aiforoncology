// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import { queryKeys } from "../../../api/queryKeys";
import {
  usePaginatedApi,
  type BasePaginatedOptions,
  type PaginationInfo,
} from "@/hooks/usePaginatedApi";
import { User } from "../../../api/models";

interface UsersApiResponse {
  users: User[];
  pagination: PaginationInfo;
}

interface UseUsersOptions extends BasePaginatedOptions {
  email?: string; // Filter by email
  first_name?: string; // Filter by first name
  last_name?: string; // Filter by last name
  status?: string; // Filter by status (active/inactive)
  mustResetPassword?: string; // Filter by must reset password (yes/no)
}

export function useUsers(options: UseUsersOptions = {}) {
  const config = useMemo(
    () => ({
      endpoint: "/api/v1/users",
      queryBuilder: (opts: UseUsersOptions) => {
        const params: Record<string, string> = {};
        if (opts.q) params.q = opts.q;
        if (opts.email) params.email = opts.email;
        if (opts.first_name) params.first_name = opts.first_name;
        if (opts.last_name) params.last_name = opts.last_name;

        // Map frontend 'status' to backend 'is_active'
        if (opts.status) {
          if (opts.status === "active") {
            params.is_active = "true";
          } else if (opts.status === "inactive") {
            params.is_active = "false";
          }
        }

        // Map frontend 'mustResetPassword' to backend 'must_reset_password'
        if (opts.mustResetPassword) {
          if (opts.mustResetPassword === "yes") {
            params.must_reset_password = "true";
          } else if (opts.mustResetPassword === "no") {
            params.must_reset_password = "false";
          }
        }

        if (opts.sort) params.sort = opts.sort;
        if (opts.dir) params.dir = opts.dir;
        return params;
      },
      queryKeyFactory: (opts: UseUsersOptions) => {
        // Use centralized query keys with normalized options
        const {
          page = 1,
          limit = 20,
          mustResetPassword,
          status,
          ...filters
        } = opts;

        // Map string values to proper types for UserQuery
        const queryParams: any = { page, limit, ...filters };

        if (status) {
          queryParams.isActive = status === "active";
        }

        if (mustResetPassword) {
          queryParams.mustResetPassword = mustResetPassword === "yes";
        }

        return queryKeys.users.list(queryParams);
      },
      dataExtractor: (response: UsersApiResponse) => ({
        items: response.users,
        pagination: response.pagination,
      }),
      errorMessage: "Failed to load users. Please try again later.",
    }),
    []
  );

  const result = usePaginatedApi<User, UsersApiResponse, UseUsersOptions>(
    options,
    config
  );

  return {
    users: result.data,
    pagination: result.pagination,
    loading: result.loading,
    error: result.error,
    refetch: result.refetch,
  };
}
