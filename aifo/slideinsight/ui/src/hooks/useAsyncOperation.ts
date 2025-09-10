// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback } from "react";
import type { AsyncOperationResult } from "./types";

/**
 * Hook for managing async operations with loading, error, and success states
 * @param asyncFunction - The async function to execute
 * @returns Object with operation state and control functions
 */
export function useAsyncOperation<T = any, TArgs extends any[] = any[]>(
  asyncFunction: (...args: TArgs) => Promise<T>
): AsyncOperationResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);

  const execute = useCallback(
    async (...args: TArgs) => {
      try {
        setLoading(true);
        setError(null);
        setSuccess(false);
        setData(null);

        const result = await asyncFunction(...args);

        setData(result);
        setSuccess(true);
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : "An error occurred";
        setError(errorMessage);
        setSuccess(false);
      } finally {
        setLoading(false);
      }
    },
    [asyncFunction]
  );

  const reset = useCallback(() => {
    setData(null);
    setLoading(false);
    setError(null);
    setSuccess(false);
  }, []);

  return {
    data,
    loading,
    error,
    success,
    execute,
    reset,
  };
}
