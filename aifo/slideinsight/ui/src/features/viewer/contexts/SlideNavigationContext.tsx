// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { createContext, useContext, useMemo, ReactNode } from "react";

export interface SlideNavigationItem {
  slideUid: string;
  slideName: string;
  caseUid?: string;
  caseName?: string;
  studyUid?: string;
  studyName?: string;
}

interface SlideNavigationContextType {
  currentSlideUid: string | null;
  slides: SlideNavigationItem[];
  currentIndex: number;
  hasNext: boolean;
  hasPrevious: boolean;
  nextSlide: SlideNavigationItem | null;
  previousSlide: SlideNavigationItem | null;
  getNextSlideUid: () => string | null;
  getPreviousSlideUid: () => string | null;
  totalSlides: number;
  sourceContext?: {
    type:
      | "study"
      | "case-list"
      | "slide-list"
      | "all-slides"
      | "filtered-subset";
    id?: string;
    name?: string;
    filters?: Record<string, any>;
  };
  // Subset navigation mode
  isSubsetMode?: boolean;
  subsetFilters?: {
    searchQuery?: string;
    searchName?: string;
    sortBy?: string;
    sortDir?: string;
  };
}

const SlideNavigationContext = createContext<SlideNavigationContextType | null>(
  null
);

interface SlideNavigationProviderProps {
  children: ReactNode;
  currentSlideUid: string | null;
  slides: SlideNavigationItem[];
  sourceContext?: SlideNavigationContextType["sourceContext"];
}

export function SlideNavigationProvider({
  children,
  currentSlideUid,
  slides,
  sourceContext,
}: SlideNavigationProviderProps) {
  const navigationState = useMemo(() => {
    const currentIndex = currentSlideUid
      ? slides.findIndex((slide) => slide.slideUid === currentSlideUid)
      : -1;

    const hasNext = currentIndex >= 0 && currentIndex < slides.length - 1;
    const hasPrevious = currentIndex > 0;

    const nextSlide = hasNext ? slides[currentIndex + 1] : null;
    const previousSlide = hasPrevious ? slides[currentIndex - 1] : null;

    return {
      currentSlideUid: currentSlideUid,
      slides,
      currentIndex,
      hasNext,
      hasPrevious,
      nextSlide,
      previousSlide,
      getNextSlideUid: () => nextSlide?.slideUid || null,
      getPreviousSlideUid: () => previousSlide?.slideUid || null,
      totalSlides: slides.length,
      sourceContext,
    };
  }, [currentSlideUid, slides, sourceContext]);

  return (
    <SlideNavigationContext.Provider value={navigationState}>
      {children}
    </SlideNavigationContext.Provider>
  );
}

export function useSlideNavigation(): SlideNavigationContextType {
  const context = useContext(SlideNavigationContext);
  if (!context) {
    // Return a default state when no context is provided
    return {
      currentSlideUid: null,
      slides: [],
      currentIndex: -1,
      hasNext: false,
      hasPrevious: false,
      nextSlide: null,
      previousSlide: null,
      getNextSlideUid: () => null,
      getPreviousSlideUid: () => null,
      totalSlides: 0,
    };
  }
  return context;
}
