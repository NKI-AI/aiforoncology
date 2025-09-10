// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, {
  createContext,
  useContext,
  useEffect,
  ReactNode,
  useCallback,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
  useApiQuery,
  useApiMutation,
  createApiMutation,
} from "../utils/apiQueries";
import { apiFetch, ApiError } from "../utils/fetchUtils";

// ============================================================================
// Types and Interfaces
// ============================================================================

export interface UserData {
  userUid: string;
  email: string;
  roles: string[];
  isSwitched?: boolean;
  originalEmail?: string;
  originalUserUid?: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token: string;
  refresh_expires_in: number;
}

// ============================================================================
// Query Keys
// ============================================================================

export const authKeys = {
  me: () => ["auth", "me"] as const,
};

// ============================================================================
// Auth Context
// ============================================================================

interface AuthContextType {
  user: UserData | null;
  isLoading: boolean;
  error: Error | null;
  isLoggedIn: boolean;
  login: (email: string, password: string) => Promise<UserData>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  clearError: () => void;
  // Switch user functionality
  switchUser: (targetUserUid: string) => Promise<UserData>;
  switchBack: () => Promise<UserData>;
  isSwitched: boolean;
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Check if a user has superadmin role
 */
export function isSuperAdmin(user: UserData | null | undefined): boolean {
  if (!user || !user.roles) {
    return false;
  }
  return user.roles.includes("superadmin");
}

// ============================================================================
// Auth Provider
// ============================================================================

const AuthContext = createContext<AuthContextType | null>(null);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const queryClient = useQueryClient();

  // Query to fetch current user using standardized helper
  const {
    data: user,
    isLoading,
    error,
    refetch,
  } = useApiQuery<{
    userUid: string;
    email: string;
    scopes: string[];
    isSwitched?: boolean;
    originalUserUid?: string;
    originalEmail?: string;
  }>(authKeys.me(), "/api/v1/auth/me", {
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  });

  // Transform the raw user data to include switch user detection
  const transformedUser: UserData | null = user
    ? {
        userUid: user.userUid,
        email: user.email || user.userUid,
        roles: user.scopes || [],
        isSwitched: user.isSwitched || false,
        originalEmail: user.originalEmail,
        originalUserUid: user.originalUserUid,
      }
    : null;

  // Login mutation using standardized helper
  const loginMutation = useApiMutation(
    async ({
      email,
      password,
    }: {
      email: string;
      password: string;
    }): Promise<UserData> => {
      const formData = new URLSearchParams();
      formData.append("email", email);
      formData.append("password", password);

      // Login - this will set HTTP-only cookies
      const loginResponse = await apiFetch<LoginResponse>(
        "/api/v1/auth/login",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/x-www-form-urlencoded",
          },
          body: formData,
        }
      );

      // Get user data from /me endpoint
      const response = await apiFetch<{ userUid: string; scopes: string[] }>(
        "/api/v1/auth/me"
      );

      const userData: UserData = {
        userUid: response.userUid,
        email: email, // Use the email from login since /me doesn't return it yet
        roles: response.scopes || [],
      };

      return userData;
    },
    {
      onSuccess: (userData) => {
        queryClient.setQueryData(authKeys.me(), userData);
      },
      onError: (error) => {},
    }
  );

  // Logout mutation using standardized helper
  const logoutMutation = useApiMutation(
    (_: void) => createApiMutation.post<void, {}>("/api/v1/auth/logout")({}),
    {
      onSuccess: () => {
        queryClient.removeQueries({ queryKey: authKeys.me() });
      },
      onError: (error) => {
        // Clear cache anyway on logout failure
        queryClient.removeQueries({ queryKey: authKeys.me() });
      },
    }
  );

  // Refresh function
  const refresh = async () => {
    await refetch();
  };

  // Clear error function
  const clearError = () => {
    queryClient.setQueryData(authKeys.me(), user);
  };

  // Create wrapper functions for the login/logout mutations
  const loginWrapper = useCallback(
    async (email: string, password: string): Promise<UserData> => {
      return loginMutation.mutateAsync({ email, password });
    },
    [loginMutation.mutateAsync]
  );

  const logoutWrapper = useCallback(async (): Promise<void> => {
    return logoutMutation.mutateAsync(undefined);
  }, [logoutMutation.mutateAsync]);

  // Switch user mutation
  const switchUserMutation = useApiMutation(
    async (targetUserUid: string): Promise<UserData> => {
      const response = await apiFetch<LoginResponse>(
        "/api/v1/auth/switch-user",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ targetUserUid }),
        }
      );

      // Get updated user data from /me endpoint after switching
      const meResponse = await apiFetch<{
        userUid: string;
        email: string;
        scopes: string[];
        isSwitched?: boolean;
        originalUserUid?: string;
        originalEmail?: string;
      }>("/api/v1/auth/me");

      const userData: UserData = {
        userUid: meResponse.userUid,
        email: meResponse.email || meResponse.userUid,
        roles: meResponse.scopes || [],
        isSwitched: meResponse.isSwitched || false,
        originalEmail: meResponse.originalEmail,
        originalUserUid: meResponse.originalUserUid,
      };

      return userData;
    },
    {
      onSuccess: (userData) => {
        queryClient.setQueryData(authKeys.me(), userData);
      },
    }
  );

  // Switch back mutation
  const switchBackMutation = useApiMutation(
    async (): Promise<UserData> => {
      const response = await apiFetch<LoginResponse>(
        "/api/v1/auth/switch-back",
        {
          method: "POST",
        }
      );

      // Get updated user data from /me endpoint after switching back
      const meResponse = await apiFetch<{
        userUid: string;
        email: string;
        scopes: string[];
        isSwitched?: boolean;
        originalUserUid?: string;
        originalEmail?: string;
      }>("/api/v1/auth/me");

      const userData: UserData = {
        userUid: meResponse.userUid,
        email: meResponse.email || meResponse.userUid,
        roles: meResponse.scopes || [],
        isSwitched: meResponse.isSwitched || false,
        originalEmail: meResponse.originalEmail,
        originalUserUid: meResponse.originalUserUid,
      };

      return userData;
    },
    {
      onSuccess: (userData) => {
        queryClient.setQueryData(authKeys.me(), userData);
      },
    }
  );

  const value: AuthContextType = {
    user: transformedUser,
    isLoading,
    error: error || null,
    isLoggedIn: !!transformedUser,
    login: loginWrapper,
    logout: logoutWrapper,
    refresh,
    clearError,
    switchUser: switchUserMutation.mutateAsync,
    switchBack: async () => switchBackMutation.mutateAsync(undefined),
    isSwitched: !!transformedUser?.isSwitched,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
