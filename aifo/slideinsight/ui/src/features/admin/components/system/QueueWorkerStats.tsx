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

interface QueueWorkerStatsProps {
  stats: QueueStats;
}

export function QueueWorkerStats({ stats }: QueueWorkerStatsProps) {
  const jobsPerWorker =
    stats.workers > 0 ? (stats.completed / stats.workers).toFixed(1) : "0";
  const currentThroughput = stats.workers * stats.default_rate_limit;

  return (
    <div className="bg-background rounded-lg border border-gray-200 p-6">
      <h2 className="text-lg font-semibold text-muted-900 mb-4">
        Worker Statistics
      </h2>

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="text-center p-4 bg-blue-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {stats.workers}
            </div>
            <div className="text-sm text-blue-700">Active Workers</div>
          </div>

          <div className="text-center p-4 bg-purple-50 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">
              {currentThroughput}
            </div>
            <div className="text-sm text-purple-700">Max Throughput/sec</div>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-100">
          <div className="flex items-center justify-between py-2">
            <span className="text-sm text-muted-600">Jobs per Worker</span>
            <span className="text-sm font-medium text-muted-900">
              {jobsPerWorker}
            </span>
          </div>

          <div className="flex items-center justify-between py-2">
            <span className="text-sm text-muted-600">Worker Rate Limit</span>
            <span className="text-sm font-medium text-muted-900">
              {stats.default_rate_limit} jobs/sec
            </span>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-100">
          <div className="text-sm text-muted-600 mb-2">Worker Status</div>
          <div className="flex items-center space-x-3">
            <div className="flex items-center space-x-2">
              <div
                className={`h-2 w-2 rounded-full ${
                  stats.running ? "bg-green-500" : "bg-gray-400"
                }`}
              ></div>
              <span className="text-xs text-muted-600">
                {stats.running ? "Running" : "Idle"}
              </span>
            </div>
            {stats.pending > 0 && (
              <div className="flex items-center space-x-2">
                <div className="h-2 w-2 rounded-full bg-yellow-500"></div>
                <span className="text-xs text-muted-600">Queue backlog</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
