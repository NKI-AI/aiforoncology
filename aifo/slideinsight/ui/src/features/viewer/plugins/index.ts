// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// Export plugin system components
export { pluginManager } from "./PluginManager";
export { PluginDock, PluginRenderer, useDockedPlugins } from "./PluginDock";
export type {
  ViewerPlugin,
  PluginAPI,
  PluginContext,
  PluginState,
} from "./types";

// Export built-in plugins
export { BrightnessControlPlugin } from "./BrightnessControlPlugin";
export { AnnotationControlPlugin } from "./AnnotationControlPlugin";

// Note: Import pluginManager directly to avoid circular dependencies
// Use: import { pluginManager } from '@/features/viewer/plugins/PluginManager';
