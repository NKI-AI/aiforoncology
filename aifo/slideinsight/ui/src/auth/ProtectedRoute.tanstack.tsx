// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useEffect } from "react";
import { useRouter } from "@tanstack/react-router";
import { useAuth } from "./auth";

export default function ProtectedRoute({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { isLoggedIn, isLoading, user } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
        <div className="bg-background rounded-lg shadow p-6 max-w-md w-full">
          <div className="flex justify-center mb-4">
            <svg
              className="animate-spin h-8 w-8 text-indigo-600"
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
          <p className="text-center text-muted-700">
            Checking authentication...
          </p>
        </div>
      </div>
    );
  }

  if (!isLoggedIn) {
    // Get current path for redirect after login
    const currentPath = router.state.location.pathname;

    // Navigate to login with the current path and message in search params
    // Use setTimeout to avoid navigation conflicts during render
    setTimeout(() => {
      router.navigate({
        to: "/login",
        search: {
          from: currentPath,
          message: "You need to login to view this website",
        },
        replace: true,
      });
    }, 0);

    // Return loading state while navigating
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  // If authenticated, render the children
  return <>{children}</>;
}
