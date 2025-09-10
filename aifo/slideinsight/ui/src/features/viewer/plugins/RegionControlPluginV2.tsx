// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useState, useRef } from "react";
import { RectangleStackIcon } from "@heroicons/react/24/solid";
import {
  MapLayerPlugin,
  PluginAPI,
  PluginContext,
  MapInteractionContext,
  LayerContext,
} from "./base";
import RegionPanel, {
  type RegionItem,
} from "@/features/viewer/components/RegionPanel";
import { useRegionSyncV2 } from "@/features/viewer/hooks/useRegionSyncV2";
import {
  loadRegionSettings,
  saveRegionSettings,
  resetRegionSettings,
  type RegionSettings,
  type RegionSettingsPartial,
} from "@/features/viewer/utils/regionSettingsStorage";

import VectorSource from "ol/source/Vector";
import VectorLayer from "ol/layer/Vector";
import Draw, { createBox } from "ol/interaction/Draw";
import Modify from "ol/interaction/Modify";
import Snap from "ol/interaction/Snap";
import Transform from "ol-ext/interaction/Transform";
import Feature from "ol/Feature";
import Polygon from "ol/geom/Polygon";
import { Style, Stroke, Fill, Text, RegularShape } from "ol/style";
import GeoJSON from "ol/format/GeoJSON";

/**
 * Panel component for the region control plugin
 */
interface RegionPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
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

  // Get plugin instance to access its methods
  const pluginInstance = api.context.pluginManager?.getPlugin(
    "region-control-v2"
  ) as RegionControlPluginV2 | undefined;

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  // Ensure dock position is compatible with RegionControlPanel
  const compatibleDock: "free" | "left" =
    api.state.dock === "left" ? "left" : "free";

  const handleUpdateRegions = useCallback(
    (updatedRegions: RegionItem[]) => {
      setRegions(updatedRegions);
      pluginInstance?.updateRegions(updatedRegions);
    },
    [pluginInstance]
  );

  const handleStartDrawRegion = useCallback(() => {
    pluginInstance?.startDrawing();
    setIsDrawingRegion(true);
  }, [pluginInstance]);

  const handleStopDrawRegion = useCallback(() => {
    pluginInstance?.stopDrawing();
    setIsDrawingRegion(false);
  }, [pluginInstance]);

  const handleUpdateRegionDisplaySettings = useCallback(
    (settings: RegionSettingsPartial) => {
      const updated = saveRegionSettings(settings);
      setRegionDisplaySettings(updated);
      pluginInstance?.updateDisplaySettings(settings);
    },
    [pluginInstance]
  );

  const handleResetRegionDisplaySettings = useCallback(() => {
    const reset = resetRegionSettings();
    setRegionDisplaySettings(reset);
    pluginInstance?.resetDisplaySettings();
  }, [pluginInstance]);

  return (
    <RegionPanel
      isOpen={api.state.isOpen}
      onClose={onClose}
      regions={regions}
      onUpdateRegions={handleUpdateRegions}
      selectedId={selectedRegionId}
      onSelect={setSelectedRegionId}
      hoveredId={hoveredRegionId}
      onHoverIdChange={setHoveredRegionId}
      dockOverride={compatibleDock}
      onDockChange={handleDockChange}
      onStartDrawRegion={handleStartDrawRegion}
      onStopDraw={handleStopDrawRegion}
      isDrawing={isDrawingRegion}
      canStartDrawing={() => !isEditingRegion}
      onEditingStateChange={setIsEditingRegion}
      regionDisplaySettings={regionDisplaySettings}
      onUpdateRegionDisplaySettings={handleUpdateRegionDisplaySettings}
      onResetRegionDisplaySettings={handleResetRegionDisplaySettings}
    />
  );
}

/**
 * Class-based region control plugin
 * Demonstrates MapLayerPlugin with both map interactions and layer management
 */
export class RegionControlPluginV2 extends MapLayerPlugin {
  public readonly id = "region-control-v2";
  public readonly name = "Regions";
  public readonly version = "2.0.0";

  // Layer management
  private persistedRegionLayer?: VectorLayer<VectorSource>;
  private tempRegionLayer?: VectorLayer<VectorSource>;

