// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { regionKeys } from "./useRegionQueries";

// Types for sync queue operations
export type SyncOperation =
  | { type: "create"; localId: string; data: any; timestamp: number }
  | { type: "update"; regionId: string; data: any; timestamp: number }
  | { type: "delete"; regionId: string; data?: any; timestamp: number };

export interface QueuedOperation {
  id: string;
  operation: SyncOperation;
  retryCount: number;
  lastAttempt: number;
  status: "pending" | "syncing" | "failed" | "completed";
  error?: string;
}

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000; // Base delay in ms
const MAX_RETRY_DELAY = 30000; // Max delay in ms

// Mutation function types
export interface RegionMutationFunctions {
  createRegion: (data: {
    localId: string;
    data: any;
    timestamp: number;
  }) => Promise<any>;
  updateRegion: (data: {
    regionId: string;
    data: any;
    timestamp: number;
  }) => Promise<any>;
  deleteRegion: (data: { regionId: string; timestamp: number }) => Promise<any>;
}

/**
 * Advanced sync queue with conflict resolution and offline support
 */
export function useRegionSyncQueue(
  slideUid: string,
  mutations: RegionMutationFunctions
) {
  const queryClient = useQueryClient();
  const [queue, setQueue] = useState<QueuedOperation[]>([]);
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const syncTimeoutRef = useRef<NodeJS.Timeout>();
  const queueRef = useRef<QueuedOperation[]>([]);

  // Keep queue ref in sync
  useEffect(() => {
    queueRef.current = queue;
  }, [queue]);

  // Monitor online status
  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      processQueue(); // Process queue when coming back online
    };

    const handleOffline = () => {
      setIsOnline(false);
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  // Load queue from localStorage on mount
  useEffect(() => {
    const savedQueue = localStorage.getItem(`regionQueue:${slideUid}`);
    if (savedQueue) {
      try {
        const parsedQueue = JSON.parse(savedQueue);
        setQueue(parsedQueue);
      } catch (error) {
        console.error("Failed to load sync queue from localStorage:", error);
      }
    }
  }, [slideUid]);

  // Save queue to localStorage whenever it changes
  useEffect(() => {
    if (queue.length > 0) {
      localStorage.setItem(`regionQueue:${slideUid}`, JSON.stringify(queue));
    } else {
      localStorage.removeItem(`regionQueue:${slideUid}`);
    }
  }, [queue, slideUid]);

  // Add operation to queue
  const enqueueOperation = useCallback((operation: SyncOperation) => {
    const queuedOp: QueuedOperation = {
      id: `${operation.type}-${Date.now()}-${Math.random()}`,
      operation,
      retryCount: 0,
      lastAttempt: 0,
      status: "pending",
    };

    setQueue((prev) => {
      // Check for conflicting operations and resolve them
      const filtered = resolveConflicts([...prev, queuedOp]);
      return filtered;
    });

    // Process queue after adding operation
    setTimeout(processQueue, 100);
  }, []);

  // Resolve conflicts in the queue (e.g., delete after create, multiple updates)
  const resolveConflicts = useCallback(
    (operations: QueuedOperation[]): QueuedOperation[] => {
      const resolved: QueuedOperation[] = [];
      const operationsByTarget = new Map<string, QueuedOperation[]>();

      // Group operations by target (localId or regionId)
      operations.forEach((op) => {
        const target =
          op.operation.type === "create"
            ? op.operation.localId
            : op.operation.regionId;

        if (!operationsByTarget.has(target)) {
          operationsByTarget.set(target, []);
        }
        operationsByTarget.get(target)!.push(op);
      });

      // Resolve conflicts for each target
      operationsByTarget.forEach((ops, target) => {
        // Sort by timestamp
        ops.sort((a, b) => a.operation.timestamp - b.operation.timestamp);

        // Apply conflict resolution rules
        const hasDelete = ops.some((op) => op.operation.type === "delete");

        if (hasDelete) {
          // If there's a delete, only keep the delete operation (latest one)
          const deleteOp = ops
            .filter((op) => op.operation.type === "delete")
            .pop();
          if (deleteOp) {
            resolved.push(deleteOp);
          }
        } else {
          // Keep create (if any) and latest update
          const createOp = ops.find((op) => op.operation.type === "create");
          const updateOps = ops.filter((op) => op.operation.type === "update");

          if (createOp) {
            // Merge create with updates
            const mergedData = updateOps.reduce(
              (acc, op) => ({
                ...acc,
                ...op.operation.data,
              }),
              createOp.operation.data
            );

            resolved.push({
              ...createOp,
              operation: {
                ...createOp.operation,
                data: mergedData,
                timestamp: Math.max(...ops.map((op) => op.operation.timestamp)),
              },
            });
          } else if (updateOps.length > 0) {
            // Keep only the latest update with merged data
            const latestUpdate = updateOps[updateOps.length - 1];
            const mergedData = updateOps.reduce(
              (acc, op) => ({
                ...acc,
                ...op.operation.data,
              }),
              {}
            );

            resolved.push({
              ...latestUpdate,
              operation: {
                ...latestUpdate.operation,
                data: mergedData,
              },
            });
          }
        }
      });

      return resolved;
    },
    []
  );

  // Calculate exponential backoff delay
  const getRetryDelay = useCallback((retryCount: number): number => {
    const delay = RETRY_DELAY * Math.pow(2, retryCount);
    return Math.min(delay, MAX_RETRY_DELAY);
  }, []);

  // Process a single operation
  const processOperation = useCallback(
    async (queuedOp: QueuedOperation): Promise<boolean> => {
      try {
        setQueue((prev) =>
          prev.map((op) =>
            op.id === queuedOp.id ? { ...op, status: "syncing" } : op
          )
        );

        const { operation } = queuedOp;

        switch (operation.type) {
          case "create":
            // Execute create mutation using hook
            await mutations.createRegion({
              localId: operation.localId,
              data: operation.data,
              timestamp: operation.timestamp,
            });
            break;
          case "update":
            // Execute update mutation using hook
            await mutations.updateRegion({
              regionId: operation.regionId,
              data: operation.data,
              timestamp: operation.timestamp,
            });
            break;
          case "delete":
            // Execute delete mutation using hook
            await mutations.deleteRegion({
              regionId: operation.regionId,
              timestamp: operation.timestamp,
            });
            break;
        }

        // Mark as completed and remove from queue
        setQueue((prev) => prev.filter((op) => op.id !== queuedOp.id));
        return true;
      } catch (error) {
        console.error(`Failed to process operation ${queuedOp.id}:`, error);

        // Update operation with error and retry info
        setQueue((prev) =>
          prev.map((op) =>
            op.id === queuedOp.id
              ? {
                  ...op,
                  status: "failed",
                  retryCount: op.retryCount + 1,
                  lastAttempt: Date.now(),
                  error:
                    error instanceof Error ? error.message : "Unknown error",
                }
              : op
          )
        );

        return false;
      }
    },
    [mutations]
  );

  // Process the entire queue
  const processQueue = useCallback(async () => {
    if (!isOnline || queueRef.current.length === 0) {
      return;
    }

    // Clear any existing timeout
    if (syncTimeoutRef.current) {
      clearTimeout(syncTimeoutRef.current);
    }

    const now = Date.now();
    const processableOps = queueRef.current.filter((op) => {
      if (op.status === "syncing" || op.status === "completed") {
        return false;
      }

      if (op.status === "failed") {
        if (op.retryCount >= MAX_RETRIES) {
          return false; // Max retries reached
        }

        const retryDelay = getRetryDelay(op.retryCount);
        return now - op.lastAttempt >= retryDelay;
      }

      return true; // Pending operations
    });

    // Process operations sequentially to maintain order
    for (const op of processableOps) {
      await processOperation(op);

      // Small delay between operations to prevent overwhelming the server
      await new Promise((resolve) => setTimeout(resolve, 100));
    }

    // Schedule next processing if there are still pending operations
    const remainingOps = queueRef.current.filter(
      (op) =>
        op.status === "pending" ||
        (op.status === "failed" && op.retryCount < MAX_RETRIES)
    );

    if (remainingOps.length > 0) {
      const nextRetryTime = Math.min(
        ...remainingOps.map((op) => {
          if (op.status === "failed") {
            return op.lastAttempt + getRetryDelay(op.retryCount);
          }
          return Date.now();
        })
      );

      const delay = Math.max(0, nextRetryTime - Date.now());
      syncTimeoutRef.current = setTimeout(processQueue, delay);
    }
  }, [isOnline, processOperation, getRetryDelay]);

  // Manual retry for failed operations
  const retryFailedOperations = useCallback(() => {
    setQueue((prev) =>
      prev.map((op) =>
        op.status === "failed"
          ? { ...op, status: "pending", error: undefined }
          : op
      )
    );
    processQueue();
  }, [processQueue]);

  // Clear completed operations older than 1 hour
  const cleanupQueue = useCallback(() => {
    const oneHourAgo = Date.now() - 60 * 60 * 1000;
    setQueue((prev) =>
      prev.filter(
        (op) => op.status !== "completed" || op.operation.timestamp > oneHourAgo
      )
    );
  }, []);

  // Cleanup effect
  useEffect(() => {
    const interval = setInterval(cleanupQueue, 5 * 60 * 1000); // Every 5 minutes
    return () => clearInterval(interval);
  }, [cleanupQueue]);

  // Auto-process queue when coming online or when queue changes
  useEffect(() => {
    if (isOnline && queue.length > 0) {
      processQueue();
    }
  }, [isOnline, queue.length, processQueue]);

  return {
    queue,
    isOnline,
    enqueueOperation,
    processQueue,
    retryFailedOperations,
    cleanupQueue,

    // Queue statistics
    pendingCount: queue.filter((op) => op.status === "pending").length,
    failedCount: queue.filter((op) => op.status === "failed").length,
    syncingCount: queue.filter((op) => op.status === "syncing").length,
  };
}
