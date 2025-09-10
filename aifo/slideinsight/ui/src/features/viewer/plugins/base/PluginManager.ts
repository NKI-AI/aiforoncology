// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import {
  IPlugin,
  IPluginManager,
  PluginConfig,
  PluginMetadata,
  PluginRegistryEntry,
  PluginCapabilities,
} from "./interfaces";
import { PluginState, PluginContext } from "../types";

// Re-export for compatibility
export type { PluginState, PluginContext };

export interface PluginAPI {
  context: PluginContext;
  state: PluginState;
  setState: (updates: Partial<PluginState>) => void;
  closePanel: () => void;
}

/**
 * Clean, class-based plugin manager
 * No backward compatibility - purely class-based architecture
 */
export class PluginManager implements IPluginManager {
  private plugins = new Map<string, PluginRegistryEntry>();
  private pluginStates = new Map<string, PluginState>();
  private context: PluginContext = {};
  private stateChangeListeners = new Map<
    string,
    Set<(state: PluginState) => void>
  >();
  private pluginChangeListeners = new Set<() => void>();
  private loadOrder: string[] = [];

  /**
   * Register a class-based plugin
   */
  public async registerPlugin(
    plugin: IPlugin,
    config?: PluginConfig,
    metadata?: PluginMetadata
  ): Promise<void> {
    if (this.plugins.has(plugin.id)) {
      console.warn(`Plugin ${plugin.id} is already registered, skipping`);
      return;
    }

    // Validate plugin before registration
    if (!this.validatePlugin(plugin, config)) {
      throw new Error(`Plugin ${plugin.id} failed validation`);
    }

    // Check dependencies and conflicts
    if (!this.canLoadPlugin(plugin.id)) {
      throw new Error(
        `Plugin ${plugin.id} cannot be loaded due to dependencies or conflicts`
      );
    }

    // Create registry entry
    const registryEntry: PluginRegistryEntry = {
      plugin,
      metadata: metadata || this.createDefaultMetadata(plugin),
      config: config || this.createDefaultConfig(plugin),
      registeredAt: new Date(),
      isActive: true,
      loadOrder: this.loadOrder.length,
    };

    this.plugins.set(plugin.id, registryEntry);
    this.loadOrder.push(plugin.id);

    // Initialize plugin state
    const defaultDock = registryEntry.config.panel?.defaultDock || "free";
    const compatibleDock: "free" | "left" =
      defaultDock === "left" ? "left" : "free";
    const defaultState: PluginState = {
      isOpen: false,
      dock: compatibleDock,
      size: registryEntry.config.panel?.defaultSize,
    };

    // Load state from localStorage if available
    const storageKey =
      registryEntry.config.panel?.storageKey || `plugin_${plugin.id}`;
    const savedState = this.loadPluginState(storageKey);

    this.pluginStates.set(plugin.id, { ...defaultState, ...savedState });
    this.stateChangeListeners.set(plugin.id, new Set());

    try {
      // Initialize the plugin
      const api = this.createPluginAPI(plugin.id);
      await plugin.initialize(api);

      console.log(
        `Plugin ${plugin.id} registered and initialized successfully`
      );

      // Notify plugin change listeners
      this.pluginChangeListeners.forEach((listener) => listener());
    } catch (error) {
      // Cleanup on initialization failure
      this.plugins.delete(plugin.id);
      this.pluginStates.delete(plugin.id);
      this.stateChangeListeners.delete(plugin.id);
      this.loadOrder = this.loadOrder.filter((id) => id !== plugin.id);

      console.error(`Failed to initialize plugin ${plugin.id}:`, error);
      throw error;
    }
  }

  /**
   * Unregister a plugin
   */
  public async unregisterPlugin(pluginId: string): Promise<void> {
    const entry = this.plugins.get(pluginId);
    if (!entry) {
      console.warn(`Plugin ${pluginId} not found for unregistration`);
      return;
    }

    try {
      // Destroy the plugin
      await entry.plugin.destroy();
    } catch (error) {
      console.error(`Error destroying plugin ${pluginId}:`, error);
    }

    // Cleanup
    this.plugins.delete(pluginId);
    this.pluginStates.delete(pluginId);
    this.stateChangeListeners.delete(pluginId);
    this.loadOrder = this.loadOrder.filter((id) => id !== pluginId);

    console.log(`Plugin ${pluginId} unregistered successfully`);

    // Notify plugin change listeners
    this.pluginChangeListeners.forEach((listener) => listener());
  }

