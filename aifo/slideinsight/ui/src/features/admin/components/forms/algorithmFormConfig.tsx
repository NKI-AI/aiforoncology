// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { FormConfig, FieldConfig } from "./FormFactory";
import { Algorithm } from "../../hooks/useAdminData";
import TenantDropdown from "../tenants/TenantDropdown";

const algorithmFields: FieldConfig[] = [
  {
    name: "tenantUid",
    label: "Tenant",
    type: "dropdown",
    required: true,
    placeholder: "Select a tenant",
    dropdownComponent: TenantDropdown,
    validation: ({ value }) => (!value ? "Please select a tenant" : undefined),
  },
  {
    name: "name",
    label: "Algorithm Name",
    type: "text",
    required: true,
    placeholder: "Enter algorithm name",
    validation: ({ value }) =>
      !value?.trim() ? "Algorithm name is required" : undefined,
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    placeholder: "Enter algorithm description (optional)",
    rows: 3,
  },
  {
    name: "version",
    label: "Version",
    type: "text",
    required: true,
    placeholder: "e.g., 1.0.0",
    validation: ({ value }) => {
      if (!value?.trim()) return "Version is required";
      // Basic semantic versioning validation
      if (!/^\d+\.\d+\.\d+(-[\w\.-]+)?(\+[\w\.-]+)?$/.test(value)) {
        return "Please use semantic versioning (e.g., 1.0.0)";
      }
      return undefined;
    },
  },
  {
    name: "endpointUrl",
    label: "Endpoint URL",
    type: "text",
    required: true,
    placeholder: "https://api.example.com/algorithm",
    validation: ({ value }) => {
      if (!value?.trim()) return "Endpoint URL is required";
      try {
        new URL(value);
        return undefined;
      } catch {
        return "Please enter a valid URL";
      }
    },
  },
  {
    name: "httpMethod",
    label: "HTTP Method",
    type: "select",
    required: true,
    placeholder: "Select HTTP method",
    options: [
      { label: "POST", value: "POST" },
      { label: "PUT", value: "PUT" },
      { label: "PATCH", value: "PATCH" },
    ],
    defaultValue: "POST",
  },
  {
    name: "executionMode",
    label: "Execution Mode",
    type: "select",
    required: true,
    placeholder: "Select execution mode",
    options: [
      { label: "Batch Processing", value: "BATCH" },
      { label: "Stream Processing", value: "STREAM" },
    ],
    defaultValue: "BATCH",
  },
  {
    name: "progressTransport",
    label: "Progress Transport",
    type: "select",
    required: true,
    placeholder: "Select progress transport",
    options: [
      { label: "WebSocket", value: "WEBSOCKET" },
      { label: "Server-Sent Events", value: "SSE" },
    ],
    defaultValue: "WEBSOCKET",
  },
];

export const algorithmFormConfig: FormConfig<Algorithm> = {
  fields: algorithmFields,
  apiEndpoints: {
    create: "/api/v1/algorithms",
    update: (algorithm: Algorithm) => `/api/v1/algorithms/${algorithm.id}`,
  },
  entityName: "Algorithm",
  themeColor: "blue",
  getDefaultValues: (algorithm?: Algorithm | null) => ({
    tenantUid: algorithm?.tenantId ? String(algorithm.tenantId) : "",
    name: algorithm?.name ?? "",
    description: algorithm?.description ?? "",
    version: algorithm?.version ?? "",
    endpointUrl: algorithm?.endpointUrl ?? "",
    httpMethod: algorithm?.httpMethod ?? "POST",
    executionMode: algorithm?.executionMode ?? "BATCH",
    progressTransport: algorithm?.progressTransport ?? "WEBSOCKET",
  }),
  prepareCreateData: (values) => ({
    tenant_uid: values.tenantUid,
    name: values.name,
    description: values.description || "",
    version: values.version,
    endpoint_url: values.endpointUrl,
    http_method: values.httpMethod,
    execution_mode: values.executionMode,
    progress_transport: values.progressTransport,
  }),
  prepareUpdateData: (values) => ({
    name: values.name,
    description: values.description || "",
    version: values.version,
    endpoint_url: values.endpointUrl,
    http_method: values.httpMethod,
    execution_mode: values.executionMode,
    progress_transport: values.progressTransport,
  }),
  getEntityTitle: (algorithm) => algorithm?.name || "",
};
