// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useAuth } from "../auth";

export default function RootRedirect() {
  const { isLoggedIn, isLoading, refresh } = useAuth();
  const navigate = useNavigate();

  const isDevelopment = process.env.NODE_ENV === "development";

  useEffect(() => {
    // Re-check authentication status to ensure we have the latest state
    // This is especially important when navigating back to root from other routes
    const verifyAndRedirect = async () => {
      if (!isLoading) {
        // Re-check auth status to ensure it's fresh
        await refresh();

        // The auth state will be updated by refresh, so we need to get it fresh
        // We'll handle the redirect in a separate effect that depends on auth state
      }
    };

    verifyAndRedirect();
  }, [refresh, isLoading, isDevelopment]);

  // Separate effect to handle redirects based on auth state
  useEffect(() => {
    if (!isLoading) {
      if (isLoggedIn) {
        navigate({ to: "/studies", replace: true });
      } else {
        navigate({
          to: "/login",
          replace: true,
          search: { message: "Please log in to access SlideInsight" },
        });
      }
    }
  }, [isLoggedIn, isLoading, navigate, isDevelopment]);

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  // Return null while redirecting
  return null;
}
