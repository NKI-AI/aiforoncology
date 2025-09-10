// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { EyeIcon, EyeSlashIcon } from "../../../components/icons";
import { LogoBranding } from "../../../components/LogoBranding";
import { DarkModeToggle } from "../../../components/DarkModeToggle";

interface FormFieldProps {
  id: string;
  name: string;
  type: string;
  label: string;
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
  autoComplete?: string;
  autoCapitalize?: string;
  spellCheck?: boolean;
  className?: string;
  error?: string;
}

export function FormField({
  id,
  name,
  type,
  label,
  placeholder,
  value,
  onChange,
  required = false,
  autoComplete,
  autoCapitalize = "none",
  spellCheck = false,
  className = "",
  error,
}: FormFieldProps) {
  const hasError = !!error;

  return (
    <div className="space-y-1">
      <label
        htmlFor={id}
        className="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        {label}
        {required && (
          <span className="text-red-500 dark:text-red-400 ml-1">*</span>
        )}
      </label>
      <input
        id={id}
        name={name}
        type={type}
        required={required}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={`block w-full px-3 py-2 bg-white dark:bg-gray-700/50 border rounded-lg 
                    focus:outline-none focus:ring-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 text-sm transition-all duration-200
                    autofill:bg-white autofill:text-gray-900 dark:autofill:bg-gray-700/50 dark:autofill:text-white
                    autofill:shadow-[inset_0_0_0px_1000px_white] dark:autofill:shadow-[inset_0_0_0px_1000px_rgb(55_65_81_/_0.5)]
                    ${
                      hasError
                        ? "border-red-500/70 focus:ring-red-500/50 focus:border-red-400"
                        : "border-gray-300 dark:border-gray-600/50 focus:ring-indigo-500/50 focus:border-indigo-400 hover:border-gray-400 dark:hover:border-gray-500/70 hover:bg-gray-50 dark:hover:bg-gray-700/70"
                    } ${className}`}
        autoComplete={autoComplete}
        autoCapitalize={autoCapitalize}
        spellCheck={spellCheck}
      />
      {hasError && (
        <p className="text-sm text-red-500 dark:text-red-400 flex items-center mt-1">
          <svg
            className="w-4 h-4 mr-1 flex-shrink-0"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fillRule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
              clipRule="evenodd"
            />
          </svg>
          {error}
        </p>
      )}
    </div>
  );
}

interface PasswordFieldProps {
  id: string;
  name: string;
  label: string;
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
  autoComplete?: string;
  className?: string;
  error?: string;
}

export function PasswordField({
  id,
  name,
  label,
  placeholder,
  value,
  onChange,
  required = false,
  autoComplete = "current-password",
  className = "",
  error,
}: PasswordFieldProps) {
  const [showPassword, setShowPassword] = useState(false);
  const hasError = !!error;

  const handleMouseDown = () => {
    setShowPassword(true);
  };

  const handleMouseUp = () => {
    setShowPassword(false);
  };

  const handleMouseLeave = () => {
    setShowPassword(false);
  };

  // Touch event handlers for mobile devices
  const handleTouchStart = (e: React.TouchEvent) => {
    e.preventDefault(); // Prevent mouse events from firing
    setShowPassword(true);
  };

  const handleTouchEnd = (e: React.TouchEvent) => {
    e.preventDefault(); // Prevent mouse events from firing
    setShowPassword(false);
  };

  // Fallback click handler for accessibility
  const handleClick = () => {
    // Toggle behavior for click (useful for keyboard navigation)
    setShowPassword(!showPassword);
  };

  return (
    <div className="space-y-1">
      <label
        htmlFor={id}
        className="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        {label}
        {required && (
          <span className="text-red-500 dark:text-red-400 ml-1">*</span>
        )}
      </label>
      <div className="relative">
        <input
          id={id}
          name={name}
          type={showPassword ? "text" : "password"}
          required={required}
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className={`block w-full px-3 py-2 pr-10 bg-white dark:bg-gray-700/50 border rounded-lg 
                        focus:outline-none focus:ring-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 text-sm transition-all duration-200
                        autofill:bg-white autofill:text-gray-900 dark:autofill:bg-gray-700/50 dark:autofill:text-white
                        autofill:shadow-[inset_0_0_0px_1000px_white] dark:autofill:shadow-[inset_0_0_0px_1000px_rgb(55_65_81_/_0.5)]
                        ${
                          hasError
                            ? "border-red-500/70 focus:ring-red-500/50 focus:border-red-400"
                            : "border-gray-300 dark:border-gray-600/50 focus:ring-indigo-500/50 focus:border-indigo-400 hover:border-gray-400 dark:hover:border-gray-500/70 hover:bg-gray-50 dark:hover:bg-gray-700/70"
                        } ${className}`}
          autoComplete={autoComplete}
        />
        <button
          type="button"
          onMouseDown={handleMouseDown}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseLeave}
          onTouchStart={handleTouchStart}
          onTouchEnd={handleTouchEnd}
          onClick={handleClick}
          className="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 dark:text-muted-400 hover:text-gray-700 dark:hover:text-muted-300 
                        focus:outline-none focus:text-indigo-600 dark:focus:text-indigo-400 transition-colors duration-200"
          aria-label={showPassword ? "Hide password" : "Show password"}
        >
          {showPassword ? (
            <EyeSlashIcon className="h-4 w-4" />
          ) : (
            <EyeIcon className="h-4 w-4" />
          )}
        </button>
      </div>
      {hasError && (
        <p className="text-sm text-red-500 dark:text-red-400 flex items-center mt-1">
          <svg
            className="w-4 h-4 mr-1 flex-shrink-0"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fillRule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
              clipRule="evenodd"
            />
          </svg>
          {error}
        </p>
      )}
    </div>
  );
}

