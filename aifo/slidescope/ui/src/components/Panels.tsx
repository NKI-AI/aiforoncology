// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import MaskControlPanel from "./panels/MaskControlPanel";
import SlideInfoPanel from "./panels/SlideInfoPanel";
import HelpPanel from "./panels/HelpPanel";
import { PanelState } from "../hooks/usePanels";

interface PanelsProps {
  visiblePanels: PanelState;
  togglePanel: (panelName: keyof PanelState) => void;
  slideMetadata: any;
}

export default function Panels({
  visiblePanels,
  togglePanel,
  slideMetadata,
}: PanelsProps) {
  return (
    <>
      {visiblePanels.maskControl && (
        <MaskControlPanel onClose={() => togglePanel("maskControl")} />
      )}

      {visiblePanels.slideInfo && (
        <SlideInfoPanel
          slideMetadata={slideMetadata}
          onClose={() => togglePanel("slideInfo")}
        />
      )}

      {visiblePanels.help && <HelpPanel onClose={() => togglePanel("help")} />}
    </>
  );
}
