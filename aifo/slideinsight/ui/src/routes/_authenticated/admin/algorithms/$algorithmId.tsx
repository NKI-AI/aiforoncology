// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import React from "react";
import AlgorithmDetailPage from "@/features/admin/components/algorithms/AlgorithmDetailPage";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import AdminPageLayout from "@/features/admin/components/AdminPageLayout";
import { Link } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/admin/algorithms/$algorithmId"
)({
  component: AdminAlgorithmDetail,
});

function AdminAlgorithmDetail() {
  const { algorithmId } = Route.useParams();

  try {
    return <AlgorithmDetailPage algorithmId={algorithmId} />;
  } catch (error) {
    console.error("Error in AdminAlgorithmDetail:", error);
    return (
      <AdminPageLayout
        title="Algorithm Details"
        description="Error loading algorithm information"
        actions={
          <Link
            to="/admin/algorithms"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Algorithms
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
