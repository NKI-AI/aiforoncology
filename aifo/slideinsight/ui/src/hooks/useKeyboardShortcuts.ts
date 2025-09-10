// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useEffect } from "react";

type KeyHandler = (e: KeyboardEvent) => void;

interface UseKeyboardShortcutsOptions {
  onKeyM?: () => void;
  onKeyS?: () => void;
  onKeyH?: () => void;
  onKeyEscape?: () => void;
  onKeyT?: () => void;
  onKeyC?: () => void;
  onKeyPlus?: () => void;
  onKeyMinus?: () => void;
  onArrowLeft?: () => void;
  onArrowRight?: () => void;
  onArrowUp?: () => void;
  onArrowDown?: () => void;
}

export function useKeyboardShortcuts({
  onKeyM,
  onKeyS,
  onKeyH,
  onKeyEscape,
  onKeyT,
  onKeyC,
  onKeyPlus,
  onKeyMinus,
  onArrowLeft,
  onArrowRight,
  onArrowUp,
  onArrowDown,
}: UseKeyboardShortcutsOptions) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore key events when typing in input fields
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        e.target instanceof HTMLSelectElement
      ) {
        return;
      }

      switch (e.key.toLowerCase()) {
        case "m":
          onKeyM?.();
          break;
        case "s":
          onKeyS?.();
          break;
        case "h":
          onKeyH?.();
          break;
        case "escape":
          onKeyEscape?.();
          break;
        case "t":
          onKeyT?.();
          break;
        case "c":
          onKeyC?.();
          break;
        case "+":
        case "=": // Common keyboard layout has + as Shift+=
          onKeyPlus?.();
          break;
        case "-":
          onKeyMinus?.();
          break;
        case "arrowleft":
          e.preventDefault();
          onArrowLeft?.();
          break;
        case "arrowright":
          e.preventDefault();
          onArrowRight?.();
          break;
        case "arrowup":
          e.preventDefault();
          onArrowUp?.();
          break;
        case "arrowdown":
          e.preventDefault();
          onArrowDown?.();
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    onKeyM,
    onKeyS,
    onKeyH,
    onKeyEscape,
    onKeyT,
    onKeyC,
    onKeyPlus,
    onKeyMinus,
    onArrowLeft,
    onArrowRight,
    onArrowUp,
    onArrowDown,
  ]);
}
