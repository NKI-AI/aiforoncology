// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { SunIcon, MoonIcon } from "@/components/icons";
import { useDarkMode } from "@/hooks/useDarkMode";

interface DarkModeToggleProps {
  className?: string;
  size?: "sm" | "md" | "lg";
  variant?: "navbar" | "standalone";
}

export function DarkModeToggle({
  className = "",
  size = "md",
  variant = "standalone",
}: DarkModeToggleProps) {
  const { isDark, toggle } = useDarkMode();

  const sizeClasses = {
    sm: "h-6 w-6",
    md: "h-8 w-8",
    lg: "h-10 w-10",
  };

  const iconSizeClasses = {
    sm: "h-3 w-3",
    md: "h-4 w-4",
    lg: "h-5 w-5",
  };

  const variantClasses = {
    navbar:
      "text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:ring-white/50",
    standalone:
      "text-muted-600 dark:text-muted-400 hover:text-muted-800 dark:hover:text-muted-200 hover:bg-muted-100 dark:hover:bg-muted-800 focus:ring-indigo-500/50",
  };

  return (
    <button
      type="button"
      onClick={toggle}
      title={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={`${sizeClasses[size]} inline-flex items-center justify-center rounded-md transition-colors duration-200 focus:outline-none focus:ring-2 ${variantClasses[variant]} ${className}`}
    >
      {isDark ? (
        <SunIcon className={iconSizeClasses[size]} />
      ) : (
        <MoonIcon className={iconSizeClasses[size]} />
      )}
    </button>
  );
}
