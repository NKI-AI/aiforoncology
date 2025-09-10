// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../../utils/fetchUtils";
import { queryKeys } from "../../../api/queryKeys";
import {
  Permission,
  CreatePermissionRequest,
  PermissionsResponse,
} from "../../../types/permissions";
import { toast } from "sonner";

// Hook to get all permissions
export function usePermissions() {
  return useQuery({
    queryKey: queryKeys.permissions.list(),
    queryFn: async (): Promise<Permission[]> => {
      // Backend returns { permissions: Permission[], pagination: {...} }
      const response = await apiFetch<{ permissions: Permission[] }>(
        "/api/v1/permissions?limit=1000"
      );
      return response?.permissions || [];
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to create a permission
export function useCreatePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePermissionRequest): Promise<Permission> => {
      return await apiFetch<Permission>("/api/v1/permissions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
    },
    onSuccess: (newPermission) => {
      // Invalidate and refetch permissions list
      queryClient.invalidateQueries({
        queryKey: queryKeys.permissions.lists(),
      });
      toast.success("Permission created!", {
        description: `${newPermission.name} has been created successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to create permission:", error);
      toast.error("Failed to create permission", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to create multiple permissions (bulk)
export function useCreatePermissionsBulk() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePermissionRequest[]): Promise<any> => {
      return await apiFetch<any>("/api/v1/permissions/bulk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
    },
    onSuccess: (response) => {
      // Invalidate and refetch permissions list
      queryClient.invalidateQueries({
        queryKey: queryKeys.permissions.lists(),
      });

      const createdCount = response.count || 0;
      const errorCount = response.error_count || 0;

      if (errorCount > 0) {
        toast.warning("Bulk permission creation completed with some errors", {
          description: `${createdCount} permissions created, ${errorCount} failed.`,
        });
      } else {
        toast.success("Bulk permissions created!", {
          description: `${createdCount} permissions created successfully.`,
        });
      }
    },
    onError: (error) => {
      console.error("Failed to create permissions:", error);
      toast.error("Failed to create permissions", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}
