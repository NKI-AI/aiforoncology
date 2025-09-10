import React from "react";

interface LoadingOverlayProps {
  text?: string;
}

export function LoadingOverlay({ text = "Loading..." }: LoadingOverlayProps) {
  return (
    <div className="absolute inset-0 flex items-center justify-center bg-black/60 dark:bg-black/70">
      <div className="flex items-center space-x-3 px-4 py-3 rounded-lg shadow-lg bg-background/5 dark:bg-black/40 border border-white/10">
        <div className="h-5 w-5 rounded-full border-2 border-gray-300 dark:border-gray-400 border-t-transparent animate-spin" />
        <span className="text-sm text-muted-100">{text}</span>
      </div>
    </div>
  );
}

export default LoadingOverlay;
