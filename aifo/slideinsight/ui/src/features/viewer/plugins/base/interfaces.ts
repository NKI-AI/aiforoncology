// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import type OlMap from "ol/Map";
import type { Interaction } from "ol/interaction";
import type VectorSource from "ol/source/Vector";
import type VectorLayer from "ol/layer/Vector";
import type { PluginContext, PluginState, PluginAPI } from "./PluginManager";
import type {
  DisplayMetadata,
  SlideMetadata,
} from "@/features/viewer/components/map/types";

/**
 * Plugin capabilities declaration
 * Used to declare what features a plugin provides
 */
export interface PluginCapabilities {
  /** Plugin can interact with the map (draw, select, modify) */
  hasMapInteractions?: boolean;

  /** Plugin manages its own layers on the map */
  hasLayers?: boolean;

  /** Plugin has a panel UI component */
  hasPanel?: boolean;

  /** Plugin has a toolbar button */
  hasButton?: boolean;

  /** Plugin requires slide metadata to function */
  requiresSlideMetadata?: boolean;

  /** Plugin requires study context to function */
  requiresStudyContext?: boolean;

  /** Plugin can work with annotations */
  canManageAnnotations?: boolean;

  /** Plugin can work with regions */
  canManageRegions?: boolean;

  /** Plugin provides keyboard shortcuts */
  hasKeyboardShortcuts?: boolean;

  /** Plugin persists state to localStorage */
  persistsState?: boolean;

  /** Plugin can export/import data */
  canExportImport?: boolean;
}

/**
 * Plugin button configuration
 */
export interface PluginButtonConfig {
  id: string;
  label: string;
  icon?: React.ComponentType<{ className?: string }>;
  tooltip?: string;
  position: "left" | "right";
  order: number;
  disabled?: boolean;
  hidden?: boolean;
}

/**
 * Plugin panel configuration
 */
export interface PluginPanelConfig {
  id: string;
  title: string;
  defaultSize?: { width: number; height: number };
  defaultDock: "free" | "left" | "right" | "bottom";
  storageKey?: string;
  resizable?: boolean;
  closable?: boolean;
  minimizable?: boolean;
}

/**
 * Plugin lifecycle hooks
 * These are called at specific points in the plugin lifecycle
 */
export interface PluginLifecycleHooks {
  /** Called before plugin initialization */
  onInitialize?: (api: PluginAPI) => Promise<void> | void;

  /** Called before plugin destruction */
  onDestroy?: () => Promise<void> | void;

  /** Called when plugin context changes */
  onContextChange?: (context: PluginContext) => Promise<void> | void;

  /** Called when plugin button is clicked */
  onButtonClick?: (api: PluginAPI) => Promise<void> | void;

  /** Called when plugin panel opens */
  onPanelOpen?: (api: PluginAPI) => Promise<void> | void;

  /** Called when plugin panel closes */
  onPanelClose?: (api: PluginAPI) => Promise<void> | void;

  /** Called when slide changes */
  onSlideChange?: (
    slideUid: string | undefined,
    context: PluginContext
  ) => Promise<void> | void;

  /** Called when study changes */
  onStudyChange?: (
    studyUid: string | undefined,
    context: PluginContext
  ) => Promise<void> | void;
}

/**
 * Core plugin interface
 * All plugins must implement this interface
 */
export interface IPlugin {
  /** Unique plugin identifier */
  readonly id: string;

  /** Human-readable plugin name */
  readonly name: string;

  /** Plugin version */
  readonly version: string;

  /** Plugin capabilities */
  readonly capabilities: PluginCapabilities;

  /** Plugin button configuration (if any) */
  readonly button?: PluginButtonConfig;

  /** Plugin panel configuration (if any) */
  readonly panel?: PluginPanelConfig;

  /** Whether the plugin is initialized */
  readonly isInitialized: boolean;

  /** Whether the plugin is destroyed */
  readonly isDestroyed: boolean;

  /** Initialize the plugin */
  initialize(api: PluginAPI): Promise<void>;

  /** Clean up the plugin */
  destroy(): Promise<void>;

  /** Handle context changes */
  onContextChange(context: PluginContext): Promise<void>;

  /** Handle button click events */
  onButtonClick(api: PluginAPI): Promise<void>;

  /** Get the React component for the plugin panel */
  getPanelComponent():
    | React.ComponentType<{
        api: PluginAPI;
        onClose: () => void;
      }>
    | undefined;
}

/**
 * Context for map interaction plugins
 */
export interface MapInteractionContext {
  map: OlMap;
  slideUid?: string;
  slideMetadata?: DisplayMetadata | null;
  rawSlideMetadata?: SlideMetadata | null;
  slideLayer?: any;
}

/**
 * Interface for plugins that interact with the map
 */
export interface IMapInteractionPlugin extends IPlugin {
  /** Set up map interactions when context is available */
  setupMapInteractions(context: MapInteractionContext): Promise<void>;

  /** Clean up map interactions */
  cleanupMapInteractions(): Promise<void>;
}

/**
 * Context for layer plugins
 */
export interface LayerContext {
  map: OlMap;
  slideUid?: string;
  slideMetadata?: DisplayMetadata | null;
  rawSlideMetadata?: SlideMetadata | null;
}

/**
 * Interface for plugins that manage layers
 */
export interface ILayerPlugin extends IPlugin {
  /** Create and add layers to the map */
  createLayers(context: LayerContext): Promise<void>;

  /** Clean up and remove layers from the map */
  cleanupLayers(): Promise<void>;
}

/**
 * Interface for plugins that manage annotations
 */
export interface IAnnotationPlugin extends IPlugin {
  /** Get current annotations managed by this plugin */
  getAnnotations(): any[];

