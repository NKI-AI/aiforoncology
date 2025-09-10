// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";

interface SystemNotification {
  id: number;
  title: string;
  body: string;
  link?: string;
  type: "info" | "success" | "warning" | "error";
  priority: "low" | "normal" | "high" | "urgent";
  isRead: boolean;
  isDismissed: boolean;
  expiresAt?: string;
  createdAt: string;
  readAt?: string;
  dismissedAt?: string;
}

interface NotificationPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

interface NotificationOverviewProps {
  notifications: SystemNotification[];
  pagination?: NotificationPagination;
  typeStats: Record<string, number>;
  priorityStats: Record<string, number>;
}

export function NotificationOverview({
  notifications,
  pagination,
  typeStats,
  priorityStats,
}: NotificationOverviewProps) {
  const totalNotifications = pagination?.total || notifications.length;
  const unreadCount = notifications.filter((n) => !n.isRead).length;
  const dismissedCount = notifications.filter((n) => n.isDismissed).length;
  const activeCount = notifications.filter((n) => !n.isDismissed).length;

  const readPercentage =
    totalNotifications > 0
      ? (
          ((totalNotifications - unreadCount) / totalNotifications) *
          100
        ).toFixed(1)
      : "0";
  const dismissedPercentage =
    totalNotifications > 0
      ? ((dismissedCount / totalNotifications) * 100).toFixed(1)
      : "0";

  return (
    <div className="bg-background rounded-lg border border-gray-200 p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold text-muted-900">
          Notification Overview
        </h2>
        <div className="text-sm text-muted-500">
          Last updated: {new Date().toLocaleTimeString()}
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="text-center p-4 bg-blue-50 rounded-lg">
          <div className="text-2xl font-bold text-blue-600">
            {totalNotifications}
          </div>
          <div className="text-sm text-blue-700">Total Notifications</div>
        </div>

        <div className="text-center p-4 bg-red-50 rounded-lg">
          <div className="text-2xl font-bold text-red-600">{unreadCount}</div>
          <div className="text-sm text-red-700">Unread</div>
          <div className="text-xs text-red-600">
            {((unreadCount / totalNotifications) * 100).toFixed(1)}%
          </div>
        </div>

        <div className="text-center p-4 bg-green-50 rounded-lg">
          <div className="text-2xl font-bold text-green-600">{activeCount}</div>
          <div className="text-sm text-green-700">Active</div>
          <div className="text-xs text-green-600">
            {((activeCount / totalNotifications) * 100).toFixed(1)}%
          </div>
        </div>

        <div className="text-center p-4 bg-gray-50 rounded-lg">
          <div className="text-2xl font-bold text-muted-600">
            {dismissedCount}
          </div>
          <div className="text-sm text-muted-700">Dismissed</div>
          <div className="text-xs text-muted-600">{dismissedPercentage}%</div>
        </div>
      </div>

      {/* Progress Bars */}
      <div className="space-y-4">
        <div>
          <div className="flex justify-between text-sm text-muted-600 mb-2">
            <span>Read Status</span>
            <span>{readPercentage}% read</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="flex h-2 rounded-full overflow-hidden">
              <div
                className="bg-green-500 transition-all duration-300"
                style={{ width: `${readPercentage}%` }}
              ></div>
              <div
                className="bg-red-500 transition-all duration-300"
                style={{ width: `${100 - parseFloat(readPercentage)}%` }}
              ></div>
            </div>
          </div>
        </div>

        <div>
          <div className="flex justify-between text-sm text-muted-600 mb-2">
            <span>Dismissal Status</span>
            <span>{dismissedPercentage}% dismissed</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="flex h-2 rounded-full overflow-hidden">
              <div
                className="bg-green-500 transition-all duration-300"
                style={{ width: `${100 - parseFloat(dismissedPercentage)}%` }}
              ></div>
              <div
                className="bg-gray-500 transition-all duration-300"
                style={{ width: `${dismissedPercentage}%` }}
              ></div>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="mt-6 pt-4 border-t border-gray-100">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <div className="font-medium text-muted-900 mb-2">By Type</div>
            <div className="space-y-1">
              {Object.entries(typeStats).map(([type, count]) => (
                <div key={type} className="flex justify-between">
                  <span className="text-muted-600 capitalize">{type}</span>
                  <span className="font-medium">{count}</span>
                </div>
              ))}
            </div>
          </div>

          <div>
            <div className="font-medium text-muted-900 mb-2">By Priority</div>
            <div className="space-y-1">
              {Object.entries(priorityStats).map(([priority, count]) => (
                <div key={priority} className="flex justify-between">
                  <span className="text-muted-600 capitalize">{priority}</span>
                  <span className="font-medium">{count}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
