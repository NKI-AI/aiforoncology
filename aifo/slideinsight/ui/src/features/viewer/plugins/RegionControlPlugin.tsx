// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect, useState, useRef } from "react";
import { RectangleStackIcon } from "@heroicons/react/24/solid";
import { ViewerPlugin, PluginAPI } from "./types";
import RegionPanel, {
  type RegionItem,
} from "@/features/viewer/components/RegionPanel";
import { useCoordinateTransforms } from "@/features/viewer/components/SlideWorkspace/hooks/useCoordinateTransforms";
import { useRegionSyncV2 } from "@/features/viewer/hooks/useRegionSyncV2";
import {
  removeFeaturesById,
  findInSourcesById,
  updateFeaturesById,
} from "@/features/viewer/ol/utils";
import {
  loadRegionSettings,
  saveRegionSettings,
  resetRegionSettings,
  type RegionSettings,
  type RegionSettingsPartial,
} from "@/features/viewer/utils/regionSettingsStorage";
import DragPan from "ol/interaction/DragPan";
import Transform from "ol-ext/interaction/Transform";
import type OlMap from "ol/Map";
import VectorSource from "ol/source/Vector";
import VectorLayer from "ol/layer/Vector";
import Draw, { createBox } from "ol/interaction/Draw";
import Modify from "ol/interaction/Modify";
import Snap from "ol/interaction/Snap";
import Feature from "ol/Feature";
import Polygon from "ol/geom/Polygon";
import { Style, Stroke, Fill, Text, RegularShape } from "ol/style";
import GeoJSON from "ol/format/GeoJSON";

interface RegionPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
}

/**
 * Manages all region drawing tools and interactions
 */
class RegionDrawingManager {
  private map: OlMap | null = null;
  private slideUid: string = "";
  private rawSlideMetadata: any = null;
  private transformMapToPixelCoordinates: ((geoJson: any) => any) | null = null;
  private transformPixelToMapCoordinates: ((geoJson: any) => any) | null = null;

  // Layer management
  private tempRegionLayerRef: {
    source: VectorSource;
    layer: VectorLayer;
  } | null = null;
  private persistedRegionLayerRef: {
    source: VectorSource;
    layer: VectorLayer;
  } | null = null;

  // Interaction management
  private activeRegionDrawRef: Draw | null = null;
  private regionModifyInteractionRef: Modify | null = null;
  private regionSnapInteractionRef: Snap | null = null;
  private regionTransformRef: Transform | null = null;
  private dragPanInteractionRef: DragPan | null = null;

  // State
  private regionIdCounterRef = 1;
  private selectedRegionIdRef: string | null = null;
  private hoveredRegionIdRef: string | null = null;
  private isModifyingRegionRef = false;
  private isDrawingRegionRef = false;

  // Callbacks
  private onRegionsUpdate: ((regions: RegionItem[]) => void) | null = null;
  private onRegionSelect: ((id: string | null) => void) | null = null;
  private onRegionHover: ((id: string | null) => void) | null = null;
  private onDrawingStateChange: ((isDrawing: boolean) => void) | null = null;
  private onEditingStateChange: ((isEditing: boolean) => void) | null = null;
  private onCreateRegion: ((regionData: any) => Promise<void>) | null = null;

  // Display settings
  private regionDisplaySettings: RegionSettings = loadRegionSettings();
  private onUpdateRegionDisplaySettings:
    | ((settings: RegionSettingsPartial) => void)
    | null = null;
  private onResetRegionDisplaySettings: (() => void) | null = null;

  // Dynamic styles based on settings
  private getRegionStyles() {
    const settings = this.regionDisplaySettings;

    // Helper function to create fill or return undefined for zero opacity
    const createFill = (baseColor: string, opacity: number) => {
      if (opacity === 0) {
        return undefined; // No fill when opacity is 0
      }
      const opacityHex = Math.round(opacity * 255)
        .toString(16)
        .padStart(2, "0");
      return new Fill({ color: `${baseColor}${opacityHex}` });
    };

    const selectedOpacity = Math.min(1, settings.fillOpacity + 0.15);
    const hoverOpacity = Math.min(1, settings.fillOpacity + 0.13);

    return {
      normal: new Style({
        stroke: new Stroke({ color: "#8b5cf6", width: settings.strokeWidth }),
        fill: createFill("#8b5cf6", settings.fillOpacity),
      }),
      selected: new Style({
        stroke: new Stroke({
          color: "#0ea5e9",
          width: settings.strokeWidth + 2,
        }),
        fill: createFill("#0ea5e9", selectedOpacity),
      }),
      hover: new Style({
        stroke: new Stroke({
          color: "#f59e0b",
          width: settings.strokeWidth + 1,
        }),
        fill: createFill("#f59e0b", hoverOpacity),
      }),
    };
  }

  private modifyStyles = {
    handle: new Style({
      image: new RegularShape({
        points: 4, // square
        radius: 8, // size of the handle
        angle: Math.PI / 4, // oriented like a diamond
        fill: new Fill({ color: "rgba(14,165,233,0.9)" }), // cyan-ish
        stroke: new Stroke({ color: "#ffffff", width: 4 }),
      }),
    }),
    segment: new Style({
      stroke: new Stroke({ color: "rgba(14,165,233,0.8)", width: 4 }),
    }),
  };

  initialize(
    map: OlMap,
    slideUid: string,
    rawSlideMetadata: any,
    transformMapToPixelCoordinates: (geoJson: any) => any,
    transformPixelToMapCoordinates: (geoJson: any) => any,
    callbacks: {
      onRegionsUpdate: (regions: RegionItem[]) => void;
      onRegionSelect: (id: string | null) => void;
      onRegionHover: (id: string | null) => void;
      onDrawingStateChange: (isDrawing: boolean) => void;
      onEditingStateChange: (isEditing: boolean) => void;
      onUpdateRegionDisplaySettings: (settings: RegionSettingsPartial) => void;
      onResetRegionDisplaySettings: () => void;
      onCreateRegion: (regionData: any) => Promise<void>;
    }
  ) {
    this.map = map;
    this.slideUid = slideUid;
    this.rawSlideMetadata = rawSlideMetadata;
    this.transformMapToPixelCoordinates = transformMapToPixelCoordinates;
    this.transformPixelToMapCoordinates = transformPixelToMapCoordinates;
    this.onRegionsUpdate = callbacks.onRegionsUpdate;
    this.onRegionSelect = callbacks.onRegionSelect;
    this.onRegionHover = callbacks.onRegionHover;
    this.onDrawingStateChange = callbacks.onDrawingStateChange;
    this.onEditingStateChange = callbacks.onEditingStateChange;
    this.onUpdateRegionDisplaySettings =
      callbacks.onUpdateRegionDisplaySettings;
    this.onResetRegionDisplaySettings = callbacks.onResetRegionDisplaySettings;
    this.onCreateRegion = callbacks.onCreateRegion;

    // Find drag pan interaction
    const dragPan =
      map
        .getInteractions()
        .getArray()
        .find((i): i is DragPan => i instanceof DragPan) || null;
    this.dragPanInteractionRef = dragPan;

    this.ensurePersistedRegionLayer();
    this.initializeRegionModifyInteractions();
    this.loadPersistedRegions();
    this.setupMapEventHandlers();
    this.setupIdMappingListener();
  }

