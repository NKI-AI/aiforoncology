// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

type DockMode = "free" | "left";

interface SlidePanelContextValue {
  effectiveDock: DockMode;
  isDragging: boolean;
  headerRef: React.RefObject<HTMLDivElement>;
  onHeaderMouseDown: (e: React.MouseEvent) => void;
  onHeaderTouchStart: (e: React.TouchEvent) => void;
  toggleDock: () => void;
}

const SlidePanelContext = createContext<SlidePanelContextValue | null>(null);

function useSlidePanelContext(): SlidePanelContextValue {
  const ctx = useContext(SlidePanelContext);
  if (!ctx) {
    throw new Error("SlidePanel.Header must be used within SlidePanel");
  }
  return ctx;
}

export interface SlidePanelProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;

  // Docking control
  dockOverride?: DockMode;
  onDockChange?: (dock: DockMode) => void;
  defaultDock?: DockMode;

  // Persistence
  storageKey?: string; // base key for dock/position/size

  // Sizing/position
  defaultPosition?: { x: number; y: number };
  defaultSize?: { width: number; height: number };
  dockedWidth?: number;
  minWidth?: number;
  minHeight?: number;

  // Styling
  className?: string;
}

export function SlidePanel({
  isOpen,
  onClose,
  children,
  dockOverride,
  onDockChange,
  defaultDock = "free",
  storageKey,
  defaultPosition = { x: 100, y: 100 },
  defaultSize = { width: 320, height: 560 },
  dockedWidth,
  minWidth = 280,
  minHeight = 400,
  className,
}: SlidePanelProps) {
  // Internal state (persisted)
  const [dock, setDock] = useState<DockMode>(defaultDock);
  const [position, setPosition] = useState<{ x: number; y: number }>(
    defaultPosition
  );
  const [size, setSize] = useState<{ width: number; height: number }>(
    defaultSize
  );

  // Transient UI state
  const [isDragging, setIsDragging] = useState(false);
  const [isResizing, setIsResizing] = useState(false);
  const [dragOffset, setDragOffset] = useState<{ x: number; y: number }>({
    x: 0,
    y: 0,
  });
  const [resizeStart, setResizeStart] = useState<{
    x: number;
    y: number;
    width: number;
    height: number;
  }>({ x: 0, y: 0, width: 0, height: 0 });
  const [resizeDirection, setResizeDirection] = useState<string>("");
  const [viewport, setViewport] = useState<{ width: number; height: number }>({
    width: 0,
    height: 0,
  });

  const panelRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);

  // Load persisted state
  useEffect(() => {
    if (!storageKey) return;
    try {
      const savedDock = localStorage.getItem(`${storageKey}_dock`);
      if (savedDock === "free" || savedDock === "left") setDock(savedDock);
      const savedPos = localStorage.getItem(`${storageKey}_pos`);
      if (savedPos) {
        const parsed = JSON.parse(savedPos);
        if (typeof parsed?.x === "number" && typeof parsed?.y === "number")
          setPosition(parsed);
      }
      const savedSize = localStorage.getItem(`${storageKey}_size`);
      if (savedSize) {
        const parsed = JSON.parse(savedSize);
        if (
          typeof parsed?.width === "number" &&
          typeof parsed?.height === "number"
        )
          setSize(parsed);
      }
    } catch {
      // ignore
    }
  }, [storageKey]);

  // Persist state
  useEffect(() => {
    if (!storageKey) return;
    localStorage.setItem(`${storageKey}_dock`, dockOverride ?? dock);
  }, [dock, dockOverride, storageKey]);
  useEffect(() => {
    if (!storageKey) return;
    localStorage.setItem(`${storageKey}_pos`, JSON.stringify(position));
  }, [position, storageKey]);
  useEffect(() => {
    if (!storageKey) return;
    localStorage.setItem(`${storageKey}_size`, JSON.stringify(size));
  }, [size, storageKey]);

  // Track viewport size
  useEffect(() => {
    const update = () => {
      if (typeof window !== "undefined") {
        setViewport({ width: window.innerWidth, height: window.innerHeight });
      }
    };
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);

  const effectiveDock: DockMode = dockOverride ?? dock;

  // Focus the active panel on open so keyboard shortcuts work immediately
  useEffect(() => {
    if (!isOpen) return;
    const rafId = requestAnimationFrame(() => {
      panelRef.current?.focus();
    });
    return () => cancelAnimationFrame(rafId);
  }, [isOpen]);

  // Close the active panel with Escape
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Escape") {
        e.stopPropagation();
        onClose();
      }
    },
    [onClose]
  );

  // Helper to get coordinates from mouse or touch event
  const getEventCoordinates = (e: MouseEvent | TouchEvent) => {
    if ("touches" in e) {
      return { clientX: e.touches[0].clientX, clientY: e.touches[0].clientY };
    }
    return { clientX: e.clientX, clientY: e.clientY };
  };

  // Dragging start (only when floating) - supports both mouse and touch
  const onHeaderMouseDown = (e: React.MouseEvent) => {
    if (effectiveDock !== "free") return;
    if (!headerRef.current?.contains(e.target as Node)) return;
    e.preventDefault(); // Prevent text selection and other default behaviors
    setIsDragging(true);
    if (dockOverride) {
      onDockChange?.("free");
    } else {
      setDock("free");
    }
    const rect = panelRef.current?.getBoundingClientRect();
    if (rect)
      setDragOffset({ x: e.clientX - rect.left, y: e.clientY - rect.top });
  };

  const onHeaderTouchStart = (e: React.TouchEvent) => {
    if (effectiveDock !== "free") return;
    if (!headerRef.current?.contains(e.target as Node)) return;
    e.preventDefault(); // Prevent scrolling and other default behaviors
    setIsDragging(true);
    if (dockOverride) {
      onDockChange?.("free");
    } else {
      setDock("free");
    }
    const rect = panelRef.current?.getBoundingClientRect();
    if (rect) {
      const touch = e.touches[0];
      setDragOffset({
        x: touch.clientX - rect.left,
        y: touch.clientY - rect.top,
      });
    }
  };

  const handleResizeMouseDown = (e: React.MouseEvent, direction: string) => {
    e.stopPropagation();
    e.preventDefault();
    setIsResizing(true);
    setResizeDirection(direction);
    setResizeStart({
      x: e.clientX,
      y: e.clientY,
      width: size.width,
      height: size.height,
    });
  };

  const handleResizeTouchStart = (e: React.TouchEvent, direction: string) => {
    e.stopPropagation();
    e.preventDefault();
    setIsResizing(true);
    setResizeDirection(direction);
    const touch = e.touches[0];
    setResizeStart({
      x: touch.clientX,
      y: touch.clientY,
      width: size.width,
      height: size.height,
    });
  };

  const handleMove = (e: MouseEvent | TouchEvent) => {
    const coords = getEventCoordinates(e);

    if (isDragging) {
      const nextX = coords.clientX - dragOffset.x;
      const nextY = coords.clientY - dragOffset.y;
      const margin = 16;
      const clampedX = Math.max(
        margin,
        Math.min(nextX, Math.max(margin, viewport.width - size.width - margin))
      );
      const clampedY = Math.max(
        margin,
        Math.min(
          nextY,
          Math.max(margin, viewport.height - size.height - margin)
        )
      );
      setPosition({ x: clampedX, y: clampedY });
    } else if (isResizing) {
      const dx = coords.clientX - resizeStart.x;
      const dy = coords.clientY - resizeStart.y;
      let newWidth = size.width;
      let newHeight = size.height;
      let newX = position.x;
      let newY = position.y;
      if (resizeDirection.includes("right"))
        newWidth = Math.max(minWidth, resizeStart.width + dx);
      if (resizeDirection.includes("left")) {
        const w2 = resizeStart.width - dx;
        if (w2 >= minWidth) {
          newWidth = w2;
          newX = position.x + dx;
        }
      }
      if (resizeDirection.includes("bottom"))
        newHeight = Math.max(minHeight, resizeStart.height + dy);
      if (resizeDirection.includes("top")) {
        const h2 = resizeStart.height - dy;
        if (h2 >= minHeight) {
          newHeight = h2;
          newY = position.y + dy;
        }
      }
      setSize({ width: newWidth, height: newHeight });
      const margin = 16;
      const clampedX = Math.max(
        margin,
        Math.min(newX, Math.max(margin, viewport.width - newWidth - margin))
      );
      const clampedY = Math.max(
        margin,
        Math.min(newY, Math.max(margin, viewport.height - newHeight - margin))
      );
      setPosition({ x: clampedX, y: clampedY });
    }
  };

  const handleEnd = () => {
    setIsDragging(false);
    setIsResizing(false);
    setResizeDirection("");
    const threshold = 32;
    if (viewport.width > 0) {
      const distLeft = Math.abs(position.x - 0);
      if (distLeft < threshold) {
        if (dockOverride) {
          onDockChange?.("left");
        } else {
          setDock("left");
        }
      }
    }
  };

  useEffect(() => {
    if (isDragging || isResizing) {
      // Add both mouse and touch event listeners
      document.addEventListener("mousemove", handleMove);
      document.addEventListener("mouseup", handleEnd);
      document.addEventListener("touchmove", handleMove, { passive: false });
      document.addEventListener("touchend", handleEnd);
      document.addEventListener("touchcancel", handleEnd);

      return () => {
        document.removeEventListener("mousemove", handleMove);
        document.removeEventListener("mouseup", handleEnd);
        document.removeEventListener("touchmove", handleMove);
        document.removeEventListener("touchend", handleEnd);
        document.removeEventListener("touchcancel", handleEnd);
      };
    }
  }, [
    isDragging,
    isResizing,
    dragOffset,
    resizeStart,
    resizeDirection,
    viewport,
    position,
    size,
    dockOverride,
  ]);

  const margin = 16;
  const computedLeft = effectiveDock === "left" ? 0 : position.x;
  const computedTop = Math.max(
    margin,
    Math.min(
      position.y,
      Math.max(margin, viewport.height - size.height - margin)
    )
  );

  // Determine border radius based on docked position
  const getBorderRadius = () => {
    if (effectiveDock !== "left") return "rounded-lg";

    // When docked, remove all rounding to create a continuous panel interface
    // except for the first panel which gets left rounding if at edge
    return "rounded-none";
  };

  const toggleDock = useCallback(() => {
    const next: DockMode = effectiveDock === "left" ? "free" : "left";
    if (dockOverride) onDockChange?.(next);
    else setDock(next);
  }, [effectiveDock, dockOverride, onDockChange]);

  const ctxValue = useMemo<SlidePanelContextValue>(
    () => ({
      effectiveDock,
      isDragging,
      headerRef,
      onHeaderMouseDown,
      onHeaderTouchStart,
      toggleDock,
    }),
    [effectiveDock, isDragging]
  );

  if (!isOpen) return null;

  return (
    <SlidePanelContext.Provider value={ctxValue}>
      {effectiveDock === "left" ? (
        <div
          ref={panelRef}
          className={[
            "flex bg-card text-card-foreground overflow-hidden z-40 select-none flex-col h-full min-h-0 border-r border-border pointer-events-auto",
            getBorderRadius(),
            className ?? "",
          ].join(" ")}
          style={{ width: `${dockedWidth ?? size.width}px` }}
          onKeyDown={handleKeyDown}
          tabIndex={-1}
        >
          {children}
        </div>
      ) : (
        <div
          ref={panelRef}
          className={[
            "fixed flex bg-card text-card-foreground rounded-lg overflow-hidden shadow-xl z-50 select-none flex-col min-h-0 border border-border pointer-events-auto",
            className ?? "",
          ].join(" ")}
          style={{
            left: `${computedLeft}px`,
            top: `${computedTop}px`,
            width: `${size.width}px`,
            height: `${size.height}px`,
          }}
          onMouseDown={onHeaderMouseDown}
          onTouchStart={onHeaderTouchStart}
          onKeyDown={handleKeyDown}
          tabIndex={-1}
        >
          {children}

          {/* Resize handles - larger touch-friendly areas */}
          <div
            className="absolute top-0 left-4 right-4 h-2 cursor-n-resize hover:bg-blue-500 hover:opacity-30 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "top")}
            onTouchStart={(e) => handleResizeTouchStart(e, "top")}
          />
          <div
            className="absolute bottom-0 left-4 right-4 h-2 cursor-s-resize hover:bg-blue-500 hover:opacity-30 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "bottom")}
            onTouchStart={(e) => handleResizeTouchStart(e, "bottom")}
          />
          <div
            className="absolute top-4 bottom-4 left-0 w-2 cursor-w-resize hover:bg-blue-500 hover:opacity-30 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "left")}
            onTouchStart={(e) => handleResizeTouchStart(e, "left")}
          />
          <div
            className="absolute top-4 bottom-4 right-0 w-2 cursor-e-resize hover:bg-blue-500 hover:opacity-30 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "right")}
            onTouchStart={(e) => handleResizeTouchStart(e, "right")}
          />
          <div
            className="absolute top-0 left-0 w-4 h-4 cursor-nw-resize hover:bg-blue-500 hover:opacity-50 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "top left")}
            onTouchStart={(e) => handleResizeTouchStart(e, "top left")}
          />
          <div
            className="absolute top-0 right-0 w-4 h-4 cursor-ne-resize hover:bg-blue-500 hover:opacity-50 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "top right")}
            onTouchStart={(e) => handleResizeTouchStart(e, "top right")}
          />
          <div
            className="absolute bottom-0 left-0 w-4 h-4 cursor-sw-resize hover:bg-blue-500 hover:opacity-50 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "bottom left")}
            onTouchStart={(e) => handleResizeTouchStart(e, "bottom left")}
          />
          <div
            className="absolute bottom-0 right-0 w-4 h-4 cursor-se-resize hover:bg-blue-500 hover:opacity-50 touch-manipulation"
            onMouseDown={(e) => handleResizeMouseDown(e, "bottom right")}
            onTouchStart={(e) => handleResizeTouchStart(e, "bottom right")}
          />
        </div>
      )}
    </SlidePanelContext.Provider>
  );
}

