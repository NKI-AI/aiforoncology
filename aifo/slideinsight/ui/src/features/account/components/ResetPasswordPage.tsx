// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useCallback } from "react";
import { Link, useSearch } from "@tanstack/react-router";
import { useAccountOperations } from "../hooks/useAccountOperations";
import { useChangePassword } from "../../auth/hooks/useChangePassword";
import {
  AuthPageLayout,
  FormField,
  PasswordField,
  SubmitButton,
  Alert,
  PasswordRequirements,
} from "@/features/auth/components/FormComponents";

export default function ResetPasswordPage() {
  const [email, setEmail] = useState("");
  const [token, setToken] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [step, setStep] = useState<"request" | "confirm">("request");
  const [localError, setLocalError] = useState<string | null>(null);
  const [redirectMessage, setRedirectMessage] = useState<string | null>(null);

  const {
    loading,
    error,
    success,
    successMessage,
    requestPasswordReset,
    confirmPasswordReset,
    redirectToLogin,
    clearError,
    clearSuccess,
  } = useAccountOperations();

  // Get search params to check if we have a token, message, or forced change
  const searchParams = useSearch({ from: "/account/reset_password" }) as {
    token?: string;
    message?: string;
    forced_change?: string;
    email?: string;
  };

  // For forced password change scenario
  const {
    formData: changePasswordFormData,
    loading: changePasswordLoading,
    error: changePasswordError,
    fieldErrors: changePasswordFieldErrors,
    success: changePasswordSuccess,
    successMessage: changePasswordSuccessMessage,
    updateField: updateChangePasswordField,
    handleSubmit: handleChangePasswordSubmit,
    redirectToLogin: changePasswordRedirectToLogin,
    clearError: clearChangePasswordError,
  } = useChangePassword({
    email: searchParams.email,
    onSuccess: () => {
      // Auto-redirect to login after successful password change
      setTimeout(() => {
        // Use window.location to avoid dependency issues
        window.location.href =
          "/login?message=Password%20changed%20successfully!%20You%20can%20now%20sign%20in%20with%20your%20new%20password.";
      }, 3000);
    },
  });

  const isForcedChange = searchParams.forced_change === "true";

  // Memoize search params to prevent infinite loops
  const { token: urlToken, message: urlMessage } = searchParams;

  useEffect(() => {
    if (urlToken) {
      setToken(urlToken);
      setStep("confirm");
    }
    if (urlMessage) {
      setRedirectMessage(urlMessage);
    }
  }, [urlToken, urlMessage]);

  const handleRequestReset = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email.trim()) {
      setLocalError("Email is required");
      return;
    }

    setLocalError(null);

    try {
      await requestPasswordReset(email);
    } catch (err) {
      // Error is already handled by the hook
    }
  };

  const handleConfirmReset = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token.trim()) {
      setLocalError("Reset token is required");
      return;
    }

    if (!newPassword) {
      setLocalError("New password is required");
      return;
    }

    if (newPassword !== confirmPassword) {
      setLocalError("Passwords do not match");
      return;
    }

    if (newPassword.length < 8) {
      setLocalError("Password must be at least 8 characters long");
      return;
    }

    setLocalError(null);

    try {
      await confirmPasswordReset(token, newPassword);
      // Redirect to login page after 3 seconds on success
      setTimeout(() => {
        redirectToLogin(
          "Password reset successfully! You can now sign in with your new password."
        );
      }, 3000);
    } catch (err) {
      // Error is already handled by the hook
    }
  };

  const switchToConfirm = useCallback(() => {
    setStep("confirm");
    setLocalError(null);
    // Note: clearSuccess and clearError will be called by the useEffect below
  }, []);

  // Clear errors when switching steps - only clear local error here
  useEffect(() => {
    setLocalError(null);
    // Clear account operation errors when step changes
    if (clearError) {
      clearError();
    }
    if (clearSuccess) {
      clearSuccess();
    }
  }, [step]); // Only depend on step, not the functions

  // Get the current error (prefer local validation errors over API errors)
  const currentError = localError || error;

  // Check if we have any field errors that should prevent showing global error
  const hasFieldErrors = Object.values(changePasswordFieldErrors).some(
    (error) => !!error
  );
  const shouldShowGlobalChangePasswordError =
    changePasswordError && !hasFieldErrors;

  // If this is a forced password change, render the change password form
  if (isForcedChange) {
    return (
      <AuthPageLayout
        title="Password Change Required"
        subtitle="You must change your password to continue"
        error={shouldShowGlobalChangePasswordError ? changePasswordError : null}
        success={changePasswordSuccess}
        successMessage={changePasswordSuccessMessage}
        loading={changePasswordLoading}
        successAction={{
          text: "Continue to Sign In",
          to: "/login",
        }}
        customAlerts={
          <div className="mb-4 p-4 bg-yellow-900/20 border border-yellow-600/30 rounded-lg">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-yellow-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-300">
                  Password Change Required
                </h3>
                <div className="mt-2 text-sm text-yellow-200">
                  <p>
                    Your administrator has required you to change your password.
                    Please enter your current password and choose a new one
                    below.
                  </p>
                </div>
              </div>
            </div>
          </div>
        }
        afterForm={
          !changePasswordSuccess && (
            <div className="mt-6 text-center text-sm border-t border-gray-600/30 pt-4">
              <span className="text-muted-400">
                Can't remember your current password?{" "}
              </span>
              <button
                onClick={() =>
                  (window.location.href = "/account/reset_password")
                }
                className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
              >
                Use email reset instead
              </button>
            </div>
          )
        }
      >
        {!changePasswordSuccess && (
          <form onSubmit={handleChangePasswordSubmit} className="space-y-4">
            {/* Show email if available */}
            {searchParams.email && (
              <div className="text-sm text-muted-400 mb-4">
                Changing password for:{" "}
                <span className="text-muted-200 font-medium">
                  {searchParams.email}
                </span>
              </div>
            )}

            <PasswordField
              id="currentPassword"
              name="currentPassword"
              label="Current Password"
              placeholder="Enter your current password"
              value={changePasswordFormData.currentPassword}
              onChange={updateChangePasswordField("currentPassword")}
              required
              autoComplete="current-password"
              error={changePasswordFieldErrors.currentPassword}
            />

            <PasswordField
              id="newPassword"
              name="newPassword"
              label="New Password"
              placeholder="Enter your new password"
              value={changePasswordFormData.newPassword}
              onChange={updateChangePasswordField("newPassword")}
              required
              autoComplete="new-password"
              error={changePasswordFieldErrors.newPassword}
            />

            <PasswordField
              id="confirmPassword"
              name="confirmPassword"
              label="Confirm New Password"
              placeholder="Confirm your new password"
              value={changePasswordFormData.confirmPassword}
              onChange={updateChangePasswordField("confirmPassword")}
              required
              autoComplete="new-password"
              error={changePasswordFieldErrors.confirmPassword}
            />

            <div className="text-xs text-gray-600 dark:text-muted-400 bg-gray-100 dark:bg-gray-700/30 p-3 rounded-lg border border-gray-200 dark:border-gray-600/30">
              <div className="font-medium text-gray-800 dark:text-muted-300 mb-1">
                Password Requirements:
              </div>
              <ul className="space-y-1">
                <li>• At least 8 characters long</li>
                <li>• Contains uppercase and lowercase letters</li>
                <li>• Contains numbers and special characters</li>
                <li>• Must be different from your current password</li>
              </ul>
            </div>

            <div className="pt-2">
              <SubmitButton
                loading={changePasswordLoading}
                success={false}
                loadingText="Changing Password..."
                successText="Password Changed"
                defaultText="Change Password"
              />
            </div>
          </form>
        )}
      </AuthPageLayout>
    );
  }

  if (step === "request") {
    return (
      <AuthPageLayout
        title="Reset Password"
        subtitle="Enter your email to receive reset instructions"
        error={currentError}
        success={success}
        successMessage={successMessage}
        redirectMessage={redirectMessage}
        loading={loading}
        afterForm={
          <div className="mt-6 space-y-3">
            <div className="text-center text-sm">
              <span className="text-muted-400">Have a reset token? </span>
              <button
                onClick={switchToConfirm}
                className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
              >
                Enter token
              </button>
            </div>

            <div className="text-center text-sm">
              <span className="text-muted-400">Remember your password? </span>
              <Link
                to="/login"
                className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
              >
                Sign in
              </Link>
            </div>
          </div>
        }
      >
        {!success && (
          <form onSubmit={handleRequestReset} className="space-y-4">
            <FormField
              id="email"
              name="email"
              type="email"
              label="Email Address"
              placeholder="Enter your email address"
              value={email}
              onChange={setEmail}
              required
              autoComplete="email"
            />

            <div className="pt-2">
              <SubmitButton
                loading={loading}
                success={false}
                loadingText="Sending Reset Email..."
                successText="Email Sent"
                defaultText="Send Reset Email"
              />
            </div>
          </form>
        )}
      </AuthPageLayout>
    );
  }

  return (
    <AuthPageLayout
      title="Reset Password"
      subtitle="Enter your reset token and new password"
      error={currentError}
      success={success}
      successMessage={successMessage}
      redirectMessage={redirectMessage}
      loading={loading}
      successAction={{
        text: "Continue to Sign In",
        to: "/login",
      }}
      afterForm={
        !success && (
          <div className="mt-4 text-center text-sm">
            <span className="text-muted-400">Don't have a token? </span>
            <button
              onClick={() => setStep("request")}
              className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
            >
              Request reset
            </button>
          </div>
        )
      }
    >
      {!success && (
        <form onSubmit={handleConfirmReset} className="space-y-4">
          <FormField
            id="token"
            name="token"
            type="text"
            label="Reset Token"
            placeholder="Enter reset token from email"
            value={token}
            onChange={setToken}
            required
          />

          <PasswordField
            id="newPassword"
            name="newPassword"
            label="New Password"
            placeholder="Enter new password"
            value={newPassword}
            onChange={setNewPassword}
            required
            autoComplete="new-password"
          />

          <PasswordField
            id="confirmPassword"
            name="confirmPassword"
            label="Confirm New Password"
            placeholder="Confirm new password"
            value={confirmPassword}
            onChange={setConfirmPassword}
            required
            autoComplete="new-password"
          />

          <div className="text-xs text-gray-600 dark:text-muted-400 bg-gray-100 dark:bg-gray-700/30 p-3 rounded-lg border border-gray-200 dark:border-gray-600/30">
            <div className="font-medium text-gray-800 dark:text-muted-300 mb-1">
              Password Requirements:
            </div>
            <ul className="space-y-1">
              <li>• At least 8 characters long</li>
              <li>• Contains uppercase and lowercase letters</li>
              <li>• Contains numbers and special characters</li>
            </ul>
          </div>

          <div className="pt-2">
            <SubmitButton
              loading={loading}
              success={false}
              loadingText="Resetting Password..."
              successText="Password Reset"
              defaultText="Reset Password"
            />
          </div>
        </form>
      )}
    </AuthPageLayout>
  );
}
