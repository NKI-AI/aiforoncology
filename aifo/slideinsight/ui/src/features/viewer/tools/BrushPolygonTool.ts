// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import Map from "ol/Map";
import Feature from "ol/Feature";
import Geometry from "ol/geom/Geometry";
import Polygon from "ol/geom/Polygon";
import MultiPolygon from "ol/geom/MultiPolygon";
import LineString from "ol/geom/LineString";
import Point from "ol/geom/Point";
import VectorSource from "ol/source/Vector";
import VectorLayer from "ol/layer/Vector";
import { Style, Fill, Stroke } from "ol/style";
import GeoJSON from "ol/format/GeoJSON";
import LinearRing from "ol/geom/LinearRing";
import MultiLineString from "ol/geom/MultiLineString";
import MultiPoint from "ol/geom/MultiPoint";

// JSTS imports using OL3Parser + BufferOp/OverlayOp (ESM paths require .js)
// @ts-ignore: jsts ESM packages don't ship TS types
import OL3Parser from "jsts/org/locationtech/jts/io/OL3Parser.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import BufferOp from "jsts/org/locationtech/jts/operation/buffer/BufferOp.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import OverlayOp from "jsts/org/locationtech/jts/operation/overlay/OverlayOp.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import GeometryFactory from "jsts/org/locationtech/jts/geom/GeometryFactory.js";
// @ts-ignore: jsts ESM packages don't ship TS types
import IsValidOp from "jsts/org/locationtech/jts/operation/valid/IsValidOp.js";

type Mode = "add" | "erase";

/** Options you pass from SlideWorkspace so Brush tool can persist on commit. */
export interface BrushPolygonToolOptions {
  /** Called after we changed the feature geometry (use this to save to localStorage). */
  onCommit?: (feature: Feature<Polygon | MultiPolygon>) => void;
  /** Z-index for the preview overlay. */
  previewZIndex?: number;
  /** Optional: disable map DragPan while drawing. */
  setDragPanActive?: (active: boolean) => void;
  /** Called when brushing with no selected target polygon to create a new polygon from the stroke. */
  onCreateFromStroke?: (geom: Polygon) => void;
}

/**
 * A vector-geometry brush that *modifies a single existing* Polygon/MultiPolygon.
 * - Drag paints; pointerup applies union/difference to the target feature.
 * - Keeps a translucent stroke preview while dragging.
 */
export class BrushPolygonTool {
  private map: Map;
  private target: Feature<Polygon | MultiPolygon> | null = null;

  private mode: Mode = "add";
  private brushPx = 24;
  private drawing = false;

  private previewLayer: VectorLayer<VectorSource>;
  private strokeLine = new LineString([]);
  private rafPending = false;

  private gj = new GeoJSON();
  private parser: any = null;
  private opts: BrushPolygonToolOptions;

  private handlersBound = false;

  private hoverRing: Feature<Polygon> | null = null;
  private _onHoverMove: any = null;

  // Store bound handlers for cleanup
  private _onDown: any = null;
  private _onMove: any = null;
  private _onUp: any = null;

  constructor(map: Map, opts: BrushPolygonToolOptions = {}) {
    this.map = map;
    this.opts = opts;

    this.previewLayer = new VectorLayer({
      source: new VectorSource({ wrapX: false }),
      style: new Style({
        fill: new Fill({ color: "rgba(0,0,255,0.15)" }),
        stroke: new Stroke({ color: "rgba(0,0,200,0.6)", width: 1 }),
      }),
    });
    this.previewLayer.setZIndex(opts.previewZIndex ?? 2200);
    map.addLayer(this.previewLayer);
    this.previewLayer.setVisible(false); // only visible while drawing

    // Initialize OL3Parser and inject OpenLayers geometry classes
    try {
      // Create geometry factory for JSTS operations
      const geometryFactory = new GeometryFactory();
      this.parser = new OL3Parser(geometryFactory, undefined);
      this.parser.inject(
        Point,
        LineString,
        LinearRing,
        Polygon,
        MultiPoint,
        MultiLineString,
        MultiPolygon
      );
    } catch {
      this.parser = null;
    }
  }

