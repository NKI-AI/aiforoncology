// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import type Map from "ol/Map";
import { DisplayMetadata } from "@/features/viewer/components/map/types";

/**
 * Context available to all panels
 */
export interface PanelContext {
  /** The OpenLayers map instance */
  mapRef: Map | null;
  /** Current slide UID */
  slideUid?: string;
  /** Current study UID */
  studyUid?: string;
  /** Current case UID */
  caseUid?: string;
  /** Slide metadata */
  slideMetadata: DisplayMetadata | null;
  /** Raw slide metadata (includes channels, etc.) */
  rawSlideMetadata: any;
}

/**
 * Panel state management
 */
export interface PanelState {
  /** Whether the panel is currently open */
  isOpen: boolean;
  /** Panel docking mode */
  dock: "free" | "left";
  /** Panel size (for floating panels) */
  size?: { width: number; height: number };
  /** Custom panel-specific state */
  customState?: Record<string, any>;
}

/**
 * Panel registration interface
 */
export interface PanelRegistration {
  /** Unique panel identifier */
  id: string;
  /** Display name for the panel */
  name: string;
  /** Icon component for the dock button */
  icon: React.ComponentType<{ className?: string }>;
  /** Panel component */
  component: React.ComponentType<PanelProps>;
  /** Default panel state */
  defaultState: Partial<PanelState>;
  /** Whether this panel should be available by default */
  enabled?: boolean;
  /** Keyboard shortcut to toggle panel (optional) */
  shortcut?: string;
  /** Panel position order in the dock (lower numbers appear first) */
  order?: number;
}

/**
 * Props passed to each panel component
 */
export interface PanelProps {
  /** Panel context with shared data */
  context: PanelContext;
  /** Current panel state */
  state: PanelState;
  /** Update panel state */
  updateState: (updates: Partial<PanelState>) => void;
  /** Close the panel */
  onClose: () => void;
}

/**
 * Panel manager interface for registering and managing panels
 */
export interface PanelManager {
  /** Register a new panel */
  register: (panel: PanelRegistration) => void;
  /** Unregister a panel */
  unregister: (panelId: string) => void;
  /** Get all registered panels */
  getPanels: () => PanelRegistration[];
  /** Get panel state */
  getState: (panelId: string) => PanelState;
  /** Update panel state */
  updateState: (panelId: string, updates: Partial<PanelState>) => void;
  /** Toggle panel open/closed */
  toggle: (panelId: string) => void;
  /** Close all panels */
  closeAll: () => void;
}
