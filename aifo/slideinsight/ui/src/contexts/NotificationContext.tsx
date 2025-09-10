// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { createContext, useContext, useCallback, useRef } from "react";
import { useNotifications } from "@/hooks/useNotifications";
import { useAuth } from "@/auth";
import { showUserNotificationToast } from "@/components/NotificationToast";

interface NotificationContextType {
  notifications: any[];
  stats: {
    totalCount: number;
    unreadCount: number;
    undismissedCount: number;
  };
  loading: boolean;
  error: string | null;
  // Actions
  markAsRead: (id: number) => Promise<void>;
  markAsDismissed: (id: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  deleteNotification: (id: number) => Promise<void>;
  fetchNotifications: (
    limit?: number,
    offset?: number,
    filter?: any
  ) => Promise<void>;
  manualRefresh: () => void;
  requestNotificationPermission: () => Promise<boolean>;
  // Event handlers for external updates
  onNotificationUpdate: () => void;
  // Self-sent notification tracking
  markNotificationAsSelfSent: (targetUserUid?: string) => void;
  isRecentlySelfSent: () => boolean;
  // Debug functions
  testNotificationToast: () => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(
  undefined
);

export const useNotificationContext = () => {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error(
      "useNotificationContext must be used within a NotificationProvider"
    );
  }
  return context;
};

interface NotificationProviderProps {
  children: React.ReactNode;
}

export const NotificationProvider: React.FC<NotificationProviderProps> = ({
  children,
}) => {
  const { user } = useAuth();

  // Track self-sent notifications to avoid duplicate toasts
  const lastSelfSentTime = useRef<number>(0);
  const lastSelfSentTargetUser = useRef<string>("");

  // Check if a notification was recently self-sent (within 5 seconds)
  const isRecentlySelfSent = useCallback(() => {
    const currentTime = Date.now();
    const timeDiff = currentTime - lastSelfSentTime.current;
    const isRecent = timeDiff < 5000; // Extended to 5 seconds threshold
    const isSelfTarget = lastSelfSentTargetUser.current === user?.userUid;
    const isLikelySelfSent =
      isRecent && (isSelfTarget || lastSelfSentTargetUser.current === "");

    return isLikelySelfSent;
  }, [user?.userUid]);

  const {
    notifications,
    stats,
    loading,
    error,
    markAsRead,
    markAsDismissed,
    markAllAsRead,
    deleteNotification,
    fetchNotifications,
    manualRefresh,
    requestNotificationPermission,
    getWebSocketStatus,
    forceReconnect,
  } = useNotifications({ isRecentlySelfSent });

  // Handler for external updates (like from admin actions)
  const onNotificationUpdate = useCallback(() => {
    manualRefresh();
  }, [manualRefresh]);

  // Mark notification as self-sent (called when sending from admin)
  const markNotificationAsSelfSent = useCallback(
    (targetUserUid?: string) => {
      lastSelfSentTime.current = Date.now();
      lastSelfSentTargetUser.current = targetUserUid || "";
    },
    [user?.userUid]
  );

  // Test function to manually trigger a notification toast
  const testNotificationToast = useCallback(() => {
    try {
      const testNotification = {
        id: Date.now(),
        title: "Test Notification",
        body: "This is a test notification to verify the toast system is working.",
        type: "info" as const,
        priority: "normal" as const,
        user: {
          fullName: "Test User",
          email: "test@test.com",
        },
      };

      const toastId = showUserNotificationToast(testNotification);
    } catch (error) {}
  }, []);

  const contextValue: NotificationContextType = {
    notifications,
    stats,
    loading,
    error,
    markAsRead,
    markAsDismissed,
    markAllAsRead,
    deleteNotification,
    fetchNotifications,
    manualRefresh,
    requestNotificationPermission,
    onNotificationUpdate,
    markNotificationAsSelfSent,
    isRecentlySelfSent,
    testNotificationToast,
  };

  // Expose debug functions globally for console testing
  React.useEffect(() => {
    if (typeof window !== "undefined") {
      (window as any).debugNotifications = {
        testToast: testNotificationToast,
        checkSelfSent: isRecentlySelfSent,
        markAsSelfSent: markNotificationAsSelfSent,
        refresh: manualRefresh,
        checkWebSocket: getWebSocketStatus,
        reconnectWebSocket: forceReconnect,
        stats,
        notifications,
        user: user?.userUid,
      };
    }
  }, [
    testNotificationToast,
    isRecentlySelfSent,
    markNotificationAsSelfSent,
    manualRefresh,
    getWebSocketStatus,
    forceReconnect,
    stats,
    notifications,
    user,
  ]);

  return (
    <NotificationContext.Provider value={contextValue}>
      {children}
    </NotificationContext.Provider>
  );
};
