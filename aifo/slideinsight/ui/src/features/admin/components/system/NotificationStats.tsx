// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface NotificationStatsData {
  totalCount: number;
  unreadCount: number;
  undismissedCount: number;
}

interface NotificationStatsProps {
  stats?: NotificationStatsData;
  typeStats: Record<string, number>;
  priorityStats: Record<string, number>;
}

export function NotificationStats({
  stats,
  typeStats,
  priorityStats,
}: NotificationStatsProps) {
  const getTypeBadgeColor = (type: string) => {
    switch (type) {
      case "error":
        return "bg-red-100 text-red-800 border-red-200";
      case "warning":
        return "bg-yellow-100 text-yellow-800 border-yellow-200";
      case "success":
        return "bg-green-100 text-green-800 border-green-200";
      case "info":
      default:
        return "bg-blue-100 text-blue-800 border-blue-200";
    }
  };

  const getPriorityBadgeColor = (priority: string) => {
    switch (priority) {
      case "urgent":
        return "bg-red-100 text-red-800 border-red-200";
      case "high":
        return "bg-orange-100 text-orange-800 border-orange-200";
      case "normal":
        return "bg-blue-100 text-blue-800 border-blue-200";
      case "low":
      default:
        return "bg-gray-100 text-muted-800 border-gray-200";
    }
  };

  const totalTypes = Object.values(typeStats).reduce(
    (sum, count) => sum + count,
    0
  );
  const totalPriorities = Object.values(priorityStats).reduce(
    (sum, count) => sum + count,
    0
  );

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      {/* Type Distribution */}
      <div className="bg-background rounded-lg border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-muted-900 mb-4">
          Notification Types
        </h3>

        <div className="space-y-3">
          {Object.entries(typeStats).map(([type, count]) => {
            const percentage =
              totalTypes > 0 ? ((count / totalTypes) * 100).toFixed(1) : "0";

            return (
              <div key={type} className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded border capitalize ${getTypeBadgeColor(
                      type
                    )}`}
                  >
                    {type}
                  </span>
                  <span className="text-sm text-muted-600">{percentage}%</span>
                </div>
                <div className="flex items-center space-x-2">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full transition-all duration-300 ${
                        type === "error"
                          ? "bg-red-500"
                          : type === "warning"
                          ? "bg-yellow-500"
                          : type === "success"
                          ? "bg-green-500"
                          : "bg-blue-500"
                      }`}
                      style={{ width: `${percentage}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-muted-900 w-8 text-right">
                    {count}
                  </span>
                </div>
              </div>
            );
          })}
        </div>

        {totalTypes === 0 && (
          <div className="text-center py-8 text-muted-500">
            No type data available
          </div>
        )}
      </div>

      {/* Priority Distribution */}
      <div className="bg-background rounded-lg border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-muted-900 mb-4">
          Priority Levels
        </h3>

        <div className="space-y-3">
          {Object.entries(priorityStats).map(([priority, count]) => {
            const percentage =
              totalPriorities > 0
                ? ((count / totalPriorities) * 100).toFixed(1)
                : "0";

            return (
              <div key={priority} className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded border capitalize ${getPriorityBadgeColor(
                      priority
                    )}`}
                  >
                    {priority}
                  </span>
                  <span className="text-sm text-muted-600">{percentage}%</span>
                </div>
                <div className="flex items-center space-x-2">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full transition-all duration-300 ${
                        priority === "urgent"
                          ? "bg-red-500"
                          : priority === "high"
                          ? "bg-orange-500"
                          : priority === "normal"
                          ? "bg-blue-500"
                          : "bg-gray-500"
                      }`}
                      style={{ width: `${percentage}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-muted-900 w-8 text-right">
                    {count}
                  </span>
                </div>
              </div>
            );
          })}
        </div>

        {totalPriorities === 0 && (
          <div className="text-center py-8 text-muted-500">
            No priority data available
          </div>
        )}
      </div>

      {/* Overall Stats */}
      {stats && (
        <div className="bg-background rounded-lg border border-gray-200 p-6 lg:col-span-2">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            System-wide Statistics
          </h3>

          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-4 bg-blue-50 rounded-lg">
              <div className="text-2xl font-bold text-blue-600">
                {stats.totalCount}
              </div>
              <div className="text-sm text-blue-700">Total Count</div>
            </div>

            <div className="text-center p-4 bg-red-50 rounded-lg">
              <div className="text-2xl font-bold text-red-600">
                {stats.unreadCount}
              </div>
              <div className="text-sm text-red-700">Unread Count</div>
            </div>

            <div className="text-center p-4 bg-green-50 rounded-lg">
              <div className="text-2xl font-bold text-green-600">
                {stats.undismissedCount}
              </div>
              <div className="text-sm text-green-700">Undismissed Count</div>
            </div>
          </div>

          <div className="mt-4 text-xs text-muted-500 text-center">
            * System-wide statistics across all users
          </div>
        </div>
      )}
    </div>
  );
}
