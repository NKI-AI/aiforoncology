// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import { useRegister } from "../hooks/useRegister";
import {
  AuthPageLayout,
  FormField,
  PasswordField,
  SubmitButton,
  PasswordRequirements,
} from "./FormComponents";

export default function RegisterPage() {
  const {
    formData,
    updateField,
    loading,
    error,
    fieldErrors,
    success,
    successMessage,
    handleSubmit,
  } = useRegister();

  // Check if we have any field errors that should prevent showing global error
  const hasFieldErrors = Object.values(fieldErrors).some((error) => !!error);
  const shouldShowGlobalError = error && !hasFieldErrors;

  return (
    <AuthPageLayout
      title="Create Account"
      subtitle="Sign up for a new SlideInsight account"
      error={shouldShowGlobalError ? error : null}
      success={success}
      successMessage={successMessage}
      loading={loading}
      successAction={{
        text: "Have your token? Activate now",
        to: "/account/activate",
      }}
      beforeForm={
        success && (
          <div className="text-sm text-muted-300 text-center">
            Check your email for the activation token.
          </div>
        )
      }
      afterForm={
        <div className="mt-6 text-center text-sm">
          <span className="text-muted-400">Already have an account? </span>
          <Link
            to="/login"
            className="text-indigo-400 hover:text-indigo-300 transition-colors duration-200"
          >
            Sign in
          </Link>
        </div>
      }
    >
      {!success && (
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <FormField
              id="firstName"
              name="firstName"
              type="text"
              label="First Name"
              placeholder="Enter first name"
              value={formData.firstName}
              onChange={updateField("firstName")}
              required
              autoComplete="given-name"
              error={fieldErrors.firstName}
            />
            <FormField
              id="lastName"
              name="lastName"
              type="text"
              label="Last Name"
              placeholder="Enter last name"
              value={formData.lastName}
              onChange={updateField("lastName")}
              required
              autoComplete="family-name"
              error={fieldErrors.lastName}
            />
          </div>

          <FormField
            id="email"
            name="email"
            type="email"
            label="Email"
            placeholder="Enter your email"
            value={formData.email}
            onChange={updateField("email")}
            required
            autoComplete="email"
            error={fieldErrors.email}
          />

          <PasswordField
            id="password"
            name="password"
            label="Password"
            placeholder="Enter your password"
            value={formData.password}
            onChange={updateField("password")}
            required
            autoComplete="new-password"
            error={fieldErrors.password}
          />

          <PasswordField
            id="confirmPassword"
            name="confirmPassword"
            label="Confirm Password"
            placeholder="Confirm your password"
            value={formData.confirmPassword}
            onChange={updateField("confirmPassword")}
            required
            autoComplete="new-password"
            error={fieldErrors.confirmPassword}
          />

          <PasswordRequirements />

          <div className="pt-2">
            <SubmitButton
              loading={loading}
              success={false}
              loadingText="Creating Account..."
              successText="Account Created"
              defaultText="Create Account"
            />
          </div>
        </form>
      )}
    </AuthPageLayout>
  );
}
