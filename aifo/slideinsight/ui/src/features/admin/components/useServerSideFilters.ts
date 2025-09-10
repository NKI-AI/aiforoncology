// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useEffect, useCallback } from "react";

/**
 * Generic server-side filters hook that manages search state with debouncing.
 * This isolates search UI state from server requests and prevents table re-renders.
 */
export function useServerSideFilters<T extends Record<string, any>>(
  onFiltersChange: (filters: T) => void,
  initialFilters: T,
  debounceMs: number = 300
) {
  // Local state for immediate UI updates (no re-renders of table)
  const [filters, setFilters] = useState<T>(initialFilters);

  // Debounced effect for server requests
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      onFiltersChange(filters);
    }, debounceMs);

    return () => clearTimeout(timeoutId);
  }, [filters, onFiltersChange, debounceMs]);

  const updateFilter = useCallback((key: keyof T, value: any) => {
    setFilters((prev) => ({
      ...prev,
      [key]: value,
    }));
  }, []);

  const clearFilters = useCallback(() => {
    setFilters(initialFilters);
  }, [initialFilters]);

  const hasActiveFilters = useCallback(() => {
    return Object.values(filters).some((value) =>
      typeof value === "boolean" ? value : value !== ""
    );
  }, [filters]);

  return {
    filters,
    updateFilter,
    clearFilters,
    hasActiveFilters: hasActiveFilters(),
  };
}
