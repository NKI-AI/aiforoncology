// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React from "react";
import SlideDetailPage from "@/features/admin/components/slides/SlideDetailPage";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import AdminPageLayout from "@/features/admin/components/AdminPageLayout";
import { Link } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/admin/slides/$slideUid")({
  component: AdminSlideDetail,
});

function AdminSlideDetail() {
  const { slideUid } = Route.useParams();

  try {
    return <SlideDetailPage slideUid={slideUid} />;
  } catch (error) {
    console.error("Error in AdminSlideDetail:", error);
    return (
      <AdminPageLayout
        title="Slide Details"
        description="Error loading slide information"
        actions={
          <Link
            to="/admin/slides"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Slides
          </Link>
        }
      >
        <ErrorStateAlert
          error={error instanceof Error ? error : new Error(String(error))}
          title="Component Error"
          onRetry={() => window.location.reload()}
          variant="detailed"
        />
      </AdminPageLayout>
    );
  }
}
