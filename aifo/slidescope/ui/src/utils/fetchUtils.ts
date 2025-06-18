// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
/**
 * Thrown whenever an HTTP response is non-OK.
 * Carries status, statusText, and any parsed JSON payload.
 */
export class ApiError extends Error {
  status: number;
  statusText: string;
  data?: any;

  constructor(status: number, statusText: string, message: string, data?: any) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.statusText = statusText;
    this.data = data;
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, ApiError);
    }
  }
}

/**
 * Single fetch wrapper for all JSON-based API calls.
 *
 * - Automatically includes cookies (httpOnly) via credentials: 'include'
 * - Parses JSON responses (ignores parse errors)
 * - Throws ApiError on non-2xx
 *
 * @param url  resource URL
 * @param init fetch options (method, headers, body, etc.)
 * @returns    parsed JSON as T
 * @throws     ApiError when res.ok === false
 */
export async function apiFetch<T>(
  url: string,
  init: RequestInit = {}
): Promise<T> {
  const res = await fetch(url, {
    credentials: "include", // send cookies
    ...init,
  });

  // Try parsing JSON (if any)
  let json: any;
  const ct = res.headers.get("content-type") || "";
  if (ct.includes("application/json")) {
    try {
      json = await res.json();
    } catch {
      // ignore parse errors
    }
  }

  // On non-2xx, wrap in ApiError
  if (!res.ok) {
    const msg = json?.message ?? `HTTP ${res.status}: ${res.statusText}`;
    throw new ApiError(res.status, res.statusText, msg, json);
  }

  // Everything’s good — return the parsed JSON (could be undefined)
  return json as T;
}
