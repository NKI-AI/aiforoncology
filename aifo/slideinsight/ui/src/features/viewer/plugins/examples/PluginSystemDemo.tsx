// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useEffect } from "react";
import {
  pluginManager,
  BasePlugin,
  MapInteractionPlugin,
  LayerPlugin,
  type IPlugin,
  type PluginConfig,
  type PluginMetadata,
} from "../base";
import { brightnessControlPluginV2 } from "../BrightnessControlPluginV2";
import { regionControlPluginV2 } from "../RegionControlPluginV2";

/**
 * Example of a simple plugin that extends BasePlugin
 */
class SimpleExamplePlugin extends BasePlugin {
  public readonly id = "simple-example";
  public readonly name = "Simple Example";
  public readonly version = "1.0.0";

  protected setupDefaultCapabilities(): void {
    this.addCapability("hasButton", true);
    this.addCapability("hasPanel", true);

    this.setButton({
      id: "simple-example-button",
      label: "Simple",
      tooltip: "A simple example plugin",
      position: "right",
      order: 10,
    });

    this.setPanel({
      id: "simple-example-panel",
      title: "Simple Example",
      defaultSize: { width: 300, height: 200 },
      defaultDock: "free",
    });
  }

  protected createPanelComponent(): React.ComponentType<any> {
    return ({ api, onClose }) => (
      <div className="p-4">
        <h3 className="text-lg font-semibold mb-2">Simple Example Plugin</h3>
        <p className="text-sm text-gray-600 mb-4">
          This is a demonstration of the new class-based plugin system.
        </p>
        <div className="space-y-2">
          <p>
            <strong>Plugin ID:</strong> {this.id}
          </p>
          <p>
            <strong>Version:</strong> {this.version}
          </p>
          <p>
            <strong>Slide UID:</strong> {api.context.slideUid || "None"}
          </p>
          <p>
            <strong>Panel Open:</strong> {api.state.isOpen ? "Yes" : "No"}
          </p>
        </div>
        <button
          onClick={onClose}
          className="mt-4 px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Close
        </button>
      </div>
    );
  }
}

/**
 * Example of a map interaction plugin
 */
class MapInteractionExamplePlugin extends MapInteractionPlugin {
  public readonly id = "map-interaction-example";
  public readonly name = "Map Interaction Example";
  public readonly version = "1.0.0";

  protected setupDefaultCapabilities(): void {
    this.addCapability("hasMapInteractions", true);
    this.addCapability("hasButton", true);

    this.setButton({
      id: "map-interaction-example-button",
      label: "Map Demo",
      tooltip: "Demonstrates map interaction capabilities",
      position: "right",
      order: 11,
    });
  }

  public async setupMapInteractions(context: any): Promise<void> {
    console.log(
      "MapInteractionExamplePlugin: Setting up map interactions",
      context
    );
    // Add custom map interactions here
  }

  public async cleanupMapInteractions(): Promise<void> {
    console.log("MapInteractionExamplePlugin: Cleaning up map interactions");
    // Remove custom map interactions here
  }

  protected async handleButtonClick(api: any): Promise<void> {
    console.log("MapInteractionExamplePlugin: Button clicked");
    console.log("Map available:", !!this.getMap());
    console.log("Interactions count:", this.interactions.length);
  }
}

/**
 * Plugin system demonstration component
 */
