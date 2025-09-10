// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback } from "react";
import type { ToggleState } from "./types";

/**
 * Hook for managing toggle state (open/closed, visible/hidden, etc.)
 * @param initialState - Initial state value (default: false)
 * @returns Object with toggle state and control functions
 */
export function useToggle(initialState: boolean = false): ToggleState {
  const [isOpen, setIsOpen] = useState<boolean>(initialState);

  const open = useCallback(() => {
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
  }, []);

  const toggle = useCallback(() => {
    setIsOpen((prev) => !prev);
  }, []);

  return {
    isOpen,
    open,
    close,
    toggle,
  };
}
