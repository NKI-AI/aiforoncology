// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useSlides } from "../hooks/useSlides";
import { apiFetch } from "../utils/fetchUtils";

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
        id: slide.slideId,
        name: slide.slideName || slide.slideId,
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

  const handleSlideSelect = (slideId: string) => {
    navigate(`/v/${slideId}`);
  };

  return (
    <div className="flex items-center justify-center min-h-full p-4">
      <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-xl">
        <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center">
          Select a Slide
        </h2>

        {/* Search input */}
        <div className="mb-6">
          <div className="relative">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
              <svg
                className="w-5 h-5 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
            </div>
            <input
              type="text"
              className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-indigo-500 focus:border-indigo-500 block w-full pl-10 p-2.5"
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
                <svg
                  className="w-5 h-5 text-gray-400 hover:text-gray-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
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
              <div className="text-center p-4 text-gray-500">
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
                        <div className="text-sm text-gray-500 flex flex-col sm:flex-row sm:gap-3">
                          <span>{slide.dimensions}</span>
                          <span className="hidden sm:inline">•</span>
                          <span>{slide.resolution}</span>
                        </div>
                      </div>
                      <svg
                        className="h-5 w-5 text-gray-400"
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                      >
                        <path
                          fillRule="evenodd"
                          d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                          clipRule="evenodd"
                        />
                      </svg>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {/* Footer with link back to gallery */}
        <div className="mt-6 pt-4 border-t border-gray-200 text-center">
          <button
            onClick={() => navigate("/slides")}
            className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
          >
            View full slide gallery
          </button>
        </div>
      </div>
    </div>
  );
}
