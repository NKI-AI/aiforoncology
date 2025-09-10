// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect, useState, useRef } from "react";
import { PencilSquareIcon } from "@heroicons/react/24/outline";
import { ViewerPlugin, PluginAPI } from "./types";
import AnnotationPanel, {
  type AnnotationItem,
  type LabelId,
} from "@/features/viewer/components/AnnotationPanel";

import { useStudyAnnotationSettings } from "@/hooks/useStudyAnnotationSettings";
import { useAnnotationSettings } from "@/features/viewer/hooks/useAnnotationSettings";
import { useAnnotationSync } from "@/features/viewer/hooks/useAnnotationSync";
import { useCoordinateTransforms } from "@/features/viewer/components/SlideWorkspace/hooks/useCoordinateTransforms";
import { BrushPolygonTool } from "@/features/viewer/tools/BrushPolygonTool";
import DragPan from "ol/interaction/DragPan";
import type OlMap from "ol/Map";
import { primaryAction } from "ol/events/condition";
import VectorSource from "ol/source/Vector";
import VectorLayer from "ol/layer/Vector";
import Draw, { createBox } from "ol/interaction/Draw";
import Modify from "ol/interaction/Modify";
import Snap from "ol/interaction/Snap";
import Feature from "ol/Feature";
import Polygon from "ol/geom/Polygon";
import MultiPolygon from "ol/geom/MultiPolygon";
import Point from "ol/geom/Point";
import { Style, Stroke, Fill, Circle as CircleStyle } from "ol/style";
import GeoJSON from "ol/format/GeoJSON";
import MapBrowserEventType from "ol/MapBrowserEventType";

// JSTS imports for polygon merge operations
// @ts-ignore: jsts ESM packages don't ship TS types
import OL3Parser from "jsts/org/locationtech/jts/io/OL3Parser.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import OverlayOp from "jsts/org/locationtech/jts/operation/overlay/OverlayOp.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import GeometryFactory from "jsts/org/locationtech/jts/geom/GeometryFactory.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import BufferOp from "jsts/org/locationtech/jts/operation/buffer/BufferOp.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import IsValidOp from "jsts/org/locationtech/jts/operation/valid/IsValidOp.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import DouglasPeuckerSimplifier from "jsts/org/locationtech/jts/simplify/DouglasPeuckerSimplifier.js";
// Import LinearRing and other geometries for JSTS injection
import LineString from "ol/geom/LineString";
import LinearRing from "ol/geom/LinearRing";
import MultiLineString from "ol/geom/MultiLineString";
import MultiPoint from "ol/geom/MultiPoint";

interface AnnotationPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
}

/**
 * Manages all annotation drawing tools and interactions
 */
class AnnotationDrawingManager {
  private map: OlMap | null = null;
  private slideUid: string = "";
  private rawSlideMetadata: any = null;
  private transformMapToPixelCoordinates: ((geoJson: any) => any) | null = null;
  private transformPixelToMapCoordinates: ((geoJson: any) => any) | null = null;

  // Layer management
  private tempRoiLayerRef: { source: VectorSource; layer: VectorLayer } | null =
    null;
  private persistedRoiLayerRef: {
    source: VectorSource;
    layer: VectorLayer;
  } | null = null;

  // Interaction management
  private activeDrawRef: Draw | null = null;
  private modifyInteractionRef: Modify | null = null;
  private snapInteractionRef: Snap | null = null;
  private dragPanInteractionRef: DragPan | null = null;
  private seedPointerDownHandlerRef: ((e: any) => void) | null = null;

  // Brush tool
  private brushRef: BrushPolygonTool | null = null;

  // JSTS parser for polygon operations
  private jstsParser: any = null;

  // Preview layer for geometry operations
  private previewLayerRef: { source: VectorSource; layer: VectorLayer } | null =
    null;

  // State
  private roiIdCounterRef = 1;
  private isModifyingRef = false;
  private selectedFeatureId: string | null = null;
  private hoveredFeatureId: string | null = null;
  private activeAnnotationLabel: LabelId | null = null;

  // Callbacks
  private onAnnotationsUpdate:
    | ((annotations: AnnotationItem[]) => void)
    | null = null;
  private onAnnotationSelect: ((id: string | null) => void) | null = null;
  private onAnnotationHover: ((id: string | null) => void) | null = null;
  private annotationDisplaySettingsRef: any = null;
  private getLabelColors:
    | (() => { labelColors: Record<LabelId, string> })
    | null = null;
  private micrometersToPixels: ((micrometers: number) => number) | null = null;
  private annotationSettings: any = null;
  private annotationSettingsLoading = false;
  private getCurrentAnnotations: (() => AnnotationItem[]) | null = null;
  private slideMpp: number | null = null;

  initialize(
    map: OlMap,
    slideUid: string,
    rawSlideMetadata: any,
    transformMapToPixelCoordinates: (geoJson: any) => any,
    transformPixelToMapCoordinates: (geoJson: any) => any,
    callbacks: {
      onAnnotationsUpdate: (annotations: AnnotationItem[]) => void;
      onAnnotationSelect: (id: string | null) => void;
      onAnnotationHover: (id: string | null) => void;
      getLabelColors: () => { labelColors: Record<LabelId, string> };
      micrometersToPixels: (micrometers: number) => number;
      getCurrentAnnotations: () => AnnotationItem[];
    },
    annotationDisplaySettingsRef: any,
    annotationSettings: any,
    annotationSettingsLoading: boolean
  ) {
    this.map = map;
    this.slideUid = slideUid;
    this.rawSlideMetadata = rawSlideMetadata;
    this.transformMapToPixelCoordinates = transformMapToPixelCoordinates;
    this.transformPixelToMapCoordinates = transformPixelToMapCoordinates;
    this.onAnnotationsUpdate = callbacks.onAnnotationsUpdate;
    this.onAnnotationSelect = callbacks.onAnnotationSelect;
    this.onAnnotationHover = callbacks.onAnnotationHover;
    this.getLabelColors = callbacks.getLabelColors;
    this.micrometersToPixels = callbacks.micrometersToPixels;
    this.getCurrentAnnotations = callbacks.getCurrentAnnotations;
    this.annotationDisplaySettingsRef = annotationDisplaySettingsRef;
    this.annotationSettings = annotationSettings;
    this.annotationSettingsLoading = annotationSettingsLoading;
    this.slideMpp = rawSlideMetadata?.slideMpp || null;

    // Find drag pan interaction
    const drag =
      map
        .getInteractions()
        .getArray()
        .find((i): i is DragPan => i instanceof DragPan) || null;
    this.dragPanInteractionRef = drag;

    // Initialize JSTS parser
    this.initializeJSTSParser();

    // Initialize layers
    this.ensurePersistedRoiLayer();
    this.initializeModifyInteractions();
    this.initializeBrushTool();
    this.initializeClickHandling();
    this.initializeHoverHandling();
  }

  private initializeJSTSParser() {
    try {
      // Create geometry factory for JSTS operations
      const geometryFactory = new GeometryFactory();
      this.jstsParser = new OL3Parser(geometryFactory, undefined);
      this.jstsParser.inject(
        Point,
        LineString,
        LinearRing,
        Polygon,
        MultiPoint,
        MultiLineString,
        MultiPolygon
      );
    } catch (error) {
      console.error("Failed to initialize JSTS parser:", error);
      this.jstsParser = null;
    }
  }

  private createPreviewLayer() {
    if (!this.map) return null;
    if (this.previewLayerRef) return this.previewLayerRef;

    const source = new VectorSource({ wrapX: false });
    const layer = new VectorLayer({
      source,
      style: (feature) => {
        const previewType = feature.get("previewType");

        // Different styles for different preview types
        if (previewType === "merge") {
          return new Style({
            stroke: new Stroke({
              color: "#10b981",
              width: 3,
              lineDash: [5, 5],
            }),
            fill: new Fill({ color: "#10b98133" }),
          });
        } else if (previewType === "simplify") {
          return new Style({
            stroke: new Stroke({
              color: "#f59e0b",
              width: 3,
              lineDash: [3, 3],
            }),
            fill: new Fill({ color: "#f59e0b33" }),
          });
        }

        // Default preview style
        return new Style({
          stroke: new Stroke({ color: "#6366f1", width: 3, lineDash: [4, 4] }),
          fill: new Fill({ color: "#6366f133" }),
        });
      },
    });

    layer.setZIndex(2500); // Above annotations but below UI
    this.map.addLayer(layer);

    this.previewLayerRef = { source, layer };
    return this.previewLayerRef;
  }

  clearPreview() {
    if (this.previewLayerRef) {
      this.previewLayerRef.source.clear();
    }
  }

  private showGeometryPreview(
    geometry: Polygon | MultiPolygon,
    previewType: "merge" | "simplify" = "simplify"
  ) {
    const previewLayer = this.createPreviewLayer();
    if (!previewLayer) return;

    // Clear existing preview
    this.clearPreview();

    // Create preview feature
    const previewFeature = new Feature({
      geometry: geometry,
      previewType: previewType,
    });

    previewLayer.source.addFeature(previewFeature);
  }

  private countPolygonPoints(geometry: Polygon | MultiPolygon): number {
    let totalPoints = 0;

    if (geometry instanceof Polygon) {
      // Count points in exterior ring
      const exteriorRing = geometry.getLinearRing(0);
      if (exteriorRing) {
        totalPoints += exteriorRing.getCoordinates().length - 1; // Subtract 1 because first and last points are the same
      }

      // Count points in interior rings (holes)
      const interiorRingCount = geometry.getLinearRingCount() - 1;
      for (let i = 1; i <= interiorRingCount; i++) {
        const interiorRing = geometry.getLinearRing(i);
        if (interiorRing) {
          totalPoints += interiorRing.getCoordinates().length - 1;
        }
      }
    } else if (geometry instanceof MultiPolygon) {
      // Count points in all polygons
      const polygons = geometry.getPolygons();
      polygons.forEach((polygon) => {
        totalPoints += this.countPolygonPoints(polygon);
      });
    }

    return totalPoints;
  }

