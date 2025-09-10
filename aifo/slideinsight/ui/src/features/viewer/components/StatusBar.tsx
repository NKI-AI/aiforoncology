// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useMemo, useState, useRef, useEffect } from "react";
import { useNavigate, Link } from "@tanstack/react-router";
import { useAuth } from "@/auth";
import { useNotifications } from "@/hooks/useNotifications";
import { UserIcon, AdminIcon, LogoutIcon } from "@/components/icons";
import { onJXLLoadingStateChange } from "@/features/viewer/components/map/JXLLoader";

interface SyncStatus {
  isSyncing: boolean;
  lastSyncTime: Date | null;
  error: string | null;
  onManualSync: () => Promise<void>;
}

interface StatusBarProps {
  tileProgress: {
    inFlight: number;
    loaded: number;
    errors: number;
    started: number;
  };
  isRefreshing: boolean;
  syncStatus?: SyncStatus;
}

function WebSocketBadge() {
  // Reuse notifications websocket as general indicator
  const { connected, connecting, reconnectAttempts } = useNotifications();
  const color = connecting
    ? "bg-amber-500"
    : connected
    ? "bg-green-500"
    : "bg-red-500";
  const label = connecting
    ? "Connecting"
    : connected
    ? "Connected"
    : "Disconnected";
  return (
    <span className="inline-flex items-center gap-1">
      <span className={`inline-block w-1.5 h-1.5 rounded-full ${color}`} />
      <span>
        {label}
        {!connected && reconnectAttempts ? ` (${reconnectAttempts})` : ""}
      </span>
    </span>
  );
}

function SyncStatusBadge({ syncStatus }: { syncStatus: SyncStatus }) {
  const [isManualSyncing, setIsManualSyncing] = useState(false);

  const handleManualSync = async () => {
    if (isManualSyncing || syncStatus.isSyncing) return;

    setIsManualSyncing(true);
    try {
      await syncStatus.onManualSync();
    } catch (error) {
      console.error("Manual sync failed:", error);
    } finally {
      setIsManualSyncing(false);
    }
  };

  const formatLastSync = (date: Date | null) => {
    if (!date) return "Never";
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);

    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return date.toLocaleDateString();
  };

  const color = syncStatus.error
    ? "text-red-400"
    : syncStatus.isSyncing || isManualSyncing
    ? "text-amber-400"
    : "text-green-400";

  const icon =
    syncStatus.isSyncing || isManualSyncing
      ? "⟳"
      : syncStatus.error
      ? "⚠"
      : "✓";

  return (
    <button
      onClick={handleManualSync}
      disabled={syncStatus.isSyncing || isManualSyncing}
      className={`inline-flex items-center gap-1 px-1 py-0.5 rounded hover:bg-muted transition-colors ${color} disabled:opacity-50`}
      title={`Last sync: ${formatLastSync(syncStatus.lastSyncTime)}${
        syncStatus.error ? `\nError: ${syncStatus.error}` : ""
      }\nClick to sync now`}
    >
      <span
        className={`inline-block ${
          syncStatus.isSyncing || isManualSyncing ? "animate-spin" : ""
        }`}
      >
        {icon}
      </span>
      <span className="text-xs">
        {syncStatus.isSyncing || isManualSyncing ? "Syncing" : "Sync"}
      </span>
    </button>
  );
}

