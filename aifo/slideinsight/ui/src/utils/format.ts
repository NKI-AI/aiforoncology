// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Format bytes into human-readable format (B, KB, MB, GB, TB)
 */
export const formatBytes = (bytes: number, decimals: number = 1): string => {
  if (bytes === 0 || bytes == null) return "0 B";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
};

/**
 * Format numbers with K, M, B suffixes for large values
 */
export const formatNumber = (num: number): string => {
  if (num == null) return "0";
  if (num >= 1e9) return (num / 1e9).toFixed(1) + "B";
  if (num >= 1e6) return (num / 1e6).toFixed(1) + "M";
  if (num >= 1e3) return (num / 1e3).toFixed(1) + "K";
  return num.toString();
};

/**
 * Format decimal number as percentage with specified decimal places
 */
export const formatPercentage = (
  value: number,
  decimals: number = 1
): string => {
  if (value == null) return "0.0%";
  return `${value.toFixed(decimals)}%`;
};

/**
 * Format nanoseconds into human-readable duration (ms, s, m)
 */
export const formatDuration = (nanoseconds: number): string => {
  if (nanoseconds == null) return "0ms";
  const milliseconds = nanoseconds / 1e6;
  if (milliseconds < 1000) return `${milliseconds.toFixed(1)}ms`;
  const seconds = milliseconds / 1000;
  if (seconds < 60) return `${seconds.toFixed(1)}s`;
  const minutes = seconds / 60;
  return `${minutes.toFixed(1)}m`;
};

/**
 * Format seconds into human-readable time format (s, m, h, d)
 */
export const formatTime = (seconds: number): string => {
  if (seconds == null || seconds === 0) return "0s";

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (days > 0) return `${days}d ${hours}h ${mins}m`;
  if (hours > 0) return `${hours}h ${mins}m ${secs}s`;
  if (mins > 0) return `${mins}m ${secs}s`;
  return `${secs}s`;
};

/**
 * Format Unix timestamp into localized date string
 */
export const formatTimestamp = (timestamp: number): string => {
  if (timestamp == null || timestamp === 0) return "Never";
  return new Date(timestamp * 1000).toLocaleString();
};

/**
 * Format relative time from a date string (e.g., "2m ago", "1h ago")
 */
export const formatRelativeTime = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return "Just now";
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes}m ago`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours}h ago`;
  } else {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days}d ago`;
  }
};

/**
 * Formats a date string to a localized format
 * @param dateString - ISO date string
 * @returns Formatted date string or "Invalid date" on error
 */
export function formatDate(dateString: string): string {
  try {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return "Invalid date";
  }
}

/**
 * Formats a date string to a shorter format (without time)
 * @param dateString - ISO date string
 * @returns Formatted date string or "Invalid date" on error
 */
export function formatDateShort(dateString: string): string {
  try {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return "Invalid date";
  }
}
