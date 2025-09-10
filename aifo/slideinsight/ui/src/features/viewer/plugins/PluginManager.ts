// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import {
  ViewerPlugin,
  PluginManager,
  PluginContext,
  PluginState,
  PluginAPI,
} from "./types";

export class ViewerPluginManager implements PluginManager {
  private plugins = new Map<string, ViewerPlugin>();
  private pluginStates = new Map<string, PluginState>();
  private context: PluginContext = {};
  private stateChangeListeners = new Map<
    string,
    Set<(state: PluginState) => void>
  >();
  private pluginChangeListeners = new Set<() => void>();

  registerPlugin(plugin: ViewerPlugin): void {
    if (this.plugins.has(plugin.id)) {
      // Plugin already registered, skip silently to allow conditional re-registration
      return;
    }

    this.plugins.set(plugin.id, plugin);

    // Initialize plugin state
    const defaultDock = plugin.panel?.defaultDock || "free";
    const compatibleDock: "free" | "left" =
      defaultDock === "left" ? "left" : "free";
    const defaultState: PluginState = {
      isOpen: false,
      dock: compatibleDock,
      size: plugin.panel?.defaultSize,
    };

    // Load state from localStorage if available
    const storageKey = plugin.panel?.storageKey || `plugin_${plugin.id}`;
    const savedState = this.loadPluginState(storageKey);

    this.pluginStates.set(plugin.id, { ...defaultState, ...savedState });
    this.stateChangeListeners.set(plugin.id, new Set());

    // Initialize the plugin
    if (plugin.initialize) {
      const api = this.createPluginAPI(plugin.id);
      plugin.initialize(api);
    }

    console.log(`Plugin ${plugin.id} registered successfully`);

    // Notify plugin change listeners
    this.pluginChangeListeners.forEach((listener) => listener());
  }

  unregisterPlugin(pluginId: string): void {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) return;

    // Cleanup plugin
    if (plugin.destroy) {
      try {
        plugin.destroy();
      } catch (error) {
        console.error(`Error destroying plugin ${pluginId}:`, error);
      }
    }

    this.plugins.delete(pluginId);
    this.pluginStates.delete(pluginId);
    this.stateChangeListeners.delete(pluginId);

    console.log(`Plugin ${pluginId} unregistered`);