  /**
   * Get a plugin by ID
   */
  public getPlugin(pluginId: string): IPlugin | undefined {
    return this.plugins.get(pluginId)?.plugin;
  }

  /**
   * Get all registered plugins
   */
  public getAllPlugins(): IPlugin[] {
    return Array.from(this.plugins.values())
      .filter((entry) => entry.isActive)
      .map((entry) => entry.plugin);
  }

  /**
   * Get plugins by capability
   */
  public getPluginsByCapability(
    capability: keyof PluginCapabilities
  ): IPlugin[] {
    return Array.from(this.plugins.values())
      .filter(
        (entry) => entry.isActive && entry.plugin.capabilities[capability]
      )
      .map((entry) => entry.plugin);
  }

  /**
   * Get plugin state
   */
  public getPluginState(pluginId: string): PluginState {
    const state = this.pluginStates.get(pluginId);
    if (!state) {
      throw new Error(`Plugin ${pluginId} not found`);
    }
    return state;
  }

  /**
   * Set plugin state
   */
  public setPluginState(pluginId: string, updates: Partial<PluginState>): void {
    const currentState = this.pluginStates.get(pluginId);
    if (!currentState) {
      console.warn(`Plugin ${pluginId} not found for state update`);
      return;
    }

    const newState = { ...currentState, ...updates };
    this.pluginStates.set(pluginId, newState);

    // Save to localStorage
    const entry = this.plugins.get(pluginId);
    if (entry?.config.panel?.storageKey) {
      this.savePluginState(entry.config.panel.storageKey, newState);
    }

    // Notify listeners
    const listeners = this.stateChangeListeners.get(pluginId);
    if (listeners) {
      listeners.forEach((listener) => listener(newState));
    }
  }

  /**
   * Update context for all plugins
   */
  public async updateContext(context: Partial<PluginContext>): Promise<void> {
    this.context = { ...this.context, ...context };

    // Notify all active plugins of context change
    const promises = Array.from(this.plugins.entries())
      .filter(([_, entry]) => entry.isActive)
      .map(async ([pluginId, entry]) => {
        try {
          await entry.plugin.onContextChange(this.context);
        } catch (error) {
          console.error(
            `Error updating context for plugin ${pluginId}:`,
            error
          );
        }
      });

    await Promise.all(promises);
  }

  /**
   * Enable/disable a plugin
   */
  public async setPluginActive(
    pluginId: string,
    active: boolean
  ): Promise<void> {
    const entry = this.plugins.get(pluginId);
    if (!entry) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    if (entry.isActive === active) {
      return; // No change needed
    }

    entry.isActive = active;

    if (active) {
      // Re-initialize plugin if being activated
      const api = this.createPluginAPI(pluginId);
      await entry.plugin.initialize(api);
      await entry.plugin.onContextChange(this.context);
    } else {
      // Deactivate plugin (but don't destroy it)
      this.setPluginState(pluginId, { isOpen: false });
    }

    console.log(`Plugin ${pluginId} ${active ? "activated" : "deactivated"}`);

    // Notify plugin change listeners
    this.pluginChangeListeners.forEach((listener) => listener());
  }

  /**
   * Check if plugin can be loaded (dependencies, conflicts)
   */
  public canLoadPlugin(pluginId: string): boolean {
    // TODO: Implement dependency and conflict checking
    return true;
  }

  /**
   * Get plugin load order
   */
  public getLoadOrder(): string[] {
    return [...this.loadOrder];
  }

  /**
   * Validate plugin configuration
   */
  public validatePlugin(plugin: IPlugin, config?: PluginConfig): boolean {
    // Basic validation
    if (!plugin.id || !plugin.name || !plugin.version) {
      console.error(`Plugin missing required properties: id, name, or version`);
      return false;
    }

    // Check for ID conflicts
    if (this.plugins.has(plugin.id)) {
      console.error(`Plugin ID ${plugin.id} is already registered`);
      return false;
    }

    // Validate capabilities
    if (!plugin.capabilities || typeof plugin.capabilities !== "object") {
      console.error(`Plugin ${plugin.id} must declare capabilities`);
      return false;
    }

    return true;
  }

