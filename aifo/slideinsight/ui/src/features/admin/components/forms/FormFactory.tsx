// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useForm } from "@tanstack/react-form";
import { useApiMutation, createApiMutation } from "@/utils/apiQueries";
import ModalHeader from "../ModalHeader";
import { Button } from "../../../../components/ui/button";

// Field configuration interface
export interface FieldConfig {
  name: string;
  label: string;
  type:
    | "text"
    | "email"
    | "password"
    | "textarea"
    | "checkbox"
    | "dropdown"
    | "select";
  required?: boolean;
  placeholder?: string;
  validation?: (value: any) => string | undefined;
  disabled?: (isEdit: boolean) => boolean;
  visible?: (isEdit: boolean) => boolean;
  dropdownComponent?: React.ComponentType<any>;
  options?: { label: string; value: string }[]; // For select fields
  defaultValue?: any; // Default value for the field
  rows?: number; // For textarea
  description?: string; // Helper text
  section?: string; // For grouping fields into sections
}

// Custom content section interface
export interface CustomContentSection {
  name: string;
  component: React.ComponentType<{
    entity?: any;
    isEdit: boolean;
    isLoading: boolean;
    ref?: React.RefObject<any>;
  }>;
  position: "before" | "after";
  section: string;
}

// Form configuration interface
export interface FormConfig<T> {
  fields: FieldConfig[];
  customContent?: CustomContentSection[];
  apiEndpoints: {
    create: string;
    update: (entity: T) => string;
  };
  entityName: string;
  themeColor: "blue" | "green" | "purple";
  getDefaultValues: (entity?: T | null) => Record<string, any>;
  prepareCreateData: (values: Record<string, any>) => any;
  prepareUpdateData: (values: Record<string, any>) => any;
  getEntityTitle?: (entity?: T | null) => string;
  postCreateHook?: (
    entity: T,
    customContentRefs: Record<string, React.RefObject<any>>
  ) => Promise<void>;
}

