// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";

interface UseMaskKeyboardShortcutsProps {
  onToggleMask: () => void;
  onIncreaseMaskOpacity: () => void;
  onDecreaseMaskOpacity: () => void;
}

/**
 * Hook for managing mask-related keyboard shortcuts
 */
export function useMaskKeyboardShortcuts({
  onToggleMask,
  onIncreaseMaskOpacity,
  onDecreaseMaskOpacity,
}: UseMaskKeyboardShortcutsProps) {
  useKeyboardShortcuts({
    onKeyT: onToggleMask,
    onKeyPlus: onIncreaseMaskOpacity,
    onKeyMinus: onDecreaseMaskOpacity,
  });
}
