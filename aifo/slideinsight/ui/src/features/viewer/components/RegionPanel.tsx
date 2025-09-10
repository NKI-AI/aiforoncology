// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useEffect, useRef, useState } from "react";
import {
  XMarkIcon,
  EyeIcon,
  EyeSlashIcon,
  TrashIcon,
  CheckIcon,
  XMarkIcon as CancelIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  Cog6ToothIcon,
  QuestionMarkCircleIcon,
} from "@heroicons/react/24/outline";
import { RectangleStackIcon } from "@heroicons/react/24/solid";
import SlidePanel from "@/components/ui/slide-panel";
import {
  type RegionSettings,
  type RegionSettingsPartial,
} from "@/features/viewer/utils/regionSettingsStorage";

export interface RegionItem {
  id: string;
  name: string;
  visible?: boolean;
}

export interface RegionPanelProps {
  isOpen: boolean;
  onClose: () => void;

  regions: RegionItem[];
  onUpdateRegions?: (items: RegionItem[]) => void;
  selectedId?: string | null;
  onSelect?: (id: string) => void;
  /** Hovered region id to highlight in the list */
  hoveredId?: string | null;
  /** Callback when hovering list items; pass null on leave */
  onHoverIdChange?: (id: string | null) => void;

  dockOverride?: "free" | "left";
  onDockChange?: (dock: "free" | "left") => void;
  /** Start drawing a box region on the map */
  onStartDrawRegion?: () => void;
  /** Stop current drawing interaction (if any). */
  onStopDraw?: () => void;
  /** Whether drawing is currently active */
  isDrawing?: boolean;
  /** Callback to check if drawing should be allowed (e.g., no unnamed regions) */
  canStartDrawing?: () => boolean;
  /** Callback when editing state changes */
  onEditingStateChange?: (isEditing: boolean) => void;
  /** Region display settings */
  regionDisplaySettings?: RegionSettings;
  /** Callback to update region display settings */
  onUpdateRegionDisplaySettings?: (settings: RegionSettingsPartial) => void;
  /** Callback to reset region display settings to defaults */
  onResetRegionDisplaySettings?: () => void;
}

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

