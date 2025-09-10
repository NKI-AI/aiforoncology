// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { XMarkIcon, EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { Channel } from "./map/types";
import SlidePanel from "@/components/ui/slide-panel";
import SlidePanelList, {
  SlidePanelListItem,
} from "@/components/ui/slide-panel-list";
import {
  loadBrightnessSettings,
  saveBrightnessSettings,
  loadChannelSettings,
  saveChannelSettings,
  hasNonDefaultBrightnessSettings,
  type BrightnessContrastSettings,
} from "@/features/viewer/utils/brightnessContrastStorage";

interface ChannelState extends Channel {
  id: string;
  visible: boolean;
  min: number;
  max: number;
}

interface BrightnessContrastPanelProps {
  isOpen: boolean;
  onClose: () => void;
  channels: Record<string, Channel>;
  slideUid: string;
  onStyleChange?: (options: {
    showGrayscale: boolean;
    invertBackground: boolean;
    gamma: number;
    channels: Record<
      string,
      {
        visible: boolean;
        min: number;
        max: number;
      }
    >;
  }) => void;
  isRefreshing?: boolean; // New prop to indicate if tiles are refreshing
  // Optional external docking control. When provided, component will not manage internal docking
  dockOverride?: "free" | "left";
  onDockChange?: (dock: "free" | "left") => void;
  // Optional callback to report current settings state (for non-default indicator)
  onSettingsStateChange?: (hasNonDefaultSettings: boolean) => void;
  // Function to get current settings for auto-apply
  getCurrentSettingsRef?: React.MutableRefObject<(() => any) | null>;
}

