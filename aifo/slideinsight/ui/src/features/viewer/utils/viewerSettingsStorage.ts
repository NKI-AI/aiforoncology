// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { TileFormat } from "../components/map/SlideImage";
import { CrosshairMode } from "../components/map/CoordinateTracker";

/**
 * Settings that are stored in localStorage for viewer preferences
 */
export interface ViewerSettings {
  showCrosshair: boolean;
  panSensitivity: number;
  zoomSensitivity: number;
  tileFormat: TileFormat;
  useAutoDetection: boolean;
  quality?: number; // Quality parameter for JPG/JXL (0-100)
}

/**
 * Default settings for different image types
 */
export const DEFAULT_REGULAR_SETTINGS: ViewerSettings = {
  showCrosshair: false,
  panSensitivity: 1.0,
  zoomSensitivity: 1.0,
  tileFormat: "jpg", // Changed from 'jxl' to 'jpg' for regular images
  useAutoDetection: false, // Disable auto-detection to use JPG preference
  quality: 85, // Good quality for JPG
};

export const DEFAULT_FLUORESCENT_SETTINGS: ViewerSettings = {
  showCrosshair: false,
  panSensitivity: 1.0,
  zoomSensitivity: 1.0,
  tileFormat: "jxl", // JXL also supports multi-channel fluorescent
  useAutoDetection: false, // Auto-detection disabled for fluorescent
  quality: 100, // High quality default for fluorescent
};

/**
 * Storage keys for different image types
 */
const STORAGE_KEYS = {
  regular: "slideinsight_viewer_settings_regular",
  fluorescent: "slideinsight_viewer_settings_fluorescent",
} as const;

/**
 * Legacy storage key for backward compatibility
 */
const LEGACY_STORAGE_KEY = "slideinsight_viewer_settings";

/**
 * Get the appropriate storage key based on image type
 */
function getStorageKey(isFluorescent: boolean): string {
  return isFluorescent ? STORAGE_KEYS.fluorescent : STORAGE_KEYS.regular;
}

/**
 * Get default settings based on image type
 */
function getDefaultSettings(isFluorescent: boolean): ViewerSettings {
  return isFluorescent
    ? DEFAULT_FLUORESCENT_SETTINGS
    : DEFAULT_REGULAR_SETTINGS;
}

/**
 * Validate and sanitize settings
 */
function validateSettings(
  settings: any,
  defaultSettings: ViewerSettings
): ViewerSettings {
  return {
    showCrosshair:
      typeof settings.showCrosshair === "boolean"
        ? settings.showCrosshair
        : defaultSettings.showCrosshair,
    panSensitivity:
      typeof settings.panSensitivity === "number" && settings.panSensitivity > 0
        ? Math.max(0.1, Math.min(100, settings.panSensitivity))
        : defaultSettings.panSensitivity,
    zoomSensitivity:
      typeof settings.zoomSensitivity === "number" &&
      settings.zoomSensitivity > 0
        ? Math.max(0.1, Math.min(100, settings.zoomSensitivity))
        : defaultSettings.zoomSensitivity,
    tileFormat: ["png", "jpg", "jxl"].includes(settings.tileFormat)
      ? settings.tileFormat
      : defaultSettings.tileFormat,
    useAutoDetection:
      typeof settings.useAutoDetection === "boolean"
        ? settings.useAutoDetection
        : defaultSettings.useAutoDetection,
    quality:
      typeof settings.quality === "number" &&
      settings.quality >= 10 &&
      settings.quality <= 100
        ? settings.quality
        : settings.quality === undefined
        ? undefined
        : defaultSettings.quality,
  };
}

/**
 * Safely parse JSON from localStorage with error handling
 */
