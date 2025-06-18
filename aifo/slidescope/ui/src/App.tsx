// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useState, useEffect, Suspense } from "react";
import { useParams, useLocation } from "react-router-dom";
import React from "react";
import Navbar from "./components/Navbar";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import { usePanels } from "./hooks/usePanels";
import { MaskProvider } from "./contexts/MaskContext";

// Lazy load the heavy components
const SlideViewer = React.lazy(() => import("./components/SlideViewer"));
const Panels = React.lazy(() => import("./components/Panels"));
const SlideSelector = React.lazy(() => import("./components/SlideSelector"));

export default function App() {
  // Get slideId from URL parameters using React Router
  const { slideId } = useParams<{ slideId?: string }>();
  const location = useLocation();

  useEffect(() => {
    console.log("[DEBUG] Current route:", location.pathname);
    console.log("[DEBUG] Current slide ID:", slideId);
  }, [location, slideId]);

  // Panel state management
  const { visiblePanels, togglePanel, closeAllPanels } = usePanels();

  const [slideMetadata, setSlideMetadata] = useState<any>(null);
  const [showCrosshair, setShowCrosshair] = useState(false);

  // Keyboard shortcuts - only handle panel and crosshair controls
  // Mask-related shortcuts are handled in MaskContext
  useKeyboardShortcuts({
    onKeyM: () => togglePanel("maskControl"),
    onKeyS: () => togglePanel("slideInfo"),
    onKeyH: () => togglePanel("help"),
    onKeyEscape: closeAllPanels,
    onKeyC: () => setShowCrosshair((prev) => !prev),
  });

  const toggleCrosshair = () => {
    setShowCrosshair((prev) => !prev);
  };

  // Check if we have a slideId to determine which component to render
  const showSlideViewer = !!slideId;

  return (
    <MaskProvider>
      <div className="bg-gray-100 text-gray-800 font-sans h-full w-full flex flex-col">
        <Navbar
          onToggleMaskControl={() => togglePanel("maskControl")}
          onToggleSlideInfo={() => togglePanel("slideInfo")}
          onToggleHelp={() => togglePanel("help")}
          onToggleCrosshair={toggleCrosshair}
          showCrosshair={showCrosshair}
        />

        <div className="relative flex-1">
          {showSlideViewer ? (
            <>
              <Suspense
                fallback={
                  <div className="absolute inset-0 pt-12 bg-black flex items-center justify-center">
                    <div className="text-white text-lg">Loading panels...</div>
                  </div>
                }
              >
                <Panels
                  visiblePanels={visiblePanels}
                  togglePanel={togglePanel}
                  slideMetadata={slideMetadata}
                />
              </Suspense>
              <Suspense
                fallback={
                  <div className="absolute inset-0 pt-12 bg-black flex items-center justify-center">
                    <div className="text-white text-lg">
                      Loading slide viewer...
                    </div>
                  </div>
                }
              >
                <SlideViewer
                  slideId={slideId}
                  showCrosshair={showCrosshair}
                  onMetadataLoaded={setSlideMetadata}
                />
              </Suspense>
            </>
          ) : (
            <Suspense
              fallback={
                <div className="flex items-center justify-center min-h-full">
                  <div className="text-lg text-gray-700">
                    Loading slide selector...
                  </div>
                </div>
              }
            >
              <SlideSelector />
            </Suspense>
          )}
        </div>
      </div>
    </MaskProvider>
  );
}
