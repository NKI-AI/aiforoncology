// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import {
  Area as RechartsArea,
  AreaChart,
  CartesianGrid,
  XAxis as RechartsXAxis,
  YAxis as RechartsYAxis,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
} from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { SystemStats } from "@/hooks/useSystemMonitor";

// Properly typed wrapper components to fix TypeScript issues
const Area = (props: any) => React.createElement(RechartsArea as any, props);
const XAxis = (props: any) => React.createElement(RechartsXAxis as any, props);
const YAxis = (props: any) => React.createElement(RechartsYAxis as any, props);
const Tooltip = (props: any) =>
  React.createElement(RechartsTooltip as any, props);

interface SystemMonitorChartsProps {
  data: SystemStats[];
  className?: string;
}

const formatTime = (timestamp: number) => {
  if (timestamp == null || isNaN(timestamp)) {
    return new Date().toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  }
  return new Date(timestamp).toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

const formatBytes = (bytes: number, decimals: number = 1): string => {
  if (bytes === 0) return "0 B";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["B", "KB", "MB", "GB", "TB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
};

// CPU Usage Chart
export const CPUChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    pidCpu: stat.pid?.cpu ?? 0,
    osCpu: stat.os?.cpu ?? 0,
  }));

  const chartConfig = {
    pidCpu: {
      label: "Process CPU",
      color: "hsl(var(--chart-1))",
    },
    osCpu: {
      label: "System CPU",
      color: "hsl(var(--chart-2))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>CPU Usage</CardTitle>
        <CardDescription>
          Real-time CPU usage for process and system
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillPidCpu" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-pidCpu)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-pidCpu)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillOsCpu" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-osCpu)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-osCpu)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, 100]}
              tickFormatter={(value) => `${value}%`}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                `${Number(value).toFixed(1)}%`,
                name === "pidCpu" ? "Process CPU" : "System CPU",
              ]}
            />
            <Area
              dataKey="pidCpu"
              type="monotone"
              fill="url(#fillPidCpu)"
              stroke="var(--color-pidCpu)"
              strokeWidth={2}
            />
            <Area
              dataKey="osCpu"
              type="monotone"
              fill="url(#fillOsCpu)"
              stroke="var(--color-osCpu)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// Memory Usage Chart
export const MemoryChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    pidMemory: (stat.pid?.ram ?? 0) / (1024 * 1024), // Convert to MB
    osMemory: (stat.os?.ram ?? 0) / (1024 * 1024), // Convert to MB
    totalMemory: (stat.os?.total_ram ?? 0) / (1024 * 1024), // Convert to MB
  }));

  const chartConfig = {
    pidMemory: {
      label: "Process Memory",
      color: "hsl(var(--chart-1))",
    },
    osMemory: {
      label: "System Memory",
      color: "hsl(var(--chart-2))",
    },
    totalMemory: {
      label: "Total Memory",
      color: "hsl(var(--chart-3))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Memory Usage</CardTitle>
        <CardDescription>
          Real-time memory usage for process and system
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillPidMemory" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-pidMemory)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-pidMemory)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillOsMemory" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-osMemory)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-osMemory)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillTotalMemory" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-totalMemory)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-totalMemory)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => `${Math.round(value)} MB`}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                `${Math.round(Number(value))} MB`,
                name === "pidMemory"
                  ? "Process Memory"
                  : name === "osMemory"
                  ? "System Memory"
                  : "Total Memory",
              ]}
            />
            <Area
              dataKey="totalMemory"
              type="monotone"
              fill="url(#fillTotalMemory)"
              stroke="var(--color-totalMemory)"
              strokeWidth={1}
              fillOpacity={0.2}
            />
            <Area
              dataKey="osMemory"
              type="monotone"
              fill="url(#fillOsMemory)"
              stroke="var(--color-osMemory)"
              strokeWidth={2}
            />
            <Area
              dataKey="pidMemory"
              type="monotone"
              fill="url(#fillPidMemory)"
              stroke="var(--color-pidMemory)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// Connections Chart
