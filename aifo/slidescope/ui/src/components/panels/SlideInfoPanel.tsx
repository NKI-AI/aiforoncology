// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect } from "react";
import ControlCard from "../ControlCard";
import {
  Image,
  Aperture,
  Microscope,
  Factory,
  Grid,
  Monitor,
  HardDrive,
  Info,
  X,
} from "lucide-react";

interface SlideInfoPanelProps {
  slideMetadata: any;
  onClose: () => void;
}

export default function SlideInfoPanel({
  slideMetadata,
  onClose,
}: SlideInfoPanelProps) {
  const {
    slideId: slideId = "Loading...",
    slideMpp: slideMpp = "Loading...",
    objective = "Loading...",
    physicalDimensions = "Loading...",
    vendor = "Loading...",
    resolution = "Loading...",
    digitalSize = "Loading...",
  } = slideMetadata || {};

  const items = [
    { label: "Slide ID", value: slideId, icon: Image },
    { label: "Microns Per Pixel", value: slideMpp, icon: Aperture },
    { label: "Objective", value: objective, icon: Microscope },
    { label: "Physical Dimensions", value: physicalDimensions, icon: Grid },
    { label: "Vendor", value: vendor, icon: Factory },
    { label: "Resolution", value: resolution, icon: Monitor },
    { label: "Digital Size", value: digitalSize, icon: HardDrive },
  ];

  // Note: Since this panel has a unique animation, we're keeping its outer structure
  // but using elements from the ControlCard for consistency
  const [isVisible, setIsVisible] = useState(false);
  useEffect(() => {
    const timer = setTimeout(() => setIsVisible(true), 50);
    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(onClose, 200);
  };

  return (
    <div
      className="fixed top-16 bottom-0 left-0 z-40 w-72 bg-white shadow-lg transform transition-transform duration-200 ease-out rounded-tr-lg"
      style={{
        transform: isVisible ? "translateX(0)" : "translateX(-100%)",
      }}
    >
      <div className="bg-gradient-to-r from-indigo-600 to-indigo-500 p-3 text-white flex justify-between items-center">
        <div className="flex items-center">
          <Info className="h-5 w-5 mr-2" />
          <h3 className="font-medium">Slide Information</h3>
        </div>
        <button
          className="text-white/80 hover:text-white focus:outline-none"
          onClick={handleClose}
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      <div className="p-5 space-y-6 overflow-y-auto h-[calc(100%-56px)]">
        {items.map(({ label, value, icon: Icon }) => (
          <div key={label} className="flex items-start">
            <div className="bg-indigo-50 p-2.5 rounded-lg mr-4">
              <Icon className="w-5 h-5 text-indigo-600" />
            </div>
            <div className="flex-1">
              <p className="text-xs text-gray-500 font-medium mb-1">{label}</p>
              <p className="font-medium text-sm text-gray-800 break-words">
                {value}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
