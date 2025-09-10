// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useEffect } from "react";
import { useSystemMonitor } from "@/hooks/useSystemMonitor";
import AdminSidebar from "../AdminSidebar";
import AdminHeader from "../AdminHeader";
import { ErrorStateAlert } from "@/components/ErrorStateAlert";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { StatCard } from "@/components/ui/stat-card";
import {
  CPUChart,
  MemoryChart,
  ConnectionsChart,
  LoadChart,
  HTTPChart,
  RuntimeChart,
  WebSocketChart,
  IOChart,
} from "./SystemMonitorCharts";
import {
  Activity,
  Wifi,
  WifiOff,
  RefreshCw,
  Trash2,
  Globe,
  Zap,
  MessageSquare,
  HardDrive,
  Cpu,
  MemoryStick,
} from "lucide-react";

export default function SystemMonitorPage() {
  // Set document title
  useEffect(() => {
    document.title = "SlideInsight - System Monitor";
    return () => {
      document.title = "SlideInsight Viewer";
    };
  }, []);

  const {
    stats,
    formattedStats,
    historicalData,
    connected,
    error,
    connect,
    disconnect,
    clearHistory,
  } = useSystemMonitor();

  const handleRefresh = () => {
    if (connected) {
      disconnect();
      setTimeout(() => connect(), 500);
    } else {
      connect();
    }
  };

  const connectionStatus = connected ? (
    <Badge variant="default" className="flex items-center gap-1">
      <Wifi className="h-3 w-3" />
      Connected
    </Badge>
  ) : (
    <Badge variant="destructive" className="flex items-center gap-1">
      <WifiOff className="h-3 w-3" />
      Disconnected
    </Badge>
  );

  const headerActions = (
    <div className="flex items-center gap-2">
      {connectionStatus}
      <Button
        onClick={clearHistory}
        variant="outline"
        size="sm"
        disabled={historicalData.length === 0}
      >
        <Trash2 className="h-4 w-4 mr-2" />
        Clear History
      </Button>
      <Button
        onClick={handleRefresh}
        variant="outline"
        size="sm"
        disabled={false}
      >
        <RefreshCw className="h-4 w-4 mr-2" />
        {connected ? "Reconnect" : "Connect"}
      </Button>
    </div>
  );

  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title="System Monitor"
          description="Real-time comprehensive system performance monitoring with enhanced metrics"
          actions={headerActions}
        />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {/* Error state */}
              {error && (
                <div className="px-4 lg:px-6">
                  <ErrorStateAlert
                    error={error}
                    title="Failed to connect to monitoring service"
                    onRetry={handleRefresh}
                    variant="inline"
                  />
                </div>
              )}

              {/* Current Stats Cards */}
              <div className="px-4 lg:px-6">
                <div className="max-w-7xl">
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
                    <StatCard
                      title="System CPU"
                      value={formattedStats?.os.cpu || "0.0%"}
                      subtitle={`Process: ${formattedStats?.pid.cpu || "0.0%"}`}
                      icon={Cpu}
                    />

                    <StatCard
                      title="Memory"
                      value={formattedStats?.pid.ram || "0 B"}
                      subtitle={`System: ${formattedStats?.os.ram || "0 B"}`}
                      icon={MemoryStick}
                    />

                    <StatCard
                      title="HTTP Requests"
                      value={`${
                        formattedStats?.http.requestsPerSecond || "0"
                      }/s`}
                      subtitle={`Active: ${
                        formattedStats?.http.activeRequests || "0"
                      }`}
                      icon={Globe}
                    />

                    <StatCard
                      title="Goroutines"
                      value={formattedStats?.runtime.goroutines || "0"}
                      subtitle={`GC: ${
                        formattedStats?.runtime.gcCycles || "0"
                      }`}
                      icon={Zap}
                    />

                    <StatCard
                      title="WebSockets"
                      value={formattedStats?.websocket.activeConnections || "0"}
                      subtitle="Active connections"
                      icon={MessageSquare}
                    />

                    <StatCard
                      title="Load Average"
                      value={formattedStats?.os.loadAvg || "0.00"}
                      subtitle="1-minute average"
                      icon={Activity}
                    />
                  </div>

                  {/* Additional System Information */}
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-6 mt-4">
                    <StatCard
                      title="System Uptime"
                      value={formattedStats?.os.uptime || "0s"}
                      subtitle={`Process: ${
                        formattedStats?.pid.uptime || "0s"
                      }`}
                      icon={Activity}
                    />

                    <StatCard
                      title="Total Processes"
                      value={formattedStats?.os.totalProcesses || "0"}
                      subtitle={`Running: ${
                        formattedStats?.os.runningProcesses || "0"
                      }`}
                      icon={Zap}
                    />

                    <StatCard
                      title="Swap Usage"
                      value={formattedStats?.os.swapUsed || "0 B"}
                      subtitle={`Total: ${
                        formattedStats?.os.swapTotal || "0 B"
                      }`}
                      icon={HardDrive}
                    />

                    <StatCard
                      title="File Descriptors"
                      value={formattedStats?.pid.fds || "0"}
                      subtitle={`Threads: ${
                        formattedStats?.pid.threads || "0"
                      }`}
                      icon={HardDrive}
                    />

                    <StatCard
                      title="HTTP Response P99"
                      value={`${formattedStats?.http.p99ResponseTime || "0"}ms`}
                      subtitle={`Slow: ${
                        formattedStats?.http.slowRequests || "0"
                      }`}
                      icon={Globe}
                    />

                    <StatCard
                      title="Load Avg (5/15m)"
                      value={`${formattedStats?.os.loadAvg5 || "0.00"}`}
                      subtitle={`15m: ${
                        formattedStats?.os.loadAvg15 || "0.00"
                      }`}
                      icon={Activity}
                    />
                  </div>
                </div>
              </div>

              {/* Enhanced Details Cards */}
              <div className="px-4 lg:px-6">
                <div className="max-w-7xl">
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    {/* HTTP Details */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <Globe className="h-4 w-4" />
                          HTTP Performance
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Avg Response:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.avgResponseTime || "0"}ms
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            P95 Response:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.p95ResponseTime || "0"}ms
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            P99 Response:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.p99ResponseTime || "0"}ms
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Error Rate:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.errorRate || "0"}%
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Slow Requests:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.slowRequests || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Total Requests:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.http.totalRequests || "0"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Runtime Details */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <Zap className="h-4 w-4" />
                          Go Runtime
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Heap Alloc:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.runtime.heapAlloc || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Heap Objects:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.runtime.heapObjects || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            GC Pause:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.runtime.gcPauseTotal || "0ms"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Go Version:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.runtime.version || "N/A"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>

                    {/* WebSocket Details */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <MessageSquare className="h-4 w-4" />
                          WebSocket Activity
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Total Connections:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.websocket.totalConnections || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Messages Sent:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.websocket.messagesSent || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Messages Received:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.websocket.messagesReceived || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Errors:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.websocket.connectionErrors || "0"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>

                    {/* I/O Performance */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <HardDrive className="h-4 w-4" />
                          I/O Performance
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Disk Read:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.diskRead || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Disk Write:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.diskWrite || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Net Read:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.netRead || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Net Write:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.netWrite || "0 B"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </div>
              </div>

              {/* Additional System Details */}
              <div className="px-4 lg:px-6">
                <div className="max-w-7xl">
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {/* System Information */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <Activity className="h-4 w-4" />
                          System Information
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            System Uptime:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.uptime || "0s"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Process Uptime:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.pid.uptime || "0s"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Total Processes:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.totalProcesses || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Running Processes:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.runningProcesses || "0"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Process Details */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <Cpu className="h-4 w-4" />
                          Process Details
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            CPU Time:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.pid.cpu_time || "0s"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Context Switches:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.pid.context_switches || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            File Descriptors:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.pid.fds || "0"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Threads:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.pid.threads || "0"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Memory & Swap */}
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base flex items-center gap-2">
                          <MemoryStick className="h-4 w-4" />
                          Memory & Swap
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Total RAM:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.totalRam || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Used RAM:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.ram || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Swap Used:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.swapUsed || "0 B"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">
                            Swap Total:
                          </span>
                          <span className="text-sm font-medium">
                            {formattedStats?.os.swapTotal || "0 B"}
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </div>
              </div>

              {/* Charts */}
              {historicalData.length > 0 ? (
                <>
                  {/* System Charts */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl grid gap-6 lg:grid-cols-2">
                      <CPUChart data={historicalData} />
                      <MemoryChart data={historicalData} />
                    </div>
                  </div>

                  {/* Performance Charts */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl grid gap-6 lg:grid-cols-2">
                      <HTTPChart data={historicalData} />
                      <RuntimeChart data={historicalData} />
                    </div>
                  </div>

                  {/* Activity Charts */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl grid gap-6 lg:grid-cols-2">
                      <WebSocketChart data={historicalData} />
                      <IOChart data={historicalData} />
                    </div>
                  </div>

                  {/* Additional Charts */}
                  <div className="px-4 lg:px-6">
                    <div className="max-w-7xl grid gap-6 lg:grid-cols-2">
                      <ConnectionsChart data={historicalData} />
                      <LoadChart data={historicalData} />
                    </div>
                  </div>
                </>
              ) : (
                <div className="px-4 lg:px-6">
                  <div className="max-w-7xl">
                    <Card>
                      <CardContent className="p-6">
                        <div className="text-center">
                          <Activity className="mx-auto h-12 w-12 text-muted-foreground" />
                          <h3 className="mt-2 text-sm font-semibold text-muted-900">
                            No monitoring data yet
                          </h3>
                          <p className="mt-1 text-sm text-muted-foreground">
                            {connected
                              ? "Waiting for comprehensive system data..."
                              : "Connect to start monitoring system performance with enhanced metrics."}
                          </p>
                          {!connected && (
                            <div className="mt-6">
                              <Button
                                onClick={connect}
                                className="inline-flex items-center"
                              >
                                <Wifi className="h-4 w-4 mr-2" />
                                Connect
                              </Button>
                            </div>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </div>
              )}

              {/* Connection Status */}
              <div className="px-4 lg:px-6">
                <div className="max-w-7xl">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">
                        Connection Status & Statistics
                      </CardTitle>
                      <CardDescription>
                        Enhanced monitoring WebSocket connection status and
                        comprehensive metrics
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="grid gap-4 md:grid-cols-4">
                        <div>
                          <div className="text-sm font-medium text-muted-foreground">
                            Status
                          </div>
                          <div className="mt-1">{connectionStatus}</div>
                        </div>
                        <div>
                          <div className="text-sm font-medium text-muted-foreground">
                            Data Points
                          </div>
                          <div className="mt-1 text-sm">
                            {historicalData.length} / 50
                          </div>
                        </div>
                        <div>
                          <div className="text-sm font-medium text-muted-foreground">
                            Update Frequency
                          </div>
                          <div className="mt-1 text-sm">Every 3 seconds</div>
                        </div>
                        <div>
                          <div className="text-sm font-medium text-muted-foreground">
                            Last Update
                          </div>
                          <div className="mt-1 text-sm">
                            {formattedStats?.timestamp || "Never"}
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