interface SubmitButtonProps {
  loading: boolean;
  success: boolean;
  loadingText: string;
  successText: string;
  defaultText: string;
  className?: string;
}

export function SubmitButton({
  loading,
  success,
  loadingText,
  successText,
  defaultText,
  className = "",
}: SubmitButtonProps) {
  return (
    <button
      type="submit"
      disabled={loading || success}
      className={`w-full flex justify-center items-center py-2.5 px-4 border border-transparent 
                rounded-lg shadow-sm text-sm font-medium text-white transition-all duration-200
                ${
                  loading || success
                    ? "bg-indigo-500/50 cursor-not-allowed"
                    : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 hover:shadow-lg active:transform active:scale-98"
                } ${className}`}
    >
      {loading && (
        <svg
          className="animate-spin -ml-1 mr-3 h-4 w-4 text-white"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          ></circle>
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      )}
      {success && (
        <svg
          className="mr-2 h-4 w-4 text-green-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M5 13l4 4L19 7"
          />
        </svg>
      )}
      {loading ? loadingText : success ? successText : defaultText}
    </button>
  );
}

interface AlertProps {
  type: "error" | "success" | "info";
  message: string;
  className?: string;
}

export function Alert({ type, message, className = "" }: AlertProps) {
  const baseClasses =
    "flex items-start space-x-3 p-4 rounded-lg border text-sm";

  const typeClasses = {
    error:
      "bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800/30 text-red-800 dark:text-red-200",
    success:
      "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800/30 text-green-800 dark:text-green-200",
    info: "bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800/30 text-blue-800 dark:text-blue-200",
  };

  const iconClasses = {
    error: "text-red-500 dark:text-red-400",
    success: "text-green-500 dark:text-green-400",
    info: "text-blue-500 dark:text-blue-400",
  };

  const Icon = () => {
    if (type === "error") {
      return (
        <svg
          className={`h-5 w-5 ${iconClasses[type]} flex-shrink-0 mt-0.5`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    } else if (type === "success") {
      return (
        <svg
          className={`h-5 w-5 ${iconClasses[type]} flex-shrink-0 mt-0.5`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    } else {
      return (
        <svg
          className={`h-5 w-5 ${iconClasses[type]} flex-shrink-0 mt-0.5`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    }
  };

  return (
    <div className={`${baseClasses} ${typeClasses[type]} ${className}`}>
      <Icon />
      <div className="flex-1">{message}</div>
    </div>
  );
}

interface CompactAuthLayoutProps {
  title: string;
  subtitle: string;
  children: React.ReactNode;
  showDarkModeToggle?: boolean;
}

export function CompactAuthLayout({
  title,
  subtitle,
  children,
  showDarkModeToggle = true,
}: CompactAuthLayoutProps) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 flex flex-col items-center justify-center px-4 py-8 relative">
      {/* Dark Mode Toggle - Top right */}
      {showDarkModeToggle && (
        <div className="absolute top-4 right-4 z-10">
          <DarkModeToggle size="md" />
        </div>
      )}

      {/* Logo & Branding - Outside the card */}
      <LogoBranding className="mb-6" />

      {/* Compact Card */}
      <div className="w-full max-w-sm bg-white/95 dark:bg-gray-800/90 backdrop-blur-sm rounded-xl shadow-2xl border border-gray-200/50 dark:border-gray-700/50 overflow-hidden">
        {/* Card Header */}
        <div className="px-5 py-4 text-center border-b border-gray-200/30 dark:border-gray-700/30">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">
            {title}
          </h2>
          <p className="text-sm text-gray-600 dark:text-gray-300">{subtitle}</p>
        </div>

        {/* Card Content */}
        <div className="px-5 py-5 space-y-4">{children}</div>
      </div>

      {/* Terms and Privacy - Outside the card */}
      <div className="mt-5 text-center text-xs text-gray-600 dark:text-muted-400 max-w-sm">
        <p className="leading-relaxed">
          By continuing, you agree to our{" "}
          <Link
            to="/terms"
            className="text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 underline underline-offset-4 transition-colors duration-200"
          >
            Terms of Service
          </Link>{" "}
          and{" "}
          <Link
            to="/privacy"
            className="text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 underline underline-offset-4 transition-colors duration-200"
          >
            Privacy Policy
          </Link>
          .
        </p>
      </div>
    </div>
  );
}

interface PasswordRequirementsProps {
  className?: string;
}

export function PasswordRequirements({
  className = "",
}: PasswordRequirementsProps) {
  return (
    <div
      className={`text-xs text-gray-600 dark:text-muted-400 bg-gray-100 dark:bg-gray-700/30 p-3 rounded-lg border border-gray-200 dark:border-gray-600/30 ${className}`}
    >
      <div className="font-medium text-gray-800 dark:text-muted-300 mb-1">
        Password Requirements:
      </div>
      <ul className="space-y-1">
        <li>• At least 8 characters long</li>
        <li>• Contains uppercase and lowercase letters</li>
        <li>• Contains numbers and special characters</li>
      </ul>
    </div>
  );
}

// Re-export the new components
export { AuthPageLayout } from "./AuthPageLayout";
