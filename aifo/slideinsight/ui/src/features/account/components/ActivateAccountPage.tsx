// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useRef } from "react";
import { Link, useSearch } from "@tanstack/react-router";
import { useAccountOperations } from "../hooks/useAccountOperations";
import {
  AuthPageLayout,
  FormField,
  SubmitButton,
  Alert,
} from "@/features/auth/components/FormComponents";

export default function ActivateAccountPage() {
  const [token, setToken] = useState("");
  const [autoActivating, setAutoActivating] = useState(false);
  const [redirectCountdown, setRedirectCountdown] = useState<number | null>(
    null
  );

  const {
    loading,
    error,
    success,
    successMessage,
    activateAccount,
    redirectToLogin,
    clearError,
  } = useAccountOperations();

  // Use a ref to store the stable redirectToLogin function
  const redirectToLoginRef = useRef(redirectToLogin);
  redirectToLoginRef.current = redirectToLogin;

  // Debug logging in development
  const isDevelopment = process.env.NODE_ENV === "development";

  // Get search params to check if we have a token
  const searchParams = useSearch({ from: "/account/activate" }) as {
    token?: string;
    activation_code?: string;
    message?: string;
  };

  useEffect(() => {
    // Check for token in URL params (either 'token' or 'activation_code')
    const urlToken = searchParams.token || searchParams.activation_code;
    if (urlToken) {
      setToken(urlToken);
      setAutoActivating(true);
      // Automatically activate if token is in URL
      handleActivation(urlToken);
    }
  }, [searchParams.token, searchParams.activation_code]);

  // Handle countdown and redirect after successful activation
  useEffect(() => {
    if (success && !redirectCountdown) {
      setRedirectCountdown(3);
    }
  }, [success, redirectCountdown]);

  useEffect(() => {
    if (redirectCountdown !== null && redirectCountdown > 0) {
      const timer = setTimeout(() => {
        setRedirectCountdown(redirectCountdown - 1);
      }, 1000);
      return () => clearTimeout(timer);
    } else if (redirectCountdown === 0) {
      redirectToLoginRef.current(
        "Account activated successfully! You can now sign in."
      );
    }
  }, [redirectCountdown]);

  const handleActivation = async (activationToken: string) => {
    try {
      const result = await activateAccount(activationToken);
    } catch (err) {
      // Error is already handled by the hook
    } finally {
      setAutoActivating(false);
    }
  };

  const handleManualActivation = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token.trim()) {
      return;
    }

    await handleActivation(token);
  };

  // Clear any existing errors when component mounts or token changes
  useEffect(() => {
    clearError();
  }, [clearError]);

  return (
    <AuthPageLayout
      title="Activate Account"
      subtitle="Complete your account activation"
      error={error ? `Activation failed: ${error}` : null}
      success={success}
      successMessage={successMessage}
      redirectMessage={searchParams.message}
      loading={loading}
      customAlerts={
        <>
          {/* Auto-activation in progress */}
          {autoActivating && (
            <div className="mb-4">
              <Alert type="info" message="Activating your account..." />
            </div>
          )}

          {/* Countdown alert */}
          {redirectCountdown !== null && redirectCountdown > 0 && (
            <div className="mb-4 p-3 bg-green-900/20 border border-green-600/30 rounded-lg text-center">
              <p className="text-green-300 text-sm">
                Redirecting to sign in page in {redirectCountdown} second
                {redirectCountdown !== 1 ? "s" : ""}...
              </p>
            </div>
          )}
        </>
      }
      afterForm={
        <div className="mt-6 space-y-3">
          <div className="text-center text-sm">
            <span className="text-muted-400">Didn't receive an email? </span>
            <Link
              to="/account/resend-activation"
              className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
            >
              Send new activation email
            </Link>
          </div>

          <div className="text-center text-sm">
            <span className="text-muted-400">Already activated? </span>
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
      {/* Manual activation form - only show if no auto-activation and not successful */}
      {!autoActivating && !success && (
        <form onSubmit={handleManualActivation} className="space-y-4">
          <FormField
            id="token"
            name="token"
            type="text"
            label="Activation Token"
            placeholder="Enter activation token from email"
            value={token}
            onChange={setToken}
            required
          />

          <div className="text-xs text-gray-600 dark:text-muted-400 bg-gray-100 dark:bg-gray-700/30 p-3 rounded-lg border border-gray-200 dark:border-gray-600/30">
            Check your email for the activation token and paste it above.
          </div>

          <div className="pt-2">
            <SubmitButton
              loading={loading}
              success={false}
              loadingText="Activating Account..."
              successText="Account Activated"
              defaultText="Activate Account"
            />
          </div>
        </form>
      )}
    </AuthPageLayout>
  );
}