  /** Create a new annotation */
  createAnnotation(annotation: any): Promise<void>;

  /** Update an existing annotation */
  updateAnnotation(id: string, updates: any): Promise<void>;

  /** Delete an annotation */
  deleteAnnotation(id: string): Promise<void>;

  /** Select an annotation */
  selectAnnotation(id: string | null): void;

  /** Start drawing a new annotation */
  startDrawing(type: string, options?: any): void;

  /** Stop drawing */
  stopDrawing(): void;
}

/**
 * Interface for plugins that manage regions
 */
export interface IRegionPlugin extends IPlugin {
  /** Get current regions managed by this plugin */
  getRegions(): any[];

  /** Create a new region */
  createRegion(region: any): Promise<void>;

  /** Update an existing region */
  updateRegion(id: string, updates: any): Promise<void>;

  /** Delete a region */
  deleteRegion(id: string): Promise<void>;

  /** Select a region */
  selectRegion(id: string | null): void;

  /** Start drawing a new region */
  startDrawingRegion(type: string, options?: any): void;

  /** Stop drawing region */
  stopDrawingRegion(): void;
}

/**
 * Interface for plugins that export/import data
 */
export interface IDataPlugin extends IPlugin {
  /** Export plugin data */
  exportData(format: string): Promise<any>;

  /** Import plugin data */
  importData(data: any, format: string): Promise<void>;

  /** Get supported export formats */
  getSupportedExportFormats(): string[];

  /** Get supported import formats */
  getSupportedImportFormats(): string[];
}

/**
 * Interface for plugins with keyboard shortcuts
 */
export interface IKeyboardPlugin extends IPlugin {
  /** Get keyboard shortcuts provided by this plugin */
  getKeyboardShortcuts(): Array<{
    key: string;
    description: string;
    handler: () => void;
    modifiers?: Array<"ctrl" | "alt" | "shift" | "meta">;
  }>;

  /** Handle keyboard events */
  handleKeyboardEvent(event: KeyboardEvent): boolean;
}

/**
 * Interface for plugins that persist state
 */
export interface IStatefulPlugin extends IPlugin {
  /** Save plugin state to storage */
  saveState(): Promise<void>;

  /** Load plugin state from storage */
  loadState(): Promise<void>;

  /** Reset plugin state to defaults */
  resetState(): Promise<void>;

  /** Get current plugin state */
  getPluginState(): any;

  /** Set plugin state */
  setPluginState(state: any): void;
}

/**
 * Plugin configuration options
 */
export interface PluginConfig {
  /** Plugin capabilities */
  capabilities?: Partial<PluginCapabilities>;

  /** Button configuration */
  button?: PluginButtonConfig;

  /** Panel configuration */
  panel?: PluginPanelConfig;

  /** Lifecycle hooks */
  lifecycleHooks?: Partial<PluginLifecycleHooks>;

  /** Plugin dependencies (other plugin IDs this plugin requires) */
  dependencies?: string[];

  /** Plugin conflicts (plugin IDs that cannot be used with this plugin) */
  conflicts?: string[];

  /** Minimum required context properties */
  requiredContext?: Array<keyof PluginContext>;
}

/**
 * Plugin metadata for registration
 */
export interface PluginMetadata {
  /** Plugin author */
  author?: string;

  /** Plugin description */
  description?: string;

  /** Plugin homepage URL */
  homepage?: string;

  /** Plugin license */
  license?: string;

  /** Plugin keywords */
  keywords?: string[];

  /** Plugin category */
  category?:
    | "annotation"
    | "region"
    | "analysis"
    | "visualization"
    | "utility"
    | "other";

  /** Whether this is a core plugin (shipped with SlideInsight) */
  isCore?: boolean;

  /** Plugin settings schema (JSON Schema) */
  settingsSchema?: any;

  /** Default plugin settings */
  defaultSettings?: any;
}

/**
 * Plugin registry entry
 */
export interface PluginRegistryEntry {
  /** Plugin instance */
  plugin: IPlugin;

  /** Plugin metadata */
  metadata: PluginMetadata;

  /** Plugin configuration */
  config: PluginConfig;

  /** Registration timestamp */
  registeredAt: Date;

  /** Whether plugin is currently active */
  isActive: boolean;

  /** Plugin load order */
  loadOrder: number;
}

/**
 * Plugin manager interface
 */
export interface IPluginManager {
  /** Register a plugin */
  registerPlugin(
    plugin: IPlugin,
    config?: PluginConfig,
    metadata?: PluginMetadata
  ): Promise<void>;

  /** Unregister a plugin */
  unregisterPlugin(pluginId: string): Promise<void>;

  /** Get a plugin by ID */
  getPlugin(pluginId: string): IPlugin | undefined;

  /** Get all registered plugins */
  getAllPlugins(): IPlugin[];

  /** Get plugins by capability */
  getPluginsByCapability(capability: keyof PluginCapabilities): IPlugin[];

  /** Get plugin state */
  getPluginState(pluginId: string): PluginState;

  /** Set plugin state */
  setPluginState(pluginId: string, updates: Partial<PluginState>): void;

  /** Update context for all plugins */
  updateContext(context: Partial<PluginContext>): Promise<void>;

  /** Enable/disable a plugin */
  setPluginActive(pluginId: string, active: boolean): Promise<void>;

  /** Check if plugin can be loaded (dependencies, conflicts) */
  canLoadPlugin(pluginId: string): boolean;

  /** Get plugin load order */
  getLoadOrder(): string[];

  /** Validate plugin configuration */
  validatePlugin(plugin: IPlugin, config?: PluginConfig): boolean;
}
