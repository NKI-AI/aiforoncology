// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { RouterProvider } from "@tanstack/react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "./auth";
import { router, useRouterContext } from "./router";
import { Toaster } from "./components/ui/sonner";
import { NotificationProvider } from "./contexts/NotificationContext";

// Create a QueryClient instance
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime)
      retry: (failureCount, error: any) => {
        // Don't retry on 401, 403, or 404 errors
        if (
          error?.status === 401 ||
          error?.status === 403 ||
          error?.status === 404
        ) {
          return false;
        }
        return failureCount < 3;
      },
    },
    mutations: {
      retry: false,
    },
  },
});

// Inner component that has access to auth context
function InnerApp() {
  const context = useRouterContext();

  return (
    <RouterProvider router={router} context={{ ...context, queryClient }} />
  );
}

// Main App component that provides auth context
export default function AppRouter() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <NotificationProvider>
          <InnerApp />
          <Toaster
            position="bottom-right"
            richColors
            closeButton
            visibleToasts={5}
            expand={true}
          />
        </NotificationProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}
