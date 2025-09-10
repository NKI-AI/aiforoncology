// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useCallback, useEffect, useState, useRef } from "react";

interface GlobalSyncState {
  isSyncing: boolean;
  lastSyncTime: Date | null;
  error: string | null;
}

interface SyncableComponent {
  manualSync: () => Promise<any>;
  isSyncing: boolean;
  lastSyncTime: Date | null;
  error: string | null;
}

// Global sync coordinator that manages multiple sync sources
export function useGlobalSync() {
  const [syncState, setSyncState] = useState<GlobalSyncState>({
    isSyncing: false,
    lastSyncTime: null,
    error: null,
  });

  // Registry of syncable components
  const syncSourcesRef = useRef<Map<string, SyncableComponent>>(new Map());

  // Register a sync source (e.g., annotations, regions)
  const registerSyncSource = useCallback(
    (id: string, source: SyncableComponent) => {
      syncSourcesRef.current.set(id, source);

      // Update global state based on all sources
      updateGlobalSyncState();
    },
    []
  );

  // Unregister a sync source
  const unregisterSyncSource = useCallback((id: string) => {
    syncSourcesRef.current.delete(id);
    updateGlobalSyncState();
  }, []);

  // Update global sync state based on all registered sources
  const updateGlobalSyncState = useCallback(() => {
    const sources = Array.from(syncSourcesRef.current.values());

    const isSyncing = sources.some((source) => source.isSyncing);
    const errors = sources.map((source) => source.error).filter(Boolean);
    const lastSyncTimes = sources
      .map((source) => source.lastSyncTime)
      .filter(Boolean);

    const lastSyncTime =
      lastSyncTimes.length > 0
        ? new Date(Math.max(...lastSyncTimes.map((date) => date!.getTime())))
        : null;

    setSyncState({
      isSyncing,
      lastSyncTime,
      error: errors.length > 0 ? errors.join("; ") : null,
    });
  }, []);

  // Manual sync all registered sources
  const manualSyncAll = useCallback(async () => {
    const sources = Array.from(syncSourcesRef.current.values());

    try {
      console.log(`Starting global sync for ${sources.length} sources...`);

      // Sync all sources in parallel
      await Promise.all(
        sources.map(async (source, index) => {
          try {
            await source.manualSync();
            console.log(`Sync source ${index + 1} completed successfully`);
          } catch (error) {
            console.error(`Sync source ${index + 1} failed:`, error);
            throw error;
          }
        })
      );

      console.log("Global sync completed successfully");
    } catch (error) {
      console.error("Global sync failed:", error);
      throw error;
    }
  }, []);

  // Listen for sync state changes from registered sources
  useEffect(() => {
    const interval = setInterval(updateGlobalSyncState, 1000);
    return () => clearInterval(interval);
  }, [updateGlobalSyncState]);

  // Listen for global sync events
  useEffect(() => {
    const handleGlobalSync = () => {
      manualSyncAll().catch(console.error);
    };

    window.addEventListener("globalSync", handleGlobalSync);

    return () => {
      window.removeEventListener("globalSync", handleGlobalSync);
    };
  }, [manualSyncAll]);

  return {
    syncState,
    registerSyncSource,
    unregisterSyncSource,
    manualSyncAll,
    isSyncing: syncState.isSyncing,
    lastSyncTime: syncState.lastSyncTime,
    error: syncState.error,
  };
}