  setTargetFeature(f: Feature<Polygon | MultiPolygon> | null) {
    this.target = f;
  }

  setMode(m: Mode) {
    this.mode = m;
    // Update preview style based on mode
    const style = new Style({
      fill: new Fill({
        color: m === "add" ? "rgba(0,255,0,0.15)" : "rgba(255,0,0,0.15)",
      }),
      stroke: new Stroke({
        color: m === "add" ? "rgba(0,200,0,0.6)" : "rgba(200,0,0,0.6)",
        width: 2,
      }),
    });
    this.previewLayer.setStyle(style);
  }

  private ringPolygonAt(coord: number[]): Polygon {
    if (!this.parser) {
      // Fallback small square
      const r = this.mapBrushRadius();
      return new Polygon([
        [
          [coord[0] - r, coord[1] - r],
          [coord[0] + r, coord[1] - r],
          [coord[0] + r, coord[1] + r],
          [coord[0] - r, coord[1] + r],
          [coord[0] - r, coord[1] - r],
        ],
      ]);
    }
    const tiny = new LineString([coord, coord]);
    const jtsLine = this.parser.read(tiny);
    const radius = this.mapBrushRadius();
    const jtsCircle = BufferOp.bufferOp(jtsLine, radius);
    return this.parser.write(jtsCircle) as Polygon;
  }

  private showHoverRingAt(coord: number[]) {
    const src = this.previewLayer.getSource()!;

    // Check if we can erase from this position (for visual feedback)
    const canEraseFromHere =
      this.mode === "erase" &&
      (this.pickEditableTargetAt(this.map.getPixelFromCoordinate(coord)) ||
        this.findNearestEditableTarget(coord));

    if (!this.hoverRing) {
      this.hoverRing = new Feature<Polygon>(this.ringPolygonAt(coord));
      src.addFeature(this.hoverRing);
    } else {
      this.hoverRing.setGeometry(this.ringPolygonAt(coord));
    }

    // Update hover ring style based on mode and capability
    const style = new Style({
      fill: new Fill({
        color:
          this.mode === "add"
            ? "rgba(0,255,0,0.10)"
            : canEraseFromHere
            ? "rgba(255,100,0,0.15)" // Orange when can erase from outside
            : "rgba(255,0,0,0.10)", // Red when normal erase
      }),
      stroke: new Stroke({
        color:
          this.mode === "add"
            ? "rgba(0,255,0,0.8)"
            : canEraseFromHere
            ? "rgba(255,140,0,0.9)" // Bright orange when can erase from outside
            : "rgba(255,0,0,0.8)", // Red when normal erase
        width: canEraseFromHere ? 2 : 1.5, // Thicker when can erase from outside
      }),
    });

    this.previewLayer.setStyle(style);
  }

  private hideHoverRing() {
    if (!this.hoverRing) return;
    this.previewLayer.getSource()!.removeFeature(this.hoverRing);
    this.hoverRing = null;
  }

  setBrushSizePixels(px: number) {
    this.brushPx = Math.max(1, px | 0);
  }

  /** Enable/disable the tool (adds/removes pointer handlers). */
  setActive(active: boolean) {
    if (active && !this.handlersBound) {
      this.bindHandlers();
      this.handlersBound = true;
    } else if (!active && this.handlersBound) {
      this.unbindHandlers();
      this.handlersBound = false;
      this.cancelStroke();
    }
    // Attach hover listener to always see the ring while active
    if (active && !this._onHoverMove) {
      // Make preview layer visible immediately when activating
      this.previewLayer.setVisible(true);

      this._onHoverMove = (e: any) => {
        this.showHoverRingAt(e.coordinate);
      };
      this.map.on("pointermove" as any, this._onHoverMove);
      // optional: crosshair cursor
      try {
        (this.map.getTargetElement() as HTMLElement).style.cursor = "crosshair";
      } catch {}
    } else if (!active && this._onHoverMove) {
      try {
        this.map.un("pointermove" as any, this._onHoverMove);
      } catch {}
      this._onHoverMove = null;
      this.hideHoverRing();
      this.previewLayer.setVisible(false);
      try {
        (this.map.getTargetElement() as HTMLElement).style.cursor = "";
      } catch {}
    }
  }

