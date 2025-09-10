// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import { useLogin } from "../hooks/useLogin";
import {
  AuthPageLayout,
  FormField,
  PasswordField,
  SubmitButton,
} from "./FormComponents";

export default function LoginPage() {
  const {
    email,
    password,
    setEmail,
    setPassword,
    loading,
    error,
    loginSuccess,
    redirectMessage,
    handleSubmit,
  } = useLogin();

  return (
    <AuthPageLayout
      title="Welcome back"
      subtitle="Sign in to your account"
      error={error}
      success={loginSuccess}
      successMessage="Login successful! Redirecting you..."
      redirectMessage={redirectMessage}
      loading={loading}
      afterForm={
        <div className="mt-6 text-center text-sm">
          <span className="text-muted-400">Don't have an account? </span>
          <Link
            to="/register"
            className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
          >
            Sign up
          </Link>
        </div>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <FormField
          id="email"
          name="email"
          type="email"
          label="Email"
          placeholder="Enter your email"
          value={email}
          onChange={setEmail}
          required
          autoComplete="email"
        />

        <div className="space-y-2">
          <PasswordField
            id="password"
            name="password"
            label="Password"
            placeholder="Enter your password"
            value={password}
            onChange={setPassword}
            required
            autoComplete="current-password"
          />
          <div className="flex justify-end">
            <Link
              to="/account/reset_password"
              className="text-xs text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
            >
              Forgot password?
            </Link>
          </div>
        </div>

        <div className="pt-2">
          <SubmitButton
            loading={loading}
            success={loginSuccess}
            loadingText="Signing in..."
            successText="Signed In"
            defaultText="Sign In"
          />
        </div>
      </form>
    </AuthPageLayout>
  );
}
