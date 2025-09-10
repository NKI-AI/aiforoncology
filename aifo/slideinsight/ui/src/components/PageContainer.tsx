// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface PageContainerProps {
  maxWidth?: string;
  className?: string;
  children: React.ReactNode;
}

function PageContainer({
  maxWidth = "max-w-7xl",
  className = "",
  children,
}: PageContainerProps) {
  return (
    <div className={`container mx-auto ${maxWidth} ${className}`}>
      {children}
    </div>
  );
}

export default PageContainer;