  private getRegionStyle = (feature: any) => {
    const id = String(feature.getId?.() ?? "");
    const name = feature.get?.("name") || `Region ${id}`;

    const isSelected = id === this.selectedRegionIdRef;
    const isHovered = id === this.hoveredRegionIdRef;

    const styles = this.getRegionStyles();
    const base = isSelected
      ? styles.selected
      : isHovered
      ? styles.hover
      : styles.normal;

    // Hide label while modifying/drawing → much smoother
    const drawLabel = !this.isModifyingRegionRef && !this.isDrawingRegionRef;

    return new Style({
      stroke: base.getStroke() || undefined,
      fill: base.getFill() || undefined,
      text: drawLabel
        ? new Text({
            text: name,
            font: "14px Arial, sans-serif",
            fill: new Fill({
              color: isSelected ? "#0ea5e9" : isHovered ? "#f59e0b" : "#8b5cf6",
            }),
            // Reduce white stroke when fill opacity is very low to minimize "glow"
            stroke:
              this.regionDisplaySettings.fillOpacity > 0.1
                ? new Stroke({ color: "#ffffff", width: 2 })
                : new Stroke({ color: "#ffffff", width: 1 }),
            textAlign: "center",
            textBaseline: "middle",
          })
        : undefined,
    });
  };

  private createTemporaryRegionLayer() {
    if (!this.map) return null;
    if (this.tempRegionLayerRef) return this.tempRegionLayerRef;

    const source = new VectorSource({ wrapX: false });
    const layer = new VectorLayer({
      source,
      style: (feature) => {
        const name = feature.get("name") || "Drawing...";
        const text = new Text({
          text: name,
          font: "14px Arial, sans-serif",
          fill: new Fill({ color: "#8b5cf6" }),
          // Reduce white stroke when fill opacity is very low to minimize "glow"
          stroke:
            this.regionDisplaySettings.fillOpacity > 0.1
              ? new Stroke({ color: "#ffffff", width: 2 })
              : new Stroke({ color: "#ffffff", width: 1 }),
          textAlign: "center",
          textBaseline: "middle",
        });

        const styles = this.getRegionStyles();
        return new Style({
          stroke: styles.normal.getStroke() || undefined,
          fill: styles.normal.getFill() || undefined,
          text,
        });
      },
    });

    layer.setZIndex(2100); // Higher than annotation layers
    this.map.addLayer(layer);

    this.tempRegionLayerRef = { source, layer };
    return this.tempRegionLayerRef;
  }

  private ensurePersistedRegionLayer() {
    if (!this.map) return null;
    if (this.persistedRegionLayerRef) return this.persistedRegionLayerRef;

    const source = new VectorSource({ wrapX: false });
    const layer = new VectorLayer({
      source,
      style: this.getRegionStyle,
    });
    layer.setZIndex(1600); // Between annotations and temp regions
    this.map.addLayer(layer);
    this.persistedRegionLayerRef = { source, layer };
    return this.persistedRegionLayerRef;
  }

  private rectFromTwoCorners(a: [number, number], b: [number, number]) {
    const minX = Math.min(a[0], b[0]);
    const maxX = Math.max(a[0], b[0]);
    const minY = Math.min(a[1], b[1]);
    const maxY = Math.max(a[1], b[1]);
    return [
      [
        [minX, minY],
        [maxX, minY],
        [maxX, maxY],
        [minX, maxY],
        [minX, minY],
      ],
    ];
  }

  private initializeRegionModifyInteractions() {
    if (!this.map) return;

    // Remove previous interactions
    if (this.regionModifyInteractionRef) {
      try {
        this.map.removeInteraction(this.regionModifyInteractionRef);
      } catch {}
      this.regionModifyInteractionRef = null;
    }
    if (this.regionSnapInteractionRef) {
      try {
        this.map.removeInteraction(this.regionSnapInteractionRef);
      } catch {}
      this.regionSnapInteractionRef = null;
    }
    if (this.regionTransformRef) {
      try {
        this.map.removeInteraction(this.regionTransformRef);
      } catch {}
      this.regionTransformRef = null;
    }

    const persisted = this.persistedRegionLayerRef;
    if (!persisted) return;

    //             modifyCenter: (evt: any) => !(evt?.originalEvent?.metaKey || evt?.originalEvent?.ctrlKey), // eslint-disable-line

    const transform = new Transform({
      layers: (layer) => layer === persisted.layer,
      filter: (feature) =>
        String(feature.getId?.() ?? "") === this.selectedRegionIdRef,
      translate: true,
      scale: true,
      stretch: true,
      rotate: false,
      hitTolerance: 10,
      modifyCenter: (evt: any) =>
        !(evt?.originalEvent?.metaKey || evt?.originalEvent?.ctrlKey),
    });

    transform.on("transformstart" as any, () => {
      this.isModifyingRegionRef = true;
      try {
        this.dragPanInteractionRef?.setActive(false);
      } catch {}
    });

    transform.on("transformend" as any, (e: any) => {
      this.isModifyingRegionRef = false;
      try {
        this.dragPanInteractionRef?.setActive(true);
      } catch {}

      const feature =
        e.feature ||
        (e.features && e.features.item && e.features.item(0)) ||
        null;
      if (!feature) return;

      const geom = feature.getGeometry();
      if (geom instanceof Polygon) {
        const ext = geom.getExtent();
        const anchor: [number, number] = [ext[0], ext[1]];
        const moved: [number, number] = [ext[2], ext[3]];
        geom.setCoordinates(this.rectFromTwoCorners(anchor, moved));
      }

      this.persistRegionGeometry(feature);
    });

    this.map.addInteraction(transform);
    this.regionTransformRef = transform;

    transform.setActive(
      this.selectedRegionIdRef !== null && !this.isDrawingRegionRef
    );
  }

  private persistRegionGeometry(feature: Feature) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const format = new GeoJSON();
      const geoJsonInMapCoords = format.writeFeatureObject(feature);
      const geoJsonInPixelCoords = this.transformMapToPixelCoordinates({
        type: "FeatureCollection",
        features: [geoJsonInMapCoords],
      });

      const key = `slideRegions:${this.slideUid}`;
      const existing = localStorage.getItem(key);
      let existingData: any = { type: "FeatureCollection", features: [] };

