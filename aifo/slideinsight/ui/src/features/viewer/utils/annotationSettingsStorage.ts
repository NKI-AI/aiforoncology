// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Annotation display settings interface
 */
export interface AnnotationSettings {
  /** Stroke width for polygon and box annotations (in pixels) */
  strokeWidth: number;
  /** Fill opacity for polygon and box annotations (0-1) */
  fillOpacity: number;
  /** Point annotation size in micrometers */
  pointSizeMicrometers: number;
}

/**
 * Partial annotation settings for updates
 */
export type AnnotationSettingsPartial = Partial<AnnotationSettings>;

/**
 * Default annotation settings
 */
const DEFAULT_ANNOTATION_SETTINGS: AnnotationSettings = {
  strokeWidth: 3,
  fillOpacity: 0.2,
  pointSizeMicrometers: 5,
};

/**
 * Storage key for annotation settings
 */
const STORAGE_KEY = "annotationDisplaySettings";

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
    const parsed = JSON.parse(jsonString);
    return parsed && typeof parsed === "object"
      ? { ...fallback, ...parsed }
      : fallback;
  } catch (error) {
    console.warn("Error parsing annotation settings from localStorage:", error);
    return fallback;
  }
}

/**
 * Validate and sanitize annotation settings
 */
function validateSettings(
  settings: Partial<AnnotationSettings>,
  defaults: AnnotationSettings
): AnnotationSettings {
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
    pointSizeMicrometers:
      typeof settings.pointSizeMicrometers === "number" &&
      settings.pointSizeMicrometers > 0
        ? Math.max(0.1, Math.min(100, settings.pointSizeMicrometers))
        : defaults.pointSizeMicrometers,
  };
}

/**
 * Load annotation settings from localStorage
 */
export function loadAnnotationSettings(): AnnotationSettings {
  if (!isLocalStorageAvailable()) {
    return DEFAULT_ANNOTATION_SETTINGS;
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    const settings: Partial<AnnotationSettings> = safeParseJSON(stored, {});
    return validateSettings(settings, DEFAULT_ANNOTATION_SETTINGS);
  } catch (error) {
    console.warn("Error loading annotation settings from localStorage:", error);
    return DEFAULT_ANNOTATION_SETTINGS;
  }
}

/**
 * Save annotation settings to localStorage
 */
export function saveAnnotationSettings(
  settings: AnnotationSettingsPartial
): void {
  if (!isLocalStorageAvailable()) {
    return;
  }

  try {
    // Load current settings and merge with new ones
    const currentSettings = loadAnnotationSettings();
    const updatedSettings = validateSettings(
      { ...currentSettings, ...settings },
      DEFAULT_ANNOTATION_SETTINGS
    );

    localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedSettings));
  } catch (error) {
    console.warn("Error saving annotation settings to localStorage:", error);
  }
}

/**
 * Reset annotation settings to defaults
 */
export function resetAnnotationSettings(): AnnotationSettings {
  if (isLocalStorageAvailable()) {
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.warn(
        "Error removing annotation settings from localStorage:",
        error
      );
    }
  }
  return DEFAULT_ANNOTATION_SETTINGS;
}

/**
 * Update a single annotation setting
 */
export function updateAnnotationSetting<K extends keyof AnnotationSettings>(
  key: K,
  value: AnnotationSettings[K]
): void {
  saveAnnotationSettings({ [key]: value });
}

/**
 * Get default annotation settings
 */
export function getDefaultAnnotationSettings(): AnnotationSettings {
  return { ...DEFAULT_ANNOTATION_SETTINGS };
}
