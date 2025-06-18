// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../auth";

export default function RootRedirect() {
  const { isLoggedIn, loading } = useAuth();

  // Show loading state while checking authentication
  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  // Redirect based on authentication status
  return isLoggedIn ? (
    <Navigate to="/slides" replace />
  ) : (
    <Navigate
      to="/login"
      replace
      state={{ message: "Please log in to access SlideScope" }}
    />
  );
}
