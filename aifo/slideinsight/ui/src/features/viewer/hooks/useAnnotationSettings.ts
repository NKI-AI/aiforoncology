// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback, useEffect } from "react";
import {
  type AnnotationSettings,
  type AnnotationSettingsPartial,
  loadAnnotationSettings,
  saveAnnotationSettings,
  resetAnnotationSettings,
  updateAnnotationSetting,
} from "@/features/viewer/utils/annotationSettingsStorage";

export interface UseAnnotationSettingsReturn {
  settings: AnnotationSettings;
  updateSettings: (settings: AnnotationSettingsPartial) => void;
  updateSetting: <K extends keyof AnnotationSettings>(
    key: K,
    value: AnnotationSettings[K]
  ) => void;
  resetSettings: () => void;
}

/**
 * Hook for managing annotation display settings with localStorage persistence
 */
export function useAnnotationSettings(): UseAnnotationSettingsReturn {
  const [settings, setSettings] = useState<AnnotationSettings>(() =>
    loadAnnotationSettings()
  );

  // Update settings and persist to localStorage
  const updateSettings = useCallback(
    (newSettings: AnnotationSettingsPartial) => {
      setSettings((current) => {
        const updated = { ...current, ...newSettings };
        saveAnnotationSettings(newSettings);
        return updated;
      });
    },
    []
  );

  // Update a single setting
  const updateSetting = useCallback(
    <K extends keyof AnnotationSettings>(
      key: K,
      value: AnnotationSettings[K]
    ) => {
      setSettings((current) => {
        const updated = { ...current, [key]: value };
        updateAnnotationSetting(key, value);
        return updated;
      });
    },
    []
  );

  // Reset settings to defaults
  const resetSettings = useCallback(() => {
    const defaultSettings = resetAnnotationSettings();
    setSettings(defaultSettings);
  }, []);

  // Load settings on mount (in case they were changed elsewhere)
  useEffect(() => {
    const loadedSettings = loadAnnotationSettings();
    setSettings(loadedSettings);
  }, []);

  return {
    settings,
    updateSettings,
    updateSetting,
    resetSettings,
  };
}
