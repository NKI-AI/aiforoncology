// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute, Link } from "@tanstack/react-router";
import React, { useState, useEffect } from "react";
import { getFirstSlideUid } from "@/lib/routing/redirectHelpers";

export const Route = createFileRoute("/_authenticated/admin/cases/$caseUid")({
  component: AdminCaseDetail,
});

function AdminCaseDetail() {
  const { caseUid } = Route.useParams();
  const [firstSlideUid, setFirstSlideUid] = useState<string | null>(null);
  const [isLoadingSlides, setIsLoadingSlides] = useState(true);

  useEffect(() => {
    const fetchFirstSlide = async () => {
      setIsLoadingSlides(true);
      try {
        const slideUid = await getFirstSlideUid(caseUid);
        setFirstSlideUid(slideUid);
      } catch (error) {
        console.error("Failed to fetch first slide:", error);
        setFirstSlideUid(null);
      } finally {
        setIsLoadingSlides(false);
      }
    };

    fetchFirstSlide();
  }, [caseUid]);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <div className="bg-background rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-bold text-muted-900">
                Case Details
              </h1>
              <p className="text-sm text-muted-600 mt-1">Case ID: {caseUid}</p>
            </div>
            <Link
              to="/admin/cases"
              className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
            >
              ← Back to Cases
            </Link>
          </div>

          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* View Case Slides Link */}
              {isLoadingSlides ? (
                <div className="block p-4 bg-gray-50 border border-gray-200 rounded-lg">
                  <h3 className="font-semibold text-muted-900">
                    View Case Slides
                  </h3>
                  <p className="text-sm text-muted-700 mt-1">
                    Loading slides...
                  </p>
                </div>
              ) : firstSlideUid ? (
                <Link
                  to="/v/$caseUid/i/$slideUid"
                  params={{ caseUid, slideUid: firstSlideUid }}
                  className="block p-4 bg-blue-50 hover:bg-blue-100 border border-blue-200 rounded-lg transition"
                >
                  <h3 className="font-semibold text-blue-900">
                    View Case Slides
                  </h3>
                  <p className="text-sm text-blue-700 mt-1">
                    Browse and view all slides in this case
                  </p>
                </Link>
              ) : (
                <div className="block p-4 bg-gray-50 border border-gray-200 rounded-lg">
                  <h3 className="font-semibold text-muted-900">
                    View Case Slides
                  </h3>
                  <p className="text-sm text-muted-700 mt-1">
                    No slides available for this case
                  </p>
                </div>
              )}

              <div className="block p-4 bg-gray-50 border border-gray-200 rounded-lg">
                <h3 className="font-semibold text-muted-900">Case Analytics</h3>
                <p className="text-sm text-muted-700 mt-1">
                  View statistics and analytics for this case
                </p>
                <p className="text-xs text-muted-500 mt-2">Coming soon</p>
              </div>

              <div className="block p-4 bg-gray-50 border border-gray-200 rounded-lg">
                <h3 className="font-semibold text-muted-900">Export Data</h3>
                <p className="text-sm text-muted-700 mt-1">
                  Export case data and annotations
                </p>
                <p className="text-xs text-muted-500 mt-2">Coming soon</p>
              </div>

              <div className="block p-4 bg-gray-50 border border-gray-200 rounded-lg">
                <h3 className="font-semibold text-muted-900">Case Settings</h3>
                <p className="text-sm text-muted-700 mt-1">
                  Manage case permissions and settings
                </p>
                <p className="text-xs text-muted-500 mt-2">Coming soon</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