export function BrightnessContrastPanel({
  isOpen,
  onClose,
  channels,
  slideUid,
  onStyleChange,
  isRefreshing = false,
  dockOverride,
  onDockChange,
  onSettingsStateChange,
  getCurrentSettingsRef,
}: BrightnessContrastPanelProps) {
  const [selectedChannelId, setSelectedChannelId] = useState<string | null>(
    null
  );

  // Load settings from localStorage using utility functions
  const [showGrayscale, setShowGrayscale] = useState(() => {
    const settings = loadBrightnessSettings();
    return settings.showGrayscale;
  });
  const [invertBackground, setInvertBackground] = useState(() => {
    const settings = loadBrightnessSettings();
    return settings.invertBackground;
  });
  const [globalGamma, setGlobalGamma] = useState(() => {
    const settings = loadBrightnessSettings();
    return settings.globalGamma;
  });

  // Keyboard navigation is now handled by SlidePanelList

  // Initialize channel states with saved settings
  const [channelStates, setChannelStates] = useState<
    Record<string, ChannelState>
  >(() => {
    const initialStates: Record<string, ChannelState> = {};
    const channelIds = Object.keys(channels);
    const savedChannelSettings = loadChannelSettings(slideUid, channelIds);

    Object.entries(channels).forEach(([id, channel]) => {
      const savedChannel = savedChannelSettings[id] || {
        visible: true,
        min: 0,
        max: 65535,
      };
      initialStates[id] = {
        ...channel,
        id,
        visible: savedChannel.visible,
        min: savedChannel.min,
        max: savedChannel.max,
      };
    });
    return initialStates;
  });

  // Update channel states when channels prop changes (only if channels exist)
  useEffect(() => {
    const channelEntries = Object.entries(channels);
    if (channelEntries.length > 0) {
      const channelIds = Object.keys(channels);
      const savedChannelSettings = loadChannelSettings(slideUid, channelIds);

      const newStates: Record<string, ChannelState> = {};
      channelEntries.forEach(([id, channel]) => {
        const savedChannel = savedChannelSettings[id] || {
          visible: true,
          min: 0,
          max: 65535,
        };
        newStates[id] = {
          ...channel,
          id,
          visible: savedChannel.visible,
          min: savedChannel.min,
          max: savedChannel.max,
        };
      });
      setChannelStates(newStates);
    }
  }, [channels, slideUid]);

  // Auto-select first channel if none selected
  useEffect(() => {
    if (!selectedChannelId && Object.keys(channelStates).length > 0) {
      setSelectedChannelId(Object.keys(channelStates)[0]);
    }
  }, [selectedChannelId, channelStates]);

  // Select-all toggle applied to the currently filtered subset (provided by SlidePanelList)
  const toggleAllFiltered = (
    filtered: ChannelState[],
    newVisibility: boolean
  ) => {
    setChannelStates((prev) => {
      const newStates = { ...prev };
      filtered.forEach((channel) => {
        newStates[channel.id] = {
          ...newStates[channel.id],
          visible: newVisibility,
        };
      });
      return newStates;
    });
  };

  // Panel dragging/resize handled by SlidePanel

  // Handle channel visibility toggle
  const toggleChannelVisibility = (channelId: string) => {
    setChannelStates((prev) => {
      const newVisibility = !prev[channelId].visible;
      return {
        ...prev,
        [channelId]: {
          ...prev[channelId],
          visible: newVisibility,
        },
      };
    });
  };

  // Handle channel parameter changes
  const updateChannelParam = (
    channelId: string,
    param: keyof ChannelState,
    value: number
  ) => {
    setChannelStates((prev) => ({
      ...prev,
      [channelId]: {
        ...prev[channelId],
        [param]: value,
      },
    }));
  };

  // SlidePanel handles focus and Escape; SlidePanelList handles arrows and autoscroll

  // Update style when settings change
  const updateStyle = useCallback(() => {
    if (onStyleChange) {
      // Convert channel states to the format expected by the callback
      const channelParams: Record<
        string,
        { visible: boolean; min: number; max: number }
      > = {};
      Object.values(channelStates).forEach((channel) => {
        channelParams[channel.id] = {
          visible: channel.visible,
          min: channel.min,
          max: channel.max,
        };
      });

      onStyleChange({
        showGrayscale,
        invertBackground,
        gamma: globalGamma,
        channels: channelParams,
      });
    }
  }, [
    onStyleChange,
    showGrayscale,
    invertBackground,
    globalGamma,
    channelStates,
  ]);

  // Update style when relevant settings change
  useEffect(() => {
    updateStyle();
  }, [updateStyle]);

  // Check for non-default settings and report to parent
  useEffect(() => {
    if (onSettingsStateChange) {
      const currentSettings: BrightnessContrastSettings = {
        showGrayscale,
        invertBackground,
        globalGamma,
        channels: Object.fromEntries(
          Object.values(channelStates).map((ch) => [
            ch.id,
            { visible: ch.visible, min: ch.min, max: ch.max },
          ])
        ),
      };
      const hasNonDefault = hasNonDefaultBrightnessSettings(currentSettings);
      onSettingsStateChange(hasNonDefault);
    }
  }, [
    showGrayscale,
    invertBackground,
    globalGamma,
    channelStates,
    onSettingsStateChange,
  ]);

  // Save settings to localStorage when they change
  useEffect(() => {
    saveBrightnessSettings({ showGrayscale });
  }, [showGrayscale]);

  useEffect(() => {
    saveBrightnessSettings({ invertBackground });
  }, [invertBackground]);

  useEffect(() => {
    saveBrightnessSettings({ globalGamma });
  }, [globalGamma]);

  // Save channel settings when they change
  useEffect(() => {
    const channelSettings = Object.fromEntries(
      Object.values(channelStates).map((ch) => [
        ch.id,
        { visible: ch.visible, min: ch.min, max: ch.max },
      ])
    );
    saveChannelSettings(slideUid, channelSettings);
  }, [channelStates, slideUid]);

  // Auto-adjust function (placeholder)
  const handleAuto = () => {
    // TODO: Implement auto-adjustment based on histogram
  };

  // Reset function
  const handleReset = () => {
    setChannelStates((prev) => {
      const newStates = { ...prev };
      Object.keys(newStates).forEach((id) => {
        newStates[id] = {
          ...newStates[id],
          visible: true, // Reset visibility to true
          min: 0,
          max: 65535,
        };
      });
      return newStates;
    });
    setGlobalGamma(1.0);
    setShowGrayscale(false);
    setInvertBackground(false);
  };

  // Function to get current settings (can be used by parent to auto-apply)
  const getCurrentSettings = useCallback(() => {
    return {
      showGrayscale,
      invertBackground,
      gamma: globalGamma,
      channels: Object.fromEntries(
        Object.values(channelStates).map((ch) => [
          ch.id,
          { visible: ch.visible, min: ch.min, max: ch.max },
        ])
      ),
    };
  }, [showGrayscale, invertBackground, globalGamma, channelStates]);

  // Expose getCurrentSettings through ref for parent access
  useEffect(() => {
    if (getCurrentSettingsRef) {
      getCurrentSettingsRef.current = getCurrentSettings;
    }
  }, [getCurrentSettings, getCurrentSettingsRef]);

  if (!isOpen) return null;

  // Check if we have any channels loaded yet
  const hasChannels = Object.keys(channelStates).length > 0;

  const PanelBody = (
    <>
      <SlidePanelList<ChannelState>
        items={Object.values(channelStates)}
        getFilterText={(ch) => `${ch.biomarker} ${ch.name}`}
        searchThreshold={5}
        searchPlaceholder="Filter channels by name"
        selectAllLabel="Select all"
        getItemChecked={(ch) => ch.visible}
        onToggleAll={(filtered, newChecked) =>
          toggleAllFiltered(filtered, newChecked)
        }
        getItemId={(ch) => ch.id}
        activeId={selectedChannelId}
        onActiveIdChange={(id) => setSelectedChannelId(id)}
        onToggleActive={(id) => toggleChannelVisibility(id)}
        className="flex-shrink"
      >
        {(filtered) => {
          return hasChannels ? (
            <ul className="space-y-0.5 pb-2">
              {filtered.map((channel) => (
                <SlidePanelListItem
                  key={channel.id}
                  active={selectedChannelId === channel.id}
                  dataActive={selectedChannelId === channel.id}
                  leftColor={channel.color}
                  primary={channel.biomarker}
                  secondary={channel.name}
                  onClick={() => setSelectedChannelId(channel.id)}
                  rightSlot={
                    <button
                      className="p-1 rounded hover:bg-accent text-muted-foreground"
                      title={channel.visible === false ? "Show" : "Hide"}
                      aria-label={channel.visible === false ? "Show" : "Hide"}
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleChannelVisibility(channel.id);
                      }}
                    >
                      {channel.visible === false ? (
                        <EyeIcon className="h-4 w-4" />
                      ) : (
                        <EyeSlashIcon className="h-4 w-4" />
                      )}
                    </button>
                  }
                />
              ))}
            </ul>
          ) : (
            <div className="flex items-center justify-center h-32">
              <span className="text-sm text-muted-foreground">
                No fluorescent channels available
              </span>
            </div>
          );
        }}
      </SlidePanelList>

      {/* Fixed Controls Section */}
      <div className="flex-shrink-0 border-t border-border">
        {/* Settings */}
        <div className="px-3 py-1 space-y-1.5">
          <div className="flex items-center justify-center space-x-4 text-xs">
            <label className="inline-flex items-center space-x-1">
              <Checkbox
                checked={showGrayscale}
                onCheckedChange={(checked) => setShowGrayscale(!!checked)}
                className="h-3.5 w-3.5"
              />
              <span>Show grayscale</span>
            </label>
            <label className="inline-flex items-center space-x-1">
              <Checkbox
                checked={invertBackground}
                onCheckedChange={(checked) => setInvertBackground(!!checked)}
                className="h-3.5 w-3.5"
              />
              <span>Invert background</span>
            </label>
          </div>
        </div>

        <hr className="border-border" />

        {/* Contrast Controls */}
        <div className="px-3 py-2 space-y-3">
          {/* Global gamma control */}
          <div className="space-y-1">
            <div className="flex items-center justify-between text-xs">
              <span>Global gamma</span>
              <span>{globalGamma.toFixed(1)}</span>
            </div>
            <input
              type="range"
              min="0.1"
              max="5"
              step="0.1"
              value={globalGamma}
              onChange={(e) => setGlobalGamma(parseFloat(e.target.value))}
              className="w-full h-1.5 bg-input rounded-lg appearance-none cursor-pointer"
            />
          </div>

          {selectedChannelId && channelStates[selectedChannelId] && (
            <>
              <hr className="border-border" />

              {/* Selected channel controls */}
              <div className="space-y-3">
                {/* Min control */}
                <div className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <span>Min</span>
                    <span>
                      {channelStates[selectedChannelId].min.toFixed(0)}
                    </span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="65535"
                    value={channelStates[selectedChannelId].min}
                    onChange={(e) =>
                      updateChannelParam(
                        selectedChannelId,
                        "min",
                        parseFloat(e.target.value)
                      )
                    }
                    className="w-full h-1.5 bg-input rounded-lg appearance-none cursor-pointer"
                  />
                </div>

                {/* Max control */}
                <div className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <span>Max</span>
                    <span>
                      {channelStates[selectedChannelId].max.toFixed(0)}
                    </span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="65535"
                    value={channelStates[selectedChannelId].max}
                    onChange={(e) =>
                      updateChannelParam(
                        selectedChannelId,
                        "max",
                        parseFloat(e.target.value)
                      )
                    }
                    className="w-full h-1.5 bg-input rounded-lg appearance-none cursor-pointer"
                  />
                </div>
              </div>
            </>
          )}
        </div>

        {/* Buttons */}
        <div className="px-3 py-2 flex justify-end space-x-1.5 border-t border-border">
          <Button
            size="sm"
            variant="secondary"
            onClick={handleAuto}
            className="px-2 py-1 bg-secondary hover:bg-input text-xs h-6"
          >
            Auto
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={handleReset}
            className="px-2 py-1 bg-secondary hover:bg-input text-xs h-6"
          >
            Reset
          </Button>
        </div>
      </div>
    </>
  );

  return (
    <>
      {/* Mobile bottom sheet */}
      <div className="fixed inset-0 z-50 md:hidden">
        <div className="absolute inset-0 bg-black/40" onClick={onClose} />
        <div className="absolute inset-x-0 bottom-0 bg-card text-card-foreground rounded-t-2xl shadow-xl select-none flex flex-col max-h-[85vh] h-[70vh] min-h-0">
          {/* Mobile header */}
          <div className="bg-muted px-3 py-2 text-base font-semibold flex items-center justify-between rounded-t-2xl">
            <div className="flex items-center space-x-2">
              <span className="mx-auto h-1 w-8 rounded bg-muted-foreground" />
              <span>Brightness & contrast</span>
            </div>
            <button
              onClick={onClose}
              className="w-6 h-6 text-muted-foreground hover:text-foreground transition-colors"
              aria-label="Close"
            >
              <XMarkIcon className="w-5 h-5" />
            </button>
          </div>
          {PanelBody}
        </div>
      </div>

      {/* Desktop panel */}
      <SlidePanel
        isOpen={isOpen}
        onClose={onClose}
        dockOverride={dockOverride}
        onDockChange={onDockChange}
        storageKey={`brightnessPanel_${slideUid}`}
        defaultSize={{ width: 320, height: 600 }}
        className="font-sans"
      >
        <SlidePanel.Header title="Brightness & contrast" onClose={onClose} />
        {PanelBody}
      </SlidePanel>
    </>
  );
}
