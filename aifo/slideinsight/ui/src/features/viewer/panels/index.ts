// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// Core types and interfaces
export type {
  PanelContext,
  PanelState,
  PanelRegistration,
  PanelProps,
  PanelManager,
} from "./types";

// Panel registry and management
export {
  PanelRegistryProvider,
  usePanelRegistry,
  usePanelManager,
  usePanelContext,
} from "./PanelRegistry";

// Panel dock and renderer components
export { PanelDock, PanelRenderer, useDockedPanels } from "./PanelDock";

// Built-in panels
export {
  builtinPanels,
  registerBuiltinPanels,
  maskControlPanelRegistration,
  annotationPanelRegistration,
  brightnessPanelRegistration,
} from "./builtins";