export function PluginSystemDemo() {
  useEffect(() => {
    const registerDemoPlugins = async () => {
      try {
        // Register the new class-based plugins
        console.log("Registering demo plugins...");

        // Register simple example plugin
        const simplePlugin = new SimpleExamplePlugin();
        await pluginManager.registerPlugin(
          simplePlugin,
          {
            capabilities: {
              hasButton: true,
              hasPanel: true,
            },
          },
          {
            author: "SlideInsight Team",
            description: "A simple demonstration plugin",
            category: "utility",
            isCore: false,
          }
        );

        // Register map interaction example plugin
        const mapPlugin = new MapInteractionExamplePlugin();
        await pluginManager.registerPlugin(
          mapPlugin,
          {
            capabilities: {
              hasMapInteractions: true,
              hasButton: true,
            },
          },
          {
            author: "SlideInsight Team",
            description: "Demonstrates map interaction capabilities",
            category: "utility",
            isCore: false,
          }
        );

        // Register the enhanced brightness control plugin
        await pluginManager.registerPlugin(
          brightnessControlPluginV2,
          {
            capabilities: {
              hasPanel: true,
              hasButton: true,
              requiresSlideMetadata: true,
              persistsState: true,
            },
          },
          {
            author: "SlideInsight Team",
            description:
              "Enhanced brightness and contrast control for fluorescent images",
            category: "visualization",
            isCore: true,
          }
        );

        // Register the enhanced region control plugin
        await pluginManager.registerPlugin(
          regionControlPluginV2,
          {
            capabilities: {
              hasMapInteractions: true,
              hasLayers: true,
              hasPanel: true,
              hasButton: true,
              canManageRegions: true,
              requiresSlideMetadata: true,
              persistsState: true,
            },
          },
          {
            author: "SlideInsight Team",
            description: "Enhanced region of interest management",
            category: "annotation",
            isCore: true,
          }
        );

        console.log("Demo plugins registered successfully!");

        // Demonstrate plugin manager capabilities
        console.log(
          "All plugins:",
          pluginManager.getAllPlugins().map((p) => p.id)
        );
        console.log(
          "Plugins with map interactions:",
          pluginManager
            .getPluginsByCapability("hasMapInteractions")
            .map((p) => p.id)
        );
        console.log(
          "Plugins with panels:",
          pluginManager.getPluginsByCapability("hasPanel").map((p) => p.id)
        );
      } catch (error) {
        console.error("Error registering demo plugins:", error);
      }
    };

    registerDemoPlugins();

    // Cleanup on unmount
    return () => {
      console.log("Cleaning up demo plugins...");
      pluginManager.unregisterPlugin("simple-example");
      pluginManager.unregisterPlugin("map-interaction-example");
      pluginManager.unregisterPlugin("brightness-control-v2");
      pluginManager.unregisterPlugin("region-control-v2");
    };
  }, []);

  return (
    <div className="p-4 bg-gray-50 rounded-lg">
      <h2 className="text-xl font-bold mb-4">Plugin System Demo</h2>
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-semibold mb-2">
            New Plugin Architecture Features
          </h3>
          <ul className="list-disc list-inside space-y-1 text-sm">
            <li>
              <strong>Class-based plugins:</strong> Proper inheritance and
              encapsulation
            </li>
            <li>
              <strong>Plugin capabilities:</strong> Declarative system for
              plugin features
            </li>
            <li>
              <strong>Specialized base classes:</strong> BasePlugin,
              MapInteractionPlugin, LayerPlugin, MapLayerPlugin
            </li>
            <li>
              <strong>Lifecycle hooks:</strong> Proper initialization and
              cleanup
            </li>
            <li>
              <strong>Enhanced plugin manager:</strong> Better state management
              and validation
            </li>
            <li>
              <strong>Backward compatibility:</strong> Legacy plugins still work
            </li>
          </ul>
        </div>

        <div>
          <h3 className="text-lg font-semibold mb-2">
            Plugin Types Demonstrated
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div className="bg-white p-3 rounded border">
              <h4 className="font-semibold">BasePlugin</h4>
              <p>
                Simple Example Plugin - Basic panel and button functionality
              </p>
            </div>
            <div className="bg-white p-3 rounded border">
              <h4 className="font-semibold">MapInteractionPlugin</h4>
              <p>
                Map Interaction Example - Demonstrates map interaction
                capabilities
              </p>
            </div>
            <div className="bg-white p-3 rounded border">
              <h4 className="font-semibold">BasePlugin (Enhanced)</h4>
              <p>
                Brightness Control V2 - Enhanced version with better structure
              </p>
            </div>
            <div className="bg-white p-3 rounded border">
              <h4 className="font-semibold">MapLayerPlugin</h4>
              <p>
                Region Control V2 - Full map interaction and layer management
              </p>
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-lg font-semibold mb-2">Usage Instructions</h3>
          <ol className="list-decimal list-inside space-y-1 text-sm">
            <li>
              Open the browser developer console to see plugin registration logs
            </li>
            <li>Look for the new plugin buttons in the viewer toolbar</li>
            <li>Click the buttons to test the enhanced plugin functionality</li>
            <li>Notice the improved error handling and validation</li>
            <li>Check the console for detailed plugin lifecycle events</li>
          </ol>
        </div>
      </div>
    </div>
  );
}
