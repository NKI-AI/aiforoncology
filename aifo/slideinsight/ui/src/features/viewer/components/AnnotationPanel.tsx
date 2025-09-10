// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  XMarkIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  EyeIcon,
  EyeSlashIcon,
  TrashIcon,
  Cog6ToothIcon,
  QuestionMarkCircleIcon,
  ArrowsPointingInIcon,
  SparklesIcon,
} from "@heroicons/react/24/outline";
import SlidePanel from "@/components/ui/slide-panel";
import {
  useStudyAnnotationSettings,
  type StudyAnnotationLabel,
} from "@/hooks/useStudyAnnotationSettings";
import { useNavigate } from "@tanstack/react-router";
import Polygon from "ol/geom/Polygon";
import MultiPolygon from "ol/geom/MultiPolygon";
import {
  type AnnotationSettings,
  type AnnotationSettingsPartial,
} from "@/features/viewer/utils/annotationSettingsStorage";

export type LabelId = string;

// Default drawing tool per label type
type DrawingMode = "point" | "box" | "polygon";

type AllowedLabelDraw = {
  label: LabelId;
  mode: DrawingMode;
};

export interface AnnotationItem {
  id: string;
  name: string; // This now serves as the label identifier
  visible?: boolean;
  kind?: "polygon" | "box" | "point";
}

export interface AnnotationPanelProps {
  isOpen: boolean;
  onClose: () => void;

  activeLabel?: LabelId | null;
  onActiveLabelChange?: (label: LabelId | null) => void;

  annotations: AnnotationItem[];
  onUpdateAnnotations?: (items: AnnotationItem[]) => void;
  selectedId?: string | null;
  onSelect?: (id: string) => void;
  /** Hovered annotation id to highlight in the list */
  hoveredId?: string | null;
  /** Callback when hovering list items; pass null on leave */
  onHoverIdChange?: (id: string | null) => void;
  /** Callback when hovering group headers; pass array of IDs for group highlighting */
  onHoverGroupChange?: (ids: string[] | null) => void;

  dockOverride?: "free" | "left";
  onDockChange?: (dock: "free" | "left") => void;
  /** Start drawing geometry on the map with the current label */
  onStartDrawROI?: (mode: "point" | "box" | "polygon", label: LabelId) => void;
  /** Start drawing using a label-specific default tool. Types enforce valid pairs. */
  onStartDraw?: (args: AllowedLabelDraw) => void;
  /** Stop current drawing interaction (if any). */
  onStopDraw?: () => void;

  // Brush tool props
  brushActive?: boolean;
  brushMode?: "add" | "erase";
  brushSizePx?: number;
  onStartBrushAdd?: () => void;
  onStartBrushErase?: () => void;
  onStopBrush?: () => void;
  onBrushSizeChange?: (size: number) => void;

  // Study UID for fetching annotation settings
  studyUid?: string;

  // Annotation settings passed from parent to avoid duplicate API calls
  annotationSettings?: any;
  annotationSettingsLoading?: boolean;
  annotationSettingsError?: any;

  // Annotation display settings
  annotationDisplaySettings?: AnnotationSettings;
  onUpdateAnnotationDisplaySettings?: (
    settings: AnnotationSettingsPartial
  ) => void;
  onResetAnnotationDisplaySettings?: () => void;

  // Slide metadata for unit conversions
  slideMpp?: number;

  // Merge functionality
  onMergeAnnotations?: (annotationIds: string[]) => boolean;

  // Simplify functionality
  onSimplifyAnnotation?: (annotationId: string, tolerance: number) => boolean;
  onPreviewSimplification?: (
    annotationId: string,
    tolerance: number
  ) => {
    geometry: Polygon | MultiPolygon;
    originalPoints: number;
    simplifiedPoints: number;
    reduction: number;
    reductionPercent: string;
    toleranceUsedMicrometers: number;
  } | null;

  // Preview functionality
  onPreviewMerge?: (annotationIds: string[]) => Polygon | MultiPolygon | null;
  onClearPreview?: () => void;
}

// Helper function to get default tool for annotation type
const getDefaultToolForType = (
  type: "point" | "box" | "polygon"
): DrawingMode => {
  return type;
};

// Helper function to create label display data from study annotation settings
const createLabelsFromSettings = (
  annotations: StudyAnnotationLabel[]
): { id: LabelId; name: string; color: string; hint: DrawingMode }[] => {
  if (!annotations || !Array.isArray(annotations)) {
    return [];
  }

  return annotations
    .filter(
      (annotation) =>
        annotation && annotation.label && annotation.color && annotation.type
    )
    .map((annotation) => ({
      id: annotation.label,
      name:
        annotation.label.charAt(0).toUpperCase() + annotation.label.slice(1),
      color: annotation.color,
      hint: annotation.type,
    }));
};

function cls(...arr: (string | false | null | undefined)[]) {
  return arr.filter(Boolean).join(" ");
}

interface PanelFoldableBodyProps {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  isCollapsed: boolean;
  onToggleCollapsed: () => void;
  children: React.ReactNode;
}

function PanelFoldableBody({
  title,
  icon: Icon,
  isCollapsed,
  onToggleCollapsed,
  children,
}: PanelFoldableBodyProps) {
  return (
    <div className="mt-2 pt-2 border-t border-border700/60">
      <button
        onClick={onToggleCollapsed}
        className="flex items-center gap-1 text-xs text-muted-foreground400 hover:text-foreground transition-colors h-2"
      >
        {isCollapsed ? (
          <ChevronRightIcon className="h-3 w-3" />
        ) : (
          <ChevronDownIcon className="h-3 w-3" />
        )}
        <Icon className="h-3 w-3" />
        <span className="leading-none">{title}</span>
      </button>

      {!isCollapsed && <div className="ml-4 mt-2">{children}</div>}
    </div>
  );
}

