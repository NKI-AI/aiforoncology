// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useCallback, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useSlides } from "../hooks/useSlides";
import { apiFetch } from "../utils/fetchUtils";
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  ChevronRightIcon,
} from "@heroicons/react/24/outline";

interface SlideData {
  id: string;
  name: string;
  dimensions: string;
  resolution: string;
}

export default function SlideSelector() {
  const { slides: rawSlides, loading, error } = useSlides();
  const [searchTerm, setSearchTerm] = useState("");
  const navigate = useNavigate();

  const slides: SlideData[] = useMemo(
    () =>
      rawSlides.map((slide) => ({
        id: slide.slideUid,
        name: slide.slideName || slide.slideUid,
        dimensions: `${slide.slideWidth} × ${slide.slideHeight}`,
        resolution: slide.slideMpp
          ? `${slide.slideMpp.toFixed(4)} µm/px`
          : "Unknown",
      })),
    [rawSlides]
  );

  // Filter slides based on search term
  const filteredSlides = useMemo(() => {
    if (!searchTerm.trim()) return slides;

    const term = searchTerm.toLowerCase();
    return slides.filter(
      (slide) =>
        slide.name.toLowerCase().includes(term) ||
        slide.id.toLowerCase().includes(term)
    );
  }, [slides, searchTerm]);

  const handleSlideSelect = (slideUid: string) => {
    navigate({ to: "/i/$slideUid", params: { slideUid } });
  };

  return (
    <div className="flex items-center justify-center min-h-full p-4">
      <div className="bg-card border border-border rounded-lg shadow-lg p-6 w-full max-w-xl">
        <h2 className="text-2xl font-bold text-muted-800 mb-6 text-center">
          Select a Slide
        </h2>

        {/* Search input */}
        <div className="mb-6">
          <div className="relative">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
              <MagnifyingGlassIcon className="w-5 h-5 text-muted-400" />
            </div>
            <input
              type="text"
              className="bg-gray-50 border border-gray-300 text-muted-900 text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 p-2.5"
              placeholder="Search slides..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              autoFocus
            />
            {searchTerm && (
              <button
                className="absolute inset-y-0 right-0 flex items-center pr-3"
                onClick={() => setSearchTerm("")}
              >
                <XMarkIcon className="w-5 h-5 text-muted-400 hover:text-muted-600" />
              </button>
            )}
          </div>
        </div>

        {/* Loading state */}
        {loading && (
          <div className="flex justify-center items-center p-8">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
          </div>
        )}

        {/* Error message */}
        {error && (
          <div className="p-4 mb-4 text-sm text-red-700 bg-red-100 rounded-lg">
            {error}
          </div>
        )}

        {/* Slides list */}
        {!loading && !error && (
          <div className="overflow-y-auto max-h-80">
            {filteredSlides.length === 0 ? (
              <div className="text-center p-4 text-muted-500">
                No slides matching "{searchTerm}" found
              </div>
            ) : (
              <ul className="divide-y divide-gray-200">
                {filteredSlides.map((slide) => (
                  <li key={slide.id} className="py-3">
                    <button
                      onClick={() => handleSlideSelect(slide.id)}
                      className="w-full text-left hover:bg-gray-50 p-2 rounded transition-colors flex items-center justify-between"
                    >
                      <div>
                        <div className="font-medium text-indigo-600">
                          {slide.name}
                        </div>
                        <div className="text-sm text-muted-500 flex flex-col sm:flex-row sm:gap-3">
                          <span>{slide.dimensions}</span>
                          <span className="hidden sm:inline">•</span>
                          <span>{slide.resolution}</span>
                        </div>
                      </div>
                      <ChevronRightIcon className="h-5 w-5 text-muted-400" />
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {/* Footer with navigation links */}
        <div className="mt-6 pt-4 border-t border-gray-200 text-center">
          <div className="flex justify-center gap-4">
            <button
              onClick={() => navigate({ to: "/admin/cases" })}
              className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
            >
              Browse Cases
            </button>
            <span className="text-muted-400">•</span>
            <button
              onClick={() => navigate({ to: "/admin/slides" })}
              className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
            >
              View All Slides
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
