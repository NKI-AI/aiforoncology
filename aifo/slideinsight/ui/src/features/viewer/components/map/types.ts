// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface Channel {
  name: string;
  biomarker: string;
  color: string;
}

export interface SlideMetadata {
  slideUid: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
  imageTypeId?: string; // Image type identifier (e.g., "img_type_fluor")
  tileSize?: number;
  format?: string;
  magnification?: string;
  vendor?: string;
  metadata?: {
    channels?: Record<string, Channel>;
  };
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
  slideUid: string;
  slideMpp: string;
}

/**
 * Interface for a mask
 */
export interface Mask {
  maskUid: string;
  maskName: string;
  tilesUrl: string;
  slideUid: string;
  maskWidth: number;
  maskHeight: number;
  maskMpp: number;
  labels?: Array<{
    name: string;
    index: number;
    color: string;
  }>;
}

/**
 * Interface for a list of masks
 */
export interface MaskList {
  SlideUID: string;
  masks: Mask[];
}

/**
 * Interface for mask metadata
 */
interface MaskMetadata {
  id: string;
  type: string;
  name: string;
  description: string;
  created: string;
  author: string;
  has_mask: boolean;
  mask_color_map: Record<string, number[]>;
  slideUid: string;
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
