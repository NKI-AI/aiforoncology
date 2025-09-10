// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "@/components/LoadingFallback";

const TermsPage = React.lazy(() => import("@/pages/TermsPage"));

export const Route = createFileRoute("/terms")({
  component: TermsComponent,
});

function TermsComponent() {
  return (
    <Suspense fallback={<LoadingFallback message="Loading terms..." />}>
      <TermsPage />
    </Suspense>
  );
}
