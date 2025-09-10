// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState, useRef, useEffect } from "react";
import {
  NotificationIcon,
  CheckIcon,
  CloseIcon,
  TrashIcon,
  ErrorIcon,
  WarningIcon,
  AlertIcon,
  RefreshIcon,
} from "@/components/icons";
import { useNotificationContext } from "@/contexts/NotificationContext";

interface NotificationDropdownProps {
  className?: string;
}

const NotificationDropdown: React.FC<NotificationDropdownProps> = ({
  className = "",
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const {
    notifications,
    stats,
    loading,
    error,
    markAsRead,
    markAsDismissed,
    markAllAsRead,
    deleteNotification,
    requestNotificationPermission,
    manualRefresh,
  } = useNotificationContext();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Request notification permission when component mounts
  useEffect(() => {
    requestNotificationPermission();
  }, [requestNotificationPermission]);

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case "error":
        return <ErrorIcon className="h-4 w-4 text-red-500" />;
      case "warning":
        return <WarningIcon className="h-4 w-4 text-yellow-500" />;
      case "success":
        return <CheckIcon className="h-4 w-4 text-green-500" />;
      default:
        return <AlertIcon className="h-4 w-4 text-blue-500" />;
    }
  };

  const getPriorityBadge = (priority: string) => {
    switch (priority) {
      case "urgent":
        return "bg-red-900/30 text-red-200";
      case "high":
        return "bg-orange-900/30 text-orange-200";
      case "normal":
        return "bg-blue-900/30 text-blue-200";
      case "low":
        return "bg-muted text-muted-foreground";
      default:
        return "bg-muted text-muted-foreground";
    }
  };

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) {
      return "Just now";
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60);
      return `${minutes}m ago`;
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600);
      return `${hours}h ago`;
    } else {
      const days = Math.floor(diffInSeconds / 86400);
      return `${days}d ago`;
    }
  };

  const handleNotificationClick = async (notification: any) => {
    if (!notification.isRead) {
      await markAsRead(notification.id);
    }

    if (notification.link) {
      window.open(notification.link, "_blank");
    }
  };

  const visibleNotifications = notifications.filter((n) => !n.isDismissed);

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      {/* Notification Button */}
      <button
        type="button"
        className="relative p-2 text-white hover:text-indigo-200 transition-colors rounded-md hover:bg-indigo-600"
        onClick={() => setIsOpen(!isOpen)}
        title="Notifications"
      >
        <NotificationIcon className="h-5 w-5" />

        {/* Unread Badge */}
        {stats.unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center min-w-[20px] notification-badge">
            {stats.unreadCount > 99 ? "99+" : stats.unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div className="notification-dropdown fixed inset-x-2 top-14 z-50 flex flex-col rounded-lg border bg-popover text-popover-foreground shadow-xl sm:absolute sm:inset-x-auto sm:right-0 sm:top-full sm:mt-2 sm:w-96">
          {/* Header */}
          <div className="px-3 py-2 border-b bg-muted/50 rounded-t-lg">
            <div className="flex items-center justify-between">
              <h3 className="text-sm sm:text-base font-semibold">
                Notifications
              </h3>
              <div className="flex items-center gap-2">
                <button
                  onClick={manualRefresh}
                  className="text-muted-foreground hover:text-foreground p-1 rounded hover:bg-muted"
                  title="Refresh notifications"
                >
                  <RefreshIcon className="h-4 w-4" />
                </button>
                {stats.unreadCount > 0 && (
                  <button
                    onClick={markAllAsRead}
                    className="text-xs text-indigo-300 hover:text-indigo-200 font-medium"
                  >
                    Mark all read
                  </button>
                )}
                <button
                  onClick={() => setIsOpen(false)}
                  className="text-muted-foreground hover:text-foreground"
                  title="Close"
                >
                  <CloseIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
            <div className="flex gap-4 mt-1.5 text-xs text-muted-400">
              <span>{stats.totalCount} total</span>
              <span>{stats.unreadCount} unread</span>
            </div>
          </div>

          {/* Content */}
          <div className="max-h-[60vh] sm:max-h-96 overflow-y-auto">
            {loading && (
              <div className="p-4 text-center text-muted-400">
                Loading notifications...
              </div>
            )}

            {error && (
              <div className="p-4 text-center text-red-400">
                <ErrorIcon className="h-6 w-6 mx-auto mb-2" />
                {error}
              </div>
            )}

            {!loading && !error && visibleNotifications.length === 0 && (
              <div className="p-8 text-center text-muted-400">
                <NotificationIcon className="h-12 w-12 mx-auto mb-3 text-muted-600" />
                <p className="text-sm">No notifications yet</p>
              </div>
            )}

            {!loading && !error && visibleNotifications.length > 0 && (
              <div className="divide-y divide-gray-700">
                {visibleNotifications.map((notification) => (
                  <div
                    key={notification.id}
                    className={`p-3 sm:p-4 transition-colors ${
                      !notification.isRead
                        ? "bg-indigo-900/30"
                        : "bg-transparent"
                    } hover:bg-gray-700/50`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div
                        className="flex-1 cursor-pointer"
                        onClick={() => handleNotificationClick(notification)}
                      >
                        <div className="flex items-center gap-2 mb-1">
                          {getNotificationIcon(notification.type)}
                          <h4
                            className={`text-sm font-medium ${
                              notification.isRead
                                ? "text-muted-300"
                                : "text-white"
                            }`}
                          >
                            {notification.title}
                          </h4>
                          {!notification.isRead && (
                            <div className="w-2 h-2 bg-blue-400 rounded-full"></div>
                          )}
                        </div>

                        <p className="text-sm text-muted-300 mb-2">
                          {notification.body}
                        </p>

                        <div className="flex items-center gap-2">
                          <span
                            className={`px-2 py-1 text-xs font-medium rounded-full ${getPriorityBadge(
                              notification.priority
                            )}`}
                          >
                            {notification.priority}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {formatRelativeTime(notification.createdAt)}
                          </span>
                        </div>
                      </div>

                      {/* Action Buttons */}
                      <div className="flex items-center gap-1">
                        {!notification.isRead && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              markAsRead(notification.id);
                            }}
                            className="p-1 text-muted-foreground hover:text-green-400 transition-colors"
                            title="Mark as read"
                          >
                            <CheckIcon className="h-4 w-4" />
                          </button>
                        )}

                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            markAsDismissed(notification.id);
                          }}
                          className="p-1 text-muted-foreground hover:text-yellow-400 transition-colors"
                          title="Dismiss"
                        >
                          <CloseIcon className="h-4 w-4" />
                        </button>

                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteNotification(notification.id);
                          }}
                          className="p-1 text-muted-foreground hover:text-red-400 transition-colors"
                          title="Delete"
                        >
                          <TrashIcon className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Footer */}
          {visibleNotifications.length > 0 && (
            <div className="px-3 py-2 border-t border-gray-700 bg-gray-900 rounded-b-lg">
              <p className="text-xs text-muted-foreground text-center">
                Showing {visibleNotifications.length} notifications
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default NotificationDropdown;
