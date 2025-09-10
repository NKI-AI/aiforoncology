// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { forceOverviewPosition } from "./OverviewMapFactory";
import {
  OVERVIEW_MAP_POSITION_DELAY,
  OVERVIEW_BUTTON_VISIBILITY_DELAY,
} from "./constants";

/**
 * Utility functions for overview map positioning and styling
 */

/**
 * Force overview map position - simplified version without retry
 */
export function forceOverviewPositionWithRetry(): Promise<void> {
  return new Promise((resolve) => {
    // Single attempt with minimal delay
    setTimeout(() => {
      forceOverviewPosition();
      resolve();
    }, OVERVIEW_MAP_POSITION_DELAY);
  });
}

/**
 * Ensure button visibility - simplified version
 */
export function ensureOverviewButtonVisibility(): Promise<void> {
  return new Promise((resolve) => {
    // Single attempt with basic styling
    setTimeout(() => {
      const overviewElement = document.querySelector(
        ".ol-overviewmap-topright"
      ) as HTMLElement;

      if (overviewElement) {
        // Apply basic styles to overview element
        Object.assign(overviewElement.style, {
          display: "block",
          visibility: "visible",
          position: "absolute",
          top: "10px",
          right: "10px",
          zIndex: "1200",
        });

        // Find and fix the button
        const button = overviewElement.querySelector("button") as HTMLElement;
        if (button) {
          // Apply basic styles to button
          Object.assign(button.style, {
            display: "block",
            visibility: "visible",
            opacity: "1",
            position: "absolute",
            bottom: "2px",
            left: "2px",
            width: "24px",
            height: "24px",
            backgroundColor: "white",
            border: "1px solid #ccc",
            borderRadius: "3px",
            cursor: "pointer",
            zIndex: "1201",
          });
        }
      }
      resolve();
    }, OVERVIEW_BUTTON_VISIBILITY_DELAY);
  });
}
