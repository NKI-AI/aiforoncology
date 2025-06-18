// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect } from "react";
import { useLocation, useNavigate, Link } from "react-router-dom";
import { apiFetch, ApiError } from "../utils/fetchUtils";
import { SlideScopeIcon, ArrowLeftIcon } from "./icons";

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface ResetPasswordFormProps {
  email: string;
  setEmail: (email: string) => void;
  loading: boolean;
  resetSuccess: boolean;
  onSubmit: (e: React.FormEvent) => Promise<void>;
  error: string | null;
}

function ResetPasswordForm({
  email,
  setEmail,
  loading,
  resetSuccess,
  onSubmit,
  error,
}: ResetPasswordFormProps) {
  return (
    <>
      {/* Error Message */}
      {error && (
        <div
          className="mb-4 text-sm text-red-300 bg-red-900/30 border border-red-800 rounded px-3 py-2"
          role="alert"
        >
          {error}
        </div>
      )}

      {/* Success Message */}
      {resetSuccess && (
        <div
          className="mb-4 text-sm text-green-300 bg-green-900/30 border border-green-800 rounded px-3 py-2"
          role="alert"
        >
          Password reset instructions have been sent to your email address.
        </div>
      )}

      {!resetSuccess && (
        <>
          <div className="mb-4 text-sm text-gray-400">
            Enter your email address and we'll send you instructions to reset
            your password.
          </div>

          {/* Reset Password Form */}
          <form onSubmit={onSubmit} className="space-y-4">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-300"
              >
                Email Address
              </label>
              <input
                id="email"
                name="email"
                type="email"
                required
                placeholder="your@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-100 text-sm"
                autoComplete="email"
                autoCapitalize="none"
                spellCheck="false"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className={`w-full py-2 px-4 rounded text-white ${
                loading
                  ? "bg-indigo-800 cursor-not-allowed"
                  : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              } text-sm font-medium transition-colors`}
            >
              {loading ? "Sending..." : "Send Reset Instructions"}
            </button>
          </form>
        </>
      )}

      {/* Back to Login Link */}
      <div className="mt-6 text-center">
        <Link
          to="/login"
          className="inline-flex items-center text-sm text-indigo-400 hover:text-indigo-300 transition-colors"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Login
        </Link>
      </div>
    </>
  );
}

export default function ResetPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resetSuccess, setResetSuccess] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  // Display a message if we were redirected here
  useEffect(() => {
    if (location.state?.message) {
      // Could handle redirect messages if needed
    }
  }, [location.state]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isDevelopment) {
      console.log("Reset password form submitted for email:", email);
    }
    setLoading(true);
    setError(null);

    try {
      const formData = new URLSearchParams();
      formData.append("email", email);

      await apiFetch("/api/v1/auth/reset_password", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: formData,
      });

      setResetSuccess(true);

      // Redirect to login after 3 seconds
      setTimeout(() => {
        navigate("/login", {
          state: {
            message: "Password reset instructions sent. Check your email.",
          },
        });
      }, 3000);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`Reset failed: ${err.message}`);
      } else {
        setError("An unexpected error occurred");
      }
      if (isDevelopment) {
        console.error("Reset password error:", err);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-sm bg-gray-800 rounded-lg p-6 shadow-lg">
        {/* Logo & Heading */}
        <div className="flex items-center justify-center space-x-2 mb-6">
          <SlideScopeIcon className="h-8 w-8 text-indigo-500" />
          <h1 className="text-2xl font-semibold text-gray-100">SlideScope</h1>
        </div>
        <p className="text-center text-sm text-gray-400 mb-6">
          Reset your password
        </p>

        <ResetPasswordForm
          email={email}
          setEmail={setEmail}
          loading={loading}
          resetSuccess={resetSuccess}
          onSubmit={handleSubmit}
          error={error}
        />

        {/* Footer */}
        <p className="mt-6 text-center text-xs text-gray-500">
          © 2025 SlideScope. All rights reserved.
        </p>
      </div>
    </div>
  );
}
