// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  useMemo,
  useState,
  Suspense,
  useCallback,
  useEffect,
  useRef,
} from "react";
import { useMatches } from "@tanstack/react-router";
import { useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";
import { usePanels } from "@/hooks/usePanels";
// useLocalStorageState is now handled by the plugin system
import DragPan from "ol/interaction/DragPan";
import { MaskProvider } from "@/features/viewer/contexts/MaskContext";
import { VectorProvider } from "@/features/viewer/contexts/VectorContext";
import { SlideNavigationProvider } from "@/features/viewer/contexts/SlideNavigationContext";
import { useSlideNavigationData } from "@/features/viewer/hooks/useSlideNavigationData";
import { DisplayMetadata } from "@/features/viewer/components/map/types";
import SlideViewer from "@/features/viewer/components/SlideViewer";
import CaseSlideBar from "@/features/viewer/components/CaseSlideBar";
import type OlMap from "ol/Map";

import { Text } from "ol/style";
import GeoJSON from "ol/format/GeoJSON";

// ViewerSettingsPanel is now handled by the plugin system
// BrightnessContrastPanel is now handled by the plugin system
// AnnotationControlPanel is now handled by the plugin system
// Viewer settings are now handled by the plugin system
// resetBrightnessSettings is now handled by the plugin system
import {
  generateSlideStyle,
  type SlideStyleOptions,
} from "@/features/viewer/components/map/slideStyleUtils";
import { StatusBar } from "@/features/viewer/components/StatusBar";
// AnnotationItem and LabelId types are now only used within the AnnotationControlPlugin
import { useStudyAnnotationSettings } from "@/hooks/useStudyAnnotationSettings";

import "ol-ext/dist/ol-ext.css";
import { useCoordinateTransforms } from "@/features/viewer/components/SlideWorkspace/hooks/useCoordinateTransforms";
import { BrushPolygonTool } from "@/features/viewer/tools/BrushPolygonTool";
import {
  setTileErrorCallback,
  resetTileErrors,
} from "@/features/viewer/components/map/defaultRegistrations";
import { useAnnotationSettings } from "@/features/viewer/hooks/useAnnotationSettings";
import { useAnnotationSync } from "@/features/viewer/hooks/useAnnotationSync";

// Plugin system imports
import { PluginRenderer, PluginDock } from "@/features/viewer/plugins";
import { pluginManager } from "@/features/viewer/plugins/PluginManager";
import { BrightnessControlPlugin } from "@/features/viewer/plugins/BrightnessControlPlugin";
import { AnnotationControlPlugin } from "@/features/viewer/plugins/AnnotationControlPlugin";
import { MaskControlPlugin } from "@/features/viewer/plugins/MaskControlPlugin";
import { RegionControlPlugin } from "@/features/viewer/plugins/RegionControlPlugin";
import { ViewerSettingsPlugin } from "@/features/viewer/plugins/ViewerSettingsPlugin";

import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
} from "@/components/AlertDialog";

interface SlideWorkspaceProps {
  slideUid: string;
  caseUid?: string;
  studyUid?: string;
}

interface FilterParams {
  searchQuery?: string;
  searchName?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  limit?: number;
  isSubsetMode?: boolean;
  hasVectorAnnotations?: string;
  hasRasterAnnotations?: string;
  status?: string;
}

/**
 * Main slide viewer component that handles navigation context and panels.
 * This component manages URL parameters, navigation data, and panel state,
 * while delegating map rendering to SlideViewer.
 */
