// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React from "react";
import TenantDetailPage from "../../../../components/TenantDetailPage";
import ErrorStateAlert from "../../../../components/ErrorStateAlert";

export const Route = createFileRoute(
  "/_authenticated/admin/tenants/$tenantUid"
)({
  component: AdminTenantDetail,
});

function AdminTenantDetail() {
  const { tenantUid } = Route.useParams();

  try {
    return <TenantDetailPage tenantUid={tenantUid} />;
  } catch (error) {
    console.error("Error in AdminTenantDetail:", error);
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto">
          <ErrorStateAlert
            error={error instanceof Error ? error : new Error(String(error))}
            title="Component Error"
            variant="detailed"
          />
        </div>
      </div>
    );
  }
}
