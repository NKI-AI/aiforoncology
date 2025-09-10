// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useReducer, useCallback, useEffect } from "react";
// Crosshair removed
import {
  loadViewerSettings,
  updateViewerSetting,
  isLocalStorageAvailable,
} from "../utils/viewerSettingsStorage";

interface ViewerState {
  showSettings: boolean;
  panSensitivity: number;
  zoomSensitivity: number;
}

type ViewerAction =
  | { type: "TOGGLE_SETTINGS" }
  | { type: "SET_SHOW_SETTINGS"; show: boolean }
  | { type: "SET_PAN_SENSITIVITY"; sensitivity: number }
  | { type: "SET_ZOOM_SENSITIVITY"; sensitivity: number }
  | { type: "RESET_VIEWER_STATE" };

const getInitialState = (isFluorescent?: boolean): ViewerState => {
  // Non-persistent state defaults
  const baseState: ViewerState = {
    showSettings: false,
    panSensitivity: 1.0,
    zoomSensitivity: 1.0,
  };

  // Load persistent settings from localStorage if available and image type is known
  if (isFluorescent !== undefined && isLocalStorageAvailable()) {
    const savedSettings = loadViewerSettings(isFluorescent);
    return {
      ...baseState,
      panSensitivity: savedSettings.panSensitivity,
      zoomSensitivity: savedSettings.zoomSensitivity,
    };
  }

  // Fallback to default values
  return baseState;
};

function viewerReducer(
  state: ViewerState,
  action: ViewerAction,
  isFluorescent?: boolean
): ViewerState {
  switch (action.type) {
    case "TOGGLE_SETTINGS":
      return { ...state, showSettings: !state.showSettings };
    case "SET_SHOW_SETTINGS":
      return { ...state, showSettings: action.show };
    case "SET_PAN_SENSITIVITY":
      return { ...state, panSensitivity: action.sensitivity };
    case "SET_ZOOM_SENSITIVITY":
      return { ...state, zoomSensitivity: action.sensitivity };
    case "RESET_VIEWER_STATE":
      return getInitialState(isFluorescent);
    default:
      return state;
  }
}

interface UseViewerStateReturn {
  state: ViewerState;
  toggleSettings: () => void;
  setShowSettings: (show: boolean) => void;
  setPanSensitivity: (sensitivity: number) => void;
  setZoomSensitivity: (sensitivity: number) => void;
  resetViewerState: () => void;
}

export function useViewerState(isFluorescent?: boolean): UseViewerStateReturn {
  const [state, dispatch] = useReducer(
    (state: ViewerState, action: ViewerAction) =>
      viewerReducer(state, action, isFluorescent),
    getInitialState(isFluorescent)
  );

  // Save persistent settings to localStorage when they change
  // crosshair removed

  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      updateViewerSetting(
        "panSensitivity",
        state.panSensitivity,
        isFluorescent
      );
    }
  }, [state.panSensitivity, isFluorescent]);

  useEffect(() => {
    if (isFluorescent !== undefined && isLocalStorageAvailable()) {
      updateViewerSetting(
        "zoomSensitivity",
        state.zoomSensitivity,
        isFluorescent
      );
    }
  }, [state.zoomSensitivity, isFluorescent]);

  const toggleSettings = useCallback(() => {
    dispatch({ type: "TOGGLE_SETTINGS" });
  }, []);

  const setShowSettings = useCallback((show: boolean) => {
    dispatch({ type: "SET_SHOW_SETTINGS", show });
  }, []);

  const setPanSensitivity = useCallback((sensitivity: number) => {
    dispatch({ type: "SET_PAN_SENSITIVITY", sensitivity });
  }, []);

  const setZoomSensitivity = useCallback((sensitivity: number) => {
    dispatch({ type: "SET_ZOOM_SENSITIVITY", sensitivity });
  }, []);

  const resetViewerState = useCallback(() => {
    dispatch({ type: "RESET_VIEWER_STATE" });
  }, []);

  return {
    state,
    toggleSettings,
    setShowSettings,
    setPanSensitivity,
    setZoomSensitivity,
    resetViewerState,
  };
}
