// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useState, useEffect, useCallback } from "react";
import { apiFetch } from "../utils/fetchUtils";

interface Slide {
  slideId: string;
  slideName: string;
  slideUri: string;
  slideWidth: number;
  slideHeight: number;
  slideMpp: number;
}

interface ApiResponse {
  slides: Slide[];
}

interface Mask {
  maskId: string;
  maskName: string;
  maskUri: string;
}

interface MaskList {
  slide_id: string;
  masks: Mask[];
}

export interface UseSlidesOptions {
  withMaskCounts?: boolean;
}

export interface SlideWithCount extends Slide {
  maskCount?: number;
}

export function useSlides(options: UseSlidesOptions = {}) {
  const { withMaskCounts = false } = options;
  const [slides, setSlides] = useState<SlideWithCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSlides = useCallback(async () => {
    try {
      setLoading(true);
      const data = await apiFetch<ApiResponse>("/api/v1/slides");

      let slidesList: SlideWithCount[] = data.slides;

      if (withMaskCounts) {
        slidesList = await Promise.all(
          data.slides.map(async (slide) => {
            try {
              const masks = await apiFetch<MaskList>(
                `/api/v1/slides/${slide.slideId}/annotations/raster`
              );
              return { ...slide, maskCount: masks.masks.length };
            } catch {
              return { ...slide, maskCount: 0 };
            }
          })
        );
      }

      setSlides(slidesList);
    } catch (err) {
      console.error("Failed to fetch slides:", err);
      setError("Failed to load slides. Please try again later.");
    } finally {
      setLoading(false);
    }
  }, [withMaskCounts]);

  useEffect(() => {
    fetchSlides();
  }, [fetchSlides]);

  return { slides, loading, error, refetch: fetchSlides };
}

export type { Slide };
