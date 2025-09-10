// Copyright 2025 Jonas Teuwen. All rights reserved.
// This file is part of SlideInsight.
// Use of this source code is governed by the terms found in the LICENSE file.

import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import Map from "ol/Map";
import { useNavigate } from "@tanstack/react-router";
import {
  ArrowDownTrayIcon,
  ArrowUpTrayIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  CpuChipIcon,
  EyeIcon,
  EyeSlashIcon,
  InformationCircleIcon,
  UserIcon,
} from "@heroicons/react/24/outline";

import SlidePanel from "@/components/ui/slide-panel";
import { Input } from "@/components/ui/input";

import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { apiFetch } from "@/utils/fetchUtils";
import {
  useStudyAnnotationSettings,
  type StudyAnnotationLabel,
} from "@/hooks/useStudyAnnotationSettings";
import { useMaskContext } from "@/features/viewer/contexts/MaskContext";
import { useVectorContext } from "@/features/viewer/contexts/VectorContext";
import {
  vectorAnnotationsService,
  type AnnotationImportResult,
} from "@/services/vectorAnnotationsService";

// Types
interface EnhancedMaskAnnotation {
  maskUid: string;
  maskName: string;
  tilesUrl: string;
  slideUid: string;
  maskWidth: number;
  maskHeight: number;
  maskMpp: number;
  actorType?: "user" | "model";
  actorId?: number;
  mutable?: boolean;
  modelName?: string;
  labels?: Array<{ name: string; index: number; color: string }>;
  createdAt?: string;
  deletedAt?: string;
  opacity?: number;
}
interface EnhancedVectorAnnotation {
  vectorUid: string;
  vectorName: string;
  slideUid: string;
  fileUri?: string;
  dataBlob?: any;
  actorType?: "user" | "model";
  actorId?: number;
  mutable?: boolean;
  modelName?: string;
  labels?: Array<{ name: string; color: string }>;
  createdAt?: string;
  deletedAt?: string;
  opacity?: number;
}
interface MaskList {
  slideUid: string;
  masks: EnhancedMaskAnnotation[];
}
interface VectorAnnotationList {
  slideUid: string;
  annotations: EnhancedVectorAnnotation[];
}

interface AnnotationControlPanelProps {
  onClose: () => void;
  mapRef?: Map | null;
  slideUid?: string;
  studyUid?: string;
  dockOverride?: "free" | "left";
  onDockChange?: (dock: "free" | "left") => void;
  openOverride?: boolean;
  onOpenChange?: (open: boolean) => void;
  annotationSettings?: any;
  annotationSettingsLoading?: boolean;
  annotationSettingsError?: any;
  onImportComplete?: (result: AnnotationImportResult) => void;
}

// Helpers
const createLabelsFromSettings = (annotations: StudyAnnotationLabel[] = []) => {
  return (annotations || [])
    .filter((a) => a && a.label && a.color && a.type)
    .map((a) => ({
      id: a.label,
      name: a.label.charAt(0).toUpperCase() + a.label.slice(1),
      color: a.color,
      hint: a.type as "point" | "box" | "polygon",
    }));
};

