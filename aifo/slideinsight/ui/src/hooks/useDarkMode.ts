// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useEffect } from "react";

export interface UseDarkModeReturn {
  isDark: boolean;
  toggle: () => void;
  setDark: (dark: boolean) => void;
}

/**
 * Hook for managing dark mode state across the application.
 * Uses the same localStorage key and logic as the Navbar component.
 */
export function useDarkMode(): UseDarkModeReturn {
  const [isDark, setIsDark] = useState<boolean>(() => {
    try {
      const storedTheme = localStorage.getItem("si-theme");
      if (storedTheme === "dark") return true;
      if (storedTheme === "light") return false;
    } catch (_) {
      // Ignore access errors (e.g., privacy mode)
    }
    if (typeof window !== "undefined" && window.matchMedia) {
      return window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    return false;
  });

  // Sync theme to document and persist in localStorage
  useEffect(() => {
    const root = document.documentElement;
    if (isDark) {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
    try {
      localStorage.setItem("si-theme", isDark ? "dark" : "light");
    } catch (_) {
      // Ignore storage errors
    }
  }, [isDark]);

  // React to external changes (e.g., navbar toggle, other tabs, classList changes)
  useEffect(() => {
    const root = document.documentElement;

    const handleStorage = (e: StorageEvent) => {
      if (e.key === "si-theme" && e.newValue) {
        setIsDark(e.newValue === "dark");
      }
    };

    const mutationObserver = new MutationObserver(() => {
      const currentlyDark = root.classList.contains("dark");
      setIsDark((prev) => (prev !== currentlyDark ? currentlyDark : prev));
    });

    mutationObserver.observe(root, {
      attributes: true,
      attributeFilter: ["class"],
    });
    window.addEventListener("storage", handleStorage);

    return () => {
      mutationObserver.disconnect();
      window.removeEventListener("storage", handleStorage);
    };
  }, []);

  const toggle = () => setIsDark((prev) => !prev);
  const setDark = (dark: boolean) => setIsDark(dark);

  return {
    isDark,
    toggle,
    setDark,
  };
}
