// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

/**
 * Utility functions for handling API errors consistently across the application
 */

/**
 * Check if an error is authorization-related (401, 403, or contains specific keywords)
 */
export const isAuthorizationError = (error: any): boolean => {
  if (!error) return false;

  // Check error message
  const errorMessage = typeof error === "string" ? error : error.message || "";
  const lowerMessage = errorMessage.toLowerCase();

  // Check for common authorization error patterns
  return (
    lowerMessage.includes("unauthorized") ||
    lowerMessage.includes("insufficient permissions") ||
    lowerMessage.includes("access denied") ||
    lowerMessage.includes("forbidden") ||
    error.status === 401 ||
    error.status === 403
  );
};

/**
 * Check if an error is a not found error (404)
 */
export const isNotFoundError = (error: any): boolean => {
  return error?.status === 404;
};

/**
 * Check if an error is a server error (5xx)
 */
export const isServerError = (error: any): boolean => {
  return error?.status >= 500 && error?.status < 600;
};

/**
 * Check if an error is a network error
 */
export const isNetworkError = (error: any): boolean => {
  return (
    error?.code === "NETWORK_ERROR" ||
    error?.message?.toLowerCase().includes("network") ||
    error?.message?.toLowerCase().includes("fetch")
  );
};

/**
 * Get a user-friendly error message from an error object
 */
export const getErrorMessage = (
  error: any,
  fallbackMessage = "An unexpected error occurred"
): string => {
  if (isAuthorizationError(error)) {
    return "You are not authorized to perform this action";
  }

  if (isNotFoundError(error)) {
    return "The requested resource was not found";
  }

  if (isServerError(error)) {
    return "Server error. Please try again later";
  }

  if (isNetworkError(error)) {
    return "Network error. Please check your connection and try again";
  }

  return error?.message || fallbackMessage;
};

/**
 * Check if an error should trigger a retry
 */
export const shouldRetryError = (
  error: any,
  retryCount = 0,
  maxRetries = 3
): boolean => {
  // Don't retry authorization or not found errors
  if (isAuthorizationError(error) || isNotFoundError(error)) {
    return false;
  }

  // Don't retry if we've exceeded max retries
  if (retryCount >= maxRetries) {
    return false;
  }

  // Retry server errors and network errors
  return isServerError(error) || isNetworkError(error);
};