// Props for the generated form component
export interface GeneratedFormProps<T> {
  entity?: T | null;
  onSuccess: (entity: T) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

// Theme color mappings
const themeColors = {
  blue: {
    ring: "focus:ring-blue-500",
    border: "focus:border-blue-500",
    button: "bg-blue-600 hover:bg-blue-700 focus:ring-blue-500",
    checkbox: "text-blue-600 focus:ring-blue-500",
  },
  green: {
    ring: "focus:ring-green-500",
    border: "focus:border-green-500",
    button: "bg-green-600 hover:bg-green-700 focus:ring-green-500",
    checkbox: "text-green-600 focus:ring-green-500",
  },
  purple: {
    ring: "focus:ring-purple-500",
    border: "focus:border-purple-500",
    button: "bg-purple-600 hover:bg-purple-700 focus:ring-purple-500",
    checkbox: "text-purple-600 focus:ring-purple-500",
  },
};

export function createForm<T>(config: FormConfig<T>) {
  const FormComponent: React.FC<GeneratedFormProps<T>> = ({
    entity,
    onSuccess,
    onCancel,
    isLoading = false,
  }) => {
    const isEdit = Boolean(entity);
    const theme = themeColors[config.themeColor];

    // Create refs for custom content sections
    const customContentRefs = React.useMemo(() => {
      const refs: Record<string, React.RefObject<any>> = {};
      config.customContent?.forEach((section) => {
        refs[section.name] = React.createRef();
      });
      return refs;
    }, [config.customContent]);

    // Create and update mutations
    const createMutation = useApiMutation(
      createApiMutation.post<T, any>(config.apiEndpoints.create),
      {
        onSuccess: async (newEntity) => {
          // Run post-create hook if it exists
          if (config.postCreateHook && !isEdit) {
            try {
              await config.postCreateHook(newEntity, customContentRefs);
            } catch (error) {
              console.error("Post-create hook failed:", error);
              // Still call onSuccess since the entity was created
            }
          }
          onSuccess(newEntity);
        },
        onError: (error) => {
          console.error(
            `Failed to create ${config.entityName.toLowerCase()}:`,
            error
          );
          // TODO: Show error message to user
        },
      }
    );

    const updateMutation = useApiMutation(
      (data: { entity: T; updateData: any }) =>
        createApiMutation.put<T, any>(config.apiEndpoints.update(data.entity))(
          data.updateData
        ),
      {
        onSuccess,
        onError: (error) => {
          console.error(
            `Failed to update ${config.entityName.toLowerCase()}:`,
            error
          );
          // TODO: Show error message to user
        },
      }
    );

    const form = useForm({
      defaultValues: config.getDefaultValues(entity),
      onSubmit: async ({ value }) => {
        if (isEdit && entity) {
          // Update entity
          const updateData = config.prepareUpdateData(value);
          updateMutation.mutate({ entity, updateData });
        } else {
          // Create entity
          const createData = config.prepareCreateData(value);
          createMutation.mutate(createData);
        }
      },
    });

    const isMutating = createMutation.isPending || updateMutation.isPending;
    const isFormLoading = isLoading || isMutating;

    // Group fields by section
    const fieldsBySection = config.fields.reduce((acc, field) => {
      const section = field.section || "main";
      if (!acc[section]) acc[section] = [];
      acc[section].push(field);
      return acc;
    }, {} as Record<string, FieldConfig[]>);

    // Group custom content by section
    const customContentBySection =
      config.customContent?.reduce((acc, content) => {
        if (!acc[content.section])
          acc[content.section] = { before: [], after: [] };
        acc[content.section][content.position].push(content);
        return acc;
      }, {} as Record<string, { before: CustomContentSection[]; after: CustomContentSection[] }>) ||
      {};

    const renderCustomContent = (sections: CustomContentSection[]) => {
      return sections.map((section) => {
        const Component = section.component;
        return (
          <Component
            key={section.name}
            entity={entity}
            isEdit={isEdit}
            isLoading={isFormLoading}
            ref={customContentRefs[section.name]}
          />
        );
      });
    };

    const renderField = (field: FieldConfig) => {
      // Check visibility
      if (field.visible && !field.visible(isEdit)) {
        return null;
      }

      const isFieldDisabled =
        isFormLoading || (field.disabled ? field.disabled(isEdit) : false);
      const isRequired =
        field.required && (!isEdit || field.visible?.(isEdit) !== false);

      return (
        <form.Field
          key={field.name}
          name={field.name}
          validators={
            field.validation ? { onChange: field.validation } : undefined
          }
        >
          {(formField) => (
            <div>
              <label
                htmlFor={field.name}
                className="block text-sm font-medium text-muted-700 mb-2"
              >
                {field.label}
                {isRequired && <span className="text-red-500 ml-1">*</span>}
              </label>

              {field.type === "textarea" ? (
                <textarea
                  id={field.name}
                  disabled={isFieldDisabled}
                  value={formField.state.value}
                  onChange={(e) => formField.handleChange(e.target.value)}
                  onBlur={formField.handleBlur}
                  rows={field.rows || 3}
                  className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 ${
                    theme.ring
                  } ${theme.border} ${
                    formField.state.meta.errors.length > 0
                      ? "border-red-500"
                      : "border-gray-300"
                  } ${field.name === "metadata" ? "font-mono text-sm" : ""}`}
                  placeholder={field.placeholder}
                />
              ) : field.type === "checkbox" ? (
                <div className="flex items-center">
                  <input
                    id={field.name}
                    type="checkbox"
                    disabled={isFieldDisabled}
                    checked={formField.state.value}
                    onChange={(e) => formField.handleChange(e.target.checked)}
                    className={`h-4 w-4 ${theme.checkbox} border-gray-300 rounded`}
                  />
                  <label
                    htmlFor={field.name}
                    className="ml-2 block text-sm text-muted-700"
                  >
                    {field.description || field.label}
                  </label>
                </div>
              ) : field.type === "dropdown" && field.dropdownComponent ? (
                <field.dropdownComponent
                  value={formField.state.value}
                  onChange={formField.handleChange}
                  onBlur={formField.handleBlur}
                  disabled={isFieldDisabled}
                  placeholder={field.placeholder}
                  error={
                    formField.state.meta.errors.length > 0
                      ? formField.state.meta.errors[0]
                      : undefined
                  }
                  required={isRequired}
                />
              ) : field.type === "select" ? (
                <select
                  id={field.name}
                  disabled={isFieldDisabled}
                  value={formField.state.value || ""}
                  onChange={(e) => formField.handleChange(e.target.value)}
                  onBlur={formField.handleBlur}
                  className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 ${
                    theme.ring
                  } ${theme.border} ${
                    isFieldDisabled ? "bg-gray-50 text-muted-500" : ""
                  } ${
                    formField.state.meta.errors.length > 0
                      ? "border-red-500"
                      : "border-gray-300"
                  }`}
                >
                  {field.placeholder && (
                    <option value="" disabled>
                      {field.placeholder}
                    </option>
                  )}
                  {field.options?.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  id={field.name}
                  type={field.type}
                  disabled={isFieldDisabled}
                  value={formField.state.value}
                  onChange={(e) => formField.handleChange(e.target.value)}
                  onBlur={formField.handleBlur}
                  className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 ${
                    theme.ring
                  } ${theme.border} ${
                    isFieldDisabled ? "bg-gray-50 text-muted-500" : ""
                  } ${
                    formField.state.meta.errors.length > 0
                      ? "border-red-500"
                      : "border-gray-300"
                  }`}
                  placeholder={field.placeholder}
                />
              )}

              {formField.state.meta.errors.length > 0 && (
                <p className="mt-1 text-sm text-red-600">
                  {formField.state.meta.errors[0]}
                </p>
              )}

              {field.description && field.type !== "checkbox" && (
                <p className="mt-1 text-xs text-muted-500">
                  {field.description}
                </p>
              )}

              {isFieldDisabled && field.disabled?.(isEdit) && (
                <p className="mt-1 text-xs text-muted-500">
                  {field.name === "email"
                    ? "Email cannot be changed"
                    : field.name === "tenantUid"
                    ? "Tenant cannot be changed after user creation"
                    : "This field cannot be changed"}
                </p>
              )}
            </div>
          )}
        </form.Field>
      );
    };

    const renderSection = (sectionName: string, fields: FieldConfig[]) => {
      const visibleFields = fields.filter(
        (field) => !field.visible || field.visible(isEdit)
      );

      const sectionCustomContent = customContentBySection[sectionName] || {
        before: [],
        after: [],
      };

      if (
        visibleFields.length === 0 &&
        sectionCustomContent.before.length === 0 &&
        sectionCustomContent.after.length === 0
      ) {
        return null;
      }

      if (sectionName === "main") {
        return (
          <div key={sectionName} className="space-y-6">
            {renderCustomContent(sectionCustomContent.before)}
            {visibleFields.map(renderField)}
            {renderCustomContent(sectionCustomContent.after)}
          </div>
        );
      }

      return (
        <div key={sectionName} className="border-t pt-6">
          <h3 className="text-lg font-medium text-muted-800 mb-4">
            {sectionName.charAt(0).toUpperCase() + sectionName.slice(1)}{" "}
            Settings
          </h3>
          {renderCustomContent(sectionCustomContent.before)}
          <div className="space-y-4">{visibleFields.map(renderField)}</div>
          {renderCustomContent(sectionCustomContent.after)}
        </div>
      );
    };

    const entityTitle =
      config.getEntityTitle?.(entity) ||
      (entity as any)?.name ||
      (entity as any)?.email ||
      "";

    return (
      <div className="space-y-6">
        <ModalHeader
          title={
            isEdit
              ? `Edit ${config.entityName}: ${entityTitle}`
              : `Create New ${config.entityName}`
          }
          description={
            isEdit
              ? `Update ${config.entityName.toLowerCase()} information below`
              : `Fill in the details to create a new ${config.entityName.toLowerCase()}`
          }
        />

        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          className="space-y-6"
        >
          {Object.entries(fieldsBySection).map(([sectionName, fields]) =>
            renderSection(sectionName, fields)
          )}

          {/* Form Actions */}
          <div className="flex items-center justify-end space-x-4 pt-6 border-t">
            <Button
              type="button"
              onClick={onCancel}
              disabled={isFormLoading}
              variant="outline"
            >
              Cancel
            </Button>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
            >
              {([canSubmit, isSubmitting]) => (
                <Button
                  type="submit"
                  disabled={!canSubmit || isFormLoading || isSubmitting}
                  variant="default"
                >
                  {isSubmitting || isFormLoading
                    ? isEdit
                      ? "Updating..."
                      : "Creating..."
                    : isEdit
                    ? `Update ${config.entityName}`
                    : `Create ${config.entityName}`}
                </Button>
              )}
            </form.Subscribe>
          </div>
        </form>
      </div>
    );
  };

  return FormComponent;
}
