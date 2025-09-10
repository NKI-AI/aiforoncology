// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface MoonIconProps {
  className?: string;
}

const MoonIcon: React.FC<MoonIconProps> = ({ className = "h-5 w-5" }) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      className={className}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M21.752 15.002A9.718 9.718 0 0112 21.75c-5.385 0-9.75-4.365-9.75-9.75 0-4.112 2.565-7.62 6.174-9.02a.75.75 0 01.98.97A7.5 7.5 0 0020.25 14.27a.75.75 0 011.502.732z"
      />
    </svg>
  );
};

export default MoonIcon;