      if (existing) {
        try {
          existingData = JSON.parse(existing);
          if (!existingData || !Array.isArray(existingData.features)) {
            existingData = { type: "FeatureCollection", features: [] };
          }
        } catch {
          existingData = { type: "FeatureCollection", features: [] };
        }
      }

      const featureId = String(feature.getId());
      const idx = existingData.features.findIndex(
        (f: any) => f?.id != null && String(f.id) === featureId
      );
      if (idx >= 0) {
        const updatedFeature = geoJsonInPixelCoords.features[0];
        if (!updatedFeature.properties) updatedFeature.properties = {};
        const originalFeature = existingData.features[idx];
        updatedFeature.properties = {
          ...originalFeature.properties,
          ...updatedFeature.properties,
        };
        existingData.features[idx] = updatedFeature;
        localStorage.setItem(key, JSON.stringify(existingData));
      }
    } catch (err) {
      console.error("Error saving modified region:", err);
    }
  }

  private setRegionModifyActive(active: boolean) {
    if (this.regionModifyInteractionRef) {
      try {
        this.regionModifyInteractionRef.setActive(active);
      } catch {}
    }
    if (this.regionSnapInteractionRef) {
      try {
        this.regionSnapInteractionRef.setActive(active);
      } catch {}
    }
  }

  loadPersistedRegions() {
    if (
      !this.map ||
      !this.rawSlideMetadata ||
      !this.transformPixelToMapCoordinates
    )
      return;

    try {
      const key = `slideRegions:${this.slideUid}`;
      const raw = localStorage.getItem(key);
      if (!raw) return;

      let fc: any;
      try {
        fc = JSON.parse(raw);
      } catch {
        localStorage.removeItem(key);
        return;
      }
      if (!fc.features || fc.features.length === 0) return;

      const transformedGeoJSON = this.transformPixelToMapCoordinates(fc);
      const persisted = this.ensurePersistedRegionLayer();
      if (!persisted) return;

      const { source } = persisted;
      const format = new GeoJSON();
      const features = format.readFeatures(transformedGeoJSON);

      features.forEach((f, index) => {
        const originalFeature = transformedGeoJSON.features[index];
        if (originalFeature?.id) {
          f.setId(originalFeature.id);
        }
        source.addFeature(f);
      });
      persisted.layer.changed();

      const regions: RegionItem[] = features.map((f, index) => {
        const originalFeature = transformedGeoJSON.features[index];
        const id = originalFeature?.id
          ? String(originalFeature.id)
          : String(f.getId() ?? `${Date.now()}-${Math.random()}`);
        return {
          id,
          name: (f.get("name") as string) ?? `Region ${id}`,
          visible: (f.get("visible") as boolean) ?? true,
        };
      });

      this.onRegionsUpdate?.(regions);

      setTimeout(() => {
        this.initializeRegionModifyInteractions();
      }, 100);
    } catch (e) {
      console.error("Error loading persisted regions:", e);
    }
  }

  private setupMapEventHandlers() {
    if (!this.map) return;

    // Hover handler
    const hoverHandler = (e: any) => {
      if (this.isModifyingRegionRef) {
        this.onRegionHover?.(null);
        return;
      }

      let foundId: string | null = null;

      this.map?.forEachFeatureAtPixel(e.pixel, function (f) {
        const feature = f as Feature;
        const id = feature.getId?.();
        if (
          id != null &&
          (String(id).startsWith("region-") || String(id).startsWith("temp-"))
        ) {
          foundId = String(id);
          return true;
        }
        return false;
      });

      if (this.hoveredRegionIdRef !== foundId) {
        this.hoveredRegionIdRef = foundId;
        this.onRegionHover?.(foundId);
        const layer = this.persistedRegionLayerRef?.layer;
        if (layer) layer.changed();
      }
    };

    // Click handler
    const clickHandler = (e: any) => {
      const hit = this.map?.forEachFeatureAtPixel(
        e.pixel,
        (f) => f as Feature | undefined
      );
      const id = hit?.getId?.();
      if (!id) return;
      const idStr = String(id);

      if (idStr.startsWith("region-") || idStr.startsWith("temp-")) {
        this.selectedRegionIdRef = idStr;
        this.onRegionSelect?.(idStr);
        const layer = this.persistedRegionLayerRef?.layer;
        if (layer) layer.changed();
        this.setRegionModifyActive(true);
      }
    };

    this.map.on("pointermove", hoverHandler);
    this.map.on("singleclick", clickHandler);
  }

  private setupIdMappingListener() {
    const handleIdMapping = (event: CustomEvent) => {
      const { tempId, serverRegionId } = event.detail;
      console.log(
        "RegionDrawingManager: Mapping temp ID to server ID:",
        tempId,
        "->",
        serverRegionId
      );

      // Update the feature ID in the map
      if (this.persistedRegionLayerRef) {
        const features = this.persistedRegionLayerRef.source.getFeatures();
        const feature = features.find((f) => String(f.getId()) === tempId);
        if (feature) {
          // Update feature ID and properties
          feature.setId(serverRegionId);
          feature.set("isTemporary", false);

          // Update localStorage with new ID
          this.updateLocalStorageId(tempId, serverRegionId);

          // Update selection if this was the selected region
          if (this.selectedRegionIdRef === tempId) {
            this.selectedRegionIdRef = serverRegionId;
            this.onRegionSelect?.(serverRegionId);
          }

          // Update the UI state to replace temp ID with server ID
          const currentRegions = this.getCurrentRegions();
          const updatedRegions = currentRegions.map((region) =>
            region.id === tempId ? { ...region, id: serverRegionId } : region
          );
          this.onRegionsUpdate?.(updatedRegions);

          // Refresh the layer
          this.persistedRegionLayerRef.layer.changed();
          console.log(
            "RegionDrawingManager: Updated feature ID from",
            tempId,
            "to",
            serverRegionId
          );
        }
      }
    };

    window.addEventListener(
      "regionIdMapping",
      handleIdMapping as EventListener
    );
  }

  private updateLocalStorageId(tempId: string, serverRegionId: string) {
    try {
      const key = `slideRegions:${this.slideUid}`;
      const raw = localStorage.getItem(key);
      if (raw) {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          const feature = data.features.find(
            (f: any) => String(f.id) === tempId
          );
          if (feature) {
            // Store the original temp ID for reference
            if (feature.properties) {
              feature.properties.originalId = tempId;
              feature.properties.regionId = serverRegionId;
              feature.properties.synced = true;
            } else {
              feature.properties = {
                originalId: tempId,
                regionId: serverRegionId,
                synced: true,
              };
            }
            // Update the main ID to server ID
            feature.id = serverRegionId;
            localStorage.setItem(key, JSON.stringify(data));
            console.log(
              "RegionDrawingManager: Updated localStorage ID from",
              tempId,
              "to",
              serverRegionId
            );
          }
        }
      }
    } catch (error) {
      console.error("Error updating localStorage ID:", error);
    }
  }

  updateSelectedRegion(id: string | null) {
    this.selectedRegionIdRef = id;
    const layer = this.persistedRegionLayerRef?.layer;
    if (layer) layer.changed();

    if (this.regionTransformRef) {
      const active = id !== null && !this.isDrawingRegionRef;

      // Always clear selection first
      // @ts-ignore
      this.regionTransformRef.setSelection([]);

      this.regionTransformRef.setActive(active);

      if (active && this.persistedRegionLayerRef) {
        const features = this.persistedRegionLayerRef.source.getFeatures();
        const f =
          features.find((ff) => String(ff.getId?.() ?? "") === id) || null;
        if (f) {
          // @ts-ignore
          this.regionTransformRef.setSelection([f]);
          console.log(
            "RegionDrawingManager: Selected feature for transform:",
            id
          );
        } else {
          console.warn(
            "RegionDrawingManager: Could not find feature for selection:",
            id
          );
        }
      }
    } else {
      console.warn("RegionDrawingManager: Transform interaction not available");
    }

    this.setRegionModifyActive(id !== null && !this.isDrawingRegionRef);
  }

  updateHoveredRegion(id: string | null) {
    this.hoveredRegionIdRef = id;
    const layer = this.persistedRegionLayerRef?.layer;
    if (layer) layer.changed();
  }

  updateDrawingState(isDrawing: boolean) {
    this.isDrawingRegionRef = isDrawing;
    if (this.regionTransformRef) {
      const active = this.selectedRegionIdRef !== null && !isDrawing;
      this.regionTransformRef.setActive(active);
      if (!active) {
        // @ts-ignore
        this.regionTransformRef.setSelection([]);
      }
    }
  }

  canStartDrawing(): boolean {
    // For now, just return true. Add logic for editing states if needed
    return true;
  }

  startRegionDraw() {
    if (!this.map) return;

    if (!this.canStartDrawing()) {
      console.log("Cannot start drawing: editing in progress");
      return;
    }

    // Remove existing draw interaction
    if (this.activeRegionDrawRef) {
      try {
        this.map.removeInteraction(this.activeRegionDrawRef);
      } catch {}
      this.activeRegionDrawRef = null;
    }

    this.setRegionModifyActive(false);

    const tempLayer = this.createTemporaryRegionLayer();
    if (!tempLayer) return;

    const { source } = tempLayer;
    const draw = new Draw({
      source,
      type: "Circle",
      geometryFunction: createBox(),
    });

    draw.on("drawend", (evt) => {
      const feature = evt.feature as Feature;
      const tempId = `temp-${Date.now()}-${Math.random()
        .toString(36)
        .substring(2, 9)}`;

      // Get current regions to generate unique name
      const currentRegions = this.getCurrentRegions();

      const generateUniqueName = (
        baseName: string,
        existingRegions: RegionItem[]
      ): string => {
        const existingNames = new Set(
          existingRegions.map((r) => r.name.toLowerCase())
        );

        if (!existingNames.has(baseName.toLowerCase())) {
          return baseName;
        }

        let counter = 2;
        while (existingNames.has(`${baseName} ${counter}`.toLowerCase())) {
          counter++;
        }
        return `${baseName} ${counter}`;
      };

      const name = generateUniqueName("Region", currentRegions);

      // 1. RENDER IMMEDIATELY ON SCREEN
      feature.setId(tempId);
      feature.set("visible", true);
      feature.set("name", name);
      feature.set("kind", "box");
      feature.set("isTemporary", true); // Mark as temporary

      // Move to persisted layer immediately for visual feedback
      try {
        const persistedSource = this.persistedRegionLayerRef?.source;
        if (persistedSource) {
          persistedSource.addFeature(feature);
          source.removeFeature(feature);
        }
      } catch {}

      // Update UI state immediately
      const newRegion = { id: tempId, name, visible: true };
      const updatedRegions = [...currentRegions, newRegion];
      this.onRegionsUpdate?.(updatedRegions);

      // 2. SAVE TO LOCALSTORAGE
      this.saveRegionToLocalStorage(feature);

      // 3. SUBMIT TO API IN BACKGROUND
      setTimeout(async () => {
        if (!this.transformMapToPixelCoordinates) {
          console.error("No coordinate transform available");
          return;
        }

        try {
          const format = new GeoJSON();
          const geoJsonInMapCoords = format.writeFeatureObject(feature);
          const geoJsonInPixelCoords = this.transformMapToPixelCoordinates({
            type: "FeatureCollection",
            features: [geoJsonInMapCoords],
          });

          const regionGeometry = geoJsonInPixelCoords.features[0].geometry;

          // Create region using TanStack Query
          const regionData = {
            regionName: name,
            regionType: "roi" as const,
            geometry: regionGeometry,
            coordinateSystem: "pixel" as const,
            visible: true,
            tempId, // Include temp ID for mapping
          };

          // Use the TanStack Query createRegion function
          if (this.onCreateRegion) {
            await this.onCreateRegion(regionData);
            console.log("RegionDrawingManager: Submitted region to API:", name);
          } else {
            console.error("No createRegion callback available");
          }
        } catch (error) {
          console.error("Error submitting region to API:", error);
          // On error, keep the local region - don't remove it
          // User can manually delete if needed
        }
      }, 100);

      // Trigger selection and editing
      setTimeout(() => {
        this.onRegionSelect?.(tempId);
        const editEvent = new CustomEvent("editRegion", {
          detail: { regionId: tempId, name },
        });
        window.dispatchEvent(editEvent);
      }, 100);
    });

    this.map.addInteraction(draw);
    this.activeRegionDrawRef = draw;
    this.isDrawingRegionRef = true;
    this.onDrawingStateChange?.(true);
  }

  stopRegionDraw() {
    if (!this.map) return;
    if (this.activeRegionDrawRef) {
      try {
        this.map.removeInteraction(this.activeRegionDrawRef);
      } catch {}
      this.activeRegionDrawRef = null;
    }
    this.isDrawingRegionRef = false;
    this.onDrawingStateChange?.(false);

    this.setRegionModifyActive(this.selectedRegionIdRef !== null);
  }

  private saveRegionToLocalStorage(feature: Feature) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const format = new GeoJSON();
      const geoJsonInMapCoords = format.writeFeatureObject(feature);
      const geoJsonInPixelCoords = this.transformMapToPixelCoordinates({
        type: "FeatureCollection",
        features: [geoJsonInMapCoords],
      });

      const key = `slideRegions:${this.slideUid}`;
      const existing = localStorage.getItem(key);
      let existingData: any = { type: "FeatureCollection", features: [] };
      if (existing) {
        try {
          existingData = JSON.parse(existing);
          if (!existingData || !Array.isArray(existingData.features)) {
            existingData = { type: "FeatureCollection", features: [] };
          }
        } catch {
          existingData = { type: "FeatureCollection", features: [] };
        }
      }

      const newFeature = geoJsonInPixelCoords.features[0];
      if (!newFeature.properties) {
        newFeature.properties = {};
      }
      newFeature.properties.name = feature.get("name");
      newFeature.properties.visible = feature.get("visible");
      newFeature.properties.kind = feature.get("kind");

      existingData.features.push(newFeature);
      localStorage.setItem(key, JSON.stringify(existingData));
    } catch (e) {
      console.error("Error saving region:", e);
    }
  }

  updateRegions(regions: RegionItem[]) {
    // Update localStorage
    const key = `slideRegions:${this.slideUid}`;
    const raw = localStorage.getItem(key);
    if (raw) {
      try {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          data.features.forEach((f: any) => {
            const fid = f?.id != null ? String(f.id) : undefined;
            if (fid) {
              const region = regions.find((r) => r.id === fid);
              if (region && f.properties) {
                f.properties.name = region.name;
                f.properties.visible = region.visible;
              }
            }
          });
          localStorage.setItem(key, JSON.stringify(data));
        }
      } catch {
        // ignore parse errors
      }
    }

    // Update feature properties in map layers
    updateFeaturesById(
      this.persistedRegionLayerRef?.source,
      regions,
      (feat, region) => {
        feat.set("name", region.name);
        feat.set("visible", region.visible);
        feat.changed();
      }
    );
    updateFeaturesById(
      this.tempRegionLayerRef?.source,
      regions,
      (feat, region) => {
        feat.set("name", region.name);
        feat.set("visible", region.visible);
        feat.changed();
      }
    );

    const persistedLayer = this.persistedRegionLayerRef?.layer;
    if (persistedLayer) {
      persistedLayer.changed();
    }
  }

  removeRegions(removedIds: string[]) {
    console.log("RegionDrawingManager: Removing regions:", removedIds);

    // Update localStorage
    const key = `slideRegions:${this.slideUid}`;
    const raw = localStorage.getItem(key);
    if (raw) {
      try {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          data.features = data.features.filter((f: any) => {
            const fid = f?.id != null ? String(f.id) : undefined;
            return fid ? !removedIds.includes(fid) : true;
          });
          localStorage.setItem(key, JSON.stringify(data));
        }
      } catch {
        // ignore parse errors
      }
    }

    // Remove from map layers and force refresh
    removeFeaturesById(this.persistedRegionLayerRef?.source, removedIds);
    removeFeaturesById(this.tempRegionLayerRef?.source, removedIds);

    // Force layer refresh to update the map display
    const persistedLayer = this.persistedRegionLayerRef?.layer;
    if (persistedLayer) {
      persistedLayer.changed();
      console.log(
        "RegionDrawingManager: Forced persisted layer refresh after removal"
      );
    }

    const tempLayer = this.tempRegionLayerRef?.layer;
    if (tempLayer) {
      tempLayer.changed();
    }

    // Clear selection if the selected region was removed
    if (
      this.selectedRegionIdRef &&
      removedIds.includes(this.selectedRegionIdRef)
    ) {
      this.selectedRegionIdRef = null;
      this.onRegionSelect?.(null);

      // Deactivate transform interaction
      if (this.regionTransformRef) {
        try {
          this.regionTransformRef.setActive(false);
          // @ts-ignore
          this.regionTransformRef.setSelection([]);
        } catch {}
      }
    }
  }

  findRegionFeatureById(id: string): Feature | null {
    return findInSourcesById(
      [this.persistedRegionLayerRef?.source, this.tempRegionLayerRef?.source],
      id
    );
  }

  // Methods to access private properties for reactivation
  getRegionTransformRef() {
    return this.regionTransformRef;
  }

  getRegionModifyInteractionRef() {
    return this.regionModifyInteractionRef;
  }

  getRegionSnapInteractionRef() {
    return this.regionSnapInteractionRef;
  }

  getSelectedRegionIdRef() {
    return this.selectedRegionIdRef;
  }

  getIsDrawingRegionRef() {
    return this.isDrawingRegionRef;
  }

  // Method to update callbacks when panel reopens
  updateCallbacks(callbacks: {
    onRegionsUpdate: (regions: RegionItem[]) => void;
    onRegionSelect: (id: string | null) => void;
    onRegionHover: (id: string | null) => void;
    onDrawingStateChange: (isDrawing: boolean) => void;
    onEditingStateChange: (isEditing: boolean) => void;
    onUpdateRegionDisplaySettings: (settings: RegionSettingsPartial) => void;
    onResetRegionDisplaySettings: () => void;
    onCreateRegion: (regionData: any) => Promise<void>;
  }) {
    this.onRegionsUpdate = callbacks.onRegionsUpdate;
    this.onRegionSelect = callbacks.onRegionSelect;
    this.onRegionHover = callbacks.onRegionHover;
    this.onDrawingStateChange = callbacks.onDrawingStateChange;
    this.onEditingStateChange = callbacks.onEditingStateChange;
    this.onUpdateRegionDisplaySettings =
      callbacks.onUpdateRegionDisplaySettings;
    this.onResetRegionDisplaySettings = callbacks.onResetRegionDisplaySettings;
    this.onCreateRegion = callbacks.onCreateRegion;
  }

  // Method to get current regions from the persisted layer
  getCurrentRegions(): RegionItem[] {
    const persistedSource = this.persistedRegionLayerRef?.source;
    if (!persistedSource) return [];

    const features = persistedSource.getFeatures();
    return features.map((f) => {
      const id = String(f.getId() ?? `${Date.now()}-${Math.random()}`);
      return {
        id,
        name: (f.get("name") as string) ?? `Region ${id}`,
        visible: (f.get("visible") as boolean) ?? true,
      };
    });
  }

  // Method to reinitialize transform interaction (useful when features are loaded)
  reinitializeTransformInteraction() {
    console.log("RegionDrawingManager: Reinitializing transform interaction");
    this.initializeRegionModifyInteractions();

    // If there's a selected region, reselect it
    if (this.selectedRegionIdRef) {
      setTimeout(() => {
        this.updateSelectedRegion(this.selectedRegionIdRef);
      }, 100);
    }
  }

  // Update display settings and refresh styles
  updateDisplaySettings(settings: RegionSettingsPartial) {
    this.regionDisplaySettings = saveRegionSettings(settings);
    // Force style refresh
    const persistedLayer = this.persistedRegionLayerRef?.layer;
    if (persistedLayer) {
      persistedLayer.changed();
    }
    const tempLayer = this.tempRegionLayerRef?.layer;
    if (tempLayer) {
      tempLayer.changed();
    }
  }

  // Reset display settings to defaults
  resetDisplaySettings() {
    this.regionDisplaySettings = resetRegionSettings();
    // Force style refresh
    const persistedLayer = this.persistedRegionLayerRef?.layer;
    if (persistedLayer) {
      persistedLayer.changed();
    }
    const tempLayer = this.tempRegionLayerRef?.layer;
    if (tempLayer) {
      tempLayer.changed();
    }
  }

  // Get current display settings
  getDisplaySettings(): RegionSettings {
    return this.regionDisplaySettings;
  }

  // Load regions from GeoJSON (for API sync)
  loadRegionsFromGeoJSON(geoJsonData: any) {
    if (!this.map || !geoJsonData?.features) return;

    const persisted = this.ensurePersistedRegionLayer();
    if (!persisted) return;

    const { source } = persisted;

    // Clear existing features
    source.clear();

    const format = new GeoJSON();
    const features = format.readFeatures(geoJsonData);

    features.forEach((f) => {
      // Ensure the feature has proper properties
      if (!f.get("name")) {
        f.set("name", `Region ${f.getId()}`);
      }
      if (!f.get("visible")) {
        f.set("visible", true);
      }
      if (!f.get("kind")) {
        f.set("kind", "box");
      }
      source.addFeature(f);
    });

    persisted.layer.changed();
    console.log(
      "RegionDrawingManager: Loaded",
      features.length,
      "regions from GeoJSON"
    );
  }

  // Update feature IDs after sync (map local IDs to server UUIDs)
  updateFeatureIdsAfterSync(idMapping: { [localId: string]: string }) {
    const persisted = this.persistedRegionLayerRef;
    if (!persisted) return;

    const { source } = persisted;
    const features = source.getFeatures();

    features.forEach((feature) => {
      const currentId = String(feature.getId() || "");
      if (idMapping[currentId]) {
        const newId = idMapping[currentId];
        console.log(
          `RegionDrawingManager: Updating feature ID from ${currentId} to ${newId}`
        );

        // Create a new feature with the same geometry and properties but new ID
        const geometry = feature.getGeometry();
        const properties = feature.getProperties();

        // Remove the old feature
        source.removeFeature(feature);

        // Create new feature with server UUID
        const newFeature = new Feature(geometry);
        newFeature.setId(newId);

        // Copy all properties
        Object.keys(properties).forEach((key) => {
          if (key !== "geometry") {
            newFeature.set(key, properties[key]);
          }
        });

        // Add the new feature
        source.addFeature(newFeature);
      }
    });

    persisted.layer.changed();
    console.log("RegionDrawingManager: Updated feature IDs after sync");
  }

  cleanup() {
    if (this.map) {
      // Only deactivate Transform interaction, don't remove it
      if (this.regionTransformRef) {
        try {
          this.regionTransformRef.setActive(false);
        } catch {}
      }
      // Remove active drawing interaction
      if (this.activeRegionDrawRef) {
        try {
          this.map.removeInteraction(this.activeRegionDrawRef);
        } catch {}
      }
      // Deactivate modify and snap interactions
      if (this.regionModifyInteractionRef) {
        try {
          this.regionModifyInteractionRef.setActive(false);
        } catch {}
      }
      if (this.regionSnapInteractionRef) {
        try {
          this.regionSnapInteractionRef.setActive(false);
        } catch {}
      }
    }

    // NOTE: We do NOT remove the region layers from the map here!
    // The layers should persist even when the plugin panel is closed.
    // Only remove layers when the plugin is actually unregistered.
  }

  // New method for complete cleanup when plugin is unregistered
  cleanupCompletely() {
    if (this.map) {
      // Remove all interactions completely
      if (this.regionTransformRef) {
        try {
          this.map.removeInteraction(this.regionTransformRef);
        } catch {}
        this.regionTransformRef = null;
      }
      if (this.activeRegionDrawRef) {
        try {
          this.map.removeInteraction(this.activeRegionDrawRef);
        } catch {}
        this.activeRegionDrawRef = null;
      }
      if (this.regionModifyInteractionRef) {
        try {
          this.map.removeInteraction(this.regionModifyInteractionRef);
        } catch {}
        this.regionModifyInteractionRef = null;
      }
      if (this.regionSnapInteractionRef) {
        try {
          this.map.removeInteraction(this.regionSnapInteractionRef);
        } catch {}
        this.regionSnapInteractionRef = null;
      }

      // Remove layers from map
      if (this.tempRegionLayerRef) {
        try {
          this.map.removeLayer(this.tempRegionLayerRef.layer);
        } catch {}
        this.tempRegionLayerRef = null;
      }
      if (this.persistedRegionLayerRef) {
        try {
          this.map.removeLayer(this.persistedRegionLayerRef.layer);
        } catch {}
        this.persistedRegionLayerRef = null;
      }
    }
  }
}

