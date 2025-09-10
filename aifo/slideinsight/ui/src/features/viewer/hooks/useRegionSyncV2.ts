// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/api/queryKeys";
import {
  regionsService,
  type Region,
  type CreateRegionRequest,
  type UpdateRegionRequest,
} from "@/services/regionsService";
import type { RegionItem } from "@/features/viewer/components/RegionPanel";

export interface OptimisticRegionItem extends RegionItem {
  isCreating?: boolean;
  isDeleting?: boolean;
  isUpdating?: boolean;
  error?: string;
}

// Helper to convert Region API type to local RegionItem
const regionToRegionItem = (region: Region): RegionItem => ({
  id: region.regionId,
  name: region.regionName,
  visible: region.visible,
});

// Helper to convert geometry to GeoJSON feature format
const regionToGeoJson = (region: Region) => ({
  type: "Feature",
  id: region.regionId,
  geometry: region.geometry,
  properties: {
    regionId: region.regionId,
    name: region.regionName,
    visible: region.visible,
    regionType: region.regionType,
    coordinateSystem: region.coordinateSystem,
    synced: true,
  },
});

// Helper function to generate a hash from geometry for deduplication
function generateGeometryHash(geoJson: any): string {
  if (!geoJson?.geometry) return "";
  try {
    return btoa(JSON.stringify(geoJson.geometry)).slice(0, 16);
  } catch {
    return "";
  }
}

// LocalStorage utilities
const getLocalStorageKey = (slideUid: string) => `slideRegions:${slideUid}`;

// Helper to find server ID for a temp ID in localStorage
const findServerIdInLocalStorage = async (
  slideUid: string,
  tempId: string
): Promise<string | null> => {
  try {
    const key = getLocalStorageKey(slideUid);
    const raw = localStorage.getItem(key);
    if (!raw) return null;

    const fc = JSON.parse(raw);
    if (!fc.features || !Array.isArray(fc.features)) return null;

    // Look for a feature that has the temp ID but also has a server regionId
    const feature = fc.features.find((f: any) => {
      return (
        (String(f.id) === tempId ||
          String(f.properties?.originalId) === tempId) &&
        f.properties?.regionId &&
        !f.properties.regionId.startsWith("temp-")
      );
    });

    return feature?.properties?.regionId || null;
  } catch (error) {
    console.error("Error finding server ID in localStorage:", error);
    return null;
  }
};

