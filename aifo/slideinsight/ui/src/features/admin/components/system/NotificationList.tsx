// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState } from "react";
import { apiFetch } from "@/utils/fetchUtils";
import { formatDate, formatRelativeTime } from "@/utils/format";
import { EditNotificationForm } from "./EditNotificationForm";

interface SystemNotificationUser {
  id: number;
  userUid: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  fullName: string;
}

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
  user?: SystemNotificationUser;
}

interface NotificationPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

interface NotificationListProps {
  notifications: SystemNotification[];
  pagination?: NotificationPagination;
  onNotificationUpdated?: () => void;
}

export function NotificationList({
  notifications,
  pagination,
  onNotificationUpdated,
}: NotificationListProps) {
  const [editingNotification, setEditingNotification] =
    useState<SystemNotification | null>(null);
  const [deletingIds, setDeletingIds] = useState<Set<number>>(new Set());

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

  const truncateText = (text: string, maxLength: number = 100) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + "...";
  };

  const handleEdit = (notification: SystemNotification) => {
    setEditingNotification(notification);
  };

  const handleEditSuccess = () => {
    setEditingNotification(null);
    if (onNotificationUpdated) {
      onNotificationUpdated();
    }
  };

  const handleEditCancel = () => {
    setEditingNotification(null);
  };

  const handleDelete = async (notification: SystemNotification) => {
    const confirmMessage = `Are you sure you want to delete this notification?\n\nTitle: ${
      notification.title
    }\nUser: ${
      notification.user?.fullName || "Unknown"
    }\n\nThis action cannot be undone.`;

    if (!confirm(confirmMessage)) {
      return;
    }

    setDeletingIds((prev) => new Set(prev).add(notification.id));

    try {
      await apiFetch(`/api/v1/admin/notifications/${notification.id}`, {
        method: "DELETE",
      });

      if (onNotificationUpdated) {
        onNotificationUpdated();
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to delete notification";
      alert(`Error: ${errorMessage}`);
      console.error("Delete notification error:", err);
    } finally {
      setDeletingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(notification.id);
        return newSet;
      });
    }
  };

  if (notifications.length === 0) {
    return (
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-lg font-semibold text-card-foreground mb-4">
          Recent Notifications
        </h2>
        <div className="text-center py-8">
          <div className="text-muted-foreground">No notifications found</div>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="bg-card rounded-lg border border-border p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-card-foreground">
            Recent Notifications
          </h2>
          {pagination && (
            <div className="text-sm text-muted-foreground">
              Showing {notifications.length} of {pagination.total} notifications
            </div>
          )}
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Notification
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Priority
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-background divide-y divide-gray-200">
              {notifications.map((notification) => (
                <tr
                  key={notification.id}
                  className={`hover:bg-gray-50 ${
                    !notification.isRead ? "bg-blue-50" : ""
                  }`}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-start">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center">
                          <div
                            className={`text-sm font-medium ${
                              !notification.isRead
                                ? "text-muted-900"
                                : "text-muted-700"
                            }`}
                          >
                            {notification.title}
                          </div>
                          {!notification.isRead && (
                            <div className="ml-2 w-2 h-2 bg-blue-500 rounded-full"></div>
                          )}
                        </div>
                        <div className="text-sm text-muted-500 mt-1">
                          {truncateText(notification.body)}
                        </div>
                        {notification.link && (
                          <div className="text-xs text-blue-600 mt-1">
                            <a
                              href={notification.link}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="hover:underline"
                            >
                              View Link
                            </a>
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {notification.user ? (
                      <div className="flex flex-col">
                        <div className="text-sm font-medium text-muted-900">
                          {notification.user.fullName}
                        </div>
                        <div className="text-sm text-muted-500">
                          @{notification.user.firstName}
                        </div>
                        {notification.user.email && (
                          <div className="text-xs text-muted-400">
                            {notification.user.email}
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="text-sm text-muted-400">Unknown User</div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded border capitalize ${getTypeBadgeColor(
                        notification.type
                      )}`}
                    >
                      {notification.type}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-1 text-xs font-medium rounded border capitalize ${getPriorityBadgeColor(
                        notification.priority
                      )}`}
                    >
                      {notification.priority}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-col space-y-1">
                      {notification.isDismissed ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-muted-800">
                          Dismissed
                        </span>
                      ) : notification.isRead ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                          Read
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                          Unread
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-muted-900">
                      {formatRelativeTime(notification.createdAt)}
                    </div>
                    <div className="text-xs text-muted-500">
                      {formatDate(notification.createdAt)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div className="flex space-x-2">
                      <button
                        onClick={() => handleEdit(notification)}
                        className="text-indigo-600 hover:text-indigo-900"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(notification)}
                        disabled={deletingIds.has(notification.id)}
                        className="text-red-600 hover:text-red-900 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {deletingIds.has(notification.id)
                          ? "Deleting..."
                          : "Delete"}
                      </button>
                      <button
                        onClick={() => {
                          const userInfo = notification.user
                            ? `User: ${notification.user.fullName} ${
                                notification.user.email
                                  ? "\nEmail: " + notification.user.email
                                  : ""
                              }\n`
                            : "User: Unknown\n";
                          const details = `${userInfo}ID: ${notification.id}
Title: ${notification.title}
Body: ${notification.body}
Type: ${notification.type}
Priority: ${notification.priority}
Created: ${formatDate(notification.createdAt)}
${notification.readAt ? `Read: ${formatDate(notification.readAt)}` : ""}
${
  notification.dismissedAt
    ? `Dismissed: ${formatDate(notification.dismissedAt)}`
    : ""
}
${
  notification.expiresAt ? `Expires: ${formatDate(notification.expiresAt)}` : ""
}`;
                          alert(details);
                        }}
                        className="text-muted-600 hover:text-muted-900"
                      >
                        View Details
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {pagination && pagination.total > pagination.limit && (
          <div className="mt-4 flex items-center justify-between border-t border-gray-200 pt-4">
            <div className="text-sm text-muted-700">
              Showing page {pagination.page} of {pagination.totalPages}
            </div>
            <div className="text-sm text-muted-500">
              Total: {pagination.total} notifications
            </div>
          </div>
        )}
      </div>

      {/* Edit Notification Modal */}
      {editingNotification && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
          <div className="bg-background rounded-lg max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
            <h2 className="text-lg font-semibold mb-4">Edit Notification</h2>
            <EditNotificationForm
              notification={editingNotification}
              onSuccess={handleEditSuccess}
              onCancel={handleEditCancel}
            />
          </div>
        </div>
      )}
    </>
  );
}
