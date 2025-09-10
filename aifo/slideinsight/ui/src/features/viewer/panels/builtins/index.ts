// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { PanelManager } from "../types";
import { maskControlPanelRegistration } from "./MaskControlPanelWrapper";
import { annotationPanelRegistration } from "./AnnotationPanelWrapper";
import { brightnessPanelRegistration } from "./BrightnessPanelWrapper";

/**
 * All built-in panel registrations
 */
export const builtinPanels = [
  maskControlPanelRegistration,
  annotationPanelRegistration,
  brightnessPanelRegistration,
];

/**
 * Register all built-in panels with the panel manager
 */
export function registerBuiltinPanels(manager: PanelManager) {
  builtinPanels.forEach((panel) => {
    manager.register(panel);
  });
}

// Re-export individual registrations for direct use
export { maskControlPanelRegistration } from "./MaskControlPanelWrapper";
export { annotationPanelRegistration } from "./AnnotationPanelWrapper";
export { brightnessPanelRegistration } from "./BrightnessPanelWrapper";
