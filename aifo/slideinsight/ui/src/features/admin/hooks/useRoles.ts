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
  Role,
  CreateRoleRequest,
  RolePermissionAssignment,
  UserRoleAssignment,
} from "../../../types/roles";
import { Permission } from "../../../types/permissions";
import { toast } from "sonner";

// Hook to get all roles
export function useRoles() {
  return useQuery({
    queryKey: queryKeys.roles.list(),
    queryFn: async (): Promise<Role[]> => {
      const response = await apiFetch<Role[]>("/api/v1/roles");
      return response || [];
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to create a role
export function useCreateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRoleRequest): Promise<Role> => {
      return await apiFetch<Role>("/api/v1/roles", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
    },
    onSuccess: (newRole) => {
      // Invalidate and refetch roles list
      queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
      toast.success("Role created!", {
        description: `${newRole.name} has been created successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to create role:", error);
      toast.error("Failed to create role", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to delete a role
export function useDeleteRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (roleName: string): Promise<void> => {
      await apiFetch(`/api/v1/roles/${roleName}`, {
        method: "DELETE",
      });
    },
    onSuccess: (_, roleName) => {
      // Invalidate and refetch roles list
      queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
      toast.success("Role deleted!", {
        description: `${roleName} has been permanently deleted.`,
      });
    },
    onError: (error) => {
      console.error("Failed to delete role:", error);
      toast.error("Failed to delete role", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to get role permissions
export function useRolePermissions(roleName: string) {
  return useQuery({
    queryKey: queryKeys.roles.permissions(roleName),
    queryFn: async (): Promise<Permission[]> => {
      const response = await apiFetch<Permission[]>(
        `/api/v1/roles/${roleName}/permissions`
      );
      return response || [];
    },
    enabled: !!roleName,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to assign permissions to a role
export function useAssignPermissionsToRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      roleName,
      permissionIDs,
    }: {
      roleName: string;
      permissionIDs: number[];
    }): Promise<any> => {
      return await apiFetch<any>(`/api/v1/roles/${roleName}/permissions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ permission_ids: permissionIDs }),
      });
    },
    onSuccess: (response, { roleName }) => {
      // Invalidate role permissions
      queryClient.invalidateQueries({
        queryKey: queryKeys.roles.permissions(roleName),
      });

      const assignedCount = response.count || 0;
      const errorCount = response.error_count || 0;

      if (errorCount > 0) {
        toast.warning("Permission assignment completed with some errors", {
          description: `${assignedCount} permissions assigned, ${errorCount} failed.`,
        });
      } else {
        toast.success("Permissions assigned!", {
          description: `${assignedCount} permissions assigned to ${roleName}.`,
        });
      }
    },
    onError: (error) => {
      console.error("Failed to assign permissions to role:", error);
      toast.error("Failed to assign permissions", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to remove a permission from a role
export function useRemovePermissionFromRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      roleName,
      permissionID,
    }: {
      roleName: string;
      permissionID: number;
    }): Promise<void> => {
      await apiFetch(`/api/v1/roles/${roleName}/permissions/${permissionID}`, {
        method: "DELETE",
      });
    },
    onSuccess: (_, { roleName, permissionID }) => {
      // Invalidate role permissions
      queryClient.invalidateQueries({
        queryKey: queryKeys.roles.permissions(roleName),
      });
      toast.success("Permission removed!", {
        description: `Permission removed from ${roleName}.`,
      });
    },
    onError: (error) => {
      console.error("Failed to remove permission from role:", error);
      toast.error("Failed to remove permission", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to get role users
export function useRoleUsers(roleName: string) {
  return useQuery({
    queryKey: queryKeys.roles.users(roleName),
    queryFn: async (): Promise<string[]> => {
      const response = await apiFetch<string[]>(
        `/api/v1/roles/${roleName}/users`
      );
      return response || [];
    },
    enabled: !!roleName,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to assign users to a role
export function useAssignUsersToRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      roleName,
      userUIDs,
    }: {
      roleName: string;
      userUIDs: string[];
    }): Promise<any> => {
      return await apiFetch<any>(`/api/v1/roles/${roleName}/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ user_uids: userUIDs }),
      });
    },
    onSuccess: (response, { roleName }) => {
      // Invalidate role users
      queryClient.invalidateQueries({
        queryKey: queryKeys.roles.users(roleName),
      });

      const assignedCount = response.count || 0;
      const errorCount = response.error_count || 0;

      if (errorCount > 0) {
        toast.warning("User assignment completed with some errors", {
          description: `${assignedCount} users assigned, ${errorCount} failed.`,
        });
      } else {
        toast.success("Users assigned!", {
          description: `${assignedCount} users assigned to ${roleName}.`,
        });
      }
    },
    onError: (error) => {
      console.error("Failed to assign users to role:", error);
      toast.error("Failed to assign users", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to remove a user from a role
export function useRemoveUserFromRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      roleName,
      userUID,
    }: {
      roleName: string;
      userUID: string;
    }): Promise<void> => {
      await apiFetch(`/api/v1/roles/${roleName}/users/${userUID}`, {
        method: "DELETE",
      });
    },
    onSuccess: (_, { roleName, userUID }) => {
      // Invalidate role users
      queryClient.invalidateQueries({
        queryKey: queryKeys.roles.users(roleName),
      });
      toast.success("User removed!", {
        description: `User removed from ${roleName}.`,
      });
    },
    onError: (error) => {
      console.error("Failed to remove user from role:", error);
      toast.error("Failed to remove user", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}
