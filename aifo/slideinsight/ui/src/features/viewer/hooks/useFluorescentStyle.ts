// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useMemo, useRef } from "react";
import type TileLayer from "ol/layer/WebGLTile";
import { generateSlideStyle } from "@/features/viewer/components/map/slideStyleUtils";
import type { SlideMetadata } from "@/features/viewer/components/map/types";
import type { SlideStyleOptions } from "@/features/viewer/components/map/slideStyleUtils";
import {
  loadBrightnessSettings,
  loadChannelSettings,
  hasNonDefaultBrightnessSettings,
} from "@/features/viewer/utils/brightnessContrastStorage";

export function useFluorescentStyle(
  slideLayer: TileLayer | null,
  rawMetadata: SlideMetadata | null,
  slideUid?: string,
  onHasNonDefaultSettings?: (hasNonDefault: boolean) => void
) {
  const hasAppliedRef = useRef(false);
  const prevSlideLayerRef = useRef<TileLayer | null>(null);

  const channelMetadata = rawMetadata?.metadata?.channels ?? null;

  // Reset hasApplied flag when slide layer changes
  useEffect(() => {
    if (slideLayer !== prevSlideLayerRef.current) {
      hasAppliedRef.current = false;
      prevSlideLayerRef.current = slideLayer;
    }
  }, [slideLayer]);

  const defaultChannelParams = useMemo(() => {
    if (!channelMetadata)
      return null as Record<
        string,
        { visible: boolean; min: number; max: number }
      > | null;
    const params: Record<
      string,
      { visible: boolean; min: number; max: number }
    > = {};
    Object.keys(channelMetadata).forEach((id) => {
      params[id] = { visible: true, min: 0, max: 65535 };
    });
    return params;
  }, [channelMetadata]);

  useEffect(() => {
    if (!slideLayer || !channelMetadata || hasAppliedRef.current) return;

    // Check if we have stored settings to apply
    if (slideUid) {
      try {
        // Load global brightness settings
        const globalSettings = loadBrightnessSettings();

        // Load channel-specific settings
        const channelIds = Object.keys(channelMetadata);
        const channelSettings = loadChannelSettings(slideUid, channelIds);

        // Check if we have any non-default settings
        const fullSettings = {
          ...globalSettings,
          channels: channelSettings,
        };
        const hasStoredSettings = hasNonDefaultBrightnessSettings(fullSettings);

        if (hasStoredSettings) {
          // Apply stored settings instead of defaults
          const style = generateSlideStyle({
            showGrayscale: globalSettings.showGrayscale,
            invertBackground: globalSettings.invertBackground,
            gamma: globalSettings.globalGamma,
            channels: channelSettings,
            channelMetadata,
          });
          slideLayer.setStyle(style);
          hasAppliedRef.current = true;

          // Report non-default settings state
          if (onHasNonDefaultSettings) {
            onHasNonDefaultSettings(true);
          }
          return;
        }
      } catch (error) {
        console.warn("Error loading stored brightness settings:", error);
      }
    }

    // Apply default settings if no stored settings found
    const style = generateSlideStyle({
      showGrayscale: false,
      invertBackground: false,
      gamma: 1.0,
      channels: defaultChannelParams ?? undefined,
      channelMetadata,
    });
    slideLayer.setStyle(style);
    hasAppliedRef.current = true;

    // Report default settings state
    if (onHasNonDefaultSettings) {
      onHasNonDefaultSettings(false);
    }
  }, [
    slideLayer,
    channelMetadata,
    defaultChannelParams,
    slideUid,
    onHasNonDefaultSettings,
  ]);

  const applyStyle = useCallback(
    (options: SlideStyleOptions) => {
      if (!slideLayer) return;
      const style = generateSlideStyle({
        ...options,
        channelMetadata: channelMetadata ?? undefined,
      });
      slideLayer.setStyle(style);
    },
    [slideLayer, channelMetadata]
  );

  return { applyStyle, hasApplied: hasAppliedRef.current } as const;
}
