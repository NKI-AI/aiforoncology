// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from "react";
import {
  PanelRegistration,
  PanelState,
  PanelManager,
  PanelContext,
} from "./types";

/**
 * Default panel state
 */
const DEFAULT_PANEL_STATE: PanelState = {
  isOpen: false,
  dock: "left",
  size: { width: 320, height: 560 },
  customState: {},
};

/**
 * Panel Registry Context
 */
const PanelRegistryContext = createContext<{
  manager: PanelManager;
  context: PanelContext;
  states: Record<string, PanelState>;
} | null>(null);

/**
 * Hook to access the panel registry
 */
export function usePanelRegistry() {
  const registry = useContext(PanelRegistryContext);
  if (!registry) {
    throw new Error(
      "usePanelRegistry must be used within a PanelRegistryProvider"
    );
  }
  return registry;
}

/**
 * Hook to access panel manager only
 */
export function usePanelManager() {
  const { manager } = usePanelRegistry();
  return manager;
}

/**
 * Hook to access panel context only
 */
export function usePanelContext() {
  const { context } = usePanelRegistry();
  return context;
}

/**
 * Panel Registry Provider Props
 */
interface PanelRegistryProviderProps {
  children: React.ReactNode;
  context: PanelContext;
}

/**
 * Panel Registry Provider
 */
export function PanelRegistryProvider({
  children,
  context,
}: PanelRegistryProviderProps) {
  const [panels, setPanels] = useState<Map<string, PanelRegistration>>(
    new Map()
  );
  const [states, setStates] = useState<Record<string, PanelState>>({});

  // Load panel states from localStorage on mount
  useEffect(() => {
    const savedStates: Record<string, PanelState> = {};
    panels.forEach((panel, panelId) => {
      const saved = localStorage.getItem(`panel_${panelId}_state`);
      if (saved) {
        try {
          const parsedState = JSON.parse(saved);
          savedStates[panelId] = {
            ...DEFAULT_PANEL_STATE,
            ...panel.defaultState,
            ...parsedState,
          };
        } catch {
          savedStates[panelId] = {
            ...DEFAULT_PANEL_STATE,
            ...panel.defaultState,
          };
        }
      } else {
        savedStates[panelId] = {
          ...DEFAULT_PANEL_STATE,
          ...panel.defaultState,
        };
      }
    });
    setStates(savedStates);
  }, [panels]);

  // Save panel state to localStorage when it changes
  const saveState = useCallback((panelId: string, state: PanelState) => {
    localStorage.setItem(`panel_${panelId}_state`, JSON.stringify(state));
  }, []);

  // Panel manager implementation
  const manager: PanelManager = {
    register: useCallback((panel: PanelRegistration) => {
      setPanels((prev) => new Map(prev).set(panel.id, panel));

      // Initialize state for new panel
      setStates((prev) => {
        if (prev[panel.id]) return prev; // Already exists

        const saved = localStorage.getItem(`panel_${panel.id}_state`);
        let initialState = { ...DEFAULT_PANEL_STATE, ...panel.defaultState };

        if (saved) {
          try {
            const parsedState = JSON.parse(saved);
            initialState = { ...initialState, ...parsedState };
          } catch {
            // Use default state if parsing fails
          }
        }

        return { ...prev, [panel.id]: initialState };
      });
    }, []),

    unregister: useCallback((panelId: string) => {
      setPanels((prev) => {
        const next = new Map(prev);
        next.delete(panelId);
        return next;
      });
      setStates((prev) => {
        const next = { ...prev };
        delete next[panelId];
        return next;
      });
    }, []),

    getPanels: useCallback(() => {
      return Array.from(panels.values()).sort(
        (a, b) => (a.order ?? 100) - (b.order ?? 100)
      );
    }, [panels]),

    getState: useCallback(
      (panelId: string) => {
        return states[panelId] || { ...DEFAULT_PANEL_STATE };
      },
      [states]
    ),

    updateState: useCallback(
      (panelId: string, updates: Partial<PanelState>) => {
        setStates((prev) => {
          const currentState = prev[panelId] || { ...DEFAULT_PANEL_STATE };
          const newState = { ...currentState, ...updates };
          saveState(panelId, newState);
          return { ...prev, [panelId]: newState };
        });
      },
      [saveState]
    ),

    toggle: useCallback(
      (panelId: string) => {
        setStates((prev) => {
          const currentState = prev[panelId] || { ...DEFAULT_PANEL_STATE };
          const newState = { ...currentState, isOpen: !currentState.isOpen };
          saveState(panelId, newState);
          return { ...prev, [panelId]: newState };
        });
      },
      [saveState]
    ),

    closeAll: useCallback(() => {
      setStates((prev) => {
        const updated: Record<string, PanelState> = {};
        Object.entries(prev).forEach(([panelId, state]) => {
          if (state.isOpen) {
            const newState = { ...state, isOpen: false };
            updated[panelId] = newState;
            saveState(panelId, newState);
          } else {
            updated[panelId] = state;
          }
        });
        return updated;
      });
    }, [saveState]),
  };

  return (
    <PanelRegistryContext.Provider value={{ manager, context, states }}>
      {children}
    </PanelRegistryContext.Provider>
  );
}
