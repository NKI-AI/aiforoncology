// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { SunIcon } from "@heroicons/react/24/outline";
import { PanelProps, PanelRegistration } from "../types";
import { BrightnessContrastPanel } from "@/features/viewer/components/BrightnessContrastPanel";

/**
 * Wrapper component for the existing BrightnessContrastPanel
 */
function BrightnessPanelWrapper({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  const customState = state.customState || {};

  return (
    <BrightnessContrastPanel
      isOpen={state.isOpen}
      onClose={onClose}
      channels={customState.channels || {}}
      slideUid={context.slideUid || ""}
      onStyleChange={customState.onStyleChange}
      isRefreshing={customState.isRefreshing || false}
      dockOverride={state.dock}
      onDockChange={(dock) => updateState({ dock })}
      onSettingsStateChange={customState.onSettingsStateChange}
      getCurrentSettingsRef={customState.getCurrentSettingsRef}
    />
  );
}

/**
 * Panel registration for the Brightness & Contrast Panel
 */
export const brightnessPanelRegistration: PanelRegistration = {
  id: "brightness-contrast",
  name: "Brightness & Contrast",
  icon: SunIcon,
  component: BrightnessPanelWrapper,
  defaultState: {
    isOpen: false,
    dock: "free",
    size: { width: 320, height: 600 },
  },
  enabled: true,
  shortcut: "b",
  order: 30,
};