// Global region manager instance to persist across panel open/close
let globalRegionManager: RegionDrawingManager | null = null;

// Function to initialize the global region manager
function initializeGlobalRegionManager(context: any) {
  if (
    globalRegionManager ||
    !context.map ||
    !context.slideUid ||
    !context.rawSlideMetadata
  ) {
    return;
  }

  // We need coordinate transforms - let's create them here
  const slideMpp = context.rawSlideMetadata?.slideMpp;
  if (!slideMpp) return;

  const transformMapToPixelCoordinates = (geoJsonData: any) => {
    if (!geoJsonData?.features || !slideMpp) {
      return geoJsonData;
    }

    const transformCoordinate = (coord: number[]): [number, number] => {
      if (coord.length < 2) return [0, 0];
      const pixelX = coord[0] / (slideMpp * 1e-6);
      const pixelY = -coord[1] / (slideMpp * 1e-6); // Note the negative for Y
      return [pixelX, pixelY];
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

    return {
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
  };

  const transformPixelToMapCoordinates = (geoJsonData: any) => {
    if (!geoJsonData?.features || !slideMpp) {
      return geoJsonData;
    }

    const transformCoordinate = (coord: number[]): [number, number] => {
      if (coord.length < 2) return [0, 0];
      const mapX = coord[0] * slideMpp * 1e-6;
      const mapY = -(coord[1] * slideMpp * 1e-6); // Note the negative for Y
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

    return {
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
  };

  const manager = new RegionDrawingManager();
  manager.initialize(
    context.map,
    context.slideUid,
    context.rawSlideMetadata,
    transformMapToPixelCoordinates,
    transformPixelToMapCoordinates,
    {
      // Dummy callbacks - these will be overridden when panel opens
      onRegionsUpdate: () => {},
      onRegionSelect: () => {},
      onRegionHover: () => {},
      onDrawingStateChange: () => {},
      onEditingStateChange: () => {},
      onUpdateRegionDisplaySettings: () => {},
      onResetRegionDisplaySettings: () => {},
      onCreateRegion: async () => {}, // Dummy callback
    }
  );

  globalRegionManager = manager;
  console.log("RegionControlPlugin: Global region manager initialized");

  // Load regions immediately when slide opens (not just when panel opens)
  // This ensures regions are visible right away
  manager.loadPersistedRegions();

  // Also trigger a background sync to ensure we have the latest data
  // Use a custom event to trigger sync when the panel opens
  setTimeout(() => {
    window.dispatchEvent(new CustomEvent("globalSync"));
  }, 1000); // Small delay to let everything initialize
}

function RegionPluginPanel({ api, onClose }: RegionPluginPanelProps) {
  const { context } = api;
  const [regions, setRegions] = useState<RegionItem[]>([]);
  const [selectedRegionId, setSelectedRegionId] = useState<string | null>(null);
  const [hoveredRegionId, setHoveredRegionId] = useState<string | null>(null);
  const [isDrawingRegion, setIsDrawingRegion] = useState(false);
  const [isEditingRegion, setIsEditingRegion] = useState(false);
  const [regionDisplaySettings, setRegionDisplaySettings] =
    useState<RegionSettings>(loadRegionSettings());

  // Region sync functionality using TanStack Query
  const {
    regions: syncedRegions,
    rawRegions,
    isLoading: isSyncLoading,
    isSyncing,
    lastSyncTime,
    error: syncError,
    createRegion,
    updateRegion,
    deleteRegion: deleteRegionFromSync,
    manualSync,
  } = useRegionSyncV2(context.slideUid || "");

  // Use global region manager instance to persist across panel open/close
  const regionManagerRef = useRef<RegionDrawingManager | null>(
    globalRegionManager
  );
  const hasLoggedManagerRef = useRef<boolean>(false);

  const { transformMapToPixelCoordinates, transformPixelToMapCoordinates } =
    useCoordinateTransforms(context.rawSlideMetadata?.slideMpp);

  // Define callbacks first
  const handleUpdateRegionDisplaySettings = useCallback(
    (settings: RegionSettingsPartial) => {
      const updated = saveRegionSettings(settings);
      setRegionDisplaySettings(updated);
      regionManagerRef.current?.updateDisplaySettings(settings);
    },
    []
  );

  const handleResetRegionDisplaySettings = useCallback(() => {
    const reset = resetRegionSettings();
    setRegionDisplaySettings(reset);
    regionManagerRef.current?.resetDisplaySettings();
  }, []);

  // Initialize region manager when map and metadata are ready
  useEffect(() => {
    // Reset logging flag when slide changes
    hasLoggedManagerRef.current = false;

    if (
      context.map &&
      context.slideUid &&
      context.rawSlideMetadata &&
      transformMapToPixelCoordinates &&
      transformPixelToMapCoordinates
    ) {
      if (globalRegionManager) {
        // Use existing global manager and update callbacks
        regionManagerRef.current = globalRegionManager;
        // Only log once per slide to avoid spam
        if (!hasLoggedManagerRef.current) {
          console.log(
            "RegionControlPlugin: Using existing global region manager"
          );
          hasLoggedManagerRef.current = true;
        }

        // Update callbacks to use current state setters
        globalRegionManager.updateCallbacks({
          onRegionsUpdate: setRegions,
          onRegionSelect: setSelectedRegionId,
          onRegionHover: setHoveredRegionId,
          onDrawingStateChange: setIsDrawingRegion,
          onEditingStateChange: setIsEditingRegion,
          onUpdateRegionDisplaySettings: handleUpdateRegionDisplaySettings,
          onResetRegionDisplaySettings: handleResetRegionDisplaySettings,
          onCreateRegion: async (regionData) => {
            await createRegion(regionData);
          },
        });

        // Regions are automatically loaded via TanStack Query
        // Just sync the TanStack Query regions with the map when available
        if (syncedRegions.length > 0) {
          console.log(
            "RegionControlPlugin: Loaded",
            syncedRegions.length,
            "regions from TanStack Query"
          );

          // Update local state
          const regionItems: RegionItem[] = syncedRegions.map((region) => ({
            id: region.id,
            name: region.name,
            visible: region.visible,
          }));
          setRegions(regionItems);

          // Load the GeoJSON features into the map from raw regions
          if (rawRegions.length > 0) {
            const featureCollection = {
              type: "FeatureCollection",
              features: rawRegions.map((region) => ({
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
              })),
            };

            // Transform and load into the map
            if (transformPixelToMapCoordinates) {
              const transformedGeoJSON =
                transformPixelToMapCoordinates(featureCollection);
              globalRegionManager?.loadRegionsFromGeoJSON(transformedGeoJSON);
              // Reinitialize transform interaction after loading regions
              setTimeout(() => {
                globalRegionManager?.reinitializeTransformInteraction();
              }, 200);
            }
          }
        } else {
          // No regions from server, check localStorage as fallback
          const existingRegions =
            globalRegionManager?.getCurrentRegions() || [];
          setRegions(existingRegions);
          // Reinitialize transform interaction for localStorage regions too
          if (existingRegions.length > 0) {
            setTimeout(() => {
              globalRegionManager?.reinitializeTransformInteraction();
            }, 200);
          }
        }

        // Reactivate interactions when panel opens
        const transformRef = globalRegionManager.getRegionTransformRef();
        if (transformRef) {
          try {
            const selectedId = globalRegionManager.getSelectedRegionIdRef();
            const active =
              selectedId !== null &&
              !globalRegionManager.getIsDrawingRegionRef();
            transformRef.setActive(active);

            // If there's a selected region, make sure it's properly selected in the transform
            if (active && selectedId) {
              const selectedFeature =
                globalRegionManager.findRegionFeatureById(selectedId);
              if (selectedFeature) {
                // @ts-ignore
                transformRef.setSelection([selectedFeature]);
              }
            }
          } catch {}
        }
        const modifyRef = globalRegionManager.getRegionModifyInteractionRef();
        if (modifyRef) {
          try {
            modifyRef.setActive(true);
          } catch {}
        }
        const snapRef = globalRegionManager.getRegionSnapInteractionRef();
        if (snapRef) {
          try {
            snapRef.setActive(true);
          } catch {}
        }
      } else if (!regionManagerRef.current) {
        // Create new manager if none exists
        const manager = new RegionDrawingManager();
        manager.initialize(
          context.map,
          context.slideUid,
          context.rawSlideMetadata,
          transformMapToPixelCoordinates,
          transformPixelToMapCoordinates,
          {
            onRegionsUpdate: setRegions,
            onRegionSelect: setSelectedRegionId,
            onRegionHover: setHoveredRegionId,
            onDrawingStateChange: setIsDrawingRegion,
            onEditingStateChange: setIsEditingRegion,
            onUpdateRegionDisplaySettings: handleUpdateRegionDisplaySettings,
            onResetRegionDisplaySettings: handleResetRegionDisplaySettings,
            onCreateRegion: async (regionData) => {
              await createRegion(regionData);
            },
          }
        );
        regionManagerRef.current = manager;
        globalRegionManager = manager; // Store in global reference
        console.log("RegionControlPlugin: Created new region manager");
      }
    }
  }, [
    context.map,
    context.slideUid,
    context.rawSlideMetadata,
    transformMapToPixelCoordinates,
    transformPixelToMapCoordinates,
    // Note: Removed sync dependencies as they cause excessive re-renders
    // The sync systems are handled separately in their own effects
  ]);

  // Note: We don't sync TanStack Query regions automatically anymore
  // Regions are managed locally first, then synced to server
  // The ID mapping events handle updating from temp IDs to server IDs

  // TanStack Query handles all sync automatically, no need for manual event listeners

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      // Only cleanup interactions when panel closes, keep layers visible
      regionManagerRef.current?.cleanup();
      // Don't clear the manager reference - keep it for persistence
    };
  }, []);

  // Update manager when selection changes
  useEffect(() => {
    if (regionManagerRef.current) {
      regionManagerRef.current.updateSelectedRegion(selectedRegionId);
    }
  }, [selectedRegionId]);

  // Update manager when hover changes
  useEffect(() => {
    if (regionManagerRef.current) {
      regionManagerRef.current.updateHoveredRegion(hoveredRegionId);
    }
  }, [hoveredRegionId]);

  // Update manager when drawing state changes
  useEffect(() => {
    if (regionManagerRef.current) {
      regionManagerRef.current.updateDrawingState(isDrawingRegion);
    }
  }, [isDrawingRegion]);

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  const handleHoverIdChange = useCallback((id: string | null) => {
    setHoveredRegionId(id);
  }, []);

  const handleUpdateRegions = useCallback(
    (updatedRegions: RegionItem[]) => {
      const prevIds = new Set(regions.map((r) => r.id));
      const nextIds = new Set(updatedRegions.map((r) => r.id));
      const removedIds: string[] = [];
      regions.forEach((r) => {
        if (!nextIds.has(r.id)) removedIds.push(r.id);
      });

      console.log(
        "RegionControlPlugin: Updating regions, removed IDs:",
        removedIds
      );

      // Clear selection if deleted
      if (
        removedIds.length &&
        selectedRegionId &&
        removedIds.includes(selectedRegionId)
      ) {
        setSelectedRegionId(null);
      }

      // Handle deletions - FIRST remove from map, THEN sync with API
      if (removedIds.length > 0) {
        // Remove from map immediately for instant visual feedback
        regionManagerRef.current?.removeRegions(removedIds);
        console.log(
          "RegionControlPlugin: Removed regions from map:",
          removedIds
        );

        // Then sync with API asynchronously using TanStack Query
        // Deletion is handled optimistically by TanStack Query
        removedIds.forEach((regionId) => {
          deleteRegionFromSync(regionId).catch((error) => {
            console.error("Error deleting region from server:", error);
            // TanStack Query will handle rollback automatically
          });
        });
      }

      // Update remaining regions
      regionManagerRef.current?.updateRegions(updatedRegions);
      setRegions(updatedRegions);

      // Trigger a manual sync to keep server updated (but don't wait for it)
      if (removedIds.length > 0) {
        // Use a shorter timeout for more responsive sync
        setTimeout(() => {
          manualSync().catch((error) => {
            console.error(
              "Error during manual sync after region update:",
              error
            );
          });
        }, 500);
      }
    },
    [regions, selectedRegionId, deleteRegionFromSync, manualSync]
  );

  const handleStartDrawRegion = useCallback(() => {
    regionManagerRef.current?.startRegionDraw();
  }, []);

  const handleStopDrawRegion = useCallback(() => {
    regionManagerRef.current?.stopRegionDraw();
  }, []);

  const canStartDrawing = useCallback(() => {
    return regionManagerRef.current?.canStartDrawing() ?? false;
  }, []);

  return (
    <RegionPanel
      isOpen={api.state.isOpen}
      onClose={onClose}
      regions={regions}
      onUpdateRegions={handleUpdateRegions}
      selectedId={selectedRegionId}
      onSelect={setSelectedRegionId}
      hoveredId={hoveredRegionId}
      onHoverIdChange={handleHoverIdChange}
      dockOverride={api.state.dock}
      onDockChange={handleDockChange}
      onStartDrawRegion={handleStartDrawRegion}
      onStopDraw={handleStopDrawRegion}
      isDrawing={isDrawingRegion}
      canStartDrawing={canStartDrawing}
      onEditingStateChange={setIsEditingRegion}
      regionDisplaySettings={regionDisplaySettings}
      onUpdateRegionDisplaySettings={handleUpdateRegionDisplaySettings}
      onResetRegionDisplaySettings={handleResetRegionDisplaySettings}
      // TODO: Add sync status props to RegionPanel interface later
    />
  );
}

export const RegionControlPlugin: ViewerPlugin = {
  id: "region-control",
  name: "Regions",
  version: "1.0.0",

  button: {
    id: "region-control-button",
    label: "Regions",
    icon: RectangleStackIcon,
    tooltip: "Create and manage regions of interest",
    position: "right",
    order: 4,
  },

  panel: {
    id: "region-control-panel",
    title: "Regions",
    defaultSize: { width: 320, height: 400 },
    defaultDock: "left",
    storageKey: "regionPanel",
  },

  PanelComponent: RegionPluginPanel,

  initialize: (api) => {
    console.log("RegionControlPlugin: Initializing plugin");
    // Plugin will initialize when context is ready
  },

  onButtonClick: (api: PluginAPI) => {
    console.log("RegionControlPlugin: Button clicked");
    console.log("RegionControlPlugin: API context:", api.context);
    console.log("RegionControlPlugin: API state:", api.state);

    // Toggle panel
    console.log("RegionControlPlugin: Toggling panel state");
    api.setState({ isOpen: !api.state.isOpen });
  },

  onContextChange: (context) => {
    console.log("RegionControlPlugin: Context changed", context);

    // Initialize region manager when map and metadata are available
    if (
      context.map &&
      context.slideUid &&
      context.rawSlideMetadata &&
      !globalRegionManager
    ) {
      console.log(
        "RegionControlPlugin: Initializing region manager from context change"
      );
      initializeGlobalRegionManager(context);
    }
  },

  destroy: () => {
    // Complete cleanup when plugin is unregistered
    if (globalRegionManager) {
      globalRegionManager.cleanupCompletely();
      globalRegionManager = null;
    }
  },
};
