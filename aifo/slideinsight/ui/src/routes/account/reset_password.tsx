// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "../../components/LoadingFallback";

const ResetPasswordPage = React.lazy(() =>
  import("../../features/account").then((module) => ({
    default: module.ResetPasswordPage,
  }))
);

export const Route = createFileRoute("/account/reset_password")({
  component: ResetPasswordComponent,
});

function ResetPasswordComponent() {
  return (
    <Suspense
      fallback={<LoadingFallback message="Loading reset password..." />}
    >
      <ResetPasswordPage />
    </Suspense>
  );
}
