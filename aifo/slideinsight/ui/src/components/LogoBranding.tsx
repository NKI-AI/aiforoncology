// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import { SlideInsightIcon } from "./icons";

interface LogoBrandingProps {
  size?: "small" | "large";
  linkTo?: string;
  className?: string;
  centered?: boolean;
}

export function LogoBranding({
  size = "large",
  linkTo,
  className = "",
  centered = true,
}: LogoBrandingProps) {
  const isSmall = size === "small";

  const content = (
    <>
      <div className="bg-indigo-500/10 p-2 rounded-xl">
        <SlideInsightIcon
          className={`text-indigo-400 ${isSmall ? "h-10 w-10" : "h-12 w-12"}`}
        />
      </div>
      <div className="text-left">
        <h1
          className={`font-bold text-muted-100 ${
            isSmall ? "text-3xl" : "text-4xl"
          }`}
        >
          SlideInsight
        </h1>
        <div
          className={`text-indigo-400/70 font-small ${
            isSmall ? "text-xs" : "text-sm"
          }`}
          style={{ fontSize: "0.73rem" }}
        >
          Computational Pathology Platform
        </div>
      </div>
    </>
  );

  const baseClasses = `flex items-center space-x-3 ${
    centered ? "justify-center" : ""
  } ${className}`;

  if (linkTo) {
    return (
      <Link
        to={linkTo}
        className={`${baseClasses} hover:opacity-80 transition-opacity duration-200`}
      >
        {content}
      </Link>
    );
  }

  return <div className={baseClasses}>{content}</div>;
}
