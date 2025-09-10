// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "../../../components/LoadingFallback";

// Lazy load the components
const SlideSelector = React.lazy(
  () => import("../../../components/SlideSelector")
);
const Navbar = React.lazy(() => import("../../../components/Navbar"));

export const Route = createFileRoute("/_authenticated/v/")({
  component: ViewerComponent,
});

function ViewerComponent() {
  return (
    <div className="bg-gray-100 text-muted-800 font-sans h-full w-full flex flex-col">
      <Suspense fallback={<LoadingFallback message="Loading navbar..." />}>
        <Navbar onToggleSlideInfo={() => {}} onToggleHelp={() => {}} />
      </Suspense>

      <div className="relative flex-1">
        <Suspense fallback={<LoadingFallback message="Loading viewer..." />}>
          <SlideSelector />
        </Suspense>
      </div>
    </div>
  );
}
