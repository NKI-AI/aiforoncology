"use client";
import {
  motion,
  MotionValue,
  useMotionValue,
  useSpring,
  useTransform,
  useAnimationFrame,
  type SpringOptions,
  AnimatePresence,
} from "framer-motion";
import React, {
  Children,
  cloneElement,
  createContext,
  useContext,
  useEffect,
  useCallback,
  useMemo,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { cn } from "@/lib/utils";

const BASE_ICON = 40;
const DEFAULT_MAGNIFICATION = 56;
const DEFAULT_DISTANCE = 150;
const DEFAULT_PANEL_SIZE = 48;

type Side = "bottom" | "top" | "left" | "right";
type Align = "start" | "center" | "end";

type DockProps = {
  children: React.ReactNode;
  className?: string;
  distance?: number;
  panelSize?: number;
  magnification?: number;
  spring?: SpringOptions;
  side?: Side; // default 'bottom'
  align?: Align; // default 'center'
  floating?: boolean; // default false
  resizable?: boolean; // default true
  edgeInsetPx?: number; // default 12
};

type DockItemProps = {
  className?: string;
  children: React.ReactNode;
};

type DockLabelProps = {
  className?: string;
  children: React.ReactNode;
  isHovered?: MotionValue<number>;
  anchorRef?: React.RefObject<HTMLDivElement>;
  width?: MotionValue<number>;
  height?: MotionValue<number>;
};

type DockIconProps = {
  className?: string;
  children: React.ReactNode;
  width?: MotionValue<number>;
  height?: MotionValue<number>;
};

type DocContextType = {
  mousePrimaryAxis: MotionValue<number>;
  spring: SpringOptions;
  magnification: number;
  distance: number;
  side: Side;
  isVertical: boolean;
  railSize: number;
};

const DockContext = createContext<DocContextType | undefined>(undefined);
function DockProvider({
  children,
  value,
}: {
  children: React.ReactNode;
  value: DocContextType;
}) {
  return <DockContext.Provider value={value}>{children}</DockContext.Provider>;
}
function useDock() {
  const ctx = useContext(DockContext);
  if (!ctx) throw new Error("useDock must be used within a DockProvider");
  return ctx;
}

export function Dock({
  children,
  className,
  spring = { mass: 0.1, stiffness: 150, damping: 12 },
  magnification: magnificationProp = DEFAULT_MAGNIFICATION,
  distance = DEFAULT_DISTANCE,
  panelSize: panelSizeProp = DEFAULT_PANEL_SIZE,
  side = "bottom",
  align = "center",
  floating = false,
  resizable = true,
  edgeInsetPx = 12,
}: DockProps) {
  const isVertical = side === "left" || side === "right";
  const mousePrimaryAxis = useMotionValue(Infinity);
  const hoverState = useMotionValue(0);

  const [magnification, setMagnification] = useState(magnificationProp);
  const [panelSize, setPanelSize] = useState(panelSizeProp);
  useEffect(() => setMagnification(magnificationProp), [magnificationProp]);
  useEffect(() => setPanelSize(panelSizeProp), [panelSizeProp]);

  const growth = Math.max(0, magnification - BASE_ICON);
  const expandedRail = panelSize + Math.ceil(growth * 0.75);
  const railThickness = useSpring(
    useTransform(hoverState, [0, 1], [panelSize, expandedRail]),
    spring
  );

  const onMove = (e: React.MouseEvent) => {
    hoverState.set(1);
    if (isVertical) mousePrimaryAxis.set(e.clientY);
    else mousePrimaryAxis.set(e.clientX);
  };
  const onLeave = () => {
    hoverState.set(0);
    mousePrimaryAxis.set(Infinity);
  };

  // Edge snapping (wrapper)
  const sidePosClass =
    side === "bottom"
      ? align === "start"
        ? "left-[env(safe-area-inset-left)] bottom-[env(safe-area-inset-bottom)]"
        : align === "end"
        ? "right-[env(safe-area-inset-right)] bottom-[env(safe-area-inset-bottom)]"
        : "left-1/2 bottom-[env(safe-area-inset-bottom)] -translate-x-1/2"
      : side === "top"
      ? align === "start"
        ? "left-[env(safe-area-inset-left)] top-[env(safe-area-inset-top)]"
        : align === "end"
        ? "right-[env(safe-area-inset-right)] top-[env(safe-area-inset-top)]"
        : "left-1/2 top-[env(safe-area-inset-top)] -translate-x-1/2"
      : side === "right"
      ? align === "start"
        ? "right-[env(safe-area-inset-right)] top-[env(safe-area-inset-top)]"
        : align === "end"
        ? "right-[env(safe-area-inset-right)] bottom-[env(safe-area-inset-bottom)]"
        : "right-[env(safe-area-inset-right)] top-1/2 -translate-y-1/2"
      : // left
      align === "start"
      ? "left-[env(safe-area-inset-left)] top-[env(safe-area-inset-top)]"
      : align === "end"
      ? "left-[env(safe-area-inset-left)] bottom-[env(safe-area-inset-bottom)]"
      : "left-[env(safe-area-inset-left)] top-1/2 -translate-y-1/2";

  // Align inside wrapper (long axis)
  const alignClass =
    side === "bottom" || side === "top"
      ? align === "start"
        ? "justify-start"
        : align === "end"
        ? "justify-end"
        : "justify-center"
      : align === "start"
      ? "items-start"
      : align === "end"
      ? "items-end"
      : "items-center";

  // Alt+Drag resize
  const dragRef = useRef<HTMLDivElement>(null);
  const draggingRef = useRef(false);
  const startRef = useRef<number>(0);
  const startMagRef = useRef<number>(0);
  const startPanelRef = useRef<number>(0);

  const onPointerDown = (e: React.PointerEvent) => {
    if (!resizable || !e.altKey) return;
    draggingRef.current = true;
    startRef.current = isVertical ? e.clientY : e.clientX;
    startMagRef.current = magnification;
    startPanelRef.current = panelSize;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  };
  const onPointerMove = (e: React.PointerEvent) => {
    if (!draggingRef.current) return;
    const now = isVertical ? e.clientY : e.clientX;
    const delta = now - startRef.current;
    const newMag = Math.min(
      120,
      Math.max(36, startMagRef.current + delta * 0.3)
    );
    const newPanel = Math.min(
      96,
      Math.max(32, startPanelRef.current + delta * 0.15)
    );
    setMagnification(newMag);
    setPanelSize(newPanel);
  };
  const onPointerUp = (e: React.PointerEvent) => {
    if (!draggingRef.current) return;
    draggingRef.current = false;
    (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
  };

  const railAxisClass = isVertical ? "flex-col" : "flex-row";
  const railCrossAlignClass = isVertical ? "items-center" : "items-end";
  const railPaddingClass = isVertical ? "py-4 px-2" : "px-4 py-2";

  const containerStyle = isVertical
    ? { width: railThickness as any, overflow: "visible" as const }
    : { height: railThickness as any, overflow: "visible" as const };

  const wrapperClasses = cn(
    "z-50 pointer-events-auto",
    floating && "fixed",
    floating && sidePosClass,
    className
  );

  const content = (
    <div
      className={wrapperClasses}
      style={floating ? { padding: edgeInsetPx } : undefined}
    >
      <motion.div
        onMouseMove={onMove}
        onMouseLeave={onLeave}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        style={containerStyle}
        className={cn(
          "flex overflow-visible", // no mx-auto to avoid recentering
          isVertical ? "h-[min(70vh,100%)]" : "items-end",
          alignClass
        )}
        role="none"
      >
        <div
          ref={dragRef}
          className={cn(
            "flex w-fit rounded-2xl bg-gray-50 dark:bg-neutral-900 gap-4",
            railAxisClass,
            railCrossAlignClass,
            railPaddingClass,
            "overflow-visible shadow-sm"
          )}
          role="toolbar"
          aria-label="Application dock"
          aria-orientation={isVertical ? "vertical" : "horizontal"}
        >
          <DockProvider
            value={{
              mousePrimaryAxis,
              spring,
              distance,
              magnification,
              side,
              isVertical,
              railSize: panelSize,
            }}
          >
            {children}
          </DockProvider>
        </div>
      </motion.div>
    </div>
  );

  // IMPORTANT: portal the floating dock to <body> to escape transformed ancestors
  return floating ? createPortal(content, document.body) : content;
}

export function DockItem({ children, className }: DockItemProps) {
  const ref = useRef<HTMLDivElement>(null);
  const {
    distance,
    magnification,
    mousePrimaryAxis,
    spring,
    isVertical,
    side,
  } = useDock();
  const isHovered = useMotionValue(0);

  const mouseDelta = useTransform(mousePrimaryAxis, (val) => {
    const r = ref.current?.getBoundingClientRect();
    if (!r) return Infinity;
    const center = isVertical ? r.top + r.height / 2 : r.left + r.width / 2;
    return val - center;
  });

  const sizeTransform = useTransform(
    mouseDelta,
    [-distance, 0, distance],
    [BASE_ICON, magnification, BASE_ICON]
  );

  const width = useSpring(sizeTransform, spring);
  const height = useSpring(sizeTransform, spring);

  const transformOrigin = isVertical
    ? side === "right"
      ? "center right"
      : "center left"
    : "bottom center";

  return (
    <motion.div
      ref={ref}
      style={{ width, height, transformOrigin }}
      onHoverStart={() => isHovered.set(1)}
      onHoverEnd={() => isHovered.set(0)}
      onFocus={() => isHovered.set(1)}
      onBlur={() => isHovered.set(0)}
      className={cn(
        "relative inline-flex items-center justify-center",
        className
      )}
      tabIndex={0}
      role="button"
      aria-haspopup="true"
    >
      {Children.map(children, (child) =>
        cloneElement(child as React.ReactElement<any>, {
          width,
          height,
          isHovered,
          anchorRef: ref,
        })
      )}
    </motion.div>
  );
}

export function DockLabel({
  children,
  className,
  isHovered,
  anchorRef,
  width,
  height,
}: DockLabelProps) {
  const { side, isVertical } = useDock();
  const [visible, setVisible] = useState(false);
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null);
  const [placement, setPlacement] = useState<
    "top" | "bottom" | "left" | "right"
  >("top");

  const EDGE_PAD = 8;
  const ARROW = 10;

  const preferred = useMemo<"top" | "bottom" | "left" | "right">(() => {
    if (isVertical) return side === "right" ? "left" : "right";
    return side === "top" ? "bottom" : "top";
  }, [isVertical, side]);

  useEffect(() => {
    if (!isHovered) return;
    const unsub = isHovered.on("change", (v) => setVisible(v === 1));
    return () => unsub?.();
  }, [isHovered]);

  const recalc = useCallback(() => {
    if (!anchorRef?.current) return;
    const r = anchorRef.current.getBoundingClientRect();
    const gap = Math.max(10, Math.min(22, r.height * 0.25));

    const base = (p: typeof placement) => {
      switch (p) {
        case "top":
          return { left: r.left + r.width / 2, top: r.top - gap };
        case "bottom":
          return { left: r.left + r.width / 2, top: r.bottom + gap };
        case "left":
          return { left: r.left - gap, top: r.top + r.height / 2 };
        case "right":
          return { left: r.right + gap, top: r.top + r.height / 2 };
      }
    };

    let place = preferred;
    let { left, top } = base(place);

    if (
      (place === "top" && top < EDGE_PAD) ||
      (place === "bottom" && top > window.innerHeight - EDGE_PAD)
    ) {
      place = place === "top" ? "bottom" : "top";
      ({ left, top } = base(place));
    }
    if (
      (place === "left" && left < EDGE_PAD) ||
      (place === "right" && left > window.innerWidth - EDGE_PAD)
    ) {
      place = place === "left" ? "right" : "left";
      ({ left, top } = base(place));
    }

    if (place === "top" || place === "bottom") {
      left = Math.min(Math.max(left, EDGE_PAD), window.innerWidth - EDGE_PAD);
    } else {
      top = Math.min(Math.max(top, EDGE_PAD), window.innerHeight - EDGE_PAD);
    }

    setPlacement(place);
    setPos({ left, top });
  }, [anchorRef, preferred]);

  useEffect(() => {
    if (visible) recalc();
  }, [visible, recalc]);

  useEffect(() => {
    if (!visible) return;
    const uW = width?.on("change", recalc);
    const uH = height?.on("change", recalc);
    const onScroll = () => recalc();
    const onResize = () => recalc();
    window.addEventListener("scroll", onScroll, true);
    window.addEventListener("resize", onResize);
    return () => {
      uW?.();
      uH?.();
      window.removeEventListener("scroll", onScroll, true);
      window.removeEventListener("resize", onResize);
    };
  }, [visible, width, height, recalc]);

  useAnimationFrame(() => {
    if (visible) recalc();
  });

  if (!visible || !pos) return null;

  const wrapperStyle: React.CSSProperties =
    placement === "top"
      ? {
          position: "fixed",
          left: pos.left,
          top: pos.top,
          transform: "translate(-50%, -100%)",
        }
      : placement === "bottom"
      ? {
          position: "fixed",
          left: pos.left,
          top: pos.top,
          transform: "translate(-50%, 0)",
        }
      : placement === "left"
      ? {
          position: "fixed",
          left: pos.left,
          top: pos.top,
          transform: "translate(-100%, -50%)",
        }
      : {
          position: "fixed",
          left: pos.left,
          top: pos.top,
          transform: "translate(0, -50%)",
        };

  const arrowPos: React.CSSProperties =
    placement === "top"
      ? {
          left: "50%",
          top: "100%",
          transform: "translate(-50%, -50%) rotate(45deg)",
        }
      : placement === "bottom"
      ? {
          left: "50%",
          top: 0,
          transform: "translate(-50%, -50%) rotate(45deg)",
        }
      : placement === "left"
      ? {
          left: "100%",
          top: "50%",
          transform: "translate(-50%, -50%) rotate(45deg)",
        }
      : {
          left: 0,
          top: "50%",
          transform: "translate(-50%, -50%) rotate(45deg)",
        };

  return createPortal(
    <div className="pointer-events-none z-[9999]" style={wrapperStyle}>
      <AnimatePresence>
        <motion.div
          initial={{ opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.98 }}
          transition={{ duration: 0.12 }}
          role="tooltip"
          className={cn(
            "relative w-max max-w-xs whitespace-nowrap rounded-md border px-2 py-1 text-xs shadow",
            "bg-gray-100 border-gray-200 text-neutral-700",
            "dark:bg-neutral-800 dark:border-neutral-900 dark:text-white",
            className
          )}
        >
          {children}
          <span
            aria-hidden
            className={cn(
              "absolute block",
              "bg-gray-100 border border-gray-200",
              "dark:bg-neutral-800 dark:border-neutral-900"
            )}
            style={{ ...arrowPos, width: ARROW, height: ARROW }}
          />
          <span
            aria-hidden
            className={cn(
              "absolute block",
              "bg-gray-100",
              "dark:bg-neutral-800"
            )}
            style={{ ...arrowPos, width: ARROW - 2, height: ARROW - 2 }}
          />
        </motion.div>
      </AnimatePresence>
    </div>,
    document.body
  );
}

export function DockIcon({
  children,
  className,
  width,
  height,
}: DockIconProps) {
  return (
    <motion.div
      style={{ width, height }}
      className={cn("flex items-center justify-center", className)}
    >
      {children}
    </motion.div>
  );
}
