// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { apiClient } from "./apiClient";

export interface RegionGeometry {
  type: string;
  coordinates: any;
}

export interface Region {
  regionId: string;
  regionName: string;
  slideUid: string;
  regionType:
    | "roi"
    | "patient"
    | "tissue"
    | "artifact"
    | "background"
    | "other";
  geometry: RegionGeometry;
  coordinateSystem: "pixel" | "physical";
  areaPixels?: number;
  areaPhysical?: number;
  labels?: any;
  metadata?: any;
  styleConfig?: any;
  actorType?: string;
  actorId?: number;
  mutable: boolean;
  visible: boolean;
  deletedAt?: string;
  deletedBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface RegionList {
  slideUid: string;
  regions: Region[];
  totalCount: number;
  hasMore: boolean;
  nextCursor?: string;
}

export interface CreateRegionRequest {
  regionName: string;
  regionType:
    | "roi"
    | "patient"
    | "tissue"
    | "artifact"
    | "background"
    | "other";
  geometry: RegionGeometry;
  coordinateSystem: "pixel" | "physical";
  areaPixels?: number;
  areaPhysical?: number;
  labels?: any;
  metadata?: any;
  styleConfig?: any;
  mutable?: boolean;
  visible?: boolean;
}

export interface UpdateRegionRequest {
  regionName?: string;
  regionType?:
    | "roi"
    | "patient"
    | "tissue"
    | "artifact"
    | "background"
    | "other";
  geometry?: RegionGeometry;
  coordinateSystem?: "pixel" | "physical";
  areaPixels?: number;
  areaPhysical?: number;
  labels?: any;
  metadata?: any;
  styleConfig?: any;
  mutable?: boolean;
  visible?: boolean;
}

export interface BulkCreateRegionsRequest {
  regions: CreateRegionRequest[];
}

export interface BulkUpdateRegionsRequest {
  updates: Record<string, UpdateRegionRequest>;
}

export interface BulkDeleteRegionsRequest {
  regionIds: string[];
}

export class RegionsService {
  /**
   * Get all regions for a specific slide
   */
  async getRegionsForSlide(slideUid: string): Promise<Region[]> {
    const response = await apiClient.get<RegionList>(
      `/slides/${slideUid}/regions`
    );
    return response.data.regions || [];
  }

  /**
   * Get a specific region by ID
   */
  async getRegionById(regionId: string): Promise<Region> {
    const response = await apiClient.get<Region>(`/regions/${regionId}`);
    return response.data;
  }

  /**
   * Create a new region
   */
  async createRegion(
    slideUid: string,
    region: CreateRegionRequest
  ): Promise<Region> {
    const response = await apiClient.post<Region>(
      `/slides/${slideUid}/regions`,
      region
    );
    return response.data;
  }

  /**
   * Update an existing region
   */
  async updateRegion(
    regionId: string,
    updates: UpdateRegionRequest
  ): Promise<Region> {
    const response = await apiClient.put<Region>(
      `/regions/${regionId}`,
      updates
    );
    return response.data;
  }

  /**
   * Delete a region
   */
  async deleteRegion(slideUid: string, regionId: string): Promise<void> {
    await apiClient.delete(`/slides/${slideUid}/regions/${regionId}`);
  }

  /**
   * Bulk create multiple regions
   */
  async bulkCreateRegions(
    slideUid: string,
    request: BulkCreateRegionsRequest
  ): Promise<Region[]> {
    const response = await apiClient.post<Region[]>(
      `/slides/${slideUid}/regions/bulk`,
      request
    );
    return response.data;
  }

  /**
   * Bulk update multiple regions
   */
  async bulkUpdateRegions(request: BulkUpdateRegionsRequest): Promise<void> {
    await apiClient.put(`/regions/bulk`, request);
  }

  /**
   * Bulk delete multiple regions
   */
  async bulkDeleteRegions(request: BulkDeleteRegionsRequest): Promise<void> {
    await apiClient.deleteWithBody(`/regions/bulk`, request);
  }

  /**
   * Get region statistics for a slide
   */
  async getRegionStatistics(slideUid: string): Promise<any> {
    const response = await apiClient.get(
      `/slides/${slideUid}/regions/statistics`
    );
    return response.data;
  }

  /**
   * Get deleted regions
   */
  async getDeletedRegions(): Promise<Region[]> {
    const response = await apiClient.get<Region[]>(`/regions/deleted`);
    return response.data;
  }

  /**
   * Restore a deleted region
   */
  async restoreRegion(regionId: string): Promise<Region> {
    const response = await apiClient.post<Region>(
      `/regions/${regionId}/restore`
    );
    return response.data;
  }
}

export const regionsService = new RegionsService();
