// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback, useEffect } from "react";
import { TileFormat } from "../components/map/SlideImage";
import {
  loadViewerSettings,
  updateViewerSetting,
  isLocalStorageAvailable,
} from "../utils/viewerSettingsStorage";

export interface UseViewerTileSettingsProps {
  isFluorescent?: boolean;
}

export interface UseViewerTileSettingsReturn {
  tileFormat: TileFormat;
  autoDetectedFormat: TileFormat | null;
  useAutoDetection: boolean;
  quality?: number;
  setTileFormat: (format: TileFormat) => void;
  setAutoDetectedFormat: (format: TileFormat | null) => void;
  setUseAutoDetection: (use: boolean) => void;
  setQuality: (quality: number | undefined) => void;
  handleTileFormatChange: (format: TileFormat) => void;
  handleAutoDetectionToggle: () => void;
  handleQualityChange: (quality: number | undefined) => void;
  handleTileFormatDetected: (format: TileFormat, bandCount?: number) => void;
}

/**
 * Hook for managing tile format and quality settings with localStorage persistence
 */
export function useViewerTileSettings({
  isFluorescent,
}: UseViewerTileSettingsProps = {}): UseViewerTileSettingsReturn {
  // Initialize state from localStorage if available
  const [tileFormat, setTileFormatState] = useState<TileFormat>(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      const settings = loadViewerSettings(isFluorescent);
      return settings.tileFormat;
    }
    // When isFluorescent is undefined, default to 'jxl' which works for both types
    // Will be updated when metadata loads
    return "jxl";
  });

  const [useAutoDetection, setUseAutoDetectionState] = useState(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      const settings = loadViewerSettings(isFluorescent);
      return settings.useAutoDetection;
    }
    return false; // Disable auto-detection by default to allow JXL preference
  });

  const [quality, setQualityState] = useState<number | undefined>(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      const settings = loadViewerSettings(isFluorescent);
      return settings.quality;
    }
    // Default to high quality when unknown
    return 100;
  });

  const [autoDetectedFormat, setAutoDetectedFormat] =
    useState<TileFormat | null>(null);

  // Save to localStorage when settings change
  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      updateViewerSetting("tileFormat", tileFormat, isFluorescent);
    }
  }, [tileFormat, isFluorescent]);

  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      updateViewerSetting("useAutoDetection", useAutoDetection, isFluorescent);
    }
  }, [useAutoDetection, isFluorescent]);

  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      updateViewerSetting("quality", quality, isFluorescent);
    }
  }, [quality, isFluorescent]);

  // Re-initialize settings when isFluorescent changes from undefined to a boolean value
  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      const settings = loadViewerSettings(isFluorescent);

      console.log(
        `🎯 Loading stored ${
          isFluorescent ? "fluorescent" : "regular"
        } settings:`,
        settings
      );

      // Update to stored settings for this image type
      setTileFormatState(settings.tileFormat);
      setUseAutoDetectionState(settings.useAutoDetection);
      setQualityState(settings.quality);
    }
  }, [isFluorescent]); // Only depend on isFluorescent to trigger when metadata loads

  // Wrapper functions that update localStorage
  const setTileFormat = useCallback((format: TileFormat) => {
    setTileFormatState(format);
  }, []);

  const setUseAutoDetection = useCallback((use: boolean) => {
    setUseAutoDetectionState(use);
  }, []);

  const setQuality = useCallback((newQuality: number | undefined) => {
    setQualityState(newQuality);
  }, []);

  // Handle tile format change (user manual selection)
  const handleTileFormatChange = useCallback(
    (format: TileFormat) => {
      setTileFormat(format);
      setUseAutoDetection(false); // Disable auto-detection when manually changed

      // Set default quality for formats that support it
      const supportsQuality = format === "jpg" || format === "jxl";
      if (supportsQuality && quality === undefined) {
        // Set sensible defaults based on format
        const defaultQuality = format === "jxl" ? 100 : 85;
        setQuality(defaultQuality);
        console.log(
          `🎯 Setting default quality ${defaultQuality} for ${format} format`
        );
      }

      // Note: Changing tile format will trigger a re-render of MapComponent
      // which will reinitialize the map with the new format
    },
    [quality, setTileFormat, setUseAutoDetection, setQuality]
  );

  // Handle auto-detection toggle
  const handleAutoDetectionToggle = useCallback(() => {
    setUseAutoDetectionState((prev) => {
      const newValue = !prev;
      if (newValue && autoDetectedFormat) {
        setTileFormat(autoDetectedFormat);
      }
      return newValue;
    });
  }, [autoDetectedFormat, setTileFormat]);

  // Handle tile format detection from map initialization
  const handleTileFormatDetected = useCallback(
    (format: TileFormat, bandCount?: number) => {
      console.log(
        `🎯 UI received tile format detection: ${format}, bands: ${bandCount}`
      );
      setAutoDetectedFormat(format);

      // Only update the active format if auto-detection is enabled
      if (useAutoDetection) {
        setTileFormat(format);
        console.log(
          `✅ Auto-detection enabled, switching UI to ${format} format`
        );

        // Set default quality for auto-detected formats that support it
        const supportsQuality = format === "jpg" || format === "jxl";
        if (supportsQuality && quality === undefined) {
          const defaultQuality = format === "jxl" ? 100 : 85;
          setQuality(defaultQuality);
          console.log(
            `🎯 Auto-detection: Setting default quality ${defaultQuality} for ${format} format`
          );
        }
      } else {
        console.log(
          `⚙️ Auto-detection disabled, keeping manual format: ${tileFormat}`
        );
      }
    },
    [useAutoDetection, tileFormat, quality, setTileFormat, setQuality]
  );

  // Handle quality parameter change
  const handleQualityChange = useCallback(
    (newQuality: number | undefined) => {
      setQuality(newQuality);
      // Quality change will trigger a re-render of MapComponent which will update URLs
    },
    [setQuality]
  );

  return {
    tileFormat,
    autoDetectedFormat,
    useAutoDetection,
    quality,
    setTileFormat,
    setAutoDetectedFormat,
    setUseAutoDetection,
    setQuality,
    handleTileFormatChange,
    handleAutoDetectionToggle,
    handleQualityChange,
    handleTileFormatDetected,
  };
}
