// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

export interface PluginSettings {
  viewerSettings: boolean;
  brightnessControl: boolean;
  annotationControl: boolean;
  maskControl: boolean;
  regionControl: boolean;
}

export interface UserStudySettingsModulesPageProps {
  pluginSettings: PluginSettings;
  setPluginSettings: (settings: PluginSettings) => void;
}

export default function UserStudySettingsModulesPage({
  pluginSettings,
  setPluginSettings,
}: UserStudySettingsModulesPageProps) {
  const uid = React.useId();

  const handlePluginToggle = (
    pluginKey: keyof PluginSettings,
    checked: boolean
  ) => {
    setPluginSettings({ ...pluginSettings, [pluginKey]: checked });
  };

  return (
    <Card className="border bg-card shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">
          Viewer modules
        </CardTitle>
        <p className="text-sm text-muted-foreground">
          Control which tools and features are available in the slide viewer for
          this study.
        </p>
      </CardHeader>

      <CardContent className="p-0">
        <div className="divide-y">
          {/* Core Modules Section */}
          <section className="px-6 py-5">
            <div className="mb-4">
              <h3 className="text-sm font-medium text-foreground">
                Core modules
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                Essential viewer functionality and controls.
              </p>
            </div>

            <div className="space-y-4">
              {/* Viewer Settings */}
              <div className="flex items-center justify-between py-2">
                <div className="flex-1">
                  <div className="text-sm font-medium">Viewer Settings</div>
                  <div className="text-xs text-muted-foreground">
                    Pan sensitivity, zoom controls, quality settings
                  </div>
                </div>
                <Switch
                  id={`${uid}-viewer-settings`}
                  checked={pluginSettings.viewerSettings}
                  onCheckedChange={(checked) =>
                    handlePluginToggle("viewerSettings", checked)
                  }
                />
              </div>

              {/* Brightness Control */}
              <div className="flex items-center justify-between py-2">
                <div className="flex-1">
                  <div className="text-sm font-medium">
                    Brightness & Contrast
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Image brightness and contrast adjustments
                  </div>
                </div>
                <Switch
                  id={`${uid}-brightness-control`}
                  checked={pluginSettings.brightnessControl}
                  onCheckedChange={(checked) =>
                    handlePluginToggle("brightnessControl", checked)
                  }
                />
              </div>
            </div>
          </section>

          {/* Annotation Modules Section */}
          <section className="px-6 py-5">
            <div className="mb-4">
              <h3 className="text-sm font-medium text-foreground">
                Annotation modules
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                Tools for creating, viewing, and managing annotations.
              </p>
            </div>

            <div className="space-y-4">
              {/* Manual Annotation Control */}
              <div className="flex items-center justify-between py-2">
                <div className="flex-1">
                  <div className="text-sm font-medium">Manual Annotations</div>
                  <div className="text-xs text-muted-foreground">
                    Draw points, boxes, and polygons manually
                  </div>
                </div>
                <Switch
                  id={`${uid}-annotation-control`}
                  checked={pluginSettings.annotationControl}
                  onCheckedChange={(checked) =>
                    handlePluginToggle("annotationControl", checked)
                  }
                />
              </div>

              {/* Mask Control */}
              <div className="flex items-center justify-between py-2">
                <div className="flex-1">
                  <div className="text-sm font-medium">Mask Visualization</div>
                  <div className="text-xs text-muted-foreground">
                    View and control mask overlays from models
                  </div>
                </div>
                <Switch
                  id={`${uid}-mask-control`}
                  checked={pluginSettings.maskControl}
                  onCheckedChange={(checked) =>
                    handlePluginToggle("maskControl", checked)
                  }
                />
              </div>
            </div>
          </section>

          {/* Advanced Modules Section */}
          <section className="px-6 py-5">
            <div className="mb-4">
              <h3 className="text-sm font-medium text-foreground">
                Advanced modules
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                Specialized tools for specific use cases.
              </p>
            </div>

            <div className="space-y-4">
              {/* Region Control */}
              <div className="flex items-center justify-between py-2">
                <div className="flex-1">
                  <div className="text-sm font-medium">Region Annotations</div>
                  <div className="text-xs text-muted-foreground">
                    Define regions for multi-patient samples or tissue areas
                  </div>
                </div>
                <Switch
                  id={`${uid}-region-control`}
                  checked={pluginSettings.regionControl}
                  onCheckedChange={(checked) =>
                    handlePluginToggle("regionControl", checked)
                  }
                />
              </div>
            </div>
          </section>

          {/* Info Section */}
          <section className="px-6 py-5 bg-muted/20">
            <div className="rounded-lg border bg-background/50 p-4">
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0 mt-0.5">
                  <svg
                    className="h-5 w-5 text-blue-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth={1.5}
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"
                    />
                  </svg>
                </div>
                <div className="flex-1 space-y-2">
                  <h4 className="text-sm font-medium text-foreground">
                    Module visibility
                  </h4>
                  <div className="text-xs text-muted-foreground space-y-1">
                    <p>• Disabled modules won't appear in the viewer toolbar</p>
                    <p>
                      • Users will only see the tools relevant to this study
                    </p>
                    <p>• Changes apply immediately to new viewer sessions</p>
                    <p>
                      • Consider your study's workflow when selecting modules
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </CardContent>
    </Card>
  );
}
