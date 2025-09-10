// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { formatDate as formatDateCore, formatRelativeTime } from "./format";

/**
 * Admin table utility functions
 */

/**
 * Format date for admin tables - wrapper around core formatDate
 */
export const formatDate = formatDateCore;

/**
 * Format relative time for admin tables - wrapper around core formatRelativeTime
 */
export const formatTableRelativeTime = formatRelativeTime;

/**
 * Common table column formatters for admin tables
 */
export const tableFormatters = {
  date: formatDate,
  relativeTime: formatRelativeTime,

  /**
   * Format boolean as Yes/No
   */
  boolean: (value: boolean): string => (value ? "Yes" : "No"),

  /**
   * Format array as comma-separated string
   */
  array: (value: any[]): string => value?.join(", ") || "",

  /**
   * Truncate text with ellipsis
   */
  truncate: (text: string, maxLength: number = 50): string => {
    if (!text) return "";
    return text.length <= maxLength
      ? text
      : `${text.substring(0, maxLength)}...`;
  },
};

/**
 * Common table utilities
 */
export const tableUtils = {
  /**
   * Get row ID from an entity
   */
  getRowId: (item: any): string => {
    return (
      item.id?.toString() || item.uid || item.userUid || item.studyUid || ""
    );
  },

  /**
   * Create a sortable column header
   */
  createSortableHeader: (label: string, sortKey: string) => ({
    label,
    sortKey,
    className: "cursor-pointer hover:bg-gray-100",
  }),
};
