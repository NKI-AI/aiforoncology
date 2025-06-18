// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useEffect } from "react";
import { useLocation, useNavigate, Link } from "react-router-dom";
import { useAuth } from "../auth";
import type { LoginResponse, UserData } from "../auth/authProvider";
import { jwtDecode } from "jwt-decode";
import { apiFetch, ApiError } from "../utils/fetchUtils";
import { SlideScopeIcon, EyeIcon, EyeSlashIcon } from "./icons";

// Check for development environment
const isDevelopment = process.env.NODE_ENV === "development";

interface LoginFormProps {
  username: string;
  setUsername: (username: string) => void;
  password: string;
  setPassword: (password: string) => void;
  loading: boolean;
  loginSuccess: boolean;
  onSubmit: (e: React.FormEvent) => Promise<void>;
  error: string | null;
  redirectMessage: string | null;
}

function LoginForm({
  username,
  setUsername,
  password,
  setPassword,
  loading,
  loginSuccess,
  onSubmit,
  error,
  redirectMessage,
}: LoginFormProps) {
  const [showPassword, setShowPassword] = useState(false);

  const handleMouseDown = () => {
    setShowPassword(true);
  };

  const handleMouseUp = () => {
    setShowPassword(false);
  };

  const handleMouseLeave = () => {
    setShowPassword(false);
  };

  return (
    <>
      {/* Redirect Message */}
      {redirectMessage && (
        <div
          className="mb-4 text-sm text-blue-300 bg-blue-900/30 border border-blue-800 rounded px-3 py-2"
          role="alert"
        >
          {redirectMessage}
        </div>
      )}

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
      {loginSuccess && (
        <div
          className="mb-4 text-sm text-green-300 bg-green-900/30 border border-green-800 rounded px-3 py-2"
          role="alert"
        >
          Login successful! Redirecting you...
        </div>
      )}

      {/* Login Form */}
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="username"
            className="block text-sm font-medium text-gray-300"
          >
            Username
          </label>
          <input
            id="username"
            name="username"
            type="text"
            required
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-100 text-sm"
            autoComplete="username"
            autoCapitalize="none"
            spellCheck="false"
          />
        </div>

        <div>
          <label
            htmlFor="password"
            className="block text-sm font-medium text-gray-300"
          >
            Password
          </label>
          <div className="relative mt-1">
            <input
              id="password"
              name="password"
              type={showPassword ? "text" : "password"}
              required
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="block w-full px-3 py-2 pr-10 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-100 text-sm"
              autoComplete="current-password"
            />
            <button
              type="button"
              onMouseDown={handleMouseDown}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseLeave}
              className="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 hover:text-gray-300 focus:outline-none focus:text-gray-300"
              aria-label={showPassword ? "Hide password" : "Show password"}
            >
              {showPassword ? (
                <EyeSlashIcon className="h-4 w-4" />
              ) : (
                <EyeIcon className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        <div className="flex items-center justify-end text-sm">
          <Link
            to="/account/reset_password"
            className="text-indigo-400 hover:text-indigo-300"
          >
            Forgot password?
          </Link>
        </div>

        <button
          type="submit"
          disabled={loading || loginSuccess}
          className={`w-full py-2 px-4 rounded text-white ${
            loading || loginSuccess
              ? "bg-indigo-800 cursor-not-allowed"
              : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          } text-sm font-medium transition-colors`}
        >
          {loading ? "Signing in..." : loginSuccess ? "Signed In" : "Sign In"}
        </button>
      </form>
    </>
  );
}

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loginSuccess, setLoginSuccess] = useState(false);
  const [redirectMessage, setRedirectMessage] = useState<string | null>(null);
  const { setAuth } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // Get the redirect location or default to /slides
  const from = location.state?.from || "/slides";

  // Display a message if we were redirected here
  useEffect(() => {
    if (location.state?.message) {
      setRedirectMessage(location.state.message);
    }
  }, [location.state]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isDevelopment) {
      console.log("Login form submitted for user:", username);
    }
    setLoading(true);
    setError(null);
    setRedirectMessage(null);

    try {
      const formData = new URLSearchParams();
      formData.append("username", username);
      formData.append("password", password);

      const data = await apiFetch<LoginResponse>("/api/v1/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: formData,
      });

      // Extract user data from JWT tokens instead of making a separate API call
      let userData: UserData = { username };

      try {
        // Decode the JWT token to get user data
        const decodedToken = jwtDecode<{
          sub: string;
          roles?: string[];
          email?: string;
          exp?: number;
        }>(data.access_token);

        let refreshExp: number | undefined = undefined;
        if (data.refresh_token) {
          try {
            const decodedRefresh = jwtDecode<{ exp?: number }>(
              data.refresh_token
            );
            refreshExp = decodedRefresh.exp;
          } catch {
            // ignore
          }
        }

        // Use data from the token
        userData = {
          username: decodedToken.sub || username,
          roles: decodedToken.roles || [],
          email: decodedToken.email,
          exp: decodedToken.exp,
          refresh_exp: refreshExp,
        };

        if (isDevelopment) {
          console.log("User info extracted from token:", userData);
        }
      } catch (decodeError) {
        if (isDevelopment) {
          console.warn(
            "Could not decode token, using minimal user data",
            decodeError
          );
        }
      }

      // Update auth context with user data
      setAuth(userData, true);

      setLoginSuccess(true);
      setTimeout(() => {
        navigate(from);
      }, 500);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(`Login failed: ${err.message}`);
      } else {
        setError("An unexpected error occurred");
      }
      if (isDevelopment) {
        console.error("Login error:", err);
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
          Sign in to your account
        </p>

        <LoginForm
          username={username}
          setUsername={setUsername}
          password={password}
          setPassword={setPassword}
          loading={loading}
          loginSuccess={loginSuccess}
          onSubmit={handleSubmit}
          error={error}
          redirectMessage={redirectMessage}
        />

        {/* Footer */}
        <p className="mt-6 text-center text-xs text-gray-500">
          © 2025 SlideScope. All rights reserved.
        </p>
      </div>
    </div>
  );
}