function SlideWorkspace({ slideUid, caseUid, studyUid }: SlideWorkspaceProps) {
  // Get search parameters from URL
  const matches = useMatches();
  const lastMatch = matches[matches.length - 1];
  const searchParams =
    (lastMatch?.search as {
      q?: string;
      name?: string;
      sort?: string;
      dir?: string;
      page?: number;
      limit?: number;
      subset?: string;
      vector?: string;
      raster?: string;
      status?: string;
    }) || {};

  // Extract filter parameters for navigation
  const filterParams: FilterParams = {
    searchQuery: searchParams.q || "",
    searchName: searchParams.name || "",
    sortBy: searchParams.sort || "",
    sortDir: searchParams.dir || "",
    page: searchParams.page || 1,
    limit: searchParams.limit || 100,
    isSubsetMode: searchParams.subset === "true",
    hasVectorAnnotations: searchParams.vector || "",
    hasRasterAnnotations: searchParams.raster || "",
    status: searchParams.status || "",
  };

  // Fetch navigation data for slide navigation
  const {
    navigationData,
    loading: navigationLoading,
    error: navigationError,
  } = useSlideNavigationData({
    slideUid,
    caseUid,
    ...filterParams,
  });

  // Determine if we have navigation context (case or study based navigation)
  const hasNavigationContext = Boolean(caseUid || studyUid || navigationData);

  // State for panels and metadata
  const { visiblePanels, togglePanel, closeAllPanels } = usePanels();
  const [slideMetadata, setSlideMetadata] = useState<DisplayMetadata | null>(
    null
  );

  const [isLoading, setIsLoading] = useState(false);
  const [tileProgress, setTileProgress] = useState<{
    inFlight: number;
    loaded: number;
    errors: number;
    started: number;
  }>({ inFlight: 0, loaded: 0, errors: 0, started: 0 });
  const [slideLayer, setSlideLayer] = useState<any>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  // Brightness control is now handled by the plugin system
  const [showTileErrorDialog, setShowTileErrorDialog] = useState(false);
  const [tileErrorCount, setTileErrorCount] = useState(0);
  const [showServerError, setShowServerError] = useState(false);
  const [serverErrorMessage, setServerErrorMessage] = useState("");
  // Annotation panel state is now managed by the plugin system

  // Fetch annotation settings from the study
  const {
    data: annotationSettings,
    isLoading: annotationSettingsLoading,
    error: annotationSettingsError,
  } = useStudyAnnotationSettings(studyUid);

  // Debug logging for annotation settings
  useEffect(() => {
    if (annotationSettings) {
      console.log(
        "SlideWorkspace - Full annotation settings:",
        annotationSettings
      );
      console.log(
        "SlideWorkspace - Annotations array:",
        annotationSettings.annotations
      );
    }
    if (annotationSettingsError) {
      console.error(
        "SlideWorkspace - Error loading annotation settings:",
        annotationSettingsError
      );
    }
  }, [annotationSettings, annotationSettingsError]);

  // Annotation display settings
  const {
    settings: annotationDisplaySettings,
    updateSettings: updateAnnotationDisplaySettings,
    resetSettings: resetAnnotationDisplaySettings,
  } = useAnnotationSettings();

  // Annotation sync functionality
  const {
    syncState,
    loadAnnotations,
    manualSync,
    isLoading: isSyncLoading,
    isSyncing,
    lastSyncTime,
    error: syncError,
  } = useAnnotationSync(slideUid);

  // Keep latest display settings in a ref so OL style functions can read live values
  const annotationDisplaySettingsRef = useRef(annotationDisplaySettings);
  useEffect(() => {
    annotationDisplaySettingsRef.current = annotationDisplaySettings;
  }, [annotationDisplaySettings]);

  // Label color mapping is now handled by the AnnotationControlPlugin
  // Annotation state is now fully managed by the AnnotationControlPlugin

  // Region styles and styling functions are now handled by the RegionControlPlugin

  // Annotation dock state is now handled by the plugin system

  // Register plugins based on study settings
  useEffect(() => {
    // Conditionally register plugins based on study settings
    const pluginSettings = annotationSettings?.pluginSettings;

    // ViewerSettings plugin - now toggleable
    if (pluginSettings?.viewerSettings !== false) {
      pluginManager.registerPlugin(ViewerSettingsPlugin);
    }

    if (pluginSettings?.brightnessControl !== false) {
      pluginManager.registerPlugin(BrightnessControlPlugin);
    }

    if (pluginSettings?.annotationControl !== false) {
      pluginManager.registerPlugin(AnnotationControlPlugin);
    }

    if (pluginSettings?.maskControl !== false) {
      pluginManager.registerPlugin(MaskControlPlugin);
    }

    if (pluginSettings?.regionControl === true) {
      pluginManager.registerPlugin(RegionControlPlugin);
    }

    return () => {
      // Cleanup all plugins on unmount
      pluginManager.unregisterPlugin(ViewerSettingsPlugin.id);
      pluginManager.unregisterPlugin(BrightnessControlPlugin.id);
      pluginManager.unregisterPlugin(AnnotationControlPlugin.id);
      pluginManager.unregisterPlugin(MaskControlPlugin.id);
      pluginManager.unregisterPlugin(RegionControlPlugin.id);
    };
  }, [annotationSettings?.pluginSettings]);

  // Keyboard shortcuts - only handle panel controls
  useKeyboardShortcuts({
    onKeyS: () => togglePanel("slideInfo"),
    onKeyH: () => togglePanel("help"),
    onKeyEscape: closeAllPanels,
  });

  const [mapRef, setMapRef] = useState<OlMap | null>(null);
  const [rawSlideMetadata, setRawSlideMetadata] = useState<any>(null);
  // Annotation-related refs are now managed by the AnnotationPlugin
  const dragPanInteractionRef = useRef<any>(null);

  // Annotation interaction helpers are now managed by the AnnotationPlugin

  // Region modify interactions are now handled by the RegionControlPlugin

  // Draw interaction helpers are now managed by the AnnotationPlugin

  // Viewer state is now handled by the plugin system
  const isFluorescent = rawSlideMetadata?.imageTypeId === "img_type_fluor";

  // Viewer settings state - updated by ViewerSettingsPlugin
  const [viewerSettings, setViewerSettings] = useState({
    panSensitivity: 1.0,
    zoomSensitivity: 1.0,
    quality: undefined as number | undefined,
    showMeasurementBar: true,
  });

  const handleViewerSettingsChange = useCallback(
    (settings: typeof viewerSettings) => {
      setViewerSettings(settings);
    },
    []
  );

  const handleMapReady = useCallback((m: OlMap) => {
    setMapRef(m);
  }, []);

  useEffect(() => {
    if (!mapRef) return;
    const drag =
      mapRef
        .getInteractions()
        .getArray()
        .find((i): i is DragPan => i instanceof DragPan) || null;
    dragPanInteractionRef.current = drag;
  }, [mapRef]);

  const handleSlideLayerCreated = useCallback((layer: any) => {
    setSlideLayer(layer);
  }, []);

  const handleTileProgress = useCallback(
    (p: {
      inFlight: number;
      loaded: number;
      errors: number;
      started: number;
    }) => {
      setTileProgress(p);
      setIsRefreshing(p.inFlight > 0);
    },
    []
  );

  const handleMaskLayerCreated = useCallback(() => {
    // Empty callback
  }, []);

  const handleVectorLayerCreated = useCallback(() => {
    // Empty callback
  }, []);

  // Store raw metadata for coordinate transformations
  const handleRawMetadataLoaded = useCallback(
    (metadata: any) => {
      setRawSlideMetadata(metadata);
    },
    [slideUid]
  );

  // Set up tile error callback when component mounts
  useEffect(() => {
    setTileErrorCallback((slideUidFromError, errorCount) => {
      if (slideUidFromError === slideUid) {
        setTileErrorCount(errorCount);
        setShowTileErrorDialog(true);
      }
    });

    // Clean up when component unmounts
    return () => {
      setTileErrorCallback(() => {});
    };
  }, [slideUid]);

  // Handle tile error dialog actions
  const handleRetryTileLoading = useCallback(async () => {
    try {
      // First check if the server is accessible by making a simple API call
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

      const response = await fetch("/api/status", {
        method: "GET",
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (response.ok) {
        // Server is accessible, safe to reload
        resetTileErrors(slideUid);
        setShowTileErrorDialog(false);
        setTileErrorCount(0);
        window.location.reload();
      } else {
        // Server returned an error
        setShowServerError(true);
        setServerErrorMessage(
          `Server returned status ${response.status}: ${response.statusText}`
        );
      }
    } catch (error) {
      // Network error or server unreachable
      setShowServerError(true);

      if (error instanceof Error && error.name === "AbortError") {
        // Timeout - don't log this as it's expected behavior
        setServerErrorMessage(
          "Server is not responding. Please check your network connection and try again later."
        );
      } else if (
        error instanceof TypeError &&
        error.message.includes("fetch")
      ) {
        // Network/fetch error - don't log this as it's expected when server is down
        setServerErrorMessage(
          "Unable to connect to the server. Please check your network connection."
        );
      } else {
        // Unexpected error - log this one for debugging
        console.warn(
          "Unexpected error during server connectivity check:",
          error
        );
        setServerErrorMessage(
          "Server is currently unavailable. Please try again later."
        );
      }
    }
  }, [slideUid]);

  const handleCloseTileErrorDialog = useCallback(() => {
    setShowTileErrorDialog(false);
  }, []);

  const handleCloseServerErrorDialog = useCallback(() => {
    setShowServerError(false);
    setServerErrorMessage("");
  }, []);

  // Update plugin context when relevant data changes
  useEffect(() => {
    pluginManager.updateContext({
      slideUid,
      studyUid,
      caseUid,
      map: mapRef,
      slideMetadata,
      rawSlideMetadata,
      slideLayer,
      // Study annotation settings for mask/vector control
      studyAnnotationSettings: annotationSettings,
      studyAnnotationSettingsLoading: annotationSettingsLoading,
      studyAnnotationSettingsError: annotationSettingsError,
      // Viewer settings callback for ViewerSettingsPlugin
      onViewerSettingsChange: handleViewerSettingsChange,
      // All annotation state and functions are now managed internally by the AnnotationControlPlugin
      // Region-related context is now managed internally by the RegionControlPlugin
    });
  }, [
    slideUid,
    studyUid,
    caseUid,
    mapRef,
    slideMetadata,
    rawSlideMetadata,
    slideLayer,
    annotationSettings,
    annotationSettingsLoading,
    annotationSettingsError,
    handleViewerSettingsChange,
  ]);

  const slideViewerComponent = (
    <>
      <SlideViewer
        slideUid={slideUid}
        studyUid={studyUid}
        onMetadataLoaded={setSlideMetadata}
        onMapReady={handleMapReady}
        onRawMetadataLoaded={handleRawMetadataLoaded}
        onSlideLayerCreated={handleSlideLayerCreated}
        onMaskLayerCreated={handleMaskLayerCreated}
        onVectorLayerCreated={handleVectorLayerCreated}
        onLoadingChange={setIsLoading}
        onTileProgress={handleTileProgress}
        panSensitivity={viewerSettings.panSensitivity}
        zoomSensitivity={viewerSettings.zoomSensitivity}
        quality={viewerSettings.quality}
        showMeasurementBar={viewerSettings.showMeasurementBar}
      />
      {/* Floating panels overlaying the viewer */}
      <div className="absolute inset-0 pointer-events-none">
        {/* Plugin system renders floating panels only */}
        <PluginRenderer renderMode="floating" />
        {/* Annotation Panel is now handled by the plugin system */}
        {/* Region Panel is now handled by the plugin system */}
      </div>
    </>
  );

  const content = (
    <div
      className="w-full flex flex-col overflow-hidden min-h-0 viewer-no-scroll"
      style={{
        height: "calc(100dvh - 44px)",
      }}
    >
      <div className="relative flex-1 min-h-0 overflow-hidden">
        {/* Left annotations sidebar within viewer area */}
        <div className="absolute inset-0 flex">
          {/* Plugin system renders docked panels only */}
          <PluginRenderer renderMode="docked" />
          {/* Annotation panel is now handled by the plugin system */}
          {/* Region panel is now handled by the plugin system */}
          <div className="flex-1 relative min-h-0">
            {slideViewerComponent}
            {/* Plugin dock for all controls */}
            <div className="absolute bottom-4 right-4 z-30 overflow-visible">
              <PluginDock />
            </div>
          </div>
        </div>
      </div>
      {hasNavigationContext && (
        <CaseSlideBar
          className="h-20 md:h-24 bg-gray-900 border-t border-gray-700 shrink-0"
          mapRef={mapRef as any}
          filterParams={filterParams}
        />
      )}
      {/* Thin bottom status bar */}
      <div className="shrink-0">
        <StatusBar
          tileProgress={tileProgress}
          isRefreshing={isRefreshing}
          syncStatus={
            syncError !== undefined
              ? {
                  isSyncing,
                  lastSyncTime,
                  error: syncError,
                  onManualSync: async () => {
                    // Trigger global sync event for all plugins (including regions)
                    window.dispatchEvent(new CustomEvent("globalSync"));

                    // Also sync annotations directly
                    await manualSync();
                    return;
                  },
                }
              : undefined
          }
        />
      </div>

      {/* Brightness control is now handled by the plugin system */}

      {/* Alert dialog for persistent tile loading errors */}
      <AlertDialog open={showTileErrorDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Having trouble loading this slide
            </AlertDialogTitle>
            <AlertDialogDescription>
              We're having some difficulty loading parts of this slide. You can
              continue viewing what's already loaded, or try refreshing to load
              the missing parts.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCloseTileErrorDialog}>
              Continue Viewing
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleRetryTileLoading}>
              Try Again
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Alert dialog for server connectivity issues */}
      <AlertDialog open={showServerError}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Can't connect right now</AlertDialogTitle>
            <AlertDialogDescription>
              We're having trouble reaching the server. This might be temporary
              - try waiting a moment and checking again.
              <br />
              <br />
              If this keeps happening, you might want to check your internet
              connection or contact support.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCloseServerErrorDialog}>
              Close
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleRetryTileLoading}>
              Check Again
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );

  // Conditionally wrap with SlideNavigationProvider if we have navigation context
  if (hasNavigationContext) {
    return (
      <MaskProvider>
        <VectorProvider>
          <SlideNavigationProvider
            currentSlideUid={slideUid}
            slides={navigationData?.slides || []}
            sourceContext={navigationData?.source}
          >
            {content}
          </SlideNavigationProvider>
        </VectorProvider>
      </MaskProvider>
    );
  }

  // Simple viewer without navigation context
  return (
    <MaskProvider>
      <VectorProvider>{content}</VectorProvider>
    </MaskProvider>
  );
}

export default function SlideWorkspaceRoute() {
  const matches = useMatches();
  const params = matches[matches.length - 1]?.params as {
    slideUid?: string;
    caseUid?: string;
    studyUid?: string;
  };

  if (!params?.slideUid) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-gray-900 text-white">
        <div className="text-center">
          <h1>Missing Slide ID</h1>
          <p>No slide ID provided in the URL parameters.</p>
        </div>
      </div>
    );
  }

  return (
    <SlideWorkspace
      slideUid={params.slideUid}
      caseUid={params.caseUid}
      studyUid={params.studyUid}
    />
  );
}
