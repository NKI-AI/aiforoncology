// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useState, useCallback } from "react";
import { useNavigate } from "@tanstack/react-router";
import { apiFetch, ApiError } from "../../../utils/fetchUtils";

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface RegistrationForm {
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  confirmPassword: string;
}

interface RegistrationData {
  email: string;
  firstName: string;
  lastName: string;
  password: string;
}

interface RegistrationResponse {
  message: string;
}

interface RegistrationFieldErrors {
  email?: string;
  firstName?: string;
  lastName?: string;
  password?: string;
  confirmPassword?: string;
}

interface UseRegisterReturn {
  // Form state
  formData: RegistrationForm;
  updateField: (field: keyof RegistrationForm) => (value: string) => void;

  // Status state
  loading: boolean;
  error: string | null;
  fieldErrors: RegistrationFieldErrors;
  success: boolean;
  successMessage: string | null;

  // Actions
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  clearError: () => void;
  clearFieldError: (field: keyof RegistrationForm) => void;
}

export function useRegister(): UseRegisterReturn {
  const [formData, setFormData] = useState<RegistrationForm>({
    email: "",
    firstName: "",
    lastName: "",
    password: "",
    confirmPassword: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<RegistrationFieldErrors>({});
  const [success, setSuccess] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const navigate = useNavigate();

  const updateField = useCallback(
    (field: keyof RegistrationForm) => (value: string) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
      // Clear errors when user starts typing
      setError(null);
      setFieldErrors((prev) => ({ ...prev, [field]: undefined }));
    },
    []
  );

  const validateForm = useCallback((): RegistrationFieldErrors | null => {
    const errors: RegistrationFieldErrors = {};

    if (!formData.email.trim()) {
      errors.email = "Email is required";
    }
    if (!formData.firstName.trim()) {
      errors.firstName = "First name is required";
    }
    if (!formData.lastName.trim()) {
      errors.lastName = "Last name is required";
    }
    if (!formData.password) {
      errors.password = "Password is required";
    } else if (formData.password.length < 8) {
      errors.password = "Password must be at least 8 characters long";
    }
    if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = "Passwords do not match";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  }, [formData]);

  const parseRegistrationError = useCallback(
    (errorMessage: string): RegistrationFieldErrors => {
      const errors: RegistrationFieldErrors = {};
      const lowerMessage = errorMessage.toLowerCase();

      // Parse backend validation errors
      if (
        lowerMessage.includes("email already registered") ||
        lowerMessage.includes("email")
      ) {
        errors.email = errorMessage;
      } else if (lowerMessage.includes("password")) {
        // Handle any password-related validation error
        // Extract the specific validation message - look for the last occurrence of ": "
        const parts = errorMessage.split(": ");
        const specificMessage =
          parts.length > 1 ? parts[parts.length - 1] : errorMessage;
        errors.password = specificMessage;
      }

      return errors;
    },
    []
  );

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      const validationErrors = validateForm();
      if (validationErrors) {
        setFieldErrors(validationErrors);
        return;
      }

      setLoading(true);
      setError(null);
      setFieldErrors({});

      try {
        const registrationData: RegistrationData = {
          email: formData.email,
          firstName: formData.firstName,
          lastName: formData.lastName,
          password: formData.password,
        };

        const data = await apiFetch<RegistrationResponse>(
          "/api/v1/auth/register",
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify(registrationData),
          }
        );

        setSuccess(true);
        setSuccessMessage(data.message);

        // Redirect to login page after 3 seconds
        setTimeout(() => {
          navigate({
            to: "/login",
            search: {
              message:
                "Registration successful! Please check your email for verification instructions and then log in.",
            },
          });
        }, 3000);
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.status === 422) {
            // Validation error - try to parse field-specific errors
            const parsedErrors = parseRegistrationError(
              err.message || "Validation failed"
            );
            if (Object.keys(parsedErrors).length > 0) {
              setFieldErrors(parsedErrors);
            } else {
              setError(`Registration failed: ${err.message}`);
            }
          } else if (err.status === 409) {
            // Conflict error (duplicate email)
            const parsedErrors = parseRegistrationError(
              err.message || "Email already exists"
            );
            if (Object.keys(parsedErrors).length > 0) {
              setFieldErrors(parsedErrors);
            } else {
              setError(`Registration failed: ${err.message}`);
            }
          } else {
            setError(`Registration failed: ${err.message}`);
          }
        } else {
          setError("An unexpected error occurred during registration");
        }
      } finally {
        setLoading(false);
      }
    },
    [formData, validateForm, parseRegistrationError, navigate]
  );

  const clearError = useCallback(() => {
    setError(null);
    setFieldErrors({});
  }, []);

  const clearFieldError = useCallback((field: keyof RegistrationForm) => {
    setFieldErrors((prev) => ({ ...prev, [field]: undefined }));
  }, []);

  return {
    // Form state
    formData,
    updateField,

    // Status state
    loading,
    error,
    fieldErrors,
    success,
    successMessage,

    // Actions
    handleSubmit,
    clearError,
    clearFieldError,
  };
}
