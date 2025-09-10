// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useRef, useState } from "react";
import Map from "ol/Map";
import GeoJSON from "ol/format/GeoJSON";
import WebGLVectorLayer from "ol/layer/WebGLVector";
import VectorSource from "ol/source/Vector";
import Overlay from "ol/Overlay";
import { apiFetch } from "@/utils/fetchUtils";
import { useVectorContext } from "@/features/viewer/contexts/VectorContext";
import { SlideMetadata } from "@/features/viewer/components/map/types";

interface VectorAnnotation {
  vectorUid: string;
  vectorName: string;
  slideUid: string;
  labels: Array<{
    name: string;
    color: string;
  }>;
}

interface VectorAnnotationList {
  slideUid: string;
  annotations: VectorAnnotation[];
}

interface SlideAnnotationMetadata {
  slideUid: string;
  rasterUrl: string;
  vectorUrl: string;
  rasterCount: number;
  vectorCount: number;
}

interface UseVectorLayerProps {
  slideUid: string | undefined;
}

export function useVectorLayer({ slideUid }: UseVectorLayerProps) {
  const vectorLayersRef = useRef<WebGLVectorLayer[]>([]);
  const mapRef = useRef<Map | null>(null);
  const overlayRef = useRef<Overlay | null>(null);
  const slideMetadataRef = useRef<SlideMetadata | null>(null);
  const {
    vectorOpacity,
    showVectors,
    vectorColors,
    setVectorLayers,
    vectorLayers,
  } = useVectorContext();

  // Track when layers are loaded to trigger opacity updates
  const [layersLoaded, setLayersLoaded] = useState(0);

  // Dynamic label-to-color lookup that gets populated from API
  const labelColorMap = useRef<{ [key: string]: string }>({});

  const assignColorToLabel = useCallback(
    (label: string): string => {
      // Check if we already have a color for this label
      if (labelColorMap.current[label]) {
        return labelColorMap.current[label];
      }

      // For any other labels, assign from the color palette as fallback
      const existingLabels = Object.keys(labelColorMap.current);
      const colorIndex = existingLabels.length % vectorColors.length;
      const color = vectorColors[colorIndex];
      labelColorMap.current[label] = color;

      return color;
    },
    [vectorColors]
  );

  // Style configuration for WebGL Vector Layer - using variables for dynamic colors
  const createVectorStyle = useCallback(
    (colorIndex: number, initialColor?: string) => {
      const fallbackColor =
        initialColor || vectorColors[colorIndex % vectorColors.length];
      const rgb = hexToRgb(fallbackColor);

      return [
        {
          filter: ["==", ["var", "highlightedId"], ["id"]],
          style: {
            "stroke-color": "white",
            "stroke-width": 4,
            "stroke-offset": 0,
            "fill-color": [255, 255, 255, 0.5],
          },
        },
        {
          else: true,
          style: {
            "stroke-color": ["var", "strokeColor"],
            "stroke-width": 3,
            "stroke-offset": 0,
            "fill-color": ["var", "fillColor"],
          },
        },
      ];
    },
    [vectorColors]
  );

  // Helper function to update layer colors efficiently using variables
  const updateLayerColor = useCallback(
    (layer: WebGLVectorLayer, color: string) => {
      const rgb = hexToRgb(color);
      layer.updateStyleVariables({
        strokeColor: [rgb.r, rgb.g, rgb.b, 1.0],
        fillColor: [rgb.r, rgb.g, rgb.b, 0.7],
      });
    },
    []
  );

  // Fetch annotation metadata to check if slide has vector annotations
  const fetchAnnotationMetadata = useCallback(
    async (slideUid: string): Promise<SlideAnnotationMetadata | null> => {
      try {
        const response = await apiFetch<SlideAnnotationMetadata>(
          `/api/v1/slides/${slideUid}/annotations`
        );
        return response;
      } catch (error) {
        console.error("Failed to fetch annotation metadata:", error);
        return null;
      }
    },
    []
  );

  // Fetch vector annotations for a slide
  const fetchVectorAnnotations = useCallback(
    async (slideUid: string): Promise<VectorAnnotation[]> => {
      try {
        const response = await apiFetch<VectorAnnotationList>(
          `/api/v1/slides/${slideUid}/annotations/vector`
        );
        // Handle case where annotations might be null instead of empty array
        return response.annotations || [];
      } catch (error) {
        console.error("Failed to fetch vector annotations:", error);
        return [];
      }
    },
    []
  );

  // Fetch GeoJSON data for a specific vector annotation
  const fetchVectorGeoJSON = useCallback(
    async (slideUid: string, vectorUid: string): Promise<any> => {
      try {
        const response = await apiFetch(
          `/api/v1/slides/${slideUid}/annotations/vector/${vectorUid}/file`
        );
        return response;
      } catch (error) {
        console.error(
          `Failed to fetch GeoJSON for vector ${vectorUid}:`,
          error
        );
        return null;
      }
    },
    []
  );

  // Load vector annotations for the current slide
  const loadVectorAnnotations = useCallback(async () => {
    if (!slideUid || !mapRef.current) {
      return;
    }

    try {
      // Clear existing vector layers
      vectorLayersRef.current.forEach((layer) => {
        mapRef.current?.removeLayer(layer);
      });
      vectorLayersRef.current = [];

      // Reset label color map
      labelColorMap.current = {};

      // First check if slide has any vector annotations
      const annotationMetadata = await fetchAnnotationMetadata(slideUid);

      if (!annotationMetadata || annotationMetadata.vectorCount === 0) {
        setVectorLayers([]);
        setLayersLoaded((prev) => prev + 1);
        return;
      }

      // Fetch vector annotations
      const vectorAnnotations = await fetchVectorAnnotations(slideUid);

      // Early return if no annotations were actually loaded
      if (!vectorAnnotations || vectorAnnotations.length === 0) {
        setVectorLayers([]);
        setLayersLoaded((prev) => prev + 1);
        return;
      }

      // Build label-to-color map from API response
      vectorAnnotations.forEach((annotation) => {
        if (annotation.labels) {
          annotation.labels.forEach((labelInfo) => {
            labelColorMap.current[labelInfo.name] = labelInfo.color;
          });
        }
      });

      // Collect all features grouped by vector annotation and then by label
      const featuresByVectorAndLabel: {
        [vectorName: string]: { [label: string]: any[] };
      } = {};

      // Initialize feature groups for all vector annotations and their labels
      vectorAnnotations.forEach((annotation) => {
        featuresByVectorAndLabel[annotation.vectorName] = {};
        if (annotation.labels) {
          annotation.labels.forEach((labelInfo) => {
            featuresByVectorAndLabel[annotation.vectorName][labelInfo.name] =
              [];
          });
        }
      });

      // Load each vector annotation and group features by vector name and label
      for (let i = 0; i < vectorAnnotations.length; i++) {
        const annotation = vectorAnnotations[i];

        const geoJsonData = await fetchVectorGeoJSON(
          slideUid,
          annotation.vectorUid
        );

        if (geoJsonData?.features) {
          // Transform the entire GeoJSON first
          const transformedGeoJSON = transformGeoJSONCoordinates(
            geoJsonData,
            mapRef.current!
          );

          transformedGeoJSON.features.forEach((feature: any) => {
            const labelProperties = [
              "names",
              "name",
              "label",
              "title",
              "NAME",
              "LABEL",
              "TITLE",
            ];
            let featureLabel: string | null = null;

            for (const prop of labelProperties) {
              const value = feature.properties?.[prop];
              if (value && typeof value === "string" && value.trim()) {
                featureLabel = value.trim();
                break;
              }
            }

            if (featureLabel) {
              // Ensure the structure exists
              if (!featuresByVectorAndLabel[annotation.vectorName]) {
                featuresByVectorAndLabel[annotation.vectorName] = {};
              }
              if (
                !featuresByVectorAndLabel[annotation.vectorName][featureLabel]
              ) {
                featuresByVectorAndLabel[annotation.vectorName][featureLabel] =
                  [];
              }

              featuresByVectorAndLabel[annotation.vectorName][
                featureLabel
              ].push(feature);
            }
          });
        }
      }

      // Create layers for each vector annotation and label combination that has features
      const vectorLayerInfos: Array<{
        id: string;
        name: string;
        color: string;
        visible: boolean;
        vectorName?: string; // Add vector annotation name for grouping
      }> = [];

      let layerIndex = 0;
      for (const [vectorName, labelGroups] of Object.entries(
        featuresByVectorAndLabel
      )) {
        for (const [label, features] of Object.entries(labelGroups)) {
          if (features.length > 0) {
            // Create GeoJSON for this vector-label combination
            const labelGeoJSON = {
              type: "FeatureCollection",
              features: features,
            };

            const assignedColor = assignColorToLabel(label);

            // Create vector layer for this vector-label combination
            const vectorLayer = new WebGLVectorLayer({
              source: new VectorSource({
                features: new GeoJSON().readFeatures(labelGeoJSON),
              }),
              style: createVectorStyle(layerIndex, assignedColor),
              variables: {
                highlightedId: -1,
                strokeColor: (() => {
                  const rgb = hexToRgb(assignedColor);
                  return [rgb.r, rgb.g, rgb.b, 1.0];
                })(),
                fillColor: (() => {
                  const rgb = hexToRgb(assignedColor);
                  return [rgb.r, rgb.g, rgb.b, 0.7];
                })(),
              },
              opacity: 0, // Start with 0 opacity, will be set by the opacity effect
              visible: true,
            });

            // Add metadata for layer management
            vectorLayer.set("vectorName", label);
            vectorLayer.set("vectorUid", `${vectorName}-${label}`);
            vectorLayer.set("layerType", "vector");
            vectorLayer.set("assignedColor", assignedColor); // Store initial color for change detection
            vectorLayer.set(
              "debugInfo",
              `Vector layer for ${vectorName}-${label}`
            );
            vectorLayer.setZIndex(1000 + layerIndex);

            vectorLayersRef.current.push(vectorLayer);
            mapRef.current?.addLayer(vectorLayer);

            // DEBUG: Log vector layer creation
            console.log(
              `🔷 DEBUG: Created vector layer ${layerIndex} for ${vectorName}-${label}`,
              vectorLayer
            );

            // Add to context - name is just the label, vectorName is for grouping
            vectorLayerInfos.push({
              id: `${vectorName}-${label}`,
              name: label,
              color: assignedColor,
              visible: true,
              vectorName: vectorName,
            });

            layerIndex++;
          }
        }
      }

      // Update vector context with loaded layer information
      setVectorLayers(vectorLayerInfos);

      // Trigger opacity update for newly loaded layers
      setLayersLoaded((prev) => prev + 1);
    } catch (error) {
      console.error("Failed to load vector annotations:", error);
    }
  }, [
    slideUid,
    fetchAnnotationMetadata,
    fetchVectorAnnotations,
    fetchVectorGeoJSON,
    createVectorStyle,
    assignColorToLabel,
    setVectorLayers,
  ]);

  // Update vector layer opacity
  useEffect(() => {
    vectorLayersRef.current.forEach((layer) => {
      // Use layer opacity to control transparency
      layer.setOpacity(showVectors ? vectorOpacity : 0);
    });
  }, [vectorOpacity, showVectors, layersLoaded]);

  // Update individual layer visibility when toggled in control panel
  useEffect(() => {
    if (vectorLayers.length === 0) return;

    vectorLayersRef.current.forEach((layer) => {
      const vectorUid = layer.get("vectorUid");
      const contextLayer = vectorLayers.find((vl) => vl.id === vectorUid);

      if (contextLayer) {
        const shouldBeVisible = showVectors && contextLayer.visible;
        // Use actual opacity when visible, 0 when hidden
        layer.setOpacity(shouldBeVisible ? vectorOpacity : 0);
      }
    });
  }, [
    vectorLayers.map((vl) => `${vl.id}:${vl.visible}`).join(","), // Only trigger on visibility changes
    showVectors,
    vectorOpacity,
  ]);

  // Update vector layer colors when changed via color picker
  useEffect(() => {
    if (vectorLayers.length === 0) return;

    vectorLayersRef.current.forEach((layer) => {
      const vectorUid = layer.get("vectorUid");
      const contextLayer = vectorLayers.find((vl) => vl.id === vectorUid);

      if (contextLayer) {
        // Check if the color has changed from what's stored in the layer
        const currentColor = layer.get("assignedColor");
        if (currentColor !== contextLayer.color) {
          // Update the stored color and efficiently update the layer color
          layer.set("assignedColor", contextLayer.color);
          updateLayerColor(layer, contextLayer.color);
        }
      }
    });
  }, [
    vectorLayers.map((vl) => `${vl.id}:${vl.color}`).join(","), // Only trigger on color changes
    updateLayerColor,
  ]);

  // Handle vector layer creation (called from MapComponent)
  const handleVectorLayerCreated = useCallback(
    (map: Map, metadata: SlideMetadata) => {
      mapRef.current = map;
      slideMetadataRef.current = metadata;

      // Create tooltip overlay
      const tooltipElement = document.createElement("div");
      tooltipElement.className = "vector-tooltip";
      tooltipElement.style.cssText = `
            background: rgba(0, 0, 0, 0.8);
            color: white;
            padding: 8px 12px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
            pointer-events: none;
            white-space: nowrap;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
            z-index: 1000;
            max-width: 200px;
            word-wrap: break-word;
            display: none;
        `;

      const overlay = new Overlay({
        element: tooltipElement,
        offset: [0, -15],
        positioning: "bottom-center",
      });

      map.addOverlay(overlay);
      overlayRef.current = overlay;

      // Set up hover interaction
      let highlightedFeature: any = null;
      let currentVectorName: string | null = null;

      const displayFeatureInfo = (
        pixel: [number, number],
        coordinate?: [number, number]
      ) => {
        const feature = map.forEachFeatureAtPixel(pixel, (feature, layer) => {
          // Only process vector annotation layers
          if (layer && layer.get("layerType") === "vector") {
            return feature;
          }
          return null;
        });

        const id = feature ? feature.getId() : -1;

        // Find the vector name for this feature
        let vectorName: string | null = null;
        if (feature) {
          // Try to get a meaningful label from the feature properties
          // User specified 'names' property, check that first, then common alternatives
          const labelProperties = [
            "names",
            "name",
            "label",
            "title",
            "NAME",
            "LABEL",
            "TITLE",
          ];

          for (const prop of labelProperties) {
            const value = feature.get(prop);
            if (value && typeof value === "string" && value.trim()) {
              vectorName = value.trim();
              break;
            }
          }

          // If no named property found, try to use the feature ID
          if (!vectorName) {
            const featureId = feature.getId();
            if (featureId && typeof featureId === "string") {
              vectorName = featureId;
            } else if (featureId && typeof featureId === "number") {
              vectorName = `Feature ${featureId}`;
            }
          }

          // Last resort: fall back to the vector layer name
          if (!vectorName) {
            for (const layer of vectorLayersRef.current) {
              if (layer.getSource()?.hasFeature(feature)) {
                vectorName = layer.get("vectorName") || "Vector Feature";
                break;
              }
            }
          }
        }

        if (highlightedFeature !== feature) {
          // Reset previous highlight
          if (highlightedFeature) {
            vectorLayersRef.current.forEach((layer) => {
              layer.updateStyleVariables({ highlightedId: -1 });
            });
          }

          // Set new highlight
          if (feature) {
            vectorLayersRef.current.forEach((layer) => {
              if (layer.getSource()?.hasFeature(feature)) {
                layer.updateStyleVariables({ highlightedId: id || -1 });
              }
            });
          }

          highlightedFeature = feature;
        }

        // Update tooltip
        if (
          vectorName &&
          vectorName !== currentVectorName &&
          overlayRef.current &&
          coordinate
        ) {
          const tooltipElement = overlayRef.current.getElement();
          if (tooltipElement) {
            tooltipElement.textContent = vectorName;
            tooltipElement.style.display = "block";
            overlayRef.current.setPosition(coordinate);
          }
          currentVectorName = vectorName;
        } else if (!vectorName && overlayRef.current) {
          const tooltipElement = overlayRef.current.getElement();
          if (tooltipElement) {
            tooltipElement.style.display = "none";
          }
          currentVectorName = null;
        }
      };

      // Add hover interaction
      map.on("pointermove", (evt) => {
        if (evt.dragging) {
          return;
        }
        const coordinate =
          evt.coordinate && evt.coordinate.length >= 2
            ? ([evt.coordinate[0], evt.coordinate[1]] as [number, number])
            : undefined;
        displayFeatureInfo([evt.pixel[0], evt.pixel[1]], coordinate);
      });

      // Add click interaction
      map.on("click", (evt) => {
        const coordinate =
          evt.coordinate && evt.coordinate.length >= 2
            ? ([evt.coordinate[0], evt.coordinate[1]] as [number, number])
            : undefined;
        displayFeatureInfo([evt.pixel[0], evt.pixel[1]], coordinate);
      });

      // Hide tooltip when mouse leaves map
      map.getViewport().addEventListener("mouseleave", () => {
        if (overlayRef.current) {
          const tooltipElement = overlayRef.current.getElement();
          if (tooltipElement) {
            tooltipElement.style.display = "none";
          }
          currentVectorName = null;
        }
      });

      // Load vector annotations for this slide
      loadVectorAnnotations();
    },
    [loadVectorAnnotations]
  );

  // Reload vector annotations when slideUid changes
  useEffect(() => {
    if (mapRef.current && slideUid) {
      loadVectorAnnotations();
    }
  }, [slideUid, loadVectorAnnotations]);

  // Transform GeoJSON coordinates from pixel space to map coordinate space using slide metadata
  const transformGeoJSONCoordinates = useCallback(
    (geoJsonData: any, map: Map): any => {
      if (!geoJsonData?.features || !map) {
        return geoJsonData;
      }

      // Use actual slide metadata instead of trying to estimate mpp
      const slideMetadata = slideMetadataRef.current;
      if (!slideMetadata) {
        return geoJsonData;
      }

      const { slideMpp, slideWidth, slideHeight } = slideMetadata;

      // Transform coordinates using the exact SlideImage formula
      const transformCoordinate = (coord: number[]): [number, number] => {
        if (coord.length < 2) return [0, 0];

        // SlideImage transformation (from SlideImage.tsx):
        // X: pixelX * mpp * 1e-6
        // Y: -(pixelY * mpp * 1e-6)
        const mapX = coord[0] * slideMpp * 1e-6;
        const mapY = -(coord[1] * slideMpp * 1e-6);

        return [mapX, mapY];
      };

      const transformCoordinates = (coords: any): any => {
        if (
          Array.isArray(coords) &&
          coords.length >= 2 &&
          typeof coords[0] === "number"
        ) {
          return transformCoordinate(coords);
        } else if (Array.isArray(coords)) {
          return coords.map(transformCoordinates);
        }
        return coords;
      };

      // Create transformed GeoJSON
      const transformedGeoJSON = {
        ...geoJsonData,
        features: geoJsonData.features.map((feature: any) => ({
          ...feature,
          geometry: feature.geometry
            ? {
                ...feature.geometry,
                coordinates: transformCoordinates(feature.geometry.coordinates),
              }
            : feature.geometry,
        })),
      };

      return transformedGeoJSON;
    },
    []
  );

  // Clean up overlay when component unmounts
  useEffect(() => {
    return () => {
      if (overlayRef.current && mapRef.current) {
        mapRef.current.removeOverlay(overlayRef.current);
      }
    };
  }, []);

  return {
    handleVectorLayerCreated,
    vectorLayers: vectorLayersRef.current,
    reloadVectorAnnotations: loadVectorAnnotations,
  };
}

// Helper function to convert hex color or rgb() string to RGB
function hexToRgb(color: string): { r: number; g: number; b: number } {
  // Handle rgb() format
  const rgbMatch = color.match(/rgb\((\d+),\s*(\d+),\s*(\d+)\)/);
  if (rgbMatch) {
    return {
      r: parseInt(rgbMatch[1], 10),
      g: parseInt(rgbMatch[2], 10),
      b: parseInt(rgbMatch[3], 10),
    };
  }

  // Handle hex format
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(color);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : { r: 255, g: 0, b: 0 }; // Default to red
}