  private createTemporaryRoiLayer() {
    if (!this.map) return null;
    if (this.tempRoiLayerRef) return this.tempRoiLayerRef;

    const source = new VectorSource({ wrapX: false });
    const layer = new VectorLayer({
      source,
      style: (feature) => {
        const label = (feature.get("name") as LabelId) ?? "roi";

        // Don't render if annotation settings haven't loaded yet
        if (this.annotationSettingsLoading || !this.annotationSettings) {
          return new Style({
            stroke: new Stroke({ color: "#9ca3af", width: 2 }),
            fill: new Fill({ color: "#9ca3af33" }),
          });
        }

        const { labelColors } = this.getLabelColors?.() || { labelColors: {} };
        const base = labelColors[label];
        if (!base) {
          console.error(
            `No color found for annotation label "${label}". Available colors:`,
            labelColors
          );
          throw new Error(
            `Missing color configuration for annotation label "${label}". Please check study annotation settings.`
          );
        }
        const settings = this.annotationDisplaySettingsRef?.current;
        if (!settings) return new Style();

        const stroke = new Stroke({ color: base, width: settings.strokeWidth });
        const fillOpacityHex = Math.round(settings.fillOpacity * 255)
          .toString(16)
          .padStart(2, "0");
        const fill = new Fill({ color: `${base}${fillOpacityHex}` });

        const geom = feature.getGeometry();
        if (geom instanceof Point) {
          const pointRadius =
            (this.micrometersToPixels?.(settings.pointSizeMicrometers) || 10) /
            2;
          return new Style({
            image: new CircleStyle({
              radius: pointRadius,
              stroke: new Stroke({
                color: base,
                width: Math.max(1, settings.strokeWidth - 1),
              }),
              fill: new Fill({ color: base }),
            }),
          });
        }
        return new Style({ stroke, fill });
      },
    });

    layer.setZIndex(2000);
    this.map.addLayer(layer);

    this.tempRoiLayerRef = { source, layer };
    return this.tempRoiLayerRef;
  }

  private ensurePersistedRoiLayer() {
    if (!this.map) return null;
    if (this.persistedRoiLayerRef) return this.persistedRoiLayerRef;

    const source = new VectorSource({ wrapX: false });
    const layer = new VectorLayer({
      source,
      style: (feature) => {
        const id = String(feature.getId() ?? "");
        const label = (feature.get("name") as LabelId) ?? "roi";

        // Don't render if annotation settings haven't loaded yet
        if (this.annotationSettingsLoading || !this.annotationSettings) {
          return new Style({
            stroke: new Stroke({ color: "#9ca3af", width: 2 }),
            fill: new Fill({ color: "#9ca3af33" }),
          });
        }

        const { labelColors } = this.getLabelColors?.() || { labelColors: {} };
        const base = labelColors[label];
        if (!base) {
          console.error(
            `No color found for annotation label "${label}". Available colors:`,
            labelColors
          );
          throw new Error(
            `Missing color configuration for annotation label "${label}". Please check study annotation settings.`
          );
        }

        const settings = this.annotationDisplaySettingsRef?.current;
        if (!settings) return new Style();

        // Check if this feature is selected or hovered
        const isSelected = this.selectedFeatureId === id;
        const isHovered = this.hoveredFeatureId === id;
        const width = isSelected
          ? settings.strokeWidth + 2
          : isHovered
          ? settings.strokeWidth + 1
          : settings.strokeWidth;
        const strokeColor = isSelected
          ? "#0ea5e9"
          : isHovered
          ? "#f59e0b"
          : base;
        const fillOpacityHex = Math.round(settings.fillOpacity * 255)
          .toString(16)
          .padStart(2, "0");
        const selectedFillOpacityHex = Math.round(
          Math.min(1, settings.fillOpacity + 0.15) * 255
        )
          .toString(16)
          .padStart(2, "0");
        const hoveredFillOpacityHex = Math.round(
          Math.min(1, settings.fillOpacity + 0.1) * 255
        )
          .toString(16)
          .padStart(2, "0");
        const fillColor = isSelected
          ? `#0ea5e9${selectedFillOpacityHex}`
          : isHovered
          ? `#f59e0b${hoveredFillOpacityHex}`
          : `${base}${fillOpacityHex}`;
        const stroke = new Stroke({ color: strokeColor, width });

        const geom = feature.getGeometry();
        if (geom instanceof Point) {
          const pointRadius =
            (this.micrometersToPixels?.(settings.pointSizeMicrometers) || 10) /
            2;
          return new Style({
            image: new CircleStyle({
              radius: isSelected
                ? pointRadius + 2
                : isHovered
                ? pointRadius + 1
                : pointRadius,
              stroke: new Stroke({
                color: strokeColor,
                width: isSelected
                  ? Math.max(2, settings.strokeWidth)
                  : isHovered
                  ? Math.max(1, settings.strokeWidth)
                  : Math.max(1, settings.strokeWidth - 1),
              }),
              fill: new Fill({ color: fillColor }),
            }),
          });
        }
        return new Style({
          stroke,
          fill: new Fill({ color: fillColor }),
        });
      },
    });
    layer.setZIndex(1500);
    this.map.addLayer(layer);
    this.persistedRoiLayerRef = { source, layer };
    return this.persistedRoiLayerRef;
  }

  private initializeModifyInteractions() {
    if (!this.map) return;

    // Clean up existing interactions
    if (this.modifyInteractionRef) {
      try {
        this.map.removeInteraction(this.modifyInteractionRef);
      } catch {}
      this.modifyInteractionRef = null;
    }
    if (this.snapInteractionRef) {
      try {
        this.map.removeInteraction(this.snapInteractionRef);
      } catch {}
      this.snapInteractionRef = null;
    }

    const persisted = this.persistedRoiLayerRef;
    if (!persisted) return;

    const modify = new Modify({
      source: persisted.source,
      pixelTolerance: 8,
      condition: (evt) => primaryAction(evt) && !this.activeDrawRef,
    });

    const snap = new Snap({ source: persisted.source });

    modify.on("modifystart", () => {
      this.isModifyingRef = true;
      if (this.dragPanInteractionRef) {
        try {
          this.dragPanInteractionRef.setActive(false);
        } catch {}
      }
    });

    modify.on("modifyend", (evt) => {
      this.isModifyingRef = false;

      if (this.dragPanInteractionRef) {
        try {
          this.dragPanInteractionRef.setActive(true);
        } catch {}
      }

      // Persist geometry changes back to localStorage
      this.persistModifiedFeatures(evt.features.getArray() as Feature[]);
    });

    this.map.addInteraction(modify);
    this.map.addInteraction(snap);
    this.modifyInteractionRef = modify;
    this.snapInteractionRef = snap;
  }

  private persistModifiedFeatures(features: Feature[]) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const format = new GeoJSON();
      if (features.length === 0) return;

      const key = `slideAnnotations:${this.slideUid}`;
      const raw = localStorage.getItem(key);
      let store: any = raw
        ? JSON.parse(raw)
        : { type: "FeatureCollection", features: [] };
      if (!Array.isArray(store.features)) store.features = [];

      features.forEach((feat) => {
        const geoJsonInMap = format.writeFeatureObject(feat);
        const geoJsonInPixel = this.transformMapToPixelCoordinates!({
          type: "FeatureCollection",
          features: [geoJsonInMap],
        });
        const updated = geoJsonInPixel.features[0];

        const fid = String(feat.getId());
        const idx = store.features.findIndex(
          (f: any) => f?.id != null && String(f.id) === fid
        );
        if (idx >= 0) {
          const existingProps = store.features[idx].properties || {};
          const newProps = updated.properties || {};

          updated.properties = {
            ...existingProps,
            ...newProps,
            lastModified: Date.now(),
          };

          // Mark as modified and needing sync by removing vectorUid
          // This will make the sync system treat it as a local annotation
          delete updated.properties.vectorUid;
          store.features[idx] = updated;
        }
      });

      localStorage.setItem(key, JSON.stringify(store));