const loadLocalRegions = (slideUid: string): Region[] => {
  if (!slideUid) return [];

  try {
    const key = getLocalStorageKey(slideUid);
    const raw = localStorage.getItem(key);
    if (!raw) return [];

    const fc = JSON.parse(raw);
    if (!fc.features || !Array.isArray(fc.features)) return [];

    return fc.features.map(
      (feature: any): Region => ({
        regionId: String(feature.id || `${Date.now()}-${Math.random()}`),
        regionName: feature.properties?.name || `Region ${feature.id}`,
        slideUid,
        regionType: feature.properties?.regionType || "roi",
        geometry: feature.geometry,
        coordinateSystem: feature.properties?.coordinateSystem || "pixel",
        mutable: true,
        visible: feature.properties?.visible !== false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
    );
  } catch (error) {
    console.error("Error loading local regions:", error);
    return [];
  }
};

const saveLocalRegions = (slideUid: string, regions: Region[]) => {
  if (!slideUid) return;

  try {
    const key = getLocalStorageKey(slideUid);
    const featureCollection = {
      type: "FeatureCollection",
      features: regions.map((region) => regionToGeoJson(region)),
    };
    localStorage.setItem(key, JSON.stringify(featureCollection));
  } catch (error) {
    console.error("Error saving local regions:", error);
  }
};

export function useRegionSyncV2(slideUid: string) {
  const queryClient = useQueryClient();
  const regionQueryKey = queryKeys.regions.list(slideUid);

  // Track pending operations to handle create-then-delete scenarios
  const pendingOperationsRef = useRef<
    Map<string, { type: "create" | "delete"; serverRegionId?: string }>
  >(new Map());

  // Query to load regions with localStorage fallback
  const {
    data: regions = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: regionQueryKey,
    queryFn: async (): Promise<Region[]> => {
      try {
        // Try to load from server first
        const serverRegions = await regionsService.getRegionsForSlide(slideUid);
        console.log(
          `RegionSyncV2: Loaded ${serverRegions.length} regions from server`
        );

        // Save to localStorage for offline access
        saveLocalRegions(slideUid, serverRegions);
        return serverRegions;
      } catch (error) {
        console.warn(
          "RegionSyncV2: Server fetch failed, falling back to localStorage:",
          error
        );
        // Fall back to localStorage
        const localRegions = loadLocalRegions(slideUid);
        console.log(
          `RegionSyncV2: Loaded ${localRegions.length} regions from localStorage`
        );
        return localRegions;
      }
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes (was cacheTime)
    retry: 1, // Only retry once for regions
    enabled: !!slideUid,
  });

  // Create region mutation with optimistic updates
  const createRegionMutation = useMutation({
    mutationFn: async (
      newRegion: CreateRegionRequest & { tempId?: string }
    ): Promise<Region> => {
      const { tempId, ...regionData } = newRegion;

      // Check if this region was marked for deletion while creating
      if (tempId) {
        const pendingOp = pendingOperationsRef.current.get(tempId);
        if (pendingOp?.type === "delete") {
          // Region was deleted before creation completed
          // Create it anyway, then immediately delete it
          const serverRegion = await regionsService.createRegion(
            slideUid,
            regionData
          );

          // Immediately delete the server region
          await regionsService.deleteRegion(slideUid, serverRegion.regionId);

          // Clean up pending operations
          pendingOperationsRef.current.delete(tempId);

          // Don't add to cache since it was immediately deleted
          throw new Error("Region was deleted during creation");
        }
      }

      // Normal creation flow
      const serverRegion = await regionsService.createRegion(
        slideUid,
        regionData
      );
      console.log(
        "RegionSyncV2: Created region on server:",
        serverRegion.regionId
      );

      // Update the pending operation with server ID
      if (tempId && pendingOperationsRef.current.has(tempId)) {
        pendingOperationsRef.current.set(tempId, {
          type: "create",
          serverRegionId: serverRegion.regionId,
        });
      }

      // Trigger ID mapping event for the drawing manager
      if (tempId) {
        console.log(
          "RegionSyncV2: Triggering ID mapping event:",
          tempId,
          "->",
          serverRegion.regionId
        );
        setTimeout(() => {
          const mappingEvent = new CustomEvent("regionIdMapping", {
            detail: { tempId, serverRegionId: serverRegion.regionId },
          });
          window.dispatchEvent(mappingEvent);
          console.log("RegionSyncV2: Dispatched ID mapping event for", tempId);
        }, 100);
      }

      // Update localStorage with server region
      const currentRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];
      const updatedRegions = [...currentRegions, serverRegion];
      saveLocalRegions(slideUid, updatedRegions);

      return serverRegion;
    },
    onMutate: async (newRegion) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: regionQueryKey });

      // Snapshot previous value
      const previousRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];

      // Track this pending creation
      if (newRegion.tempId) {
        pendingOperationsRef.current.set(newRegion.tempId, { type: "create" });
      }

      // Create optimistic region
      const optimisticRegion: Region = {
        regionId: newRegion.tempId || `temp-${Date.now()}`,
        regionName: newRegion.regionName,
        slideUid,
        regionType: newRegion.regionType,
        geometry: newRegion.geometry,
        coordinateSystem: newRegion.coordinateSystem,
        mutable: true,
        visible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      // Optimistically update cache
      queryClient.setQueryData<Region[]>(regionQueryKey, [
        ...previousRegions,
        optimisticRegion,
      ]);

      return { previousRegions, optimisticRegion };
    },
    onError: (err, newRegion, context) => {
      // Clean up pending operations on error
      if (newRegion.tempId) {
        pendingOperationsRef.current.delete(newRegion.tempId);
      }

      // Rollback on error (unless it was a create-then-delete scenario)
      if (
        context?.previousRegions &&
        !err.message.includes("deleted during creation")
      ) {
        queryClient.setQueryData(regionQueryKey, context.previousRegions);
      }

      // Don't log error for create-then-delete scenarios
      if (!err.message.includes("deleted during creation")) {
        console.error("Error creating region:", err);
      }
    },
    onSuccess: (serverRegion, variables) => {
      // Clean up pending operations on successful creation
      if (variables.tempId) {
        pendingOperationsRef.current.delete(variables.tempId);
      }
    },
    onSettled: () => {
      // Always refetch after mutation
      queryClient.invalidateQueries({ queryKey: regionQueryKey });
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
    }): Promise<Region> => {
      const updatedRegion = await regionsService.updateRegion(
        regionId,
        updates
      );

      // Update localStorage
      const currentRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];
      const updatedRegions = currentRegions.map((r) =>
        r.regionId === regionId ? updatedRegion : r
      );
      saveLocalRegions(slideUid, updatedRegions);

      return updatedRegion;
    },
    onMutate: async ({ regionId, updates }) => {
      await queryClient.cancelQueries({ queryKey: regionQueryKey });

      const previousRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];

      // Optimistically update
      const optimisticRegions = previousRegions.map((region) =>
        region.regionId === regionId
          ? { ...region, ...updates, updatedAt: new Date().toISOString() }
          : region
      );

      queryClient.setQueryData<Region[]>(regionQueryKey, optimisticRegions);

      return { previousRegions };
    },
    onError: (err, variables, context) => {
      if (context?.previousRegions) {
        queryClient.setQueryData(regionQueryKey, context.previousRegions);
      }
      console.error("Error updating region:", err);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: regionQueryKey });
    },
  });

  // Delete region mutation with optimistic updates
  const deleteRegionMutation = useMutation({
    mutationFn: async (regionId: string): Promise<void> => {
      console.log("RegionSyncV2: Delete mutation called with ID:", regionId);
      let actualRegionId = regionId;

      // Check if this is a temp ID
      if (regionId.startsWith("temp-")) {
        console.log(
          "RegionSyncV2: Detected temp ID, looking for server mapping:",
          regionId
        );

        // First check pending operations
        console.log(
          "RegionSyncV2: Pending operations:",
          Array.from(pendingOperationsRef.current.entries())
        );
        const pendingOp = pendingOperationsRef.current.get(regionId);
        console.log(
          "RegionSyncV2: Pending operation for",
          regionId,
          ":",
          pendingOp
        );

        if (pendingOp?.type === "create") {
          if (pendingOp.serverRegionId) {
            // Creation completed, use server ID to delete
            actualRegionId = pendingOp.serverRegionId;
            console.log(
              "RegionSyncV2: Found server ID in pending ops:",
              actualRegionId
            );
          } else {
            // Creation still pending, mark for deletion
            pendingOperationsRef.current.set(regionId, { type: "delete" });
            console.log("RegionSyncV2: Marked pending creation for deletion");
            return; // Let the create mutation handle the deletion
          }
        } else {
          // Check localStorage for the server ID mapping
          const localStorageId = await findServerIdInLocalStorage(
            slideUid,
            regionId
          );
          if (localStorageId) {
            actualRegionId = localStorageId;
            console.log(
              "RegionSyncV2: Found server ID in localStorage:",
              actualRegionId
            );
          } else {
            // Check current regions cache
            const currentRegions =
              queryClient.getQueryData<Region[]>(regionQueryKey) || [];
            const region = currentRegions.find((r) => r.regionId === regionId);
            if (region && !region.regionId.startsWith("temp-")) {
              actualRegionId = region.regionId;
              console.log(
                "RegionSyncV2: Using region from cache:",
                actualRegionId
              );
            } else {
              console.warn(
                "RegionSyncV2: Could not find server ID for temp ID:",
                regionId,
                "deleting locally only"
              );
              // Don't call API for temp IDs without server mapping
              // Just clean up locally - the region might not have been successfully created on server
              pendingOperationsRef.current.delete(regionId);
              const currentRegions =
                queryClient.getQueryData<Region[]>(regionQueryKey) || [];
              const filteredRegions = currentRegions.filter(
                (r) => r.regionId !== regionId
              );
              saveLocalRegions(slideUid, filteredRegions);
              console.log(
                "RegionSyncV2: Deleted temp region locally:",
                regionId
              );
              return;
            }
          }
        }
      }

      // Delete from server using the actual server ID
      console.log(
        "RegionSyncV2: Deleting from server with ID:",
        actualRegionId
      );
      await regionsService.deleteRegion(slideUid, actualRegionId);

      // Clean up pending operations
      pendingOperationsRef.current.delete(regionId);

      // Update localStorage - remove using the original regionId (temp or server)
      const currentRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];
      const filteredRegions = currentRegions.filter(
        (r) => r.regionId !== regionId && r.regionId !== actualRegionId
      );
      saveLocalRegions(slideUid, filteredRegions);
    },
    onMutate: async (regionId) => {
      await queryClient.cancelQueries({ queryKey: regionQueryKey });

      const previousRegions =
        queryClient.getQueryData<Region[]>(regionQueryKey) || [];

      // Optimistically remove from cache
      const optimisticRegions = previousRegions.filter(
        (region) => region.regionId !== regionId
      );

      queryClient.setQueryData<Region[]>(regionQueryKey, optimisticRegions);

      return { previousRegions, deletedRegionId: regionId };
    },
    onError: (err, regionId, context) => {
      // Rollback on error
      if (context?.previousRegions) {
        queryClient.setQueryData(regionQueryKey, context.previousRegions);
      }
      console.error("Error deleting region:", err);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: regionQueryKey });
    },
  });

  // Convert regions to RegionItems for UI
  const regionItems: OptimisticRegionItem[] = regions.map((region) => ({
    ...regionToRegionItem(region),
    // Add optimistic states based on pending mutations
    isCreating:
      createRegionMutation.isPending &&
      createRegionMutation.variables?.regionName === region.regionName,
    isDeleting:
      deleteRegionMutation.isPending &&
      deleteRegionMutation.variables === region.regionId,
    isUpdating:
      updateRegionMutation.isPending &&
      updateRegionMutation.variables?.regionId === region.regionId,
  }));

  // Helper functions
  const createRegion = useCallback(
    (regionData: CreateRegionRequest & { tempId?: string }) => {
      // If no tempId provided, generate one (for direct API calls)
      const tempId =
        regionData.tempId ||
        `temp-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      return createRegionMutation.mutateAsync({ ...regionData, tempId });
    },
    [createRegionMutation]
  );

  const updateRegion = useCallback(
    (regionId: string, updates: UpdateRegionRequest) => {
      return updateRegionMutation.mutateAsync({ regionId, updates });
    },
    [updateRegionMutation]
  );

  const deleteRegion = useCallback(
    (regionId: string) => {
      return deleteRegionMutation.mutateAsync(regionId);
    },
    [deleteRegionMutation]
  );

  const manualSync = useCallback(async () => {
    try {
      const freshRegions = await refetch();
      return freshRegions.data || [];
    } catch (error) {
      console.error("Manual sync failed:", error);
      throw error;
    }
  }, [refetch]);

  return {
    // Data
    regions: regionItems,
    rawRegions: regions,

    // Loading states
    isLoading,
    isCreating: createRegionMutation.isPending,
    isUpdating: updateRegionMutation.isPending,
    isDeleting: deleteRegionMutation.isPending,
    isSyncing:
      createRegionMutation.isPending ||
      updateRegionMutation.isPending ||
      deleteRegionMutation.isPending,

    // Error states
    error: error?.message || null,
    createError: createRegionMutation.error?.message || null,
    updateError: updateRegionMutation.error?.message || null,
    deleteError: deleteRegionMutation.error?.message || null,

    // Actions
    createRegion,
    updateRegion,
    deleteRegion,
    manualSync,
    refetch,

    // Sync state
    lastSyncTime: new Date(), // TanStack Query handles this internally
  };
}
