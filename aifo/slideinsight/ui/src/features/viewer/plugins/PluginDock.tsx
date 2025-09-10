// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect, useState } from "react";
import { pluginManager } from "./PluginManager";
import { ViewerPlugin, PluginState } from "./types";
import { Dock, DockIcon, DockItem, DockLabel } from "@/components/ui/dock";

/**
 * Plugin dock that renders toggle buttons for all registered plugins
 */
export function PluginDock({ className }: { className?: string }) {
  const [plugins, setPlugins] = useState<ViewerPlugin[]>([]);
  const [pluginStates, setPluginStates] = useState<Record<string, PluginState>>(
    {}
  );

  // Update plugins and states when they change
  useEffect(() => {
    const updatePlugins = () => {
      const allPlugins = pluginManager.getToolbarPlugins();
      console.log(
        "PluginDock: Found plugins:",
        allPlugins.map((p) => p.id)
      );
      setPlugins(allPlugins);

      const states: Record<string, PluginState> = {};
      allPlugins.forEach((plugin) => {
        states[plugin.id] = pluginManager.getPluginState(plugin.id);
      });
      setPluginStates(states);

      return allPlugins;
    };

    // Initial update
    let allPlugins = updatePlugins();

    // Subscribe to plugin registry changes
    const unsubscribePluginChanges = pluginManager.subscribeToPluginChanges(
      () => {
        console.log("PluginDock: Plugin registry changed, updating...");
        allPlugins = updatePlugins();
      }
    );

    // Subscribe to state changes for all current plugins
    let stateUnsubscribers: (() => void)[] = [];

    const setupStateSubscriptions = (plugins: ViewerPlugin[]) => {
      // Clear existing subscriptions
      stateUnsubscribers.forEach((unsub) => unsub());
      stateUnsubscribers = [];

      // Set up new subscriptions
      plugins.forEach((plugin) => {
        const unsubscribe = pluginManager.subscribeToStateChanges(
          plugin.id,
          (state) => {
            setPluginStates((prev) => ({ ...prev, [plugin.id]: state }));
          }
        );
        stateUnsubscribers.push(unsubscribe);
      });
    };

    setupStateSubscriptions(allPlugins);

    // Re-setup state subscriptions when plugins change
    const unsubscribePluginChangesForState =
      pluginManager.subscribeToPluginChanges(() => {
        const currentPlugins = pluginManager.getToolbarPlugins();
        setupStateSubscriptions(currentPlugins);
      });

    return () => {
      unsubscribePluginChanges();
      unsubscribePluginChangesForState();
      stateUnsubscribers.forEach((unsub) => unsub());
    };
  }, []);

  if (plugins.length === 0) {
    return null;
  }

  return (
    <div className={className}>
      <Dock side="right" align="end" edgeInsetPx={8} resizable>
        {plugins
          .slice()
          .reverse()
          .map((plugin) => {
            const state = pluginStates[plugin.id];
            const isActive = state?.isOpen || false;
            const IconComponent = plugin.button?.icon;

            return (
              <div
                onClick={() => pluginManager.handleButtonClick(plugin.id)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    pluginManager.handleButtonClick(plugin.id);
                  }
                }}
                role="button"
                tabIndex={0}
                aria-label={plugin.button?.tooltip || `Toggle ${plugin.name}`}
              >
                <DockItem
                  key={plugin.id}
                  className={`aspect-square rounded-lg transition-colors cursor-pointer overflow-visible ${
                    isActive
                      ? "bg-primary text-primary-foreground shadow-md"
                      : "bg-gray-200 hover:bg-gray-300 dark:bg-neutral-800 dark:hover:bg-neutral-700"
                  }`}
                >
                  <DockLabel>
                    <div className="text-center">
                      <div className="font-medium">
                        {plugin.button?.label || plugin.name}
                      </div>
                      {plugin.button?.tooltip && (
                        <div className="text-xs opacity-75 mt-1">
                          {plugin.button.tooltip}
                        </div>
                      )}
                    </div>
                  </DockLabel>
                  <DockIcon>
                    <div className="flex items-center justify-center text-neutral-600 dark:text-neutral-300 w-full h-full">
                      {IconComponent && <IconComponent className="h-4 w-4" />}
                      {!IconComponent && (
                        <span className="text-sm font-medium">
                          {plugin.button?.label || plugin.name.charAt(0)}
                        </span>
                      )}
                    </div>
                  </DockIcon>
                </DockItem>
              </div>
            );
          })}
      </Dock>
    </div>
  );
}

