// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { redirect } from "@tanstack/react-router";
import { apiFetch } from "@/utils/fetchUtils";

interface RedirectOptions {
  // Route template for successful redirect
  successRoute: string;
  // Route for fallback redirect if no slides found
  fallbackRoute: string;
  // Function to extract case UID from params
  getCaseUid: (params: Record<string, string>) => string;
  // Optional function to extract additional params for the redirect
  getExtraParams?: (params: Record<string, string>) => Record<string, string>;
}

/**
 * Utility function to get the first slide UID for a given case.
 * Returns null if no slides are found or if there's an error.
 */
export async function getFirstSlideUid(
  caseUid: string
): Promise<string | null> {
  try {
    const caseData = (await apiFetch(`/api/v1/cases/${caseUid}/slides`)) as any;
    const slides = caseData.slides || [];

    if (slides.length > 0) {
      const firstSlide = slides[0];
      return firstSlide.slideUid || null;
    }

    return null;
  } catch (error) {
    console.error(`Failed to fetch first slide for case ${caseUid}:`, error);
    return null;
  }
}

/**
 * Generic function to create a redirect route that fetches the first slide of a case
 * and redirects to the appropriate slide viewer route.
 */
function createCaseSlideRedirect(options: RedirectOptions) {
  return async ({ params }: { params: Record<string, string> }) => {
    const { successRoute, fallbackRoute, getCaseUid, getExtraParams } = options;
    const caseUid = getCaseUid(params);

    try {
      // Fetch the first slide of the case
      const caseData = (await apiFetch(
        `/api/v1/cases/${caseUid}/slides`
      )) as any;
      const slides = caseData.slides || [];

      if (slides.length > 0) {
        const firstSlide = slides[0];
        const slideUid = firstSlide.slideUid;

        if (slideUid) {
          // Build redirect parameters
          const redirectParams = {
            caseUid: caseUid,
            slideUid: slideUid,
            ...(getExtraParams ? getExtraParams(params) : {}),
          };

          // Redirect to the first slide
          throw redirect({
            to: successRoute,
            params: redirectParams,
          });
        }
      }

      // If no slides found, redirect to fallback
      const fallbackParams = getExtraParams ? getExtraParams(params) : {};
      throw redirect({
        to: fallbackRoute,
        params: fallbackParams,
      });
    } catch (error: any) {
      // If API call fails, redirect to fallback
      if (error.status === 302 || error.data) {
        // This is a redirect, let it pass through
        throw error;
      }

      const fallbackParams = getExtraParams ? getExtraParams(params) : {};
      throw redirect({
        to: fallbackRoute,
        params: fallbackParams,
      });
    }
  };
}

/**
 * Pre-configured redirect creators for common use cases
 */

// For /v/$caseUid/ routes (case-only context)
export const createCaseOnlyRedirect = () =>
  createCaseSlideRedirect({
    successRoute: "/v/$caseUid/i/$slideUid",
    fallbackRoute: "/admin/cases",
    getCaseUid: (params) => params.caseUid,
  });

// For /studies/$studyUid/v/$caseUid/ routes (study + case context)
export const createStudyCaseRedirect = () =>
  createCaseSlideRedirect({
    successRoute: "/studies/$studyUid/v/$caseUid/i/$slideUid",
    fallbackRoute: "/studies/$studyUid",
    getCaseUid: (params) => params.caseUid,
    getExtraParams: (params) => ({ studyUid: params.studyUid }),
  });
