// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { FormConfig, FieldConfig } from "./FormFactory";
import { User } from "../../hooks/useAdminData";
import TenantDropdown from "../tenants/TenantDropdown";

const userFields: FieldConfig[] = [
  {
    name: "tenantUid",
    label: "Tenant",
    type: "dropdown",
    required: true,
    placeholder: "Select a tenant...",
    disabled: (isEdit) => isEdit,
    dropdownComponent: TenantDropdown,
    validation: ({ value }) => (!value ? "Tenant is required" : undefined),
  },
  {
    name: "email",
    label: "Email",
    type: "email",
    required: true,
    placeholder: "Enter email address",
    validation: ({ value }) => {
      if (!value) return "Email is required";
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return !emailRegex.test(value)
        ? "Please enter a valid email address"
        : undefined;
    },
  },
  {
    name: "firstName",
    label: "First Name",
    type: "text",
    placeholder: "Enter first name",
  },
  {
    name: "lastName",
    label: "Last Name",
    type: "text",
    placeholder: "Enter last name",
  },
  {
    name: "password",
    label: "Password",
    type: "password",
    required: true,
    placeholder: "Enter password (min 8 characters)",
    visible: (isEdit) => !isEdit,
    validation: ({ value }) =>
      !value
        ? "Password is required"
        : value.length < 8
        ? "Password must be at least 8 characters"
        : undefined,
  },
  {
    name: "mustResetPassword",
    label: "Must Reset Password",
    type: "checkbox",
    section: "settings",
    visible: (isEdit) => isEdit,
    description: "Require password reset on next login",
  },
  {
    name: "isActive",
    label: "Is Active",
    type: "checkbox",
    section: "settings",
    visible: (isEdit) => isEdit,
    description: "User account is active",
  },
];

export const userFormConfig: FormConfig<User> = {
  fields: userFields,
  apiEndpoints: {
    create: "/api/v1/users",
    update: (user: User) => `/api/v1/users/${user.userUid}`,
  },
  entityName: "User",
  themeColor: "blue",
  getDefaultValues: (user?: User | null) => ({
    email: user?.email ?? "",
    firstName: user?.firstName ?? "",
    lastName: user?.lastName ?? "",
    password: "",
    mustResetPassword: user?.mustResetPassword ?? false,
    isActive: user?.isActive ?? true,
    tenantUid: user?.tenantUid ? String(user.tenantUid) : "",
  }),
  prepareCreateData: (values) => ({
    email: values.email,
    firstName: values.firstName,
    lastName: values.lastName,
    password: values.password,
    tenantUid: values.tenantUid,
  }),
  prepareUpdateData: (values) => ({
    email: values.email,
    firstName: values.firstName,
    lastName: values.lastName,
    mustResetPassword: values.mustResetPassword,
    isActive: values.isActive,
  }),
  getEntityTitle: (user) => user?.email || "New User",
};