  destroy() {
    this.setActive(false);
    try {
      this.map.removeLayer(this.previewLayer);
    } catch {}
  }

  // ---------- internals ----------

  private bindHandlers() {
    // Using bound methods so we can unbind cleanly
    this._onDown = this.onPointerDown.bind(this);
    this._onMove = this.onPointerMove.bind(this);
    this._onUp = this.onPointerUp.bind(this);

    this.map.on("pointerdown" as any, this._onDown);
    this.map.on("pointermove" as any, this._onMove);
    this.map.on("pointerup" as any, this._onUp);
    this.map.getViewport().addEventListener("mouseleave", this._onUp as any);
  }

  private unbindHandlers() {
    try {
      this.map.un("pointerdown" as any, this._onDown);
    } catch {}
    try {
      this.map.un("pointermove" as any, this._onMove);
    } catch {}
    try {
      this.map.un("pointerup" as any, this._onUp);
    } catch {}
    try {
      this.map
        .getViewport()
        .removeEventListener("mouseleave", this._onUp as any);
    } catch {}
    this._onDown = this._onMove = this._onUp = null;
  }

  /** Pick an editable annotation polygon at the given pixel (excludes Region boxes). */
  private pickEditableTargetAt(
    pixel: number[]
  ): Feature<Polygon | MultiPolygon> | null {
    let chosen: Feature<Polygon | MultiPolygon> | null = null;

    this.map.forEachFeatureAtPixel(pixel, (f: any) => {
      const geom = f?.getGeometry?.();
      if (!geom) return undefined;

      const isPoly = geom instanceof Polygon || geom instanceof MultiPolygon;
      if (!isPoly) return undefined;

      // Only treat *annotation* polygons as brush targets (not Region rectangles)
      const kind = f.get?.("kind");
      const isAnnotationPoly = kind === "polygon";
      if (!isAnnotationPoly) return undefined;

      chosen = f as Feature<Polygon | MultiPolygon>;
      return true; // stop searching
    });

    return chosen;
  }

  /**
   * Find the closest editable polygon within brush radius for erase operations from outside.
   * This allows "eating" parts of polygons from their edges.
   */
  private findNearestEditableTarget(
    coordinate: number[]
  ): Feature<Polygon | MultiPolygon> | null {
    let nearestFeature: Feature<Polygon | MultiPolygon> | null = null;
    let nearestDistance = Infinity;
    const brushRadius = this.mapBrushRadius();
    const searchRadius = brushRadius * 2; // Search within 2x brush radius

    // Get all vector layers and search for annotation polygons
    this.map.getAllLayers().forEach((layer) => {
      if (!(layer instanceof VectorLayer)) return;

      const source = layer.getSource();
      if (!source) return;

      source.forEachFeature((feature: any) => {
        const geom = feature?.getGeometry?.();
        if (!geom) return;

        const isPoly = geom instanceof Polygon || geom instanceof MultiPolygon;
        if (!isPoly) return;

        // Only consider annotation polygons
        const kind = feature.get?.("kind");
        const isAnnotationPoly = kind === "polygon";
        if (!isAnnotationPoly) return;

        // Calculate distance to polygon edge
        const distance = this.getDistanceToPolygonEdge(coordinate, geom);

        if (distance <= searchRadius && distance < nearestDistance) {
          nearestDistance = distance;
          nearestFeature = feature as Feature<Polygon | MultiPolygon>;
        }
      });
    });

    return nearestFeature;
  }

