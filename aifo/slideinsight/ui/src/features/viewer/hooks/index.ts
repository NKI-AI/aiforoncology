// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Centralized exports for viewer feature hooks
 */

// ===== VIEWER STATE HOOKS =====
export { useViewerState } from "./useViewerState";
export { useViewerTileSettings } from "./useViewerTileSettings";

// ===== MAP & LAYER HOOKS =====
export { useSlideMap } from "./useSlideMap";
export { useMapInitialization } from "./useMapInitialization";
export { useMaskLayers, type UseMaskLayersResult } from "./useMaskLayers";
export { useVectorLayers, type UseVectorLayersResult } from "./useVectorLayers";
export { useVectorLayer } from "./useVectorLayer";

export { useFluorescentStyle } from "./useFluorescentStyle";
export { useMapLoadSpinner } from "./useMapLoadSpinner";

// ===== MASK-SPECIFIC HOOKS =====
export { useMaskColors } from "./useMaskColors";
export { useMaskKeyboardShortcuts } from "./useMaskKeyboardShortcuts";

// ===== NAVIGATION HOOKS =====
export { useSlideNavigationData } from "./useSlideNavigationData";
