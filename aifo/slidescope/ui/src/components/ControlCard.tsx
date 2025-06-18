// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { ReactNode } from "react";
import { X } from "lucide-react";

interface ControlCardProps {
  title: string;
  icon?: ReactNode;
  onClose: () => void;
  position?: "left" | "right";
  width?: string;
  children: ReactNode;
  className?: string;
  id?: string;
}

export default function ControlCard({
  title,
  icon,
  onClose,
  position = "left",
  width = "w-72",
  children,
  className = "",
  id,
}: ControlCardProps) {
  const positionClasses =
    position === "left" ? "left-4 top-16" : "right-4 top-16";

  return (
    <div
      id={id}
      className={`control-card card-shadow fixed ${positionClasses} ${width} z-40 bg-white ${className}`}
    >
      <div className="bg-gradient-to-r from-indigo-600 to-indigo-500 p-3 text-white flex justify-between items-center">
        <div className="flex items-center">
          {icon}
          <h3 className="font-medium">{title}</h3>
        </div>
        <button
          className="text-white/80 hover:text-white focus:outline-none"
          onClick={onClose}
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      <div className="max-h-[calc(100vh-12rem)] overflow-y-auto p-4 bg-white">
        {children}
      </div>
    </div>
  );
}
