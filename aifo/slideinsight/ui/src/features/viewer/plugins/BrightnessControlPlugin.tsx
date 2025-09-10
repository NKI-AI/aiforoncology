// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect, useState } from "react";
import { SunIcon } from "@heroicons/react/24/outline";
import { ViewerPlugin, PluginAPI } from "./types";
import { BrightnessContrastPanel } from "@/features/viewer/components/BrightnessContrastPanel";
import {
  generateSlideStyle,
  SlideStyleOptions,
} from "@/features/viewer/components/map/slideStyleUtils";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
} from "@/components/AlertDialog";

interface BrightnessPluginPanelProps {
  api: PluginAPI;
  onClose: () => void;
}

function BrightnessPluginPanel({ api, onClose }: BrightnessPluginPanelProps) {
  const [hasNonDefaultSettings, setHasNonDefaultSettings] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showUnavailableDialog, setShowUnavailableDialog] = useState(false);

  const handleStyleChange = useCallback(
    (options: SlideStyleOptions) => {
      const { context } = api;
      if (context.slideLayer && context.rawSlideMetadata?.metadata?.channels) {
        const styleOptions = {
          ...options,
          channelMetadata: context.rawSlideMetadata.metadata.channels,
        };
        const style = generateSlideStyle(styleOptions);
        context.slideLayer.setStyle(style);
      }
    },
    [api]
  );

  const handleDockChange = useCallback(
    (dock: "free" | "left") => {
      api.setState({ dock });
    },
    [api]
  );

  // Check if we have fluorescent metadata available
  const hasChannels =
    api.context.rawSlideMetadata?.metadata?.channels &&
    Object.keys(api.context.rawSlideMetadata.metadata.channels).length > 0;

  // Show dialog when panel is opened but no channels are available
  useEffect(() => {
    if (api.state.isOpen && !hasChannels) {
      setShowUnavailableDialog(true);
    }
  }, [api.state.isOpen, hasChannels]);

  return (
    <>
      {/* Only show the brightness panel if channels are available */}
      {hasChannels && (
        <BrightnessContrastPanel
          isOpen={api.state.isOpen}
          onClose={onClose}
          channels={api.context.rawSlideMetadata?.metadata?.channels || {}}
          slideUid={api.context.slideUid || ""}
          onStyleChange={handleStyleChange}
          isRefreshing={isRefreshing}
          dockOverride={api.state.dock}
          onDockChange={handleDockChange}
          onSettingsStateChange={setHasNonDefaultSettings}
        />
      )}

      {/* Show unavailable dialog when appropriate */}
      <AlertDialog
        open={showUnavailableDialog}
        onOpenChange={(open) => {
          setShowUnavailableDialog(open);
          if (!open) {
            // When dialog is closed, also close the panel
            api.setState({ isOpen: false });
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Brightness & Contrast Unavailable
            </AlertDialogTitle>
            <AlertDialogDescription className="space-y-3">
              <p>
                This feature is only available for fluorescent images with
                channel metadata. The current image appears to be an RGB image
                without fluorescent channel information.
              </p>
              <p>
                If you have specific requirements and would like this feature
                implemented for RGB images, please open an issue at{" "}
                <a
                  href="https://github.com/NKI-AI/aiforoncology"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-800 underline"
                >
                  our GitHub repository
                </a>
                .
              </p>
              <p className="text-sm">
                You can also disable this module entirely in the study settings
                to remove it from the viewer toolbar.
              </p>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction
              onClick={() => {
                setShowUnavailableDialog(false);
                api.setState({ isOpen: false });
              }}
            >
              OK
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

export const BrightnessControlPlugin: ViewerPlugin = {
  id: "brightness-control",
  name: "Brightness & Contrast",
  version: "1.0.0",

  button: {
    id: "brightness-control-button",
    label: "Brightness & Contrast",
    icon: SunIcon,
    tooltip: "Adjust brightness and contrast for fluorescent images",
    position: "right",
    order: 1,
  },

  panel: {
    id: "brightness-control-panel",
    title: "Brightness & Contrast",
    defaultSize: { width: 320, height: 600 },
    defaultDock: "free",
    storageKey: "brightnessPanel",
  },

  PanelComponent: BrightnessPluginPanel,

  onButtonClick: (api: PluginAPI) => {
    console.log("BrightnessControlPlugin: Button clicked");
    console.log("BrightnessControlPlugin: API context:", api.context);
    console.log("BrightnessControlPlugin: API state:", api.state);

    // Check if we have fluorescent metadata available before showing panel
    const hasChannels =
      api.context.rawSlideMetadata?.metadata?.channels &&
      Object.keys(api.context.rawSlideMetadata.metadata.channels).length > 0;

    console.log("BrightnessControlPlugin: Has channels:", hasChannels);

    // Always toggle panel - the component will handle showing dialog if needed
    console.log("BrightnessControlPlugin: Toggling panel state");
    api.setState({ isOpen: !api.state.isOpen });
  },

  onContextChange: (context) => {
    // Could update plugin availability based on context
    // For example, hide/show button based on image type
  },
};
