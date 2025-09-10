// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface ModalSectionProps {
  title?: string;
  children: React.ReactNode;
  className?: string;
  background?: boolean; // Whether to show background styling
}

const ModalSection: React.FC<ModalSectionProps> = ({
  title,
  children,
  className = "",
  background = false,
}) => {
  const sectionClasses = background ? "bg-gray-50 rounded-lg p-4" : "";

  return (
    <div className={`space-y-4 ${className}`}>
      {title && <h3 className="text-lg font-medium text-muted-900">{title}</h3>}
      <div className={sectionClasses}>{children}</div>
    </div>
  );
};

export default ModalSection;
