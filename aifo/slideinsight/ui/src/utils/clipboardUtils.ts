// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { toast } from "sonner";

/**
 * Safely copies text to clipboard with fallback for environments where clipboard API is not available
 */
export async function copyToClipboard(
  text: string,
  description?: string
): Promise<boolean> {
  try {
    // Check if the Clipboard API is available
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      toast.success("Copied to clipboard!", {
        description: description || `Copied: ${text}`,
      });
      return true;
    } else {
      // Fallback for environments without clipboard API
      return copyToClipboardFallback(text, description);
    }
  } catch (error) {
    console.warn("Failed to copy to clipboard:", error);
    // Try fallback method
    return copyToClipboardFallback(text, description);
  }
}

/**
 * Fallback method using document.execCommand (deprecated but widely supported)
 */
function copyToClipboardFallback(text: string, description?: string): boolean {
  try {
    // Create a temporary textarea element
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    textarea.style.pointerEvents = "none";

    document.body.appendChild(textarea);
    textarea.select();
    textarea.setSelectionRange(0, 99999); // For mobile devices

    const successful = document.execCommand("copy");
    document.body.removeChild(textarea);

    if (successful) {
      toast.success("Copied to clipboard!", {
        description: description || `Copied: ${text}`,
      });
      return true;
    } else {
      throw new Error("execCommand copy failed");
    }
  } catch (error) {
    console.error("All clipboard methods failed:", error);

    // Final fallback - show the text to user
    toast.error("Could not copy to clipboard", {
      description: `Please manually copy: ${text}`,
      duration: 5000,
    });

    return false;
  }
}

/**
 * Check if clipboard functionality is available
 */
function isClipboardAvailable(): boolean {
  return (
    !!(navigator.clipboard && window.isSecureContext) ||
    document.queryCommandSupported?.("copy")
  );
}