      // Log the modification for debugging
      console.log(
        `Marked ${features.length} modified annotations for sync to server`
      );
    } catch (e) {
      console.error("Error persisting modified features:", e);
    }
  }

  private initializeBrushTool() {
    if (!this.map) return;

    this.brushRef = new BrushPolygonTool(this.map, {
      previewZIndex: 2300,
      setDragPanActive: (active) => {
        try {
          this.dragPanInteractionRef?.setActive(active);
        } catch {}
      },
      onCommit: (feat) => {
        // Check if the geometry is empty after erase operation
        const geom = feat.getGeometry();
        if (
          geom &&
          geom instanceof MultiPolygon &&
          geom.getPolygons().length === 0
        ) {
          this.removeEmptyFeature(feat);
        } else {
          this.persistFeatureGeometry(feat);
        }
        this.persistedRoiLayerRef?.layer?.changed();
      },
      onCreateFromStroke: (strokePoly: Polygon) => {
        this.handleBrushCreateFromStroke(strokePoly);
      },
    });
  }

  private initializeClickHandling() {
    if (!this.map) return;

    this.map.on("singleclick", (evt) => {
      // Don't handle clicks when drawing or modifying
      if (this.activeDrawRef || this.isModifyingRef) {
        return;
      }

      // Check if we clicked on a feature
      const features = this.map!.getFeaturesAtPixel(evt.pixel, {
        layerFilter: (layer) => layer === this.persistedRoiLayerRef?.layer,
      });

      if (features && features.length > 0) {
        const feature = features[0] as Feature;
        const featureId = String(feature.getId());

        // Toggle selection - if already selected, deselect
        if (this.selectedFeatureId === featureId) {
          this.onAnnotationSelect?.(null);
        } else {
          this.onAnnotationSelect?.(featureId);
        }
      } else {
        // Clicked on empty space, deselect
        this.onAnnotationSelect?.(null);
      }
    });
  }

  private initializeHoverHandling() {
    if (!this.map) return;

    this.map.on("pointermove", (evt) => {
      // Don't handle hover when drawing or modifying
      if (this.activeDrawRef || this.isModifyingRef) {
        if (this.hoveredFeatureId) {
          this.hoveredFeatureId = null;
          this.onAnnotationHover?.(null);
          this.refreshLayers();
        }
        return;
      }

      let foundId: string | null = null;

      // Check if we're hovering over an annotation feature
      const features = this.map!.getFeaturesAtPixel(evt.pixel, {
        layerFilter: (layer) => layer === this.persistedRoiLayerRef?.layer,
      });

      if (features && features.length > 0) {
        const feature = features[0] as Feature;
        const featureId = String(feature.getId());
        if (featureId.startsWith("roi-")) {
          foundId = featureId;
        }
      }

      // Update hover state if changed
      if (this.hoveredFeatureId !== foundId) {
        this.hoveredFeatureId = foundId;
        this.onAnnotationHover?.(foundId);
        this.refreshLayers();
      }
    });
  }

  private removeEmptyFeature(feat: Feature) {
    const fid = String(feat.getId());

    // Remove from annotations state
    if (this.onAnnotationsUpdate) {
      // This would need to be handled differently - we need access to current annotations
      // For now, we'll emit an event or use a callback
    }

    // Remove from layers
    const persistedSource = this.persistedRoiLayerRef?.source;
    if (persistedSource) {
      const feature = persistedSource.getFeatureById(fid);
      if (feature) persistedSource.removeFeature(feature);
    }

    // Remove from localStorage
    const key = `slideAnnotations:${this.slideUid}`;
    const raw = localStorage.getItem(key);
    if (raw) {
      try {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          data.features = data.features.filter((f: any) => {
            const id = f?.id != null ? String(f.id) : undefined;
            return id !== fid;
          });
          localStorage.setItem(key, JSON.stringify(data));
        }
      } catch {}
    }
  }

  private persistFeatureGeometry(feat: Feature) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const format = new GeoJSON();
      const geoJsonInMap = format.writeFeatureObject(feat);
      const geoJsonInPixel = this.transformMapToPixelCoordinates({
        type: "FeatureCollection",
        features: [geoJsonInMap],
      });
      const updated = geoJsonInPixel.features[0];

      const key = `slideAnnotations:${this.slideUid}`;
      const raw = localStorage.getItem(key);
      let store: any = raw
        ? JSON.parse(raw)
        : { type: "FeatureCollection", features: [] };
      if (!Array.isArray(store.features)) store.features = [];

      const fid = String(feat.getId());
      const idx = store.features.findIndex(
        (f: any) => f?.id != null && String(f.id) === fid
      );
      if (idx >= 0) {
        const existingProps = store.features[idx].properties || {};
        const newProps = updated.properties || {};

        updated.properties = {
          ...existingProps,
          ...newProps,
          lastModified: Date.now(),
        };

        // Mark as modified and needing sync by removing vectorUid
        // This will make the sync system treat it as a local annotation
        delete updated.properties.vectorUid;
        store.features[idx] = updated;
      }
      localStorage.setItem(key, JSON.stringify(store));
    } catch (e) {
      console.error("Brush persist failed:", e);
    }
  }

  private handleBrushCreateFromStroke(strokePoly: Polygon) {
    // Create a new annotation from brush stroke
    if (!this.map || !strokePoly) return;

    // Use active label or default to "roi" if none is selected
    const label = this.activeAnnotationLabel || "roi";

    const idNum = this.roiIdCounterRef++;
    const id = `roi-${idNum}`;

    // Create new feature with the stroke polygon geometry
    const feature = new Feature({
      geometry: strokePoly,
    });

    // Set feature properties
    feature.setId(id);
    feature.set("visible", true);
    feature.set("name", label);
    feature.set("kind", "polygon");

    // Add color information from study settings
    const { labelColors } = this.getLabelColors?.() || { labelColors: {} };
    const color = labelColors[label];
    if (color) {
      feature.set("color", color);
    }

    // Add to persisted layer
    const persistedSource = this.persistedRoiLayerRef?.source;
    if (persistedSource) {
      persistedSource.addFeature(feature);
    }

    // Save to localStorage
    this.saveFeatureToStorage(feature);

    // Immediately add the new annotation to the list
    if (this.onAnnotationsUpdate && this.getCurrentAnnotations) {
      const newAnnotationItem = {
        id: id,
        name: label,
        visible: true,
        kind: "polygon" as const,
      };

      const currentAnnotations = this.getCurrentAnnotations();
      const updatedAnnotations = [...currentAnnotations, newAnnotationItem];
      this.onAnnotationsUpdate(updatedAnnotations);
    }

    console.log(`Created new annotation "${label}" (${id}) from brush stroke`);
  }

  startRoiDraw(mode: "point" | "box" | "polygon", label: LabelId) {
    if (!this.map) return;

    // Implementation moved from SlideWorkspace
    this.stopRoiDraw(); // Clean up any existing draw

    // Create temporary layer for immediate visual feedback
    const tempLayer = this.createTemporaryRoiLayer();
    if (!tempLayer) return;

    const { source } = tempLayer;

    const options: any = { source };
    if (mode === "box") {
      options.type = "Circle";
      options.geometryFunction = createBox();
    } else if (mode === "point") {
      options.type = "Point";
    } else {
      options.type = "Polygon";
      options.freehand = false;
      options.freehandCondition = this.freehandOnDrag;
      options.dragVertexDelay = 0;
    }

    const draw = new Draw(options);

    // Handle polygon seeding for smooth freehand drawing
    if (mode === "polygon") {
      this.setupPolygonSeeding(draw);
    }

    draw.on("drawend", (evt) => {
      this.handleDrawEnd(evt.feature as Feature, mode, label, source);
    });

    this.map.addInteraction(draw);
    this.activeDrawRef = draw;

    // Disable modify while drawing
    this.setModifyActive(false);
  }

  private freehandOnDrag = (evt: any) => {
    const oe = evt.originalEvent as PointerEvent | MouseEvent;
    const primaryDown =
      "buttons" in oe
        ? ((oe as PointerEvent).buttons & 1) !== 0
        : "which" in oe
        ? (oe as any).which === 1
        : (oe as any).button === 0;
    return evt.type === MapBrowserEventType.POINTERDRAG && primaryDown;
  };

  private setupPolygonSeeding(draw: Draw) {
    if (!this.map) return;

    let hasSketch = false;

    draw.on("drawstart", () => {
      hasSketch = true;
    });
    draw.on("drawend", () => {
      hasSketch = false;
    });
    draw.on("drawabort", () => {
      hasSketch = false;
    });

    const seedOnDown = (e: any) => {
      if (this.activeDrawRef !== draw) return;
      if (hasSketch) return;

      const oe = e.originalEvent as PointerEvent | MouseEvent;
      const primaryDown =
        "buttons" in oe
          ? ((oe as PointerEvent).buttons & 1) !== 0
          : "which" in oe
          ? (oe as any).which === 1
          : (oe as any).button === 0;
      if (!primaryDown) return;

      try {
        draw.appendCoordinates([e.coordinate, e.coordinate]);
      } catch {}
    };

    this.seedPointerDownHandlerRef = seedOnDown;
    this.map.on("pointerdown" as any, seedOnDown as any);
  }

  private handleDrawEnd(
    feature: Feature,
    mode: "point" | "box" | "polygon",
    label: LabelId,
    source: VectorSource
  ) {
    const idNum = this.roiIdCounterRef++;
    const id = `roi-${idNum}`;

    // Set feature properties
    feature.setId(id);
    feature.set("visible", true);
    feature.set("name", label); // Just set the label as the name
    feature.set("kind", mode);

    // Add color information from study settings
    const { labelColors } = this.getLabelColors?.() || { labelColors: {} };
    const color = labelColors[label];
    if (color) {
      feature.set("color", color);
    }

    // Make sure we don't set the old label property
    feature.unset("label");

    // Save to localStorage
    this.saveFeatureToStorage(feature);

    // Move feature from temp layer to persisted layer
    try {
      const persistedSource = this.persistedRoiLayerRef?.source;
      if (persistedSource) {
        persistedSource.addFeature(feature);
        source.removeFeature(feature);
      }
    } catch {}

    // Immediately add the new annotation to the list
    if (this.onAnnotationsUpdate && this.getCurrentAnnotations) {
      // Create the annotation item from the feature
      const newAnnotationItem = {
        id: id,
        name: feature.get("name") || label,
        visible: feature.get("visible") !== false,
        kind: mode,
      };

      // Get current annotations and add the new one
      const currentAnnotations = this.getCurrentAnnotations();
      const updatedAnnotations = [...currentAnnotations, newAnnotationItem];

      // Update the annotation list immediately
      this.onAnnotationsUpdate(updatedAnnotations);
    }
  }

  private saveFeatureToStorage(feature: Feature) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const format = new GeoJSON();
      const geoJsonInMapCoords = format.writeFeatureObject(feature);

      const geoJsonInPixelCoords = this.transformMapToPixelCoordinates({
        type: "FeatureCollection",
        features: [geoJsonInMapCoords],
      });

      const key = `slideAnnotations:${this.slideUid}`;
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
      const updated = geoJsonInPixelCoords.features[0];

      // Check if this feature already exists in storage
      const existingIdx = existingData.features.findIndex(
        (f: any) => f?.id != null && String(f.id) === featureId
      );

      if (existingIdx >= 0) {
        // Update existing feature and mark as modified for sync
        const existingProps =
          existingData.features[existingIdx].properties || {};
        const newProps = updated.properties || {};

        updated.properties = {
          ...existingProps,
          ...newProps,
          lastModified: Date.now(),
        };

        // Mark as modified and needing sync by removing vectorUid
        delete updated.properties.vectorUid;
        existingData.features[existingIdx] = updated;
        console.log(
          `Updated existing annotation ${featureId} and marked for sync`
        );
      } else {
        // Add new feature
        existingData.features.push(updated);
        console.log(`Added new annotation ${featureId}`);
      }

      localStorage.setItem(key, JSON.stringify(existingData));

      // Immediately add the annotation to the list if it's a new one
      if (this.onAnnotationsUpdate && this.getCurrentAnnotations) {
        const featureId = String(feature.getId());
        const currentAnnotations = this.getCurrentAnnotations();

        // Check if this annotation is already in the list
        const existingAnnotation = currentAnnotations.find(
          (ann) => ann.id === featureId
        );

        if (!existingAnnotation) {
          // This is a new annotation (likely from brush tool), add it to the list
          // Add color information for brush-created features
          const labelName = feature.get("name") || "roi";
          const { labelColors } = this.getLabelColors?.() || {
            labelColors: {},
          };
          const color = labelColors[labelName];
          if (color) {
            feature.set("color", color);
          }

          const newAnnotationItem = {
            id: featureId,
            name: labelName,
            visible: feature.get("visible") !== false,
            kind: feature.get("kind") || "polygon",
          };

          const updatedAnnotations = [...currentAnnotations, newAnnotationItem];
          this.onAnnotationsUpdate(updatedAnnotations);
        }
      }
    } catch (e) {
      console.error("Error saving annotation:", e);
    }
  }

  stopRoiDraw() {
    if (!this.map) return;

    if (this.activeDrawRef) {
      try {
        this.map.removeInteraction(this.activeDrawRef);
      } catch {}
      this.activeDrawRef = null;
    }

    // Unbind polygon seeding handler if present
    if (this.seedPointerDownHandlerRef) {
      try {
        this.map.un(
          "pointerdown" as any,
          this.seedPointerDownHandlerRef as any
        );
      } catch {}
      this.seedPointerDownHandlerRef = null;
    }

    // Re-enable modify after drawing stops
    this.setModifyActive(true);
  }

  private setModifyActive(active: boolean) {
    if (this.modifyInteractionRef) {
      try {
        this.modifyInteractionRef.setActive(active);
      } catch {}
    }
    if (this.snapInteractionRef) {
      try {
        this.snapInteractionRef.setActive(active);
      } catch {}
    }
  }

  // Brush tool methods
  setBrushActive(active: boolean) {
    this.brushRef?.setActive(active);

    if (active) {
      this.stopRoiDraw();
      this.setModifyActive(false);
    } else {
      this.setModifyActive(true);
    }
  }

  setBrushMode(mode: "add" | "erase") {
    this.brushRef?.setMode(mode);
  }

  setBrushSize(sizePx: number) {
    this.brushRef?.setBrushSizePixels(sizePx);
  }

  setSelectedFeature(feature: Feature | null) {
    if (
      feature?.getGeometry() instanceof Polygon ||
      feature?.getGeometry() instanceof MultiPolygon
    ) {
      this.brushRef?.setTargetFeature(feature as any);
    } else {
      this.brushRef?.setTargetFeature(null);
    }
  }

  loadAnnotations(geoJsonData: any) {
    if (!this.transformPixelToMapCoordinates) return;

    try {
      const transformedGeoJSON =
        this.transformPixelToMapCoordinates(geoJsonData);
      const persisted = this.ensurePersistedRoiLayer();
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

      // Update ROI counter
      const maxIdNum = features.reduce((max, f) => {
        const id = String(f.getId() || "");
        if (id.startsWith("roi-")) {
          const num = parseInt(id.replace("roi-", ""), 10);
          return isNaN(num) ? max : Math.max(max, num);
        }
        return max;
      }, 0);
      this.roiIdCounterRef = maxIdNum + 1;
    } catch (e) {
      console.error("Error loading annotations:", e);
    }
  }

  refreshLayers() {
    this.persistedRoiLayerRef?.layer?.changed();
    this.tempRoiLayerRef?.layer?.changed();
  }

  getFeatureById(id: string): Feature | null {
    return this.persistedRoiLayerRef?.source.getFeatureById(id) || null;
  }

  // Set the selected feature for highlighting
  setSelectedAnnotationId(id: string | null) {
    this.selectedFeatureId = id;
    this.refreshLayers();
  }

  // Set the active annotation label for brush creation
  setActiveAnnotationLabel(label: LabelId | null) {
    this.activeAnnotationLabel = label;
  }

  // Set the hovered annotation for highlighting (called from list hover)
  setHoveredAnnotationId(id: string | null) {
    if (this.hoveredFeatureId !== id) {
      this.hoveredFeatureId = id;
      this.refreshLayers();
    }
  }

  // Simplify a polygon annotation using Douglas-Peucker algorithm
  // tolerance is in micrometers, will be converted to map coordinates internally
  simplifyPolygonAnnotation(
    annotationId: string,
    toleranceMicrometers: number
  ): {
    geometry: Polygon | MultiPolygon;
    originalPoints: number;
    simplifiedPoints: number;
    reduction: number;
    reductionPercent: string;
    toleranceUsedMicrometers: number;
  } | null {
    if (
      !this.jstsParser ||
      !this.persistedRoiLayerRef ||
      toleranceMicrometers <= 0
    ) {
      console.error(
        "Cannot simplify: invalid parameters or JSTS not available"
      );
      return null;
    }

    const source = this.persistedRoiLayerRef.source;
    const feature = source.getFeatureById(annotationId);
    if (!feature) {
      console.error(`Feature with ID ${annotationId} not found`);
      return null;
    }

    const geometry = feature.getGeometry();
    if (!(geometry instanceof Polygon) && !(geometry instanceof MultiPolygon)) {
      console.error(
        `Feature ${annotationId} is not a polygon, cannot simplify`
      );
      return null;
    }

    try {
      // Get original point count for comparison
      const originalPointCount = this.countPolygonPoints(geometry);
      const originalArea = geometry.getArea();

      // Convert tolerance from micrometers to map coordinates
      let toleranceMapCoords: number;
      if (this.slideMpp && this.slideMpp > 0) {
        // Convert micrometers to map coordinates: micrometers * 1e-6 * mpp
        toleranceMapCoords = toleranceMicrometers * 1e-6 * this.slideMpp;
        console.log(
          `Converting tolerance: ${toleranceMicrometers}μm → ${toleranceMapCoords.toFixed(
            8
          )} map units (MPP: ${this.slideMpp})`
        );
      } else {
        // Fallback: assume tolerance is already in map coordinates
        toleranceMapCoords = toleranceMicrometers * 1e-6; // Convert to very small map units
        console.warn(
          `No MPP available, using fallback tolerance conversion: ${toleranceMicrometers}μm → ${toleranceMapCoords.toFixed(
            8
          )}`
        );
      }

      // Debug polygon information
      const extent = geometry.getExtent();
      const width = extent[2] - extent[0];
      const height = extent[3] - extent[1];
      const diagonal = Math.sqrt(width * width + height * height);

      console.log(`Polygon debug info:`, {
        annotationId,
        extent,
        width: width.toFixed(8),
        height: height.toFixed(8),
        diagonal: diagonal.toFixed(8),
        area: originalArea.toFixed(8),
        toleranceMicrometers,
        toleranceMapCoords: toleranceMapCoords.toFixed(8),
        slideMpp: this.slideMpp,
      });

      // Calculate maximum safe tolerance in map coordinates
      // Use area-based approach for robustness: tolerance shouldn't reduce area by more than 10%
      const maxSafeToleranceMapCoords = Math.sqrt(originalArea) / 50; // Conservative approach

      let actualToleranceMapCoords = toleranceMapCoords;
      if (toleranceMapCoords > maxSafeToleranceMapCoords) {
        console.warn(
          `Tolerance ${toleranceMapCoords.toFixed(
            8
          )} is too high for polygon (max safe: ${maxSafeToleranceMapCoords.toFixed(
            8
          )}), clamping`
        );
        actualToleranceMapCoords = maxSafeToleranceMapCoords;
      }

      // Convert back to micrometers for logging
      const actualToleranceMicrometers =
        this.slideMpp && this.slideMpp > 0
          ? actualToleranceMapCoords / (1e-6 * this.slideMpp)
          : actualToleranceMapCoords * 1e6;

      // Convert to JSTS geometry
      const jstsGeom = this.jstsParser.read(geometry);

      // Ensure JSTS geometry is polygonal
      if (
        !jstsGeom ||
        (jstsGeom.getGeometryType() !== "Polygon" &&
          jstsGeom.getGeometryType() !== "MultiPolygon")
      ) {
        console.error(
          `JSTS conversion resulted in non-polygonal geometry: ${jstsGeom?.getGeometryType()}`
        );
        return null;
      }

      // Validate geometry before simplification
      const validatedGeom = this.validateAndFixJSTSGeometry(jstsGeom);

      // Apply Douglas-Peucker simplification with validation
      let simplified;
      try {
        simplified = DouglasPeuckerSimplifier.simplify(
          validatedGeom,
          actualToleranceMapCoords
        );
      } catch (error) {
        console.error("Douglas-Peucker simplification failed:", error);
        return null;
      }

      // Ensure simplified result is still polygonal
      if (
        !simplified ||
        (simplified.getGeometryType() !== "Polygon" &&
          simplified.getGeometryType() !== "MultiPolygon")
      ) {
        console.error(
          `Simplification resulted in non-polygonal geometry: ${simplified?.getGeometryType()}`
        );
        return null;
      }

      // Check if simplified geometry is empty or has zero area at JSTS level
      try {
        const simplifiedArea = simplified.getArea();
        if (simplifiedArea <= 0) {
          console.error(
            `JSTS simplified geometry has zero or negative area (${simplifiedArea})`
          );
          return null;
        }

        // Check if area reduction is too extreme at JSTS level
        const jstsOriginalArea = validatedGeom.getArea();
        const jstsAreaReduction =
          ((jstsOriginalArea - simplifiedArea) / jstsOriginalArea) * 100;
        if (jstsAreaReduction > 95) {
          console.error(
            `Extreme area reduction at JSTS level (${jstsAreaReduction.toFixed(
              1
            )}%), aborting simplification`
          );
          return null;
        }
      } catch (error) {
        console.error("Error checking JSTS geometry area:", error);
        return null;
      }

      // Clean the result with zero-buffer operation to fix any topology issues
      let cleanedResult;
      try {
        cleanedResult = BufferOp.bufferOp(simplified, 0);
      } catch (error) {
        console.error("Buffer operation failed:", error);
        return null;
      }

      if (
        !cleanedResult ||
        (cleanedResult.getGeometryType() !== "Polygon" &&
          cleanedResult.getGeometryType() !== "MultiPolygon")
      ) {
        console.error(
          `Buffer operation after simplification resulted in non-polygonal geometry: ${cleanedResult?.getGeometryType()}`
        );
        return null;
      }

      // Convert back to OpenLayers geometry
      const simplifiedGeometry = this.jstsParser.write(cleanedResult);

      // Final validation
      if (
        !(simplifiedGeometry instanceof Polygon) &&
        !(simplifiedGeometry instanceof MultiPolygon)
      ) {
        console.error(
          "Final simplification result is not a polygon:",
          simplifiedGeometry?.getType()
        );
        return null;
      }

      // Check that simplified geometry has valid area
      const area = simplifiedGeometry.getArea();
      const areaReduction = ((originalArea - area) / originalArea) * 100;

      if (area <= 0) {
        console.error(
          `Simplified geometry has zero or negative area (${area})`
        );
        return null;
      }

      // Check if area reduction is too extreme (more than 90% area loss is suspicious)
      if (areaReduction > 90) {
        console.warn(
          `Simplification caused extreme area reduction (${areaReduction.toFixed(
            1
          )}%), this may indicate over-simplification`
        );
      }

      // Check if area reduction is reasonable (less than 50% is usually safe)
      if (areaReduction > 50) {
        console.warn(
          `Significant area reduction (${areaReduction.toFixed(
            1
          )}%) - consider lower tolerance`
        );
      }

      // Count points in simplified geometry
      const simplifiedPointCount = this.countPolygonPoints(simplifiedGeometry);

      // Ensure we actually simplified something (if no reduction, tolerance may be too low)
      if (simplifiedPointCount >= originalPointCount) {
        console.warn(
          `No point reduction achieved (${originalPointCount} → ${simplifiedPointCount}), tolerance may be too low`
        );
      }
      const reduction = originalPointCount - simplifiedPointCount;
      const reductionPercent = ((reduction / originalPointCount) * 100).toFixed(
        1
      );

      console.log(
        `Simplification preview (tolerance: ${actualToleranceMicrometers.toFixed(
          2
        )}μm): ${originalPointCount} → ${simplifiedPointCount} points ` +
          `(${reduction} points removed, ${reductionPercent}% reduction)`
      );

      // Show preview on map
      this.showGeometryPreview(simplifiedGeometry, "simplify");

      return {
        geometry: simplifiedGeometry,
        originalPoints: originalPointCount,
        simplifiedPoints: simplifiedPointCount,
        reduction: reduction,
        reductionPercent: reductionPercent,
        toleranceUsedMicrometers: actualToleranceMicrometers,
      };
    } catch (error) {
      console.error("Error during polygon simplification:", error);
      return null;
    }
  }

  // Apply simplification to a polygon annotation (commits the change)
  applyPolygonSimplification(
    annotationId: string,
    toleranceMicrometers: number
  ): boolean {
    const simplificationResult = this.simplifyPolygonAnnotation(
      annotationId,
      toleranceMicrometers
    );
    if (!simplificationResult) {
      return false;
    }

    const { geometry: simplifiedGeometry } = simplificationResult;

    try {
      const source = this.persistedRoiLayerRef?.source;
      const feature = source?.getFeatureById(annotationId);
      if (!feature || !source) {
        console.error(
          "Feature or source not found during simplification apply"
        );
        return false;
      }

      // Update the feature with simplified geometry
      feature.setGeometry(simplifiedGeometry);

      // Save to localStorage
      this.saveFeatureToStorage(feature);

      // Clear preview and refresh layers
      this.clearPreview();
      this.refreshLayers();

      console.log(
        `Successfully simplified annotation ${annotationId} with tolerance ${toleranceMicrometers}μm`
      );
      return true;
    } catch (error) {
      console.error("Error applying polygon simplification:", error);
      return false;
    }
  }

  // Preview merge of multiple polygon annotations
  previewPolygonMerge(annotationIds: string[]): Polygon | MultiPolygon | null {
    if (
      !this.jstsParser ||
      !this.persistedRoiLayerRef ||
      annotationIds.length < 2
    ) {
      console.error(
        "Cannot preview merge: insufficient annotations or JSTS not available"
      );
      return null;
    }

    const source = this.persistedRoiLayerRef.source;
    const polygonGeometries: (Polygon | MultiPolygon)[] = [];

    // Collect all polygon features to merge
    for (const id of annotationIds) {
      const feature = source.getFeatureById(id);
      if (!feature) {
        console.error(`Feature with ID ${id} not found`);
        return null;
      }

      const geometry = feature.getGeometry();
      if (geometry instanceof Polygon || geometry instanceof MultiPolygon) {
        // Additional validation: ensure the polygon has valid area
        const area = geometry.getArea();
        if (area <= 0) {
          console.error(
            `Feature ${id} has zero or negative area (${area}), cannot merge`
          );
          return null;
        }

        polygonGeometries.push(geometry);
      } else {
        console.error(
          `Feature ${id} is not a polygon (type: ${geometry?.getType()}), cannot merge`
        );
        return null;
      }
    }

    if (polygonGeometries.length < 2) {
      console.error("Need at least 2 polygons to merge");
      return null;
    }

    try {
      // Convert all polygons to JSTS geometries and validate them
      const jstsGeometries = polygonGeometries.map((geom, index) => {
        const jstsGeom = this.jstsParser.read(geom);

        // Ensure JSTS geometry is polygonal
        if (
          !jstsGeom ||
          (jstsGeom.getGeometryType() !== "Polygon" &&
            jstsGeom.getGeometryType() !== "MultiPolygon")
        ) {
          throw new Error(
            `JSTS conversion resulted in non-polygonal geometry for feature ${
              annotationIds[index]
            }: ${jstsGeom?.getGeometryType()}`
          );
        }

        return this.validateAndFixJSTSGeometry(jstsGeom);
      });

      // Perform union operation on all geometries
      let result = jstsGeometries[0];
      for (let i = 1; i < jstsGeometries.length; i++) {
        result = OverlayOp.overlayOp(
          result,
          jstsGeometries[i],
          OverlayOp.UNION
        );

        // Validate intermediate result is still polygonal
        if (
          !result ||
          (result.getGeometryType() !== "Polygon" &&
            result.getGeometryType() !== "MultiPolygon")
        ) {
          throw new Error(
            `Union operation ${i} resulted in non-polygonal geometry: ${result?.getGeometryType()}`
          );
        }
      }

      // Clean the result with zero-buffer operation
      const cleanedResult = BufferOp.bufferOp(result, 0);

      // Ensure cleaned result is still polygonal
      if (
        !cleanedResult ||
        (cleanedResult.getGeometryType() !== "Polygon" &&
          cleanedResult.getGeometryType() !== "MultiPolygon")
      ) {
        throw new Error(
          `Buffer operation resulted in non-polygonal geometry: ${cleanedResult?.getGeometryType()}`
        );
      }

      const mergedGeometry = this.jstsParser.write(cleanedResult);

      // Ensure the result is a valid Polygon or MultiPolygon
      if (!mergedGeometry) {
        console.error("Merge operation resulted in empty geometry");
        return null;
      }

      if (
        !(mergedGeometry instanceof Polygon) &&
        !(mergedGeometry instanceof MultiPolygon)
      ) {
        console.error(
          "Merge operation resulted in non-polygon geometry:",
          mergedGeometry.getType()
        );
        return null;
      }

      // Additional check: ensure the geometry has valid area
      const area = mergedGeometry.getArea();
      if (area <= 0) {
        console.error(
          "Merge operation resulted in geometry with zero or negative area"
        );
        return null;
      }

      // Show preview on map
      this.showGeometryPreview(mergedGeometry, "merge");

      return mergedGeometry;
    } catch (error) {
      console.error("Error during polygon merge preview:", error);
      return null;
    }
  }

  // Merge multiple polygon annotations into a single polygon using JSTS union
  mergePolygonAnnotations(annotationIds: string[]): boolean {
    if (
      !this.jstsParser ||
      !this.persistedRoiLayerRef ||
      annotationIds.length < 2
    ) {
      console.error(
        "Cannot merge: insufficient annotations or JSTS not available"
      );
      return false;
    }

    const source = this.persistedRoiLayerRef.source;
    const features: Feature[] = [];
    const polygonGeometries: (Polygon | MultiPolygon)[] = [];

    // Collect all polygon features to merge
    for (const id of annotationIds) {
      const feature = source.getFeatureById(id);
      if (!feature) {
        console.error(`Feature with ID ${id} not found`);
        return false;
      }

      const geometry = feature.getGeometry();
      if (geometry instanceof Polygon || geometry instanceof MultiPolygon) {
        // Additional validation: ensure the polygon has valid area
        const area = geometry.getArea();
        if (area <= 0) {
          console.error(
            `Feature ${id} has zero or negative area (${area}), cannot merge`
          );
          return false;
        }

        // Ensure polygon has valid coordinates
        const coordinates =
          geometry instanceof Polygon
            ? geometry.getCoordinates()
            : geometry.getCoordinates();

        if (!coordinates || coordinates.length === 0) {
          console.error(`Feature ${id} has invalid coordinates, cannot merge`);
          return false;
        }

        features.push(feature);
        polygonGeometries.push(geometry);
      } else {
        console.error(
          `Feature ${id} is not a polygon (type: ${geometry?.getType()}), cannot merge`
        );
        return false;
      }
    }

    if (polygonGeometries.length < 2) {
      console.error("Need at least 2 polygons to merge");
      return false;
    }

    try {
      // Convert all polygons to JSTS geometries and validate them
      const jstsGeometries = polygonGeometries.map((geom, index) => {
        const jstsGeom = this.jstsParser.read(geom);

        // Ensure JSTS geometry is polygonal
        if (
          !jstsGeom ||
          (jstsGeom.getGeometryType() !== "Polygon" &&
            jstsGeom.getGeometryType() !== "MultiPolygon")
        ) {
          throw new Error(
            `JSTS conversion resulted in non-polygonal geometry for feature ${
              annotationIds[index]
            }: ${jstsGeom?.getGeometryType()}`
          );
        }

        return this.validateAndFixJSTSGeometry(jstsGeom);
      });

      // Perform union operation on all geometries
      let result = jstsGeometries[0];
      for (let i = 1; i < jstsGeometries.length; i++) {
        result = OverlayOp.overlayOp(
          result,
          jstsGeometries[i],
          OverlayOp.UNION
        );

        // Validate intermediate result is still polygonal
        if (
          !result ||
          (result.getGeometryType() !== "Polygon" &&
            result.getGeometryType() !== "MultiPolygon")
        ) {
          throw new Error(
            `Union operation ${i} resulted in non-polygonal geometry: ${result?.getGeometryType()}`
          );
        }
      }

      // Clean the result with zero-buffer operation
      const cleanedResult = BufferOp.bufferOp(result, 0);

      // Ensure cleaned result is still polygonal
      if (
        !cleanedResult ||
        (cleanedResult.getGeometryType() !== "Polygon" &&
          cleanedResult.getGeometryType() !== "MultiPolygon")
      ) {
        throw new Error(
          `Buffer operation resulted in non-polygonal geometry: ${cleanedResult?.getGeometryType()}`
        );
      }

      const mergedGeometry = this.jstsParser.write(cleanedResult);

      // Ensure the result is a valid Polygon or MultiPolygon
      if (!mergedGeometry) {
        console.error("Merge operation resulted in empty geometry");
        return false;
      }

      if (
        !(mergedGeometry instanceof Polygon) &&
        !(mergedGeometry instanceof MultiPolygon)
      ) {
        console.error(
          "Merge operation resulted in non-polygon geometry:",
          mergedGeometry.getType()
        );
        return false;
      }

      // Additional check: ensure the geometry has valid area
      const area = mergedGeometry.getArea();
      if (area <= 0) {
        console.error(
          "Merge operation resulted in geometry with zero or negative area"
        );
        return false;
      }

      // Use the first feature as the base for the merged annotation
      const baseFeature = features[0];
      const mergedId = String(baseFeature.getId());

      // Update the base feature with merged geometry
      baseFeature.setGeometry(mergedGeometry);

      // Remove all other features from the map and storage
      const featuresToRemove = features.slice(1);
      featuresToRemove.forEach((feature) => {
        const id = String(feature.getId());
        source.removeFeature(feature);
      });

      // Clean up localStorage completely and save only the merged feature
      this.cleanupMergedFeaturesInStorage(annotationIds, baseFeature);

      // Update annotations list by removing merged annotations and keeping the base
      if (this.onAnnotationsUpdate && this.getCurrentAnnotations) {
        const currentAnnotations = this.getCurrentAnnotations();
        const idsToRemove = new Set(annotationIds.slice(1)); // Remove all except the first
        const updatedAnnotations = currentAnnotations.filter(
          (ann) => !idsToRemove.has(ann.id)
        );
        this.onAnnotationsUpdate(updatedAnnotations);
      }

      // Clear preview and refresh layers
      this.clearPreview();
      this.refreshLayers();

      console.log(
        `Successfully merged ${annotationIds.length} polygons into ${mergedId}`
      );
      return true;
    } catch (error) {
      console.error("Error during polygon merge:", error);
      return false;
    }
  }

  private validateAndFixJSTSGeometry(jstsGeom: any): any {
    if (!jstsGeom) return jstsGeom;

    try {
      // Check if geometry is valid
      const validOp = new IsValidOp(jstsGeom);
      if (validOp.isValid()) {
        return jstsGeom;
      }

      console.warn(
        "Invalid geometry detected during merge, attempting to fix:",
        validOp.getValidationError()
      );

      // Try to fix with zero-buffer operation
      const buffered = BufferOp.bufferOp(jstsGeom, 0);
      if (buffered && new IsValidOp(buffered).isValid()) {
        return buffered;
      }

      // If still invalid, try a small buffer to clean up topology
      const smallBuffer = BufferOp.bufferOp(jstsGeom, 1e-10);
      if (smallBuffer && new IsValidOp(smallBuffer).isValid()) {
        return smallBuffer;
      }

      // Last resort: return original geometry
      return jstsGeom;
    } catch (error) {
      console.warn("Failed to validate/fix geometry during merge:", error);
      return jstsGeom;
    }
  }

  private cleanupMergedFeaturesInStorage(
    mergedIds: string[],
    mergedFeature: Feature
  ) {
    if (!this.transformMapToPixelCoordinates) return;

    try {
      const key = `slideAnnotations:${this.slideUid}`;
      const raw = localStorage.getItem(key);
      if (!raw) return;

      const data = JSON.parse(raw);
      if (!data || !Array.isArray(data.features)) return;

      // Remove all merged features from storage
      const mergedIdSet = new Set(mergedIds);
      data.features = data.features.filter((f: any) => {
        const id = f?.id != null ? String(f.id) : undefined;
        return !mergedIdSet.has(id || "");
      });

      // Add the merged feature to storage
      const format = new GeoJSON();
      const geoJsonInMap = format.writeFeatureObject(mergedFeature);
      const geoJsonInPixel = this.transformMapToPixelCoordinates({
        type: "FeatureCollection",
        features: [geoJsonInMap],
      });

      const mergedFeatureData = geoJsonInPixel.features[0];
      // Mark merged feature as modified and needing sync
      mergedFeatureData.properties = {
        ...(mergedFeatureData.properties || {}),
        lastModified: Date.now(),
      };
      // Remove vectorUid to mark as local
      delete mergedFeatureData.properties.vectorUid;

      data.features.push(mergedFeatureData);
      localStorage.setItem(key, JSON.stringify(data));

      console.log(
        `Cleaned up localStorage: removed ${mergedIds.length} features, added 1 merged feature`
      );
    } catch (e) {
      console.error("Error cleaning up merged features in localStorage:", e);
    }
  }

  private removeFromLocalStorage(annotationId: string) {
    const key = `slideAnnotations:${this.slideUid}`;
    const raw = localStorage.getItem(key);
    if (raw) {
      try {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          data.features = data.features.filter((f: any) => {
            const id = f?.id != null ? String(f.id) : undefined;
            return id !== annotationId;
          });
          localStorage.setItem(key, JSON.stringify(data));
        }
      } catch (e) {
        console.error("Error removing annotation from localStorage:", e);
      }
    }
  }

  // Remove annotation by ID from both map layers and localStorage
  removeAnnotationById(id: string) {
    // Remove from layers
    const persistedSource = this.persistedRoiLayerRef?.source;
    if (persistedSource) {
      const feature = persistedSource.getFeatureById(id);
      if (feature) {
        persistedSource.removeFeature(feature);
      }
    }

    const tempSource = this.tempRoiLayerRef?.source;
    if (tempSource) {
      const feature = tempSource.getFeatureById(id);
      if (feature) {
        tempSource.removeFeature(feature);
      }
    }

    // Remove from localStorage
    const key = `slideAnnotations:${this.slideUid}`;
    const raw = localStorage.getItem(key);
    if (raw) {
      try {
        const data = JSON.parse(raw);
        if (data && Array.isArray(data.features)) {
          data.features = data.features.filter((f: any) => {
            const fid = f?.id != null ? String(f.id) : undefined;
            return fid !== id;
          });
          localStorage.setItem(key, JSON.stringify(data));
        }
      } catch (e) {
        console.error("Error removing annotation from localStorage:", e);
      }
    }

    // Refresh layers
    this.refreshLayers();
  }

  destroy() {
    // Clean up interactions
    if (this.map) {
      if (this.activeDrawRef) {
        try {
          this.map.removeInteraction(this.activeDrawRef);
        } catch {}
      }
      if (this.modifyInteractionRef) {
        try {
          this.map.removeInteraction(this.modifyInteractionRef);
        } catch {}
      }
      if (this.snapInteractionRef) {
        try {
          this.map.removeInteraction(this.snapInteractionRef);
        } catch {}
      }
    }

    // Clean up brush tool
    this.brushRef?.destroy();

    // Clear preview
    this.clearPreview();

    // Clean up handlers
    if (this.seedPointerDownHandlerRef && this.map) {
      try {
        this.map.un(
          "pointerdown" as any,
          this.seedPointerDownHandlerRef as any
        );
      } catch {}
    }

    // NOTE: We do NOT remove the annotation layers from the map here!
    // The layers should persist even when the plugin panel is closed.
    // Only remove layers when the plugin is actually unregistered.
  }

  // New method for complete cleanup when plugin is unregistered
  destroyCompletely() {
    // First do regular cleanup
    this.destroy();

    // Then remove layers from map
    if (this.map && this.persistedRoiLayerRef) {
      try {
        this.map.removeLayer(this.persistedRoiLayerRef.layer);
      } catch {}
    }
    if (this.map && this.tempRoiLayerRef) {
      try {
        this.map.removeLayer(this.tempRoiLayerRef.layer);
      } catch {}
    }
    if (this.map && this.previewLayerRef) {
      try {
        this.map.removeLayer(this.previewLayerRef.layer);
      } catch {}
    }

    // Clear references
    this.persistedRoiLayerRef = null;
    this.tempRoiLayerRef = null;
    this.previewLayerRef = null;
  }
}

