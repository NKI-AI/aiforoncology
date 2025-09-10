// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useCallback } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useMutation } from "@tanstack/react-query";
import { apiFetch, ApiError } from "../../../utils/fetchUtils";
import { useApiMutation, createApiMutation } from "../../../utils/apiQueries";

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface ActivationResponse {
  message: string;
}

interface ResetRequestResponse {
  message: string;
}

interface ResetConfirmResponse {
  message: string;
}

interface ResendActivationResponse {
  message: string;
}

interface ResetRequestData {
  email: string;
}

interface ResetConfirmData {
  token: string;
  newPassword: string;
}

interface ResendActivationData {
  email: string;
}

export function useAccountOperations() {
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const navigate = useNavigate();

  // Account activation mutation
  const activationMutation = useMutation({
    mutationFn: async (token: string): Promise<ActivationResponse> => {
      const data = await apiFetch<any>(
        `/api/v1/auth/verify-email?token=${encodeURIComponent(token)}`,
        { method: "GET" }
      );

      // Ensure we return a consistent format
      const response: ActivationResponse = {
        message:
          data?.message || data?.error || "Account activated successfully",
      };

      return response;
    },
    onSuccess: (data) => {
      setSuccessMessage(data.message);
    },
    onError: (err) => {},
  });

  // Password reset request mutation
  const resetRequestMutation = useApiMutation(
    createApiMutation.post<ResetRequestResponse, ResetRequestData>(
      "/api/v1/auth/reset-password"
    ),
    {
      onSuccess: (data) => {
        setSuccessMessage(data.message);
      },
      onError: (err) => {},
    }
  );

  // Password reset confirmation mutation
  const resetConfirmMutation = useApiMutation(
    createApiMutation.post<ResetConfirmResponse, ResetConfirmData>(
      "/api/v1/auth/reset-password/confirm"
    ),
    {
      onSuccess: (data) => {
        setSuccessMessage(data.message);
      },
      onError: (err) => {},
    }
  );

  // Resend activation email mutation
  const resendActivationMutation = useApiMutation(
    createApiMutation.post<ResendActivationResponse, ResendActivationData>(
      "/api/v1/auth/resend-verification"
    ),
    {
      onSuccess: (data) => {
        setSuccessMessage(data.message);
      },
      onError: (err) => {},
    }
  );

  // Helper functions
  const activateAccount = useCallback(
    async (token: string) => {
      return activationMutation.mutateAsync(token);
    },
    [activationMutation]
  );

  const requestPasswordReset = useCallback(
    async (email: string) => {
      return resetRequestMutation.mutateAsync({ email });
    },
    [resetRequestMutation]
  );

  const confirmPasswordReset = useCallback(
    async (token: string, newPassword: string) => {
      return resetConfirmMutation.mutateAsync({ token, newPassword });
    },
    [resetConfirmMutation]
  );

  const resendActivationEmail = useCallback(
    async (email: string) => {
      return resendActivationMutation.mutateAsync({ email });
    },
    [resendActivationMutation]
  );

  const redirectToLogin = useCallback(
    (message?: string) => {
      navigate({
        to: "/login",
        search: message ? { message } : undefined,
      });
    },
    [navigate]
  );

  const redirectToActivation = useCallback(
    (message?: string) => {
      navigate({
        to: "/account/activate",
      });
    },
    [navigate]
  );

  const clearError = useCallback(() => {
    activationMutation.reset();
    resetRequestMutation.reset();
    resetConfirmMutation.reset();
    resendActivationMutation.reset();
  }, [
    activationMutation,
    resetRequestMutation,
    resetConfirmMutation,
    resendActivationMutation,
  ]);

  const clearSuccess = useCallback(() => {
    setSuccessMessage(null);
  }, []);

  // Aggregate loading and error states
  const loading =
    activationMutation.isPending ||
    resetRequestMutation.isPending ||
    resetConfirmMutation.isPending ||
    resendActivationMutation.isPending;

  const error =
    activationMutation.error?.message ||
    resetRequestMutation.error?.message ||
    resetConfirmMutation.error?.message ||
    resendActivationMutation.error?.message ||
    null;

  const success =
    activationMutation.isSuccess ||
    resetRequestMutation.isSuccess ||
    resetConfirmMutation.isSuccess ||
    resendActivationMutation.isSuccess;

  return {
    loading,
    error,
    success,
    successMessage,
    activateAccount,
    requestPasswordReset,
    confirmPasswordReset,
    resendActivationEmail,
    redirectToLogin,
    redirectToActivation,
    clearError,
    clearSuccess,
    // Expose individual mutations for more granular control if needed
    mutations: {
      activation: activationMutation,
      resetRequest: resetRequestMutation,
      resetConfirm: resetConfirmMutation,
      resendActivation: resendActivationMutation,
    },
  };
}
