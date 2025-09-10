// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface JobTypeStats {
  avg_ms: number;
  completed: number;
  failed: number;
}

interface QueueStats {
  by_type: Record<string, JobTypeStats>;
}

interface QueueJobTypesProps {
  stats: QueueStats;
}

export function QueueJobTypes({ stats }: QueueJobTypesProps) {
  const jobTypes = Object.entries(stats.by_type);

  if (jobTypes.length === 0) {
    return (
      <div className="bg-background rounded-lg border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-muted-900 mb-4">Job Types</h2>
        <div className="text-center py-8">
          <div className="text-muted-500">No job type data available</div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-background rounded-lg border border-gray-200 p-6">
      <h2 className="text-lg font-semibold text-muted-900 mb-4">Job Types</h2>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                Job Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                Completed
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                Failed
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                Avg Time
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                Success Rate
              </th>
            </tr>
          </thead>
          <tbody className="bg-background divide-y divide-gray-200">
            {jobTypes.map(([jobType, typeStats]) => {
              const total = typeStats.completed + typeStats.failed;
              const successRate =
                total > 0
                  ? ((typeStats.completed / total) * 100).toFixed(1)
                  : "0";
              const avgTimeSeconds = (typeStats.avg_ms / 1000).toFixed(2);

              return (
                <tr key={jobType} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="text-sm font-medium text-muted-900 capitalize">
                        {jobType}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-green-600 font-medium">
                      {typeStats.completed}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-red-600 font-medium">
                      {typeStats.failed}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-muted-900">
                      {avgTimeSeconds}s
                    </div>
                    <div className="text-xs text-muted-500">
                      {typeStats.avg_ms}ms
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="text-sm font-medium text-muted-900">
                        {successRate}%
                      </div>
                      <div className="ml-2 w-16 bg-gray-200 rounded-full h-2">
                        <div
                          className={`h-2 rounded-full transition-all duration-300 ${
                            parseFloat(successRate) >= 90
                              ? "bg-green-500"
                              : parseFloat(successRate) >= 70
                              ? "bg-yellow-500"
                              : "bg-red-500"
                          }`}
                          style={{ width: `${successRate}%` }}
                        ></div>
                      </div>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="mt-4 text-xs text-muted-500">
        * Success rate is calculated as completed jobs / (completed + failed)
        jobs
      </div>
    </div>
  );
}
