// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useEffect, useCallback } from "react";
import {
  loadBrightnessSettings,
  loadChannelSettings,
  hasNonDefaultBrightnessSettings,
  type BrightnessContrastSettings,
} from "@/features/viewer/utils/brightnessContrastStorage";
import type { SlideStyleOptions } from "@/features/viewer/components/map/slideStyleUtils";

interface UseBrightnessAutoApplyProps {
  slideUid: string;
  channels: Record<string, any>;
  onStyleChange?: (options: SlideStyleOptions) => void;
  onHasNonDefaultSettings?: (hasNonDefault: boolean) => void;
}

/**
 * Hook to automatically load and apply brightness/contrast settings from localStorage
 */
export function useBrightnessAutoApply({
  slideUid,
  channels,
  onStyleChange,
  onHasNonDefaultSettings,
}: UseBrightnessAutoApplyProps) {
  const applyStoredSettings = useCallback(() => {
    if (
      !slideUid ||
      !channels ||
      Object.keys(channels).length === 0 ||
      !onStyleChange
    ) {
      return null;
    }

    // Load global brightness settings
    const globalSettings = loadBrightnessSettings();

    // Load channel-specific settings
    const channelIds = Object.keys(channels);
    const channelSettings = loadChannelSettings(slideUid, channelIds);

    // Create the settings object for style application
    const styleOptions: SlideStyleOptions = {
      showGrayscale: globalSettings.showGrayscale,
      invertBackground: globalSettings.invertBackground,
      gamma: globalSettings.globalGamma,
      channels: channelSettings,
      channelMetadata: channels,
    };

    // Check if settings are non-default
    const fullSettings: BrightnessContrastSettings = {
      ...globalSettings,
      channels: channelSettings,
    };
    const hasNonDefault = hasNonDefaultBrightnessSettings(fullSettings);

    // Report non-default state
    if (onHasNonDefaultSettings) {
      onHasNonDefaultSettings(hasNonDefault);
    }

    // Apply the settings
    onStyleChange(styleOptions);

    return styleOptions;
  }, [slideUid, channels, onStyleChange, onHasNonDefaultSettings]);

  // Auto-apply when dependencies change
  useEffect(() => {
    applyStoredSettings();
  }, [applyStoredSettings]);

  return {
    applyStoredSettings,
  };
}
