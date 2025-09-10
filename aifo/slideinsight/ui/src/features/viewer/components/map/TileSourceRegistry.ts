// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// Registry for mapping tile formats to OpenLayers tile source factories.
// Keeping this in a separate file avoids circular deps and enables pluggability.

import type ImageTileSource from "ol/source/ImageTile.js";
import type DataTileSource from "ol/source/DataTile.js";
import type TileGrid from "ol/tilegrid/TileGrid.js";
import type Projection from "ol/proj/Projection.js";
import type { Size } from "ol/size.js";

export type TileFormat = "png" | "jpg" | "jxl";

export interface BaseImageContext {
  getProjection(): Projection;
  getTileGrid(): TileGrid;
  readonly tilePixelSize: Size;
  readonly crossOrigin: string;
  readonly isFluorescent: boolean;
  readonly tileUrlFunction?: (z: number, x: number, y: number) => string;
}

export interface WholeSlideImageFactoryOptions {
  interpolate?: boolean;
  transition?: number;
  attributions?: string | string[];
  bandCount?: number;
}

export type TileSourceFactory = (
  options: WholeSlideImageFactoryOptions,
  base: BaseImageContext
) => ImageTileSource | DataTileSource<any>;

const sourceRegistry = new Map<TileFormat, TileSourceFactory>();

export function registerTileSource(
  format: TileFormat,
  factory: TileSourceFactory
): void {
  sourceRegistry.set(format, factory);
}

export function unregisterTileSource(format: TileFormat): void {
  sourceRegistry.delete(format);
}

export function hasTileSource(format: TileFormat): boolean {
  return sourceRegistry.has(format);
}

export function getTileSourceFactory(format: TileFormat): TileSourceFactory {
  const factory = sourceRegistry.get(format);
  if (!factory) {
    throw new Error(`No tile source registered for format: ${format}`);
  }
  return factory;
}

export function getRegisteredTileFormats(): TileFormat[] {
  return Array.from(sourceRegistry.keys());
}
