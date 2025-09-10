// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useApiQuery } from "@/utils/apiQueries";
import { queryKeys } from "@/api/queryKeys";
import { useMemo } from "react";

export interface StudyAnnotationLabel {
  label: string;
  color: string;
  type: "point" | "box" | "polygon";
}

export interface StudyPluginSettings {
  viewerSettings: boolean; // Always enabled for basic viewer controls
  brightnessControl: boolean; // Brightness/contrast adjustments
  annotationControl: boolean; // Manual annotation tools
  maskControl: boolean; // Mask visualization controls
  regionControl: boolean; // Region annotation tools for multi-patient samples
}

export interface StudyAnnotationSettings {
  allowAnnotation: boolean;
  annotations: StudyAnnotationLabel[];
  colorMap: Record<string, string>;
  indexMap: Record<string, any>;
  pluginSettings: StudyPluginSettings;
}

export interface StudyMetadata {
  allow_annotation: boolean;
  annotations: StudyAnnotationLabel[];
  color_map: Record<string, string>;
  index_map: Record<string, any>;
  plugin_settings?: {
    viewer_settings?: boolean;
    brightness_control?: boolean;
    annotation_control?: boolean;
    mask_control?: boolean;
    region_control?: boolean;
  };
  study_info: {
    organs: string[]; // Renamed from categories to organs
    diseases: string[];
    stainings: string[];
  };
  // Support nested metadata structures
  metadata?: StudyMetadata;
}

export interface StudyResponse {
  tenantUid: string;
  studyUid: string;
  creatorUid: string;
  name: string;
  description: string;
  metadata: StudyMetadata;
  isPublished: boolean;
  caseCount: number;
  slideCount: number;
  createdAt: string;
}

/**
 * Hook to fetch study annotation settings from the API
 */
export function useStudyAnnotationSettings(studyUid: string | undefined) {
  const queryKey = useMemo(
    () => (studyUid ? queryKeys.studies.detail(studyUid) : null),
    [studyUid]
  );

  const url = useMemo(
    () => (studyUid ? `/api/v1/studies/${studyUid}` : null),
    [studyUid]
  );

  const queryResult = useApiQuery<StudyResponse>(
    queryKey || ["studies", "annotation-settings", "disabled"],
    url || "",
    {
      enabled: !!studyUid,
      staleTime: 1000 * 60 * 5, // 5 minutes
    }
  );

  // Transform the API response to a more convenient format
  const annotationSettings = useMemo((): StudyAnnotationSettings | null => {
    if (!queryResult.data?.metadata) return null;

    const metadata = queryResult.data.metadata;

    // Handle both direct metadata and nested metadata structures
    const actualMetadata = metadata.metadata || metadata;

    // Default plugin settings - all can be toggled based on study config
    const defaultPluginSettings: StudyPluginSettings = {
      viewerSettings: true, // Default enabled but can be toggled
      brightnessControl: true, // Default enabled
      annotationControl: actualMetadata.allow_annotation || false, // Based on annotation availability
      maskControl: true, // Default enabled
      regionControl: false, // Default disabled, needs explicit enabling
    };

    // Override with actual plugin settings from metadata if available
    const pluginSettings: StudyPluginSettings = {
      ...defaultPluginSettings,
      ...(actualMetadata.plugin_settings
        ? {
            viewerSettings:
              actualMetadata.plugin_settings.viewer_settings ??
              defaultPluginSettings.viewerSettings,
            brightnessControl:
              actualMetadata.plugin_settings.brightness_control ??
              defaultPluginSettings.brightnessControl,
            annotationControl:
              actualMetadata.plugin_settings.annotation_control ??
              defaultPluginSettings.annotationControl,
            maskControl:
              actualMetadata.plugin_settings.mask_control ??
              defaultPluginSettings.maskControl,
            regionControl:
              actualMetadata.plugin_settings.region_control ??
              defaultPluginSettings.regionControl,
          }
        : {}),
    };

    return {
      allowAnnotation: actualMetadata.allow_annotation || false,
      annotations: Array.isArray(actualMetadata.annotations)
        ? actualMetadata.annotations
        : [],
      colorMap: actualMetadata.color_map || {},
      indexMap: actualMetadata.index_map || {},
      pluginSettings,
    };
  }, [queryResult.data]);

  return {
    data: annotationSettings,
    isLoading: queryResult.isLoading,
    error: queryResult.error,
    refetch: queryResult.refetch,
    studyData: queryResult.data,
  };
}
