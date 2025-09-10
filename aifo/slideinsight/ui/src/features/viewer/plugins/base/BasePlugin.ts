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
  IPlugin,
  IMapInteractionPlugin,
  ILayerPlugin,
  PluginCapabilities,
  PluginButtonConfig,
  PluginPanelConfig,
  PluginLifecycleHooks,
  MapInteractionContext,
  LayerContext,
} from "./interfaces";

/**
 * Abstract base class for all SlideInsight plugins.
 * Provides common functionality and enforces plugin interface compliance.
 */
export abstract class BasePlugin implements IPlugin {
  // Required plugin identification
  public abstract readonly id: string;
  public abstract readonly name: string;
  public abstract readonly version: string;

  // Plugin configuration
  protected _capabilities: PluginCapabilities = {};
  protected _button?: PluginButtonConfig;
  protected _panel?: PluginPanelConfig;
  protected _lifecycleHooks: PluginLifecycleHooks = {};

  // Plugin state and context
  protected _api?: PluginAPI;
  protected _context: PluginContext = {};
  protected _state: PluginState = {
    isOpen: false,
    dock: "free",
  };

  // Internal state
  private _initialized = false;
  private _destroyed = false;

  constructor() {
    this.setupDefaultCapabilities();
  }

  // Getters for plugin configuration
  public get capabilities(): PluginCapabilities {
    return { ...this._capabilities };
  }

  public get button(): PluginButtonConfig | undefined {
    return this._button;
  }

  public get panel(): PluginPanelConfig | undefined {
    return this._panel;
  }

  public get isInitialized(): boolean {
    return this._initialized;
  }

  public get isDestroyed(): boolean {
    return this._destroyed;
  }

  // Abstract methods that must be implemented by subclasses
  protected abstract setupDefaultCapabilities(): void;

  /**
   * Initialize the plugin. Called when the plugin is registered.
   */
  public async initialize(api: PluginAPI): Promise<void> {
    if (this._initialized) {
      console.warn(`Plugin ${this.id} is already initialized`);
      return;
    }

    this._api = api;
    this._context = api.context;
    this._state = api.state;

    try {
      // Call lifecycle hook
      if (this._lifecycleHooks.onInitialize) {
        await this._lifecycleHooks.onInitialize(api);
      }

      // Call subclass implementation
      await this.onInitialize(api);

      this._initialized = true;
      console.log(`Plugin ${this.id} initialized successfully`);
    } catch (error) {
      console.error(`Failed to initialize plugin ${this.id}:`, error);
      throw error;
    }
  }

  /**
   * Clean up the plugin. Called when the plugin is unregistered.
   */
  public async destroy(): Promise<void> {
    if (this._destroyed) {
      console.warn(`Plugin ${this.id} is already destroyed`);
      return;
    }

    try {
      // Call lifecycle hook
      if (this._lifecycleHooks.onDestroy) {
        await this._lifecycleHooks.onDestroy();
      }

      // Call subclass implementation
      await this.onDestroy();

      this._destroyed = true;
      this._initialized = false;
      console.log(`Plugin ${this.id} destroyed successfully`);
    } catch (error) {
      console.error(`Failed to destroy plugin ${this.id}:`, error);
      throw error;
    }
  }

  /**
   * Handle context changes (slide change, metadata loaded, etc.)
   */
  public async onContextChange(context: PluginContext): Promise<void> {
    this._context = context;

    try {
      // Call lifecycle hook
      if (this._lifecycleHooks.onContextChange) {
        await this._lifecycleHooks.onContextChange(context);
      }

      // Call subclass implementation
      await this.handleContextChange(context);
    } catch (error) {
      console.error(
        `Error handling context change in plugin ${this.id}:`,
        error
      );
    }
  }

  /**
   * Handle button click events
   */
  public async onButtonClick(api: PluginAPI): Promise<void> {
    this._api = api;

    try {
      // Call lifecycle hook
      if (this._lifecycleHooks.onButtonClick) {
        await this._lifecycleHooks.onButtonClick(api);
      }

      // Call subclass implementation
      await this.handleButtonClick(api);
    } catch (error) {
      console.error(`Error handling button click in plugin ${this.id}:`, error);
    }
  }

  /**
   * Get the React component for the plugin panel
   */
  public getPanelComponent():
    | React.ComponentType<{
        api: PluginAPI;
        onClose: () => void;
      }>
    | undefined {
    return this.createPanelComponent();
  }

  // Protected methods for subclasses to override
  protected async onInitialize(api: PluginAPI): Promise<void> {
    // Default implementation - subclasses can override
  }