  /**
   * Calculate the minimum distance from a point to the edge of a polygon
   */
  private getDistanceToPolygonEdge(
    coordinate: number[],
    geometry: Polygon | MultiPolygon
  ): number {
    if (!this.parser) return Infinity;

    try {
      const point = this.parser.read(new Point(coordinate));
      const poly = this.parser.read(geometry);
      return point.distance(poly);
    } catch (error) {
      // Fallback: simple distance calculation
      let minDistance = Infinity;

      const polygons =
        geometry instanceof MultiPolygon ? geometry.getPolygons() : [geometry];

      for (const polygon of polygons) {
        const coordinates = polygon.getCoordinates();
        for (const ring of coordinates) {
          for (let i = 0; i < ring.length - 1; i++) {
            const p1 = ring[i];
            const p2 = ring[i + 1];
            const distance = this.pointToLineSegmentDistance(
              coordinate,
              p1,
              p2
            );
            minDistance = Math.min(minDistance, distance);
          }
        }
      }

      return minDistance;
    }
  }

  /**
   * Calculate distance from a point to a line segment
   */
  private pointToLineSegmentDistance(
    point: number[],
    lineStart: number[],
    lineEnd: number[]
  ): number {
    const A = point[0] - lineStart[0];
    const B = point[1] - lineStart[1];
    const C = lineEnd[0] - lineStart[0];
    const D = lineEnd[1] - lineStart[1];

    const dot = A * C + B * D;
    const lenSq = C * C + D * D;

    if (lenSq === 0) {
      // Line start and end are the same point
      return Math.sqrt(A * A + B * B);
    }

    let param = dot / lenSq;
    param = Math.max(0, Math.min(1, param));

    const xx = lineStart[0] + param * C;
    const yy = lineStart[1] + param * D;

    const dx = point[0] - xx;
    const dy = point[1] - yy;

    return Math.sqrt(dx * dx + dy * dy);
  }

  private onPointerDown(e: any) {
    const oe = e.originalEvent as PointerEvent | MouseEvent;
    const primaryDown =
      "buttons" in oe
        ? ((oe as PointerEvent).buttons & 1) !== 0
        : "which" in oe
        ? (oe as any).which === 1
        : (oe as any).button === 0;

    if (!primaryDown) return;

    // ✅ Auto-pick a polygon target under the cursor at stroke start (or null)
    this.target = this.pickEditableTargetAt(e.pixel);

    // 🆕 If in erase mode and no direct target found, look for nearby polygons
    if (!this.target && this.mode === "erase") {
      this.target = this.findNearestEditableTarget(e.coordinate);
    }

    this.drawing = true;
    this.strokeLine.setCoordinates([e.coordinate]);
    this.previewLayer.getSource()!.clear();
    this.previewLayer.setVisible(true);
    this.opts.setDragPanActive?.(false);
    this.updatePreviewSoon();
  }

  private onPointerMove(e: any) {
    if (!this.drawing) return;
    const coords = this.strokeLine.getCoordinates();
    const last = coords[coords.length - 1];

    // Always add the second coordinate, then use distance check for subsequent ones
    if (
      coords.length === 1 ||
      !last ||
      this.sqDist(last, e.coordinate) > this.minMoveSq()
    ) {
      coords.push(e.coordinate);
      this.strokeLine.setCoordinates(coords);
      this.updatePreviewSoon();
    }
  }

  private onPointerUp = () => {
    if (!this.drawing) return;
    this.drawing = false;
    this.commitStroke();
    this.previewLayer.getSource()!.clear();
    this.previewLayer.setVisible(false);
    this.strokeLine.setCoordinates([]);
    this.opts.setDragPanActive?.(true);
  };

  private cancelStroke() {
    this.drawing = false;
    this.previewLayer.getSource()!.clear();
    this.previewLayer.setVisible(false);
    this.strokeLine.setCoordinates([]);
    this.opts.setDragPanActive?.(true);
  }

