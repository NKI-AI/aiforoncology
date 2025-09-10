// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/auth";
import { showUserNotificationToast } from "@/components/NotificationToast";
import { useWebSocketManager, WebSocketMessage } from "./useWebSocketManager";
import { useApiMutation, createApiMutation } from "@/utils/apiQueries";
import { apiFetch, ApiError } from "@/utils/fetchUtils";

// Types
interface Notification {
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

interface NotificationStats {
  totalCount: number;
  unreadCount: number;
  undismissedCount: number;
}

interface NotificationFilter {
  type?: string;
  priority?: string;
  isRead?: boolean;
  isDismissed?: boolean;
}

interface NotificationsApiResponse {
  status: string;
  data: {
    notifications: Notification[];
    stats: NotificationStats;
  };
}

interface NotificationWebSocketMessage extends WebSocketMessage {
  data?: {
    notification?: Notification;
    stats?: NotificationStats;
  };
}

// We need to accept a function to check if notification is self-sent
interface UseNotificationsOptions {
  isRecentlySelfSent?: () => boolean;
}

export const useNotifications = (options?: UseNotificationsOptions) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [stats, setStats] = useState<NotificationStats>({
    totalCount: 0,
    unreadCount: 0,
    undismissedCount: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { user, isLoading: authLoading } = useAuth();

  // WebSocket URL factory - only create URL when user is available
  const createWebSocketUrl = useCallback(() => {
    if (!user?.userUid) {
      throw new Error("User not available for WebSocket connection");
    }
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    return `${protocol}//${host}/api/v1/notifications/ws?userUid=${encodeURIComponent(
      user.userUid
    )}`;
  }, [user?.userUid]);

  // Handle incoming WebSocket messages
  const handleMessage = useCallback(
    (message: NotificationWebSocketMessage) => {
      switch (message.type) {
        case "notification":
          if (message.data?.notification) {
            const newNotification = message.data.notification;

            setNotifications((prev) => {
              const updated = [newNotification, ...prev];
              return updated;
            });

            // Update stats when a new notification arrives (increment counters)
            setStats((prev) => {
              const newStats = {
                totalCount: prev.totalCount + 1,
                unreadCount: prev.unreadCount + 1,
                undismissedCount: prev.undismissedCount + 1,
              };
              return newStats;
            });

            // Check if this notification was self-sent with detailed logging
            const selfSentCheckResult = options?.isRecentlySelfSent?.();
            const isLikelySelfSent = selfSentCheckResult === true;

            // Always show the toast for incoming notifications unless explicitly marked as self-sent
            if (!isLikelySelfSent) {
              try {
                // Show Facebook-style toast notification
                const toastId = showUserNotificationToast({
                  id: newNotification.id,
                  title: newNotification.title,
                  body: newNotification.body,
                  link: newNotification.link,
                  type: newNotification.type,
                  priority: newNotification.priority,
                  user: {
                    fullName: "System", // Since we don't have sender info in notifications
                    email: "system@slideinsight.net",
                  },
                });
              } catch (toastError) {}
            }
            // Show browser notification if supported (as fallback/additional notification)
            if (
              "Notification" in window &&
              Notification.permission === "granted"
            ) {
              try {
                const browserNotification = new Notification(
                  newNotification.title,
                  {
                    body: newNotification.body,
                    icon: "/favicon.ico",
                    tag: `notification-${newNotification.id}`, // Prevent duplicates
                  }
                );

                // Auto-close browser notification after 5 seconds
                setTimeout(() => {
                  browserNotification.close();
                }, 5000);
              } catch (browserNotificationError) {
                console.warn(
                  "⚠️ Failed to show browser notification:",
                  browserNotificationError
                );
              }
            }
          }
          break;

        case "stats":
          if (message.data?.stats) {
            setStats(message.data.stats);
          }
          break;

        case "connection":
          break;

        case "error":
          console.error("❌ WebSocket error received:", message);
          setError("WebSocket error received");
          break;

        default:
          console.warn("❓ Unknown WebSocket message type:", message.type);
      }
    },
    [options?.isRecentlySelfSent, user?.userUid]
  );

  // Use the WebSocket manager - only connect when user is available
  const {
    connected,
    connecting,
    error: wsError,
    reconnectAttempts,
    connect: wsConnect,
    disconnect: wsDisconnect,
    forceReconnect,
    getReadyState,
    getReadyStateText,
  } = useWebSocketManager(createWebSocketUrl, handleMessage, {
    autoConnect: false, // We'll manually control connection based on user state
    maxReconnectAttempts: 0, // Unlimited retries for notifications
    reconnectDelay: 3000,
    useExponentialBackoff: false,
    enablePing: true,
    pingInterval: 30000,
  });

  // Fetch notifications from API
  const fetchNotifications = useCallback(
    async (
      limit: number = 20,
      offset: number = 0,
      filter: NotificationFilter = {}
    ) => {
      setLoading(true);
      setError(null);

      try {
        const params = new URLSearchParams({
          limit: limit.toString(),
          offset: offset.toString(),
          ...Object.entries(filter).reduce((acc, [key, value]) => {
            if (value !== undefined) {
              acc[key] = value.toString();
            }
            return acc;
          }, {} as Record<string, string>),
        });

        const result = await apiFetch<NotificationsApiResponse>(
          `/api/v1/notifications?${params}`
        );

        if (result.status === "success") {
          setNotifications(result.data.notifications || []);
          if (result.data.stats) {
            setStats(result.data.stats);
          }
        } else {
          throw new Error("Failed to fetch notifications");
        }
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err.message);
        } else {
          setError("Failed to fetch notifications");
        }
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // Mutation for marking notification as read
  const markAsReadMutation = useApiMutation(
    (id: number) =>
      createApiMutation.put<void, {}>(`/api/v1/notifications/${id}/read`)({}),
    {
      onSuccess: (_, id) => {
        setNotifications((prev) =>
          prev.map((notification) =>
            notification.id === id
              ? {
                  ...notification,
                  isRead: true,
                  readAt: new Date().toISOString(),
                }
              : notification
          )
        );
        setStats((prev) => ({
          ...prev,
          unreadCount: Math.max(0, prev.unreadCount - 1),
        }));
      },
      onError: (err) => {
        setError(
          err instanceof ApiError ? err.message : "Failed to mark as read"
        );
      },
    }
  );

  // Mutation for marking notification as dismissed
  const markAsDismissedMutation = useApiMutation(
    (id: number) =>
      createApiMutation.put<void, {}>(`/api/v1/notifications/${id}/dismiss`)(
        {}
      ),
    {
      onSuccess: (_, id) => {
        setNotifications((prev) =>
          prev.map((notification) =>
            notification.id === id
              ? {
                  ...notification,
                  isDismissed: true,
                  dismissedAt: new Date().toISOString(),
                }
              : notification
          )
        );
        setStats((prev) => ({
          ...prev,
          undismissedCount: Math.max(0, prev.undismissedCount - 1),
        }));
      },
      onError: (err) => {
        setError(
          err instanceof ApiError
            ? err.message
            : "Failed to dismiss notification"
        );
      },
    }
  );

  // Mutation for marking all notifications as read
  const markAllAsReadMutation = useApiMutation(
    (_: void) =>
      createApiMutation.put<void, {}>("/api/v1/notifications/read-all")({}),
    {
      onSuccess: () => {
        setNotifications((prev) =>
          prev.map((notification) => ({
            ...notification,
            isRead: true,
            readAt: new Date().toISOString(),
          }))
        );
        setStats((prev) => ({
          ...prev,
          unreadCount: 0,
        }));
      },
      onError: (err) => {
        setError(
          err instanceof ApiError ? err.message : "Failed to mark all as read"
        );
      },
    }
  );

  // Mutation for deleting notification
  const deleteNotificationMutation = useApiMutation(
    (id: number) =>
      createApiMutation.delete<void>(`/api/v1/notifications/${id}`)(),
    {
      onSuccess: (_, id) => {
        const deletedNotification = notifications.find((n) => n.id === id);
        setNotifications((prev) =>
          prev.filter((notification) => notification.id !== id)
        );
        if (deletedNotification) {
          setStats((prev) => ({
            totalCount: Math.max(0, prev.totalCount - 1),
            unreadCount: deletedNotification.isRead
              ? prev.unreadCount
              : Math.max(0, prev.unreadCount - 1),
            undismissedCount: deletedNotification.isDismissed
              ? prev.undismissedCount
              : Math.max(0, prev.undismissedCount - 1),
          }));
        }
      },
      onError: (err) => {
        setError(
          err instanceof ApiError
            ? err.message
            : "Failed to delete notification"
        );
      },
    }
  );

  // Request browser notification permission
  const requestNotificationPermission = useCallback(async () => {
    if ("Notification" in window) {
      const permission = await Notification.requestPermission();
      return permission === "granted";
    }
    return false;
  }, []);

  // Add a manual refresh function for debugging
  const manualRefresh = useCallback(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  // WebSocket status check function
  const getWebSocketStatus = useCallback(() => {
    const status = {
      connected,
      connecting,
      readyState: getReadyState(),
      readyStateText: getReadyStateText(),
      userUid: user?.userUid,
      hasUser: !!user,
      reconnectAttempts,
    };
    return status;
  }, [
    connected,
    connecting,
    getReadyState,
    getReadyStateText,
    user?.userUid,
    reconnectAttempts,
  ]);

  // Initialize WebSocket and fetch notifications when user changes
  useEffect(() => {
    if (user?.userUid) {
      wsConnect();
      fetchNotifications();
    } else if (!authLoading) {
      // Only clean up if auth is not loading - this prevents cleanup during page refresh
      wsDisconnect();
      setNotifications([]);
      setStats({ totalCount: 0, unreadCount: 0, undismissedCount: 0 });
    } else {
      console.log("🔌 Auth still loading, keeping current notifications state");
    }
  }, [user?.userUid, authLoading, wsConnect, wsDisconnect, fetchNotifications]);

  // Combine WebSocket and local errors
  const combinedError = error || wsError;

  // Mark notification as read
  const markAsRead = useCallback(
    async (id: number) => {
      markAsReadMutation.mutate(id);
    },
    [markAsReadMutation]
  );

  // Mark notification as dismissed
  const markAsDismissed = useCallback(
    async (id: number) => {
      markAsDismissedMutation.mutate(id);
    },
    [markAsDismissedMutation]
  );

  // Mark all notifications as read
  const markAllAsRead = useCallback(async () => {
    markAllAsReadMutation.mutate(undefined);
  }, [markAllAsReadMutation]);

  // Delete notification
  const deleteNotification = useCallback(
    async (id: number) => {
      deleteNotificationMutation.mutate(id);
    },
    [deleteNotificationMutation, notifications]
  );

  return {
    notifications,
    stats,
    loading,
    error: combinedError,
    connected,
    connecting,
    reconnectAttempts,
    fetchNotifications,
    manualRefresh,
    markAsRead,
    markAsDismissed,
    markAllAsRead,
    deleteNotification,
    requestNotificationPermission,
    getWebSocketStatus,
    forceReconnect,
  };
};
