// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useRef, useCallback, useEffect } from "react";

export interface WebSocketMessage {
  type: string;
  timestamp?: number;
  [key: string]: any;
}

interface WebSocketManagerOptions {
  /**
   * Whether to automatically connect when the hook is initialized
   */
  autoConnect?: boolean;

  /**
   * Maximum number of reconnection attempts (0 = unlimited)
   */
  maxReconnectAttempts?: number;

  /**
   * Base delay for reconnection in milliseconds
   */
  reconnectDelay?: number;

  /**
   * Whether to use exponential backoff for reconnection delays
   */
  useExponentialBackoff?: boolean;

  /**
   * Maximum delay for reconnection in milliseconds (when using exponential backoff)
   */
  maxReconnectDelay?: number;

  /**
   * Whether to automatically send ping messages to keep connection alive
   */
  enablePing?: boolean;

  /**
   * Interval for sending ping messages in milliseconds
   */
  pingInterval?: number;
}

interface WebSocketManagerState {
  connected: boolean;
  connecting: boolean;
  error: string | null;
  reconnectAttempts: number;
}

interface WebSocketManagerActions {
  connect: () => void;
  disconnect: () => void;
  forceReconnect: () => void;
  send: (message: WebSocketMessage) => boolean;
  getReadyState: () => number | null;
  getReadyStateText: () => string;
}

type WebSocketMessageHandler = (
  message: WebSocketMessage,
  rawEvent: MessageEvent
) => void;
type WebSocketEventHandler = (event: Event) => void;

export const useWebSocketManager = (
  urlFactory: () => string,
  onMessage: WebSocketMessageHandler,
  options: WebSocketManagerOptions = {}
): WebSocketManagerState & WebSocketManagerActions => {
  const {
    autoConnect = true,
    maxReconnectAttempts = 5,
    reconnectDelay = 1000,
    useExponentialBackoff = true,
    maxReconnectDelay = 30000,
    enablePing = true,
    pingInterval = 30000,
  } = options;

  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const clearTimeouts = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
  }, []);

  const startPingInterval = useCallback(() => {
    if (!enablePing || pingIntervalRef.current) return;

    pingIntervalRef.current = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        try {
          wsRef.current.send(
            JSON.stringify({
              type: "ping",
              timestamp: Date.now(),
            })
          );
        } catch (err) {
          console.error("❌ Failed to send ping:", err);
        }
      }
    }, pingInterval);
  }, [enablePing, pingInterval]);

  const stopPingInterval = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
  }, []);

  const scheduleReconnect = useCallback(() => {
    if (maxReconnectAttempts > 0 && reconnectAttempts >= maxReconnectAttempts) {
      console.error(
        `❌ Max reconnection attempts (${maxReconnectAttempts}) reached`
      );
      return;
    }

    let delay = reconnectDelay;
    if (useExponentialBackoff) {
      delay = Math.min(
        reconnectDelay * Math.pow(2, reconnectAttempts),
        maxReconnectDelay
      );
    }

    console.log(
      `🔄 Scheduling reconnect in ${delay}ms (attempt ${reconnectAttempts + 1})`
    );

    reconnectTimeoutRef.current = setTimeout(() => {
      setReconnectAttempts((prev) => prev + 1);
      connect();
    }, delay);
  }, [
    reconnectAttempts,
    maxReconnectAttempts,
    reconnectDelay,
    useExponentialBackoff,
    maxReconnectDelay,
  ]);

  const connect = useCallback(() => {
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
      return;
    }

    clearTimeouts();
    setConnecting(true);
    setError(null);

    try {
      const url = urlFactory();

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        setConnecting(false);
        setError(null);
        setReconnectAttempts(0);
        startPingInterval();
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          // Handle pong messages internally
          if (data.type === "pong") {
            console.log("🏓 Received pong");
            return;
          }

          // Pass other messages to the provided handler
          onMessage(data, event);
        } catch (err) {
          console.error("❌ Error parsing WebSocket message:", err);
          setError("Failed to parse message");
        }
      };

      ws.onclose = (event) => {
        console.log("🔌 WebSocket closed:", {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        });

        setConnected(false);
        setConnecting(false);
        stopPingInterval();
        wsRef.current = null;

        // Only attempt to reconnect if autoConnect is enabled and it wasn't a clean close
        if (autoConnect && (!event.wasClean || event.code !== 1000)) {
          scheduleReconnect();
        }
      };

      ws.onerror = (error) => {
        console.error("❌ WebSocket error:", error);
        setError("WebSocket connection error");
        setConnected(false);
        setConnecting(false);
        stopPingInterval();
      };
    } catch (err) {
      console.error("❌ Failed to create WebSocket:", err);
      setError("Failed to create WebSocket connection");
      setConnected(false);
      setConnecting(false);
    }
  }, [
    urlFactory,
    onMessage,
    autoConnect,
    scheduleReconnect,
    clearTimeouts,
    startPingInterval,
    stopPingInterval,
  ]);

  const disconnect = useCallback(() => {
    clearTimeouts();
    stopPingInterval();

    if (wsRef.current) {
      // Use close code 1000 (normal closure) to prevent automatic reconnection
      wsRef.current.close(1000, "Manual disconnect");
      wsRef.current = null;
    }

    setConnected(false);
    setConnecting(false);
    setReconnectAttempts(0);
  }, [clearTimeouts, stopPingInterval]);

  const forceReconnect = useCallback(() => {
    console.warn("🔄 Forcing WebSocket reconnect...");
    disconnect();
    setTimeout(() => {
      setReconnectAttempts(0);
      connect();
    }, 1000);
  }, [disconnect, connect]);

  const send = useCallback((message: WebSocketMessage): boolean => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      try {
        wsRef.current.send(JSON.stringify(message));
        return true;
      } catch (err) {
        console.error("❌ Failed to send WebSocket message:", err);
        setError("Failed to send message");
        return false;
      }
    }
    console.warn("⚠️ WebSocket not connected, cannot send message");
    return false;
  }, []);

  const getReadyState = useCallback((): number | null => {
    return wsRef.current?.readyState ?? null;
  }, []);

  const getReadyStateText = useCallback((): string => {
    const state = wsRef.current?.readyState;
    if (state === undefined || state === null) return "NO_CONNECTION";
    return ["CONNECTING", "OPEN", "CLOSING", "CLOSED"][state] || "UNKNOWN";
  }, []);

  // Auto-connect on mount if enabled
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      clearTimeouts();
      stopPingInterval();
    };
  }, [clearTimeouts, stopPingInterval]);

  return {
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
  };
};
