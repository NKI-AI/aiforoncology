// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface QueueStats {
  total: number;
  completed: number;
  failed: number;
  pending: number;
  running: boolean;
  workers: number;
  default_rate_limit: number;
}

interface QueueOverviewProps {
  stats: QueueStats;
}

export function QueueOverview({ stats }: QueueOverviewProps) {
  const activeJobs = stats.total - stats.completed - stats.failed;

  return (
    <div className="bg-background rounded-lg border border-gray-200 p-6">
      <h2 className="text-lg font-semibold text-muted-900 mb-4">
        Queue Overview
      </h2>

      <div className="space-y-4">
        <div className="flex items-center justify-between py-2 border-b border-gray-100">
          <span className="text-sm text-muted-600">Queue Status</span>
          <span
            className={`text-sm font-medium ${
              stats.running ? "text-green-600" : "text-red-600"
            }`}
          >
            {stats.running ? "Active" : "Stopped"}
          </span>
        </div>

        <div className="flex items-center justify-between py-2 border-b border-gray-100">
          <span className="text-sm text-muted-600">Active Workers</span>
          <span className="text-sm font-medium text-muted-900">
            {stats.workers}
          </span>
        </div>

        <div className="flex items-center justify-between py-2 border-b border-gray-100">
          <span className="text-sm text-muted-600">Rate Limit</span>
          <span className="text-sm font-medium text-muted-900">
            {stats.default_rate_limit} jobs/sec
          </span>
        </div>

        <div className="flex items-center justify-between py-2 border-b border-gray-100">
          <span className="text-sm text-muted-600">Jobs in Progress</span>
          <span className="text-sm font-medium text-blue-600">
            {activeJobs}
          </span>
        </div>

        <div className="flex items-center justify-between py-2">
          <span className="text-sm text-muted-600">Pending Jobs</span>
          <span className="text-sm font-medium text-yellow-600">
            {stats.pending}
          </span>
        </div>
      </div>

      <div className="mt-6 pt-4 border-t border-gray-100">
        <div className="text-xs text-muted-500">
          Last updated: {new Date().toLocaleTimeString()}
        </div>
      </div>
    </div>
  );
}
