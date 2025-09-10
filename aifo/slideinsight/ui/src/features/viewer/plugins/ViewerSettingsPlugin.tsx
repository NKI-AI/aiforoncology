// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback } from "react";
import { Cog6ToothIcon } from "@heroicons/react/24/outline";
import { ViewerPlugin, PluginAPI } from "./types";
import SlidePanel from "@/components/ui/slide-panel";
import { Switch } from "@/components/ui/switch";
import { useViewerState } from "@/features/viewer/hooks";
import { resetViewerSettings } from "@/features/viewer/utils/viewerSettingsStorage";
import { useLocalStorageState } from "@/hooks/useLocalStorageState";

interface ViewerSettingsPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
}

function ViewerSettingsPluginPanel({
  api,
  onClose,
}: ViewerSettingsPluginPanelProps) {
  const { context } = api;

  // Determine if image is fluorescent
  const isFluorescent =
    context.rawSlideMetadata?.imageTypeId === "img_type_fluor";

  // Viewer state settings
  const {
    state: viewerState,
    toggleSettings,
    setPanSensitivity,
    setZoomSensitivity,
  } = useViewerState(isFluorescent);

  // Quality state (could be lifted to context if needed globally)
  const [quality, setQuality] = React.useState<number | undefined>(undefined);

  // Measurement bar visibility setting with localStorage persistence
  const [showMeasurementBar, setShowMeasurementBar] =
    useLocalStorageState<boolean>("showMeasurementBar", true);

  // Update context when settings change
  React.useEffect(() => {
    if (context.onViewerSettingsChange) {
      context.onViewerSettingsChange({
        panSensitivity: viewerState.panSensitivity,
        zoomSensitivity: viewerState.zoomSensitivity,
        quality,
        showMeasurementBar,
      });
    }
  }, [
    context,
    viewerState.panSensitivity,
    viewerState.zoomSensitivity,
    quality,
    showMeasurementBar,
  ]);

  const handleResetSettings = useCallback(() => {
    if (isFluorescent !== undefined) {
      const resetSettings = resetViewerSettings(isFluorescent);
      setPanSensitivity(resetSettings.panSensitivity);
      setZoomSensitivity(resetSettings.zoomSensitivity);
      setQuality(resetSettings.quality);
      // Reset measurement bar to default
      setShowMeasurementBar(true);
    }
  }, [
    isFluorescent,
    setPanSensitivity,
    setZoomSensitivity,
    setShowMeasurementBar,
  ]);

  const handleMeasurementBarChange = useCallback(
    (show: boolean) => {
      setShowMeasurementBar(show);
    },
    [setShowMeasurementBar]
  );

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  const handleClose = () => {
    onClose();
  };

  // Determine current tile format based on simple rule
  const currentTileFormat = isFluorescent ? "jxl" : "jpg";
  const supportsQuality =
    currentTileFormat === "jpg" || currentTileFormat === "jxl";

  // Get quality range based on format
  const getQualityRange = () => {
    if (currentTileFormat === "jxl") {
      return { min: 10, max: 100, step: 5, default: 100 }; // JXL supports wide quality range, default to lossless
    } else {
      return { min: 60, max: 100, step: 5, default: 85 }; // Standard JPEG range
    }
  };

  const qualityRange = getQualityRange();

  // Get image type indicator
  const getImageTypeIndicator = () => {
    if (isFluorescent) {
      return {
        icon: "🧬",
        label: "Fluorescence Image",
        description: "Multi-channel imaging using JXL tiles",
        format: "JXL",
        color: "purple",
      };
    }
    return {
      icon: "🖼️",
      label: "Regular Image",
      description: "Standard RGB imaging using JPG tiles",
      format: "JPG",
      color: "blue",
    };
  };

  const imageType = getImageTypeIndicator();

  return (
    <SlidePanel
      isOpen={api.state.isOpen}
      onClose={handleClose}
      dockOverride={api.state.dock}
      onDockChange={handleDockChange}
      storageKey="viewerSettingsPanel"
      defaultSize={{ width: 288, height: 400 }}
    >
      <SlidePanel.Header title="Viewer Settings" onClose={handleClose} />
      <div className="flex-1 min-h-0 overflow-y-auto p-3">
        <div className="space-y-3">
          {/* Image Type Indicator */}
          <div className="bg-muted border border-border p-2 rounded-lg">
            <div className="flex items-center justify-between mb-1">
              <div className="flex items-center space-x-2">
                <span className="text-base">{imageType.icon}</span>
                <span className="text-xs font-medium text-foreground">
                  {imageType.label}
                </span>
              </div>
              <span className="px-2 py-0.5 text-xs font-mono rounded bg-secondary text-secondary-foreground">
                {imageType.format}
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {imageType.description}
            </p>
          </div>

          {/* Quality Control */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">
              Compression Quality
            </label>

            {supportsQuality ? (
              quality !== undefined ? (
                <div className="space-y-1">
                  <div className="flex items-center space-x-2">
                    <input
                      type="range"
                      min={qualityRange.min}
                      max={qualityRange.max}
                      step={qualityRange.step}
                      value={quality}
                      onChange={(e) => setQuality(parseInt(e.target.value))}
                      className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
                    />
                    <span className="text-xs text-muted-foreground font-mono w-6 text-right">
                      {quality}
                    </span>
                    <button
                      onClick={() => setQuality(undefined)}
                      className="text-xs text-primary hover:text-primary/80"
                      title="Use lossless compression"
                    >
                      Lossless
                    </button>
                  </div>
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>Smaller</span>
                    <span>Better</span>
                  </div>
                </div>
              ) : (
                <div className="space-y-1">
                  <div className="text-xs text-emerald-600 font-medium">
                    Using lossless compression
                  </div>
                  <button
                    onClick={() =>
                      setQuality(currentTileFormat === "jxl" ? 100 : 85)
                    }
                    className="text-xs text-primary hover:text-primary/80 underline"
                  >
                    Enable compression controls
                  </button>
                </div>
              )
            ) : (
              <div className="text-xs text-muted-foreground">
                Lossless format
              </div>
            )}
          </div>

          {/* Pan Sensitivity Setting */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">
              Pan Sensitivity
            </label>
            <div className="flex items-center space-x-2">
              <input
                type="range"
                min="0.1"
                max="100.0"
                step="0.1"
                value={viewerState.panSensitivity}
                onChange={(e) => setPanSensitivity(parseFloat(e.target.value))}
                className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
              />
              <span className="text-xs text-muted-foreground font-mono w-10 text-right">
                {viewerState.panSensitivity >= 10
                  ? viewerState.panSensitivity.toFixed(0)
                  : viewerState.panSensitivity.toFixed(1)}
                x
              </span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>Precise</span>
              <span>Rapid</span>
            </div>
          </div>

          {/* Zoom Sensitivity Setting */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">
              Zoom Sensitivity
            </label>
            <div className="flex items-center space-x-2">
              <input
                type="range"
                min="0.1"
                max="100.0"
                step="0.1"
                value={viewerState.zoomSensitivity}
                onChange={(e) => setZoomSensitivity(parseFloat(e.target.value))}
                className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
              />
              <span className="text-xs text-muted-foreground font-mono w-10 text-right">
                {viewerState.zoomSensitivity >= 10
                  ? viewerState.zoomSensitivity.toFixed(0)
                  : viewerState.zoomSensitivity.toFixed(1)}
                x
              </span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>Slow</span>
              <span>Fast</span>
            </div>
          </div>

          {/* Measurement Bar Toggle */}
          <div>
            <label className="flex items-center justify-between cursor-pointer">
              <span className="text-xs font-medium text-foreground">
                Show Measurement Bar
              </span>
              <Switch
                checked={showMeasurementBar}
                onCheckedChange={handleMeasurementBarChange}
              />
            </label>
            <p className="text-xs text-muted-foreground mt-1">
              Toggle the scale bar in the bottom left corner
            </p>
          </div>

          {/* Settings Storage Info */}
          <div className="text-xs text-muted-foreground pt-2 border-t border-border">
            <p>
              Settings saved separately for{" "}
              {isFluorescent ? "fluorescence" : "regular"} images
            </p>
          </div>

          {/* Reset Settings Button */}
          <button
            onClick={handleResetSettings}
            className="w-full px-2 py-1 text-xs font-medium rounded transition-colors bg-destructive/10 text-destructive border border-destructive/20 hover:bg-destructive/20"
            title="Reset settings to default for this image type"
          >
            Reset Settings
          </button>
        </div>
      </div>
    </SlidePanel>
  );
}

export const ViewerSettingsPlugin: ViewerPlugin = {
  id: "viewer-settings",
  name: "Viewer Settings",
  version: "1.0.0",

  button: {
    id: "viewer-settings-button",
    label: "Viewer Settings",
    icon: Cog6ToothIcon,
    tooltip: "Adjust viewer settings and controls",
    position: "right",
    order: 0, // Make it first in the dock
  },

  panel: {
    id: "viewer-settings-panel",
    title: "Viewer Settings",
    defaultSize: { width: 288, height: 400 }, // Match the original panel size
    defaultDock: "left",
    storageKey: "viewerSettingsPanel",
  },

  PanelComponent: ViewerSettingsPluginPanel,

  onButtonClick: (api: PluginAPI) => {
    console.log("ViewerSettingsPlugin: Button clicked");
    // Toggle panel
    api.setState({ isOpen: !api.state.isOpen });
  },

  onContextChange: (context) => {
    // Plugin is always available, no special context requirements
  },
};