  protected async onDestroy(): Promise<void> {
    // Default implementation - subclasses can override
  }

  protected async handleContextChange(context: PluginContext): Promise<void> {
    // Default implementation - subclasses can override
  }

  protected async handleButtonClick(api: PluginAPI): Promise<void> {
    // Default implementation: toggle panel if it exists
    if (this._panel) {
      api.setState({ isOpen: !api.state.isOpen });
    }
  }

  protected createPanelComponent():
    | React.ComponentType<{
        api: PluginAPI;
        onClose: () => void;
      }>
    | undefined {
    // Default implementation - subclasses should override if they have a panel
    return undefined;
  }

  // Utility methods for subclasses
  protected setState(updates: Partial<PluginState>): void {
    if (this._api) {
      this._api.setState(updates);
    }
  }

  protected closePanel(): void {
    if (this._api) {
      this._api.closePanel();
    }
  }

  protected getContext(): PluginContext {
    return this._context;
  }

  protected getMap(): OlMap | null {
    return this._context.map || null;
  }

  // Validation methods
  protected validateCapabilities(): boolean {
    // Validate that declared capabilities match actual implementation
    if (
      this._capabilities.hasMapInteractions &&
      !this.implementsMapInteractions()
    ) {
      console.error(
        `Plugin ${this.id} declares map interactions but doesn't implement IMapInteractionPlugin`
      );
      return false;
    }

    if (this._capabilities.hasLayers && !this.implementsLayers()) {
      console.error(
        `Plugin ${this.id} declares layers but doesn't implement ILayerPlugin`
      );
      return false;
    }

    return true;
  }

  private implementsMapInteractions(): boolean {
    return (
      "setupMapInteractions" in this &&
      typeof (this as any).setupMapInteractions === "function"
    );
  }

  private implementsLayers(): boolean {
    return (
      "createLayers" in this && typeof (this as any).createLayers === "function"
    );
  }

  // Configuration methods for subclasses
  protected setButton(config: PluginButtonConfig): void {
    this._button = config;
  }

  protected setPanel(config: PluginPanelConfig): void {
    this._panel = config;
  }

  protected setLifecycleHooks(hooks: Partial<PluginLifecycleHooks>): void {
    this._lifecycleHooks = { ...this._lifecycleHooks, ...hooks };
  }

  protected addCapability(
    capability: keyof PluginCapabilities,
    value: boolean = true
  ): void {
    this._capabilities[capability] = value;
  }
}

/**
 * Abstract base class for plugins that interact with the map
 */
export abstract class MapInteractionPlugin
  extends BasePlugin
  implements IMapInteractionPlugin
{
  protected interactions: Interaction[] = [];
  protected mapContext?: MapInteractionContext;

  protected setupDefaultCapabilities(): void {
    this._capabilities.hasMapInteractions = true;
  }

  public async onContextChange(context: PluginContext): Promise<void> {
    await super.onContextChange(context);

    if (context.map) {
      this.mapContext = {
        map: context.map,
        slideUid: context.slideUid,
        slideMetadata: context.slideMetadata,
        rawSlideMetadata: context.rawSlideMetadata,
        slideLayer: context.slideLayer,
      };

      await this.setupMapInteractions(this.mapContext);
    }
  }

  public async destroy(): Promise<void> {
    await this.cleanupMapInteractions();
    await super.destroy();
  }

  // Abstract methods for map interaction plugins
  public abstract setupMapInteractions(
    context: MapInteractionContext
  ): Promise<void>;
  public abstract cleanupMapInteractions(): Promise<void>;

  // Utility methods for map interactions
  protected addInteraction(interaction: Interaction): void {
    const map = this.getMap();
    if (map) {
      map.addInteraction(interaction);
      this.interactions.push(interaction);
    }
  }

  protected removeInteraction(interaction: Interaction): void {
    const map = this.getMap();
    if (map) {
      map.removeInteraction(interaction);
      const index = this.interactions.indexOf(interaction);
      if (index > -1) {
        this.interactions.splice(index, 1);
      }
    }
  }

  protected removeAllInteractions(): void {
    const map = this.getMap();
    if (map) {
      this.interactions.forEach((interaction) => {
        map.removeInteraction(interaction);
      });
      this.interactions = [];
    }
  }
}

/**
 * Abstract base class for plugins that manage layers
 */
export abstract class LayerPlugin extends BasePlugin implements ILayerPlugin {
  protected layers: VectorLayer<VectorSource>[] = [];
  protected layerContext?: LayerContext;

  protected setupDefaultCapabilities(): void {
    this._capabilities.hasLayers = true;
  }

