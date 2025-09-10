// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface SlideInsightIconProps {
  className?: string;
}

const SlideInsightIcon: React.FC<SlideInsightIconProps> = ({
  className = "h-8 w-8 text-indigo-300",
}) => {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle
        cx="12"
        cy="12"
        r="9"
        stroke="currentColor"
        strokeWidth="2"
        fill="none"
      />
      <circle cx="12" cy="12" r="5" fill="currentColor" />
    </svg>
  );
};

export default SlideInsightIcon;
