// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
  useMemo,
} from "react";
import { useNavigate, useRouter } from "@tanstack/react-router";
import { useSlideNavigation } from "@/features/viewer/contexts/SlideNavigationContext";
import { apiFetch } from "@/utils/fetchUtils";
import {
  useCases,
  useCaseNeighbors,
  useCasesByStudy,
  type CaseNeighborsOptions,
} from "@/hooks/useCases";
import Map from "ol/Map";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
} from "@/components/AlertDialog";

interface CaseSlideBarProps {
  className?: string;
  mapRef?: Map | null;
  filterParams?: {
    searchQuery?: string;
    searchName?: string;
    sortBy?: string;
    sortDir?: string;
    page?: number;
    limit?: number;
    isSubsetMode?: boolean;
    hasVectorAnnotations?: string;
    hasRasterAnnotations?: string;
    status?: string;
  };
}

interface CaseSlide {
  slideUid: string;
  slideName: string;
  slideWidth?: number;
  slideHeight?: number;
  slideMpp?: number;
}

const SNAP_HEIGHTS = {
  collapsed: 0,
  expanded: 60,
  mobile: 48, // Smaller height for mobile devices
} as const;

export default function CaseSlideBar({
  className = "",
  mapRef,
  filterParams,
}: CaseSlideBarProps) {
  const navigate = useNavigate();
  const router = useRouter();
  const {
    currentSlideUid,
    slides,
    currentIndex,
    totalSlides,
    hasNext,
    hasPrevious,
    getNextSlideUid,
    getPreviousSlideUid,
  } = useSlideNavigation();
  const [caseSlides, setCaseSlides] = useState<CaseSlide[]>([]);
  const [caseName, setCaseName] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [showCasesListModal, setShowCasesListModal] = useState(false);
  const [showServerError, setShowServerError] = useState(false);
  const [serverErrorMessage, setServerErrorMessage] = useState("");

  // Track previous case ID to prevent unnecessary fetching
  const [lastFetchedCaseUid, setLastFetchedCaseUid] = useState<string | null>(
    null
  );

  // Modern loading UX: Add delay to prevent flash for fast requests
  const [showLoadingState, setShowLoadingState] = useState(false);
  const loadingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Simple collapse state
  const [isCollapsed, setIsCollapsed] = useState(false);
  const panelGroupRef = useRef<HTMLDivElement | null>(null);

  // Filter slides to only show those in the current case
  const currentSlide = currentIndex >= 0 ? slides[currentIndex] : null;
  const resolvedCaseUid = currentSlide?.caseUid;

  // Extract studyUid from current route
  const currentPath = router.state.location.pathname;
  const studyUidMatch = currentPath.match(/\/studies\/([^\/]+)/);
  const studyUid = studyUidMatch ? studyUidMatch[1] : undefined;

  // Fetch case neighbors for navigation
  const neighborsOptions: CaseNeighborsOptions = {
    q: filterParams?.searchQuery || "",
    name: filterParams?.searchName || "",
    sort: filterParams?.sortBy || "",
    dir: filterParams?.sortDir || "",
    status: filterParams?.status || "",
    has_vector_annotations: filterParams?.hasVectorAnnotations || "",
    has_raster_annotations: filterParams?.hasRasterAnnotations || "",
  };
  const { neighbors, loading: neighborsLoading } = useCaseNeighbors(
    studyUid,
    resolvedCaseUid,
    neighborsOptions
  );

  // Handle clearing filters by navigating to the same route without search parameters
  const handleClearFilters = useCallback(() => {
    if (studyUid && resolvedCaseUid && currentSlideUid) {
      navigate({
        to: "/studies/$studyUid/v/$caseUid/i/$slideUid",
        params: {
          studyUid: studyUid,
          caseUid: resolvedCaseUid,
          slideUid: currentSlideUid,
        },
        search: {}, // Clear all search parameters
      });
    } else if (resolvedCaseUid && currentSlideUid) {
      navigate({
        to: "/v/$caseUid/i/$slideUid",
        params: {
          caseUid: resolvedCaseUid,
          slideUid: currentSlideUid,
        },
        search: {}, // Clear all search parameters
      });
    }
  }, [navigate, studyUid, resolvedCaseUid, currentSlideUid]);

  const handleCloseServerError = useCallback(() => {
    setShowServerError(false);
    setServerErrorMessage("");
  }, []);

  const handleRetryFetchCaseSlides = useCallback(() => {
    setShowServerError(false);
    setServerErrorMessage("");
    // Reset lastFetchedCaseUid to force refetch
    setLastFetchedCaseUid(null);
  }, []);

  // Check if there are active filters
  const hasActiveFilters = useMemo(() => {
    return Boolean(
      filterParams?.searchQuery ||
        filterParams?.searchName ||
        filterParams?.status ||
        filterParams?.hasVectorAnnotations ||
        filterParams?.hasRasterAnnotations ||
        filterParams?.sortBy
    );
  }, [filterParams]);

  // Load saved collapse state from localStorage
  useEffect(() => {
    const savedCollapsed = localStorage.getItem("caseSlideBarCollapsed");

    if (savedCollapsed) {
      setIsCollapsed(savedCollapsed === "true");
    }
  }, []);

  // Save collapse state to localStorage
  useEffect(() => {
    localStorage.setItem("caseSlideBarCollapsed", isCollapsed.toString());
  }, [isCollapsed]);

  // Fetch slides directly from case API when we have a NEW case ID
  useEffect(() => {
    if (!resolvedCaseUid) {
      setCaseSlides([]);
      setCaseName("");
      setLastFetchedCaseUid(null);
      return;
    }

    // Only fetch if this is a different case than we last fetched
    if (resolvedCaseUid === lastFetchedCaseUid) {
      return;
    }

    const fetchCaseSlides = async () => {
      try {
        setLoading(true);

        // Google-style loading delay: Only show loading UI after 300ms
        // This prevents flash for fast requests (< 300ms)
        loadingTimeoutRef.current = setTimeout(() => {
          setShowLoadingState(true);
        }, 300);

        const caseInfo = (await apiFetch(
          `/api/v1/cases/${resolvedCaseUid}`
        )) as any;
        const caseSlidesResponse = (await apiFetch(
          `/api/v1/cases/${resolvedCaseUid}/slides`
        )) as any;
        const slides = caseSlidesResponse.slides || caseSlidesResponse || [];
        setCaseSlides(slides);
        setCaseName(caseInfo.name || "Unknown Case");
        setLastFetchedCaseUid(resolvedCaseUid);
      } catch (error) {
        console.error("❌ Failed to fetch case slides:", error);

        // Check if this is a network/connectivity error
        if (error instanceof TypeError && error.message.includes("fetch")) {
          setShowServerError(true);
          setServerErrorMessage(
            "Unable to connect to the server. Please check your network connection."
          );
        } else {
          // Try fallback to navigation context if API fails for other reasons
          const navigationCaseSlides = slides.filter(
            (slide) => slide.caseUid === resolvedCaseUid
          );

          if (navigationCaseSlides.length > 0) {
            setCaseSlides(
              navigationCaseSlides.map((slide) => ({
                slideUid: slide.slideUid,
                slideName: slide.slideName,
              }))
            );
            setCaseName("Unknown Case");
            setLastFetchedCaseUid(resolvedCaseUid);
          } else {
            // No fallback available, show server error
            setShowServerError(true);
            setServerErrorMessage(
              "Failed to load case information. Please try again later."
            );
          }
        }
      } finally {
        // Clear the loading timeout and reset states
        if (loadingTimeoutRef.current) {
          clearTimeout(loadingTimeoutRef.current);
          loadingTimeoutRef.current = null;
        }
        setLoading(false);
        setShowLoadingState(false);
      }
    };

    fetchCaseSlides();
  }, [resolvedCaseUid, lastFetchedCaseUid, filterParams]);

  const handleSlideClick = useCallback(
    (slideUid: string) => {
      if (resolvedCaseUid) {
        navigate({
          to: "/v/$caseUid/i/$slideUid",
          params: {
            caseUid: resolvedCaseUid,
            slideUid: slideUid,
          },
        });
      }
    },
    [navigate, resolvedCaseUid]
  );

  const handleCaseSelect = useCallback(
    async (caseUid: string) => {
      try {
        // Fetch the first slide of the selected case
        const caseData = (await apiFetch(
          `/api/v1/cases/${caseUid}/slides`
        )) as any;
        const slides = caseData.slides || [];

        if (slides.length > 0) {
          const firstSlide = slides[0];

          const slideUid = firstSlide.slideUid;

          if (slideUid) {
            if (studyUid) {
              // Navigate to study-specific route
              navigate({
                to: "/studies/$studyUid/v/$caseUid/i/$slideUid",
                params: {
                  studyUid: studyUid,
                  caseUid: caseUid,
                  slideUid: slideUid,
                },
              });
            } else {
              // Navigate to study-agnostic route
              navigate({
                to: "/v/$caseUid/i/$slideUid",
                params: {
                  caseUid: caseUid,
                  slideUid: slideUid,
                },
              });
            }
          } else {
            console.error("❌ No valid slide ID found in slide:", firstSlide);
            console.error("Available properties:", Object.keys(firstSlide));
          }
        } else {
          console.warn("❌ No slides found in case:", caseUid);
        }
      } catch (error) {
        console.error("❌ Failed to fetch case slides:", error);
        console.error("Case caseUid used:", caseUid);

        // Show server error dialog for navigation failures too
        if (error instanceof TypeError && error.message.includes("fetch")) {
          setShowServerError(true);
          setServerErrorMessage(
            "Unable to connect to the server. Please check your network connection."
          );
        } else {
          setShowServerError(true);
          setServerErrorMessage(
            "Failed to load case information. Please try again later."
          );
        }
      }
    },
    [navigate, studyUid]
  );

  // Handle case navigation
  const handlePreviousCase = useCallback(async () => {
    if (!neighbors?.prev || !studyUid) return;

    const prevCaseUid = neighbors.prev.caseUid;
    await handleCaseSelect(prevCaseUid);
  }, [neighbors, studyUid, handleCaseSelect]);

  const handleNextCase = useCallback(async () => {
    if (!neighbors?.next || !studyUid) return;

    const nextCaseUid = neighbors.next.caseUid;
    await handleCaseSelect(nextCaseUid);
  }, [neighbors, studyUid, handleCaseSelect]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is typing in input fields
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        e.target instanceof HTMLSelectElement
      ) {
        return;
      }

      // Case navigation with Shift + Arrow keys
      if (e.shiftKey) {
        switch (e.key) {
          case "ArrowLeft":
          case "ArrowUp":
            e.preventDefault();
            handlePreviousCase();
            break;
          case "ArrowRight":
          case "ArrowDown":
            e.preventDefault();
            handleNextCase();
            break;
        }
        return;
      }

      // Slide navigation with Arrow keys
      switch (e.key) {
        case "ArrowLeft":
        case "ArrowUp":
          e.preventDefault();
          if (hasPrevious) {
            const prevSlideUid = getPreviousSlideUid();
            if (prevSlideUid && resolvedCaseUid) {
              if (studyUid) {
                navigate({
                  to: "/studies/$studyUid/v/$caseUid/i/$slideUid",
                  params: {
                    studyUid: studyUid,
                    caseUid: resolvedCaseUid,
                    slideUid: prevSlideUid,
                  },
                });
              } else {
                navigate({
                  to: "/v/$caseUid/i/$slideUid",
                  params: {
                    caseUid: resolvedCaseUid,
                    slideUid: prevSlideUid,
                  },
                });
              }
            }
          }
          break;
        case "ArrowRight":
        case "ArrowDown":
          e.preventDefault();
          if (hasNext) {
            const nextSlideUid = getNextSlideUid();
            if (nextSlideUid && resolvedCaseUid) {
              if (studyUid) {
                navigate({
                  to: "/studies/$studyUid/v/$caseUid/i/$slideUid",
                  params: {
                    studyUid: studyUid,
                    caseUid: resolvedCaseUid,
                    slideUid: nextSlideUid,
                  },
                });
              } else {
                navigate({
                  to: "/v/$caseUid/i/$slideUid",
                  params: {
                    caseUid: resolvedCaseUid,
                    slideUid: nextSlideUid,
                  },
                });
              }
            }
          }
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    hasNext,
    hasPrevious,
    getNextSlideUid,
    getPreviousSlideUid,
    navigate,
    resolvedCaseUid,
    studyUid,
    handlePreviousCase,
    handleNextCase,
  ]);

  // Cleanup loading timeout on unmount
  useEffect(() => {
    return () => {
      if (loadingTimeoutRef.current) {
        clearTimeout(loadingTimeoutRef.current);
      }
    };
  }, []);

  // When layout changes (resize), refresh OpenLayers map
  const refreshMapSize = useCallback(() => {
    if (!mapRef) return;
    // Defer a tick to ensure DOM has settled
    setTimeout(() => {
      mapRef.updateSize();
    }, 50);
  }, [mapRef]);

  const actualHeight = isCollapsed
    ? SNAP_HEIGHTS.collapsed
    : SNAP_HEIGHTS.expanded;

  // Update OpenLayers map size when bar collapses/expands
  useEffect(() => {
    if (mapRef) {
      setTimeout(() => {
        mapRef.updateSize();
      }, 100);
    }
  }, [isCollapsed, mapRef]);

  // Determine what to render - always show the bar, but with different content
  const hasValidData =
    resolvedCaseUid &&
    (caseSlides.length > 0 || lastFetchedCaseUid === resolvedCaseUid);
  const isLoadingNewCase =
    showLoadingState &&
    resolvedCaseUid &&
    resolvedCaseUid !== lastFetchedCaseUid;

  return (
    <div
      className={`relative w-full ${className}`}
      style={{ height: isCollapsed ? 0 : SNAP_HEIGHTS.expanded }}
    >
      {/* Minimal expand control when collapsed */}
      {isCollapsed && (
        <button
          onClick={() => setIsCollapsed(false)}
          className="fixed bottom-2 left-1/2 -translate-x-1/2 h-7 w-7 rounded-full border border-border bg-background hover:bg-muted text-muted-foreground shadow-sm flex items-center justify-center z-50"
          title="Expand"
        >
          <svg
            className="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M5 15l7-7 7 7"
            />
          </svg>
        </button>
      )}

      {!isCollapsed && (
        <ResizablePanelGroup
          direction="horizontal"
          className="border-t border-border bg-background/80 backdrop-blur supports-[backdrop-filter]:bg-background/70 shadow-lg rounded-none w-full h-full"
          onLayout={refreshMapSize}
        >
          <ResizablePanel
            defaultSize={25}
            minSize={20}
            maxSize={60}
            onResize={refreshMapSize}
          >
            <div className="h-full flex items-center px-3 py-2">
              {/* Cases List Button */}
              <Button
                variant="secondary"
                size="icon"
                className="h-7 w-7 shrink-0"
                onClick={() => setShowCasesListModal(true)}
                disabled={Boolean(isLoadingNewCase)}
                title="Browse all cases"
              >
                <svg
                  className="w-3.5 h-3.5 text-foreground/80"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                </svg>
              </Button>

              {/* Case icon */}
              <div className="mx-2 h-7 w-7 rounded bg-primary/10 text-primary flex items-center justify-center shrink-0">
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                  />
                </svg>
              </div>

              {/* Case info */}
              <div className="flex-1 min-w-0">
                {isLoadingNewCase ? (
                  <div className="animate-pulse">
                    <div className="h-4 bg-muted rounded w-32 mb-1" />
                    <div className="h-3 bg-muted rounded w-20" />
                  </div>
                ) : hasValidData ? (
                  <div className="transition-opacity duration-300 ease-out opacity-100">
                    <div className="text-sm font-semibold truncate text-foreground">
                      {caseName || "Loading..."}
                    </div>
                    <div className="text-xs text-muted-foreground flex items-center">
                      {caseSlides.length} image
                      {caseSlides.length !== 1 ? "s" : ""}
                      {neighbors && (
                        <div className="ml-2 flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={handlePreviousCase}
                            disabled={Boolean(
                              !neighbors?.prev ||
                                neighborsLoading ||
                                isLoadingNewCase
                            )}
                            title={
                              neighbors?.prev
                                ? `Previous case: ${neighbors.prev.name}`
                                : "No previous case"
                            }
                          >
                            ◀
                          </Button>
                          <span className="text-primary font-medium">
                            {neighbors.number}/{neighbors.count}
                          </span>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={handleNextCase}
                            disabled={Boolean(
                              !neighbors?.next ||
                                neighborsLoading ||
                                isLoadingNewCase
                            )}
                            title={
                              neighbors?.next
                                ? `Next case: ${neighbors.next.name}`
                                : "No next case"
                            }
                          >
                            ▶
                          </Button>
                        </div>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="animate-pulse">
                    <div className="h-4 bg-muted rounded w-24 mb-1" />
                    <div className="h-3 bg-muted rounded w-16" />
                  </div>
                )}
              </div>

              {/* Clear Filters */}
              {hasActiveFilters && (
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleClearFilters}
                  disabled={Boolean(isLoadingNewCase)}
                  className="ml-2 shrink-0"
                >
                  Clear Filters
                </Button>
              )}

              {/* Collapse Button */}
              <Button
                variant="ghost"
                size="icon"
                className="ml-2 h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={() => setIsCollapsed(true)}
                title="Collapse"
              >
                <svg
                  className="h-4 w-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </Button>
            </div>
          </ResizablePanel>
          <ResizableHandle withHandle />
          <ResizablePanel
            defaultSize={75}
            minSize={40}
            onResize={refreshMapSize}
          >
            <div className="h-full overflow-hidden">
              {isLoadingNewCase ? (
                <div className="h-full flex items-center py-2 px-2 transition-opacity duration-300 ease-out">
                  <div className="flex gap-2">
                    {Array.from(
                      { length: Math.max(1, caseSlides.length || 1) },
                      (_, i) => i + 1
                    ).map((i) => {
                      const availableHeight = SNAP_HEIGHTS.expanded - 20;
                      const cardHeight = Math.max(
                        32,
                        Math.min(availableHeight, 160)
                      );
                      const isCompact = cardHeight <= 50;
                      const isVeryCompact = cardHeight <= 36;
                      return (
                        <div
                          key={i}
                          className="flex-shrink-0 flex flex-row items-center gap-2 px-2 py-1 rounded border border-border animate-pulse opacity-40 bg-muted/30"
                          style={{
                            height: `${cardHeight}px`,
                            minWidth: isVeryCompact
                              ? "80px"
                              : isCompact
                              ? "100px"
                              : "80px",
                            animationDuration: "3s",
                          }}
                        >
                          <div
                            className={`${
                              isVeryCompact
                                ? "w-5 h-5"
                                : isCompact
                                ? "w-7 h-7"
                                : "w-10 h-10"
                            } bg-muted rounded flex-shrink-0`}
                          />
                          <div className="flex-1 min-w-0">
                            <div
                              className={`${
                                isVeryCompact ? "h-2" : "h-3"
                              } bg-muted rounded mb-1`}
                            />
                            {!isVeryCompact && (
                              <div className="h-2 bg-muted rounded w-8" />
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              ) : hasValidData && caseSlides.length === 0 ? (
                <div className="h-full flex items-center justify-center transition-opacity duration-300 ease-out opacity-100">
                  <div className="text-sm text-muted-foreground">
                    No images found in this case
                  </div>
                </div>
              ) : hasValidData ? (
                <div className="h-full flex items-center py-1 md:py-2 px-1 md:px-2 transition-opacity duration-300 ease-out opacity-100">
                  <div className="flex gap-1 md:gap-2 overflow-x-auto scrollbar-thin scrollbar-thumb-muted scrollbar-track-transparent">
                    {caseSlides.map((slide, index) => {
                      const isActive = slide.slideUid === currentSlideUid;
                      // Use base height for calculations - CSS will handle responsive sizing
                      const availableHeight = SNAP_HEIGHTS.expanded - 20;
                      const cardHeight = Math.max(
                        32,
                        Math.min(availableHeight, 160)
                      );
                      const isCompact = cardHeight <= 50;
                      const isVeryCompact = cardHeight <= 36;
                      return (
                        <button
                          key={slide.slideUid}
                          onClick={() => handleSlideClick(slide.slideUid)}
                          className={`flex-shrink-0 flex ${
                            isVeryCompact
                              ? "flex-row items-center gap-1 px-2 py-1"
                              : isCompact
                              ? "flex-row items-center gap-2 px-2 py-1"
                              : "flex-col items-center justify-center p-2"
                          } rounded border transition-all hover:bg-muted ${
                            isActive
                              ? "border-primary bg-primary/10 shadow-sm ring-1 ring-primary/20"
                              : "border-border hover:shadow-sm"
                          }`}
                          style={{
                            height: `${cardHeight}px`,
                            minWidth: isVeryCompact
                              ? "80px"
                              : isCompact
                              ? "100px"
                              : "80px",
                            maxWidth: isVeryCompact ? "140px" : "none",
                            transitionDelay: `${Math.min(index * 50, 300)}ms`,
                            transitionDuration: "400ms",
                          }}
                          title={`${slide.slideName} (${slide.slideUid})`}
                        >
                          <div
                            className={`${
                              isVeryCompact
                                ? "w-5 h-5"
                                : isCompact
                                ? "w-7 h-7"
                                : "w-10 h-10"
                            } rounded border flex items-center justify-center flex-shrink-0 ${
                              isActive
                                ? "bg-primary/10 text-primary border-primary/40"
                                : "bg-muted text-muted-foreground border-border"
                            }`}
                          >
                            <svg
                              className={`${
                                isVeryCompact
                                  ? "w-3 h-3"
                                  : isCompact
                                  ? "w-4 h-4"
                                  : "w-5 h-5"
                              }`}
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={1.5}
                                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2 2v12a2 2 0 002 2z"
                              />
                            </svg>
                          </div>
                          <div
                            className={`${
                              isCompact || isVeryCompact
                                ? "flex-1 min-w-0"
                                : "text-center flex flex-col justify-center"
                            } ${!isCompact && !isVeryCompact ? "mt-1" : ""}`}
                          >
                            <div
                              className={`text-xs font-medium truncate ${
                                isCompact || isVeryCompact
                                  ? "text-left"
                                  : "text-center"
                              } ${
                                isActive
                                  ? "text-foreground"
                                  : "text-foreground/80"
                              }`}
                            >
                              {slide.slideName || "Unnamed"}
                            </div>
                            {!isVeryCompact && (
                              <div
                                className={`flex ${
                                  isCompact
                                    ? "justify-start gap-2"
                                    : "flex-col justify-center"
                                } items-center text-xs text-muted-foreground ${
                                  !isCompact ? "mt-1" : ""
                                }`}
                              >
                                <span
                                  className={`${
                                    isActive
                                      ? "text-primary font-medium"
                                      : "text-muted-foreground"
                                  }`}
                                >
                                  #{index + 1}
                                </span>
                                {!isCompact &&
                                  slide.slideWidth &&
                                  slide.slideHeight && (
                                    <span className="text-xs">
                                      {Math.round(slide.slideWidth / 1000)}k×
                                      {Math.round(slide.slideHeight / 1000)}k
                                    </span>
                                  )}
                              </div>
                            )}
                          </div>
                          {isActive && (
                            <div
                              className={`${
                                isCompact || isVeryCompact ? "ml-2" : "mt-1"
                              } w-2 h-2 bg-primary rounded-full flex-shrink-0`}
                            />
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ) : (
                <div className="h-full flex items-center justify-center transition-opacity duration-300 ease-out opacity-100">
                  <div className="text-center">
                    <div className="text-sm text-muted-foreground mb-1">
                      No case selected
                    </div>
                    <div className="text-xs text-muted-foreground/70">
                      Navigate to an image to view case information
                    </div>
                  </div>
                </div>
              )}
            </div>
          </ResizablePanel>
        </ResizablePanelGroup>
      )}

      {/* Server Error Dialog */}
      <AlertDialog open={showServerError}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Can't connect right now</AlertDialogTitle>
            <AlertDialogDescription>
              We're having trouble reaching the server. This might be temporary
              - try waiting a moment and checking again.
              <br />
              <br />
              If this keeps happening, you might want to check your internet
              connection or contact support.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCloseServerError}>
              Close
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleRetryFetchCaseSlides}>
              Try Again
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Cases List Modal */}
      <CasesListModal
        isOpen={showCasesListModal}
        onClose={() => setShowCasesListModal(false)}
        onCaseSelect={handleCaseSelect}
        studyUid={studyUid}
        filterParams={filterParams}
      />
    </div>
  );
}

// Simple Cases List Modal Component
interface CasesListModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCaseSelect: (caseUid: string) => void;
  studyUid?: string; // Optional study context
  filterParams?: {
    searchQuery?: string;
    searchName?: string;
    sortBy?: string;
    sortDir?: string;
    page?: number;
    limit?: number;
    isSubsetMode?: boolean;
    hasVectorAnnotations?: string;
    hasRasterAnnotations?: string;
    status?: string;
  };
}

function CasesListModal({
  isOpen,
  onClose,
  onCaseSelect,
  studyUid,
  filterParams,
}: CasesListModalProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1); // Reset to first page when search changes
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Reset modal state when opened/closed
  useEffect(() => {
    if (isOpen) {
      setCurrentPage(1);
      setSearchQuery("");
      setDebouncedSearchQuery("");
    }
  }, [isOpen]);

  // Use study-specific or global cases based on context
  const hookOptions = {
    page: currentPage,
    limit: 20,
    q: debouncedSearchQuery || filterParams?.searchQuery || "",
    name: filterParams?.searchName || "",
    status: filterParams?.status || "",
    has_vector_annotations: filterParams?.hasVectorAnnotations || "",
    has_raster_annotations: filterParams?.hasRasterAnnotations || "",
    sort: filterParams?.sortBy || "",
    dir: filterParams?.sortDir || "",
    withAnnotations: true, // Show annotation counts
  };

  const allCasesResult = useCases(studyUid ? {} : hookOptions);
  const studyCasesResult = useCasesByStudy(
    studyUid,
    studyUid ? hookOptions : {}
  );

  const { cases, pagination, loading, error } = studyUid
    ? studyCasesResult
    : allCasesResult;

  const handleCaseClick = useCallback(
    (caseUid: string) => {
      onCaseSelect(caseUid);
      onClose();
    },
    [onCaseSelect, onClose, studyUid]
  );

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-background dark:bg-card rounded-lg shadow-xl w-full max-w-4xl max-h-[80vh] m-4 border border-border"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div className="flex items-center justify-between p-6 border-b border-border">
          <div>
            <h2 className="text-xl font-semibold text-foreground">
              {studyUid ? "Study Cases" : "All Cases"}
              {filterParams &&
                (filterParams.searchQuery ||
                  filterParams.searchName ||
                  filterParams.status ||
                  filterParams.hasVectorAnnotations ||
                  filterParams.hasRasterAnnotations) && (
                  <span className="ml-2 text-sm font-normal text-blue-600">
                    (Filtered)
                  </span>
                )}
            </h2>
            {studyUid && (
              <p className="text-sm text-muted-foreground mt-1">
                Browsing cases in current study
                {filterParams &&
                  (filterParams.searchQuery ||
                    filterParams.searchName ||
                    filterParams.status ||
                    filterParams.hasVectorAnnotations ||
                    filterParams.hasRasterAnnotations) && (
                    <span className="text-blue-600"> with active filters</span>
                  )}
              </p>
            )}
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 bg-muted hover:bg-muted/80 rounded-full flex items-center justify-center transition-colors"
          >
            <svg
              className="w-4 h-4 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {/* Search */}
        <div className="p-6 border-b border-border">
          <input
            type="text"
            placeholder={`Search ${studyUid ? "study " : ""}cases...`}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 border border-input rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring"
          />
        </div>

        {/* Cases List */}
        <div
          className="flex-1 overflow-y-auto p-6"
          style={{ maxHeight: "400px" }}
        >
          {loading && (
            <div className="flex justify-center items-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          )}

          {error && (
            <div className="text-center py-8 text-red-600">{error}</div>
          )}

          {!loading && !error && cases.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              {searchQuery
                ? `No cases found matching "${searchQuery}"`
                : "No cases found"}
            </div>
          )}

          {!loading && !error && cases.length > 0 && (
            <div className="grid gap-3">
              {cases.map((caseItem) => (
                <button
                  key={caseItem.caseUid}
                  onClick={() => handleCaseClick(caseItem.caseUid)}
                  className="flex items-center p-4 border border-border rounded-lg hover:border-primary hover:bg-muted transition-colors text-left"
                >
                  <div className="w-8 h-8 bg-primary/10 rounded flex items-center justify-center mr-3">
                    <svg
                      className="w-5 h-5 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                      />
                    </svg>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-foreground truncate">
                      {caseItem.name}
                    </div>
                    <div className="text-sm text-muted-foreground flex items-center space-x-3">
                      <span>
                        {caseItem.slideCount} image
                        {caseItem.slideCount !== 1 ? "s" : ""}
                      </span>
                      {caseItem.annotationCount !== undefined && (
                        <span className="text-primary">
                          {caseItem.annotationCount} annotation
                          {caseItem.annotationCount !== 1 ? "s" : ""}
                        </span>
                      )}
                      <span className="text-muted-foreground/70">
                        ID: {caseItem.caseUid}
                      </span>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Pagination */}
        {pagination.totalPages > 1 && (
          <div className="flex items-center justify-between p-6 border-t border-border">
            <div className="text-sm text-muted-foreground">
              Showing {(currentPage - 1) * 20 + 1}-
              {Math.min(currentPage * 20, pagination.total)} of{" "}
              {pagination.total} cases
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => setCurrentPage(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-3 py-2 text-sm font-medium text-foreground bg-background border border-border rounded-md hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <span className="px-3 py-2 text-sm font-medium text-muted-foreground">
                Page {currentPage} of {pagination.totalPages}
              </span>
              <button
                onClick={() => setCurrentPage(currentPage + 1)}
                disabled={currentPage === pagination.totalPages}
                className="px-3 py-2 text-sm font-medium text-foreground bg-background border border-border rounded-md hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
