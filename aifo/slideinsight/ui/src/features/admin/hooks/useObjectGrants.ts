// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../../../utils/fetchUtils";
import { toast } from "sonner";

// Object grant types
export interface ObjectGrant {
  id: number;
  grantee_type: "user" | "group" | "role";
  grantee_id: number;
  grantee_name?: string;
  permission: string;
  resource_type: "study" | "case" | "slide";
  resource_id: number;
  created_at: string;
  updated_at: string;
  grantee_info?: {
    name: string;
    email?: string;
    first_name?: string;
    last_name?: string;
  };
}

interface CreateObjectGrantRequest {
  grantee_type: "user" | "group" | "role";
  grantee_id?: number;
  grantee_uid?: string;
  permission: string;
  resource_type: "study" | "case" | "slide";
  resource_id?: number;
  resource_uid?: string;
}

interface BulkCreateObjectGrantsRequest {
  grantee_type: "user" | "group" | "role";
  grantee_id?: number;
  grantee_uid?: string;
  permissions: string[];
  resource_type: "study" | "case" | "slide";
  resource_id?: number;
  resource_uid?: string;
}

interface DeleteObjectGrantRequest {
  grantee_type: "user" | "group" | "role";
  grantee_id: number;
  permission: string;
}

interface BulkDeleteObjectGrantsRequest {
  resourceType: string;
  resourceId: string | number;
  grants: DeleteObjectGrantRequest[];
}

// Query keys for object grants
const objectGrantKeys = {
  all: ["object-grants"] as const,
  resource: (resourceType: string, resourceId: string | number) =>
    [...objectGrantKeys.all, "resource", resourceType, resourceId] as const,
};

// Hook to get object grants for a specific resource
export function useObjectGrants(
  resourceType: string,
  resourceId: string | number
) {
  return useQuery({
    queryKey: objectGrantKeys.resource(resourceType, resourceId),
    queryFn: async (): Promise<ObjectGrant[]> => {
      const response = await apiFetch<ObjectGrant[]>(
        `/api/v1/object-grants/${resourceType}/${resourceId}`
      );
      return response || [];
    },
    enabled: !!resourceType && !!resourceId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to create an object grant
export function useCreateObjectGrant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateObjectGrantRequest): Promise<void> => {
      await apiFetch<any>("/api/v1/object-grants", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
    },
    onSuccess: (_, variables) => {
      // Invalidate and refetch grants for this resource
      queryClient.invalidateQueries({
        queryKey: objectGrantKeys.resource(
          variables.resource_type,
          variables.resource_uid || variables.resource_id!
        ),
      });

      toast.success("Permission granted!", {
        description: `${variables.permission} permission granted successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to create object grant:", error);
      toast.error("Failed to grant permission", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to create multiple object grants for the same user/resource
export function useBulkCreateObjectGrants() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: BulkCreateObjectGrantsRequest): Promise<void> => {
      // Create multiple grants by making parallel requests
      const promises = data.permissions.map((permission) =>
        apiFetch<any>("/api/v1/object-grants", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            grantee_type: data.grantee_type,
            grantee_id: data.grantee_id,
            grantee_uid: data.grantee_uid,
            permission,
            resource_type: data.resource_type,
            resource_id: data.resource_id,
            resource_uid: data.resource_uid,
          }),
        })
      );

      await Promise.all(promises);
    },
    onSuccess: (_, variables) => {
      // Invalidate and refetch grants for this resource
      queryClient.invalidateQueries({
        queryKey: objectGrantKeys.resource(
          variables.resource_type,
          variables.resource_uid || variables.resource_id!
        ),
      });

      const count = variables.permissions.length;
      toast.success("Permissions granted!", {
        description: `${count} permission${
          count > 1 ? "s" : ""
        } granted successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to create object grants:", error);
      toast.error("Failed to grant permissions", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to delete an object grant
export function useDeleteObjectGrant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      resourceType,
      resourceId,
      data,
    }: {
      resourceType: string;
      resourceId: string | number;
      data: DeleteObjectGrantRequest;
    }): Promise<void> => {
      await apiFetch<any>(
        `/api/v1/object-grants/${resourceType}/${resourceId}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        }
      );
    },
    onSuccess: (_, variables) => {
      // Invalidate and refetch grants for this resource
      queryClient.invalidateQueries({
        queryKey: objectGrantKeys.resource(
          variables.resourceType,
          variables.resourceId
        ),
      });

      toast.success("Permission revoked!", {
        description: `${variables.data.permission} permission revoked successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to delete object grant:", error);
      toast.error("Failed to revoke permission", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}

// Hook to delete multiple object grants
export function useBulkDeleteObjectGrants() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: BulkDeleteObjectGrantsRequest): Promise<void> => {
      // Delete multiple grants by making parallel requests
      const promises = data.grants.map((grant) =>
        apiFetch<any>(
          `/api/v1/object-grants/${data.resourceType}/${data.resourceId}`,
          {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify(grant),
          }
        )
      );

      await Promise.all(promises);
    },
    onSuccess: (_, variables) => {
      // Invalidate and refetch grants for this resource
      queryClient.invalidateQueries({
        queryKey: objectGrantKeys.resource(
          variables.resourceType,
          variables.resourceId
        ),
      });

      const count = variables.grants.length;
      toast.success("Permissions revoked!", {
        description: `${count} permission${
          count > 1 ? "s" : ""
        } revoked successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to delete object grants:", error);
      toast.error("Failed to revoke permissions", {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    },
  });
}
