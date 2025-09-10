// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { Suspense } from "react";
import { Link } from "@tanstack/react-router";
import { CompactAuthLayout, Alert } from "./FormComponents";

interface AuthPageLayoutProps {
  title: string;
  subtitle: string;
  children: React.ReactNode;

  // State management
  error?: string | null;
  success?: boolean;
  successMessage?: string | null;
  redirectMessage?: string | null;
  loading?: boolean;

  // Success actions
  successAction?: {
    text: string;
    to: string;
    search?: Record<string, string>;
  };

  // Additional content slots
  beforeForm?: React.ReactNode;
  afterForm?: React.ReactNode;

  // Custom alerts
  customAlerts?: React.ReactNode;
}

function LoadingFallback() {
  return (
    <CompactAuthLayout
      title="Loading..."
      subtitle="Please wait while we load the page"
    >
      <div className="flex justify-center items-center py-8">
        <svg
          className="animate-spin h-8 w-8 text-indigo-500"
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
      </div>
    </CompactAuthLayout>
  );
}

export function AuthPageLayout({
  title,
  subtitle,
  children,
  error,
  success,
  successMessage,
  redirectMessage,
  loading,
  successAction,
  beforeForm,
  afterForm,
  customAlerts,
}: AuthPageLayoutProps) {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <CompactAuthLayout title={title} subtitle={subtitle}>
        {/* Custom alerts - highest priority */}
        {customAlerts}

        {/* Redirect Message */}
        {redirectMessage && (
          <div className="mb-4">
            <Alert type="info" message={redirectMessage} />
          </div>
        )}

        {/* Error Alert */}
        {error && (
          <div className="mb-4">
            <Alert type="error" message={error} />
          </div>
        )}

        {/* Success Alert */}
        {success && successMessage && (
          <>
            <div className="mb-4">
              <Alert type="success" message={successMessage} />
            </div>
            {successAction && (
              <div className="text-center mb-4">
                <Link
                  to={successAction.to}
                  search={successAction.search}
                  className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors duration-200"
                >
                  {successAction.text} →
                </Link>
              </div>
            )}
          </>
        )}

        {/* Before form content */}
        {beforeForm}

        {/* Main content */}
        {children}

        {/* After form content */}
        {afterForm}
      </CompactAuthLayout>
    </Suspense>
  );
}
