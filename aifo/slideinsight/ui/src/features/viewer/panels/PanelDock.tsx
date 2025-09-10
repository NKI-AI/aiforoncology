// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useEffect } from "react";
import { usePanelRegistry } from "./PanelRegistry";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/**
 * Panel dock that renders toggle buttons for all registered panels
 */
export function PanelDock({ className }: { className?: string }) {
  const { manager, states } = usePanelRegistry();
  const panels = manager.getPanels().filter((panel) => panel.enabled !== false);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Only handle shortcuts when not in input/textarea
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.contentEditable === "true"
      ) {
        return;
      }

      panels.forEach((panel) => {
        if (
          panel.shortcut &&
          e.key.toLowerCase() === panel.shortcut.toLowerCase()
        ) {
          e.preventDefault();
          manager.toggle(panel.id);
        }
      });

      // Escape to close all panels
      if (e.key === "Escape") {
        manager.closeAll();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [panels, manager]);

  return (
    <div className={className}>
      <TooltipProvider>
        {panels.map((panel) => {
          const state = states[panel.id];
          const isActive = state?.isOpen || false;

          return (
            <Tooltip key={panel.id}>
              <TooltipTrigger asChild>
                <button
                  onClick={() => manager.toggle(panel.id)}
                  className={`
                    p-2 rounded-md transition-colors
                    ${
                      isActive
                        ? "bg-primary text-primary-foreground shadow-md"
                        : "bg-muted hover:bg-accent text-muted-foreground hover:text-foreground"
                    }
                  `}
                  aria-label={`Toggle ${panel.name}`}
                  title={`${panel.name}${
                    panel.shortcut ? ` (${panel.shortcut.toUpperCase()})` : ""
                  }`}
                >
                  <panel.icon className="h-5 w-5" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="right">
                <div className="text-center">
                  <div>{panel.name}</div>
                  {panel.shortcut && (
                    <div className="text-xs opacity-75">
                      Press {panel.shortcut.toUpperCase()}
                    </div>
                  )}
                </div>
              </TooltipContent>
            </Tooltip>
          );
        })}
      </TooltipProvider>
    </div>
  );
}

/**
 * Panel renderer that renders all open panels
 */
export function PanelRenderer() {
  const { manager, context, states } = usePanelRegistry();
  const panels = manager.getPanels().filter((panel) => panel.enabled !== false);

  const updatePanelState = useCallback(
    (panelId: string) => {
      return (updates: Partial<(typeof states)[string]>) => {
        manager.updateState(panelId, updates);
      };
    },
    [manager]
  );

  const closePanelHandler = useCallback(
    (panelId: string) => {
      return () => {
        manager.updateState(panelId, { isOpen: false });
      };
    },
    [manager]
  );

  // Render docked panels (left dock)
  const dockedPanels = panels.filter((panel) => {
    const state = states[panel.id];
    return state?.isOpen && state?.dock === "left";
  });

  // Render floating panels
  const floatingPanels = panels.filter((panel) => {
    const state = states[panel.id];
    return state?.isOpen && state?.dock === "free";
  });

  return (
    <>
      {/* Docked panels - rendered in the left sidebar */}
      {dockedPanels.map((panel) => {
        const state = states[panel.id];
        const PanelComponent = panel.component;

        return (
          <PanelComponent
            key={`docked-${panel.id}`}
            context={context}
            state={state}
            updateState={updatePanelState(panel.id)}
            onClose={closePanelHandler(panel.id)}
          />
        );
      })}

      {/* Floating panels - rendered as overlays */}
      {floatingPanels.map((panel) => {
        const state = states[panel.id];
        const PanelComponent = panel.component;

        return (
          <div
            key={`floating-${panel.id}`}
            className="absolute inset-0 pointer-events-none"
          >
            <PanelComponent
              context={context}
              state={state}
              updateState={updatePanelState(panel.id)}
              onClose={closePanelHandler(panel.id)}
            />
          </div>
        );
      })}
    </>
  );
}

/**
 * Hook to get information about docked panels for layout calculations
 */
export function useDockedPanels() {
  const { manager, states } = usePanelRegistry();
  const panels = manager.getPanels().filter((panel) => panel.enabled !== false);

  const dockedPanels = panels.filter((panel) => {
    const state = states[panel.id];
    return state?.isOpen && state?.dock === "left";
  });

  const totalDockedWidth = dockedPanels.reduce((total, panel) => {
    const state = states[panel.id];
    return total + (state?.size?.width || 320);
  }, 0);

  return {
    dockedPanels,
    totalDockedWidth,
    hasDockedPanels: dockedPanels.length > 0,
  };
}