  private updatePreviewSoon() {
    if (this.rafPending) return;
    this.rafPending = true;
    requestAnimationFrame(() => {
      this.rafPending = false;
      const strokePoly = this.makeStrokePolygon();
      const src = this.previewLayer.getSource()!;
      src.clear();
      if (this.hoverRing) src.addFeature(this.hoverRing); // keep the ring on top
      if (strokePoly) src.addFeature(new Feature(strokePoly));
    });
  }

  private commitStroke() {
    const strokePoly = this.makeStrokePolygon();
    if (!strokePoly) return;

    if (!this.target) {
      // No selected polygon → create a new one from this stroke
      this.opts.onCreateFromStroke?.(strokePoly);
      return;
    }

    // If there is no target, just end (preview already showed the stroke)
    if (!this.target) {
      // No polygon under cursor when the stroke started:
      // - in add mode: create a new polygon from this stroke
      // - in erase mode: nothing to erase
      if (this.mode === "add") {
        this.opts.onCreateFromStroke?.(strokePoly);
      }
      return;
    }

    const base = this.target.getGeometry();
    if (!base) return;

    try {
      const newGeom = this.applyBoolean(base, strokePoly, this.mode);
      this.target.setGeometry(newGeom);
      this.opts.onCommit?.(this.target);
    } catch (error) {
      console.error("Error applying brush stroke:", error);
    }
  }

  /** Build buffered polygon from the current stroke path. */
  private makeStrokePolygon(): Polygon | null {
    const coords = this.strokeLine.getCoordinates();
    if (coords.length < 2) return null;

    if (!this.parser) return null;

    // Clean up the stroke coordinates to avoid degenerate cases
    const cleanedCoords = this.cleanStrokeCoordinates(coords);
    if (cleanedCoords.length < 2) return null;

    // Create a cleaned LineString
    const cleanedLine = new LineString(cleanedCoords);

    try {
      const jtsLine = this.parser.read(cleanedLine);
      const radius = this.mapBrushRadius();

      // Validate the line geometry before buffering
      const validLine = this.validateAndFixGeometry(jtsLine);

      const jtsStroke = BufferOp.bufferOp(validLine, radius);
      const olGeom = this.parser.write(jtsStroke) as Geometry;

      if (olGeom instanceof Polygon) return olGeom;
      if (olGeom instanceof MultiPolygon) {
        // pick largest for preview
        let best: Polygon | null = null,
          area = -Infinity;
        for (const p of olGeom.getPolygons()) {
          const a = Math.abs(p.getArea?.() ?? 0);
          if (a > area) {
            best = p;
            area = a;
          }
        }
        return best;
      }
    } catch (error) {
      console.warn("Error creating stroke polygon:", error);
      return null;
    }

    return null;
  }

  /**
   * Clean stroke coordinates to remove duplicates and ensure minimum distance
   */
  private cleanStrokeCoordinates(coords: number[][]): number[][] {
    if (coords.length === 0) return coords;

    const cleaned: number[][] = [coords[0]];
    const minDistSq = this.minMoveSq();

    for (let i = 1; i < coords.length; i++) {
      const current = coords[i];
      const last = cleaned[cleaned.length - 1];

      // Only add if it's far enough from the last point
      if (this.sqDist(current, last) >= minDistSq) {
        cleaned.push(current);
      }
    }

    // Ensure we have at least 2 points for a valid line
    if (cleaned.length < 2 && coords.length >= 2) {
      // If cleaning removed too many points, keep first and last
      cleaned.length = 0;
      cleaned.push(coords[0], coords[coords.length - 1]);
    }

    return cleaned;
  }

