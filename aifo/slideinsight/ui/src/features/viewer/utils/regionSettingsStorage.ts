// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Region display settings interface
 */
export interface RegionSettings {
  /** Stroke width for region borders (in pixels) */
  strokeWidth: number;
  /** Fill opacity for regions (0-1) */
  fillOpacity: number;
}

/**
 * Partial region settings for updates
 */
export type RegionSettingsPartial = Partial<RegionSettings>;

/**
 * Default region settings
 */
const DEFAULT_REGION_SETTINGS: RegionSettings = {
  strokeWidth: 4,
  fillOpacity: 0,
};

/**
 * Storage key for region settings
 */
const STORAGE_KEY = "regionDisplaySettings";

/**
 * Check if localStorage is available
 */
function isLocalStorageAvailable(): boolean {
  try {
    const test = "__localStorage_test__";
    localStorage.setItem(test, test);
    localStorage.removeItem(test);
    return true;
  } catch {
    return false;
  }
}

/**
 * Safely parse JSON with fallback
 */
function safeParseJSON<T>(jsonString: string | null, fallback: T): T {
  if (!jsonString) {
    return fallback;
  }
  try {
    return JSON.parse(jsonString) as T;
  } catch (error) {
    console.warn("Failed to parse JSON from localStorage:", error);
    return fallback;
  }
}

/**
 * Validate and sanitize region settings
 */
function validateSettings(
  settings: Partial<RegionSettings>,
  defaults: RegionSettings
): RegionSettings {
  return {
    strokeWidth:
      typeof settings.strokeWidth === "number" && settings.strokeWidth > 0
        ? Math.max(1, Math.min(20, settings.strokeWidth))
        : defaults.strokeWidth,
    fillOpacity:
      typeof settings.fillOpacity === "number" &&
      settings.fillOpacity >= 0 &&
      settings.fillOpacity <= 1
        ? settings.fillOpacity
        : defaults.fillOpacity,
  };
}

/**
 * Load region settings from localStorage
 */
export function loadRegionSettings(): RegionSettings {
  if (!isLocalStorageAvailable()) {
    return DEFAULT_REGION_SETTINGS;
  }

  const stored = localStorage.getItem(STORAGE_KEY);
  const parsed = safeParseJSON(stored, {});
  return validateSettings(parsed, DEFAULT_REGION_SETTINGS);
}

/**
 * Save region settings to localStorage
 */
export function saveRegionSettings(
  settings: RegionSettingsPartial
): RegionSettings {
  const current = loadRegionSettings();
  const updated = validateSettings(
    { ...current, ...settings },
    DEFAULT_REGION_SETTINGS
  );

  if (isLocalStorageAvailable()) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
    } catch (error) {
      console.warn("Failed to save region settings to localStorage:", error);
    }
  }

  return updated;
}

/**
 * Reset region settings to defaults
 */
export function resetRegionSettings(): RegionSettings {
  if (isLocalStorageAvailable()) {
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.warn(
        "Failed to remove region settings from localStorage:",
        error
      );
    }
  }
  return DEFAULT_REGION_SETTINGS;
}