export function StatusBar({
  tileProgress,
  isRefreshing,
  syncStatus,
}: StatusBarProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const [isJXLLoading, setIsJXLLoading] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const isAdmin =
    user?.roles?.includes("superadmin") ||
    user?.roles?.some((role) => role.includes("platform.admin"));

  useEffect(() => {
    const onDocClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  // Track JXL loading state
  useEffect(() => {
    const cleanup = onJXLLoadingStateChange(setIsJXLLoading);
    return cleanup;
  }, []);

  const handleLogout = async () => {
    await logout();
    navigate({ to: "/login" });
    setIsUserMenuOpen(false);
  };

  const percent = useMemo(() => {
    if (tileProgress.started === 0) return 0;
    const done = tileProgress.loaded + tileProgress.errors;
    return Math.min(100, Math.round((done / tileProgress.started) * 100));
  }, [tileProgress]);

  const tileStatusText = useMemo(() => {
    if (isJXLLoading) {
      return "Decoder";
    }
    if (tileProgress.inFlight === 0) {
      return "Ready";
    }
    return `${tileProgress.inFlight} loading`;
  }, [tileProgress.inFlight, isJXLLoading]);

  return (
    <div className="h-5 md:h-5 w-full bg-background/80 backdrop-blur-sm border-t border-border text-muted-foreground flex items-center px-2 md:px-3 text-xs">
      <div className="flex items-center gap-3 flex-1 min-w-0">
        {user ? (
          <div className="relative" ref={menuRef}>
            <button
              type="button"
              onClick={() => setIsUserMenuOpen((v) => !v)}
              className="inline-flex max-w-[8rem] md:max-w-[14rem] items-center gap-1 truncate rounded px-1 md:px-1.5 py-0.5 hover:bg-muted transition-colors"
            >
              <span className="truncate">{user.email}</span>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="currentColor"
                className="h-3 w-3"
              >
                <path d="M12 15l-5-5h10l-5 5z" />
              </svg>
            </button>
            {isUserMenuOpen && (
              <div className="absolute bottom-full left-0 mb-1 w-48 rounded-lg border border-border bg-background text-foreground shadow-lg ring-1 ring-black/5 dark:ring-white/10 z-50">
                <div className="px-3 py-2 border-b border-border flex items-center gap-2">
                  <UserIcon className="h-4 w-4 text-muted-foreground" />
                  <div className="truncate font-medium text-xs">
                    {user.email}
                  </div>
                </div>
                <div className="py-1">
                  <Link
                    to="/account"
                    className="flex items-center px-3 py-2 text-xs text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                    onClick={() => setIsUserMenuOpen(false)}
                  >
                    <UserIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                    Account Settings
                  </Link>
                  {isAdmin && (
                    <Link
                      to="/admin"
                      className="flex items-center px-3 py-2 text-xs text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <AdminIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                      Admin Dashboard
                    </Link>
                  )}
                  <button
                    type="button"
                    className="flex items-center w-full px-3 py-2 text-left text-xs text-destructive hover:bg-destructive/10 transition-colors"
                    onClick={handleLogout}
                  >
                    <LogoutIcon className="h-4 w-4 mr-2 text-destructive" />
                    Logout
                  </button>
                </div>
              </div>
            )}
          </div>
        ) : (
          <span className="truncate">Guest</span>
        )}
        <WebSocketBadge />
        {syncStatus && <SyncStatusBadge syncStatus={syncStatus} />}
      </div>

      <div className="flex items-center gap-2 md:gap-3">
        {/* Progress bar - show when tiles are loading or JXL is loading */}
        {(tileProgress.started > 0 || isJXLLoading) && (
          <div className="w-16 md:w-32 h-1 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full transition-all duration-300 ${
                isJXLLoading
                  ? "bg-orange-400/80"
                  : isRefreshing
                  ? "bg-indigo-400/80"
                  : "bg-muted-foreground/80"
              }`}
              style={{
                width: isJXLLoading ? "100%" : `${percent}%`,
                animation: isJXLLoading
                  ? "pulse 1.5s ease-in-out infinite"
                  : undefined,
              }}
            />
          </div>
        )}

        {/* Tile status - right aligned */}
        <div className="text-right min-w-[3rem] md:min-w-[4rem]">
          <span
            className={`tabular-nums text-xs ${
              isJXLLoading
                ? "text-orange-300"
                : tileProgress.inFlight > 0
                ? "text-indigo-300"
                : "text-muted-foreground"
            }`}
          >
            {tileStatusText}
          </span>
        </div>
      </div>
    </div>
  );
}
