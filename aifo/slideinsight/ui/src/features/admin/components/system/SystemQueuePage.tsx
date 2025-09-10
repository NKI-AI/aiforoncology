// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/utils/fetchUtils";
import AdminSidebar from "../AdminSidebar";
import AdminHeader from "../AdminHeader";
import { AdminErrorAlert } from "../AdminErrorAlert";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { QueueOverview } from "./QueueOverview";
import { QueueWorkerStats } from "./QueueWorkerStats";
import { QueueJobTypes } from "./QueueJobTypes";
import { QueueStatusCard } from "./QueueStatusCard";

interface QueueStats {
  total: number;
  completed: number;
  failed: number;
  pending: number;
  by_type: Record<
    string,
    {
      avg_ms: number;
      completed: number;
      failed: number;
    }
  >;
  running: boolean;
  workers: number;
  default_rate_limit: number;
}

interface QueueResponse {
  status: string;
  data: QueueStats;
}

export default function SystemQueuePage() {
  // Set document title
  useEffect(() => {
    document.title = "SlideInsight - Queue Monitoring";
    return () => {
      document.title = "SlideInsight Viewer";
    };
  }, []);

  const {
    data: queueData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["admin", "system", "queue"],
    queryFn: async () => {
      return await apiFetch<QueueResponse>("/api/v1/system/queue");
    },
    staleTime: 30 * 1000, // 30 seconds
    gcTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 30 * 1000, // Auto refresh every 30 seconds
  });

  const handleRefresh = () => {
    refetch();
  };

  const headerActions = (
    <button
      onClick={handleRefresh}
      disabled={isLoading}
      className="inline-flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-400 text-white text-sm font-medium rounded-lg transition shadow-sm hover:shadow focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50"
    >
      {isLoading ? (
        <>
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
          Refreshing...
        </>
      ) : (
        "Refresh"
      )}
    </button>
  );

  const stats = queueData?.data;

  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title="Queue Monitoring"
          description="Monitor queue status, job processing, and worker performance"
          actions={headerActions}
        />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {/* Error state */}
              {error && (
                <AdminErrorAlert
                  error={
                    error instanceof Error
                      ? error.message
                      : "Failed to load queue data"
                  }
                  loading={isLoading}
                  onRetry={handleRefresh}
                />
              )}

              {/* Main content */}
              {stats && (
                <>
                  {/* Status Cards */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl">
                      <QueueStatusCard stats={stats} />
                    </div>
                  </div>

                  {/* Overview and Worker Stats */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl grid gap-6 lg:grid-cols-2">
                      <QueueOverview stats={stats} />
                      <QueueWorkerStats stats={stats} />
                    </div>
                  </div>

                  {/* Job Types */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl">
                      <QueueJobTypes stats={stats} />
                    </div>
                  </div>
                </>
              )}

              {/* Loading state */}
              {isLoading && !stats && (
                <div className="px-4 lg:px-6">
                  <div className="max-w-7xl">
                    <div className="space-y-4">
                      <div className="animate-pulse bg-gray-200 h-32 rounded-lg"></div>
                      <div className="grid gap-4 lg:grid-cols-2">
                        <div className="animate-pulse bg-gray-200 h-48 rounded-lg"></div>
                        <div className="animate-pulse bg-gray-200 h-48 rounded-lg"></div>
                      </div>
                      <div className="animate-pulse bg-gray-200 h-64 rounded-lg"></div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
