// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import "./Crosshair.css";

interface CrosshairProps {
  show: boolean;
  position: { x: number; y: number };
  coordinates: { x: string; y: string };
}

/**
 * Component that displays a crosshair at the current cursor position
 */
export default function Crosshair({
  show,
  position,
  coordinates,
}: CrosshairProps) {
  if (!show) return null;

  return (
    <>
      <div className="crosshair">
        <div
          id="vertical-line"
          className="vertical"
          style={{ transform: `translateX(${position.x}px)` }}
        ></div>
        <div
          id="horizontal-line"
          className="horizontal"
          style={{ transform: `translateY(${position.y}px)` }}
        ></div>
        <div
          id="center-dot"
          className="center"
          style={{
            left: `${position.x}px`,
            top: `${position.y}px`,
          }}
        ></div>
      </div>
      <div className="coordinates-display">
        {coordinates.x} mm, {coordinates.y} mm
      </div>
    </>
  );
}
