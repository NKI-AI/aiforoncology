// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useEffect, useState, useCallback } from "react";
import Map from "ol/Map";

export type CrosshairMode = "follow" | "lock";

interface CoordinateTrackerProps {
  map: Map | null;
  mapContainer: HTMLDivElement | null;
  onPositionChange: (position: { x: number; y: number }) => void;
  onCoordinatesChange: (coordinates: { x: string; y: string }) => void;
  mode?: CrosshairMode;
  onModeChange?: (mode: CrosshairMode) => void;
}

/**
 * Component that tracks mouse/touch coordinates on the map with follow and lock modes
 */
export default function CoordinateTracker({
  map,
  mapContainer,
  onPositionChange,
  onCoordinatesChange,
  mode = "follow",
  onModeChange,
}: CoordinateTrackerProps) {
  const [isTouch, setIsTouch] = useState(false);
  const [isPanning, setIsPanning] = useState(false);
  const [lockedMapCoordinate, setLockedMapCoordinate] = useState<
    [number, number] | null
  >(null);
  const [isLKeyPressed, setIsLKeyPressed] = useState(false);

  // Detect if device supports touch
  useEffect(() => {
    setIsTouch("ontouchstart" in window || navigator.maxTouchPoints > 0);
  }, []);

  // Update coordinates based on map coordinate
  const updateCoordinatesFromMapCoord = useCallback(
    (coordinate: number[]) => {
      const mapCoord = coordinate as [number, number];
      const xCoord = mapCoord[0] * 1000; // Convert meters to mm
      const yCoord = Math.abs(mapCoord[1]) * 1000; // Convert meters to mm

      if (!isNaN(xCoord) && !isNaN(yCoord)) {
        onCoordinatesChange({
          x: xCoord.toFixed(2),
          y: yCoord.toFixed(2),
        });
      }
    },
    [onCoordinatesChange]
  );

  // Update position and coordinates from screen position
  const updateFromScreenPosition = useCallback(
    (x: number, y: number) => {
      if (!map) return;

      onPositionChange({ x, y });

      const coordinate = map.getCoordinateFromPixel([x, y]);
      if (coordinate) {
        updateCoordinatesFromMapCoord(coordinate);
      }
    },
    [map, onPositionChange, updateCoordinatesFromMapCoord]
  );

  // Update screen position from locked map coordinate
  const updateFromLockedCoordinate = useCallback(() => {
    if (!map || !lockedMapCoordinate) return;

    const pixel = map.getPixelFromCoordinate(lockedMapCoordinate);
    if (pixel) {
      onPositionChange({ x: pixel[0], y: pixel[1] });
      updateCoordinatesFromMapCoord(lockedMapCoordinate);
    }
  }, [
    map,
    lockedMapCoordinate,
    onPositionChange,
    updateCoordinatesFromMapCoord,
  ]);

  // Get event position from mouse or touch event
  const getEventPosition = useCallback(
    (e: MouseEvent | TouchEvent) => {
      if (!mapContainer) return null;

      const rect = mapContainer.getBoundingClientRect();
      let clientX: number, clientY: number;

      if (e instanceof TouchEvent && e.touches.length > 0) {
        clientX = e.touches[0].clientX;
        clientY = e.touches[0].clientY;
      } else if (e instanceof MouseEvent) {
        clientX = e.clientX;
        clientY = e.clientY;
      } else {
        return null;
      }

      return {
        x: clientX - rect.left,
        y: clientY - rect.top,
      };
    },
    [mapContainer]
  );

  // Lock crosshair to a specific map coordinate
  const lockCrosshairAt = useCallback(
    (screenX: number, screenY: number) => {
      if (!map) return;

      // Get the map coordinate at this screen position
      const mapCoord = map.getCoordinateFromPixel([screenX, screenY]);
      if (!mapCoord) return;

      if (mode === "follow") {
        // Switch to lock mode and store the map coordinate
        setLockedMapCoordinate(mapCoord as [number, number]);
        updateFromScreenPosition(screenX, screenY);
        onModeChange?.("lock");
      } else if (mode === "lock") {
        // Update locked position to new map coordinate
        setLockedMapCoordinate(mapCoord as [number, number]);
        updateFromScreenPosition(screenX, screenY);
      }
    },
    [mode, map, updateFromScreenPosition, onModeChange]
  );

  // Handle pointer events with debouncing
  const handlePointerMove = useCallback(
    (e: MouseEvent | TouchEvent) => {
      // Only update in follow mode and when not panning
      if (mode === "lock" || isPanning) return;

      const position = getEventPosition(e);
      if (!position) return;

      // Debounce using requestAnimationFrame
      requestAnimationFrame(() => {
        updateFromScreenPosition(position.x, position.y);
      });
    },
    [mode, isPanning, getEventPosition, updateFromScreenPosition]
  );

  // Handle click events with L key modifier
  const handleClick = useCallback(
    (e: MouseEvent | TouchEvent) => {
      // Only lock if L key is pressed
      if (!isLKeyPressed) return;

      const position = getEventPosition(e);
      if (!position) return;

      lockCrosshairAt(position.x, position.y);
      e.preventDefault();
    },
    [isLKeyPressed, getEventPosition, lockCrosshairAt]
  );

  // Handle keyboard events
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Ignore if typing in input fields
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      switch (e.key.toLowerCase()) {
        case "l":
          setIsLKeyPressed(true);
          e.preventDefault();
          break;
        case "arrowup":
        case "arrowdown":
        case "arrowleft":
        case "arrowright":
          // Handle arrow key movement in lock mode
          if (mode !== "lock" || !map || !mapContainer || !lockedMapCoordinate)
            break;

          const step = e.shiftKey ? 10 : 1;
          const currentPixel = map.getPixelFromCoordinate(lockedMapCoordinate);
          if (!currentPixel) break;

          let newX = currentPixel[0];
          let newY = currentPixel[1];

          switch (e.key) {
            case "ArrowUp":
              newY = Math.max(0, newY - step);
              break;
            case "ArrowDown":
              newY = Math.min(mapContainer.clientHeight, newY + step);
              break;
            case "ArrowLeft":
              newX = Math.max(0, newX - step);
              break;
            case "ArrowRight":
              newX = Math.min(mapContainer.clientWidth, newX + step);
              break;
          }

          // Convert new screen position back to map coordinate
          const newMapCoord = map.getCoordinateFromPixel([newX, newY]);
          if (newMapCoord) {
            setLockedMapCoordinate(newMapCoord as [number, number]);
            updateFromScreenPosition(newX, newY);
          }
          e.preventDefault();
          break;
        case "enter":
        case "escape":
          // Return to follow mode and clear locked coordinate
          if (mode === "lock") {
            setLockedMapCoordinate(null);
            onModeChange?.("follow");
            e.preventDefault();
          }
          break;
      }
    },
    [
      mode,
      map,
      mapContainer,
      lockedMapCoordinate,
      updateFromScreenPosition,
      onModeChange,
    ]
  );

  const handleKeyUp = useCallback((e: KeyboardEvent) => {
    // Ignore if typing in input fields
    if (
      e.target instanceof HTMLInputElement ||
      e.target instanceof HTMLTextAreaElement
    ) {
      return;
    }

    if (e.key.toLowerCase() === "l") {
      setIsLKeyPressed(false);
    }
  }, []);

  // Handle pan start detection
  const handlePanStart = useCallback(() => {
    setIsPanning(true);
  }, []);

  // Handle pan end detection
  const handlePanEnd = useCallback(() => {
    setIsPanning(false);
    // Update locked crosshair position after pan
    if (mode === "lock") {
      updateFromLockedCoordinate();
    }
  }, [mode, updateFromLockedCoordinate]);

  // Handle map view changes (pan/zoom) to update locked crosshair position
  const handleMapViewChange = useCallback(() => {
    if (mode === "lock") {
      updateFromLockedCoordinate();
    }
  }, [mode, updateFromLockedCoordinate]);

  // Clear locked coordinate when switching back to follow mode
  useEffect(() => {
    if (mode === "follow") {
      setLockedMapCoordinate(null);
    }
  }, [mode]);

  useEffect(() => {
    if (!map || !mapContainer) return;

    // Mouse events for desktop
    if (!isTouch) {
      mapContainer.addEventListener("mousemove", handlePointerMove);
      mapContainer.addEventListener("click", handleClick);
    }

    // Touch events for mobile
    if (isTouch) {
      mapContainer.addEventListener("touchmove", handlePointerMove, {
        passive: true,
      });
      mapContainer.addEventListener("touchstart", handleClick, {
        passive: false,
      });
    }

    // Map pan detection (works for both mouse and touch)
    map.on("movestart", handlePanStart);
    map.on("moveend", handlePanEnd);

    // Listen to map view changes to update locked crosshair
    const view = map.getView();
    view.on("change:center", handleMapViewChange);
    view.on("change:resolution", handleMapViewChange);

    // Global keyboard events
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);

    return () => {
      if (mapContainer) {
        mapContainer.removeEventListener("mousemove", handlePointerMove);
        mapContainer.removeEventListener("click", handleClick);
        mapContainer.removeEventListener("touchmove", handlePointerMove);
        mapContainer.removeEventListener("touchstart", handleClick);
      }

      if (map) {
        map.un("movestart", handlePanStart);
        map.un("moveend", handlePanEnd);

        const view = map.getView();
        view.un("change:center", handleMapViewChange);
        view.un("change:resolution", handleMapViewChange);
      }

      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [
    map,
    mapContainer,
    isTouch,
    handlePointerMove,
    handleClick,
    handlePanStart,
    handlePanEnd,
    handleMapViewChange,
    handleKeyDown,
    handleKeyUp,
  ]);

  // This component doesn't render anything
  return null;
}
