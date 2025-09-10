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

interface QueueStatusCardProps {
  stats: QueueStats;
}

export function QueueStatusCard({ stats }: QueueStatusCardProps) {
  const completionRate =
    stats.total > 0 ? ((stats.completed / stats.total) * 100).toFixed(1) : "0";
  const failureRate =
    stats.total > 0 ? ((stats.failed / stats.total) * 100).toFixed(1) : "0";

  return (
    <div className="bg-card rounded-lg border border-border p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-card-foreground">
          Queue Status
        </h2>
        <div className="flex items-center space-x-2">
          <div
            className={`h-3 w-3 rounded-full ${
              stats.running ? "bg-green-500" : "bg-red-500"
            }`}
          ></div>
          <span className="text-sm font-medium text-muted-700">
            {stats.running ? "Running" : "Stopped"}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="text-center">
          <div className="text-2xl font-bold text-muted-900">{stats.total}</div>
          <div className="text-sm text-muted-600">Total Jobs</div>
        </div>

        <div className="text-center">
          <div className="text-2xl font-bold text-green-600">
            {stats.completed}
          </div>
          <div className="text-sm text-muted-600">Completed</div>
          <div className="text-xs text-muted-500">{completionRate}%</div>
        </div>

        <div className="text-center">
          <div className="text-2xl font-bold text-red-600">{stats.failed}</div>
          <div className="text-sm text-muted-600">Failed</div>
          <div className="text-xs text-muted-500">{failureRate}%</div>
        </div>

        <div className="text-center">
          <div className="text-2xl font-bold text-yellow-600">
            {stats.pending}
          </div>
          <div className="text-sm text-muted-600">Pending</div>
        </div>
      </div>

      {/* Progress bar */}
      {stats.total > 0 && (
        <div className="mt-4">
          <div className="flex justify-between text-xs text-muted-600 mb-1">
            <span>Progress</span>
            <span>{completionRate}% complete</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="flex h-2 rounded-full overflow-hidden">
              <div
                className="bg-green-500 transition-all duration-300"
                style={{ width: `${completionRate}%` }}
              ></div>
              <div
                className="bg-red-500 transition-all duration-300"
                style={{ width: `${failureRate}%` }}
              ></div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
