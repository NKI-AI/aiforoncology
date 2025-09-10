// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface ModalHeaderProps {
  title: string;
  description?: string;
  className?: string;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({
  title,
  description,
  className = "",
}) => {
  return (
    <div className={`mb-6 ${className}`}>
      <h2 className="text-xl font-semibold text-muted-900 mb-2">{title}</h2>
      {description && <p className="text-sm text-muted-600">{description}</p>}
    </div>
  );
};

export default ModalHeader;
