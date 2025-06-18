// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useState } from "react";

export interface PanelState {
  maskControl: boolean;
  slideInfo: boolean;
  help: boolean;
}

export function usePanels(initialState?: Partial<PanelState>) {
  const [visiblePanels, setVisiblePanels] = useState<PanelState>({
    maskControl: false,
    slideInfo: false,
    help: false,
    ...initialState,
  });

  const togglePanel = (panelName: keyof PanelState) => {
    setVisiblePanels((prev) => ({
      ...prev,
      [panelName]: !prev[panelName],
    }));
  };

  const closeAllPanels = () => {
    setVisiblePanels({
      maskControl: false,
      slideInfo: false,
      help: false,
    });
  };

  return {
    visiblePanels,
    togglePanel,
    closeAllPanels,
  };
}
