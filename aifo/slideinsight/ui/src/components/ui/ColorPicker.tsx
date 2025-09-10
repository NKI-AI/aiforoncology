// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { Block, Sketch, ColorResult } from "@uiw/react-color";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { PaintBrushIcon } from "@heroicons/react/24/outline";
import { useDarkMode } from "@/hooks/useDarkMode";

// Default color palette with commonly used colors
const DEFAULT_COLORS = [
  "#FF6B6B",
  "#4ECDC4",
  "#45B7D1",
  "#96CEB4",
  "#FFEAA7",
  "#DDA0DD",
  "#98D8C8",
  "#F7DC6F",
  "#BB8FCE",
  "#85C1E9",
  "#F8C471",
  "#82E0AA",
  "#F1948A",
  "#85C1E9",
  "#D7BDE2",
  "#A3E4D7",
  "#F9E79F",
  "#D5A6BD",
  "#AED6F1",
  "#A9DFBF",
  "#FAD7A0",
  "#E8DAEF",
  "#D6EAF8",
  "#D1F2EB",
  "#FCF3CF",
  "#FADBD8",
  "#E6E6FA",
  "#F0F8FF",
  "#F5F5DC",
  "#FFF8DC",
  "#FF0000",
  "#00FF00",
  "#0000FF",
  "#FFFF00",
  "#FF00FF",
  "#00FFFF",
  "#800000",
  "#008000",
  "#000080",
  "#808000",
  "#800080",
  "#008080",
  "#C0C0C0",
  "#808080",
  "#000000",
];

export interface ColorPickerProps {
  color: string;
  onChange: (result: ColorResult) => void;
  onChangeComplete?: (result: ColorResult) => void;
  colors?: string[];
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
  showAdvanced?: boolean;
  side?: "top" | "right" | "bottom" | "left";
  align?: "start" | "center" | "end";
}

export default function ColorPicker({
  color,
  onChange,
  onChangeComplete,
  colors = DEFAULT_COLORS,
  open,
  onOpenChange,
  children,
  showAdvanced = true,
  side = "bottom",
  align = "center",
}: ColorPickerProps) {
  const [showSketch, setShowSketch] = React.useState(false);
  const [internalOpen, setInternalOpen] = React.useState(false);
  const { isDark } = useDarkMode();

  // Support controlled and uncontrolled modes for the popover
  const isOpen = open ?? internalOpen;
  const handleOpenChange = React.useCallback(
    (newOpen: boolean) => {
      if (open === undefined) {
        setInternalOpen(newOpen);
      }
      onOpenChange?.(newOpen);
    },
    [open, onOpenChange]
  );

  const handleColorChange = (result: ColorResult) => {
    onChange(result);
    onChangeComplete?.(result);
  };

  const handleAdvancedClick = () => {
    setShowSketch(true);
  };

  const handleSketchChange = (result: ColorResult) => {
    onChange(result);
    onChangeComplete?.(result);
  };

  const handleSketchComplete = () => {
    setShowSketch(false);
  };

  return (
    <Popover open={isOpen} onOpenChange={handleOpenChange}>
      <PopoverTrigger>
        {children || (
          <Button variant="outline" size="sm" aria-label="Pick color">
            <PaintBrushIcon className="mr-2 h-4 w-4" />
            <div
              className="h-4 w-4 rounded border"
              style={{ backgroundColor: color }}
              aria-label="Color preview"
            />
          </Button>
        )}
      </PopoverTrigger>
      <PopoverContent
        align={align}
        side={side}
        withArrow
        className="w-auto p-1"
        data-color-mode={isDark ? "dark" : "light"}
      >
        {showSketch ? (
          <div className="space-y-1">
            <Sketch color={color} onChange={handleSketchChange} disableAlpha />
            <Button
              variant="outline"
              size="sm"
              onClick={handleSketchComplete}
              className="w-full"
            >
              Back to palette
            </Button>
          </div>
        ) : (
          <div className="space-y-1">
            <Block color={color} colors={colors} onChange={handleColorChange} />
            {showAdvanced && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleAdvancedClick}
                className="w-full"
              >
                <PaintBrushIcon className="mr-2 h-4 w-4" />
                Advanced colors
              </Button>
            )}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
