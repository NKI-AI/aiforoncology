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
import { Tenant } from "../../../api/models";

interface TenantsApiResponse {
  tenants: Tenant[];
  pagination: PaginationInfo;
}

interface UseTenantsOptions extends BasePaginatedOptions {
  name?: string; // Filter by name
  description?: string; // Filter by description
}

export function useTenants(options: UseTenantsOptions = {}) {
  const config = useMemo(
    () => ({
      endpoint: "/api/v1/tenants",
      queryBuilder: (opts: UseTenantsOptions) => {
        const params: Record<string, string> = {};
        if (opts.q) params.q = opts.q;
        if (opts.name) params.name = opts.name;
        if (opts.description) params.description = opts.description;
        if (opts.sort) params.sort = opts.sort;
        if (opts.dir) params.dir = opts.dir;
        return params;
      },
      dataExtractor: (response: TenantsApiResponse) => ({
        items: response.tenants,
        pagination: response.pagination,
      }),
      errorMessage: "Failed to load tenants. Please try again later.",
    }),
    []
  );

  const result = usePaginatedApi<Tenant, TenantsApiResponse, UseTenantsOptions>(
    options,
    config
  );

  return {
    tenants: result.data,
    pagination: result.pagination,
    loading: result.loading,
    error: result.error,
    refetch: result.refetch,
  };
}
