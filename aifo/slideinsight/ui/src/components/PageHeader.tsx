// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface PageHeaderProps {
  title?: string;
  subtitle?: string;
  className?: string;
  children?: React.ReactNode;
}

function PageHeader({
  title,
  subtitle,
  className = "",
  children,
}: PageHeaderProps) {
  if (!title && !subtitle && !children) {
    return null;
  }

  return (
    <div className={`mb-8 ${className}`}>
      {title && (
        <h1 className="text-3xl font-bold text-foreground mb-2">{title}</h1>
      )}
      {subtitle && <p className="text-muted-foreground">{subtitle}</p>}
      {children}
    </div>
  );
}

export default PageHeader;
