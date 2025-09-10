// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useCallback } from "react";
import { useWebSocketManager, WebSocketMessage } from "./useWebSocketManager";
import {
  formatBytes,
  formatNumber,
  formatPercentage,
  formatDuration,
  formatTime,
  formatTimestamp,
} from "@/utils/format";

interface SystemStatsPID {
  cpu: number;
  ram: number;
  conns: number;
  fds: number;
  threads: number;
  uptime: number;
  cpu_time: number;
  context_switches: number;
}

interface SystemStatsOS {
  cpu: number;
  ram: number;
  total_ram: number;
  swap_used: number;
  swap_total: number;
  load_avg: number;
  load_avg_5: number;
  load_avg_15: number;
  conns: number;
  disk_read_bytes: number;
  disk_write_bytes: number;
  net_read_bytes: number;
  net_write_bytes: number;
  uptime: number;
  boot_time: number;
  total_processes: number;
  running_processes: number;
}

interface SystemStatsHTTP {
  requests_per_second: number;
  total_requests: number;
  active_requests: number;
  avg_response_time_ms: number;
  p95_response_time_ms: number;
  p99_response_time_ms: number;
  status_codes: Record<string, number>;
  error_rate: number;
  requests_by_method: Record<string, number>;
  requests_by_endpoint: Record<string, number>;
  last_reset_time: number;
  total_response_time: number;
  slow_requests: number;
}

interface SystemStatsRuntime {
  goroutines: number;
  heap_alloc: number;
  heap_sys: number;
  heap_inuse: number;
  heap_objects: number;
  stack_inuse: number;
  gc_cycles: number;
  gc_pause_total_ns: number;
  next_gc: number;
  version: string;
}

interface SystemStatsWebSocket {
  total_connections: number;
  active_connections: number;
  messages_sent: number;
  messages_received: number;
  connection_errors: number;
  last_connection_time: number;
}

export interface SystemStats {
  pid: SystemStatsPID;
  os: SystemStatsOS;
  http: SystemStatsHTTP;
  runtime: SystemStatsRuntime;
  websocket: SystemStatsWebSocket;
  timestamp: number;
}

interface SystemMonitorOptions {
  maxDataPoints?: number;
  autoConnect?: boolean;
}

