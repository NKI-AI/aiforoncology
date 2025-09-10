// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { FormConfig, FieldConfig } from "./FormFactory";
import { Study } from "../../../../hooks/useStudies";

const studyFields: FieldConfig[] = [
  {
    name: "name",
    label: "Study Name",
    type: "text",
    required: true,
    placeholder: "Enter study name",
    validation: ({ value }) =>
      !value
        ? "Study name is required"
        : value.length < 3
        ? "Study name must be at least 3 characters"
        : undefined,
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    placeholder: "Enter study description",
    rows: 3,
  },
  {
    name: "metadata",
    label: "Metadata (JSON)",
    type: "textarea",
    placeholder: '{"key": "value"}',
    rows: 4,
    description: "Enter valid JSON metadata for the study",
    validation: ({ value }) => {
      if (!value) return undefined;
      try {
        JSON.parse(value);
        return undefined;
      } catch {
        return "Metadata must be valid JSON";
      }
    },
  },
  {
    name: "isPublished",
    label: "Is Published",
    type: "checkbox",
    section: "settings",
    description: "Study is published and visible to users",
  },
];

export const studyFormConfig: FormConfig<Study> = {
  fields: studyFields,
  apiEndpoints: {
    create: "/api/v1/studies",
    update: (study: Study) => `/api/v1/studies/${study.studyUid}`,
  },
  entityName: "Study",
  themeColor: "green",
  getDefaultValues: (study?: Study | null) => ({
    name: study?.name ?? "",
    description: study?.description ?? "",
    metadata: study?.metadata ?? "{}",
    isPublished: study?.isPublished ?? false,
  }),
  prepareCreateData: (values) => ({
    name: values.name,
    description: values.description,
    metadata: values.metadata,
  }),
  prepareUpdateData: (values) => ({
    name: values.name,
    description: values.description,
    metadata: values.metadata,
    isPublished: values.isPublished,
  }),
  getEntityTitle: (study) => study?.name || "New Study",
};
