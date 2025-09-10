// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect } from "react";
import { Link } from "@tanstack/react-router";
import { useAccountOperations } from "../hooks/useAccountOperations";
import {
  AuthPageLayout,
  FormField,
  SubmitButton,
} from "@/features/auth/components/FormComponents";

export default function ResendActivationPage() {
  const [email, setEmail] = useState("");
  const {
    loading,
    error,
    success,
    successMessage,
    resendActivationEmail,
    redirectToActivation,
    clearError,
    clearSuccess,
  } = useAccountOperations();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email.trim()) {
      return;
    }

    try {
      await resendActivationEmail(email);
      // Redirect to activation page after 3 seconds on success
      setTimeout(() => {
        redirectToActivation();
      }, 3000);
    } catch (err) {
      // Error is already handled by the hook
    }
  };

  // Clear any existing state when component mounts
  useEffect(() => {
    clearError();
    clearSuccess();
  }, [clearError, clearSuccess]);

  return (
    <AuthPageLayout
      title="Resend Activation Email"
      subtitle="Get a new activation token sent to your email"
      error={error ? `Failed to send activation email: ${error}` : null}
      success={success}
      successMessage={successMessage}
      loading={loading}
      successAction={{
        text: "Go to Activation Page",
        to: "/account/activate",
      }}
      afterForm={
        <div className="mt-6 space-y-3">
          <div className="text-center text-sm">
            <span className="text-muted-400">Have your activation token? </span>
            <Link
              to="/account/activate"
              className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
            >
              Activate account
            </Link>
          </div>

          <div className="text-center text-sm">
            <span className="text-muted-400">Need to create an account? </span>
            <Link
              to="/register"
              className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
            >
              Sign up
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
      {!success && (
        <form onSubmit={handleSubmit} className="space-y-4">
          <FormField
            id="email"
            name="email"
            type="email"
            label="Email Address"
            placeholder="Enter your registered email address"
            value={email}
            onChange={setEmail}
            required
            autoComplete="email"
          />

          <div className="text-xs text-gray-600 dark:text-muted-400 bg-gray-100 dark:bg-gray-700/30 p-3 rounded-lg border border-gray-200 dark:border-gray-600/30">
            We'll send a new activation token to this email address if an
            unverified account exists.
          </div>

          <div className="pt-2">
            <SubmitButton
              loading={loading}
              success={false}
              loadingText="Sending Email..."
              successText="Email Sent"
              defaultText="Send Activation Email"
            />
          </div>
        </form>
      )}
    </AuthPageLayout>
  );
}
