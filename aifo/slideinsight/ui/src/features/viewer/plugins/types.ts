// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import type OlMap from "ol/Map";
import type {
  DisplayMetadata,
  SlideMetadata,
} from "@/features/viewer/components/map/types";
import type {
  AnnotationItem,
  LabelId,
} from "@/features/viewer/components/AnnotationPanel";

export interface PluginContext {
  slideUid?: string;
  studyUid?: string;
  caseUid?: string;
  map?: OlMap | null;
  mapRef?: React.RefObject<OlMap | null>;
  pluginManager?: PluginManager;
  slideMetadata?: DisplayMetadata | null;
  rawSlideMetadata?: SlideMetadata | null;
  slideLayer?: any;
  // Annotation-related context
  annotations?: AnnotationItem[];
  selectedAnnotationId?: string | null;
  hoveredAnnotationId?: string | null;
  activeAnnotationLabel?: LabelId | null;
  brushActive?: boolean;
  brushMode?: "add" | "erase";
  brushSizePx?: number;
  // Study annotation settings for mask/vector control
  studyAnnotationSettings?: any;
  studyAnnotationSettingsLoading?: boolean;
  studyAnnotationSettingsError?: any;
  // Viewer settings callback
  onViewerSettingsChange?: (settings: {
    panSensitivity: number;
    zoomSensitivity: number;
    quality: number | undefined;
    showMeasurementBar: boolean;
  }) => void;
  // Region-related context (now managed internally by RegionControlPlugin)
  // regions?: Array<{ id: string; name: string; visible?: boolean }>;
  // selectedRegionId?: string | null;
  // hoveredRegionId?: string | null;
  // isDrawingRegion?: boolean;
  // isEditingRegion?: boolean;
  // Annotation control functions
  onUpdateAnnotations?: (items: AnnotationItem[]) => void;
  onSelectAnnotation?: (id: string) => void;
  onHoverAnnotation?: (id: string | null) => void;
  onActiveLabelChange?: (label: LabelId | null) => void;
  onStartDrawROI?: (mode: "point" | "box" | "polygon", label: LabelId) => void;
  onStopDraw?: () => void;
  onStartBrushAdd?: () => void;
  onStartBrushErase?: () => void;
  onStopBrush?: () => void;
  onBrushSizeChange?: (size: number) => void;
  onHighlightFeatureFromList?: (id: string | null) => void;
  onHighlightGroupFromList?: (ids: string[] | null) => void;
  // Region control functions (now managed internally by RegionControlPlugin)
  // onUpdateRegions?: (items: Array<{ id: string; name: string; visible?: boolean }>) => void;
  // onSelectRegion?: (id: string) => void;
  // onHoverRegion?: (id: string | null) => void;
  // onStartDrawRegion?: () => void;
  // onStopDrawRegion?: () => void;
  // canStartDrawing?: () => boolean;
  // onEditingStateChange?: (isEditing: boolean) => void;
  // onHighlightRegionFromList?: (id: string | null) => void;
}

export interface PluginButtonConfig {
  id: string;
  label: string;
  icon?: React.ComponentType<{ className?: string }>;
  tooltip?: string;
  position?: "left" | "right";
  order?: number;
}

export interface PluginPanelConfig {
  id: string;
  title: string;
  defaultSize?: { width: number; height: number };
  defaultDock?: "free" | "left";
  storageKey?: string;
}

export interface PluginState {
  isOpen: boolean;
  dock: "free" | "left";
  size?: { width: number; height: number };
  customState?: any;
}

export interface PluginAPI {
  context: PluginContext;
  state: PluginState;
  setState: (updates: Partial<PluginState>) => void;
  closePanel: () => void;
}

export interface ViewerPlugin {
  id: string;
  name: string;
  version?: string;

  // Button configuration for the toolbar
  button?: PluginButtonConfig;

  // Panel configuration if the plugin has a panel
  panel?: PluginPanelConfig;

  // Initialize the plugin
  initialize?: (api: PluginAPI) => void | Promise<void>;

  // Cleanup when plugin is destroyed
  destroy?: () => void | Promise<void>;

  // React component for the panel content (if panel is defined)
  PanelComponent?: React.ComponentType<{
    api: PluginAPI;
    onClose: () => void;
  }>;

  // Called when context changes (slide change, metadata loaded, etc.)
  onContextChange?: (context: PluginContext) => void;

  // Called when the plugin button is clicked
  onButtonClick?: (api: PluginAPI) => void;
}

export interface PluginRegistration extends ViewerPlugin {
  dependencies?: string[];
  defaultState?: Partial<PluginState>;
  lifecycle?: {
    onRegister?: (context: PluginContext) => void | Promise<void>;
    onUnregister?: (context: PluginContext) => void | Promise<void>;
    onContextChange?: (context: PluginContext) => void | Promise<void>;
    onSlideChange?: (
      slideUid: string | undefined,
      context: PluginContext
    ) => void | Promise<void>;
    onStudyChange?: (
      studyUid: string | undefined,
      context: PluginContext
    ) => void | Promise<void>;
    onMapReady?: (context: PluginContext) => void | Promise<void>;
    onSlideLoad?: (context: PluginContext) => void | Promise<void>;
    onActivate?: (context: PluginContext) => void | Promise<void>;
    onDeactivate?: (context: PluginContext) => void | Promise<void>;
  };
  order?: number;
}

export interface PluginCommunication {
  sendMessage: (pluginId: string, message: any) => void;
  onMessage: (callback: (from: string, message: any) => void) => void;
  broadcast: (message: any) => void;
  onBroadcast: (
    callback: (sourcePluginId: string, message: any) => void
  ) => () => void;
}

export interface PluginManager {
  registerPlugin: (plugin: ViewerPlugin) => void;
  unregister: (pluginId: string) => void;
  unregisterPlugin: (pluginId: string) => void;
  getPlugin: (pluginId: string) => ViewerPlugin | undefined;
  getAllPlugins: () => ViewerPlugin[];
  getPluginState: (pluginId: string) => PluginState;
  getState: (pluginId: string) => PluginState;
  setPluginState: (pluginId: string, updates: Partial<PluginState>) => void;
  updateState: (pluginId: string, updates: Partial<PluginState>) => void;
  toggle: (pluginId: string) => void;
  closeAll: () => void;
  getCommunication: (pluginId: string) => PluginCommunication;
  updateContext: (context: Partial<PluginContext>) => void;
}
