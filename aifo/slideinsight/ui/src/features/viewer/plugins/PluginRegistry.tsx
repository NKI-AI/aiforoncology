// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
} from "react";
import {
  PluginRegistration,
  PluginState,
  PluginManager,
  PluginContext,
  PluginCommunication,
} from "./types";

/**
 * Default plugin state
 */
const DEFAULT_PLUGIN_STATE: PluginState = {
  isOpen: false,
  dock: "left",
  size: { width: 320, height: 560 },
  customState: {},
};

/**
 * Plugin Registry Context
 */
const PluginRegistryContext = createContext<{
  manager: PluginManager;
  context: PluginContext;
  states: Record<string, PluginState>;
} | null>(null);

/**
 * Hook to access the plugin registry
 */
export function usePluginRegistry() {
  const registry = useContext(PluginRegistryContext);
  if (!registry) {
    throw new Error(
      "usePluginRegistry must be used within a PluginRegistryProvider"
    );
  }
  return registry;
}

/**
 * Hook to access plugin manager only
 */
export function usePluginManager() {
  const { manager } = usePluginRegistry();
  return manager;
}

/**
 * Hook to access plugin context only
 */
export function usePluginContext() {
  const { context } = usePluginRegistry();
  return context;
}

/**
 * Plugin Registry Provider Props
 */
interface PluginRegistryProviderProps {
  children: React.ReactNode;
  context: PluginContext;
}

/**
 * Plugin Registry Provider
 */
