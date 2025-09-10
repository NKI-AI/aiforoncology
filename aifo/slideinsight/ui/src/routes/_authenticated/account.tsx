// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React, { Suspense } from "react";
import { LoadingFallback } from "../../components/LoadingFallback";

const AccountPage = React.lazy(() =>
  import("@/features/account").then((module) => ({
    default: module.AccountPage,
  }))
);

export const Route = createFileRoute("/_authenticated/account")({
  component: AccountComponent,
});

function AccountComponent() {
  return (
    <Suspense fallback={<LoadingFallback message="Loading account..." />}>
      <AccountPage />
    </Suspense>
  );
}