export function RegionPanel({
  isOpen,
  onClose,
  regions,
  onUpdateRegions,
  selectedId = null,
  onSelect,
  dockOverride,
  onDockChange,
  onStartDrawRegion,
  onStopDraw,
  isDrawing = false,
  hoveredId = null,
  onHoverIdChange,
  canStartDrawing,
  onEditingStateChange,
  regionDisplaySettings,
  onUpdateRegionDisplaySettings,
  onResetRegionDisplaySettings,
}: RegionPanelProps) {
  const [size] = useState({ width: 320, height: 400 });
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState<string>("");
  const [nameError, setNameError] = useState<string>("");
  const [isNewlyCreated, setIsNewlyCreated] = useState<boolean>(false);
  const [flashingId, setFlashingId] = useState<string | null>(null);
  const [settingsCollapsed, setSettingsCollapsed] = useState(true);
  const [instructionsCollapsed, setInstructionsCollapsed] = useState(false);
  const editInputRef = useRef<HTMLInputElement>(null);

  const panelRef = useRef<HTMLDivElement>(null);

  // Focus input when starting to edit
  useEffect(() => {
    if (editingId && editInputRef.current) {
      editInputRef.current.focus();
      editInputRef.current.select();
    }
  }, [editingId]);

  // Notify parent when editing state changes
  useEffect(() => {
    onEditingStateChange?.(editingId !== null);
  }, [editingId, onEditingStateChange]);

  // Listen for auto-edit events from region creation
  useEffect(() => {
    const handleEditRegion = (event: CustomEvent) => {
      const { regionId, name } = event.detail;
      if (isOpen && regions.some((r) => r.id === regionId)) {
        setEditingId(regionId);
        setEditingName(name);
        setIsNewlyCreated(true); // Mark as newly created
      }
    };

    window.addEventListener("editRegion", handleEditRegion as EventListener);
    return () => {
      window.removeEventListener(
        "editRegion",
        handleEditRegion as EventListener
      );
    };
  }, [isOpen, regions]);

  if (!isOpen) return null;

  const handleClose = () => {
    // Stop drawing when closing the panel
    if (isDrawing) {
      onStopDraw?.();
    }
    onClose();
  };

  const flashUnnamedRegion = () => {
    // Find the currently editing region and flash it
    if (editingId) {
      setFlashingId(editingId);
      setTimeout(() => setFlashingId(null), 1000); // Flash for 1 second
    }
  };

  const handleDrawButtonClick = () => {
    if (isDrawing) {
      onStopDraw?.();
      return;
    }

    // Check if drawing should be allowed
    const canDraw = canStartDrawing ? canStartDrawing() : true;
    if (!canDraw) {
      flashUnnamedRegion();
      return;
    }

    onStartDrawRegion?.();
  };

  const toggleVisible = (id: string) => {
    const next = regions.map((r) =>
      r.id === id ? { ...r, visible: !r.visible } : r
    );
    onUpdateRegions?.(next);
  };

  const removeItem = (id: string) => {
    onUpdateRegions?.(regions.filter((r) => r.id !== id));
  };

  const startEdit = (region: RegionItem) => {
    setEditingId(region.id);
    setEditingName(region.name);
    setNameError("");
    setIsNewlyCreated(false); // This is manual editing, not auto-edit from creation
  };

  const saveEdit = () => {
    if (!editingId || !editingName.trim()) {
      // If this was a newly created region with no name, delete it
      if (isNewlyCreated && editingId) {
        removeItem(editingId);
      }
      setEditingId(null);
      setEditingName("");
      setNameError("");
      setIsNewlyCreated(false);
      return;
    }

    const trimmedName = editingName.trim();

    // Check for duplicate names (excluding the current region being edited)
    const isDuplicate = regions.some(
      (r) =>
        r.id !== editingId && r.name.toLowerCase() === trimmedName.toLowerCase()
    );

    if (isDuplicate) {
      setNameError(`"${trimmedName}" already exists`);
      return;
    }

    const next = regions.map((r) =>
      r.id === editingId ? { ...r, name: trimmedName } : r
    );
    onUpdateRegions?.(next);
    setEditingId(null);
    setEditingName("");
    setNameError("");
    setIsNewlyCreated(false);
  };

  const cancelEdit = () => {
    // If this was a newly created region, delete it
    if (isNewlyCreated && editingId) {
      removeItem(editingId);
    }

    setEditingId(null);
    setEditingName("");
    setNameError("");
    setIsNewlyCreated(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      saveEdit();
    } else if (e.key === "Escape") {
      cancelEdit();
    }
  };

  const PanelHeader = (
    <SlidePanel.Header title="Regions" onClose={handleClose} />
  );

  const PanelBody = (
    <>
      {/* Upper: regions list */}
      <div className="flex-1 min-h-0 overflow-y-auto px-3 py-2 space-y-2">
        {regions.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <RectangleStackIcon className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No regions yet</p>
            <p className="text-xs mt-1">Click "Draw Region" to start</p>
          </div>
        ) : (
          <ul className="space-y-1">
            {regions.map((region) => (
              <li
                key={region.id}
                className={cls(
                  "rounded transition-colors",
                  selectedId === region.id && "bg-sky-600/30",
                  hoveredId === region.id &&
                    selectedId !== region.id &&
                    "bg-amber-500/25",
                  flashingId === region.id &&
                    "animate-pulse bg-yellow-400/50 dark:bg-yellow-500/50"
                )}
                onMouseEnter={() => onHoverIdChange?.(region.id)}
                onMouseLeave={() => onHoverIdChange?.(null)}
              >
                <div className="flex items-center justify-between px-2 py-2">
                  <div className="flex min-w-0 items-center gap-2 flex-1">
                    <RectangleStackIcon className="h-4 w-4 text-muted-400 dark:text-muted-500 flex-shrink-0" />
                    {editingId === region.id ? (
                      <div className="flex-1 min-w-0">
                        <input
                          ref={editInputRef}
                          type="text"
                          value={editingName}
                          onChange={(e) => {
                            setEditingName(e.target.value);
                            if (nameError) setNameError(""); // Clear error on typing
                          }}
                          onKeyDown={handleKeyDown}
                          onBlur={saveEdit}
                          className={cls(
                            "w-full px-1 py-0.5 text-sm bg-gray-700 dark:bg-gray-800 border rounded text-muted-200 dark:text-muted-300 focus:outline-none",
                            nameError
                              ? "border-red-500 focus:border-red-400"
                              : "border-gray-600 dark:border-gray-500 focus:border-indigo-500"
                          )}
                        />
                      </div>
                    ) : (
                      <button
                        onClick={() => {
                          // If this is a different region, select it first
                          if (selectedId !== region.id) {
                            onSelect?.(region.id);
                          } else {
                            // If it's already selected, start editing
                            startEdit(region);
                          }
                        }}
                        className="flex-1 min-w-0 text-left hover:bg-gray-700/50 dark:hover:bg-gray-600/50 rounded px-1 py-0.5 transition-colors"
                        title="Click to select, click again to edit name"
                      >
                        <span className="truncate text-sm text-muted-200 dark:text-muted-300">
                          {region.name}
                        </span>
                      </button>
                    )}
                  </div>

                  <div className="flex items-center gap-1 ml-2">
                    {editingId === region.id ? (
                      <>
                        <button
                          className="p-1 rounded hover:bg-gray-700 dark:hover:bg-gray-600 text-green-400 dark:text-green-400"
                          title="Save"
                          onClick={saveEdit}
                        >
                          <CheckIcon className="h-4 w-4" />
                        </button>
                        <button
                          className="p-1 rounded hover:bg-gray-700 dark:hover:bg-gray-600 text-muted-400 dark:text-muted-500"
                          title="Cancel"
                          onClick={cancelEdit}
                        >
                          <CancelIcon className="h-4 w-4" />
                        </button>
                      </>
                    ) : (
                      <>
                        <button
                          className="p-1 rounded hover:bg-accent text-muted-foreground"
                          title="Edit name"
                          onClick={(e) => {
                            e.stopPropagation();
                            startEdit(region);
                          }}
                        >
                          <svg
                            className="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2.5}
                            stroke="currentColor"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125"
                            />
                          </svg>
                        </button>
                        <button
                          className="p-1 rounded hover:bg-accent text-muted-foreground"
                          title={region.visible === false ? "Show" : "Hide"}
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleVisible(region.id);
                          }}
                        >
                          {region.visible === false ? (
                            <EyeIcon className="h-4 w-4" />
                          ) : (
                            <EyeSlashIcon className="h-4 w-4" />
                          )}
                        </button>
                        <button
                          className="p-1 rounded hover:bg-accent text-muted-foreground"
                          title="Delete"
                          onClick={(e) => {
                            e.stopPropagation();
                            removeItem(region.id);
                          }}
                        >
                          <TrashIcon className="h-4 w-4" />
                        </button>
                      </>
                    )}
                  </div>
                </div>
                {editingId === region.id && nameError && (
                  <div className="flex items-center gap-1 px-2 pb-2">
                    <svg
                      className="h-3 w-3 text-red-400 dark:text-red-300 flex-shrink-0"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fillRule="evenodd"
                        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
                        clipRule="evenodd"
                      />
                    </svg>
                    <span className="text-xs text-red-400 dark:text-red-300">
                      {nameError}
                    </span>
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Bottom: draw controls and settings */}
      <div className="px-3 pt-2 pb-2 border-t border-border700/60 bg-muted/70 flex-shrink-0">
        {/* Draw button */}
        <button
          type="button"
          onClick={handleDrawButtonClick}
          className={cls(
            "w-full flex items-center justify-center gap-2 px-3 py-2 rounded transition-colors mb-2",
            isDrawing
              ? "bg-red-600 hover:bg-red-500 text-white"
              : canStartDrawing && !canStartDrawing()
              ? "bg-gray-500 hover:bg-gray-400 text-white cursor-not-allowed"
              : "bg-indigo-600 hover:bg-indigo-500 text-white"
          )}
          title={
            isDrawing
              ? "Stop drawing regions"
              : canStartDrawing && !canStartDrawing()
              ? "Complete naming the current region before drawing a new one"
              : "Start drawing regions"
          }
        >
          <RectangleStackIcon className="h-4 w-4" />
          <span className="text-sm font-medium">
            {isDrawing ? "Stop Drawing" : "Draw Region"}
          </span>
        </button>

        {/* Region Display Settings - collapsible */}
        {regionDisplaySettings && onUpdateRegionDisplaySettings && (
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
                  Stroke Width: {regionDisplaySettings.strokeWidth}px
                </label>
                <input
                  type="range"
                  min={1}
                  max={10}
                  step={1}
                  value={regionDisplaySettings.strokeWidth}
                  onChange={(e) =>
                    onUpdateRegionDisplaySettings({
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
                  {Math.round(regionDisplaySettings.fillOpacity * 100)}%
                </label>
                <input
                  type="range"
                  min={0}
                  max={1}
                  step={0.05}
                  value={regionDisplaySettings.fillOpacity}
                  onChange={(e) =>
                    onUpdateRegionDisplaySettings({
                      fillOpacity: Number(e.target.value),
                    })
                  }
                  className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                />
              </div>

              {/* Reset Button */}
              {onResetRegionDisplaySettings && (
                <button
                  onClick={onResetRegionDisplaySettings}
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
              • <strong>Draw:</strong> Click "Draw Region" to start drawing
            </li>
            <li>
              • <strong>Regions:</strong> Click and drag to create rectangular
              regions
            </li>
            <li>
              • <strong>Select:</strong> Click on a region to select it
            </li>
            <li>
              • <strong>Modify:</strong> Drag handles on selected region to
              resize
            </li>
            <li>
              • <strong>Edit:</strong> Click again on selected region to rename
            </li>
            <li>
              • <strong>Visibility:</strong> Use eye icon to show/hide regions
            </li>
          </ul>
        </PanelFoldableBody>
      </div>
    </>
  );

  return (
    <SlidePanel
      isOpen={isOpen}
      onClose={handleClose}
      dockOverride={dockOverride}
      onDockChange={onDockChange}
      storageKey="regionPanel"
      defaultSize={size}
    >
      {PanelHeader}
      {PanelBody}
    </SlidePanel>
  );
}

export default RegionPanel;
