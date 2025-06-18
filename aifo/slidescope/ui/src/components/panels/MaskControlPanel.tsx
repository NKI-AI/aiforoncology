// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect } from "react";
import ControlCard from "../ControlCard";
import { SketchPicker } from "react-color";
import { MaskColor } from "../../types";
import { useMaskContext } from "../../contexts/MaskContext";

interface MaskControlPanelProps {
  onClose: () => void;
}

export default function MaskControlPanel({ onClose }: MaskControlPanelProps) {
  const {
    maskOpacity,
    showMask,
    maskColors,
    setMaskOpacity,
    setShowMask,
    handleMaskColorChange,
  } = useMaskContext();

  const [openColorPicker, setOpenColorPicker] = useState<number | null>(null);
  const [colors, setColors] = useState(maskColors);
  const [layerVisibility, setLayerVisibility] = useState<boolean[]>(
    maskColors.map(() => true)
  );

  // Update local colors when context colors change
  useEffect(() => {
    setColors(maskColors);
  }, [maskColors]);

  const handleOpacityChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newOpacity = parseFloat(e.target.value);
    setMaskOpacity(newOpacity);
  };

  const handleToggleVisibility = (e: React.ChangeEvent<HTMLInputElement>) => {
    setShowMask(e.target.checked);
  };

  const handleColorClick = (index: number) => {
    setOpenColorPicker(openColorPicker === index ? null : index);
  };

  const handleColorClose = () => {
    setOpenColorPicker(null);
  };

  const handleColorChange = (index: number, color: any) => {
    console.log(`[DEBUG] Color picker change for index ${index}:`, color);
    const newColors = [...colors];
    newColors[index] = color.rgb;
    setColors(newColors);

    // Make sure we're passing a proper color object with r,g,b,a properties
    const rgbColor = {
      r: color.rgb.r,
      g: color.rgb.g,
      b: color.rgb.b,
      a: layerVisibility[index] ? 1.0 : 0.0, // Set alpha based on layer visibility
    };
    console.log(
      `[DEBUG] Sending color change to parent for label ${index + 1}:`,
      rgbColor
    );
    handleMaskColorChange(index + 1, rgbColor); // +1 because mask labels start at 1
  };

  const handleLayerToggle = (index: number) => {
    const newLayerVisibility = [...layerVisibility];
    newLayerVisibility[index] = !newLayerVisibility[index];
    setLayerVisibility(newLayerVisibility);

    const color = colors[index];
    const rgbColor = {
      r: color.r,
      g: color.g,
      b: color.b,
      a: newLayerVisibility[index] ? 1.0 : 0.0, // Set alpha to 0 when layer is hidden
    };
    console.log(
      `[DEBUG] Toggling visibility for label ${index + 1}:`,
      rgbColor
    );
    handleMaskColorChange(index + 1, rgbColor);
  };

  const icon = (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="h-5 w-5 mr-2"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"
      />
    </svg>
  );

  return (
    <ControlCard
      id="mask-card"
      title="Mask Controls"
      icon={icon}
      onClose={onClose}
      position="left"
    >
      <div className="space-y-4">
        <div className="p-3 bg-indigo-50 rounded-md border border-indigo-100">
          <label
            className="flex items-center mb-2 cursor-pointer"
            htmlFor="mask-toggle"
          >
            <input
              type="checkbox"
              id="mask-toggle"
              checked={showMask}
              onChange={handleToggleVisibility}
              className="form-checkbox h-4 w-4 text-indigo-600 rounded"
            />
            <span className="ml-2 text-indigo-800 font-medium">Show Mask</span>
          </label>

          <div className="pt-2">
            <label
              htmlFor="mask-opacity"
              className="block mb-1 text-indigo-700"
            >
              <span>
                Opacity:
                <span id="opacity-value" className="font-medium">
                  {" "}
                  {maskOpacity.toFixed(1)}
                </span>
              </span>
            </label>
            <input
              type="range"
              id="mask-opacity"
              min="0"
              max="1"
              step="0.1"
              value={maskOpacity}
              onChange={handleOpacityChange}
              className="w-full h-2 bg-indigo-200 rounded-lg appearance-none cursor-pointer"
            />
          </div>
        </div>

        <div className="p-3 bg-indigo-50 rounded-md border border-indigo-100">
          <h3 className="font-medium text-indigo-800 mb-3">Mask Colors</h3>
          <div className="space-y-4">
            {colors.map((color, index) => (
              <div key={index} className="relative">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <label className="relative inline-flex items-center cursor-pointer mr-3">
                      <input
                        type="checkbox"
                        className="sr-only peer"
                        checked={layerVisibility[index]}
                        onChange={() => handleLayerToggle(index)}
                      />
                      <div className="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-indigo-600"></div>
                    </label>
                    <span className="text-sm text-indigo-700">
                      Label {index + 1}
                    </span>
                  </div>
                  <div
                    className="w-8 h-8 rounded cursor-pointer border border-gray-300"
                    style={{
                      backgroundColor: `rgba(${color.r}, ${color.g}, ${
                        color.b
                      }, ${layerVisibility[index] ? 1.0 : 0.3})`,
                      opacity: layerVisibility[index] ? 1.0 : 0.5,
                    }}
                    onClick={() => handleColorClick(index)}
                  />
                </div>

                {openColorPicker === index && (
                  <div className="absolute right-0 mt-1 z-10">
                    <div className="fixed inset-0" onClick={handleColorClose} />
                    <SketchPicker
                      color={color}
                      onChange={(newColor) =>
                        handleColorChange(index, newColor)
                      }
                      disableAlpha={true}
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </ControlCard>
  );
}