function AnnotationPluginPanel({ api, onClose }: AnnotationPluginPanelProps) {
  const { context } = api;

  // Use global drawing manager instance to persist across panel open/close
  const drawingManagerRef = useRef<AnnotationDrawingManager | null>(
    globalDrawingManager
  );
  // Plugin manages its own state - no longer dependent on context initialization
  const [annotations, setAnnotations] = useState<AnnotationItem[]>([]);
  const [selectedAnnotationId, setSelectedAnnotationId] = useState<
    string | null
  >(null);
  const [hoveredAnnotationId, setHoveredAnnotationId] = useState<string | null>(
    null
  );
  const [activeAnnotationLabel, setActiveAnnotationLabel] =
    useState<LabelId | null>(null);
  const [brushActive, setBrushActive] = useState(false);
  const [brushMode, setBrushMode] = useState<"add" | "erase">("add");
  const [brushSizePx, setBrushSizePx] = useState(24);

  // Fetch annotation settings from the study
  const {
    data: annotationSettings,
    isLoading: annotationSettingsLoading,
    error: annotationSettingsError,
  } = useStudyAnnotationSettings(context.studyUid);

  // Annotation display settings
  const {
    settings: annotationDisplaySettings,
    updateSettings: updateAnnotationDisplaySettings,
    resetSettings: resetAnnotationDisplaySettings,
  } = useAnnotationSettings();

  // Coordinate transforms
  const { transformMapToPixelCoordinates, transformPixelToMapCoordinates } =
    useCoordinateTransforms(context.rawSlideMetadata?.slideMpp);

  // Annotation sync functionality
  const {
    syncState,
    loadAnnotations,
    manualSync,
    deleteAnnotation,
    isLoading: isSyncLoading,
    isSyncing,
    lastSyncTime,
    error: syncError,
  } = useAnnotationSync(context.slideUid || "");

  // Keep latest display settings in a ref so OL style functions can read live values
  const annotationDisplaySettingsRef = useRef(annotationDisplaySettings);
  useEffect(() => {
    annotationDisplaySettingsRef.current = annotationDisplaySettings;
  }, [annotationDisplaySettings]);

  // Helper function to get label colors from annotation settings
  const getLabelColors = useCallback(() => {
    const labelColors: Record<LabelId, string> = {};

    if (
      annotationSettings?.annotations &&
      Array.isArray(annotationSettings.annotations)
    ) {
      annotationSettings.annotations
        .filter(
          (annotation) =>
            annotation &&
            annotation.label &&
            annotation.color &&
            annotation.type
        )
        .forEach((annotation) => {
          labelColors[annotation.label] = annotation.color;
        });
    }

    return { labelColors };
  }, [annotationSettings]);

  // Helper function to convert micrometers to pixels for point annotations
  const micrometersToPixels = useCallback(
    (micrometers: number): number => {
      const mpp = context.rawSlideMetadata?.slideMpp;
      if (!mpp || mpp <= 0) {
        return micrometers / 0.25; // Fallback: assume 0.25 micrometers per pixel
      }
      return micrometers / mpp;
    },
    [context.rawSlideMetadata?.slideMpp]
  );

  // Keep annotations in a ref so the drawing manager can access current state
  const annotationsRef = useRef<AnnotationItem[]>([]);
  useEffect(() => {
    annotationsRef.current = annotations;
  }, [annotations]);

  // Track the current slideUid to detect changes
  const currentSlideUidRef = useRef<string | null>(null);

  // Initialize the drawing manager when map and metadata are available
  useEffect(() => {
    if (!context.map || !context.slideUid || !context.rawSlideMetadata) return;

    // Detect slide change - clear previous slide's annotations
    if (
      currentSlideUidRef.current &&
      currentSlideUidRef.current !== context.slideUid
    ) {
      console.log(
        "AnnotationControlPlugin: Slide changed from",
        currentSlideUidRef.current,
        "to",
        context.slideUid
      );
      console.log(
        "AnnotationControlPlugin: Clearing annotation layers from previous slide"
      );

      // Clear the drawing manager completely for the new slide
      if (drawingManagerRef.current) {
        drawingManagerRef.current.destroyCompletely();
        drawingManagerRef.current = null;
        globalDrawingManager = null;
      }

      // Clear annotations state
      setAnnotations([]);
      setSelectedAnnotationId(null);
      setHoveredAnnotationId(null);
    }

    // Update current slide reference
    currentSlideUidRef.current = context.slideUid;

    if (!drawingManagerRef.current) {
      drawingManagerRef.current = new AnnotationDrawingManager();
      globalDrawingManager = drawingManagerRef.current; // Store in global reference
    }

    drawingManagerRef.current.initialize(
      context.map,
      context.slideUid,
      context.rawSlideMetadata,
      transformMapToPixelCoordinates,
      transformPixelToMapCoordinates,
      {
        onAnnotationsUpdate: setAnnotations,
        onAnnotationSelect: setSelectedAnnotationId,
        onAnnotationHover: setHoveredAnnotationId,
        getLabelColors,
        micrometersToPixels,
        getCurrentAnnotations: () => annotationsRef.current,
      },
      annotationDisplaySettingsRef,
      annotationSettings,
      annotationSettingsLoading
    );

    return () => {
      // Only cleanup interactions when panel closes, keep layers visible
      // unless we're changing slides (in which case destroyCompletely was already called)
      if (
        drawingManagerRef.current &&
        currentSlideUidRef.current === context.slideUid
      ) {
        drawingManagerRef.current.destroy();
      }
      // Don't clear the manager reference - keep it for persistence across panel open/close
    };
  }, [
    context.map,
    context.slideUid,
    context.rawSlideMetadata,
    transformMapToPixelCoordinates,
    transformPixelToMapCoordinates,
    getLabelColors,
    micrometersToPixels,
    annotationSettings,
    annotationSettingsLoading,
  ]);

  // Load annotations when component mounts or slide changes
  useEffect(() => {
    if (
      !drawingManagerRef.current ||
      !context.rawSlideMetadata ||
      !context.slideUid
    )
      return;

    // Load annotations for the current slide
    loadAnnotations()
      .then((syncedAnnotations) => {
        console.log(
          "Loaded synced annotations for slide",
          context.slideUid,
          ":",
          syncedAnnotations.length
        );
        setAnnotations(syncedAnnotations);

        // Load GeoJSON data from localStorage
        const key = `slideAnnotations:${context.slideUid}`;
        const raw = localStorage.getItem(key);
        if (raw) {
          try {
            const fc = JSON.parse(raw);
            if (fc.features && fc.features.length > 0) {
              console.log(
                "Loading",
                fc.features.length,
                "annotation features from localStorage for slide",
                context.slideUid
              );
              drawingManagerRef.current?.loadAnnotations(fc);
            }
          } catch (e) {
            console.error("Error loading annotation GeoJSON:", e);
          }
        }
      })
      .catch((error) => {
        console.error("Error during annotation sync:", error);
        // Fall back to localStorage-only loading
        const key = `slideAnnotations:${context.slideUid}`;
        const raw = localStorage.getItem(key);
        if (raw) {
          try {
            const fc = JSON.parse(raw);
            if (fc.features && fc.features.length > 0) {
              console.log(
                "Fallback: Loading",
                fc.features.length,
                "annotation features from localStorage for slide",
                context.slideUid
              );
              drawingManagerRef.current?.loadAnnotations(fc);
              // Extract annotations from features
              const anns: AnnotationItem[] = fc.features.map((f: any) => ({
                id: String(f.id || `${Date.now()}-${Math.random()}`),
                name: f.properties?.name ?? "roi",
                visible: f.properties?.visible ?? true,
                kind: f.properties?.kind ?? undefined,
              }));
              setAnnotations(anns);
            }
          } catch (e) {
            console.error("Error in fallback annotation loading:", e);
          }
        }
      });
  }, [
    context.slideUid,
    context.rawSlideMetadata,
    loadAnnotations,
    drawingManagerRef.current,
  ]);

  // Sync brush state with drawing manager
  useEffect(() => {
    drawingManagerRef.current?.setBrushActive(brushActive);
  }, [brushActive]);

  useEffect(() => {
    drawingManagerRef.current?.setBrushMode(brushMode);
  }, [brushMode]);

  useEffect(() => {
    drawingManagerRef.current?.setBrushSize(brushSizePx);
  }, [brushSizePx]);

  // Update selected feature for brush tool and highlighting
  useEffect(() => {
    if (!drawingManagerRef.current) return;

    const feature = selectedAnnotationId
      ? drawingManagerRef.current.getFeatureById(selectedAnnotationId)
      : null;

    // Set selected feature for brush tool
    drawingManagerRef.current.setSelectedFeature(feature);

    // Set selected annotation for highlighting
    drawingManagerRef.current.setSelectedAnnotationId(selectedAnnotationId);
  }, [selectedAnnotationId]);

  // Update active annotation label in drawing manager
  useEffect(() => {
    if (!drawingManagerRef.current) return;
    drawingManagerRef.current.setActiveAnnotationLabel(activeAnnotationLabel);
  }, [activeAnnotationLabel]);

  // Update hovered annotation in drawing manager (for list -> map highlighting)
  useEffect(() => {
    if (!drawingManagerRef.current) return;
    drawingManagerRef.current.setHoveredAnnotationId(hoveredAnnotationId);
  }, [hoveredAnnotationId]);

  // Refresh layers when display settings change
  useEffect(() => {
    drawingManagerRef.current?.refreshLayers();
  }, [annotationDisplaySettings]);

  // Refresh layers when annotation settings finish loading
  useEffect(() => {
    if (annotationSettings && !annotationSettingsLoading) {
      drawingManagerRef.current?.refreshLayers();
    }
  }, [annotationSettings, annotationSettingsLoading]);

  // Only use periodic refresh for major changes (like when annotations are deleted externally)
  // The immediate updates handle creation, so we only need occasional sync
  useEffect(() => {
    const syncInterval = setInterval(async () => {
      try {
        const refreshedAnnotations = await loadAnnotations();
        // Only update if there's a significant change (not just +1 annotation)
        const currentCount = annotations.length;
        const newCount = refreshedAnnotations.length;

        // Update if:
        // 1. Count decreased (deletions)
        // 2. Count increased by more than 1 (bulk operations)
        // 3. No annotations currently but some loaded (initial load)
        if (
          newCount < currentCount ||
          newCount > currentCount + 1 ||
          (currentCount === 0 && newCount > 0)
        ) {
          console.log(
            "Sync refresh detected significant change:",
            currentCount,
            "->",
            newCount
          );
          setAnnotations(refreshedAnnotations);
        }
      } catch (error) {
        console.error("Error in sync refresh:", error);
      }
    }, 15000); // Check every 15 seconds, less aggressive

    return () => clearInterval(syncInterval);
  }, [loadAnnotations, annotations.length]);

  // Drawing tool handlers
  const handleStartDrawROI = useCallback(
    (mode: "point" | "box" | "polygon", label: LabelId) => {
      drawingManagerRef.current?.startRoiDraw(mode, label);
    },
    []
  );

  const handleStopDraw = useCallback(() => {
    drawingManagerRef.current?.stopRoiDraw();
  }, []);

  const handleStartBrushAdd = useCallback(() => {
    setBrushMode("add");
    setBrushActive(true);
  }, []);

  const handleStartBrushErase = useCallback(() => {
    setBrushMode("erase");
    setBrushActive(true);
  }, []);

  const handleStopBrush = useCallback(() => {
    setBrushActive(false);
  }, []);

  const handleMergeAnnotations = useCallback((annotationIds: string[]) => {
    if (!drawingManagerRef.current) {
      console.error("Drawing manager not available for merge");
      return false;
    }
    return drawingManagerRef.current.mergePolygonAnnotations(annotationIds);
  }, []);

  const handleSimplifyAnnotation = useCallback(
    (annotationId: string, tolerance: number) => {
      if (!drawingManagerRef.current) {
        console.error("Drawing manager not available for simplify");
        return false;
      }
      return drawingManagerRef.current.applyPolygonSimplification(
        annotationId,
        tolerance
      );
    },
    []
  );

  const handlePreviewSimplification = useCallback(
    (annotationId: string, tolerance: number) => {
      if (!drawingManagerRef.current) {
        console.error("Drawing manager not available for preview");
        return null;
      }
      return drawingManagerRef.current.simplifyPolygonAnnotation(
        annotationId,
        tolerance
      );
    },
    []
  );

  const handlePreviewMerge = useCallback((annotationIds: string[]) => {
    if (!drawingManagerRef.current) {
      console.error("Drawing manager not available for merge preview");
      return null;
    }
    return drawingManagerRef.current.previewPolygonMerge(annotationIds);
  }, []);

  const handleClearPreview = useCallback(() => {
    if (!drawingManagerRef.current) return;
    drawingManagerRef.current.clearPreview();
  }, []);

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  const handleUpdateAnnotations = useCallback(
    async (newAnnotations: AnnotationItem[]) => {
      const prevIds = new Set(annotations.map((a) => a.id));
      const nextIds = new Set(newAnnotations.map((a) => a.id));

      // Find removed annotation IDs
      const removedIds: string[] = [];
      annotations.forEach((a) => {
        if (!nextIds.has(a.id)) removedIds.push(a.id);
      });

      // Delete annotations from server and local storage
      if (removedIds.length > 0) {
        try {
          // Delete each annotation properly
          await Promise.all(removedIds.map((id) => deleteAnnotation(id)));

          // Remove from map layers
          removedIds.forEach((id) => {
            drawingManagerRef.current?.removeAnnotationById(id);
          });

          // Reload annotations to get fresh state
          const refreshedAnnotations = await loadAnnotations();
          setAnnotations(refreshedAnnotations);
        } catch (error) {
          console.error("Error deleting annotations:", error);
          // Still update the UI state even if deletion failed
          setAnnotations(newAnnotations);
        }
      } else {
        // No deletions, just update the state
        setAnnotations(newAnnotations);
      }
    },
    [annotations, deleteAnnotation, loadAnnotations]
  );

  return (
    <>
      <AnnotationPanel
        isOpen={api.state.isOpen}
        onClose={onClose}
        activeLabel={activeAnnotationLabel}
        onActiveLabelChange={setActiveAnnotationLabel}
        annotations={annotations}
        onUpdateAnnotations={handleUpdateAnnotations}
        selectedId={selectedAnnotationId}
        onSelect={setSelectedAnnotationId}
        hoveredId={hoveredAnnotationId}
        onHoverIdChange={setHoveredAnnotationId}
        onHoverGroupChange={undefined} // Plugin manages its own highlighting
        dockOverride={api.state.dock}
        onDockChange={handleDockChange}
        onStartDrawROI={handleStartDrawROI}
        onStopDraw={handleStopDraw}
        brushActive={brushActive}
        brushMode={brushMode}
        brushSizePx={brushSizePx}
        onStartBrushAdd={handleStartBrushAdd}
        onStartBrushErase={handleStartBrushErase}
        onStopBrush={handleStopBrush}
        onBrushSizeChange={setBrushSizePx}
        studyUid={context.studyUid}
        annotationSettings={annotationSettings}
        annotationSettingsLoading={annotationSettingsLoading}
        annotationSettingsError={annotationSettingsError}
        annotationDisplaySettings={annotationDisplaySettings}
        onUpdateAnnotationDisplaySettings={updateAnnotationDisplaySettings}
        onResetAnnotationDisplaySettings={resetAnnotationDisplaySettings}
        slideMpp={context.rawSlideMetadata?.slideMpp}
        onMergeAnnotations={handleMergeAnnotations}
        onSimplifyAnnotation={handleSimplifyAnnotation}
        onPreviewSimplification={handlePreviewSimplification}
        onPreviewMerge={handlePreviewMerge}
        onClearPreview={handleClearPreview}
      />
    </>
  );
}