  /**
   * Validates and fixes a JSTS geometry to avoid topology exceptions
   */
  private validateAndFixGeometry(jtsGeom: any): any {
    if (!jtsGeom) return jtsGeom;

    try {
      // Check if geometry is valid
      const validOp = new IsValidOp(jtsGeom);
      if (validOp.isValid()) {
        return jtsGeom;
      }

      console.warn(
        "Invalid geometry detected, attempting to fix:",
        validOp.getValidationError()
      );

      // Try to fix with zero-buffer operation
      const buffered = BufferOp.bufferOp(jtsGeom, 0);
      if (buffered && new IsValidOp(buffered).isValid()) {
        return buffered;
      }

      // If still invalid, try a small buffer to clean up topology
      const smallBuffer = BufferOp.bufferOp(jtsGeom, 1e-10);
      if (smallBuffer && new IsValidOp(smallBuffer).isValid()) {
        return smallBuffer;
      }

      // Last resort: return original geometry
      return jtsGeom;
    } catch (error) {
      console.warn("Failed to validate/fix geometry:", error);
      return jtsGeom;
    }
  }

  /**
   * Checks if stroke geometry would cause issues and simplifies if needed
   */
  private preprocessStrokeGeometry(strokeGeom: Polygon): Polygon {
    const coords = strokeGeom.getCoordinates();
    if (!coords || coords.length === 0) return strokeGeom;

    // Check for degenerate cases
    const ring = coords[0];
    if (!ring || ring.length < 4) return strokeGeom;

    // Remove consecutive duplicate points
    const cleanedRing = ring.filter((coord, index) => {
      if (index === 0) return true;
      const prev = ring[index - 1];
      const dx = Math.abs(coord[0] - prev[0]);
      const dy = Math.abs(coord[1] - prev[1]);
      return dx > 1e-12 || dy > 1e-12;
    });

    // Ensure we have enough points for a valid polygon
    if (cleanedRing.length < 4) {
      return strokeGeom; // Return original if we can't clean it properly
    }

    // Ensure ring is closed
    const first = cleanedRing[0];
    const last = cleanedRing[cleanedRing.length - 1];
    if (first[0] !== last[0] || first[1] !== last[1]) {
      cleanedRing.push([first[0], first[1]]);
    }

    return new Polygon([cleanedRing]);
  }

  private applyBoolean(
    base: Polygon | MultiPolygon,
    stroke: Polygon,
    mode: Mode
  ): Polygon | MultiPolygon {
    try {
      if (!this.parser) return base;

      // Preprocess stroke geometry to avoid degenerate cases
      const cleanStroke = this.preprocessStrokeGeometry(stroke);

      const A = this.parser.read(base);
      const B = this.parser.read(cleanStroke);

      // Validate both geometries before operation
      const validA = this.validateAndFixGeometry(A);
      const validB = this.validateAndFixGeometry(B);

      // Perform boolean operation with validated geometries
      const out =
        mode === "add"
          ? OverlayOp.overlayOp(validA, validB, OverlayOp.UNION)
          : OverlayOp.overlayOp(validA, validB, OverlayOp.DIFFERENCE);

      // Clean the result
      const cleaned = BufferOp.bufferOp(out, 0);
      const olOut = this.parser.write(cleaned) as Geometry;

      if (olOut instanceof Polygon) return olOut;
      if (olOut instanceof MultiPolygon) return olOut;

      // If empty, return an empty MultiPolygon (host may decide to delete feature).
      return new MultiPolygon([]);
    } catch (error) {
      console.error("Error in boolean operation:", error);

      // Enhanced error reporting for topology exceptions
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      if (errorMessage && errorMessage.includes("TopologyException")) {
        console.warn(
          "Topology exception encountered - this usually indicates self-intersecting or degenerate geometry"
        );
        console.warn("Original stroke coordinates:", stroke.getCoordinates());
        console.warn("Base geometry type:", base.getType());
      }

      // Return original geometry on error
      return base;
    }
  }

  private mapBrushRadius() {
    const res = this.map.getView().getResolution() || 1;
    return Math.max(1e-9, (this.brushPx / 2) * res);
  }

  private minMoveSq() {
    const res = this.map.getView().getResolution() || 1;
    const px = 4 * res; // ~4px
    return px * px;
  }

  private sqDist(a: number[], b: number[]) {
    const dx = a[0] - b[0],
      dy = a[1] - b[1];
    return dx * dx + dy * dy;
  }
}
