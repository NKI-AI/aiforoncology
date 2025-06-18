// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { ButtonHTMLAttributes, ReactNode } from "react";
import { Link } from "react-router-dom";

type ButtonVariant = "primary" | "secondary";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  variant?: ButtonVariant;
  fullWidth?: boolean;
}

interface LinkButtonProps {
  children: ReactNode;
  to: string;
  variant?: ButtonVariant;
  fullWidth?: boolean;
  className?: string;
}

const getVariantClasses = (variant: ButtonVariant): string => {
  switch (variant) {
    case "primary":
      return "bg-indigo-600 text-white hover:bg-indigo-700";
    case "secondary":
      return "bg-white text-indigo-600 border border-indigo-600 hover:bg-indigo-50";
    default:
      return "bg-indigo-600 text-white hover:bg-indigo-700";
  }
};

export function Button({
  children,
  variant = "primary",
  fullWidth = false,
  className = "",
  ...props
}: ButtonProps) {
  const baseClasses = "px-4 py-2 rounded-md transition-colors";
  const widthClasses = fullWidth ? "w-full" : "";
  const variantClasses = getVariantClasses(variant);

  return (
    <button
      className={`${baseClasses} ${variantClasses} ${widthClasses} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function LinkButton({
  children,
  to,
  variant = "primary",
  fullWidth = false,
  className = "",
}: LinkButtonProps) {
  const baseClasses =
    "px-4 py-2 rounded-md transition-colors inline-block text-center";
  const widthClasses = fullWidth ? "w-full" : "";
  const variantClasses = getVariantClasses(variant);

  return (
    <Link
      to={to}
      className={`${baseClasses} ${variantClasses} ${widthClasses} ${className}`}
    >
      {children}
    </Link>
  );
}