function safeParseJSON<T>(jsonString: string | null, fallback: T): T {
  if (!jsonString) return fallback;

  try {
    const parsed = JSON.parse(jsonString);
    // Validate that parsed object has the expected structure
    if (typeof parsed === "object" && parsed !== null) {
      return { ...fallback, ...parsed };
    }
    return fallback;
  } catch (error) {
    console.warn("Failed to parse viewer settings from localStorage:", error);
    return fallback;
  }
}

/**
 * Load viewer settings from localStorage based on image type
 */
export function loadViewerSettings(isFluorescent: boolean): ViewerSettings {
  const storageKey = getStorageKey(isFluorescent);
  const defaultSettings = getDefaultSettings(isFluorescent);

  try {
    // First try to load from the new format
    let stored = localStorage.getItem(storageKey);
    let settings: ViewerSettings = safeParseJSON(stored, defaultSettings);

    // Validate and sanitize the loaded settings
    return validateSettings(settings, defaultSettings);
  } catch (error) {
    console.warn("Error loading viewer settings from localStorage:", error);
    return defaultSettings;
  }
}

/**
 * Save viewer settings to localStorage based on image type
 */
export function saveViewerSettings(
  settings: ViewerSettings,
  isFluorescent: boolean
): void {
  const storageKey = getStorageKey(isFluorescent);

  try {
    // Sanitize settings before saving
    const sanitizedSettings: ViewerSettings = {
      showCrosshair: settings.showCrosshair,
      panSensitivity: Math.max(0.1, Math.min(100, settings.panSensitivity)),
      zoomSensitivity: Math.max(0.1, Math.min(100, settings.zoomSensitivity)),
      tileFormat: settings.tileFormat,
      useAutoDetection: settings.useAutoDetection,
      quality:
        settings.quality !== undefined
          ? Math.max(10, Math.min(100, settings.quality))
          : undefined,
    };

    localStorage.setItem(storageKey, JSON.stringify(sanitizedSettings));
  } catch (error) {
    console.warn("Error saving viewer settings to localStorage:", error);
  }
}

/**
 * Reset viewer settings to defaults for a specific image type
 */
export function resetViewerSettings(isFluorescent: boolean): ViewerSettings {
  const storageKey = getStorageKey(isFluorescent);
  const defaultSettings = getDefaultSettings(isFluorescent);

  try {
    localStorage.removeItem(storageKey);
    console.log(
      `🔄 Reset settings for ${
        isFluorescent ? "fluorescent" : "regular"
      } images`
    );
  } catch (error) {
    console.warn("Error removing viewer settings from localStorage:", error);
  }

  return defaultSettings;
}

/**
 * Update a specific setting and save to localStorage
 */
export function updateViewerSetting<K extends keyof ViewerSettings>(
  key: K,
  value: ViewerSettings[K],
  isFluorescent: boolean
): ViewerSettings {
  const currentSettings = loadViewerSettings(isFluorescent);
  const updatedSettings = {
    ...currentSettings,
    [key]: value,
  };

  saveViewerSettings(updatedSettings, isFluorescent);
  return updatedSettings;
}

/**
 * Get all stored settings for both image types
 */
export function getAllStoredSettings(): {
  regular: ViewerSettings;
  fluorescent: ViewerSettings;
} {
  return {
    regular: loadViewerSettings(false),
    fluorescent: loadViewerSettings(true),
  };
}

/**
 * Clear all stored settings
 */
export function clearAllStoredSettings(): void {
  try {
    Object.values(STORAGE_KEYS).forEach((key) => {
      localStorage.removeItem(key);
    });
    // Also remove legacy key if it exists
    localStorage.removeItem(LEGACY_STORAGE_KEY);
    console.log("🗑️ Cleared all viewer settings");
  } catch (error) {
    console.warn("Error clearing all viewer settings:", error);
  }
}

/**
 * Check if localStorage is available and functional
 */
export function isLocalStorageAvailable(): boolean {
  try {
    const testKey = "__localStorage_test__";
    localStorage.setItem(testKey, "test");
    localStorage.removeItem(testKey);
    return true;
  } catch {
    return false;
  }
}
