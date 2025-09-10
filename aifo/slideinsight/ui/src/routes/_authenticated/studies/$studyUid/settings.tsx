// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { useParams } from "@tanstack/react-router";
import { LoadingFallback } from "@/components/LoadingFallback";

const StudySettings = React.lazy(() => import("@/components/StudySettings"));

export const Route = createFileRoute(
  "/_authenticated/studies/$studyUid/settings"
)({
  component: StudySettingsComponent,
});

function StudySettingsComponent() {
  const { tab } = useParams({
    from: "/_authenticated/studies/$studyUid/settings" as any,
  });
  const initialTab = typeof tab === "string" ? tab : undefined;
  return (
    <Suspense
      fallback={<LoadingFallback message="Loading study settings..." />}
    >
      <StudySettings initialTab={initialTab} />
    </Suspense>
  );
}
