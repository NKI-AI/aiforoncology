// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { SettingsIcon } from "@/components/icons";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { Switch } from "@/components/ui/switch";
import { TileFormat } from "./map/SlideImage";

interface ViewerSettingsPanelProps {
  showSettings: boolean;
  panSensitivity: number;
  zoomSensitivity: number;
  quality?: number; // Quality parameter for JPG/JPEG XL (0-100)
  isFluorescent?: boolean; // Whether the image is fluorescent
  showMeasurementBar?: boolean; // Whether to show the measurement bar
  onToggleSettings: () => void;
  // Crosshair toggle moved out of this panel
  // Brightness control is now handled by the plugin system
  // onToggleAnnotations is now handled by the plugin system
  // onToggleRegionPanel is now handled by the plugin system
  onPanSensitivityChange: (value: number) => void;
  onZoomSensitivityChange: (value: number) => void;
  onQualityChange?: (quality: number | undefined) => void;
  onMeasurementBarChange?: (show: boolean) => void;
  onResetSettings?: () => void; // New prop for reset functionality
}

export function ViewerSettingsPanel({
  showSettings,
  panSensitivity,
  zoomSensitivity,
  quality,
  isFluorescent = false,
  showMeasurementBar = true,
  onToggleSettings,
  onPanSensitivityChange,
  onZoomSensitivityChange,
  onQualityChange,
  onMeasurementBarChange,
  onResetSettings,
}: ViewerSettingsPanelProps) {
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
    <>
      {/* Floating action buttons container - positioned above StatusBar */}
      <div className="absolute bottom-6 right-2 flex flex-col-reverse gap-2 z-[9998]">
        {/* Settings Button */}
        <button
          onClick={onToggleSettings}
          className={`w-10 h-10 sm:w-8 sm:h-8 rounded-full shadow-lg transition-all duration-200 hover:scale-110 ${
            showSettings
              ? "bg-muted-foreground hover:bg-muted-foreground/90 text-background"
              : "bg-card hover:bg-accent text-card-foreground border border-border"
          }`}
          title={`Viewer Settings (${
            isFluorescent ? "Fluorescence" : "Regular"
          } Image)`}
        >
          <SettingsIcon className="w-5 h-5 mx-auto" />
        </button>

        {/* Brightness control is now handled by the plugin system */}

        {/* Overlay button is now handled by the plugin system */}
        {/* Annotation Editor is now handled by the plugin system */}
        {/* Region Panel is now handled by the plugin system */}
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <div className="absolute bottom-6 right-14 sm:right-12 bg-card border border-border rounded-lg shadow-xl p-3 z-[9999] w-72 max-w-[calc(100vw-4rem)] max-h-[calc(100vh-8rem)] overflow-y-auto">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-semibold text-card-foreground">
              Viewer Settings
            </h3>
            <button
              onClick={onToggleSettings}
              className="w-6 h-6 text-muted-foreground hover:text-foreground transition-colors"
            >
              <XMarkIcon className="w-4 h-4" />
            </button>
          </div>

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
            {onQualityChange && (
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
                          onChange={(e) =>
                            onQualityChange(parseInt(e.target.value))
                          }
                          className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
                        />
                        <span className="text-xs text-muted-foreground font-mono w-6 text-right">
                          {quality}
                        </span>
                        <button
                          onClick={() => onQualityChange(undefined)}
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
                          onQualityChange(
                            currentTileFormat === "jxl" ? 100 : 85
                          )
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
            )}

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
                  value={panSensitivity}
                  onChange={(e) =>
                    onPanSensitivityChange(parseFloat(e.target.value))
                  }
                  className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
                />
                <span className="text-xs text-muted-foreground font-mono w-10 text-right">
                  {panSensitivity >= 10
                    ? panSensitivity.toFixed(0)
                    : panSensitivity.toFixed(1)}
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
                  value={zoomSensitivity}
                  onChange={(e) =>
                    onZoomSensitivityChange(parseFloat(e.target.value))
                  }
                  className="flex-1 h-2 bg-muted rounded appearance-none cursor-pointer"
                />
                <span className="text-xs text-muted-foreground font-mono w-10 text-right">
                  {zoomSensitivity >= 10
                    ? zoomSensitivity.toFixed(0)
                    : zoomSensitivity.toFixed(1)}
                  x
                </span>
              </div>
              <div className="flex justify-between text-xs text-muted-foreground mt-1">
                <span>Slow</span>
                <span>Fast</span>
              </div>
            </div>

            {/* Measurement Bar Toggle */}
            {onMeasurementBarChange && (
              <div>
                <label className="flex items-center justify-between cursor-pointer">
                  <span className="text-xs font-medium text-foreground">
                    Show Measurement Bar
                  </span>
                  <Switch
                    checked={showMeasurementBar}
                    onCheckedChange={onMeasurementBarChange}
                  />
                </label>
                <p className="text-xs text-muted-foreground mt-1">
                  Toggle the scale bar in the bottom left corner
                </p>
              </div>
            )}

            {/* Settings Storage Info */}
            <div className="text-xs text-muted-foreground pt-2 border-t border-border">
              <p>
                Settings saved separately for{" "}
                {isFluorescent ? "fluorescence" : "regular"} images
              </p>
            </div>

            {/* Reset Settings Button */}
            {onResetSettings && (
              <button
                onClick={onResetSettings}
                className="w-full px-2 py-1 text-xs font-medium rounded transition-colors bg-destructive/10 text-destructive border border-destructive/20 hover:bg-destructive/20"
                title="Reset settings to default for this image type"
              >
                Reset Settings
              </button>
            )}
          </div>
        </div>
      )}
    </>
  );
}