    // Notify plugin change listeners
    this.pluginChangeListeners.forEach((listener) => listener());
  }

  // Alias for unregisterPlugin
  unregister(pluginId: string): void {
    this.unregisterPlugin(pluginId);
  }

  getPlugin(pluginId: string): ViewerPlugin | undefined {
    return this.plugins.get(pluginId);
  }

  getAllPlugins(): ViewerPlugin[] {
    return Array.from(this.plugins.values());
  }

  getPluginState(pluginId: string): PluginState {
    const state = this.pluginStates.get(pluginId);
    if (!state) {
      throw new Error(`Plugin ${pluginId} not found`);
    }
    return state;
  }

  // Alias for getPluginState
  getState(pluginId: string): PluginState {
    return this.getPluginState(pluginId);
  }

  setPluginState(pluginId: string, updates: Partial<PluginState>): void {
    const currentState = this.pluginStates.get(pluginId);
    if (!currentState) return;

    const newState = { ...currentState, ...updates };
    this.pluginStates.set(pluginId, newState);

    // Save to localStorage
    const plugin = this.plugins.get(pluginId);
    if (plugin?.panel?.storageKey) {
      this.savePluginState(plugin.panel.storageKey, newState);
    }

    // Notify listeners
    const listeners = this.stateChangeListeners.get(pluginId);
    if (listeners) {
      listeners.forEach((listener) => listener(newState));
    }
  }

  // Alias for setPluginState
  updateState(pluginId: string, updates: Partial<PluginState>): void {
    this.setPluginState(pluginId, updates);
  }

  // Alias for handleButtonClick
  toggle(pluginId: string): void {
    this.handleButtonClick(pluginId);
  }

  // Close all open plugin panels
  closeAll(): void {
    this.plugins.forEach((plugin, pluginId) => {
      if (plugin.panel) {
        this.setPluginState(pluginId, { isOpen: false });
      }
    });
  }

  // Get communication interface for a plugin
  getCommunication(pluginId: string): any {
    // Return a basic communication interface
    return {
      sendMessage: (targetPluginId: string, message: any) => {
        console.log(`Message from ${pluginId} to ${targetPluginId}:`, message);
      },
      onMessage: (callback: (from: string, message: any) => void) => {
        console.log(`Plugin ${pluginId} registered message listener`);
        return () =>
          console.log(`Plugin ${pluginId} unregistered message listener`);
      },
      broadcast: (message: any) => {
        console.log(`Broadcast from ${pluginId}:`, message);
      },
      onBroadcast: (
        callback: (sourcePluginId: string, message: any) => void
      ) => {
        console.log(`Plugin ${pluginId} registered broadcast listener`);
        return () =>
          console.log(`Plugin ${pluginId} unregistered broadcast listener`);
      },
    };
  }

  updateContext(context: Partial<PluginContext>): void {
    this.context = { ...this.context, ...context };

    // Notify all plugins of context change
    this.plugins.forEach((plugin) => {
      if (plugin.onContextChange) {
        plugin.onContextChange(this.context);
      }
    });
  }

  // Subscribe to plugin state changes
  subscribeToStateChanges(
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

  // Subscribe to plugin registry changes (when plugins are added/removed)
  subscribeToPluginChanges(listener: () => void): () => void {
    this.pluginChangeListeners.add(listener);
    return () => this.pluginChangeListeners.delete(listener);
  }

  // Create API object for a specific plugin
  createPluginAPI(pluginId: string): PluginAPI {
    return {
      context: this.context,
      state: this.getPluginState(pluginId),
      setState: (updates: Partial<PluginState>) =>
        this.setPluginState(pluginId, updates),
      closePanel: () => this.setPluginState(pluginId, { isOpen: false }),
    };
  }

  // Get current context (for external access)
  getContext(): PluginContext {
    return this.context;
  }

  // Handle plugin button click
  handleButtonClick(pluginId: string): void {
    console.log(`PluginManager: Button clicked for plugin ${pluginId}`);
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      console.log(`PluginManager: Plugin ${pluginId} not found`);
      return;
    }

    const api = this.createPluginAPI(pluginId);

    if (plugin.onButtonClick) {
      console.log(
        `PluginManager: Calling custom onButtonClick for ${pluginId}`
      );
      plugin.onButtonClick(api);
    } else if (plugin.panel) {
      // Default behavior: toggle panel
      console.log(
        `PluginManager: Using default toggle behavior for ${pluginId}`
      );
      const currentState = this.getPluginState(pluginId);
      console.log(
        `PluginManager: Current state for ${pluginId}:`,
        currentState
      );
      this.setPluginState(pluginId, { isOpen: !currentState.isOpen });
    } else {
      console.log(`PluginManager: No button handler or panel for ${pluginId}`);
    }
  }

  // Get plugins that should show buttons in toolbar
  getToolbarPlugins(): ViewerPlugin[] {
    return this.getAllPlugins()
      .filter((plugin) => plugin.button)
      .sort((a, b) => (a.button?.order || 0) - (b.button?.order || 0));
  }

  // Get plugins that have open panels
  getOpenPanels(): Array<{ plugin: ViewerPlugin; state: PluginState }> {
    return this.getAllPlugins()
      .filter((plugin) => plugin.panel && this.getPluginState(plugin.id).isOpen)
      .map((plugin) => ({
        plugin,
        state: this.getPluginState(plugin.id),
      }));
  }

  private loadPluginState(storageKey: string): Partial<PluginState> {
    try {
      const saved = localStorage.getItem(storageKey);
      if (saved) {
        const parsedState = JSON.parse(saved);
        // Ensure dock value is compatible
        if (
          parsedState.dock &&
          parsedState.dock !== "left" &&
          parsedState.dock !== "free"
        ) {
          parsedState.dock = "free";
        }
        return parsedState;
      }
      return {};
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
export const pluginManager = new ViewerPluginManager();
