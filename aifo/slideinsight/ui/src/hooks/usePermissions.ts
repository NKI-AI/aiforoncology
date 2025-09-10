// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/utils/fetchUtils";

export interface PermissionGrantResource {
  resourceType: string; // "role", "study", "case", "slide"
  resourceUid?: string;
  resourceName?: string;
}

export interface PermissionCheck {
  checkType: string; // "role_based", "direct_object", "inherited_object"
  resourceType?: string;
  resourceUid?: string;
  resourceName?: string;
  result: boolean; // Whether this specific check passed
  description: string; // Human-readable description of what was checked
  grantingEntity?: PermissionGrantResource; // What granted this permission
}

export interface PermissionExplanation {
  userUid: string;
  permission: string;
  resourceType: string;
  resourceUid: string;
  hasAccess: boolean;
  grantType?: string; // "role_based_grant", "direct_object_grant", "inherited_grant", "access_denied"
  inheritancePath?: string; // e.g., "slide->case->study" for inherited access
  grantingResource?: PermissionGrantResource; // The resource that actually grants the permission
  checksPerformed: PermissionCheck[]; // All permission checks that were performed
  message: string; // Human-readable explanation
}

/**
 * Hook to fetch permission explanation for a specific resource
 */
export function usePermissionExplanation(
  resourceType: string,
  resourceUid: string | null,
  permission: string = "studies.view",
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options;

  return useQuery<PermissionExplanation>({
    queryKey: ["permission-explanation", resourceType, resourceUid, permission],
    queryFn: async () => {
      if (!resourceUid) {
        throw new Error("Resource UID is required");
      }

      const resourcePlural =
        resourceType === "study" ? "studies" : `${resourceType}s`;
      const url = `/api/v1/${resourcePlural}/${resourceUid}/access-explanation?permission=${encodeURIComponent(
        permission
      )}`;
      return apiFetch<PermissionExplanation>(url);
    },
    enabled: enabled && !!resourceUid,
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on 401/403 errors as they're likely permission issues
      if (error?.status === 401 || error?.status === 403) {
        return false;
      }
      return failureCount < 2;
    },
  });
}

/**
 * Hook to check multiple permissions at once
 */
export function useMultiplePermissionExplanations(
  checks: Array<{
    resourceType: string;
    resourceUid: string;
    permission?: string;
  }>,
  options: { enabled?: boolean } = {}
) {
  const { enabled = true } = options;

  const results = checks.map((check) =>
    usePermissionExplanation(
      check.resourceType,
      check.resourceUid,
      check.permission,
      { enabled }
    )
  );

  return {
    results,
    isLoading: results.some((result) => result.isLoading),
    hasErrors: results.some((result) => result.error),
    allLoaded: results.every((result) => !result.isLoading),
  };
}
