// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect, useCallback, useMemo } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../auth";
import { useAuthApi } from "../utils/authApiClient";
import { apiFetch } from "../utils/fetchUtils";
import { useSlides } from "../hooks/useSlides";
import Navbar from "./Navbar";
import Footer from "./Footer";
import {
  ArrowLeftIcon,
  ErrorIcon,
  ImageIcon,
  DimensionsIcon,
  ResolutionIcon,
  SecurityIcon,
  CircleIcon,
  EmptyStateIcon,
} from "./icons";

interface SlideEntry {
  slide_id: string;
  slide_name: string;
  dimensions: string;
  mpp: number;
  maskCount: number;
  [key: string]: any;
}

interface SlideListProps {}

const SlideList: React.FC<SlideListProps> = () => {
  const {
    slides: rawSlides,
    loading,
    error,
  } = useSlides({ withMaskCounts: true });
  const { logout } = useAuth();
  const api = useAuthApi();
  const navigate = useNavigate();
  const [hoveredSlide, setHoveredSlide] = useState<string | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });

  const slides = useMemo(() => {
    return rawSlides.map((slide) => ({
      slide_id: slide.slideId,
      slide_name: slide.slideName,
      dimensions: `${slide.slideWidth} × ${slide.slideHeight}`,
      mpp: slide.slideMpp,
      maskCount: slide.maskCount || 0,
    }));
  }, [rawSlides]);

  // Handle mouse enter for tooltip
  const handleMouseEnter = (slideId: string, event: React.MouseEvent) => {
    const rect = event.currentTarget.getBoundingClientRect();
    setTooltipPosition({
      x: rect.left + rect.width / 2,
      y: rect.top - 10,
    });
    setHoveredSlide(slideId);
  };

  // Handle mouse leave for tooltip
  const handleMouseLeave = () => {
    setHoveredSlide(null);
  };

  // Handle logout
  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  // Set document title
  useEffect(() => {
    document.title = "SlideScope - Slide Selection";
    return () => {
      document.title = "SlideScope Viewer";
    };
  }, []);

  return (
    <div className="bg-gray-100 text-gray-800 font-sans min-h-screen flex flex-col">
      {/* Use the shared Navbar component */}
      <Navbar
        onToggleMaskControl={() => {}}
        onToggleSlideInfo={() => {}}
        onToggleHelp={() => {}}
        onToggleCrosshair={() => {}}
        showCrosshair={false}
      />

      {/* Main content container */}
      <div className="container mx-auto px-4 py-8 mt-16 max-w-7xl flex-grow">
        {/* Header section with title and count */}
        <div className="mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between bg-white rounded-xl shadow-sm p-6 border border-indigo-100">
          <div>
            <h1 className="text-2xl font-bold text-indigo-800 mb-1">
              Available Slides
            </h1>
            <p className="text-indigo-500">
              {slides.length
                ? `${slides.length} slide${
                    slides.length !== 1 ? "s" : ""
                  } available for viewing`
                : "No slides available"}
            </p>
          </div>

          <div className="mt-4 sm:mt-0">
            <Link
              to="/"
              className="inline-flex items-center text-indigo-600 hover:text-indigo-800 transition-colors"
            >
              <ArrowLeftIcon />
              Return to Home
            </Link>
          </div>
        </div>

        {/* Loading state */}
        {loading && (
          <div className="flex justify-center items-center p-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
          </div>
        )}

        {/* Error state */}
        {error && (
          <div className="bg-red-50 border-l-4 border-red-500 p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <ErrorIcon />
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Slides table - Desktop */}
        {!loading && !error && slides.length > 0 && (
          <>
            {/* Desktop Table View */}
            <div className="hidden md:block bg-white rounded-xl shadow-md overflow-hidden border border-indigo-50">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gradient-to-r from-indigo-600 to-indigo-500 text-white">
                    <tr>
                      <th className="px-6 py-4 text-left text-sm font-semibold uppercase tracking-wider">
                        Slide Name
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold uppercase tracking-wider">
                        Annotations
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {slides.map((slide, index) => (
                      <tr
                        key={slide.slide_id}
                        className="hover:bg-gray-50 transition-colors"
                        style={{ "--index": index } as React.CSSProperties}
                      >
                        <td className="px-6 py-4">
                          <div
                            className="flex items-center cursor-help"
                            onMouseEnter={(e) =>
                              handleMouseEnter(slide.slide_id, e)
                            }
                            onMouseLeave={handleMouseLeave}
                          >
                            <ImageIcon />
                            <div className="ml-3">
                              <div className="text-sm font-medium text-gray-900">
                                {slide.slide_name}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center">
                            <SecurityIcon />
                            <div className="ml-2">
                              {slide.maskCount > 0 ? (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                  <CircleIcon className="-ml-0.5 mr-1.5 h-2 w-2 text-green-500" />
                                  {slide.maskCount}{" "}
                                  {slide.maskCount === 1 ? "Mask" : "Masks"}
                                </span>
                              ) : (
                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                  <CircleIcon className="-ml-0.5 mr-1.5 h-2 w-2 text-gray-500" />
                                  No Masks
                                </span>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <Link
                            to={`/v/${slide.slide_id}`}
                            className="inline-flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition shadow-sm hover:shadow focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50"
                          >
                            View Slide
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Mobile Card View */}
            <div className="md:hidden space-y-4">
              {slides.map((slide, index) => (
                <div
                  key={slide.slide_id}
                  className="bg-white rounded-xl shadow-md overflow-hidden border border-indigo-50"
                  style={{ "--index": index } as React.CSSProperties}
                >
                  <div className="bg-gradient-to-r from-indigo-600 to-indigo-500 p-4 text-white">
                    <div className="flex items-center">
                      <ImageIcon />
                      <h3 className="ml-2 text-lg font-semibold truncate">
                        {slide.slide_name}
                      </h3>
                    </div>
                  </div>

                  <div className="p-4 space-y-3">
                    {/* Slide Details */}
                    <div className="space-y-2 text-sm">
                      <div className="flex items-center text-gray-600">
                        <span className="font-medium w-20">ID:</span>
                        <span className="text-gray-800">{slide.slide_id}</span>
                      </div>
                      <div className="flex items-center text-gray-600">
                        <DimensionsIcon className="w-4 h-4 mr-1" />
                        <span className="font-medium w-20">Size:</span>
                        <span className="text-gray-800">
                          {slide.dimensions}
                        </span>
                      </div>
                      <div className="flex items-center text-gray-600">
                        <ResolutionIcon className="w-4 h-4 mr-1" />
                        <span className="font-medium w-20">MPP:</span>
                        <span className="text-gray-800">
                          {slide.mpp?.toFixed(4)} µm/px
                        </span>
                      </div>
                    </div>

                    {/* Annotations */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <SecurityIcon />
                        <span className="ml-2 text-sm font-medium text-gray-700">
                          Annotations:
                        </span>
                      </div>
                      <div>
                        {slide.maskCount > 0 ? (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            <CircleIcon className="-ml-0.5 mr-1.5 h-2 w-2 text-green-500" />
                            {slide.maskCount}{" "}
                            {slide.maskCount === 1 ? "Mask" : "Masks"}
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                            <CircleIcon className="-ml-0.5 mr-1.5 h-2 w-2 text-gray-500" />
                            No Masks
                          </span>
                        )}
                      </div>
                    </div>

                    {/* Action Button */}
                    <Link
                      to={`/v/${slide.slide_id}`}
                      className="block w-full py-3 px-4 bg-indigo-600 hover:bg-indigo-700 text-white text-center font-medium rounded-lg transition shadow-sm hover:shadow focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50"
                    >
                      View Slide
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}

        {/* Fixed position tooltip - Desktop only */}
        {hoveredSlide && (
          <div
            className="hidden md:block fixed z-50 bg-gray-900 text-white text-xs rounded-lg p-3 shadow-xl pointer-events-none transform -translate-x-1/2 -translate-y-full"
            style={{
              left: tooltipPosition.x,
              top: tooltipPosition.y,
            }}
          >
            {(() => {
              const slide = slides.find((s) => s.slide_id === hoveredSlide);
              return slide ? (
                <div className="space-y-2 min-w-max">
                  <div className="flex items-center">
                    <span className="font-medium text-gray-300 w-16">ID:</span>
                    <span>{slide.slide_id}</span>
                  </div>
                  <div className="flex items-center">
                    <DimensionsIcon className="w-3 h-3 mr-1 text-gray-400" />
                    <span className="font-medium text-gray-300 w-16">
                      Size:
                    </span>
                    <span>{slide.dimensions}</span>
                  </div>
                  <div className="flex items-center">
                    <ResolutionIcon className="w-3 h-3 mr-1 text-gray-400" />
                    <span className="font-medium text-gray-300 w-16">MPP:</span>
                    <span>{slide.mpp?.toFixed(4)} µm/px</span>
                  </div>
                </div>
              ) : null;
            })()}
            {/* Arrow pointing down */}
            <div className="absolute top-full left-1/2 transform -translate-x-1/2 w-2 h-2 bg-gray-900 rotate-45"></div>
          </div>
        )}

        {/* Empty state */}
        {!loading && !error && slides.length === 0 && (
          <div className="bg-white rounded-xl shadow-md p-10 text-center">
            <div className="max-w-md mx-auto">
              <div className="bg-indigo-50 rounded-full w-20 h-20 flex items-center justify-center mx-auto mb-4">
                <EmptyStateIcon />
              </div>
              <h3 className="text-xl font-medium text-gray-700 mb-2">
                No slides available
              </h3>
              <p className="text-gray-500 mb-6">
                Please check your manifest file configuration or upload some
                slides to get started.
              </p>
              <Link
                to="/"
                className="inline-flex items-center justify-center px-5 py-2 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Return to Home
              </Link>
            </div>
          </div>
        )}
      </div>

      {/* Footer */}
      <Footer showDescription={false} />

      {/* Add animation for rows using regular style tag */}
      <style>
        {`
        @keyframes fadeIn {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        tbody tr {
          animation: fadeIn 0.3s ease-out forwards;
          animation-delay: calc(0.05s * var(--index, 0));
        }
        `}
      </style>
    </div>
  );
};

export default SlideList;