// Global drawing manager instance to persist across panel open/close
let globalDrawingManager: AnnotationDrawingManager | null = null;

export const AnnotationControlPlugin: ViewerPlugin = {
  id: "annotation-control",
  name: "Annotations",
  version: "1.0.0",

  button: {
    id: "annotation-control-button",
    label: "Annotations",
    icon: PencilSquareIcon,
    tooltip: "Create and manage annotations",
    position: "right",
    order: 2,
  },

  panel: {
    id: "annotation-control-panel",
    title: "Annotations",
    defaultSize: { width: 320, height: 560 },
    defaultDock: "left",
    storageKey: "annotationPanel",
  },

  PanelComponent: AnnotationPluginPanel,

  onButtonClick: (api: PluginAPI) => {
    console.log("AnnotationControlPlugin: Button clicked");
    console.log("AnnotationControlPlugin: API context:", api.context);
    console.log("AnnotationControlPlugin: API state:", api.state);

    // Toggle panel
    console.log("AnnotationControlPlugin: Toggling panel state");
    api.setState({ isOpen: !api.state.isOpen });
  },

  onContextChange: (context) => {
    // Could update plugin availability based on context
    // For example, hide/show button based on study permissions
  },

  destroy: () => {
    // Complete cleanup when plugin is unregistered
    console.log(
      "AnnotationControlPlugin: Destroying plugin and clearing all annotation layers"
    );
    if (globalDrawingManager) {
      globalDrawingManager.destroyCompletely();
      globalDrawingManager = null;
    }
  },
};
