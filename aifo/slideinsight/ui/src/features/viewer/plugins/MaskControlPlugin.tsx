// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback } from "react";
import { PaintBrushIcon } from "@heroicons/react/24/outline";
import { ViewerPlugin, PluginAPI } from "./types";
import MaskControlPanel from "@/components/panels/MaskControlPanel";

interface MaskPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
}

function MaskPluginPanel({ api, onClose }: MaskPluginPanelProps) {
  const { context } = api;

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  const handleOpenChange = useCallback(
    (isOpen: boolean) => {
      api.setState({ isOpen });
    },
    [api]
  );

  return (
    <MaskControlPanel
      onClose={onClose}
      mapRef={context.map}
      slideUid={context.slideUid}
      studyUid={context.studyUid}
      dockOverride={api.state.dock}
      onDockChange={handleDockChange}
      openOverride={api.state.isOpen}
      onOpenChange={handleOpenChange}
      annotationSettings={context.studyAnnotationSettings}
      annotationSettingsLoading={context.studyAnnotationSettingsLoading}
      annotationSettingsError={context.studyAnnotationSettingsError}
    />
  );
}

export const MaskControlPlugin: ViewerPlugin = {
  id: "mask-control",
  name: "Mask & Vector Annotations",
  version: "1.0.0",

  button: {
    id: "mask-control-button",
    label: "Mask & Vector Annotations",
    icon: PaintBrushIcon,
    tooltip: "View and manage mask and vector annotations",
    position: "right",
    order: 3,
  },

  panel: {
    id: "mask-control-panel",
    title: "Mask & Vector Annotations",
    defaultSize: { width: 320, height: 560 },
    defaultDock: "left",
    storageKey: "maskControlPanel",
  },

  PanelComponent: MaskPluginPanel,

  onButtonClick: (api: PluginAPI) => {
    console.log("MaskControlPlugin: Button clicked");
    console.log("MaskControlPlugin: API context:", api.context);
    console.log("MaskControlPlugin: API state:", api.state);

    // Toggle panel
    console.log("MaskControlPlugin: Toggling panel state");
    api.setState({ isOpen: !api.state.isOpen });
  },

  onContextChange: (context) => {
    // Could update plugin availability based on context
    // For example, hide/show button based on study permissions
  },
};
