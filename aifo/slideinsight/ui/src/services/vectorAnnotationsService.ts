// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { apiClient } from "./apiClient";

export interface VectorLabel {
  name: string;
  color: string;
}

export interface VectorAnnotation {
  vectorUid: string;
  vectorName: string;
  slideUid: string;
  fileUri?: string;
  dataBlob?: any; // Data blob (GeoJSON, etc.)
  labels?: VectorLabel[];
  actorType?: string;
  actorId?: number;
  deletedAt?: string;
  deletedBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface VectorAnnotationList {
  slideUid: string;
  annotations: VectorAnnotation[];
}

export interface WorkspaceAnnotationItem {
  id: string;
  label: string;
  name: string;
  visible: boolean;
  kind: string;
}

export interface AnnotationImportResult {
  slideUid: string;
  vectorUid: string;
  importedCount: number;
  skippedCount: number;
  overwrittenCount: number;
  geoJsonFeatures: any;
  importedAnnotations: WorkspaceAnnotationItem[];
  skippedLabels: string[];
  studyLabels: string[];
}

export interface VectorAnnotationsResponse {
  annotations: VectorAnnotation[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

export class VectorAnnotationsService {
  /**
   * Get all vector annotations for a specific slide
   */
  async getVectorAnnotationsForSlide(
    slideUid: string
  ): Promise<VectorAnnotation[]> {
    const response = await apiClient.get<VectorAnnotationList>(
      `/slides/${slideUid}/annotations/vector`
    );
    return response.data.annotations || [];
  }

  /**
   * Create a new vector annotation
   */
  async createVectorAnnotation(
    slideUid: string,
    annotation: Omit<VectorAnnotation, "vectorUid" | "createdAt" | "updatedAt">
  ): Promise<VectorAnnotation> {
    const response = await apiClient.post<VectorAnnotation>(
      `/slides/${slideUid}/annotations/vector`,
      {
        ...annotation,
        slideUid,
      }
    );
    return response.data;
  }

  /**
   * Update an existing vector annotation
   */
  async updateVectorAnnotation(
    slideUid: string,
    vectorUid: string,
    annotation: Partial<VectorAnnotation>
  ): Promise<VectorAnnotation> {
    const response = await apiClient.put<VectorAnnotation>(
      `/slides/${slideUid}/annotations/vector/${vectorUid}`,
      annotation
    );
    return response.data;
  }

  /**
   * Delete a vector annotation
   */
  async deleteVectorAnnotation(
    slideUid: string,
    vectorUid: string
  ): Promise<void> {
    await apiClient.delete(
      `/slides/${slideUid}/annotations/vector/${vectorUid}`
    );
  }

  /**
   * Get the GeoJSON file content for a vector annotation
   */
  async getVectorAnnotationFile(
    slideUid: string,
    vectorUid: string
  ): Promise<any> {
    const response = await apiClient.get(
      `/slides/${slideUid}/annotations/vector/${vectorUid}/file`
    );
    return response.data;
  }

  /**
   * Import a vector annotation as workspace annotations
   */
  async importVectorAnnotationToWorkspace(
    slideUid: string,
    vectorUid: string,
    importSettings: {
      allowedLabels: Array<{
        label: string;
        type: string;
        color: string;
      }>;
    }
  ): Promise<AnnotationImportResult> {
    const response = await apiClient.post<AnnotationImportResult>(
      `/slides/${slideUid}/annotations/vector/${vectorUid}/import`,
      importSettings
    );
    return response.data;
  }
}

export const vectorAnnotationsService = new VectorAnnotationsService();
