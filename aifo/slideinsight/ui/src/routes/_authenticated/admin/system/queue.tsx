// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "@/components/LoadingFallback";

const SystemQueuePage = React.lazy(
  () => import("@/features/admin/components/system/SystemQueuePage")
);

export const Route = createFileRoute("/_authenticated/admin/system/queue")({
  component: SystemQueueComponent,
});

function SystemQueueComponent() {
  return (
    <Suspense fallback={<LoadingFallback message="Loading queue status..." />}>
      <SystemQueuePage />
    </Suspense>
  );
}
