// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useEffect, useState, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/utils/fetchUtils";
import AdminSidebar from "../AdminSidebar";
import AdminHeader from "../AdminHeader";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { NotificationOverview } from "./NotificationOverview";
import { NotificationStats } from "./NotificationStats";
import { NotificationList } from "./NotificationList";
import { SendNotificationForm } from "../users/SendNotificationForm";
import { NotificationIcon } from "@/components/icons";
import { BellSlashIcon } from "@heroicons/react/24/outline";
import { Button } from "../../../../components/ui/button";

interface SystemNotificationUser {
  id: number;
  userUid: string;
  email: string;
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

interface NotificationStats {
  totalCount: number;
  unreadCount: number;
  undismissedCount: number;
}

interface NotificationsResponse {
  notifications: SystemNotification[];
  pagination: NotificationPagination;
  stats?: NotificationStats;
}

interface NotificationAdminResponse {
  status: string;
  data: NotificationsResponse;
}

export default function NotificationAdminPage() {
  // Send notification modal state
  const [isSendNotificationModalOpen, setIsSendNotificationModalOpen] =
    useState(false);

  // Set document title
  useEffect(() => {
    document.title = "SlideInsight - Notification Management";
    return () => {
      document.title = "SlideInsight Viewer";
    };
  }, []);

  const {
    data: notificationData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["admin", "notifications"],
    queryFn: async () => {
      return await apiFetch<NotificationAdminResponse>(
        "/api/v1/admin/notifications?limit=50"
      );
    },
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 60 * 1000, // Auto refresh every minute
  });

  const handleRefresh = () => {
    refetch();
  };

  // Send notification handlers
  const handleSendNotificationOpen = useCallback(() => {
    setIsSendNotificationModalOpen(true);
  }, []);

  const handleSendNotificationSuccess = useCallback(() => {
    setIsSendNotificationModalOpen(false);
    // Refresh the notifications list after sending
    refetch();
  }, [refetch]);

  const handleSendNotificationCancel = useCallback(() => {
    setIsSendNotificationModalOpen(false);
  }, []);

  const headerActions = (
    <div className="flex items-center space-x-2">
      <Button onClick={handleSendNotificationOpen} variant="default">
        <NotificationIcon className="h-4 w-4 mr-2" />
        Send Notification
      </Button>
      <Button onClick={handleRefresh} disabled={isLoading} variant="secondary">
        {isLoading ? (
          <>
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
            Refreshing...
          </>
        ) : (
          "Refresh"
        )}
      </Button>
    </div>
  );

  const notifications = notificationData?.data.notifications || [];
  const pagination = notificationData?.data.pagination;
  const stats = notificationData?.data.stats;

  // Calculate additional stats
  const typeStats = notifications.reduce((acc, notification) => {
    acc[notification.type] = (acc[notification.type] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const priorityStats = notifications.reduce((acc, notification) => {
    acc[notification.priority] = (acc[notification.priority] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title="Notification Management"
          description="Monitor and manage all system notifications"
          actions={headerActions}
        />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {/* Error state */}
              {error && (
                <div className="px-4 lg:px-6">
                  <div className="max-w-7xl">
                    <ErrorStateAlert
                      error={
                        error instanceof Error
                          ? error.message
                          : "Failed to load notification data"
                      }
                      title="Unable to load admin data"
                      onRetry={handleRefresh}
                      isRetrying={isLoading}
                      variant="detailed"
                    />
                  </div>
                </div>
              )}

              {/* Main content */}
              {notifications.length > 0 && (
                <>
                  {/* Overview */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl">
                      <NotificationOverview
                        notifications={notifications}
                        pagination={pagination}
                        typeStats={typeStats}
                        priorityStats={priorityStats}
                      />
                    </div>
                  </div>

                  {/* Stats and Overview */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl">
                      <NotificationStats
                        stats={stats}
                        typeStats={typeStats}
                        priorityStats={priorityStats}
                      />
                    </div>
                  </div>

                  {/* Notification List */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl">
                      <NotificationList
                        notifications={notifications}
                        pagination={pagination}
                        onNotificationUpdated={handleRefresh}
                      />
                    </div>
                  </div>
                </>
              )}

              {/* Empty state */}
              {!isLoading && !error && notifications.length === 0 && (
                <div className="px-4 lg:px-6">
                  <div className="max-w-7xl">
                    <div className="bg-background rounded-lg border border-gray-200 p-12 text-center">
                      <div className="text-muted-400 mb-4">
                        <BellSlashIcon className="h-16 w-16 mx-auto" />
                      </div>
                      <h3 className="text-lg font-medium text-muted-900 mb-2">
                        No notifications found
                      </h3>
                      <p className="text-muted-500">
                        There are currently no notifications in the system.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Loading state */}
              {isLoading && notifications.length === 0 && (
                <div className="px-4 lg:px-6">
                  <div className="max-w-7xl">
                    <div className="space-y-4">
                      <div className="animate-pulse bg-gray-200 h-32 rounded-lg"></div>
                      <div className="animate-pulse bg-gray-200 h-48 rounded-lg"></div>
                      <div className="animate-pulse bg-gray-200 h-64 rounded-lg"></div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Send Notification Modal */}
        {isSendNotificationModalOpen && (
          <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
            <div className="bg-background rounded-lg max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
              <h2 className="text-lg font-semibold mb-4">Send Notification</h2>
              <SendNotificationForm
                onSuccess={handleSendNotificationSuccess}
                onCancel={handleSendNotificationCancel}
              />
            </div>
          </div>
        )}
      </SidebarInset>
    </SidebarProvider>
  );
}
