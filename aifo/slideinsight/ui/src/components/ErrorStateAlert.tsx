// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface ErrorStateAlertProps {
  error: string | Error | null;
  title?: string;
  onRetry?: () => void;
  retryText?: string;
  isRetrying?: boolean;
  variant?: "simple" | "detailed" | "inline";
  className?: string;
}

export function ErrorStateAlert({
  error,
  title = "Unable to load data",
  onRetry,
  retryText = "Try again",
  isRetrying = false,
  variant = "simple",
  className = "",
}: ErrorStateAlertProps) {
  if (!error) return null;

  const errorMessage = error instanceof Error ? error.message : String(error);

  if (variant === "inline") {
    return (
      <div
        className={`bg-red-50 border-l-4 border-red-500 p-4 mb-6 rounded-r-lg ${className}`}
      >
        <div className="flex">
          <div className="flex-shrink-0">
            <div className="h-5 w-5 text-red-400">⚠</div>
          </div>
          <div className="ml-3">
            <p className="text-sm text-red-700">
              <strong>{title}:</strong> {errorMessage}
            </p>
            {onRetry && (
              <button
                onClick={onRetry}
                disabled={isRetrying}
                className="mt-2 text-sm text-red-600 hover:text-red-800 underline disabled:text-red-400 disabled:no-underline"
              >
                {isRetrying ? "Retrying..." : retryText}
              </button>
            )}
          </div>
        </div>
      </div>
    );
  }

  if (variant === "detailed") {
    return (
      <div
        className={`bg-red-50 border border-red-200 rounded-md p-4 m-4 ${className}`}
      >
        <div className="flex">
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">{title}</h3>
            <div className="mt-2 text-sm text-red-700">{errorMessage}</div>
            {onRetry && (
              <button
                onClick={onRetry}
                disabled={isRetrying}
                className="mt-3 text-sm bg-red-100 hover:bg-red-200 text-red-700 px-3 py-1 rounded disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isRetrying ? "Retrying..." : retryText}
              </button>
            )}
          </div>
        </div>
      </div>
    );
  }

  // Default "simple" variant
  return (
    <div
      className={`bg-red-50 border border-red-200 rounded-md p-4 ${className}`}
    >
      <div className="flex">
        <div className="ml-3">
          <h3 className="text-sm font-medium text-red-800">{title}</h3>
          <div className="mt-2 text-sm text-red-700">{errorMessage}</div>
          {onRetry && (
            <button
              onClick={onRetry}
              disabled={isRetrying}
              className="mt-2 text-sm text-red-600 hover:text-red-800 underline disabled:text-red-400 disabled:no-underline"
            >
              {isRetrying ? "Retrying..." : retryText}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default ErrorStateAlert;
