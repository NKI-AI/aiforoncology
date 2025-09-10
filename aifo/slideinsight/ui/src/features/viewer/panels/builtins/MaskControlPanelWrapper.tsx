// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { PaintBrushIcon } from "@heroicons/react/24/outline";
import { PanelProps, PanelRegistration } from "../types";
import MaskControlPanel from "@/components/panels/MaskControlPanel";

/**
 * Wrapper component for the existing MaskControlPanel
 */
function MaskControlPanelWrapper({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  return (
    <MaskControlPanel
      onClose={onClose}
      mapRef={context.mapRef}
      slideUid={context.slideUid}
      studyUid={context.studyUid}
      dockOverride={state.dock}
      onDockChange={(dock) => updateState({ dock })}
      openOverride={state.isOpen}
      onOpenChange={(isOpen) => updateState({ isOpen })}
      annotationSettings={state.customState?.annotationSettings}
      annotationSettingsLoading={state.customState?.annotationSettingsLoading}
      annotationSettingsError={state.customState?.annotationSettingsError}
      onImportComplete={state.customState?.onImportComplete}
    />
  );
}

/**
 * Panel registration for the Mask Control Panel
 */
export const maskControlPanelRegistration: PanelRegistration = {
  id: "mask-control",
  name: "Annotations",
  icon: PaintBrushIcon,
  component: MaskControlPanelWrapper,
  defaultState: {
    isOpen: false,
    dock: "left",
    size: { width: 320, height: 560 },
  },
  enabled: true,
  shortcut: "a",
  order: 10,
};
