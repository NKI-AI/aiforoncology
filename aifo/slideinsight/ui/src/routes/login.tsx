// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "@/components/LoadingFallback";

const LoginPage = React.lazy(() =>
  import("@/features/auth").then((module) => ({ default: module.LoginPage }))
);

export const Route = createFileRoute("/login")({
  component: LoginComponent,
});

function LoginComponent() {
  return (
    <Suspense fallback={<LoadingFallback message="Loading login..." />}>
      <LoginPage />
    </Suspense>
  );
}
