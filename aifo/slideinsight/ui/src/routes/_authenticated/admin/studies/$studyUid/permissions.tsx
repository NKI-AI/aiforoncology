// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "@/components/LoadingFallback";

const StudyPermissions = React.lazy(
  () => import("@/features/admin/components/studies/StudyPermissionsPage")
);

export const Route = createFileRoute(
  "/_authenticated/admin/studies/$studyUid/permissions"
)({
  component: StudyPermissionsComponent,
});

function StudyPermissionsComponent() {
  return (
    <Suspense
      fallback={<LoadingFallback message="Loading study permissions..." />}
    >
      <StudyPermissions />
    </Suspense>
  );
}