  // Interaction management
  private drawInteraction?: Draw;
  private modifyInteraction?: Modify;
  private snapInteraction?: Snap;
  private transformInteraction?: Transform;

  // State
  private regionIdCounter = 1;
  private selectedRegionId: string | null = null;
  private hoveredRegionId: string | null = null;
  private isDrawing = false;
  private regionDisplaySettings: RegionSettings = loadRegionSettings();

  // Coordinate transforms
  private transformMapToPixelCoordinates?: (geoJson: any) => any;
  private transformPixelToMapCoordinates?: (geoJson: any) => any;

  protected setupDefaultCapabilities(): void {
    // Declare plugin capabilities
    this.addCapability("hasMapInteractions", true);
    this.addCapability("hasLayers", true);
    this.addCapability("hasPanel", true);
    this.addCapability("hasButton", true);
    this.addCapability("canManageRegions", true);
    this.addCapability("requiresSlideMetadata", true);
    this.addCapability("persistsState", true);

    // Configure button
    this.setButton({
      id: "region-control-button-v2",
      label: "Regions",
      icon: RectangleStackIcon,
      tooltip: "Create and manage regions of interest",
      position: "right",
      order: 4,
    });

    // Configure panel
    this.setPanel({
      id: "region-control-panel-v2",
      title: "Regions",
      defaultSize: { width: 320, height: 400 },
      defaultDock: "left",
      storageKey: "regionPanel",
      resizable: true,
      closable: true,
    });
  }

  protected async onInitialize(api: PluginAPI): Promise<void> {
    console.log("RegionControlPluginV2: Initializing plugin");

    // Set up coordinate transforms if slide metadata is available
    this.setupCoordinateTransforms(api.context);
  }

  protected async onDestroy(): Promise<void> {
    console.log("RegionControlPluginV2: Destroying plugin");
    // Cleanup is handled by parent class
  }

  protected async handleContextChange(context: PluginContext): Promise<void> {
    console.log("RegionControlPluginV2: Context changed", context);

    // Update coordinate transforms when slide metadata changes
    this.setupCoordinateTransforms(context);
  }

  public async createLayers(context: LayerContext): Promise<void> {
    const { map } = context;

    // Create persisted region layer
    const persistedSource = new VectorSource({ wrapX: false });
    this.persistedRegionLayer = new VectorLayer({
      source: persistedSource,
      style: this.getRegionStyle.bind(this),
    });
    this.addLayer(this.persistedRegionLayer, 1600);

    // Create temporary region layer for drawing
    const tempSource = new VectorSource({ wrapX: false });
    this.tempRegionLayer = new VectorLayer({
      source: tempSource,
      style: this.getTempRegionStyle.bind(this),
    });
    this.addLayer(this.tempRegionLayer, 2100);

    console.log("RegionControlPluginV2: Created layers");
  }

  public async cleanupLayers(): Promise<void> {
    // Parent class handles layer cleanup
    this.persistedRegionLayer = undefined;
    this.tempRegionLayer = undefined;
    console.log("RegionControlPluginV2: Cleaned up layers");
  }

  public async setupMapInteractions(
    context: MapInteractionContext
  ): Promise<void> {
    const { map } = context;

    // Create modify interaction for existing regions
    if (this.persistedRegionLayer) {
      this.modifyInteraction = new Modify({
        source: this.persistedRegionLayer.getSource()!,
      });
      this.addInteraction(this.modifyInteraction);

      // Create snap interaction
      this.snapInteraction = new Snap({
        source: this.persistedRegionLayer.getSource()!,
      });
      this.addInteraction(this.snapInteraction);

      // Create transform interaction
      this.transformInteraction = new Transform({
        layers: (layer) => layer === this.persistedRegionLayer,
        filter: (feature) =>
          String(feature.getId?.() ?? "") === this.selectedRegionId,
        translate: true,
        scale: true,
        stretch: true,
        rotate: false,
        hitTolerance: 10,
      });
      this.addInteraction(this.transformInteraction);
    }

    // Set up map event handlers
    this.setupMapEventHandlers();

    console.log("RegionControlPluginV2: Set up map interactions");
  }

