// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { PencilIcon } from "@heroicons/react/24/outline";
import { PanelProps, PanelRegistration } from "../types";
import AnnotationPanel from "@/features/viewer/components/AnnotationPanel";

/**
 * Wrapper component for the existing AnnotationPanel
 */
function AnnotationPanelWrapper({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  const customState = state.customState || {};

  return (
    <AnnotationPanel
      isOpen={state.isOpen}
      onClose={onClose}
      activeLabel={customState.activeLabel}
      onActiveLabelChange={(label) =>
        updateState({ customState: { ...customState, activeLabel: label } })
      }
      annotations={customState.annotations || []}
      onUpdateAnnotations={(annotations) =>
        updateState({ customState: { ...customState, annotations } })
      }
      selectedId={customState.selectedId}
      onSelect={(id) =>
        updateState({ customState: { ...customState, selectedId: id } })
      }
      hoveredId={customState.hoveredId}
      onHoverIdChange={(id) =>
        updateState({ customState: { ...customState, hoveredId: id } })
      }
      onHoverGroupChange={customState.onHoverGroupChange}
      dockOverride={state.dock}
      onDockChange={(dock) => updateState({ dock })}
      onStartDrawROI={customState.onStartDrawROI}
      onStopDraw={customState.onStopDraw}
      brushActive={customState.brushActive || false}
      brushMode={customState.brushMode || "add"}
      brushSizePx={customState.brushSizePx || 24}
      onStartBrushAdd={customState.onStartBrushAdd}
      onStartBrushErase={customState.onStartBrushErase}
      onStopBrush={customState.onStopBrush}
      onBrushSizeChange={(size) =>
        updateState({ customState: { ...customState, brushSizePx: size } })
      }
      studyUid={context.studyUid}
      annotationSettings={customState.annotationSettings}
      annotationSettingsLoading={customState.annotationSettingsLoading}
      annotationSettingsError={customState.annotationSettingsError}
      annotationDisplaySettings={customState.annotationDisplaySettings}
      onUpdateAnnotationDisplaySettings={
        customState.onUpdateAnnotationDisplaySettings
      }
      onResetAnnotationDisplaySettings={
        customState.onResetAnnotationDisplaySettings
      }
      slideMpp={customState.slideMpp}
    />
  );
}

/**
 * Panel registration for the Annotation Panel
 */
export const annotationPanelRegistration: PanelRegistration = {
  id: "annotation-editor",
  name: "Annotation Editor",
  icon: PencilIcon,
  component: AnnotationPanelWrapper,
  defaultState: {
    isOpen: false,
    dock: "free",
    size: { width: 320, height: 560 },
  },
  enabled: true,
  shortcut: "e",
  order: 20,
};
