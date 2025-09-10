// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import {
  regionsService,
  type Region,
  type CreateRegionRequest,
  type UpdateRegionRequest,
} from "@/services/regionsService";
import type { RegionItem } from "@/features/viewer/components/RegionPanel";

// Query keys for consistent cache management
export const regionKeys = {
  all: ["regions"] as const,
  slide: (slideUid: string) => [...regionKeys.all, "slide", slideUid] as const,
  region: (regionId: string) =>
    [...regionKeys.all, "region", regionId] as const,
};

// Extended RegionItem with optimistic state
export interface OptimisticRegionItem extends RegionItem {
  isOptimistic?: boolean;
  isDeleting?: boolean;
  isSyncing?: boolean;
  localId?: string; // Original local ID for tracking
  error?: string;
  visible: boolean; // Override to make it required
  geoJson?: any; // GeoJSON feature for rendering
}

// Convert server Region to RegionItem
function serverRegionToRegionItem(region: Region): OptimisticRegionItem {
  return {
    id: region.regionId,
    name: region.regionName,
    visible: region.visible ?? true, // Ensure visible is always boolean
    isOptimistic: false,
    isSyncing: false,
  };
}

// Generate optimistic region item for immediate UI updates
function createOptimisticRegion(
  localId: string,
  name: string,
  visible: boolean = true,
  geoJson?: any
): OptimisticRegionItem {
  return {
    id: localId,
    name,
    visible,
    isOptimistic: true,
    isSyncing: true,
    localId,
    geoJson,
  };
}

/**
 * Hook for managing regions with optimistic updates and sync queue
 */