  public async cleanupMapInteractions(): Promise<void> {
    // Parent class handles interaction cleanup
    this.drawInteraction = undefined;
    this.modifyInteraction = undefined;
    this.snapInteraction = undefined;
    this.transformInteraction = undefined;
    console.log("RegionControlPluginV2: Cleaned up map interactions");
  }

  protected createPanelComponent(): React.ComponentType<{
    api: PluginAPI;
    onClose: () => void;
  }> {
    return RegionPluginPanel;
  }

  // Public API methods for region management
  public startDrawing(): void {
    if (this.isDrawing || !this.tempRegionLayer) return;

    const map = this.getMap();
    if (!map) return;

    // Remove existing draw interaction
    if (this.drawInteraction) {
      this.removeInteraction(this.drawInteraction);
    }

    // Create new draw interaction
    this.drawInteraction = new Draw({
      source: this.tempRegionLayer.getSource()!,
      type: "Circle",
      geometryFunction: createBox(),
    });

    this.drawInteraction.on("drawend", this.handleDrawEnd.bind(this));
    this.addInteraction(this.drawInteraction);

    this.isDrawing = true;
    console.log("RegionControlPluginV2: Started drawing");
  }

  public stopDrawing(): void {
    if (!this.isDrawing) return;

    if (this.drawInteraction) {
      this.removeInteraction(this.drawInteraction);
      this.drawInteraction = undefined;
    }

    this.isDrawing = false;
    console.log("RegionControlPluginV2: Stopped drawing");
  }

  public updateRegions(regions: RegionItem[]): void {
    // Update region features on the map
    if (!this.persistedRegionLayer) return;

    const source = this.persistedRegionLayer.getSource()!;

    // Update feature properties
    regions.forEach((region) => {
      const features = source.getFeatures();
      const feature = features.find((f) => String(f.getId()) === region.id);
      if (feature) {
        feature.set("name", region.name);
        feature.set("visible", region.visible);
        feature.changed();
      }
    });

    this.persistedRegionLayer.changed();
    console.log("RegionControlPluginV2: Updated regions");
  }

  public updateDisplaySettings(settings: RegionSettingsPartial): void {
    this.regionDisplaySettings = saveRegionSettings(settings);

    // Force style refresh
    if (this.persistedRegionLayer) {
      this.persistedRegionLayer.changed();
    }
    if (this.tempRegionLayer) {
      this.tempRegionLayer.changed();
    }

    console.log("RegionControlPluginV2: Updated display settings");
  }

  public resetDisplaySettings(): void {
    this.regionDisplaySettings = resetRegionSettings();

    // Force style refresh
    if (this.persistedRegionLayer) {
      this.persistedRegionLayer.changed();
    }
    if (this.tempRegionLayer) {
      this.tempRegionLayer.changed();
    }

    console.log("RegionControlPluginV2: Reset display settings");
  }

