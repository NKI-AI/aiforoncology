// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface CpuChipIconProps {
  className?: string;
}

const CpuChipIcon: React.FC<CpuChipIconProps> = ({ className = "h-5 w-5" }) => {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M8.25 3v1.5M16.5 3v1.5M8.25 22.5V21M16.5 22.5V21M3 8.25h1.5M22.5 8.25H21M3 16.5h1.5M22.5 16.5H21M6.75 6.75h10.5v10.5H6.75V6.75Z"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 9v6m-3-3h6"
      />
    </svg>
  );
};

export default CpuChipIcon;