export function useRegionQueries(slideUid: string) {
  const queryClient = useQueryClient();

  // Query for fetching regions
  const regionsQuery = useQuery({
    queryKey: regionKeys.slide(slideUid),
    queryFn: async (): Promise<OptimisticRegionItem[]> => {
      if (!slideUid) return [];

      // First try to load from localStorage for immediate rendering
      const localRegions = loadLocalRegions();
      console.log(
        `useRegionQueries: Loaded ${localRegions.length} regions from localStorage`
      );

      // Then try to load from server for sync
      try {
        const serverRegions = await regionsService.getRegionsForSlide(slideUid);
        console.log(
          `useRegionQueries: Loaded ${serverRegions.length} regions from server`
        );

        // Merge server regions with local geometry data
        const mergedRegions: OptimisticRegionItem[] = [];

        // Add server regions (these are the authoritative source)
        serverRegions.forEach((serverRegion) => {
          const localRegion = localRegions.find(
            (local) =>
              local.regionId === serverRegion.regionId ||
              local.id === serverRegion.regionId
          );

          mergedRegions.push({
            id: serverRegion.regionId,
            name: serverRegion.regionName,
            visible: serverRegion.visible,
            isOptimistic: false,
            isSyncing: false,
            geoJson: localRegion?.geoJson || {
              type: "Feature",
              id: serverRegion.regionId,
              geometry: serverRegion.geometry,
              properties: {
                name: serverRegion.regionName,
                visible: serverRegion.visible,
                kind: "box",
              },
            },
          });
        });

        // Add any local-only regions (not yet synced)
        localRegions.forEach((localRegion) => {
          if (
            localRegion.isLocal &&
            !mergedRegions.find((m) => m.id === localRegion.id)
          ) {
            mergedRegions.push({
              id: localRegion.id,
              name: localRegion.name,
              visible: localRegion.visible,
              isOptimistic: true,
              isSyncing: true,
              localId: localRegion.id,
              geoJson: localRegion.geoJson,
            });
          }
        });

        return mergedRegions;
      } catch (error) {
        console.error("Error loading server regions, using local only:", error);
        // Fall back to local regions only
        return localRegions.map((local) => ({
          id: local.id,
          name: local.name,
          visible: local.visible,
          isOptimistic: local.isLocal,
          isSyncing: local.isLocal,
          localId: local.isLocal ? local.id : undefined,
          geoJson: local.geoJson,
        }));
      }
    },
    enabled: !!slideUid,
    staleTime: 30000, // Consider data fresh for 30 seconds
    gcTime: 5 * 60 * 1000, // Keep in cache for 5 minutes
  });

  // Helper to load local regions from localStorage
  const loadLocalRegions = useCallback((): Array<{
    id: string;
    name: string;
    visible: boolean;
    geoJson: any;
    isLocal: boolean;
    regionId?: string;
  }> => {
    if (!slideUid) return [];

    try {
      const key = `slideRegions:${slideUid}`;
      const raw = localStorage.getItem(key);
      if (!raw) return [];

      const fc = JSON.parse(raw);
      if (!fc.features || !Array.isArray(fc.features)) return [];

      return fc.features.map((feature: any) => ({
        id: String(feature.id || `${Date.now()}-${Math.random()}`),
        name: feature.properties?.name || `Region ${feature.id}`,
        visible: feature.properties?.visible !== false,
        geoJson: feature,
        isLocal: !feature.properties?.synced,
        regionId:
          feature.properties?.regionId ||
          (feature.properties?.synced ? String(feature.id) : undefined),
      }));
    } catch (error) {
      console.error("Error loading local regions:", error);
      return [];
    }
  }, [slideUid]);

  // Helper to get current regions from cache
  const getCurrentRegions = useCallback((): OptimisticRegionItem[] => {
    return queryClient.getQueryData(regionKeys.slide(slideUid)) || [];
  }, [queryClient, slideUid]);

  // Helper to update regions cache
  const updateRegionsCache = useCallback(
    (updater: (old: OptimisticRegionItem[]) => OptimisticRegionItem[]) => {
      queryClient.setQueryData(regionKeys.slide(slideUid), updater);
    },
    [queryClient, slideUid]
  );

  // Helper function to validate region name uniqueness
  const validateRegionName = useCallback(
    (name: string, excludeId?: string): string | null => {
      if (!name.trim()) {
        return "Region name cannot be empty";
      }

      const trimmedName = name.trim();

      // Check for duplicate names (case-insensitive)
      const currentRegions = getCurrentRegions();
      const existingNames = currentRegions
        .filter((r) => r.id !== excludeId)
        .map((r) => r.name.toLowerCase());

      if (existingNames.includes(trimmedName.toLowerCase())) {
        return "A region with this name already exists";
      }

      // Check for reserved names
      const reservedNames = ["region", "area", "roi", "annotation"];
      if (reservedNames.includes(trimmedName.toLowerCase())) {
        return "This name is reserved. Please choose a different name";
      }

      // Check length
      if (trimmedName.length > 50) {
        return "Region name cannot be longer than 50 characters";
      }

      return null; // Valid
    },
    [getCurrentRegions]
  );

  // Create region mutation with optimistic updates
  const createRegionMutation = useMutation({
    mutationFn: async (request: CreateRegionRequest & { localId: string }) => {
      const { localId, ...createRequest } = request;

      // Validate name before sending to server
      const nameError = validateRegionName(createRequest.regionName);
      if (nameError) {
        throw new Error(nameError);
      }

      const createdRegion = await regionsService.createRegion(
        slideUid,
        createRequest
      );
      return { createdRegion, localId };
    },
    onMutate: async (request) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: regionKeys.slide(slideUid) });

      // Snapshot the previous value
      const previousRegions = getCurrentRegions();

      // Create geoJson from the geometry data
      const geoJson = {
        type: "Feature",
        id: request.localId,
        geometry: request.geometry,
        properties: {
          name: request.regionName,
          visible: request.visible,
          kind: "box",
        },
      };

      // Optimistically update the cache
      const optimisticRegion = createOptimisticRegion(
        request.localId,
        request.regionName,
        request.visible,
        geoJson
      );

      updateRegionsCache((old) => [...old, optimisticRegion]);

      // Return context for rollback
      return { previousRegions, localId: request.localId };
    },
    onSuccess: (data, variables, context) => {
      // Replace optimistic region with real server data, keeping the geoJson
      updateRegionsCache((old) =>
        old.map((region) =>
          region.localId === context?.localId
            ? {
                ...serverRegionToRegionItem(data.createdRegion),
                geoJson: region.geoJson, // Preserve the geoJson from the optimistic update
              }
            : region
        )
      );

      // Trigger map feature ID update
      window.dispatchEvent(
        new CustomEvent("regionIdUpdated", {
          detail: {
            localId: context?.localId,
            serverId: data.createdRegion.regionId,
            region: data.createdRegion,
          },
        })
      );
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousRegions) {
        queryClient.setQueryData(
          regionKeys.slide(slideUid),
          context.previousRegions
        );
      }

      // Mark region as failed
      updateRegionsCache((old) =>
        old.map((region) =>
          region.localId === context?.localId
            ? { ...region, error: error.message, isSyncing: false }
            : region
        )
      );
    },
    onSettled: () => {
      // Always refetch after mutation settles
      queryClient.invalidateQueries({ queryKey: regionKeys.slide(slideUid) });
    },
  });

  // Update region mutation
  const updateRegionMutation = useMutation({
    mutationFn: async ({
      regionId,
      updates,
    }: {
      regionId: string;
      updates: UpdateRegionRequest;
    }) => {
      // Validate name if it's being updated
      if (updates.regionName) {
        const nameError = validateRegionName(updates.regionName, regionId);
        if (nameError) {
          throw new Error(nameError);
        }
      }

      return await regionsService.updateRegion(regionId, updates);
    },
    onMutate: async ({ regionId, updates }) => {
      await queryClient.cancelQueries({ queryKey: regionKeys.slide(slideUid) });

      const previousRegions = getCurrentRegions();

      // Optimistically update
      updateRegionsCache((old) =>
        old.map((region) =>
          region.id === regionId
            ? {
                ...region,
                ...updates,
                name: updates.regionName || region.name,
                visible: updates.visible ?? region.visible,
                isSyncing: true,
              }
            : region
        )
      );

      return { previousRegions, regionId };
    },
    onSuccess: (updatedRegion, { regionId }) => {
      // Update with server response
      updateRegionsCache((old) =>
        old.map((region) =>
          region.id === regionId
            ? { ...serverRegionToRegionItem(updatedRegion), isSyncing: false }
            : region
        )
      );
    },
    onError: (error, { regionId }, context) => {
      // Rollback on error
      if (context?.previousRegions) {
        queryClient.setQueryData(
          regionKeys.slide(slideUid),
          context.previousRegions
        );
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: regionKeys.slide(slideUid) });
    },
  });

  // Delete region mutation with optimistic updates
  const deleteRegionMutation = useMutation({
    mutationFn: async (regionId: string) => {
      await regionsService.deleteRegion(slideUid, regionId);
      return regionId;
    },
    onMutate: async (regionId) => {
      await queryClient.cancelQueries({ queryKey: regionKeys.slide(slideUid) });

      const previousRegions = getCurrentRegions();

      // Optimistically mark as deleting
      updateRegionsCache((old) =>
        old.map((region) =>
          region.id === regionId
            ? { ...region, isDeleting: true, isSyncing: true }
            : region
        )
      );

      return { previousRegions, regionId };
    },
    onSuccess: (regionId) => {
      // Remove from cache after successful deletion
      updateRegionsCache((old) =>
        old.filter((region) => region.id !== regionId)
      );

      // Trigger map feature removal
      window.dispatchEvent(
        new CustomEvent("regionDeleted", {
          detail: { regionId },
        })
      );
    },
    onError: (error, regionId, context) => {
      // Rollback on error
      if (context?.previousRegions) {
        queryClient.setQueryData(
          regionKeys.slide(slideUid),
          context.previousRegions
        );
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: regionKeys.slide(slideUid) });
    },
  });

  // Batch operations for bulk actions
  const batchOperations = useCallback(
    async (operations: Array<() => Promise<any>>) => {
      // Execute all operations
      const results = await Promise.allSettled(operations.map((op) => op()));

      // Check for failures
      const failures = results.filter((result) => result.status === "rejected");
      if (failures.length > 0) {
        console.warn(`${failures.length} operations failed:`, failures);
      }

      return results;
    },
    []
  );

  // Retry failed mutations
  const retryFailedMutations = useCallback(() => {
    const regions = getCurrentRegions();
    const failedRegions = regions.filter((region) => region.error);

    failedRegions.forEach((region) => {
      if (region.isOptimistic) {
        // Retry creation
        createRegionMutation.mutate({
          localId: region.localId || region.id,
          regionName: region.name,
          visible: region.visible,
          regionType: "roi", // Default type
          geometry: { type: "Polygon", coordinates: [] }, // Would need actual geometry
          coordinateSystem: "pixel",
        });
      }
    });
  }, [getCurrentRegions, createRegionMutation]);

  return {
    // Query data
    regions: regionsQuery.data || [],
    isLoading: regionsQuery.isLoading,
    isError: regionsQuery.isError,
    error: regionsQuery.error,

    // Mutations
    createRegion: createRegionMutation.mutate,
    updateRegion: updateRegionMutation.mutate,
    deleteRegion: deleteRegionMutation.mutate,

    // Mutation states
    isCreating: createRegionMutation.isPending,
    isUpdating: updateRegionMutation.isPending,
    isDeleting: deleteRegionMutation.isPending,

    // Utility functions
    getCurrentRegions,
    validateRegionName,
    batchOperations,
    retryFailedMutations,

    // Manual cache updates (for map integration)
    updateRegionsCache,
  };
}
