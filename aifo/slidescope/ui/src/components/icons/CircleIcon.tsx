// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";

interface CircleIconProps {
  className?: string;
}

const CircleIcon: React.FC<CircleIconProps> = ({
  className = "-ml-0.5 mr-1.5 h-2 w-2",
}) => {
  return (
    <svg className={className} fill="currentColor" viewBox="0 0 8 8">
      <circle cx="4" cy="4" r="3" />
    </svg>
  );
};

export default CircleIcon;