export default function AnnotationControlPanel({
  onClose,
  mapRef,
  slideUid,
  studyUid,
  dockOverride,
  onDockChange,
  openOverride,
  onOpenChange,
  annotationSettings,
  annotationSettingsLoading,
  annotationSettingsError,
  onImportComplete,
}: AnnotationControlPanelProps) {
  const navigate = useNavigate();

  // Open/dock
  const [isOpen, setIsOpen] = useState(false);
  const [size, setSize] = useState({ width: 320, height: 560 });
  const [dock, setDock] = useState<"free" | "left">("left");

  // Filters/state
  const [layerFilter, setLayerFilter] = useState("");
  const [actorFilter, setActorFilter] = useState<"all" | "user" | "model">(
    "all"
  );
  const [infoCollapsed, setInfoCollapsed] = useState(true);

  // Data
  const [maskAnnotations, setMaskAnnotations] = useState<
    EnhancedMaskAnnotation[]
  >([]);
  const [vectorAnnotations, setVectorAnnotations] = useState<
    EnhancedVectorAnnotation[]
  >([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Per-item UI state
  const [collapsedGroups, setCollapsedGroups] = useState<
    Record<string, boolean>
  >({});
  const [layerOpacities, setLayerOpacities] = useState<Record<string, number>>(
    {}
  );
  const [visible, setVisible] = useState<Record<string, boolean>>({});
  const [selectedMaskId, setSelectedMaskId] = useState<string | null>(null);
  const [selectedVectorId, setSelectedVectorId] = useState<string | null>(null);

  // Feature flag (server-side filtering implemented)
  const USE_BACKEND_FILTERING = true;

  // Study annotation settings (labels)
  const {
    data: fallbackSettings,
    isLoading: fallbackLoading,
    error: fallbackError,
  } = useStudyAnnotationSettings(annotationSettings ? undefined : studyUid);

  const settingsLoading = annotationSettingsLoading ?? fallbackLoading;
  const settingsError = annotationSettingsError ?? fallbackError;
  const finalSettings = annotationSettings ?? fallbackSettings;
  const LABELS = useMemo(
    () => createLabelsFromSettings(finalSettings?.annotations),
    [finalSettings]
  );

  // Contexts
  const {
    showMask,
    setShowMask,
    maskOpacity, // global (kept for compatibility)
    setMaskOpacity,
    maskLayers,
    toggleMaskLayerVisibility,
  } = useMaskContext();

  const {
    showVectors,
    setShowVectors,
    vectorOpacity, // global (kept for compatibility)
  } = useVectorContext();

  // Fetch annotations
  const fetchAnnotations = useCallback(async () => {
    if (!slideUid) return;
    setLoading(true);
    setError(null);
    try {
      let maskUrl = `/api/v1/slides/${slideUid}/annotations/raster`;
      let vectorUrl = `/api/v1/slides/${slideUid}/annotations/vector`;
      if (USE_BACKEND_FILTERING && actorFilter !== "all") {
        const params = new URLSearchParams({ actor_type: actorFilter });
        maskUrl += `?${params.toString()}`;
        vectorUrl += `?${params.toString()}`;
      }
      const [maskRes, vecRes] = await Promise.all([
        apiFetch<MaskList>(maskUrl),
        apiFetch<VectorAnnotationList>(vectorUrl),
      ]);
      setMaskAnnotations(maskRes.masks || []);
      setVectorAnnotations(vecRes.annotations || []);

      // Note: Mask layers in context are populated by useMaskLayers.handleMaskLayerCreated
      // when the actual WebGL layer is created. We don't create them here to avoid duplicates.
    } catch (e) {
      console.error(e);
      setError("Failed to load annotations");
    } finally {
      setLoading(false);
    }
  }, [slideUid, actorFilter]);

  useEffect(() => {
    fetchAnnotations();
  }, [fetchAnnotations]);

  // Persist basic UI state
  useEffect(() => {
    const savedOpen = localStorage.getItem("annotationPanelOpen");
    if (savedOpen) setIsOpen(savedOpen === "true");
    const savedDock = localStorage.getItem("annotationPanelDock");
    if (savedDock === "free" || savedDock === "left") setDock(savedDock);
    const savedSize = localStorage.getItem("annotationPanelSize");
    if (savedSize) {
      try {
        const parsed = JSON.parse(savedSize);
        if (
          typeof parsed?.width === "number" &&
          typeof parsed?.height === "number"
        )
          setSize(parsed);
      } catch {}
    }
  }, []);

  useEffect(() => {
    const effectiveOpen = openOverride ?? isOpen;
    localStorage.setItem("annotationPanelOpen", String(effectiveOpen));
    if (mapRef && (dockOverride ?? dock) === "left") {
      setTimeout(() => mapRef.updateSize(), 100);
    }
  }, [openOverride, isOpen, mapRef, dock, dockOverride]);

  useEffect(() => {
    localStorage.setItem("annotationPanelDock", dockOverride ?? dock);
  }, [dock, dockOverride]);
  useEffect(() => {
    localStorage.setItem("annotationPanelSize", JSON.stringify(size));
  }, [size]);

  // Derived helpers
  const matchesFilter = (txt: string) =>
    txt?.toLowerCase().includes(layerFilter.trim().toLowerCase());

  const groupedMasks = useMemo(() => {
    const filtered = maskAnnotations.filter((m) =>
      matchesFilter(m.maskName || m.maskUid)
    );
    return {
      user: filtered.filter((m) => m.actorType === "user"),
      model: filtered.filter((m) => m.actorType === "model"),
      unknown: filtered.filter(
        (m) =>
          !m.actorType || (m.actorType !== "user" && m.actorType !== "model")
      ),
    } as const;
  }, [maskAnnotations, layerFilter]);

  const groupedVectors = useMemo(() => {
    const filtered = vectorAnnotations.filter((v) =>
      matchesFilter(v.vectorName || v.vectorUid)
    );
    return {
      user: filtered.filter((v) => v.actorType === "user"),
      model: filtered.filter((v) => v.actorType === "model"),
      unknown: filtered.filter(
        (v) =>
          !v.actorType || (v.actorType !== "user" && v.actorType !== "model")
      ),
    } as const;
  }, [vectorAnnotations, layerFilter]);

  const hasMaskData = maskAnnotations.length > 0;
  const hasVectorData = vectorAnnotations.length > 0;

  // UI utils
  const toggleGroup = (key: string) =>
    setCollapsedGroups((prev) => ({ ...prev, [key]: !prev[key] }));
  // Note: Mask layers don't support per-layer opacity, only global opacity
  // Individual opacity sliders show global mask opacity for consistency
  const getOpacity = (id: string) => {
    const maskLayer = maskLayers.find((layer) => layer.id === id);
    if (maskLayer) {
      return maskOpacity; // Use global mask opacity for mask layers
    }
    return layerOpacities[id] ?? 1.0; // Use per-layer opacity for vector layers
  };

  const setOpacity = (id: string, val: number) => {
    const maskLayer = maskLayers.find((layer) => layer.id === id);
    if (maskLayer) {
      // Update global mask opacity for mask layers
      setMaskOpacity(val);
    } else {
      // Update per-layer opacity for vector layers
      setLayerOpacities((p) => ({ ...p, [id]: val }));
    }
  };

  // Get visibility from mask layers context, fallback to local state for non-mask items
  const isVisible = (id: string) => {
    const maskLayer = maskLayers.find((layer) => layer.id === id);
    if (maskLayer) {
      return maskLayer.visible;
    }
    return visible[id] ?? true;
  };

  const toggleVisible = (id: string) => {
    const maskLayer = maskLayers.find((layer) => layer.id === id);
    if (maskLayer) {
      // Use context function for mask layers
      toggleMaskLayerVisibility(id);
    } else {
      // Use local state for vector layers or other items
      setVisible((p) => ({ ...p, [id]: !isVisible(id) }));
    }
  };

  const snapToTenth = (v: number) => {
    const snapped = Math.round(v * 10) / 10;
    return Math.abs(v - snapped) < 0.02 ? snapped : v;
  };

  // Download functions
  const downloadAllVectorAnnotations = useCallback(async () => {
    if (!slideUid || vectorAnnotations.length === 0) return;

    try {
      // Create a combined GeoJSON FeatureCollection with all vector annotations
      const features: any[] = [];

      for (const annotation of vectorAnnotations) {
        try {
          const geoJsonData =
            await vectorAnnotationsService.getVectorAnnotationFile(
              slideUid,
              annotation.vectorUid
            );

          if (geoJsonData && geoJsonData.features) {
            // Add annotation metadata to each feature's properties
            const annotatedFeatures = geoJsonData.features.map(
              (feature: any) => ({
                ...feature,
                properties: {
                  ...feature.properties,
                  vectorUid: annotation.vectorUid,
                  vectorName: annotation.vectorName,
                  actorType: annotation.actorType,
                  modelName: annotation.modelName,
                  mutable: annotation.mutable,
                  createdAt: annotation.createdAt,
                },
              })
            );
            features.push(...annotatedFeatures);
          }
        } catch (error) {
          console.error(
            `Failed to download annotation ${annotation.vectorUid}:`,
            error
          );
        }
      }

      const featureCollection = {
        type: "FeatureCollection",
        features,
        properties: {
          slideUid,
          exportedAt: new Date().toISOString(),
          totalAnnotations: vectorAnnotations.length,
        },
      };

      // Create and download the file
      const blob = new Blob([JSON.stringify(featureCollection, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `vector-annotations-${slideUid}-${
        new Date().toISOString().split("T")[0]
      }.geojson`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Failed to download all vector annotations:", error);
    }
  }, [slideUid, vectorAnnotations]);

  // Direct import functionality
  const importAllVectorAnnotations = useCallback(async () => {
    if (!slideUid || vectorAnnotations.length === 0) return;

    try {
      // Convert study annotation settings to the format expected by the API
      const allowedLabels = finalSettings?.annotations?.map(
        (annotation: any) => ({
          label: annotation.label,
          type: annotation.type,
          color: annotation.color,
        })
      ) || [
        // Fallback labels if study settings are not available
        { label: "roi", type: "polygon", color: "#ff0000" },
        { label: "tumor", type: "polygon", color: "#00ff00" },
        { label: "stroma", type: "polygon", color: "#0000ff" },
        { label: "artifact", type: "polygon", color: "#ffff00" },
      ];

      // Import all vector annotations
      const results = await Promise.allSettled(
        vectorAnnotations.map((annotation) =>
          vectorAnnotationsService.importVectorAnnotationToWorkspace(
            slideUid,
            annotation.vectorUid,
            { allowedLabels }
          )
        )
      );

      // Process results
      let totalImported = 0;
      let totalSkipped = 0;
      let totalOverwritten = 0;
      const allSkippedLabels = new Set<string>();

      results.forEach((result) => {
        if (result.status === "fulfilled") {
          totalImported += result.value.importedCount;
          totalSkipped += result.value.skippedCount;
          totalOverwritten += result.value.overwrittenCount;
          result.value.skippedLabels.forEach((label) =>
            allSkippedLabels.add(label)
          );
        }
      });

      const aggregatedResult: AnnotationImportResult = {
        slideUid: slideUid,
        vectorUid: "bulk-import", // Placeholder for bulk import
        importedCount: totalImported,
        skippedCount: totalSkipped,
        overwrittenCount: totalOverwritten,
        skippedLabels: Array.from(allSkippedLabels),
        importedAnnotations: [], // Not needed for this use case
        geoJsonFeatures: null, // Not needed for this use case
        studyLabels: allowedLabels.map((label) => label.label),
      };

      console.log("Bulk import completed:", aggregatedResult);

      // Refresh annotations after import
      await fetchAnnotations();

      // Call the parent callback if provided
      onImportComplete?.(aggregatedResult);
    } catch (error) {
      console.error("Failed to import vector annotations:", error);
    }
  }, [
    slideUid,
    vectorAnnotations,
    finalSettings,
    fetchAnnotations,
    onImportComplete,
  ]);

  const importSingleVectorAnnotation = useCallback(
    async (vectorUid: string) => {
      if (!slideUid) return;

      try {
        // Convert study annotation settings to the format expected by the API
        const allowedLabels = finalSettings?.annotations?.map(
          (annotation: any) => ({
            label: annotation.label,
            type: annotation.type,
            color: annotation.color,
          })
        ) || [
          // Fallback labels if study settings are not available
          { label: "roi", type: "polygon", color: "#ff0000" },
          { label: "tumor", type: "polygon", color: "#00ff00" },
          { label: "stroma", type: "polygon", color: "#0000ff" },
          { label: "artifact", type: "polygon", color: "#ffff00" },
        ];

        const result =
          await vectorAnnotationsService.importVectorAnnotationToWorkspace(
            slideUid,
            vectorUid,
            { allowedLabels }
          );

        console.log("Single import completed:", result);

        // Refresh annotations after import
        await fetchAnnotations();

        // Call the parent callback if provided
        onImportComplete?.(result);
      } catch (error) {
        console.error("Failed to import vector annotation:", error);
      }
    },
    [slideUid, finalSettings, fetchAnnotations, onImportComplete]
  );

  const downloadSingleVectorAnnotation = useCallback(
    async (vectorUid: string, vectorName: string) => {
      if (!slideUid) return;

      try {
        const geoJsonData =
          await vectorAnnotationsService.getVectorAnnotationFile(
            slideUid,
            vectorUid
          );

        if (!geoJsonData) {
          console.error("No GeoJSON data found for annotation:", vectorUid);
          return;
        }

        // Create and download the file
        const blob = new Blob([JSON.stringify(geoJsonData, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = `${vectorName || vectorUid}.geojson`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      } catch (error) {
        console.error(`Failed to download annotation ${vectorUid}:`, error);
      }
    },
    [slideUid]
  );

  // Badges
  const MutabilityBadge = ({ editable }: { editable?: boolean }) => (
    <Badge variant="secondary" className="h-5 px-2 text-[10px] rounded-full">
      {editable ? "Editable" : "Read-only"}
    </Badge>
  );
  const ActorBadge = ({
    actorType,
    modelName,
  }: {
    actorType?: string;
    modelName?: string;
  }) => {
    if (actorType === "model")
      return (
        <Badge variant="outline" className="h-5 px-2 text-[10px] rounded-full">
          {modelName || "Model"}
        </Badge>
      );
    if (actorType === "user")
      return (
        <Badge variant="outline" className="h-5 px-2 text-[10px] rounded-full">
          User
        </Badge>
      );
    return null;
  };
  const ActorIcon = ({ type }: { type?: string }) =>
    type === "user" ? (
      <UserIcon className="w-3.5 h-3.5" />
    ) : type === "model" ? (
      <CpuChipIcon className="w-3.5 h-3.5" />
    ) : null;

  // Sections
  const GroupSection = ({
    title,
    count,
    groupKey,
    children,
    icon,
  }: {
    title: string;
    count: number;
    groupKey: string;
    icon?: React.ReactNode;
    children: React.ReactNode;
  }) => {
    const collapsed = !!collapsedGroups[groupKey];
    return (
      <div className="rounded-md border border-border/60 bg-muted/40">
        <Collapsible
          open={!collapsed}
          onOpenChange={() => toggleGroup(groupKey)}
        >
          <CollapsibleTrigger className="w-full flex items-center justify-between px-3 py-2 hover:bg-accent transition-colors">
            <div className="flex items-center gap-2 text-sm font-medium">
              {collapsed ? (
                <ChevronRightIcon className="w-4 h-4 text-muted-foreground" />
              ) : (
                <ChevronDownIcon className="w-4 h-4 text-muted-foreground" />
              )}
              {icon}
              <span className="capitalize">{title}</span>
            </div>
            <span className="text-xs text-muted-foreground">{count}</span>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <Separator className="opacity-50" />
            <div className="divide-y divide-border/30">{children}</div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    );
  };

  const LayerRow = ({
    id,
    name,
    actorType,
    modelName,
    editable,
    selected,
    onSelect,
    isVector,
    onDownload,
    onImport,
  }: {
    id: string;
    name: string;
    actorType?: string;
    modelName?: string;
    editable?: boolean;
    selected?: boolean;
    onSelect?: () => void;
    isVector?: boolean;
    onDownload?: () => void;
    onImport?: () => void;
  }) => (
    <div
      className={`px-3 py-2.5 ${
        selected ? "bg-primary/10" : "hover:bg-accent/50"
      }`}
    >
      <div className="flex items-center gap-2">
        <button onClick={onSelect} className="min-w-0 flex-1 text-left">
          <div className="flex items-center gap-2 min-w-0">
            <span className="truncate text-sm font-medium">{name}</span>
          </div>
        </button>
        <div className="flex items-center gap-1">
          {isVector && onImport && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    className="p-1 rounded hover:bg-accent text-muted-foreground"
                    onClick={(e) => {
                      e.stopPropagation();
                      onImport();
                    }}
                    aria-label="Import annotation"
                  >
                    <ArrowUpTrayIcon className="w-4 h-4" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="left" className="text-xs">
                  Import into workspace
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
          {isVector && onDownload && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    className="p-1 rounded hover:bg-accent text-muted-foreground"
                    onClick={(e) => {
                      e.stopPropagation();
                      onDownload();
                    }}
                    aria-label="Download annotation"
                  >
                    <ArrowDownTrayIcon className="w-4 h-4" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="left" className="text-xs">
                  Download GeoJSON
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  onClick={(e) => {
                    e.stopPropagation();
                    toggleVisible(id);
                    // Note: Individual layer visibility is now controlled via the visible state
                    // The parent component should listen to these visibility changes
                    // and apply them to the actual OpenLayers layers
                  }}
                  aria-label={isVisible(id) ? "Hide layer" : "Show layer"}
                >
                  {isVisible(id) ? (
                    <EyeIcon className="w-4 h-4" />
                  ) : (
                    <EyeSlashIcon className="w-4 h-4" />
                  )}
                </button>
              </TooltipTrigger>
              <TooltipContent side="left" className="text-xs">
                {isVisible(id) ? "Visible" : "Hidden"}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      <div className="mt-1.5 flex items-center gap-1">
        <MutabilityBadge editable={editable} />
        <ActorBadge actorType={actorType} modelName={modelName} />
      </div>

      <div className="mt-2 flex items-center gap-2">
        <span className="text-[11px] text-muted-foreground w-12">Opacity</span>
        <div className="flex-1">
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={getOpacity(id)}
            onChange={(e) => {
              const v = parseFloat(e.target.value);
              setOpacity(id, snapToTenth(v));
              // Note: Opacity changes are stored in layerOpacities state
              // The parent component should listen to these changes
              // and apply them to the actual OpenLayers layers
            }}
            className="w-full h-1 bg-input rounded-lg appearance-none cursor-pointer"
          />
        </div>
        <span className="text-[11px] text-muted-foreground font-mono w-10 text-right">
          {Math.round(getOpacity(id) * 100)}%
        </span>
      </div>
    </div>
  );

  const RasterTab = (
    <div className="rounded-md border border-border/60 bg-muted/60">
      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium">Raster Annotations</h3>
          <span className="text-xs text-muted-foreground">
            {maskAnnotations.length}
          </span>
        </div>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                className="p-1 rounded hover:bg-accent text-muted-foreground"
                onClick={() => setShowMask(!showMask)}
                aria-label={
                  showMask
                    ? "Hide all raster annotations"
                    : "Show all raster annotations"
                }
              >
                {showMask ? (
                  <EyeIcon className="w-4 h-4" />
                ) : (
                  <EyeSlashIcon className="w-4 h-4" />
                )}
              </button>
            </TooltipTrigger>
            <TooltipContent side="left" className="text-xs">
              {showMask ? "Hide all raster" : "Show all raster"}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
      <Separator />
      <div className="px-3 py-2 space-y-1">
        {(["user", "model", "unknown"] as const).map((t) => {
          const list =
            t === "user"
              ? groupedMasks.user
              : t === "model"
              ? groupedMasks.model
              : groupedMasks.unknown;
          if (list.length === 0) return null;
          const key = `masks-${t}`;
          return (
            <GroupSection
              key={t}
              title={t === "unknown" ? "Other" : t}
              count={list.length}
              groupKey={key}
              icon={<ActorIcon type={t} />}
            >
              {list.map((m) => (
                <LayerRow
                  key={m.maskUid}
                  id={m.maskUid}
                  name={m.maskName || m.maskUid}
                  actorType={m.actorType}
                  modelName={m.modelName}
                  editable={m.mutable}
                  selected={selectedMaskId === m.maskUid}
                  onSelect={() => setSelectedMaskId(m.maskUid)}
                />
              ))}
            </GroupSection>
          );
        })}
      </div>
    </div>
  );

  const VectorTab = (
    <div className="rounded-md border border-border/60 bg-muted/60">
      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium">Vector Annotations</h3>
          <span className="text-xs text-muted-foreground">
            {vectorAnnotations.length}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={importAllVectorAnnotations}
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  title="Import all vector annotations into workspace"
                  aria-label="Import all vector annotations"
                  disabled={vectorAnnotations.length === 0}
                >
                  <ArrowUpTrayIcon className="w-4 h-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="left" className="text-xs">
                Import all into workspace
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={downloadAllVectorAnnotations}
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  title="Download vector annotations as GeoJSON"
                  aria-label="Download vector annotations"
                  disabled={vectorAnnotations.length === 0}
                >
                  <ArrowDownTrayIcon className="w-4 h-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="left" className="text-xs">
                Download GeoJSON
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                  onClick={() => setShowVectors(!showVectors)}
                  aria-label={
                    showVectors
                      ? "Hide all vector annotations"
                      : "Show all vector annotations"
                  }
                >
                  {showVectors ? (
                    <EyeIcon className="w-4 h-4" />
                  ) : (
                    <EyeSlashIcon className="w-4 h-4" />
                  )}
                </button>
              </TooltipTrigger>
              <TooltipContent side="left" className="text-xs">
                {showVectors ? "Hide all vectors" : "Show all vectors"}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
      <Separator />
      <div className="px-3 py-2 space-y-1">
        {(["user", "model", "unknown"] as const).map((t) => {
          const list =
            t === "user"
              ? groupedVectors.user
              : t === "model"
              ? groupedVectors.model
              : groupedVectors.unknown;
          if (list.length === 0) return null;
          const key = `vectors-${t}`;
          return (
            <GroupSection
              key={t}
              title={t === "unknown" ? "Other" : t}
              count={list.length}
              groupKey={key}
              icon={<ActorIcon type={t} />}
            >
              {list.map((v) => (
                <LayerRow
                  key={v.vectorUid}
                  id={v.vectorUid}
                  name={v.vectorName || v.vectorUid}
                  actorType={v.actorType}
                  modelName={v.modelName}
                  editable={v.mutable}
                  selected={selectedVectorId === v.vectorUid}
                  onSelect={() => setSelectedVectorId(v.vectorUid)}
                  isVector={true}
                  onImport={() => importSingleVectorAnnotation(v.vectorUid)}
                  onDownload={() =>
                    downloadSingleVectorAnnotation(
                      v.vectorUid,
                      v.vectorName || v.vectorUid
                    )
                  }
                />
              ))}
            </GroupSection>
          );
        })}
      </div>
    </div>
  );

  const effectiveOpen = openOverride ?? isOpen;
  const effectiveDock = dockOverride ?? dock;

  return (
    <>
      <SlidePanel
        isOpen={effectiveDock === "left" ? true : effectiveOpen}
        onClose={() =>
          openOverride !== undefined ? onOpenChange?.(false) : setIsOpen(false)
        }
        dockOverride={effectiveDock}
        onDockChange={(d) => (dockOverride ? onDockChange?.(d) : setDock(d))}
        storageKey="annotationPanel"
        defaultSize={size}
        dockedWidth={effectiveOpen ? 320 : 0}
      >
        <SlidePanel.Header
          title="Annotations"
          onClose={() =>
            openOverride !== undefined
              ? onOpenChange?.(false)
              : setIsOpen(false)
          }
        />

        {effectiveOpen && (
          <div className="flex-1 min-h-0 overflow-hidden">
            {/* Sticky filter bar */}
            <div className="sticky top-0 z-10 border-b border-border/60 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/70">
              <div className="px-3 py-2 space-y-2">
                <Input
                  type="search"
                  placeholder="Filter annotations by name"
                  value={layerFilter}
                  onChange={(e) => setLayerFilter(e.target.value)}
                  className="h-8 text-xs"
                />

                {/* Actor segmented filter */}
                <div className="flex gap-1 text-xs">
                  {["all", "user", "model"].map((f) => (
                    <button
                      key={f}
                      onClick={() => setActorFilter(f as any)}
                      className={`flex-1 px-2 py-1 rounded border transition-colors ${
                        actorFilter === f
                          ? "bg-primary/15 border-primary/50 text-primary"
                          : "hover:bg-accent border-border"
                      }`}
                      title={`Filter by ${f}`}
                    >
                      <span className="flex items-center justify-center gap-1">
                        {f === "user" && <UserIcon className="w-3 h-3" />}
                        {f === "model" && <CpuChipIcon className="w-3 h-3" />}
                        <span className="capitalize">{f}</span>
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Scrollable content */}
            <div className="h-[calc(100%-80px)] overflow-y-auto">
              {/* approximate; panel header + filter bar */}
              <div className="px-3 py-3 space-y-3">
                {loading && (
                  <div className="text-center py-6 text-sm text-muted-foreground">
                    Loading annotations…
                  </div>
                )}
                {error && (
                  <div className="text-center py-6 text-sm text-red-500">
                    {error}
                  </div>
                )}

                {!loading && !error && (
                  <>
                    {/* Raster Annotations Section */}
                    {hasMaskData ? (
                      <div className="mb-3">{RasterTab}</div>
                    ) : (
                      <div className="rounded-md border border-border/60 bg-muted/60">
                        <div className="flex items-center justify-between px-3 py-2">
                          <div className="flex items-center gap-2">
                            <h3 className="text-sm font-medium">
                              Raster Annotations
                            </h3>
                            <span className="text-xs text-muted-foreground">
                              0
                            </span>
                          </div>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <button
                                  className="p-1 rounded hover:bg-accent text-muted-foreground"
                                  onClick={() => setShowMask(!showMask)}
                                  aria-label={
                                    showMask
                                      ? "Hide all raster annotations"
                                      : "Show all raster annotations"
                                  }
                                >
                                  {showMask ? (
                                    <EyeIcon className="w-4 h-4" />
                                  ) : (
                                    <EyeSlashIcon className="w-4 h-4" />
                                  )}
                                </button>
                              </TooltipTrigger>
                              <TooltipContent side="left" className="text-xs">
                                {showMask
                                  ? "Hide all raster"
                                  : "Show all raster"}
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Separator />
                        <div className="text-center py-8 text-muted-foreground">
                          <p className="text-sm">No raster annotations</p>
                          <p className="text-xs mt-1">
                            Items will appear here when available
                          </p>
                        </div>
                      </div>
                    )}

                    {/* Vector Annotations Section */}
                    {hasVectorData ? (
                      VectorTab
                    ) : (
                      <div className="rounded-md border border-border/60 bg-muted/60">
                        <div className="flex items-center justify-between px-3 py-2">
                          <div className="flex items-center gap-2">
                            <h3 className="text-sm font-medium">
                              Vector Annotations
                            </h3>
                            <span className="text-xs text-muted-foreground">
                              0
                            </span>
                          </div>
                          <div className="flex items-center gap-1">
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    onClick={importAllVectorAnnotations}
                                    className="p-1 rounded hover:bg-accent text-muted-foreground"
                                    title="Import all vector annotations into workspace"
                                    aria-label="Import all vector annotations"
                                    disabled={vectorAnnotations.length === 0}
                                  >
                                    <ArrowUpTrayIcon className="w-4 h-4" />
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent side="left" className="text-xs">
                                  Import all into workspace
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    className="p-1 rounded hover:bg-accent text-muted-foreground"
                                    onClick={() => setShowVectors(!showVectors)}
                                    aria-label={
                                      showVectors
                                        ? "Hide all vector annotations"
                                        : "Show all vector annotations"
                                    }
                                  >
                                    {showVectors ? (
                                      <EyeIcon className="w-4 h-4" />
                                    ) : (
                                      <EyeSlashIcon className="w-4 h-4" />
                                    )}
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent side="left" className="text-xs">
                                  {showVectors
                                    ? "Hide all vectors"
                                    : "Show all vectors"}
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          </div>
                        </div>
                        <Separator />
                        <div className="text-center py-8 text-muted-foreground">
                          <p className="text-sm">No vector annotations</p>
                          <p className="text-xs mt-1">
                            Items will appear here when available
                          </p>
                        </div>
                      </div>
                    )}

                    {/* Study Labels */}
                    <div className="mt-3">
                      <button
                        onClick={() => setInfoCollapsed(!infoCollapsed)}
                        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors mb-2"
                      >
                        {infoCollapsed ? (
                          <ChevronRightIcon className="h-3 w-3" />
                        ) : (
                          <ChevronDownIcon className="h-3 w-3" />
                        )}
                        <InformationCircleIcon className="h-3 w-3" />
                        <span>Study Labels</span>
                      </button>

                      {!infoCollapsed && (
                        <div className="ml-1 rounded-md border border-border/60 bg-card p-2">
                          {settingsLoading && (
                            <div className="text-xs text-muted-foreground">
                              Loading annotation settings…
                            </div>
                          )}
                          {settingsError && (
                            <div className="text-xs text-red-400">
                              Error loading annotation settings
                            </div>
                          )}
                          {!settingsLoading &&
                            !settingsError &&
                            LABELS.length > 0 && (
                              <div className="space-y-2">
                                <div className="text-xs text-muted-foreground">
                                  Configured labels for this study:
                                </div>
                                <div className="grid grid-cols-1 gap-1 max-h-36 overflow-y-auto">
                                  {LABELS.map((l) => (
                                    <div
                                      key={l.id}
                                      className="flex items-center justify-between px-1.5 py-1 rounded"
                                    >
                                      <span className="flex items-center gap-2 min-w-0">
                                        <span
                                          className="h-2.5 w-2.5 rounded-sm shrink-0"
                                          style={{ backgroundColor: l.color }}
                                        />
                                        <span className="text-sm truncate">
                                          {l.name}
                                        </span>
                                        <span className="text-[11px] text-muted-foreground">
                                          ({l.hint})
                                        </span>
                                      </span>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                          {!settingsLoading &&
                            !settingsError &&
                            LABELS.length === 0 && (
                              <div className="text-xs text-muted-foreground text-center">
                                <p className="mb-2">No labels defined.</p>
                                {studyUid && (
                                  <button
                                    onClick={() =>
                                      navigate({
                                        to: `/studies/${studyUid}/settings/annotations`,
                                      })
                                    }
                                    className="text-blue-500 hover:text-blue-400 underline"
                                  >
                                    Configure
                                  </button>
                                )}
                              </div>
                            )}
                        </div>
                      )}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
        )}
      </SlidePanel>
    </>
  );
}
