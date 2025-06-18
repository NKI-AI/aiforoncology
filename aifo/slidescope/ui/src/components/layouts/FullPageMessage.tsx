// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { ReactNode } from "react";

interface FullPageMessageProps {
  icon: ReactNode;
  title: string;
  message: string | ReactNode;
  actions?: ReactNode;
  iconBgColor?: string;
}

export default function FullPageMessage({
  icon,
  title,
  message,
  actions,
  iconBgColor = "bg-indigo-100",
}: FullPageMessageProps) {
  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-5">
      <div className="bg-white p-8 rounded-lg shadow-md max-w-md w-full text-center">
        <div
          className={`inline-flex items-center justify-center w-16 h-16 rounded-full ${iconBgColor} mb-5`}
        >
          {icon}
        </div>
        <h1 className="text-3xl font-bold text-gray-800 mb-3">{title}</h1>
        <div className="text-gray-600 mb-6">{message}</div>
        {actions && <div className="flex flex-col space-y-3">{actions}</div>}
      </div>
    </div>
  );
}