export const ConnectionsChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    pidConnections: stat.pid?.conns ?? 0,
    osConnections: stat.os?.conns ?? 0,
  }));

  const chartConfig = {
    pidConnections: {
      label: "Process Connections",
      color: "hsl(var(--chart-1))",
    },
    osConnections: {
      label: "System Connections",
      color: "hsl(var(--chart-2))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Network Connections</CardTitle>
        <CardDescription>
          Active TCP connections for process and system
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient
                id="fillPidConnections"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-pidConnections)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-pidConnections)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient
                id="fillOsConnections"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-osConnections)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-osConnections)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 10"]}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                `${value} connections`,
                name === "pidConnections" ? "Process" : "System",
              ]}
            />
            <Area
              dataKey="pidConnections"
              type="monotone"
              fill="url(#fillPidConnections)"
              stroke="var(--color-pidConnections)"
              strokeWidth={2}
            />
            <Area
              dataKey="osConnections"
              type="monotone"
              fill="url(#fillOsConnections)"
              stroke="var(--color-osConnections)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// System Load Chart
export const LoadChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    loadAvg: stat.os?.load_avg ?? 0,
  }));

  const chartConfig = {
    loadAvg: {
      label: "Load Average",
      color: "hsl(var(--chart-1))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>System Load</CardTitle>
        <CardDescription>1-minute system load average</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillLoadAvg" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-loadAvg)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-loadAvg)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 1"]}
              tickFormatter={(value) => value.toFixed(1)}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value) => [Number(value).toFixed(2), "Load Average"]}
            />
            <Area
              dataKey="loadAvg"
              type="monotone"
              fill="url(#fillLoadAvg)"
              stroke="var(--color-loadAvg)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// HTTP Performance Chart
export const HTTPChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    requestsPerSecond: stat.http?.requests_per_second ?? 0,
    avgResponseTime: stat.http?.avg_response_time_ms ?? 0,
    activeRequests: stat.http?.active_requests ?? 0,
  }));

  const chartConfig = {
    requestsPerSecond: {
      label: "Requests/sec",
      color: "hsl(var(--chart-1))",
    },
    avgResponseTime: {
      label: "Avg Response (ms)",
      color: "hsl(var(--chart-2))",
    },
    activeRequests: {
      label: "Active Requests",
      color: "hsl(var(--chart-3))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>HTTP Performance</CardTitle>
        <CardDescription>
          Real-time HTTP request metrics and response times
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient
                id="fillRequestsPerSecond"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-requestsPerSecond)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-requestsPerSecond)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient
                id="fillActiveRequests"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-activeRequests)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-activeRequests)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 1"]}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                name === "requestsPerSecond"
                  ? `${Number(value).toFixed(1)}/s`
                  : name === "avgResponseTime"
                  ? `${Number(value).toFixed(1)}ms`
                  : `${value} requests`,
                name === "requestsPerSecond"
                  ? "Requests/sec"
                  : name === "avgResponseTime"
                  ? "Avg Response"
                  : "Active Requests",
              ]}
            />
            <Area
              dataKey="requestsPerSecond"
              type="monotone"
              fill="url(#fillRequestsPerSecond)"
              stroke="var(--color-requestsPerSecond)"
              strokeWidth={2}
            />
            <Area
              dataKey="activeRequests"
              type="monotone"
              fill="url(#fillActiveRequests)"
              stroke="var(--color-activeRequests)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// Go Runtime Chart
export const RuntimeChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    goroutines: stat.runtime?.goroutines ?? 0,
    heapAllocMB: (stat.runtime?.heap_alloc ?? 0) / (1024 * 1024),
    gcCycles: stat.runtime?.gc_cycles ?? 0,
  }));

  const chartConfig = {
    goroutines: {
      label: "Goroutines",
      color: "hsl(var(--chart-1))",
    },
    heapAllocMB: {
      label: "Heap Alloc (MB)",
      color: "hsl(var(--chart-2))",
    },
    gcCycles: {
      label: "GC Cycles",
      color: "hsl(var(--chart-3))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Go Runtime</CardTitle>
        <CardDescription>
          Go runtime metrics: goroutines, memory, and garbage collection
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillGoroutines" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-goroutines)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-goroutines)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillHeapAllocMB" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-heapAllocMB)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-heapAllocMB)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 10"]}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                name === "heapAllocMB"
                  ? `${Number(value).toFixed(1)} MB`
                  : `${value}`,
                name === "goroutines"
                  ? "Goroutines"
                  : name === "heapAllocMB"
                  ? "Heap Alloc"
                  : "GC Cycles",
              ]}
            />
            <Area
              dataKey="goroutines"
              type="monotone"
              fill="url(#fillGoroutines)"
              stroke="var(--color-goroutines)"
              strokeWidth={2}
            />
            <Area
              dataKey="heapAllocMB"
              type="monotone"
              fill="url(#fillHeapAllocMB)"
              stroke="var(--color-heapAllocMB)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// WebSocket Activity Chart