export function PluginRegistryProvider({
  children,
  context,
}: PluginRegistryProviderProps) {
  const [plugins, setPlugins] = useState<Map<string, PluginRegistration>>(
    new Map()
  );
  const [states, setStates] = useState<Record<string, PluginState>>({});
  const messageListenersRef = useRef<
    Map<string, Set<(sourcePluginId: string, message: any) => void>>
  >(new Map());
  const broadcastListenersRef = useRef<
    Set<(sourcePluginId: string, message: any) => void>
  >(new Set());

  // Save plugin state to localStorage when it changes
  const saveState = useCallback((pluginId: string, state: PluginState) => {
    localStorage.setItem(`plugin_${pluginId}_state`, JSON.stringify(state));
  }, []);

  // Create communication system for a plugin
  const createCommunication = useCallback(
    (pluginId: string): PluginCommunication => {
      return {
        sendMessage: (targetPluginId: string, message: any) => {
          const listeners = messageListenersRef.current.get(targetPluginId);
          if (listeners) {
            listeners.forEach((listener) => listener(pluginId, message));
          }
        },
        onMessage: (
          callback: (sourcePluginId: string, message: any) => void
        ) => {
          if (!messageListenersRef.current.has(pluginId)) {
            messageListenersRef.current.set(pluginId, new Set());
          }
          const listeners = messageListenersRef.current.get(pluginId)!;
          listeners.add(callback);

          return () => {
            listeners.delete(callback);
            if (listeners.size === 0) {
              messageListenersRef.current.delete(pluginId);
            }
          };
        },
        broadcast: (message: any) => {
          broadcastListenersRef.current.forEach((listener) =>
            listener(pluginId, message)
          );
        },
        onBroadcast: (
          callback: (sourcePluginId: string, message: any) => void
        ) => {
          broadcastListenersRef.current.add(callback);
          return () => {
            broadcastListenersRef.current.delete(callback);
          };
        },
      };
    },
    []
  );

  // Plugin manager implementation
  const manager: PluginManager = {
    registerPlugin: useCallback(
      async (plugin: PluginRegistration) => {
        // Check dependencies
        if (plugin.dependencies) {
          const missingDeps = plugin.dependencies.filter(
            (dep) => !plugins.has(dep)
          );
          if (missingDeps.length > 0) {
            throw new Error(
              `Plugin ${plugin.id} has missing dependencies: ${missingDeps.join(
                ", "
              )}`
            );
          }
        }

        setPlugins((prev) => new Map(prev).set(plugin.id, plugin));

        // Initialize state for new plugin
        setStates((prev) => {
          if (prev[plugin.id]) return prev; // Already exists

          const saved = localStorage.getItem(`plugin_${plugin.id}_state`);
          let initialState = {
            ...DEFAULT_PLUGIN_STATE,
            ...plugin.defaultState,
          };

          if (saved) {
            try {
              const parsedState = JSON.parse(saved);
              initialState = { ...initialState, ...parsedState };
            } catch {
              // Use default state if parsing fails
            }
          }

          return { ...prev, [plugin.id]: initialState };
        });

        // Call lifecycle hook
        if (plugin.lifecycle?.onRegister) {
          await plugin.lifecycle.onRegister(context);
        }

        // Call onMapReady if map is already available
        if (context.mapRef && plugin.lifecycle?.onMapReady) {
          await plugin.lifecycle.onMapReady(context);
        }

        // Call onSlideLoad if slide is already loaded
        if (
          context.slideMetadata &&
          context.rawSlideMetadata &&
          plugin.lifecycle?.onSlideLoad
        ) {
          await plugin.lifecycle.onSlideLoad(context);
        }
      },
      [context, plugins]
    ),

    unregister: useCallback(
      async (pluginId: string) => {
        const plugin = plugins.get(pluginId);
        if (plugin?.lifecycle?.onUnregister) {
          await plugin.lifecycle.onUnregister(context);
        }

        setPlugins((prev) => {
          const next = new Map(prev);
          next.delete(pluginId);
          return next;
        });
        setStates((prev) => {
          const next = { ...prev };
          delete next[pluginId];
          return next;
        });

        // Clean up message listeners
        messageListenersRef.current.delete(pluginId);
      },
      [plugins, context]
    ),

    getAllPlugins: useCallback(() => {
      return Array.from(plugins.values()).sort(
        (a, b) => (a.order ?? 100) - (b.order ?? 100)
      );
    }, [plugins]),

    getState: useCallback(
      (pluginId: string) => {
        return states[pluginId] || { ...DEFAULT_PLUGIN_STATE };
      },
      [states]
    ),

    updateState: useCallback(
      (pluginId: string, updates: Partial<PluginState>) => {
        setStates((prev) => {
          const currentState = prev[pluginId] || { ...DEFAULT_PLUGIN_STATE };
          const newState = { ...currentState, ...updates };
          saveState(pluginId, newState);
          return { ...prev, [pluginId]: newState };
        });
      },
      [saveState]
    ),

    toggle: useCallback(
      async (pluginId: string) => {
        const currentState = states[pluginId] || { ...DEFAULT_PLUGIN_STATE };
        const newIsOpen = !currentState.isOpen;

        setStates((prev) => {
          const newState = { ...currentState, isOpen: newIsOpen };
          saveState(pluginId, newState);
          return { ...prev, [pluginId]: newState };
        });

        // Call lifecycle hooks
        const plugin = plugins.get(pluginId);
        if (plugin?.lifecycle) {
          if (newIsOpen && plugin.lifecycle.onActivate) {
            await plugin.lifecycle.onActivate(context);
          } else if (!newIsOpen && plugin.lifecycle.onDeactivate) {
            await plugin.lifecycle.onDeactivate(context);
          }
        }
      },
      [states, plugins, context, saveState]
    ),

    closeAll: useCallback(async () => {
      const updates: Record<string, PluginState> = {};
      const deactivationPromises: Promise<void>[] = [];

      Object.entries(states).forEach(([pluginId, state]) => {
        if (state.isOpen) {
          const newState = { ...state, isOpen: false };
          updates[pluginId] = newState;
          saveState(pluginId, newState);

          // Queue deactivation lifecycle call
          const plugin = plugins.get(pluginId);
          if (plugin?.lifecycle?.onDeactivate) {
            deactivationPromises.push(
              Promise.resolve(plugin.lifecycle.onDeactivate(context))
            );
          }
        } else {
          updates[pluginId] = state;
        }
      });

      setStates(updates);
      await Promise.all(deactivationPromises);
    }, [states, plugins, context, saveState]),

    getCommunication: createCommunication,

    // Additional required methods
    unregisterPlugin: useCallback((pluginId: string) => {
      setPlugins((prev) => {
        const newMap = new Map(prev);
        newMap.delete(pluginId);
        return newMap;
      });
      setStates((prev) => {
        const { [pluginId]: removed, ...rest } = prev;
        return rest;
      });
    }, []),

    getPlugin: useCallback(
      (pluginId: string) => {
        return plugins.get(pluginId);
      },
      [plugins]
    ),

    getPluginState: useCallback(
      (pluginId: string) => {
        return states[pluginId] || { ...DEFAULT_PLUGIN_STATE };
      },
      [states]
    ),

    setPluginState: useCallback(
      (pluginId: string, updates: Partial<PluginState>) => {
        setStates((prev) => {
          const currentState = prev[pluginId] || { ...DEFAULT_PLUGIN_STATE };
          const newState = { ...currentState, ...updates };
          saveState(pluginId, newState);
          return { ...prev, [pluginId]: newState };
        });
      },
      [saveState]
    ),

    updateContext: useCallback((newContext: Partial<PluginContext>) => {
      // This would be handled by the parent component
      console.log("updateContext called with:", newContext);
    }, []),
  };

  // Handle context changes - call lifecycle hooks for registered plugins
  useEffect(() => {
    if (context.mapRef) {
      plugins.forEach(async (plugin, pluginId) => {
        if (plugin.lifecycle?.onMapReady) {
          await plugin.lifecycle.onMapReady(context);
        }
      });
    }
  }, [context.mapRef, plugins, context]);

  useEffect(() => {
    if (context.slideMetadata && context.rawSlideMetadata) {
      plugins.forEach(async (plugin, pluginId) => {
        if (plugin.lifecycle?.onSlideLoad) {
          await plugin.lifecycle.onSlideLoad(context);
        }
      });
    }
  }, [context.slideMetadata, context.rawSlideMetadata, plugins, context]);

  return (
    <PluginRegistryContext.Provider value={{ manager, context, states }}>
      {children}
    </PluginRegistryContext.Provider>
  );
}