/**
 * Plugin renderer that renders all open plugin panels
 */
export function PluginRenderer({
  renderMode = "all",
}: { renderMode?: "all" | "docked" | "floating" } = {}) {
  const [openPanels, setOpenPanels] = useState<
    Array<{ plugin: ViewerPlugin; state: PluginState }>
  >([]);

  useEffect(() => {
    const updatePanels = () => {
      const panels = pluginManager.getOpenPanels();
      setOpenPanels(panels);
    };

    // Initial update
    updatePanels();

    // Subscribe to plugin registry changes
    const unsubscribePluginChanges = pluginManager.subscribeToPluginChanges(
      () => {
        console.log(
          "PluginRenderer: Plugin registry changed, updating subscriptions"
        );
        updatePanels();
        setupSubscriptions();
      }
    );

    // Subscribe to state changes for all plugins
    let stateUnsubscribers: (() => void)[] = [];

    const setupSubscriptions = () => {
      // Clear existing subscriptions
      stateUnsubscribers.forEach((unsub) => unsub());
      stateUnsubscribers = [];

      const allPlugins = pluginManager.getAllPlugins();
      console.log(
        "PluginRenderer: Setting up subscriptions for plugins:",
        allPlugins.map((p) => p.id)
      );

      stateUnsubscribers = allPlugins.map((plugin) =>
        pluginManager.subscribeToStateChanges(plugin.id, () => {
          console.log(
            `PluginRenderer: State changed for plugin ${plugin.id}, updating panels`
          );
          updatePanels();
        })
      );
    };

    setupSubscriptions();

    return () => {
      unsubscribePluginChanges();
      stateUnsubscribers.forEach((unsub) => unsub());
    };
  }, []);

  const closePluginHandler = useCallback((pluginId: string) => {
    return () => {
      pluginManager.setPluginState(pluginId, { isOpen: false });
    };
  }, []);

  // Separate docked and floating panels
  const dockedPanels = openPanels.filter(({ state }) => state.dock === "left");
  const floatingPanels = openPanels.filter(
    ({ state }) => state.dock === "free"
  );

  return (
    <>
      {/* Docked plugins - rendered in the left sidebar */}
      {(renderMode === "all" || renderMode === "docked") &&
        dockedPanels.map(({ plugin, state }) => {
          const PanelComponent = plugin.PanelComponent;
          if (!PanelComponent) return null;

          const api = pluginManager.createPluginAPI(plugin.id);

          return (
            <PanelComponent
              key={`docked-${plugin.id}`}
              api={api}
              onClose={closePluginHandler(plugin.id)}
            />
          );
        })}

      {/* Floating plugins - rendered as overlays */}
      {(renderMode === "all" || renderMode === "floating") &&
        floatingPanels.map(({ plugin, state }) => {
          const PanelComponent = plugin.PanelComponent;
          if (!PanelComponent) return null;

          const api = pluginManager.createPluginAPI(plugin.id);

          return (
            <div
              key={`floating-${plugin.id}`}
              className="absolute inset-0 pointer-events-none"
            >
              <PanelComponent
                api={api}
                onClose={closePluginHandler(plugin.id)}
              />
            </div>
          );
        })}
    </>
  );
}

/**
 * Hook to get information about docked plugins for layout calculations
 */
export function useDockedPlugins() {
  const [dockedInfo, setDockedInfo] = useState({
    dockedPlugins: [] as ViewerPlugin[],
    totalDockedWidth: 0,
    hasDockedPlugins: false,
  });

  useEffect(() => {
    const updateDockedInfo = () => {
      const openPanels = pluginManager.getOpenPanels();
      const dockedPanels = openPanels.filter(
        ({ state }) => state.dock === "left"
      );

      const totalWidth = dockedPanels.reduce((total, { state }) => {
        return total + (state.size?.width || 320);
      }, 0);

      setDockedInfo({
        dockedPlugins: dockedPanels.map(({ plugin }) => plugin),
        totalDockedWidth: totalWidth,
        hasDockedPlugins: dockedPanels.length > 0,
      });
    };

    updateDockedInfo();

    // Subscribe to state changes for all plugins
    const allPlugins = pluginManager.getAllPlugins();
    const unsubscribers = allPlugins.map((plugin) =>
      pluginManager.subscribeToStateChanges(plugin.id, updateDockedInfo)
    );

    return () => {
      unsubscribers.forEach((unsub) => unsub());
    };
  }, []);

  return dockedInfo;
}