  public async onContextChange(context: PluginContext): Promise<void> {
    await super.onContextChange(context);

    if (context.map) {
      this.layerContext = {
        map: context.map,
        slideUid: context.slideUid,
        slideMetadata: context.slideMetadata,
        rawSlideMetadata: context.rawSlideMetadata,
      };

      await this.createLayers(this.layerContext);
    }
  }

  public async destroy(): Promise<void> {
    await this.cleanupLayers();
    await super.destroy();
  }

  // Abstract methods for layer plugins
  public abstract createLayers(context: LayerContext): Promise<void>;
  public abstract cleanupLayers(): Promise<void>;

  // Utility methods for layer management
  protected addLayer(layer: VectorLayer<VectorSource>, zIndex?: number): void {
    const map = this.getMap();
    if (map) {
      if (zIndex !== undefined) {
        layer.setZIndex(zIndex);
      }
      map.addLayer(layer);
      this.layers.push(layer);
    }
  }

  protected removeLayer(layer: VectorLayer<VectorSource>): void {
    const map = this.getMap();
    if (map) {
      map.removeLayer(layer);
      const index = this.layers.indexOf(layer);
      if (index > -1) {
        this.layers.splice(index, 1);
      }
    }
  }

  protected removeAllLayers(): void {
    const map = this.getMap();
    if (map) {
      this.layers.forEach((layer) => {
        map.removeLayer(layer);
      });
      this.layers = [];
    }
  }
}

/**
 * Abstract base class for plugins that both interact with the map and manage layers
 */
export abstract class MapLayerPlugin
  extends BasePlugin
  implements IMapInteractionPlugin, ILayerPlugin
{
  protected interactions: Interaction[] = [];
  protected layers: VectorLayer<VectorSource>[] = [];
  protected mapContext?: MapInteractionContext;
  protected layerContext?: LayerContext;

  protected setupDefaultCapabilities(): void {
    this._capabilities.hasMapInteractions = true;
    this._capabilities.hasLayers = true;
  }

  public async onContextChange(context: PluginContext): Promise<void> {
    await super.onContextChange(context);

    if (context.map) {
      this.mapContext = {
        map: context.map,
        slideUid: context.slideUid,
        slideMetadata: context.slideMetadata,
        rawSlideMetadata: context.rawSlideMetadata,
        slideLayer: context.slideLayer,
      };

      this.layerContext = {
        map: context.map,
        slideUid: context.slideUid,
        slideMetadata: context.slideMetadata,
        rawSlideMetadata: context.rawSlideMetadata,
      };

      await this.createLayers(this.layerContext);
      await this.setupMapInteractions(this.mapContext);
    }
  }

  public async destroy(): Promise<void> {
    await this.cleanupMapInteractions();
    await this.cleanupLayers();
    await super.destroy();
  }

  // Abstract methods that must be implemented
  public abstract setupMapInteractions(
    context: MapInteractionContext
  ): Promise<void>;
  public abstract cleanupMapInteractions(): Promise<void>;
  public abstract createLayers(context: LayerContext): Promise<void>;
  public abstract cleanupLayers(): Promise<void>;

  // Utility methods from both base classes
  protected addInteraction(interaction: Interaction): void {
    const map = this.getMap();
    if (map) {
      map.addInteraction(interaction);
      this.interactions.push(interaction);
    }
  }

  protected removeInteraction(interaction: Interaction): void {
    const map = this.getMap();
    if (map) {
      map.removeInteraction(interaction);
      const index = this.interactions.indexOf(interaction);
      if (index > -1) {
        this.interactions.splice(index, 1);
      }
    }
  }

  protected removeAllInteractions(): void {
    const map = this.getMap();
    if (map) {
      this.interactions.forEach((interaction) => {
        map.removeInteraction(interaction);
      });
      this.interactions = [];
    }
  }

  protected addLayer(layer: VectorLayer<VectorSource>, zIndex?: number): void {
    const map = this.getMap();
    if (map) {
      if (zIndex !== undefined) {
        layer.setZIndex(zIndex);
      }
      map.addLayer(layer);
      this.layers.push(layer);
    }
  }

  protected removeLayer(layer: VectorLayer<VectorSource>): void {
    const map = this.getMap();
    if (map) {
      map.removeLayer(layer);
      const index = this.layers.indexOf(layer);
      if (index > -1) {
        this.layers.splice(index, 1);
      }
    }
  }

  protected removeAllLayers(): void {
    const map = this.getMap();
    if (map) {
      this.layers.forEach((layer) => {
        map.removeLayer(layer);
      });
      this.layers = [];
    }
  }
}
