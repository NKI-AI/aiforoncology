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

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface ChangePasswordForm {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

interface ChangePasswordData {
  currentPassword: string;
  newPassword: string;
}

interface ChangePasswordResponse {
  message: string;
}

interface UseChangePasswordOptions {
  email?: string;
  onSuccess?: () => void;
}

interface FieldErrors {
  currentPassword?: string;
  newPassword?: string;
  confirmPassword?: string;
}

interface UseChangePasswordReturn {
  formData: ChangePasswordForm;
  setFormData: React.Dispatch<React.SetStateAction<ChangePasswordForm>>;
  loading: boolean;
  error: string | null;
  fieldErrors: FieldErrors;
  success: boolean;
  successMessage: string | null;
  updateField: (field: keyof ChangePasswordForm) => (value: string) => void;
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  redirectToLogin: (message?: string) => void;
  clearError: () => void;
  clearFieldError: (field: keyof ChangePasswordForm) => void;
}

export function useChangePassword({
  email,
  onSuccess,
}: UseChangePasswordOptions = {}): UseChangePasswordReturn {
  const [formData, setFormData] = useState<ChangePasswordForm>({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [success, setSuccess] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const navigate = useNavigate();

  const updateField = useCallback(
    (field: keyof ChangePasswordForm) => (value: string) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
      // Clear errors when user starts typing
      setError(null);
      setFieldErrors((prev) => ({ ...prev, [field]: undefined }));
    },
    []
  );

  const validateForm = useCallback((): FieldErrors | null => {
    const errors: FieldErrors = {};

    if (!formData.currentPassword.trim()) {
      errors.currentPassword = "Current password is required";
    }
    if (!formData.newPassword) {
      errors.newPassword = "New password is required";
    }
    if (formData.newPassword !== formData.confirmPassword) {
      errors.confirmPassword = "New passwords do not match";
    }
    if (formData.newPassword && formData.newPassword.length < 8) {
      errors.newPassword = "New password must be at least 8 characters long";
    }
    if (
      formData.newPassword &&
      formData.newPassword === formData.currentPassword
    ) {
      errors.newPassword =
        "New password must be different from current password";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  }, [formData]);

  const parsePasswordValidationError = useCallback(
    (errorMessage: string): FieldErrors => {
      const errors: FieldErrors = {};

      // Parse backend password validation errors
      // Format: "password validation failed: <specific error message>"
      const lowerMessage = errorMessage.toLowerCase();

      if (lowerMessage.includes("password")) {
        // Handle any password-related validation error
        // Extract the specific validation message - look for the last occurrence of ": "
        const parts = errorMessage.split(": ");
        const specificMessage =
          parts.length > 1 ? parts[parts.length - 1] : errorMessage;
        errors.newPassword = specificMessage;
      }

      return errors;
    },
    []
  );

  const changePasswordMutation = useMutation({
    mutationFn: async (
      data: ChangePasswordData
    ): Promise<ChangePasswordResponse> => {
      // Choose the appropriate endpoint based on whether we have an email (forced change)
      const endpoint = email
        ? "/api/v1/auth/forced-change-password" // No auth required
        : "/api/v1/auth/change-password"; // Requires auth

      // Prepare the request data
      const requestData = email
        ? {
            email: email,
            currentPassword: data.currentPassword,
            newPassword: data.newPassword,
          }
        : {
            currentPassword: data.currentPassword,
            newPassword: data.newPassword,
          };

      const response = await apiFetch<ChangePasswordResponse>(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestData),
      });

      return response;
    },
    onSuccess: (data) => {
      setSuccess(true);
      setSuccessMessage(data.message);
      setError(null);
      setFieldErrors({});

      // Call the optional success callback
      if (onSuccess) {
        onSuccess();
      }
    },
    onError: (err) => {
      setSuccess(false);
      setSuccessMessage(null);
      setFieldErrors({});

      if (err instanceof ApiError) {
        if (err.status === 401) {
          setFieldErrors({ currentPassword: "Current password is incorrect" });
        } else if (err.status === 422) {
          // Password validation error - try to parse field-specific errors
          const parsedErrors = parsePasswordValidationError(
            err.message || "Password validation failed"
          );
          if (Object.keys(parsedErrors).length > 0) {
            setFieldErrors(parsedErrors);
          } else {
            setError(err.message || "Password validation failed");
          }
        } else {
          setError(`Password change failed: ${err.message}`);
        }
      } else {
        setError("An unexpected error occurred while changing password");
      }
    },
  });

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      // Validate form
      const validationErrors = validateForm();
      if (validationErrors) {
        setFieldErrors(validationErrors);
        return;
      }

      setError(null);
      setFieldErrors({});

      try {
        await changePasswordMutation.mutateAsync({
          currentPassword: formData.currentPassword,
          newPassword: formData.newPassword,
        });
      } catch (err) {
        // Error is already handled by the mutation
      }
    },
    [formData, validateForm, changePasswordMutation]
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

  const clearError = useCallback(() => {
    setError(null);
    setFieldErrors({});
  }, []);

  const clearFieldError = useCallback((field: keyof ChangePasswordForm) => {
    setFieldErrors((prev) => ({ ...prev, [field]: undefined }));
  }, []);

  return {
    formData,
    setFormData,
    loading: changePasswordMutation.isPending,
    error,
    fieldErrors,
    success,
    successMessage,
    updateField,
    handleSubmit,
    redirectToLogin,
    clearError,
    clearFieldError,
  };
}
