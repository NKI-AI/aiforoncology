// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface BrightnessContrastSettings {
  showGrayscale: boolean;
  invertBackground: boolean;
  globalGamma: number;
  channels: Record<
    string,
    {
      visible: boolean;
      min: number;
      max: number;
    }
  >;
}

export interface BrightnessContrastSettingsPartial {
  showGrayscale?: boolean;
  invertBackground?: boolean;
  globalGamma?: number;
  channels?: Record<
    string,
    {
      visible?: boolean;
      min?: number;
      max?: number;
    }
  >;
}

/**
 * Default brightness/contrast settings
 */
export function getDefaultBrightnessSettings(): BrightnessContrastSettings {
  return {
    showGrayscale: false,
    invertBackground: false,
    globalGamma: 1.0,
    channels: {},
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
    console.warn(
      "Failed to parse brightness settings from localStorage:",
      error
    );
    return fallback;
  }
}

/**
 * Load brightness/contrast settings from localStorage
 */
export function loadBrightnessSettings(): BrightnessContrastSettings {
  const defaultSettings = getDefaultBrightnessSettings();

  if (!isLocalStorageAvailable()) {
    return defaultSettings;
  }

  try {
    // Load global settings
    const showGrayscale = localStorage.getItem("brightnessPanel_showGrayscale");
    const invertBackground = localStorage.getItem(
      "brightnessPanel_invertBackground"
    );
    const globalGamma = localStorage.getItem("brightnessPanel_globalGamma");

    return {
      showGrayscale: showGrayscale
        ? JSON.parse(showGrayscale)
        : defaultSettings.showGrayscale,
      invertBackground: invertBackground
        ? JSON.parse(invertBackground)
        : defaultSettings.invertBackground,
      globalGamma: globalGamma
        ? JSON.parse(globalGamma)
        : defaultSettings.globalGamma,
      channels: defaultSettings.channels, // Channel settings are handled per-slide
    };
  } catch (error) {
    console.warn("Error loading brightness settings from localStorage:", error);
    return defaultSettings;
  }
}

/**
 * Save brightness/contrast settings to localStorage
 */
export function saveBrightnessSettings(
  settings: BrightnessContrastSettingsPartial
): void {
  if (!isLocalStorageAvailable()) {
    return;
  }

  try {
    if (settings.showGrayscale !== undefined) {
      localStorage.setItem(
        "brightnessPanel_showGrayscale",
        JSON.stringify(settings.showGrayscale)
      );
    }
    if (settings.invertBackground !== undefined) {
      localStorage.setItem(
        "brightnessPanel_invertBackground",
        JSON.stringify(settings.invertBackground)
      );
    }
    if (settings.globalGamma !== undefined) {
      localStorage.setItem(
        "brightnessPanel_globalGamma",
        JSON.stringify(settings.globalGamma)
      );
    }
  } catch (error) {
    console.warn("Error saving brightness settings to localStorage:", error);
  }
}

/**
 * Load channel-specific settings for a given slide
 */
export function loadChannelSettings(
  slideUid: string,
  channelIds: string[]
): Record<string, { visible: boolean; min: number; max: number }> {
  if (!isLocalStorageAvailable()) {
    return createDefaultChannelSettings(channelIds);
  }

  try {
    const stored = localStorage.getItem(`brightnessPanel_channels_${slideUid}`);
    if (stored) {
      const parsed = JSON.parse(stored);
      // Merge with defaults for any missing channels
      const defaultChannels = createDefaultChannelSettings(channelIds);
      return { ...defaultChannels, ...parsed };
    }
    return createDefaultChannelSettings(channelIds);
  } catch (error) {
    console.warn("Error loading channel settings from localStorage:", error);
    return createDefaultChannelSettings(channelIds);
  }
}

/**
 * Save channel-specific settings for a given slide
 */
export function saveChannelSettings(
  slideUid: string,
  channels: Record<string, { visible: boolean; min: number; max: number }>
): void {
  if (!isLocalStorageAvailable()) {
    return;
  }

  try {
    localStorage.setItem(
      `brightnessPanel_channels_${slideUid}`,
      JSON.stringify(channels)
    );
  } catch (error) {
    console.warn("Error saving channel settings to localStorage:", error);
  }
}

/**
 * Create default channel settings for given channel IDs
 */
function createDefaultChannelSettings(
  channelIds: string[]
): Record<string, { visible: boolean; min: number; max: number }> {
  const settings: Record<
    string,
    { visible: boolean; min: number; max: number }
  > = {};
  channelIds.forEach((id) => {
    settings[id] = {
      visible: true,
      min: 0,
      max: 65535, // Default max for 16-bit
    };
  });
  return settings;
}

/**
 * Check if current settings are different from defaults
 */
export function hasNonDefaultBrightnessSettings(
  settings: BrightnessContrastSettings
): boolean {
  const defaults = getDefaultBrightnessSettings();

  return (
    settings.showGrayscale !== defaults.showGrayscale ||
    settings.invertBackground !== defaults.invertBackground ||
    settings.globalGamma !== defaults.globalGamma ||
    Object.keys(settings.channels).some((channelId) => {
      const channel = settings.channels[channelId];
      return (
        channel.visible !== true || channel.min !== 0 || channel.max !== 65535
      );
    })
  );
}

/**
 * Reset brightness/contrast settings to defaults
 */
export function resetBrightnessSettings(
  slideUid?: string
): BrightnessContrastSettings {
  if (!isLocalStorageAvailable()) {
    return getDefaultBrightnessSettings();
  }

  try {
    // Clear global settings
    localStorage.removeItem("brightnessPanel_showGrayscale");
    localStorage.removeItem("brightnessPanel_invertBackground");
    localStorage.removeItem("brightnessPanel_globalGamma");

    // Clear slide-specific channel settings if slideUid provided
    if (slideUid) {
      localStorage.removeItem(`brightnessPanel_channels_${slideUid}`);
    }

    console.log("🔄 Reset brightness/contrast settings");
  } catch (error) {
    console.warn("Error resetting brightness settings:", error);
  }

  return getDefaultBrightnessSettings();
}
