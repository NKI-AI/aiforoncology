// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import ControlCard from "../ControlCard";
import { HelpCircle } from "lucide-react";

interface HelpPanelProps {
  onClose: () => void;
}

export default function HelpPanel({ onClose }: HelpPanelProps) {
  return (
    <ControlCard
      id="help-card"
      title="Keyboard Shortcuts"
      icon={<HelpCircle className="h-5 w-5 mr-2" />}
      onClose={onClose}
      position="right"
      width="w-80"
    >
      <div className="space-y-4">
        <div>
          <h4 className="font-medium text-indigo-700 mb-2">Panel Controls</h4>
          <div className="grid grid-cols-2 gap-2">
            <div className="text-gray-700 font-medium">M</div>
            <div className="text-gray-600">Toggle Mask panel</div>

            <div className="text-gray-700 font-medium">S</div>
            <div className="text-gray-600">Toggle Slide Info</div>

            <div className="text-gray-700 font-medium">H</div>
            <div className="text-gray-600">Toggle Help</div>

            <div className="text-gray-700 font-medium">ESC</div>
            <div className="text-gray-600">Close all panels</div>
          </div>
        </div>

        <div id="mask-controls-help">
          <h4 className="font-medium text-indigo-700 mb-2">Mask Controls</h4>
          <div className="grid grid-cols-2 gap-2">
            <div className="text-gray-700 font-medium">+</div>
            <div className="text-gray-600">Increase mask opacity</div>

            <div className="text-gray-700 font-medium">-</div>
            <div className="text-gray-600">Decrease mask opacity</div>

            <div className="text-gray-700 font-medium">T</div>
            <div className="text-gray-600">Toggle mask visibility</div>
          </div>
        </div>

        <div>
          <h4 className="font-medium text-indigo-700 mb-2">Navigation</h4>
          <div className="grid grid-cols-2 gap-2">
            <div className="text-gray-700 font-medium">Mouse Wheel</div>
            <div className="text-gray-600">Zoom in/out</div>

            <div className="text-gray-700 font-medium">Click & Drag</div>
            <div className="text-gray-600">Pan around slide</div>

            <div className="text-gray-700 font-medium">Double Click</div>
            <div className="text-gray-600">Reset view</div>

            <div className="text-gray-700 font-medium">C</div>
            <div className="text-gray-600">
              Toggle coordinates and crosshair
            </div>
          </div>
        </div>
      </div>
    </ControlCard>
  );
}