export const WebSocketChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    activeConnections: stat.websocket?.active_connections ?? 0,
    messagesSent: stat.websocket?.messages_sent ?? 0,
    messagesReceived: stat.websocket?.messages_received ?? 0,
  }));

  const chartConfig = {
    activeConnections: {
      label: "Active Connections",
      color: "hsl(var(--chart-1))",
    },
    messagesSent: {
      label: "Messages Sent",
      color: "hsl(var(--chart-2))",
    },
    messagesReceived: {
      label: "Messages Received",
      color: "hsl(var(--chart-3))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>WebSocket Activity</CardTitle>
        <CardDescription>
          Real-time WebSocket connections and message flow
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient
                id="fillActiveConnections"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-activeConnections)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-activeConnections)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillMessagesSent" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-messagesSent)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-messagesSent)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient
                id="fillMessagesReceived"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor="var(--color-messagesReceived)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-messagesReceived)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 1"]}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                `${value}`,
                name === "activeConnections"
                  ? "Active Connections"
                  : name === "messagesSent"
                  ? "Messages Sent"
                  : "Messages Received",
              ]}
            />
            <Area
              dataKey="activeConnections"
              type="monotone"
              fill="url(#fillActiveConnections)"
              stroke="var(--color-activeConnections)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};

// I/O Performance Chart
export const IOChart: React.FC<SystemMonitorChartsProps> = ({
  data,
  className,
}) => {
  const chartData = data.map((stat) => ({
    time: formatTime(stat.timestamp),
    timestamp: stat.timestamp,
    diskReadMB: (stat.os?.disk_read_bytes ?? 0) / (1024 * 1024),
    diskWriteMB: (stat.os?.disk_write_bytes ?? 0) / (1024 * 1024),
    netReadMB: (stat.os?.net_read_bytes ?? 0) / (1024 * 1024),
    netWriteMB: (stat.os?.net_write_bytes ?? 0) / (1024 * 1024),
  }));

  const chartConfig = {
    diskReadMB: {
      label: "Disk Read (MB/s)",
      color: "hsl(var(--chart-1))",
    },
    diskWriteMB: {
      label: "Disk Write (MB/s)",
      color: "hsl(var(--chart-2))",
    },
    netReadMB: {
      label: "Net Read (MB/s)",
      color: "hsl(var(--chart-3))",
    },
    netWriteMB: {
      label: "Net Write (MB/s)",
      color: "hsl(var(--chart-4))",
    },
  } satisfies ChartConfig;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>I/O Performance</CardTitle>
        <CardDescription>Disk and network I/O rates in MB/s</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[250px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillDiskReadMB" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-diskReadMB)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-diskReadMB)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillDiskWriteMB" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-diskWriteMB)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-diskWriteMB)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillNetReadMB" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-netReadMB)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-netReadMB)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillNetWriteMB" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-netWriteMB)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-netWriteMB)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="time"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, "dataMax + 1"]}
              tickFormatter={(value) => `${value.toFixed(1)} MB/s`}
            />
            <Tooltip
              labelFormatter={(value) => `Time: ${value}`}
              formatter={(value, name) => [
                `${Number(value).toFixed(2)} MB/s`,
                name === "diskReadMB"
                  ? "Disk Read"
                  : name === "diskWriteMB"
                  ? "Disk Write"
                  : name === "netReadMB"
                  ? "Net Read"
                  : "Net Write",
              ]}
            />
            <Area
              dataKey="diskReadMB"
              type="monotone"
              fill="url(#fillDiskReadMB)"
              stroke="var(--color-diskReadMB)"
              strokeWidth={2}
            />
            <Area
              dataKey="diskWriteMB"
              type="monotone"
              fill="url(#fillDiskWriteMB)"
              stroke="var(--color-diskWriteMB)"
              strokeWidth={2}
            />
            <Area
              dataKey="netReadMB"
              type="monotone"
              fill="url(#fillNetReadMB)"
              stroke="var(--color-netReadMB)"
              strokeWidth={2}
            />
            <Area
              dataKey="netWriteMB"
              type="monotone"
              fill="url(#fillNetWriteMB)"
              stroke="var(--color-netWriteMB)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
};