export const useSystemMonitor = (options: SystemMonitorOptions = {}) => {
  const { maxDataPoints = 50, autoConnect = true } = options;

  const [stats, setStats] = useState<SystemStats | null>(null);
  const [historicalData, setHistoricalData] = useState<SystemStats[]>([]);

  // WebSocket URL factory
  const createWebSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    return `${protocol}//${host}/api/v1/system/monitor/ws`;
  }, []);

  // Handle incoming WebSocket messages
  const handleMessage = useCallback(
    (message: WebSocketMessage) => {
      // Assume it's system stats data - validate structure
      const newStats: SystemStats = {
        pid: message.pid || {
          cpu: 0,
          ram: 0,
          conns: 0,
          fds: 0,
          threads: 0,
          uptime: 0,
          cpu_time: 0,
          context_switches: 0,
        },
        os: message.os || {
          cpu: 0,
          ram: 0,
          total_ram: 0,
          swap_used: 0,
          swap_total: 0,
          load_avg: 0,
          load_avg_5: 0,
          load_avg_15: 0,
          conns: 0,
          disk_read_bytes: 0,
          disk_write_bytes: 0,
          net_read_bytes: 0,
          net_write_bytes: 0,
          uptime: 0,
          boot_time: 0,
          total_processes: 0,
          running_processes: 0,
        },
        http: message.http || {
          requests_per_second: 0,
          total_requests: 0,
          active_requests: 0,
          avg_response_time_ms: 0,
          p95_response_time_ms: 0,
          p99_response_time_ms: 0,
          status_codes: {},
          error_rate: 0,
          requests_by_method: {},
          requests_by_endpoint: {},
          last_reset_time: 0,
          total_response_time: 0,
          slow_requests: 0,
        },
        runtime: message.runtime || {
          goroutines: 0,
          heap_alloc: 0,
          heap_sys: 0,
          heap_inuse: 0,
          heap_objects: 0,
          stack_inuse: 0,
          gc_cycles: 0,
          gc_pause_total_ns: 0,
          next_gc: 0,
          version: "",
        },
        websocket: message.websocket || {
          total_connections: 0,
          active_connections: 0,
          messages_sent: 0,
          messages_received: 0,
          connection_errors: 0,
          last_connection_time: 0,
        },
        timestamp: message.timestamp || Date.now(),
      };

      setStats(newStats);

      // Add to historical data
      setHistoricalData((prev) => {
        const updated = [...prev, newStats];
        // Keep only the last maxDataPoints
        if (updated.length > maxDataPoints) {
          return updated.slice(-maxDataPoints);
        }
        return updated;
      });
    },
    [maxDataPoints]
  );

  // Use the WebSocket manager
  const {
    connected,
    connecting,
    error,
    reconnectAttempts,
    connect,
    disconnect,
    forceReconnect,
    send,
    getReadyState,
    getReadyStateText,
  } = useWebSocketManager(createWebSocketUrl, handleMessage, {
    autoConnect,
    maxReconnectAttempts: 5,
    useExponentialBackoff: true,
    maxReconnectDelay: 30000,
    enablePing: true,
    pingInterval: 30000,
  });

  const clearHistory = useCallback(() => {
    setHistoricalData([]);
  }, []);

  // Format current stats for display
  const formattedStats = stats
    ? {
        pid: {
          cpu: formatPercentage(stats.pid.cpu),
          ram: formatBytes(stats.pid.ram),
          conns: (stats.pid.conns ?? 0).toString(),
          fds: (stats.pid.fds ?? 0).toString(),
          threads: (stats.pid.threads ?? 0).toString(),
          uptime: formatTime(stats.pid.uptime),
          cpu_time: formatTime(stats.pid.cpu_time),
          context_switches: (stats.pid.context_switches ?? 0).toString(),
        },
        os: {
          cpu: formatPercentage(stats.os.cpu),
          ram: formatBytes(stats.os.ram),
          totalRam: formatBytes(stats.os.total_ram),
          swapUsed: formatBytes(stats.os.swap_used),
          swapTotal: formatBytes(stats.os.swap_total),
          loadAvg: (stats.os.load_avg ?? 0).toFixed(2),
          loadAvg5: (stats.os.load_avg_5 ?? 0).toFixed(2),
          loadAvg15: (stats.os.load_avg_15 ?? 0).toFixed(2),
          conns: (stats.os.conns ?? 0).toString(),
          diskRead: formatBytes(stats.os.disk_read_bytes),
          diskWrite: formatBytes(stats.os.disk_write_bytes),
          netRead: formatBytes(stats.os.net_read_bytes),
          netWrite: formatBytes(stats.os.net_write_bytes),
          uptime: formatTime(stats.os.uptime),
          bootTime: formatTimestamp(stats.os.boot_time),
          totalProcesses: (stats.os.total_processes ?? 0).toString(),
          runningProcesses: (stats.os.running_processes ?? 0).toString(),
        },
        http: {
          requestsPerSecond: (stats.http.requests_per_second ?? 0).toFixed(1),
          totalRequests: formatNumber(stats.http.total_requests),
          activeRequests: (stats.http.active_requests ?? 0).toString(),
          avgResponseTime: (stats.http.avg_response_time_ms ?? 0).toFixed(1),
          p95ResponseTime: (stats.http.p95_response_time_ms ?? 0).toFixed(1),
          p99ResponseTime: (stats.http.p99_response_time_ms ?? 0).toFixed(1),
          errorRate: (stats.http.error_rate ?? 0).toFixed(1),
          statusCodes: stats.http.status_codes,
          requestsByMethod: stats.http.requests_by_method,
          requestsByEndpoint: stats.http.requests_by_endpoint,
          totalResponseTime: `${(stats.http.total_response_time ?? 0).toFixed(
            1
          )}ms`,
          slowRequests: (stats.http.slow_requests ?? 0).toString(),
        },
        runtime: {
          goroutines: formatNumber(stats.runtime.goroutines),
          heapAlloc: formatBytes(stats.runtime.heap_alloc),
          heapSys: formatBytes(stats.runtime.heap_sys),
          heapInuse: formatBytes(stats.runtime.heap_inuse),
          heapObjects: formatNumber(stats.runtime.heap_objects),
          stackInuse: formatBytes(stats.runtime.stack_inuse),
          gcCycles: formatNumber(stats.runtime.gc_cycles),
          gcPauseTotal: formatDuration(stats.runtime.gc_pause_total_ns),
          nextGC: formatBytes(stats.runtime.next_gc),
          version: stats.runtime.version || "Unknown",
        },
        websocket: {
          totalConnections: formatNumber(stats.websocket.total_connections),
          activeConnections: (
            stats.websocket.active_connections ?? 0
          ).toString(),
          messagesSent: formatNumber(stats.websocket.messages_sent),
          messagesReceived: formatNumber(stats.websocket.messages_received),
          connectionErrors: formatNumber(stats.websocket.connection_errors),
          lastConnectionTime:
            stats.websocket.last_connection_time > 0
              ? new Date(
                  stats.websocket.last_connection_time * 1000
                ).toLocaleTimeString()
              : "Never",
        },
        timestamp: new Date(
          (stats.timestamp ?? Date.now()) * 1000
        ).toLocaleTimeString(),
      }
    : null;

  return {
    stats,
    formattedStats,
    historicalData,
    connected,
    connecting,
    error,
    reconnectAttempts,
    connect,
    disconnect,
    forceReconnect,
    clearHistory,
    getReadyState,
    getReadyStateText,
    // Re-export formatting utilities for convenience
    formatBytes,
    formatNumber,
    formatPercentage,
    formatDuration,
    formatTime,
    formatTimestamp,
  };
};
