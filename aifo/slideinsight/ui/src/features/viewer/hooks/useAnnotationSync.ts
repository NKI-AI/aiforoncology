// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  vectorAnnotationsService,
  type VectorAnnotation,
} from "@/services/vectorAnnotationsService";
import type {
  AnnotationItem,
  LabelId,
} from "@/features/viewer/components/AnnotationPanel";

interface AnnotationSyncState {
  isLoading: boolean;
  isSyncing: boolean;
  lastSyncTime: Date | null;
  error: string | null;
}

interface LocalAnnotation {
  id: string;
  label: LabelId; // Keep this for server communication
  name: string; // This is what we display (same as label)
  visible: boolean;
  kind?: "point" | "box" | "polygon";
  color?: string; // Hex color value from study settings
  geoJson: any; // GeoJSON feature
  isLocal: boolean; // Flag to indicate if this is a local-only annotation
  vectorUid?: string; // Server-side vector UID if synced
  lastModified: number; // Timestamp for conflict resolution
  geometryHash?: string; // Hash of geometry for deduplication
}

// Helper function to generate a hash from geometry for deduplication
async function generateGeometryHash(geoJson: any): Promise<string> {
  if (!geoJson?.geometry) return "";

  // Create a simplified representation of the geometry for hashing
  const geomString = JSON.stringify({
    type: geoJson.geometry.type,
    coordinates: geoJson.geometry.coordinates,
  });

  try {
    // Use crypto.subtle.digest for robust hashing
    const encoder = new TextEncoder();
    const data = encoder.encode(geomString);
    const hashBuffer = await crypto.subtle.digest("SHA-256", data);

    // Convert ArrayBuffer to hex string
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");

    return hashHex;
  } catch (error) {
    console.warn(
      "Failed to generate crypto hash, falling back to simple hash:",
      error
    );

    // Fallback to simple hash function if crypto.subtle is not available
    let hash = 0;
    for (let i = 0; i < geomString.length; i++) {
      const char = geomString.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return hash.toString();
  }
}

// Helper function to check if two annotations represent the same feature
function areAnnotationsSimilar(
  a: LocalAnnotation,
  b: LocalAnnotation
): boolean {
  // First check if they have the same geometry hash
  if (a.geometryHash && b.geometryHash && a.geometryHash === b.geometryHash) {
    return true;
  }

  // Fallback to checking if they have the same vector UID
  if (a.vectorUid && b.vectorUid && a.vectorUid === b.vectorUid) {
    return true;
  }

  // Check if they're from the same server annotation (same vector UID prefix)
  if (a.id.includes("-") && b.id.includes("-")) {
    const aPrefix = a.id.split("-")[0];
    const bPrefix = b.id.split("-")[0];
    if (aPrefix === bPrefix && aPrefix.length > 10) {
      // Likely a vector UID
      return true;
    }
  }

  return false;
}

export function useAnnotationSync(slideUid: string) {
  const [syncState, setSyncState] = useState<AnnotationSyncState>({
    isLoading: false,
    isSyncing: false,
    lastSyncTime: null,
    error: null,
  });

  const syncInProgressRef = useRef(false);
  const lastSyncTimeRef = useRef<number>(0);

  // Convert all local annotations to a single server vector annotation
  const convertAllLocalToServer = useCallback(
    (
      localAnnotations: LocalAnnotation[]
    ): Omit<VectorAnnotation, "vectorUid" | "createdAt" | "updatedAt"> => {
      // Combine all local annotations into a single FeatureCollection
      const features = localAnnotations.map((annotation) => ({
        ...annotation.geoJson,
        properties: {
          ...annotation.geoJson?.properties,
          name: annotation.name,
          kind: annotation.kind,
          color: annotation.color,
        },
      }));

      // Collect unique labels from all annotations
      const uniqueLabels = new Map<string, { name: string; color: string }>();
      localAnnotations.forEach((annotation) => {
        if (!uniqueLabels.has(annotation.name)) {
          uniqueLabels.set(annotation.name, {
            name: annotation.name,
            color: "#ff0000", // Default color, should be replaced with actual color from study settings
          });
        }
      });

      return {
        vectorName: "version 0",
        slideUid,
        dataBlob: {
          type: "FeatureCollection",
          features: features,
        },
        labels: Array.from(uniqueLabels.values()),
        actorType: "user",
      };
    },
    [slideUid]
  );

  // Convert server annotation to local format (handles multiple features)
  const convertServerToLocal = useCallback(
    async (serverAnnotation: VectorAnnotation): Promise<LocalAnnotation[]> => {
      // Safety check: ensure serverAnnotation exists and has required properties
      if (!serverAnnotation || !serverAnnotation.vectorUid) {
        console.warn("Invalid server annotation:", serverAnnotation);
        return [];
      }

      const features = serverAnnotation.dataBlob?.features || [
        serverAnnotation.dataBlob,
      ];

      const filteredFeatures = features.filter((feature) => feature);

      return Promise.all(
        filteredFeatures.map(async (feature: any, index: number) => {
          // Use the label and other info from feature properties, with fallbacks
          const featureLabel =
            feature?.properties?.label ||
            serverAnnotation.labels?.[0]?.name ||
            "roi";
          const featureName =
            feature?.properties?.name ||
            `${serverAnnotation.vectorName || "annotation"} ${index + 1}`;
          const featureKind =
            feature?.properties?.kind ||
            (feature?.geometry?.type === "Point"
              ? "point"
              : feature?.geometry?.type === "Polygon"
              ? "polygon"
              : "box");

          // Generate geometry hash for deduplication
          const geometryHash = await generateGeometryHash(feature);

          // Use geometry hash as part of ID for more stable identification
          const stableId =
            feature?.id || `${serverAnnotation.vectorUid}-${geometryHash}`;

          return {
            id: stableId,
            label: featureLabel as LabelId,
            name: featureName || featureLabel,
            visible: true,
            kind: featureKind,
            color: feature.properties?.color,
            geoJson: feature,
            isLocal: false,
            vectorUid: serverAnnotation.vectorUid,
            lastModified: new Date(
              serverAnnotation.updatedAt ||
                serverAnnotation.createdAt ||
                Date.now()
            ).getTime(),
            geometryHash,
          };
        })
      );
    },
    []
  );

  // Load annotations from localStorage
  const loadLocalAnnotations = useCallback(async (): Promise<
    LocalAnnotation[]
  > => {
    try {
      const key = `slideAnnotations:${slideUid}`;
      const raw = localStorage.getItem(key);
      console.log(
        `Loading local annotations for ${slideUid}:`,
        raw ? "found data" : "no data"
      );
      if (!raw) return [];

      const data = JSON.parse(raw);
      if (!data.features || !Array.isArray(data.features)) {
        console.log("No features found in localStorage data");
        return [];
      }

      const localAnnotations = await Promise.all(
        data.features.map(async (feature: any) => {
          const geometryHash = await generateGeometryHash(feature);
          const vectorUid = feature.properties?.vectorUid;
          const isLocal =
            !vectorUid || vectorUid === undefined || vectorUid === null;

          const annotation = {
            id: feature.id || `local-${Date.now()}-${Math.random()}`,
            label:
              feature.properties?.name || feature.properties?.label || "roi", // Use name as label
            name:
              feature.properties?.name || feature.properties?.label || "roi",
            visible: feature.properties?.visible !== false,
            kind: feature.properties?.kind,
            color: feature.properties?.color,
            geoJson: feature,
            isLocal,
            vectorUid,
            lastModified: feature.properties?.lastModified || Date.now(),
            geometryHash,
          };

          // Debug logging for modified annotations
          if (feature.properties?.lastModified && isLocal) {
            console.log(`Found modified local annotation ${annotation.id}:`, {
              isLocal,
              vectorUid,
              lastModified: new Date(annotation.lastModified).toISOString(),
            });
          }

          return annotation;
        })
      );
      console.log(`Loaded ${localAnnotations.length} local annotations`);
      return localAnnotations;
    } catch (error) {
      console.error("Error loading local annotations:", error);
      return [];
    }
  }, [slideUid]);

  // Save annotations to localStorage
  const saveLocalAnnotations = useCallback(
    (annotations: LocalAnnotation[]) => {
      try {
        const key = `slideAnnotations:${slideUid}`;
        const featureCollection = {
          type: "FeatureCollection",
          features: annotations.map((ann) => ({
            ...ann.geoJson,
            id: ann.id,
            properties: {
              ...ann.geoJson.properties,
              name: ann.name,
              visible: ann.visible,
              kind: ann.kind,
              color: ann.color,
              lastModified: ann.lastModified,
              vectorUid: ann.vectorUid,
              geometryHash: ann.geometryHash,
            },
          })),
        };
        localStorage.setItem(key, JSON.stringify(featureCollection));
      } catch (error) {
        console.error("Error saving local annotations:", error);
      }
    },
    [slideUid]
  );

  // Load server annotations
  const loadServerAnnotations = useCallback(async (): Promise<
    VectorAnnotation[]
  > => {
    try {
      const annotations =
        await vectorAnnotationsService.getVectorAnnotationsForSlide(slideUid);
      return Array.isArray(annotations) ? annotations : [];
    } catch (error) {
      console.error("Error loading server annotations:", error);
      throw error;
    }
  }, [slideUid]);

  // Sync local annotations to server as a single vector annotation
  const syncToServer = useCallback(
    async (localAnnotations: LocalAnnotation[]) => {
      if (localAnnotations.length === 0) {
        return [];
      }

      // Check if we already have a "version 0" vector annotation for this slide
      const existingAnnotations =
        await vectorAnnotationsService.getVectorAnnotationsForSlide(slideUid);
      const existingVersionAnnotation = existingAnnotations.find(
        (ann) => ann && ann.vectorName === "version 0" && ann.vectorUid
      );

      let serverData: Omit<
        VectorAnnotation,
        "vectorUid" | "createdAt" | "updatedAt"
      >;

      if (existingVersionAnnotation && existingVersionAnnotation.vectorUid) {
        // Merge local annotations with existing server annotations
        console.log(
          `Updating existing "version 0" annotation - merging ${localAnnotations.length} local annotations with server data`
        );

        const existingFeatures =
          existingVersionAnnotation.dataBlob?.features || [];
        const localFeatures = localAnnotations.map((annotation) => ({
          ...annotation.geoJson,
          properties: {
            ...annotation.geoJson?.properties,
            name: annotation.name,
            kind: annotation.kind,
            color: annotation.color,
            // Keep the vectorUid in the server data to maintain the relationship
            vectorUid: existingVersionAnnotation.vectorUid,
          },
        }));

        // Start with existing server features
        const allFeatures = [...existingFeatures];
        const processedIds = new Set<string>();

        // Track existing feature IDs
        existingFeatures.forEach((feature: any) => {
          if (feature.id) {
            processedIds.add(String(feature.id));
          }
        });

        // Update or add local features
        localFeatures.forEach((localFeature) => {
          const localId = String(localFeature.id);

          if (processedIds.has(localId)) {
            // Replace existing feature with local version (this handles modifications)
            const existingIndex = allFeatures.findIndex(
              (f: any) => String(f.id) === localId
            );
            if (existingIndex >= 0) {
              allFeatures[existingIndex] = localFeature;
              console.log(
                `Updated existing server feature ${localId} with local changes`
              );
            }
          } else {
            // Add new local feature
            allFeatures.push(localFeature);
            processedIds.add(localId);
            console.log(`Added new local feature ${localId} to server`);
          }
        });

        // Collect all unique labels from both existing and local annotations
        const uniqueLabels = new Map<string, { name: string; color: string }>();

        // Add existing labels from server
        if (existingVersionAnnotation.labels) {
          existingVersionAnnotation.labels.forEach((label) => {
            uniqueLabels.set(label.name, label);
          });
        }

        // Add labels from local annotations (may override existing ones)
        localAnnotations.forEach((annotation) => {
          uniqueLabels.set(annotation.label, {
            name: annotation.label,
            color: annotation.color || "#ff0000", // Use annotation color or default
          });
        });

        serverData = {
          vectorName: "version 0",
          slideUid,
          dataBlob: {
            type: "FeatureCollection",
            features: allFeatures,
          },
          labels: Array.from(uniqueLabels.values()),
          actorType: "user",
        };
      } else {
        // No existing annotation, create new one with just the local annotations
        serverData = convertAllLocalToServer(localAnnotations);
      }

      let result: VectorAnnotation;
      if (existingVersionAnnotation && existingVersionAnnotation.vectorUid) {
        // Update existing "version 0" annotation
        result = await vectorAnnotationsService.updateVectorAnnotation(
          slideUid,
          existingVersionAnnotation.vectorUid,
          serverData
        );
      } else {
        // Create new "version 0" annotation
        result = await vectorAnnotationsService.createVectorAnnotation(
          slideUid,
          serverData
        );
      }

      return [result];
    },
    [slideUid, convertAllLocalToServer]
  );

  // Perform full sync
  const performSync = useCallback(async (): Promise<LocalAnnotation[]> => {
    if (syncInProgressRef.current) {
      throw new Error("Sync already in progress");
    }

    syncInProgressRef.current = true;
    setSyncState((prev) => ({ ...prev, isSyncing: true, error: null }));

    try {
      // Load both local and server annotations
      const [localAnnotations, serverAnnotations] = await Promise.all([
        loadLocalAnnotations(),
        loadServerAnnotations(),
      ]);

      // Convert server annotations to local format
      const serverAsLocalArrays = await Promise.all(
        (serverAnnotations || []).map(convertServerToLocal)
      );
      const serverAsLocal = serverAsLocalArrays.flat();

      // Merge annotations with proper deduplication
      const mergedAnnotations: LocalAnnotation[] = [];
      const processedHashes = new Set<string>();
      const processedIds = new Set<string>();

      // First, add server annotations (they take precedence)
      for (const serverAnnotation of serverAsLocal) {
        mergedAnnotations.push(serverAnnotation);
        if (serverAnnotation.geometryHash) {
          processedHashes.add(serverAnnotation.geometryHash);
        }
        processedIds.add(serverAnnotation.id);
      }

      // Then add local annotations that don't duplicate server ones
      for (const localAnnotation of localAnnotations) {
        let isDuplicate = false;
        let duplicateReason = "";

        // Special handling for modified local annotations - they should override server versions
        if (localAnnotation.isLocal && localAnnotation.lastModified) {
          // Check if this is a modified version of a server annotation
          const serverMatch = serverAsLocal.find(
            (s) => s.id === localAnnotation.id
          );
          if (serverMatch) {
            // This is a modified local annotation that should replace the server version
            console.log(
              `Found modified local annotation ${localAnnotation.id} that overrides server version`
            );

            // Remove the server version from merged annotations and add the local version
            const serverIndex = mergedAnnotations.findIndex(
              (m) => m.id === localAnnotation.id
            );
            if (serverIndex >= 0) {
              mergedAnnotations[serverIndex] = localAnnotation;
              console.log(
                `Replaced server annotation ${localAnnotation.id} with modified local version`
              );
            } else {
              mergedAnnotations.push(localAnnotation);
            }

            if (localAnnotation.geometryHash) {
              processedHashes.add(localAnnotation.geometryHash);
            }
            processedIds.add(localAnnotation.id);
            continue; // Skip the normal duplicate checking
          }
        }

        // Check for duplicates using multiple strategies
        if (processedIds.has(localAnnotation.id)) {
          isDuplicate = true;
          duplicateReason = "ID match";
        } else if (
          localAnnotation.geometryHash &&
          processedHashes.has(localAnnotation.geometryHash)
        ) {
          isDuplicate = true;
          duplicateReason = "geometry hash match";
        } else {
          // Check if this local annotation is similar to any server annotation
          for (const serverAnnotation of serverAsLocal) {
            if (areAnnotationsSimilar(localAnnotation, serverAnnotation)) {
              isDuplicate = true;
              duplicateReason = "similarity match";
              break;
            }
          }
        }

        if (isDuplicate) {
          console.log(
            `Skipping duplicate local annotation ${localAnnotation.id} (reason: ${duplicateReason})`
          );
        } else {
          mergedAnnotations.push(localAnnotation);
          if (localAnnotation.geometryHash) {
            processedHashes.add(localAnnotation.geometryHash);
          }
          processedIds.add(localAnnotation.id);
        }
      }

      // Sync all local annotations to server (both new and modified)
      const localAnnotationsToSync = mergedAnnotations.filter(
        (ann) => ann.isLocal
      );

      console.log(`Sync analysis:`, {
        totalMerged: mergedAnnotations.length,
        localToSync: localAnnotationsToSync.length,
        localAnnotations: localAnnotationsToSync.map((ann) => ({
          id: ann.id,
          isLocal: ann.isLocal,
          vectorUid: ann.vectorUid,
          lastModified: new Date(ann.lastModified).toISOString(),
        })),
      });

      if (localAnnotationsToSync.length > 0) {
        console.log(
          `Syncing ${localAnnotationsToSync.length} local annotations to server (including modified ones)`
        );

        const createdAnnotations = await syncToServer(localAnnotationsToSync);

        // Update local annotations with server info
        if (createdAnnotations.length > 0) {
          const serverAnnotation = createdAnnotations[0]; // The single "version 0" annotation

          // Convert the server annotation back to local format to get updated IDs
          const updatedServerAsLocal = await convertServerToLocal(
            serverAnnotation
          );

          // Mark all synced local annotations as no longer local
          for (let i = 0; i < mergedAnnotations.length; i++) {
            const annotation = mergedAnnotations[i];
            if (
              localAnnotationsToSync.some((local) => local.id === annotation.id)
            ) {
              // Mark as synced
              mergedAnnotations[i] = {
                ...annotation,
                vectorUid: serverAnnotation.vectorUid,
                isLocal: false,
                lastModified: new Date(
                  serverAnnotation.updatedAt ||
                    serverAnnotation.createdAt ||
                    Date.now()
                ).getTime(),
              };
              console.log(
                `Marked annotation ${annotation.id} as synced with server`
              );
            }
          }

          console.log(
            `Successfully synced ${localAnnotationsToSync.length} annotations. Server annotation updated: ${serverAnnotation.vectorUid}`
          );
        }
      }

      // Save merged annotations back to localStorage
      saveLocalAnnotations(mergedAnnotations);

      setSyncState((prev) => ({
        ...prev,
        lastSyncTime: new Date(),
        error: null,
      }));

      lastSyncTimeRef.current = Date.now();
      return mergedAnnotations;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown sync error";
      setSyncState((prev) => ({ ...prev, error: errorMessage }));
      throw error;
    } finally {
      syncInProgressRef.current = false;
      setSyncState((prev) => ({ ...prev, isSyncing: false }));
    }
  }, [
    loadLocalAnnotations,
    loadServerAnnotations,
    convertServerToLocal,
    syncToServer,
    saveLocalAnnotations,
  ]);

  // Load initial annotations (server + local merge)
  const loadAnnotations = useCallback(async (): Promise<AnnotationItem[]> => {
    setSyncState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const mergedAnnotations = await performSync();

      // Convert to AnnotationItem format
      return mergedAnnotations.map((ann) => ({
        id: ann.id,
        name: ann.name, // Use name field
        visible: ann.visible,
        kind: ann.kind,
      }));
    } catch (error) {
      console.error("Error loading annotations:", error);
      // Fall back to local annotations only
      const localAnnotations = await loadLocalAnnotations();
      return localAnnotations.map((ann) => ({
        id: ann.id,
        name: ann.name, // Use name field
        visible: ann.visible,
        kind: ann.kind,
      }));
    } finally {
      setSyncState((prev) => ({ ...prev, isLoading: false }));
    }
  }, [performSync, loadLocalAnnotations]);

  // Periodic sync (every 30 seconds)
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      // Only sync if it's been more than 30 seconds since last sync
      if (now - lastSyncTimeRef.current > 30000 && !syncInProgressRef.current) {
        performSync().catch((error) => {
          console.error("Periodic sync failed:", error);
        });
      }
    }, 30000);

    return () => clearInterval(interval);
  }, [performSync]);

  // Delete annotation function
  const deleteAnnotation = useCallback(
    async (annotationId: string): Promise<void> => {
      try {
        // Load local annotations to find the one to delete
        const localAnnotations = await loadLocalAnnotations();
        const annotationToDelete = localAnnotations.find(
          (ann) => ann.id === annotationId
        );

        if (!annotationToDelete) {
          console.warn(`Annotation ${annotationId} not found locally`);
          return;
        }

        // Remove from localStorage first
        const updatedLocalAnnotations = localAnnotations.filter(
          (ann) => ann.id !== annotationId
        );
        saveLocalAnnotations(updatedLocalAnnotations);

        // If the annotation has a vectorUid, we need to update the server vector annotation
        // by removing just this feature, not deleting the entire vector annotation
        if (annotationToDelete.vectorUid) {
          try {
            // Find all other annotations that belong to the same vector annotation
            const sameVectorAnnotations = updatedLocalAnnotations.filter(
              (ann) => ann.vectorUid === annotationToDelete.vectorUid
            );

            if (sameVectorAnnotations.length > 0) {
              // There are still other features in this vector annotation
              // Update the vector annotation with the remaining features
              const updatedVectorData = convertAllLocalToServer(
                sameVectorAnnotations
              );
              await vectorAnnotationsService.updateVectorAnnotation(
                slideUid,
                annotationToDelete.vectorUid,
                updatedVectorData
              );
              console.log(
                `Updated vector annotation ${annotationToDelete.vectorUid} after removing feature ${annotationId}`
              );
            } else {
              // This was the last feature in the vector annotation, delete the entire vector
              await vectorAnnotationsService.deleteVectorAnnotation(
                slideUid,
                annotationToDelete.vectorUid
              );
              console.log(
                `Deleted entire vector annotation ${annotationToDelete.vectorUid} as it had no remaining features`
              );
            }
          } catch (error) {
            console.error(
              `Failed to update server after deleting annotation ${annotationId}:`,
              error
            );
            // Don't re-throw - we've already removed it locally
          }
        }

        console.log(`Deleted annotation ${annotationId} locally`);
      } catch (error) {
        console.error(`Error deleting annotation ${annotationId}:`, error);
        throw error;
      }
    },
    [
      slideUid,
      loadLocalAnnotations,
      saveLocalAnnotations,
      convertAllLocalToServer,
    ]
  );

  // Manual sync function - just use the same performSync that automatic sync uses
  const manualSync = performSync;

  return {
    syncState,
    loadAnnotations,
    manualSync,
    deleteAnnotation,
    isLoading: syncState.isLoading,
    isSyncing: syncState.isSyncing,
    lastSyncTime: syncState.lastSyncTime,
    error: syncState.error,
  };
}
