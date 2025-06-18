// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
export interface SlideMetadata {
  slideId: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
  tileSize?: number;
  format?: string;
  magnification?: string;
  vendor?: string;
}

/**
 * Interface for display metadata shown in the UI
 */
export interface DisplayMetadata {
  resolution: string;
  objective: string;
  vendor: string;
  digitalSize: string;
  physicalDimensions: string;
  slideId: string;
  slideMpp: string;
}

/**
 * Interface for a mask
 */
export interface Mask {
  maskId: string;
  maskName: string;
  maskUri: string;
  tilesUrl: string;
  slideId: string;
  maskWidth: number;
  maskHeight: number;
  maskMpp: number;
}

/**
 * Interface for a list of masks
 */
export interface MaskList {
  slide_id: string;
  masks: Mask[];
}

/**
 * Interface for mask metadata
 */
export interface MaskMetadata {
  id: string;
  type: string;
  name: string;
  description: string;
  created: string;
  author: string;
  has_mask: boolean;
  mask_color_map: Record<string, number[]>;
  slide_id: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
}

/**
 * Interface for error types to improve error handling
 */
export class FetchError extends Error {
  status: number;
  statusText: string;

  constructor(message: string, status: number, statusText: string) {
    super(message);
    this.name = "FetchError";
    this.status = status;
    this.statusText = statusText;
  }
}

export class ConfigurationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ConfigurationError";
  }
}