export interface SlidePanelHeaderProps {
  title: string;
  onClose?: () => void;
  rightActions?: React.ReactNode;
}

export function SlidePanelHeader({
  title,
  onClose,
  rightActions,
}: SlidePanelHeaderProps) {
  const {
    effectiveDock,
    isDragging,
    headerRef,
    onHeaderTouchStart,
    toggleDock,
  } = useSlidePanelContext();
  const isDocked = effectiveDock === "left";
  return (
    <div
      ref={headerRef}
      className={[
        "bg-muted px-3 py-1.5 text-base font-semibold flex items-center justify-between flex-shrink-0 touch-manipulation",
        isDocked
          ? "cursor-default"
          : isDragging
          ? "cursor-grabbing"
          : "cursor-grab",
      ].join(" ")}
      onTouchStart={onHeaderTouchStart}
      style={{ touchAction: isDocked ? "auto" : "none" }}
    >
      <span>{title}</span>
      <div className="flex items-center space-x-2">
        <button
          onClick={toggleDock}
          className="px-1.5 h-5 text-xs rounded bg-secondary/60 text-secondary-foreground hover:text-foreground hover:bg-secondary transition-colors"
          title={isDocked ? "Float" : "Dock to left"}
          aria-label={isDocked ? "Float" : "Dock to left"}
        >
          {isDocked ? "Float" : "Dock"}
        </button>
        {rightActions}
        {onClose && (
          <button
            onClick={onClose}
            className="w-5 h-5 text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Close"
          >
            {/* Heroicons XMarkIcon path consumer will import where needed; we keep markup generic here */}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="currentColor"
              className="w-4 h-4"
            >
              <path
                fillRule="evenodd"
                d="M5.72 5.72a.75.75 0 011.06 0L12 10.94l5.22-5.22a.75.75 0 111.06 1.06L13.06 12l5.22 5.22a.75.75 0 11-1.06 1.06L12 13.06l-5.22 5.22a.75.75 0 11-1.06-1.06L10.94 12 5.72 6.78a.75.75 0 010-1.06z"
                clipRule="evenodd"
              />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
}

export function SlidePanelBody({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex-1 min-h-0 bg-card text-card-foreground overflow-hidden">
      {children}
    </div>
  );
}

export default Object.assign(SlidePanel, {
  Header: SlidePanelHeader,
  Body: SlidePanelBody,
});
