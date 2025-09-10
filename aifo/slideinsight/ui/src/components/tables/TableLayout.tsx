// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import * as React from "react";

interface TableLayoutProps {
  header?: React.ReactNode;
  controls?: React.ReactNode;
  body: React.ReactNode;
  footer?: React.ReactNode;
  loading?: boolean;
  className?: string;
}

export function TableLayout({
  header,
  controls,
  body,
  footer,
  loading = false,
  className = "",
}: TableLayoutProps) {
  if (loading) {
    return (
      <div className={`w-full space-y-4 ${className}`}>
        {header}
        <div className="flex justify-center items-center p-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  return (
    <div className={`w-full space-y-4 ${className}`}>
      {header}
      {controls}
      {body}
      {footer}
    </div>
  );
}
