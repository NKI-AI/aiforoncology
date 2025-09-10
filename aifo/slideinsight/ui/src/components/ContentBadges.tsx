// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { FolderOpen, Image } from "lucide-react";

interface ContentBadgesProps {
  caseCount?: number;
  slideCount?: number;
  showCases?: boolean;
  showSlides?: boolean;
  loading?: boolean;
}

const ContentBadges: React.FC<ContentBadgesProps> = ({
  caseCount = 0,
  slideCount = 0,
  showCases = true,
  showSlides = true,
  loading = false,
}) => {
  if (loading) {
    return (
      <div className="flex flex-col gap-1">
        {showCases && (
          <div className="inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap bg-muted text-muted-foreground border gap-1 animate-pulse">
            <FolderOpen className="h-3 w-3 text-muted-foreground" />
            <span className="font-semibold">--</span>
            <span>cases</span>
          </div>
        )}
        {showSlides && (
          <div className="inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap bg-muted text-muted-foreground border gap-1 animate-pulse">
            <Image className="h-3 w-3 text-muted-foreground" />
            <span className="font-semibold">--</span>
            <span>slides</span>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-1">
      {/* Cases Counter */}
      {showCases && (
        <div className="inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap bg-blue-50 text-blue-700 border-blue-200 gap-1">
          <FolderOpen className="h-3 w-3 text-blue-600" />
          <span className="font-semibold">{caseCount}</span>
          <span>case{caseCount !== 1 ? "s" : ""}</span>
        </div>
      )}

      {/* Slides Counter */}
      {showSlides && (
        <div className="inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap bg-purple-50 text-purple-700 border-purple-200 gap-1">
          <Image className="h-3 w-3 text-purple-600" />
          <span className="font-semibold">{slideCount}</span>
          <span>slide{slideCount !== 1 ? "s" : ""}</span>
        </div>
      )}
    </div>
  );
};

export default ContentBadges;
