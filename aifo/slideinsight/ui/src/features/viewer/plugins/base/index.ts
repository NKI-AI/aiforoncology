// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// Export base classes
export {
  BasePlugin,
  MapInteractionPlugin,
  LayerPlugin,
  MapLayerPlugin,
} from "./BasePlugin";

// Export interfaces
export type {
  IPlugin,
  IMapInteractionPlugin,
  ILayerPlugin,
  IAnnotationPlugin,
  IRegionPlugin,
  IDataPlugin,
  IKeyboardPlugin,
  IStatefulPlugin,
  IPluginManager,
  PluginCapabilities,
  PluginButtonConfig,
  PluginPanelConfig,
  PluginLifecycleHooks,
  PluginConfig,
  PluginMetadata,
  PluginRegistryEntry,
  MapInteractionContext,
  LayerContext,
} from "./interfaces";

// Export plugin manager and types
export {
  PluginManager,
  pluginManager,
  type PluginContext,
  type PluginState,
  type PluginAPI,
} from "./PluginManager";