export function AnnotationPanel({
  isOpen,
  onClose,
  activeLabel = null,
  onActiveLabelChange,
  annotations,
  onUpdateAnnotations,
  selectedId = null,
  onSelect,
  dockOverride,
  onDockChange,
  onStartDrawROI,
  onStartDraw,
  onStopDraw,
  hoveredId = null,
  onHoverIdChange,
  onHoverGroupChange,
  brushActive = false,
  brushMode = "add",
  brushSizePx = 24,
  onStartBrushAdd,
  onStartBrushErase,
  onStopBrush,
  onBrushSizeChange,
  studyUid,
  annotationSettings,
  annotationSettingsLoading,
  annotationSettingsError,
  annotationDisplaySettings,
  onUpdateAnnotationDisplaySettings,
  onResetAnnotationDisplaySettings,
  slideMpp,
  onMergeAnnotations,
  onSimplifyAnnotation,
  onPreviewSimplification,
  onPreviewMerge,
  onClearPreview,
}: AnnotationPanelProps) {
  const navigate = useNavigate();

  // Use annotation settings passed from parent (or fallback to API call if not provided)
  const {
    data: fallbackAnnotationSettings,
    isLoading: fallbackSettingsLoading,
    error: fallbackSettingsError,
  } = useStudyAnnotationSettings(annotationSettings ? undefined : studyUid);

  // Use passed props or fallback to API call
  const settingsLoading = annotationSettingsLoading ?? fallbackSettingsLoading;
  const settingsError = annotationSettingsError ?? fallbackSettingsError;
  const finalAnnotationSettings =
    annotationSettings ?? fallbackAnnotationSettings;
  const [toolMode, setToolMode] = useState<"pointer" | "brush">("pointer");
  const [size] = useState({ width: 320, height: 560 });
  const [collapsedGeom, setCollapsedGeom] = useState<{
    polygons: boolean;
    boxes: boolean;
    points: boolean;
  }>({ polygons: false, boxes: false, points: false });
  // Track which label groups are collapsed within each geometry type
  const [collapsedLabelGroups, setCollapsedLabelGroups] = useState<
    Record<string, boolean>
  >({});
  const [instructionsCollapsed, setInstructionsCollapsed] = useState(false);
  const [settingsCollapsed, setSettingsCollapsed] = useState(true);

  // Simplify dialog state
  const [simplifyDialogOpen, setSimplifyDialogOpen] = useState(false);
  const [simplifyAnnotationId, setSimplifyAnnotationId] = useState<
    string | null
  >(null);
  const [simplifyTolerance, setSimplifyTolerance] = useState(100.0); // Default 100 micrometers
  const [isEditingTolerance, setIsEditingTolerance] = useState(false);
  const [toleranceInputValue, setToleranceInputValue] = useState("100.0");
  const [previewGeometry, setPreviewGeometry] = useState<
    Polygon | MultiPolygon | null
  >(null);
  const [simplifyStats, setSimplifyStats] = useState<{
    originalPoints: number;
    simplifiedPoints: number;
    reduction: number;
    reductionPercent: string;
    toleranceUsedMicrometers: number;
  } | null>(null);

  // Dialog drag state
  const [dialogPosition, setDialogPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const dialogRef = useRef<HTMLDivElement>(null);

  // Merge preview dialog state
  const [mergePreviewOpen, setMergePreviewOpen] = useState(false);
  const [mergeAnnotationIds, setMergeAnnotationIds] = useState<string[]>([]);
  const [mergePreviewGeometry, setMergePreviewGeometry] = useState<
    Polygon | MultiPolygon | null
  >(null);

  // Merge dialog drag state
  const [mergeDialogPosition, setMergeDialogPosition] = useState({
    x: 0,
    y: 0,
  });
  const [isMergeDragging, setIsMergeDragging] = useState(false);
  const [mergeDragOffset, setMergeDragOffset] = useState({ x: 0, y: 0 });
  const mergeDialogRef = useRef<HTMLDivElement>(null);

  const panelRef = useRef<HTMLDivElement>(null);

  // Generate initial preview when simplify dialog opens
  useEffect(() => {
    if (simplifyDialogOpen && simplifyAnnotationId && onPreviewSimplification) {
      const result = onPreviewSimplification(
        simplifyAnnotationId,
        simplifyTolerance
      );
      if (result) {
        setPreviewGeometry(result.geometry);
        setSimplifyStats({
          originalPoints: result.originalPoints,
          simplifiedPoints: result.simplifiedPoints,
          reduction: result.reduction,
          reductionPercent: result.reductionPercent,
          toleranceUsedMicrometers: result.toleranceUsedMicrometers,
        });
      } else {
        setPreviewGeometry(null);
        setSimplifyStats(null);
      }
    }
  }, [simplifyDialogOpen, simplifyAnnotationId, onPreviewSimplification]);

  // Generate merge preview when merge dialog opens
  useEffect(() => {
    if (mergePreviewOpen && mergeAnnotationIds.length > 0 && onPreviewMerge) {
      const preview = onPreviewMerge(mergeAnnotationIds);
      setMergePreviewGeometry(preview);
    }
  }, [mergePreviewOpen, mergeAnnotationIds, onPreviewMerge]);

  // Handle dialog dragging
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (!dialogRef.current) return;

    const rect = dialogRef.current.getBoundingClientRect();
    setIsDragging(true);
    setDragOffset({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    });
  }, []);

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !dialogRef.current) return;

      const newX = e.clientX - dragOffset.x;
      const newY = e.clientY - dragOffset.y;

      // Keep dialog within viewport bounds
      const maxX = window.innerWidth - dialogRef.current.offsetWidth;
      const maxY = window.innerHeight - dialogRef.current.offsetHeight;

      setDialogPosition({
        x: Math.max(0, Math.min(newX, maxX)),
        y: Math.max(0, Math.min(newY, maxY)),
      });
    },
    [isDragging, dragOffset]
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  // Handle merge dialog dragging
  const handleMergeMouseDown = useCallback((e: React.MouseEvent) => {
    if (!mergeDialogRef.current) return;

    const rect = mergeDialogRef.current.getBoundingClientRect();
    setIsMergeDragging(true);
    setMergeDragOffset({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    });
  }, []);

  const handleMergeMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isMergeDragging || !mergeDialogRef.current) return;

      const newX = e.clientX - mergeDragOffset.x;
      const newY = e.clientY - mergeDragOffset.y;

      // Keep dialog within viewport bounds
      const maxX = window.innerWidth - mergeDialogRef.current.offsetWidth;
      const maxY = window.innerHeight - mergeDialogRef.current.offsetHeight;

      setMergeDialogPosition({
        x: Math.max(0, Math.min(newX, maxX)),
        y: Math.max(0, Math.min(newY, maxY)),
      });
    },
    [isMergeDragging, mergeDragOffset]
  );

  const handleMergeMouseUp = useCallback(() => {
    setIsMergeDragging(false);
  }, []);

  // Add event listeners for dragging
  useEffect(() => {
    if (isDragging) {
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
      return () => {
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  // Add event listeners for merge dialog dragging
  useEffect(() => {
    if (isMergeDragging) {
      document.addEventListener("mousemove", handleMergeMouseMove);
      document.addEventListener("mouseup", handleMergeMouseUp);
      return () => {
        document.removeEventListener("mousemove", handleMergeMouseMove);
        document.removeEventListener("mouseup", handleMergeMouseUp);
      };
    }
  }, [isMergeDragging, handleMergeMouseMove, handleMergeMouseUp]);

  // Reset dialog position when opening
  useEffect(() => {
    if (simplifyDialogOpen) {
      setDialogPosition({ x: 100, y: 100 }); // Start in a reasonable position
    }
  }, [simplifyDialogOpen]);

  // Reset merge dialog position when opening
  useEffect(() => {
    if (mergePreviewOpen) {
      setMergeDialogPosition({ x: 150, y: 150 }); // Start slightly offset from simplify dialog
    }
  }, [mergePreviewOpen]);

  // Clear preview when dialogs close
  useEffect(() => {
    if (!simplifyDialogOpen && !mergePreviewOpen && onClearPreview) {
      onClearPreview();
    }
  }, [simplifyDialogOpen, mergePreviewOpen, onClearPreview]);

  // Helper function to convert micrometers to pixels for display
  const micrometersToPixels = useCallback(
    (micrometers: number): number => {
      if (!slideMpp || slideMpp <= 0) {
        // Fallback: assume 0.25 micrometers per pixel (typical for 40x magnification)
        return micrometers / 0.25;
      }
      return micrometers / slideMpp;
    },
    [slideMpp]
  );

  // Helper function to get display text for point size
  const getPointSizeDisplayText = useCallback(
    (micrometers: number): string => {
      const pixels = micrometersToPixels(micrometers);
      return `${micrometers}μm (≈${Math.round(pixels)}px)`;
    },
    [micrometersToPixels]
  );

  // Create labels from annotation settings
  const LABELS = useMemo(() => {
    if (!finalAnnotationSettings?.annotations) {
      return [];
    }
    return createLabelsFromSettings(finalAnnotationSettings.annotations);
  }, [finalAnnotationSettings]);

  // Create default tool mapping
  const DEFAULT_TOOL_FOR_LABEL = useMemo(() => {
    const mapping: Record<string, DrawingMode> = {};
    LABELS.forEach((label) => {
      mapping[label.id] = getDefaultToolForType(label.hint);
    });
    return mapping;
  }, [LABELS]);

  // Keyboard shortcuts: 1/2/3/etc to switch labels (up to 9 labels)
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!isOpen) return;
      const target = e.target as HTMLElement | null;
      if (target && /input|textarea/i.test(target.tagName)) return;

      const keyNum = parseInt(e.key);
      if (keyNum >= 1 && keyNum <= 9 && keyNum <= LABELS.length) {
        const labelIndex = keyNum - 1;
        onActiveLabelChange?.(LABELS[labelIndex]?.id || null);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [isOpen, onActiveLabelChange, LABELS]);

  if (!isOpen) return null;

  const handleClose = () => {
    // Stop drawing when closing the panel
    onStopDraw?.();
    onStopBrush?.();
    onClose();
  };

  // Classify annotations by drawing kind (with fallback based on label settings)
  const { polygons, boxes, points } = useMemo(() => {
    const withFallback = annotations.map((a) => {
      // Find the label configuration to determine default kind
      const labelConfig = LABELS.find((l) => l.id === a.name);
      const defaultKind = labelConfig?.hint || "polygon";

      return {
        visible: true,
        ...a,
        kind: a.kind ?? defaultKind,
      };
    });
    return {
      polygons: withFallback.filter((a) => a.kind === "polygon"),
      boxes: withFallback.filter((a) => a.kind === "box"),
      points: withFallback.filter((a) => a.kind === "point"),
    };
  }, [annotations, LABELS]);

  const isPolygonLabel = (label: LabelId | null): boolean => {
    if (!label) return false;
    const labelConfig = LABELS.find((l) => l.id === label);
    return labelConfig?.hint === "polygon";
  };

  const isBrushAllowed = (label: LabelId | null): boolean => {
    if (label === null) return true; // Allow brush when no label is selected
    return isPolygonLabel(label);
  };

  // Keep the active tool in sync with the selected label and mode
  useEffect(() => {
    if (!activeLabel) {
      onStopDraw?.();
      onStopBrush?.();
      return;
    }
    if (toolMode === "brush") {
      if (isBrushAllowed(activeLabel)) {
        onStopDraw?.();
        onStartBrushAdd?.();
      } else {
        // Invalid state: brush on non-allowed label → fall back to pointer
        setToolMode("pointer");
        onStopBrush?.();
        if (activeLabel) startDrawForLabel(activeLabel);
      }
    } else {
      onStopBrush?.();
      if (activeLabel) startDrawForLabel(activeLabel);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeLabel, toolMode]);

  const toggleVisible = (id: string) => {
    const next = annotations.map((a) =>
      a.id === id ? { ...a, visible: !a.visible } : a
    );
    onUpdateAnnotations?.(next);
  };
  const removeItem = (id: string) =>
    onUpdateAnnotations?.(annotations.filter((a) => a.id !== id));

  // Group consecutive annotations by label within a list
  const groupConsecutiveByLabel = (items: AnnotationItem[]) => {
    if (items.length === 0) return [];

    const groups: Array<{ label: LabelId; items: AnnotationItem[] }> = [];
    let currentGroup: { label: LabelId; items: AnnotationItem[] } | null = null;

    for (const item of items) {
      if (!currentGroup || currentGroup.label !== item.name) {
        // Start new group
        currentGroup = { label: item.name, items: [item] };
        groups.push(currentGroup);
      } else {
        // Add to current group
        currentGroup.items.push(item);
      }
    }

    return groups;
  };

  const PanelHeader = (
    <SlidePanel.Header title="Annotations" onClose={handleClose} />
  );

  const startDrawForLabel = (label: LabelId) => {
    const mode = DEFAULT_TOOL_FOR_LABEL[label] as DrawingMode;
    if (!mode) return;
    onStartDraw?.({ label, mode });
    onStartDrawROI?.(mode, label);
  };

  const LabelPill = ({
    id,
    name,
    color,
    hint,
  }: {
    id: LabelId;
    name: string;
    color: string;
    hint?: string;
  }) => (
    <button
      type="button"
      onClick={() => {
        if (activeLabel === id) {
          onStopDraw?.();
          onStopBrush?.();
          onActiveLabelChange?.(null);
          return;
        }

        onActiveLabelChange?.(id);

        // Respect the global tool mode
        if (toolMode === "brush") {
          if (isBrushAllowed(id)) {
            onStopDraw?.();
            onStartBrushAdd?.();
          } else {
            // Non-brush-allowed label cannot use brush; fall back to pointer mode
            setToolMode("pointer");
            onStopBrush?.();
            startDrawForLabel(id);
          }
        } else {
          onStopBrush?.();
          startDrawForLabel(id);
        }
      }}
      className={cls(
        "w-full flex items-center justify-between px-1.5 py-1 rounded cursor-pointer transition-colors",
        id === activeLabel
          ? "bg-primary text-primary-foreground hover:bg-primary/90"
          : "hover:bg-accent"
      )}
      aria-pressed={id === activeLabel}
      title={`${name}${hint ? ` • ${hint}` : ""}`}
    >
      <span className="flex items-center space-x-2 min-w-0 flex-1">
        <span
          className="h-2.5 w-2.5 rounded-sm flex-shrink-0"
          style={{ backgroundColor: color }}
        />
        <span
          className={cls(
            "text-sm truncate",
            id === activeLabel
              ? "text-primary-foreground font-medium"
              : "text-foreground"
          )}
        >
          {name}
        </span>
        <small
          className={cls(
            id === activeLabel
              ? "text-primary-foreground/80"
              : "text-muted-foreground"
          )}
        >
          ({DEFAULT_TOOL_FOR_LABEL[id] || "polygon"})
        </small>
      </span>
    </button>
  );

  const getLabelMeta = (id: LabelId) => {
    const meta = LABELS.find((l) => l.id === id);
    if (!meta) {
      // Fallback for missing label metadata
      return {
        id,
        name: id.charAt(0).toUpperCase() + id.slice(1),
        color: "#6b7280", // Default gray color
        hint: "polygon" as DrawingMode,
      };
    }
    return meta;
  };

  const LabelSubGroup = ({
    label,
    items,
    geometryType,
  }: {
    label: LabelId;
    items: AnnotationItem[];
    geometryType: string;
  }) => {
    const groupKey = `${geometryType}-${label}`;
    const isCollapsed = collapsedLabelGroups[groupKey] ?? false;
    const labelMeta = getLabelMeta(label);

    const toggleCollapse = () => {
      setCollapsedLabelGroups((prev) => ({
        ...prev,
        [groupKey]: !prev[groupKey],
      }));
    };

    // If only one item, don't show grouping - render directly
    if (items.length === 1) {
      const a = items[0];
      return (
        <li
          key={a.id}
          className={cls(
            "flex items-center justify-between px-2 py-1",
            selectedId === a.id && "bg-sky-600/30",
            hoveredId === a.id && selectedId !== a.id && "bg-amber-500/25"
          )}
          onMouseEnter={() => onHoverIdChange?.(a.id)}
          onMouseLeave={() => onHoverIdChange?.(null)}
        >
          <button
            onClick={() => onSelect?.(a.id)}
            className="flex min-w-0 items-center gap-2 text-left"
          >
            <span
              className="h-2.5 w-2.5 rounded-sm"
              style={{ backgroundColor: labelMeta.color }}
            />
            <span className="truncate text-sm text-foreground">
              {a.name
                ? a.name.charAt(0).toUpperCase() + a.name.slice(1)
                : `Annotation ${a.id.slice(0, 4)}`}
            </span>
          </button>
          <div className="flex items-center gap-1">
            {a.kind === "polygon" && onSimplifyAnnotation && (
              <button
                className="p-1 rounded hover:bg-accent text-muted-foreground"
                title="Simplify polygon geometry"
                aria-label="Simplify polygon"
                onClick={(e) => {
                  e.stopPropagation();
                  setSimplifyAnnotationId(a.id);
                  setSimplifyTolerance(100.0);
                  setPreviewGeometry(null);
                  setSimplifyStats(null);
                  setSimplifyDialogOpen(true);
                }}
              >
                <SparklesIcon className="h-4 w-4" />
              </button>
            )}
            <button
              className="p-1 rounded hover:bg-accent text-muted-foreground"
              title={a.visible === false ? "Show" : "Hide"}
              aria-label={a.visible === false ? "Show" : "Hide"}
              onClick={(e) => {
                e.stopPropagation();
                toggleVisible(a.id);
              }}
            >
              {a.visible === false ? (
                <EyeIcon className="h-4 w-4" />
              ) : (
                <EyeSlashIcon className="h-4 w-4" />
              )}
            </button>
            <button
              className="p-1 rounded hover:bg-accent text-muted-foreground"
              title="Delete"
              aria-label="Delete"
              onClick={(e) => {
                e.stopPropagation();
                removeItem(a.id);
              }}
            >
              <TrashIcon className="h-4 w-4" />
            </button>
          </div>
        </li>
      );
    }

    // Multiple items - show as collapsible group
    return (
      <>
        <li className="border-b border-border700/30">
          <button
            className="w-full flex items-center justify-between px-2 py-1.5 hover:bg-accent/50"
            onClick={toggleCollapse}
            onMouseEnter={() =>
              onHoverGroupChange?.(items.map((item) => item.id))
            }
            onMouseLeave={() => onHoverGroupChange?.(null)}
          >
            <div className="flex items-center gap-2">
              {isCollapsed ? (
                <ChevronRightIcon className="h-3 w-3 text-muted-foreground400" />
              ) : (
                <ChevronDownIcon className="h-3 w-3 text-muted-foreground400" />
              )}
              <span
                className="h-2.5 w-2.5 rounded-sm"
                style={{ backgroundColor: labelMeta.color }}
              />
              <span className="text-sm text-foreground">{labelMeta.name}</span>
            </div>
            <div className="flex items-center gap-1">
              {geometryType === "polygons" &&
                items.length >= 2 &&
                items.every((item) => item.kind === "polygon") &&
                onMergeAnnotations && (
                  <button
                    className="p-1 rounded hover:bg-accent text-muted-foreground hover:text-foreground transition-colors"
                    title={`Merge ${items.length} "${labelMeta.name}" polygons into a single polygon using geometric union`}
                    onClick={(e) => {
                      e.stopPropagation();
                      setMergeAnnotationIds(items.map((item) => item.id));
                      setMergePreviewGeometry(null);
                      setMergePreviewOpen(true);
                    }}
                  >
                    <ArrowsPointingInIcon className="h-3 w-3" />
                  </button>
                )}
              <span className="text-xs text-muted-foreground400">
                {items.length}
              </span>
            </div>
          </button>
        </li>
        {!isCollapsed &&
          items.map((a) => (
            <li
              key={a.id}
              className={cls(
                "flex items-center justify-between px-4 py-1", // Indented with px-4
                selectedId === a.id && "bg-sky-600/30",
                hoveredId === a.id && selectedId !== a.id && "bg-amber-500/25"
              )}
              onMouseEnter={() => onHoverIdChange?.(a.id)}
              onMouseLeave={() => onHoverIdChange?.(null)}
            >
              <button
                onClick={() => onSelect?.(a.id)}
                className="flex min-w-0 items-center gap-2 text-left"
              >
                <span className="truncate text-sm text-foreground">
                  {a.name
                    ? a.name.charAt(0).toUpperCase() + a.name.slice(1)
                    : `Annotation ${a.id.slice(0, 4)}`}
                </span>
              </button>
              <div className="flex items-center gap-1">
                {a.kind === "polygon" && onSimplifyAnnotation && (
                  <button
                    className="p-1 rounded hover:bg-accent text-muted-foreground"
                    title="Simplify polygon geometry"
                    aria-label="Simplify polygon"
                    onClick={(e) => {
                      e.stopPropagation();
                      setSimplifyAnnotationId(a.id);
                      setSimplifyTolerance(100.0);
                      setToleranceInputValue("100.0");
                      setPreviewGeometry(null);
                      setSimplifyStats(null);
                      setIsEditingTolerance(false);
                      setSimplifyDialogOpen(true);
                    }}
                  >
                    <SparklesIcon className="h-4 w-4" />
                  </button>
                )}
                <button
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  title={a.visible === false ? "Show" : "Hide"}
                  aria-label={a.visible === false ? "Show" : "Hide"}
                  onClick={(e) => {
                    e.stopPropagation();
                    toggleVisible(a.id);
                  }}
                >
                  {a.visible === false ? (
                    <EyeIcon className="h-4 w-4" />
                  ) : (
                    <EyeSlashIcon className="h-4 w-4" />
                  )}
                </button>
                <button
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  title="Delete"
                  aria-label="Delete"
                  onClick={(e) => {
                    e.stopPropagation();
                    removeItem(a.id);
                  }}
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
            </li>
          ))}
      </>
    );
  };

  const GeometryGroup = ({
    title,
    items,
    collapseKey,
  }: {
    title: string;
    items: AnnotationItem[];
    collapseKey: "polygons" | "boxes" | "points";
  }) => {
    const isCollapsed = collapsedGeom[collapseKey];
    return (
      <div className="rounded-md border border-border700/60 bg-muted/60">
        <button
          className="w-full flex items-center justify-between px-2 py-1.5 hover:bg-accent"
          onClick={() =>
            setCollapsedGeom((c) => ({ ...c, [collapseKey]: !c[collapseKey] }))
          }
        >
          <div className="flex items-center gap-2">
            {isCollapsed ? (
              <ChevronRightIcon className="h-4 w-4 text-muted-foreground400" />
            ) : (
              <ChevronDownIcon className="h-4 w-4 text-muted-foreground400" />
            )}
            <span className="text-sm text-foreground">{title}</span>
          </div>
          <span className="text-xs text-muted-foreground400">
            {items.length}
          </span>
        </button>
        {!isCollapsed && items.length > 0 && (
          <ul className="divide-y divide-gray-700/60">
            {groupConsecutiveByLabel(items).map((group, groupIndex) => (
              <LabelSubGroup
                key={`${group.label}-${groupIndex}`}
                label={group.label}
                items={group.items}
                geometryType={collapseKey}
              />
            ))}
          </ul>
        )}
        {!isCollapsed && items.length === 0 && (
          <div className="px-2 py-2 text-xs text-muted-foreground400">
            No annotations
          </div>
        )}
      </div>
    );
  };

  const PanelBody = (
    <>
      {/* Upper: grouped annotations by geometry (foldable) - only show groups that have items */}
      <div className="flex-1 min-h-0 overflow-y-auto px-3 py-2 space-y-2">
        {polygons.length > 0 && (
          <GeometryGroup
            title="Polygons"
            items={polygons}
            collapseKey="polygons"
          />
        )}
        {boxes.length > 0 && (
          <GeometryGroup title="Boxes" items={boxes} collapseKey="boxes" />
        )}
        {points.length > 0 && (
          <GeometryGroup title="Points" items={points} collapseKey="points" />
        )}
        {polygons.length === 0 && boxes.length === 0 && points.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No annotations yet</p>
            <p className="text-xs mt-1">Start drawing to create annotations</p>
          </div>
        )}
      </div>

      {/* Bottom: label picker + global tool mode */}
      <div className="px-3 pt-2 pb-2 border-t border-border700/60 bg-muted/70 flex-shrink-0">
        {/* Show loading state */}
        {settingsLoading && (
          <div className="text-xs text-muted-foreground400 mb-2">
            Loading annotation settings...
          </div>
        )}

        {/* Show error state */}
        {settingsError && (
          <div className="text-xs text-red-400 mb-2">
            Error loading annotation settings
          </div>
        )}

        {/* Show empty state with navigation link */}
        {!settingsLoading &&
          !settingsError &&
          (!finalAnnotationSettings?.allowAnnotation ||
            LABELS.length === 0) && (
            <div className="text-xs text-muted-foreground400 mb-2 text-center">
              <p className="mb-2">
                No labels defined.{" "}
                <button
                  onClick={() => {
                    if (studyUid) {
                      navigate({
                        to: `/studies/${studyUid}/settings/annotations`,
                      });
                    }
                  }}
                  className="text-blue-400 hover:text-blue-300 underline"
                >
                  Configure
                </button>
              </p>
            </div>
          )}

        {/* Label list - only show if we have labels */}
        {!settingsLoading &&
          !settingsError &&
          finalAnnotationSettings?.allowAnnotation &&
          LABELS.length > 0 && (
            <>
              <div className="flex items-center justify-between text-xs text-muted-foreground400 mb-1">
                <span>Label for new annotations</span>
                {studyUid && (
                  <button
                    onClick={() => {
                      navigate({
                        to: `/studies/${studyUid}/settings/annotations`,
                      });
                    }}
                    className="inline-flex items-center text-muted-foreground hover:text-foreground transition-colors"
                    title="Configure annotation settings"
                  >
                    <Cog6ToothIcon className="h-3 w-3" />
                  </button>
                )}
              </div>
              <div className="flex flex-col space-y-0.5 max-h-36 overflow-y-auto relative">
                {LABELS.map((l) => (
                  <LabelPill key={l.id} {...l} />
                ))}
              </div>
            </>
          )}

        {/* Global Tool Mode */}
        <div className="mt-2 pt-2 border-t border-border700/60">
          <div className="text-xs text-muted-foreground400 mb-2">Tool Mode</div>
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => {
                setToolMode("pointer");
                onStopBrush?.();
                if (activeLabel) startDrawForLabel(activeLabel);
              }}
              className={cls(
                "flex-1 px-2 py-1 text-xs rounded transition-colors border",
                toolMode === "pointer"
                  ? "bg-primary/20 border-primary/60 text-primary"
                  : "hover:bg-accent border-border700/30 text-foreground"
              )}
              title="Draw the default geometry for the selected label"
            >
              Pointer
            </button>

            <button
              type="button"
              disabled={!isBrushAllowed(activeLabel)}
              onClick={() => {
                if (!isBrushAllowed(activeLabel)) return;
                setToolMode("brush");
                onStopDraw?.();
                onStartBrushAdd?.();
              }}
              className={cls(
                "flex-1 px-2 py-1 text-xs rounded transition-colors border",
                !isBrushAllowed(activeLabel)
                  ? "opacity-50 cursor-not-allowed border-border700/20 text-muted-foreground"
                  : toolMode === "brush"
                  ? "bg-emerald-600/15 border-emerald-500/60 text-emerald-400"
                  : "hover:bg-accent border-border700/30 text-foreground"
              )}
              title={
                isBrushAllowed(activeLabel)
                  ? "Brush-edit polygons (add/erase) or create new polygons"
                  : "Brush is only available for polygon labels and when no label is selected"
              }
            >
              Brush
            </button>
          </div>

          {/* Brush controls (only when brush mode is active AND brush is allowed) */}
          {toolMode === "brush" && isBrushAllowed(activeLabel) && (
            <div className="mt-2 space-y-2">
              <div className="flex gap-1">
                <button
                  onClick={onStartBrushAdd}
                  className={cls(
                    "flex-1 px-2 py-1 text-xs rounded transition-colors border",
                    brushActive && brushMode === "add"
                      ? "bg-emerald-600 text-white border-emerald-400 shadow-md"
                      : brushMode === "add" && !brushActive
                      ? "bg-emerald-600/40 text-emerald-100 border-emerald-500/60"
                      : "bg-emerald-600/20 text-emerald-400 border-emerald-600/30 hover:bg-emerald-600/30 hover:border-emerald-500/50"
                  )}
                  title="Paint to add area to selected polygon or create a new polygon"
                >
                  {brushActive && brushMode === "add" ? "● Add" : "Add"}
                </button>
                <button
                  onClick={onStartBrushErase}
                  className={cls(
                    "flex-1 px-2 py-1 text-xs rounded transition-colors border",
                    brushActive && brushMode === "erase"
                      ? "bg-red-600 text-white border-red-400 shadow-md"
                      : brushMode === "erase" && !brushActive
                      ? "bg-red-600/40 text-red-100 border-red-500/60"
                      : "bg-red-600/20 text-red-400 border-red-600/30 hover:bg-red-600/30 hover:border-red-500/50"
                  )}
                  title={
                    selectedId &&
                    annotations.find((a) => a.id === selectedId)?.kind ===
                      "polygon"
                      ? "Paint to remove area from the selected polygon, or paint from outside to 'eat' parts from edges"
                      : "Paint near polygon edges to 'eat' parts from outside, or select a polygon first to erase from inside"
                  }
                >
                  {brushActive && brushMode === "erase"
                    ? "● Erase"
                    : brushMode === "erase" &&
                      !(
                        selectedId &&
                        annotations.find((a) => a.id === selectedId)?.kind ===
                          "polygon"
                      )
                    ? "Erase (select polygon)"
                    : "Erase"}
                </button>
              </div>

              {/* Size slider only when Brush is active */}
              {brushActive && (
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground400 flex-shrink-0">
                    Size:
                  </span>
                  <input
                    type="range"
                    min={4}
                    max={96}
                    value={brushSizePx}
                    onChange={(e) =>
                      onBrushSizeChange?.(Number(e.target.value))
                    }
                    className="flex-1 h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                    title={`Brush size: ${brushSizePx}px`}
                  />
                  <span className="text-xs text-muted-foreground400 flex-shrink-0 w-6">
                    {brushSizePx}
                  </span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Annotation Display Settings - collapsible */}
        {annotationDisplaySettings && onUpdateAnnotationDisplaySettings && (
          <PanelFoldableBody
            title="Display Settings"
            icon={Cog6ToothIcon}
            isCollapsed={settingsCollapsed}
            onToggleCollapsed={() => setSettingsCollapsed(!settingsCollapsed)}
          >
            <div className="space-y-3">
              {/* Stroke Width */}
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground400">
                  Stroke Width: {annotationDisplaySettings.strokeWidth}px
                </label>
                <input
                  type="range"
                  min={1}
                  max={10}
                  step={1}
                  value={annotationDisplaySettings.strokeWidth}
                  onChange={(e) =>
                    onUpdateAnnotationDisplaySettings({
                      strokeWidth: Number(e.target.value),
                    })
                  }
                  className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                />
              </div>

              {/* Fill Opacity */}
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground400">
                  Fill Opacity:{" "}
                  {Math.round(annotationDisplaySettings.fillOpacity * 100)}%
                </label>
                <input
                  type="range"
                  min={0}
                  max={1}
                  step={0.05}
                  value={annotationDisplaySettings.fillOpacity}
                  onChange={(e) =>
                    onUpdateAnnotationDisplaySettings({
                      fillOpacity: Number(e.target.value),
                    })
                  }
                  className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                />
              </div>

              {/* Point Size */}
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground400">
                  Point Size:{" "}
                  {getPointSizeDisplayText(
                    annotationDisplaySettings.pointSizeMicrometers
                  )}
                </label>
                <input
                  type="range"
                  min={1}
                  max={50}
                  step={0.5}
                  value={annotationDisplaySettings.pointSizeMicrometers}
                  onChange={(e) =>
                    onUpdateAnnotationDisplaySettings({
                      pointSizeMicrometers: Number(e.target.value),
                    })
                  }
                  className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                />
              </div>

              {/* Reset Button */}
              {onResetAnnotationDisplaySettings && (
                <button
                  onClick={onResetAnnotationDisplaySettings}
                  className="text-xs text-muted-foreground400 hover:text-foreground transition-colors underline"
                >
                  Reset to defaults
                </button>
              )}
            </div>
          </PanelFoldableBody>
        )}

        {/* Drawing instructions - collapsible */}
        <PanelFoldableBody
          title="Instructions"
          icon={QuestionMarkCircleIcon}
          isCollapsed={instructionsCollapsed}
          onToggleCollapsed={() =>
            setInstructionsCollapsed(!instructionsCollapsed)
          }
        >
          <ul className="text-xs text-muted-foreground400/70 space-y-1">
            <li>
              • <strong>Points:</strong> Single click
            </li>
            <li>
              • <strong>Boxes:</strong> Click two opposite corners
            </li>
            <li>
              • <strong>Polygons:</strong> Drag for curves, click for straight
              lines
            </li>
            <li>
              • <strong>Brush Add:</strong> Paint to create new polygons or
              expand existing ones
            </li>
            <li>
              • <strong>Brush Erase:</strong> Paint inside to remove areas, or
              paint from outside to "eat" edges
            </li>
          </ul>
        </PanelFoldableBody>
      </div>

      {/* Simplify Dialog */}
      {simplifyDialogOpen && simplifyAnnotationId && (
        <div className="fixed inset-0 bg-black/30 z-50">
          <div
            ref={dialogRef}
            className="absolute bg-background border border-border rounded-lg shadow-lg w-80"
            style={{
              left: dialogPosition.x,
              top: dialogPosition.y,
              cursor: isDragging ? "grabbing" : "default",
            }}
          >
            <div
              className="px-4 py-3 border-b border-border cursor-grab active:cursor-grabbing select-none"
              onMouseDown={handleMouseDown}
            >
              <h3 className="text-lg font-medium">Simplify Polygon</h3>
            </div>
            <div className="p-4">
              <div className="space-y-4">
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <label className="text-sm text-muted-foreground">
                      Tolerance:
                    </label>
                    {isEditingTolerance ? (
                      <input
                        type="text"
                        value={toleranceInputValue}
                        onChange={(e) => {
                          const value = e.target.value;
                          // Allow typing numbers and decimal points, but don't validate yet
                          if (/^\d*\.?\d*$/.test(value) || value === "") {
                            setToleranceInputValue(value);
                          }
                        }}
                        onBlur={() => {
                          // Validate and apply the value when leaving the input
                          const tolerance = Math.max(
                            0.1,
                            Number(toleranceInputValue) || 100.0
                          );
                          setSimplifyTolerance(tolerance);
                          setToleranceInputValue(tolerance.toString());
                          setIsEditingTolerance(false);

                          // Update preview with final value
                          if (onPreviewSimplification && simplifyAnnotationId) {
                            try {
                              const result = onPreviewSimplification(
                                simplifyAnnotationId,
                                tolerance
                              );
                              if (result) {
                                setPreviewGeometry(result.geometry);
                                setSimplifyStats({
                                  originalPoints: result.originalPoints,
                                  simplifiedPoints: result.simplifiedPoints,
                                  reduction: result.reduction,
                                  reductionPercent: result.reductionPercent,
                                  toleranceUsedMicrometers:
                                    result.toleranceUsedMicrometers,
                                });
                              } else {
                                setPreviewGeometry(null);
                                setSimplifyStats(null);
                              }
                            } catch (error) {
                              console.error(
                                "Error generating simplification preview:",
                                error
                              );
                              setPreviewGeometry(null);
                              setSimplifyStats(null);
                            }
                          }
                        }}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            // Same logic as onBlur
                            const tolerance = Math.max(
                              0.1,
                              Number(toleranceInputValue) || 100.0
                            );
                            setSimplifyTolerance(tolerance);
                            setToleranceInputValue(tolerance.toString());
                            setIsEditingTolerance(false);

                            if (
                              onPreviewSimplification &&
                              simplifyAnnotationId
                            ) {
                              try {
                                const result = onPreviewSimplification(
                                  simplifyAnnotationId,
                                  tolerance
                                );
                                if (result) {
                                  setPreviewGeometry(result.geometry);
                                  setSimplifyStats({
                                    originalPoints: result.originalPoints,
                                    simplifiedPoints: result.simplifiedPoints,
                                    reduction: result.reduction,
                                    reductionPercent: result.reductionPercent,
                                    toleranceUsedMicrometers:
                                      result.toleranceUsedMicrometers,
                                  });
                                } else {
                                  setPreviewGeometry(null);
                                  setSimplifyStats(null);
                                }
                              } catch (error) {
                                console.error(
                                  "Error generating simplification preview:",
                                  error
                                );
                                setPreviewGeometry(null);
                                setSimplifyStats(null);
                              }
                            }
                          }
                          if (e.key === "Escape") {
                            // Cancel editing and restore original value
                            setToleranceInputValue(
                              simplifyTolerance.toString()
                            );
                            setIsEditingTolerance(false);
                          }
                        }}
                        autoFocus
                        className="w-20 px-2 py-1 text-sm border border-border700/60 rounded bg-muted/60 text-foreground focus:bg-background focus:border-primary/60 focus:outline-none"
                      />
                    ) : (
                      <button
                        onClick={() => {
                          setToleranceInputValue(simplifyTolerance.toString());
                          setIsEditingTolerance(true);
                        }}
                        className="w-20 px-2 py-1 text-sm border border-border700/60 rounded bg-muted/60 hover:bg-muted/80 text-foreground text-left transition-colors"
                      >
                        {simplifyTolerance.toFixed(1)}
                      </button>
                    )}
                    <span className="text-sm text-muted-foreground">μm</span>
                  </div>
                  <input
                    type="range"
                    min={50.0}
                    max={5000.0}
                    step={10.0}
                    value={simplifyTolerance}
                    onChange={(e) => {
                      const tolerance = Number(e.target.value);
                      setSimplifyTolerance(tolerance);
                      setToleranceInputValue(tolerance.toString());

                      // Update preview
                      if (onPreviewSimplification && simplifyAnnotationId) {
                        try {
                          const result = onPreviewSimplification(
                            simplifyAnnotationId,
                            tolerance
                          );
                          if (result) {
                            setPreviewGeometry(result.geometry);
                            setSimplifyStats({
                              originalPoints: result.originalPoints,
                              simplifiedPoints: result.simplifiedPoints,
                              reduction: result.reduction,
                              reductionPercent: result.reductionPercent,
                              toleranceUsedMicrometers:
                                result.toleranceUsedMicrometers,
                            });
                          } else {
                            setPreviewGeometry(null);
                            setSimplifyStats(null);
                          }
                        } catch (error) {
                          console.error(
                            "Error generating simplification preview:",
                            error
                          );
                          setPreviewGeometry(null);
                          setSimplifyStats(null);
                        }
                      }
                    }}
                    className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                  />
                  <div className="flex justify-between text-xs text-muted-foreground mt-1">
                    <span>50μm (Fine)</span>
                    <span>5000μm (5mm - Heavy)</span>
                  </div>

                  {/* Quick preset buttons */}
                  <div className="flex flex-wrap gap-1 mt-2">
                    {[50, 100, 250, 500].map((preset) => (
                      <button
                        key={preset}
                        onClick={() => {
                          setSimplifyTolerance(preset);
                          setToleranceInputValue(preset.toString());
                          // Trigger preview update
                          if (onPreviewSimplification && simplifyAnnotationId) {
                            try {
                              const result = onPreviewSimplification(
                                simplifyAnnotationId,
                                preset
                              );
                              if (result) {
                                setPreviewGeometry(result.geometry);
                                setSimplifyStats({
                                  originalPoints: result.originalPoints,
                                  simplifiedPoints: result.simplifiedPoints,
                                  reduction: result.reduction,
                                  reductionPercent: result.reductionPercent,
                                  toleranceUsedMicrometers:
                                    result.toleranceUsedMicrometers,
                                });
                              } else {
                                setPreviewGeometry(null);
                                setSimplifyStats(null);
                              }
                            } catch (error) {
                              console.error(
                                "Error generating preset simplification preview:",
                                error
                              );
                              setPreviewGeometry(null);
                              setSimplifyStats(null);
                            }
                          }
                        }}
                        className={`px-2 py-1 text-xs rounded border transition-colors ${
                          Math.abs(simplifyTolerance - preset) < 0.1
                            ? "bg-primary/20 border-primary/60 text-primary"
                            : "border-border hover:bg-accent"
                        }`}
                      >
                        {preset}μm
                      </button>
                    ))}
                  </div>
                </div>

                {previewGeometry && simplifyStats ? (
                  <div className="text-sm text-muted-foreground">
                    <div className="p-2 bg-muted/50 rounded space-y-1">
                      <p className="text-xs font-medium">
                        Preview: Simplified geometry ready
                      </p>
                      <p className="text-xs">
                        Points: {simplifyStats.originalPoints} →{" "}
                        {simplifyStats.simplifiedPoints}
                        <span className="text-amber-600 ml-1">
                          (-{simplifyStats.reduction},{" "}
                          {simplifyStats.reductionPercent}% reduction)
                        </span>
                      </p>
                      <p className="text-xs">
                        Area: {previewGeometry.getArea().toFixed(2)} px²
                      </p>
                      <p className="text-xs">
                        Tolerance used:{" "}
                        {simplifyStats.toleranceUsedMicrometers.toFixed(2)}μm
                      </p>
                      <p className="text-xs text-amber-600">
                        ✓ Orange dashed outline shows the result on the map
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="text-sm text-red-400">
                    <div className="p-2 bg-red-500/10 rounded">
                      <p className="text-xs">
                        ⚠️ Cannot simplify with {simplifyTolerance.toFixed(1)}μm
                        tolerance - try a lower value
                      </p>
                    </div>
                  </div>
                )}

                <div className="flex gap-2 justify-end">
                  <button
                    onClick={() => {
                      setSimplifyDialogOpen(false);
                      setSimplifyAnnotationId(null);
                      setPreviewGeometry(null);
                      setSimplifyStats(null);
                      setIsEditingTolerance(false);
                      setToleranceInputValue("100.0");
                    }}
                    className="px-3 py-2 text-sm border border-border rounded hover:bg-accent"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      if (onSimplifyAnnotation && simplifyAnnotationId) {
                        const success = onSimplifyAnnotation(
                          simplifyAnnotationId,
                          simplifyTolerance
                        );
                        if (success) {
                          setSimplifyDialogOpen(false);
                          setSimplifyAnnotationId(null);
                          setPreviewGeometry(null);
                          setSimplifyStats(null);
                        }
                      }
                    }}
                    disabled={!previewGeometry}
                    className="px-3 py-2 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Apply
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Merge Preview Dialog */}
      {mergePreviewOpen && mergeAnnotationIds.length > 0 && (
        <div className="fixed inset-0 bg-black/30 z-50">
          <div
            ref={mergeDialogRef}
            className="absolute bg-background border border-border rounded-lg shadow-lg w-80"
            style={{
              left: mergeDialogPosition.x,
              top: mergeDialogPosition.y,
              cursor: isMergeDragging ? "grabbing" : "default",
            }}
          >
            <div
              className="px-4 py-3 border-b border-border cursor-grab active:cursor-grabbing select-none"
              onMouseDown={handleMergeMouseDown}
            >
              <h3 className="text-lg font-medium">Merge Polygons Preview</h3>
            </div>
            <div className="p-4">
              <div className="space-y-4">
                <div className="text-sm text-muted-foreground">
                  <p>
                    Merging {mergeAnnotationIds.length} polygons into a single
                    polygon.
                  </p>
                  {mergePreviewGeometry && (
                    <div className="mt-2 p-2 bg-muted/50 rounded">
                      <p className="text-xs">Preview: Merged geometry ready</p>
                      <p className="text-xs">
                        Area: {mergePreviewGeometry.getArea().toFixed(2)} px²
                      </p>
                      <p className="text-xs text-emerald-600">
                        ✓ Green dashed outline shows the result on the map
                      </p>
                    </div>
                  )}
                </div>

                <div className="flex gap-2 justify-end">
                  <button
                    onClick={() => {
                      setMergePreviewOpen(false);
                      setMergeAnnotationIds([]);
                      setMergePreviewGeometry(null);
                      onClearPreview?.();
                    }}
                    className="px-3 py-2 text-sm border border-border rounded hover:bg-accent"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      if (onMergeAnnotations && mergeAnnotationIds.length > 0) {
                        const success = onMergeAnnotations(mergeAnnotationIds);
                        if (success) {
                          setMergePreviewOpen(false);
                          setMergeAnnotationIds([]);
                          setMergePreviewGeometry(null);
                        }
                      }
                    }}
                    disabled={!mergePreviewGeometry}
                    className="px-3 py-2 text-sm bg-emerald-600 text-white rounded hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Merge
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );

  // Render docked-left or floating
  return (
    <SlidePanel
      isOpen={isOpen}
      onClose={handleClose}
      dockOverride={dockOverride}
      onDockChange={onDockChange}
      storageKey={`annotationPanel`}
      defaultSize={size}
    >
      {PanelHeader}
      {PanelBody}
    </SlidePanel>
  );
}

export default AnnotationPanel;