  // Private helper methods
  private setupCoordinateTransforms(context: PluginContext): void {
    const slideMpp = context.rawSlideMetadata?.slideMpp;
    if (!slideMpp) return;

    this.transformMapToPixelCoordinates = (geoJsonData: any) => {
      if (!geoJsonData?.features || !slideMpp) {
        return geoJsonData;
      }

      const transformCoordinate = (coord: number[]): [number, number] => {
        if (coord.length < 2) return [0, 0];
        const pixelX = coord[0] / (slideMpp * 1e-6);
        const pixelY = -coord[1] / (slideMpp * 1e-6);
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

    this.transformPixelToMapCoordinates = (geoJsonData: any) => {
      if (!geoJsonData?.features || !slideMpp) {
        return geoJsonData;
      }

      const transformCoordinate = (coord: number[]): [number, number] => {
        if (coord.length < 2) return [0, 0];
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
  }

  private setupMapEventHandlers(): void {
    const map = this.getMap();
    if (!map) return;

    // Set up hover and click handlers
    map.on("pointermove", this.handlePointerMove.bind(this));
    map.on("singleclick", this.handleSingleClick.bind(this));
  }

  private handlePointerMove(event: any): void {
    const map = this.getMap();
    if (!map) return;

    let foundId: string | null = null;
    map.forEachFeatureAtPixel(event.pixel, (feature) => {
      const id = feature.getId?.();
      if (id != null && String(id).startsWith("region-")) {
        foundId = String(id);
        return true;
      }
      return false;
    });

    if (this.hoveredRegionId !== foundId) {
      this.hoveredRegionId = foundId;
      if (this.persistedRegionLayer) {
        this.persistedRegionLayer.changed();
      }
    }
  }

  private handleSingleClick(event: any): void {
    const map = this.getMap();
    if (!map) return;

    const hit = map.forEachFeatureAtPixel(
      event.pixel,
      (f) => f as Feature | undefined
    );
    const id = hit?.getId?.();
    if (!id) return;

    const idStr = String(id);
    if (idStr.startsWith("region-")) {
      this.selectedRegionId = idStr;
      if (this.persistedRegionLayer) {
        this.persistedRegionLayer.changed();
      }
    }
  }

  private handleDrawEnd(event: any): void {
    const feature = event.feature as Feature;
    const tempId = `region-${Date.now()}-${Math.random()
      .toString(36)
      .substring(2, 9)}`;

    // Set feature properties
    feature.setId(tempId);
    feature.set("name", `Region ${this.regionIdCounter++}`);
    feature.set("visible", true);
    feature.set("kind", "box");

    // Move to persisted layer
    if (this.persistedRegionLayer && this.tempRegionLayer) {
      const persistedSource = this.persistedRegionLayer.getSource()!;
      const tempSource = this.tempRegionLayer.getSource()!;

      persistedSource.addFeature(feature);
      tempSource.removeFeature(feature);
    }

    console.log("RegionControlPluginV2: Created region:", tempId);
    this.stopDrawing();
  }

  private getRegionStyle(feature: any): Style {
    const id = String(feature.getId?.() ?? "");
    const name = feature.get?.("name") || `Region ${id}`;

    const isSelected = id === this.selectedRegionId;
    const isHovered = id === this.hoveredRegionId;

    const settings = this.regionDisplaySettings;

    const createFill = (baseColor: string, opacity: number) => {
      if (opacity === 0) return undefined;
      const opacityHex = Math.round(opacity * 255)
        .toString(16)
        .padStart(2, "0");
      return new Fill({ color: `${baseColor}${opacityHex}` });
    };

    const selectedOpacity = Math.min(1, settings.fillOpacity + 0.15);
    const hoverOpacity = Math.min(1, settings.fillOpacity + 0.13);

    let strokeColor = "#8b5cf6";
    let fillOpacity = settings.fillOpacity;
    let strokeWidth = settings.strokeWidth;

    if (isSelected) {
      strokeColor = "#0ea5e9";
      fillOpacity = selectedOpacity;
      strokeWidth += 2;
    } else if (isHovered) {
      strokeColor = "#f59e0b";
      fillOpacity = hoverOpacity;
      strokeWidth += 1;
    }

    return new Style({
      stroke: new Stroke({ color: strokeColor, width: strokeWidth }),
      fill: createFill(strokeColor, fillOpacity),
      text: new Text({
        text: name,
        font: "14px Arial, sans-serif",
        fill: new Fill({ color: strokeColor }),
        stroke: new Stroke({ color: "#ffffff", width: 2 }),
        textAlign: "center",
        textBaseline: "middle",
      }),
    });
  }

  private getTempRegionStyle(feature: any): Style {
    const name = feature.get("name") || "Drawing...";
    const settings = this.regionDisplaySettings;

    return new Style({
      stroke: new Stroke({ color: "#8b5cf6", width: settings.strokeWidth }),
      fill:
        settings.fillOpacity > 0
          ? new Fill({
              color: `#8b5cf6${Math.round(settings.fillOpacity * 255)
                .toString(16)
                .padStart(2, "0")}`,
            })
          : undefined,
      text: new Text({
        text: name,
        font: "14px Arial, sans-serif",
        fill: new Fill({ color: "#8b5cf6" }),
        stroke: new Stroke({ color: "#ffffff", width: 2 }),
        textAlign: "center",
        textBaseline: "middle",
      }),
    });
  }
}

// Create and export plugin instance
export const regionControlPluginV2 = new RegionControlPluginV2();
