// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface StudyCellProps {
  title: string;
  description?: string | null;
}

const StudyCell: React.FC<StudyCellProps> = ({ title, description }) => {
  return (
    <div className="py-2 min-h-[3.5rem] flex flex-col justify-center space-y-1 min-w-0 w-full max-w-xs">
      {/* Title with ellipsis for long text */}
      <div
        className="font-medium text-foreground leading-tight truncate"
        title={title}
      >
        {title}
      </div>

      {/* Description with exactly 2 lines and ellipsis */}
      {description && (
        <div
          className="text-xs text-muted-foreground leading-normal"
          title={description}
          style={{
            display: "-webkit-box",
            WebkitLineClamp: 2,
            WebkitBoxOrient: "vertical",
            overflow: "hidden",
            textOverflow: "ellipsis",
            wordBreak: "break-word",
          }}
        >
          {description}
        </div>
      )}
    </div>
  );
};

export default StudyCell;
