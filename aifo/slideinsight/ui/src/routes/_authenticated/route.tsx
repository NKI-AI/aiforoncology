// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { useLocation } from "@tanstack/react-router";
import ProtectedRoute from "@/auth/ProtectedRoute.tanstack";
import { LoadingFallback } from "@/components/LoadingFallback";

const MainLayout = React.lazy(() =>
  import("@/components/MainLayout").then((m) => ({ default: m.MainLayout }))
);

export const Route = createFileRoute("/_authenticated")({
  component: AuthenticatedComponent,
});

function AuthenticatedComponent() {
  const location = useLocation();

  // Check if we're on a viewer route (routes containing /v/ or /i/) or admin route
  const isViewerRoute =
    location.pathname.includes("/v/") || location.pathname.includes("/i/");
  const isAdminRoute = location.pathname.includes("/admin");

  return (
    <ProtectedRoute>
      <Suspense fallback={<LoadingFallback message="Loading..." />}>
        <MainLayout
          showFooter={!isViewerRoute && !isAdminRoute}
          backgroundColor={isViewerRoute ? "bg-black" : "bg-background"}
          isViewerRoute={isViewerRoute}
        />
      </Suspense>
    </ProtectedRoute>
  );
}
