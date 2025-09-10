// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useEffect, useCallback } from "react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "../../../auth";
import { authKeys, UserData, LoginResponse } from "../../../auth";
import { useApiMutation } from "../../../utils/apiQueries";
import { apiFetch, ApiError } from "../../../utils/fetchUtils";

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface UseLoginOptions {
  defaultRedirectPath?: string;
}

interface UseLoginReturn {
  // Form state
  email: string;
  password: string;
  setEmail: (value: string) => void;
  setPassword: (value: string) => void;

  // Status state
  loading: boolean;
  error: string | null;
  loginSuccess: boolean;
  redirectMessage: string | null;

  // Actions
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  clearError: () => void;
}

interface LoginCredentials {
  email: string;
  password: string;
}

export function useLogin({
  defaultRedirectPath = "/studies",
}: UseLoginOptions = {}): UseLoginReturn {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loginSuccess, setLoginSuccess] = useState(false);
  const [redirectMessage, setRedirectMessage] = useState<string | null>(null);

  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Get search params from TanStack Router
  const searchParams = useSearch({ from: "/login" }) as {
    from?: string;
    message?: string;
  };

  // Get the redirect location or default
  const from = searchParams.from || defaultRedirectPath;

  // Safeguard: Don't redirect to login or root to avoid loops
  const safeFrom =
    from === "/login" || from === "/" || !from ? defaultRedirectPath : from;

  // Login mutation using standardized API mutation
  const loginMutation = useApiMutation(
    async ({ email, password }: LoginCredentials): Promise<UserData> => {
      const formData = new URLSearchParams();
      formData.append("email", email);
      formData.append("password", password);

      try {
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
        const response = await apiFetch<{ userUid: string; scopes: string[] }>(
          "/api/v1/auth/me"
        );

        const userData: UserData = {
          userUid: response.userUid,
          email: email, // Use email from login form since /me doesn't return it yet
          roles: response.scopes || [],
        };

        return userData;
      } catch (error) {
        throw error;
      }
    },
    {
      onSuccess: (userData) => {
        // Update the auth context and query cache
        queryClient.setQueryData(authKeys.me(), userData);
        setLoginSuccess(true);

        // Navigate immediately
        setTimeout(() => {
          navigate({ to: safeFrom });
        }, 100);
      },
      onError: (err) => {
        if (err instanceof ApiError) {
          // Handle specific error cases
          if (err.status === 423) {
            // Password reset required - redirect to change password form
            navigate({
              to: "/account/reset_password",
              search: {
                forced_change: "true",
                email: email,
                message:
                  "You must change your password before logging in. Use your current password below.",
              },
            });
            return;
          } else if (err.status === 403) {
            setError(
              "Your account is not active. Please contact your administrator."
            );
          } else {
            setError(`Login failed: ${err.message}`);
          }
        } else {
          setError("An unexpected error occurred");
        }
      },
    }
  );

  // Display a message if we were redirected here
  useEffect(() => {
    if (searchParams.message) {
      setRedirectMessage(searchParams.message);
    }
  }, [searchParams.message]);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      if (!email.trim() || !password.trim()) {
        setError("Please fill in all fields");
        return;
      }

      setError(null);
      setRedirectMessage(null);

      try {
        await loginMutation.mutateAsync({ email, password });
      } catch (error) {
        // Error is already handled in onError
      }
    },
    [email, password, loginMutation]
  );

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    // Form state
    email,
    password,
    setEmail,
    setPassword,

    // Status state
    loading: loginMutation.isPending,
    error,
    loginSuccess,
    redirectMessage,

    // Actions
    handleSubmit,
    clearError,
  };
}
