// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth";
import { ApiError, apiFetch } from "../utils/fetchUtils";

export function useAuthApi() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  /**
   * Core request wrapper:
   *  - Calls apiFetch<T>()
   *  - On 401: signs out & redirects to /login
   */
  const request = useCallback(
    async <T>(url: string, init: RequestInit = {}): Promise<T> => {
      try {
        return await apiFetch<T>(url, {
          credentials: "include",
          ...init,
        });
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          logout();
          navigate("/login");
          throw new Error("Session expired. Please sign in again.");
        }
        throw err;
      }
    },
    [logout, navigate]
  );

  /**
   * Log out both server- and client-side
   */
  const logoutFn: () => Promise<void> = useCallback(async () => {
    // Best effort to call server logout
    try {
      await request("/api/v1/auth/logout", { method: "POST" });
    } catch {
      console.warn("Logout endpoint failed; continuing client-side sign out");
    }

    // Clear client auth state & redirect
    logout();
    navigate("/login");
  }, [request, logout, navigate]);

  // Shorthand methods for common HTTP verbs
  const get = useCallback(
    <T>(url: string) => request<T>(url, { method: "GET" }),
    [request]
  );
  const post = useCallback(
    <T>(url: string, body: any) =>
      request<T>(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    [request]
  );
  const put = useCallback(
    <T>(url: string, body: any) =>
      request<T>(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    [request]
  );
  const patch = useCallback(
    <T>(url: string, body: any) =>
      request<T>(url, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    [request]
  );
  const del = useCallback(
    <T>(url: string) => request<T>(url, { method: "DELETE" }),
    [request]
  );

  return {
    request,
    get,
    post,
    put,
    patch,
    delete: del,
    logout: logoutFn,
    // Cookie-based auth can't be introspected client-side
    isLoggedIn: () => true,
  };
}
