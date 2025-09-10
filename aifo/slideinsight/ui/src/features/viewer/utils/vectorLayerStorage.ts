// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface VectorLayerSettings {
  id: string;
  name: string;
  color: string;
  visible: boolean;
  defaultColor?: string;
  vectorName?: string;
}

export interface VectorContextSettings {
  vectorOpacity: number;
  showVectors: boolean;
  vectorColors: string[];
  vectorLayers: VectorLayerSettings[];
}

/**
 * Default vector layer settings
 */
export function getDefaultVectorSettings(): VectorContextSettings {
  return {
    vectorOpacity: 0.7,
    showVectors: true,
    vectorColors: [
      "#ff0000", // Red
      "#00ff00", // Green
      "#0000ff", // Blue
      "#ffff00", // Yellow
      "#ff00ff", // Magenta
      "#00ffff", // Cyan
      "#ffa500", // Orange
      "#800080", // Purple
      "#ffc0cb", // Pink
      "#a52a2a", // Brown
    ],
    vectorLayers: [],
  };
}

/**
 * Check if localStorage is available
 */
function isLocalStorageAvailable(): boolean {
  try {
    return typeof window !== "undefined" && "localStorage" in window;
  } catch {
    return false;
  }
}

/**
 * Safely parse JSON from localStorage with error handling
 */
function safeParseJSON<T>(jsonString: string | null, fallback: T): T {
  if (!jsonString) return fallback;

  try {
    const parsed = JSON.parse(jsonString);
    if (typeof parsed === "object" && parsed !== null) {
      return { ...fallback, ...parsed };
    }
    return fallback;
  } catch (error) {
    console.warn("Failed to parse vector settings from localStorage:", error);
    return fallback;
  }
}

/**
 * Load vector context settings from localStorage
 */
export function loadVectorSettings(): VectorContextSettings {
  const defaultSettings = getDefaultVectorSettings();

  if (!isLocalStorageAvailable()) {
    return defaultSettings;
  }

  try {
    // Load global vector settings
    const vectorOpacity = localStorage.getItem("vectorPanel_opacity");
    const showVectors = localStorage.getItem("vectorPanel_showVectors");
    const vectorColors = localStorage.getItem("vectorPanel_colors");

    return {
      vectorOpacity: vectorOpacity
        ? JSON.parse(vectorOpacity)
        : defaultSettings.vectorOpacity,
      showVectors: showVectors
        ? JSON.parse(showVectors)
        : defaultSettings.showVectors,
      vectorColors: vectorColors
        ? JSON.parse(vectorColors)
        : defaultSettings.vectorColors,
      vectorLayers: defaultSettings.vectorLayers, // Layers are handled per-slide
    };
  } catch (error) {
    console.warn("Error loading vector settings from localStorage:", error);
    return defaultSettings;
  }
}

/**
 * Save vector context settings to localStorage
 */
export function saveVectorSettings(
  settings: Partial<VectorContextSettings>
): void {
  if (!isLocalStorageAvailable()) {
    return;
  }

  try {
    if (settings.vectorOpacity !== undefined) {
      localStorage.setItem(
        "vectorPanel_opacity",
        JSON.stringify(settings.vectorOpacity)
      );
    }
    if (settings.showVectors !== undefined) {
      localStorage.setItem(
        "vectorPanel_showVectors",
        JSON.stringify(settings.showVectors)
      );
    }
    if (settings.vectorColors !== undefined) {
      localStorage.setItem(
        "vectorPanel_colors",
        JSON.stringify(settings.vectorColors)
      );
    }
  } catch (error) {
    console.warn("Error saving vector settings to localStorage:", error);
  }
}

/**
 * Load vector layer settings for a given slide
 */
export function loadVectorLayerSettings(
  slideUid: string
): VectorLayerSettings[] {
  if (!isLocalStorageAvailable()) {
    return [];
  }

  try {
    const stored = localStorage.getItem(`vectorPanel_layers_${slideUid}`);
    if (stored) {
      const parsed = JSON.parse(stored);
      return Array.isArray(parsed) ? parsed : [];
    }
    return [];
  } catch (error) {
    console.warn(
      "Error loading vector layer settings from localStorage:",
      error
    );
    return [];
  }
}

/**
 * Save vector layer settings for a given slide
 */
export function saveVectorLayerSettings(
  slideUid: string,
  layers: VectorLayerSettings[]
): void {
  if (!isLocalStorageAvailable()) {
    return;
  }

  try {
    localStorage.setItem(
      `vectorPanel_layers_${slideUid}`,
      JSON.stringify(layers)
    );
  } catch (error) {
    console.warn("Error saving vector layer settings to localStorage:", error);
  }
}

/**
 * Check if current vector settings are different from defaults
 */
export function hasNonDefaultVectorSettings(
  settings: VectorContextSettings
): boolean {
  const defaults = getDefaultVectorSettings();

  return (
    settings.vectorOpacity !== defaults.vectorOpacity ||
    settings.showVectors !== defaults.showVectors ||
    JSON.stringify(settings.vectorColors) !==
      JSON.stringify(defaults.vectorColors) ||
    settings.vectorLayers.length > 0
  );
}

/**
 * Reset vector settings to defaults
 */
export function resetVectorSettings(slideUid?: string): VectorContextSettings {
  if (!isLocalStorageAvailable()) {
    return getDefaultVectorSettings();
  }

  try {
    // Clear global settings
    localStorage.removeItem("vectorPanel_opacity");
    localStorage.removeItem("vectorPanel_showVectors");
    localStorage.removeItem("vectorPanel_colors");

    // Clear slide-specific layer settings if slideUid provided
    if (slideUid) {
      localStorage.removeItem(`vectorPanel_layers_${slideUid}`);
    }

    console.log("🔄 Reset vector layer settings");
  } catch (error) {
    console.warn("Error resetting vector settings:", error);
  }

  return getDefaultVectorSettings();
}
