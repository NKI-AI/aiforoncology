// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect, useState } from "react";
import { SunIcon } from "@heroicons/react/24/outline";
import { BasePlugin, PluginAPI, PluginContext } from "./base";
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

/**
 * Panel component for the brightness control plugin
 */
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

  // Ensure dock position is compatible with BrightnessContrastPanel
  const compatibleDock: "free" | "left" =
    api.state.dock === "left" ? "left" : "free";

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
          dockOverride={compatibleDock}
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

/**
 * Class-based brightness control plugin
 * Demonstrates the new plugin architecture with proper separation of concerns
 */
export class BrightnessControlPluginV2 extends BasePlugin {
  public readonly id = "brightness-control-v2";
  public readonly name = "Brightness & Contrast";
  public readonly version = "2.0.0";

  protected setupDefaultCapabilities(): void {
    // Declare plugin capabilities
    this.addCapability("hasPanel", true);
    this.addCapability("hasButton", true);
    this.addCapability("requiresSlideMetadata", true);
    this.addCapability("persistsState", true);

    // Configure button
    this.setButton({
      id: "brightness-control-button-v2",
      label: "Brightness & Contrast",
      icon: SunIcon,
      tooltip: "Adjust brightness and contrast for fluorescent images",
      position: "right",
      order: 1,
    });

    // Configure panel
    this.setPanel({
      id: "brightness-control-panel-v2",
      title: "Brightness & Contrast",
      defaultSize: { width: 320, height: 600 },
      defaultDock: "free",
      storageKey: "brightnessPanel",
      resizable: true,
      closable: true,
    });

    // Set up lifecycle hooks
    this.setLifecycleHooks({
      onContextChange: this.onContextChangeHook.bind(this),
      onButtonClick: this.onButtonClickHook.bind(this),
    });
  }

  protected async onInitialize(api: PluginAPI): Promise<void> {
    console.log("BrightnessControlPluginV2: Initializing plugin");
    // Plugin-specific initialization logic can go here
  }

  protected async onDestroy(): Promise<void> {
    console.log("BrightnessControlPluginV2: Destroying plugin");
    // Plugin-specific cleanup logic can go here
  }

  protected async handleContextChange(context: PluginContext): Promise<void> {
    // React to context changes if needed
    console.log("BrightnessControlPluginV2: Context changed", context);
  }

  protected async handleButtonClick(api: PluginAPI): Promise<void> {
    console.log("BrightnessControlPluginV2: Button clicked");
    console.log("BrightnessControlPluginV2: API context:", api.context);
    console.log("BrightnessControlPluginV2: API state:", api.state);

    // Check if we have fluorescent metadata available before showing panel
    const hasChannels =
      api.context.rawSlideMetadata?.metadata?.channels &&
      Object.keys(api.context.rawSlideMetadata.metadata.channels).length > 0;

    console.log("BrightnessControlPluginV2: Has channels:", hasChannels);

    // Always toggle panel - the component will handle showing dialog if needed
    console.log("BrightnessControlPluginV2: Toggling panel state");
    api.setState({ isOpen: !api.state.isOpen });
  }

  protected createPanelComponent(): React.ComponentType<{
    api: PluginAPI;
    onClose: () => void;
  }> {
    return BrightnessPluginPanel;
  }

  // Lifecycle hook implementations
  private async onContextChangeHook(context: PluginContext): Promise<void> {
    // Additional context change handling if needed
    await this.handleContextChange(context);
  }

  private async onButtonClickHook(api: PluginAPI): Promise<void> {
    // Additional button click handling if needed
    await this.handleButtonClick(api);
  }

  // Public API methods specific to this plugin
  public hasChannelSupport(): boolean {
    const context = this.getContext();
    return !!(
      context.rawSlideMetadata?.metadata?.channels &&
      Object.keys(context.rawSlideMetadata.metadata.channels).length > 0
    );
  }

  public getAvailableChannels(): string[] {
    const context = this.getContext();
    const channels = context.rawSlideMetadata?.metadata?.channels;
    return channels ? Object.keys(channels) : [];
  }

  public applyStyleOptions(options: SlideStyleOptions): void {
    const context = this.getContext();
    if (context.slideLayer && context.rawSlideMetadata?.metadata?.channels) {
      const styleOptions = {
        ...options,
        channelMetadata: context.rawSlideMetadata.metadata.channels,
      };
      const style = generateSlideStyle(styleOptions);
      context.slideLayer.setStyle(style);
    }
  }
}

// Create and export plugin instance
export const brightnessControlPluginV2 = new BrightnessControlPluginV2();