  /**
   * Get plugins that should show buttons in toolbar
   */
  public getToolbarPlugins(): IPlugin[] {
    return Array.from(this.plugins.values())
      .filter(
        (entry) =>
          entry.isActive &&
          entry.plugin.capabilities.hasButton &&
          entry.plugin.button
      )
      .sort(
        (a, b) => (a.plugin.button?.order || 0) - (b.plugin.button?.order || 0)
      )
      .map((entry) => entry.plugin);
  }

  /**
   * Get plugins that have open panels
   */
  public getOpenPanels(): Array<{ plugin: IPlugin; state: PluginState }> {
    return Array.from(this.plugins.values())
      .filter(
        (entry) =>
          entry.isActive &&
          entry.plugin.capabilities.hasPanel &&
          this.getPluginState(entry.plugin.id).isOpen
      )
      .map((entry) => ({
        plugin: entry.plugin,
        state: this.getPluginState(entry.plugin.id),
      }));
  }

  /**
   * Subscribe to plugin state changes
   */
  public subscribeToStateChanges(
    pluginId: string,
    listener: (state: PluginState) => void
  ): () => void {
    const listeners = this.stateChangeListeners.get(pluginId);
    if (listeners) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    }
    return () => {};
  }

  /**
   * Subscribe to plugin registry changes (when plugins are added/removed)
   */
  public subscribeToPluginChanges(listener: () => void): () => void {
    this.pluginChangeListeners.add(listener);
    return () => this.pluginChangeListeners.delete(listener);
  }

  /**
   * Create API object for a specific plugin
   */
  public createPluginAPI(pluginId: string): PluginAPI {
    return {
      context: this.context,
      state: this.getPluginState(pluginId),
      setState: (updates: Partial<PluginState>) =>
        this.setPluginState(pluginId, updates),
      closePanel: () => this.setPluginState(pluginId, { isOpen: false }),
    };
  }

  /**
   * Get current context (for external access)
   */
  public getContext(): PluginContext {
    return this.context;
  }

  /**
   * Handle plugin button click
   */
  public handleButtonClick(pluginId: string): void {
    console.log(`PluginManager: Button clicked for plugin ${pluginId}`);
    const entry = this.plugins.get(pluginId);
    if (!entry || !entry.isActive) {
      console.log(`PluginManager: Plugin ${pluginId} not found or inactive`);
      return;
    }

    const api = this.createPluginAPI(pluginId);
    entry.plugin.onButtonClick(api).catch((error) => {
      console.error(
        `Error handling button click for plugin ${pluginId}:`,
        error
      );
    });
  }

  /**
   * Get plugin metadata
   */
  public getPluginMetadata(pluginId: string): PluginMetadata | undefined {
    return this.plugins.get(pluginId)?.metadata;
  }

  /**
   * Get plugin config
   */
  public getPluginConfig(pluginId: string): PluginConfig | undefined {
    return this.plugins.get(pluginId)?.config;
  }

  /**
   * Get plugin registry entry
   */
  public getPluginEntry(pluginId: string): PluginRegistryEntry | undefined {
    return this.plugins.get(pluginId);
  }

  // Private helper methods
  private createDefaultMetadata(plugin: IPlugin): PluginMetadata {
    return {
      author: "Unknown",
      description: `Plugin: ${plugin.name}`,
      category: "other",
      isCore: true,
    };
  }

  private createDefaultConfig(plugin: IPlugin): PluginConfig {
    return {
      capabilities: plugin.capabilities,
      button: plugin.button,
      panel: plugin.panel,
    };
  }

  private loadPluginState(storageKey: string): Partial<PluginState> {
    try {
      const saved = localStorage.getItem(storageKey);
      return saved ? JSON.parse(saved) : {};
    } catch {
      return {};
    }
  }

  private savePluginState(storageKey: string, state: PluginState): void {
    try {
      localStorage.setItem(storageKey, JSON.stringify(state));
    } catch (error) {
      console.warn(`Failed to save plugin state for ${storageKey}:`, error);
    }
  }
}

// Global plugin manager instance
export const pluginManager = new PluginManager();
