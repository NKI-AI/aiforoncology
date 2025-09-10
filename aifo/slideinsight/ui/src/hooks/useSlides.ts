// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import {
  usePaginatedApi,
  type BasePaginatedOptions,
  type PaginationInfo,
} from "./usePaginatedApi";
import { apiFetch } from "../utils/fetchUtils";

interface Slide {
  slideUid: string;
  slideName: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
}

interface ApiResponse {
  slides: Slide[];
  pagination: PaginationInfo;
}

interface Mask {
  maskUid: string;
  maskName: string;
}

interface MaskList {
  slideUid: string;
  masks: Mask[];
}

interface UseSlidesOptions extends BasePaginatedOptions {
  withMaskCounts?: boolean;
  name?: string; // Filter by name
}

export interface SlideWithCount extends Slide {
  maskCount?: number;
}

export function useSlides(options: UseSlidesOptions = {}) {
  const { withMaskCounts = false, ...queryOptions } = options;

  const config = useMemo(
    () => ({
      endpoint: "/api/v1/slides",
      queryBuilder: (opts: Omit<UseSlidesOptions, "withMaskCounts">) => {
        const params: Record<string, string> = {};
        if (opts.q) params.q = opts.q;
        if (opts.name) params.name = opts.name;
        if (opts.sort) params.sort = opts.sort;
        if (opts.dir) params.dir = opts.dir;
        return params;
      },
      dataExtractor: (response: ApiResponse) => ({
        items: response.slides,
        pagination: response.pagination,
      }),
      postProcessor: withMaskCounts
        ? async (slides: Slide[]): Promise<SlideWithCount[]> => {
            return Promise.all(
              slides.map(async (slide) => {
                try {
                  const masks = await apiFetch<MaskList>(
                    `/api/v1/slides/${slide.slideUid}/annotations/raster`
                  );
                  return { ...slide, maskCount: masks.masks.length };
                } catch {
                  return { ...slide, maskCount: 0 };
                }
              })
            );
          }
        : undefined,
      errorMessage: "Failed to load slides. Please try again later.",
    }),
    [withMaskCounts]
  );

  const result = usePaginatedApi<
    SlideWithCount,
    ApiResponse,
    Omit<UseSlidesOptions, "withMaskCounts">
  >(queryOptions, config);

  return {
    slides: result.data,
    pagination: result.pagination,
    loading: result.loading,
    error: result.error,
    refetch: result.refetch,
  };
}
