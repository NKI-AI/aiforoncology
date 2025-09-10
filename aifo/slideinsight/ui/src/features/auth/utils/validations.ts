// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

type ValidationErrors<T> = Partial<Record<keyof T, string>>;

// Individual field validators
const validators = {
  email: (email: string): string | undefined => {
    if (!email.trim()) {
      return "Email is required";
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return "Please enter a valid email address";
    }
    return undefined;
  },

  password: (password: string): string | undefined => {
    if (!password) {
      return "Password is required";
    }
    if (password.length < 8) {
      return "Password must be at least 8 characters long";
    }
    return undefined;
  },

  confirmPassword: (
    password: string,
    confirmPassword: string
  ): string | undefined => {
    if (!confirmPassword) {
      return "Please confirm your password";
    }
    if (password !== confirmPassword) {
      return "Passwords do not match";
    }
    return undefined;
  },

  required:
    (fieldName: string) =>
    (value: string): string | undefined => {
      if (!value?.trim()) {
        return `${fieldName} is required`;
      }
      return undefined;
    },

  token: (token: string): string | undefined => {
    if (!token?.trim()) {
      return "Token is required";
    }
    return undefined;
  },
};

// Common form validation schemas
interface LoginFormData {
  email: string;
  password: string;
}

interface RegisterFormData {
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  confirmPassword: string;
}

interface ResetPasswordFormData {
  token: string;
  newPassword: string;
  confirmPassword: string;
}

interface RequestResetFormData {
  email: string;
}

interface ActivationFormData {
  token: string;
}

interface ChangePasswordFormData {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

// Validation functions for each form type
const validateLoginForm = (
  data: LoginFormData
): ValidationErrors<LoginFormData> | null => {
  const errors: ValidationErrors<LoginFormData> = {};

  const emailError = validators.email(data.email);
  if (emailError) errors.email = emailError;

  const passwordError = validators.required("Password")(data.password);
  if (passwordError) errors.password = passwordError;

  return Object.keys(errors).length > 0 ? errors : null;
};

const validateRegisterForm = (
  data: RegisterFormData
): ValidationErrors<RegisterFormData> | null => {
  const errors: ValidationErrors<RegisterFormData> = {};

  const emailError = validators.email(data.email);
  if (emailError) errors.email = emailError;

  const firstNameError = validators.required("First name")(data.firstName);
  if (firstNameError) errors.firstName = firstNameError;

  const lastNameError = validators.required("Last name")(data.lastName);
  if (lastNameError) errors.lastName = lastNameError;

  const passwordError = validators.password(data.password);
  if (passwordError) errors.password = passwordError;

  const confirmPasswordError = validators.confirmPassword(
    data.password,
    data.confirmPassword
  );
  if (confirmPasswordError) errors.confirmPassword = confirmPasswordError;

  return Object.keys(errors).length > 0 ? errors : null;
};

const validateResetPasswordForm = (
  data: ResetPasswordFormData
): ValidationErrors<ResetPasswordFormData> | null => {
  const errors: ValidationErrors<ResetPasswordFormData> = {};

  const tokenError = validators.token(data.token);
  if (tokenError) errors.token = tokenError;

  const passwordError = validators.password(data.newPassword);
  if (passwordError) errors.newPassword = passwordError;

  const confirmPasswordError = validators.confirmPassword(
    data.newPassword,
    data.confirmPassword
  );
  if (confirmPasswordError) errors.confirmPassword = confirmPasswordError;

  return Object.keys(errors).length > 0 ? errors : null;
};

const validateRequestResetForm = (
  data: RequestResetFormData
): ValidationErrors<RequestResetFormData> | null => {
  const errors: ValidationErrors<RequestResetFormData> = {};

  const emailError = validators.email(data.email);
  if (emailError) errors.email = emailError;

  return Object.keys(errors).length > 0 ? errors : null;
};

const validateActivationForm = (
  data: ActivationFormData
): ValidationErrors<ActivationFormData> | null => {
  const errors: ValidationErrors<ActivationFormData> = {};

  const tokenError = validators.token(data.token);
  if (tokenError) errors.token = tokenError;

  return Object.keys(errors).length > 0 ? errors : null;
};

const validateChangePasswordForm = (
  data: ChangePasswordFormData
): ValidationErrors<ChangePasswordFormData> | null => {
  const errors: ValidationErrors<ChangePasswordFormData> = {};

  const currentPasswordError = validators.required("Current password")(
    data.currentPassword
  );
  if (currentPasswordError) errors.currentPassword = currentPasswordError;

  const newPasswordError = validators.password(data.newPassword);
  if (newPasswordError) errors.newPassword = newPasswordError;

  const confirmPasswordError = validators.confirmPassword(
    data.newPassword,
    data.confirmPassword
  );
  if (confirmPasswordError) errors.confirmPassword = confirmPasswordError;

  return Object.keys(errors).length > 0 ? errors : null;
};

// Export all validation functions and types
export {
  validators,
  validateLoginForm,
  validateRegisterForm,
  validateResetPasswordForm,
  validateRequestResetForm,
  validateActivationForm,
  validateChangePasswordForm,
};

export type {
  ValidationErrors,
  LoginFormData,
  RegisterFormData,
  ResetPasswordFormData,
  RequestResetFormData,
  ActivationFormData,
  ChangePasswordFormData,
};
