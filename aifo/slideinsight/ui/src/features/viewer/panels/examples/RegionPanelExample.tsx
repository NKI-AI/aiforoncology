// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState } from "react";
import { RectangleGroupIcon } from "@heroicons/react/24/outline";
import { PanelProps, PanelRegistration } from "../types";
import SlidePanel from "@/components/ui/slide-panel";

/**
 * Example Region Panel component
 * This demonstrates how to create a new panel using the panel registry system
 */
function RegionPanelExample({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  const [regions, setRegions] = useState<
    Array<{ id: string; name: string; visible: boolean }>
  >([]);
  const [isDrawing, setIsDrawing] = useState(false);

  const handleStartDrawing = () => {
    setIsDrawing(true);
    // Here you would integrate with OpenLayers to start drawing
    console.log("Starting region drawing on map:", context.mapRef);
  };

  const handleStopDrawing = () => {
    setIsDrawing(false);
    // Here you would stop the drawing interaction
  };

  const addRegion = () => {
    const newRegion = {
      id: `region-${Date.now()}`,
      name: `Region ${regions.length + 1}`,
      visible: true,
    };
    setRegions([...regions, newRegion]);
  };

  const toggleRegionVisibility = (id: string) => {
    setRegions(
      regions.map((r) => (r.id === id ? { ...r, visible: !r.visible } : r))
    );
  };

  const deleteRegion = (id: string) => {
    setRegions(regions.filter((r) => r.id !== id));
  };

  return (
    <SlidePanel
      isOpen={state.isOpen}
      onClose={onClose}
      dockOverride={state.dock}
      onDockChange={(dock) => updateState({ dock })}
      storageKey="regionPanel"
      defaultSize={state.size}
    >
      <SlidePanel.Header title="Regions" onClose={onClose} />

      <div className="flex-1 min-h-0 overflow-hidden">
        <div className="p-3 space-y-3">
          {/* Controls */}
          <div className="space-y-2">
            <button
              onClick={isDrawing ? handleStopDrawing : handleStartDrawing}
              className={`w-full px-3 py-2 text-sm rounded transition-colors ${
                isDrawing
                  ? "bg-red-600 hover:bg-red-700 text-white"
                  : "bg-primary hover:bg-primary/90 text-primary-foreground"
              }`}
            >
              {isDrawing ? "Stop Drawing" : "Draw Region"}
            </button>

            <button
              onClick={addRegion}
              className="w-full px-3 py-2 text-sm rounded bg-secondary hover:bg-secondary/80 text-secondary-foreground"
            >
              Add Test Region
            </button>
          </div>

          {/* Region List */}
          <div className="space-y-1">
            <h3 className="text-sm font-medium">Regions ({regions.length})</h3>

            {regions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p className="text-sm">No regions yet</p>
                <p className="text-xs mt-1">
                  Draw or add regions to get started
                </p>
              </div>
            ) : (
              <div className="space-y-1">
                {regions.map((region) => (
                  <div
                    key={region.id}
                    className="flex items-center justify-between p-2 rounded border border-border bg-card"
                  >
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <span className="text-sm truncate">{region.name}</span>
                    </div>

                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => toggleRegionVisibility(region.id)}
                        className="p-1 rounded hover:bg-accent text-muted-foreground"
                        title={region.visible ? "Hide region" : "Show region"}
                      >
                        {region.visible ? "👁️" : "👁️‍🗨️"}
                      </button>

                      <button
                        onClick={() => deleteRegion(region.id)}
                        className="p-1 rounded hover:bg-accent text-muted-foreground"
                        title="Delete region"
                      >
                        🗑️
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Debug Info */}
          <div className="mt-4 p-2 rounded bg-muted text-xs">
            <div>Slide: {context.slideUid}</div>
            <div>Study: {context.studyUid || "None"}</div>
            <div>Map: {context.mapRef ? "Ready" : "Not ready"}</div>
            <div>Dock: {state.dock}</div>
          </div>
        </div>
      </div>
    </SlidePanel>
  );
}

/**
 * Panel registration for the example Region Panel
 */
export const regionPanelRegistration: PanelRegistration = {
  id: "regions",
  name: "Regions",
  icon: RectangleGroupIcon,
  component: RegionPanelExample,
  defaultState: {
    isOpen: false,
    dock: "left",
    size: { width: 280, height: 400 },
  },
  enabled: true,
  shortcut: "r",
  order: 40,
};
