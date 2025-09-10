// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useForm } from "@tanstack/react-form";
import { apiFetch } from "../../../../utils/fetchUtils";
import { type SlideWithCount } from "../../../../api";
import { Button } from "../../../../components/ui/button";

interface SlideFormData {
  name: string;
  description: string;
  metadata: string;
}

interface SlideFormProps {
  slide?: SlideWithCount | null; // null for create, SlideWithCount for edit
  entity?: SlideWithCount | null; // Alternative prop name used by AdminModal
  onSuccess: (slide: SlideWithCount) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

const SlideForm: React.FC<SlideFormProps> = ({
  slide,
  entity,
  onSuccess,
  onCancel,
  isLoading = false,
}) => {
  // Use entity prop if provided (from AdminModal), otherwise use slide prop
  const actualSlide = entity || slide;
  const isEdit = Boolean(actualSlide);

  const form = useForm({
    defaultValues: {
      name: actualSlide?.slideName ?? "",
      description: actualSlide?.metadata?.description ?? "",
      metadata: JSON.stringify(actualSlide?.metadata ?? {}, null, 2),
    },
    onSubmit: async ({ value }: { value: SlideFormData }) => {
      try {
        let result: SlideWithCount;

        if (isEdit && actualSlide) {
          // Update slide - only send fields that can be updated
          const updateData = {
            slideName: value.name,
            metadata: JSON.parse(value.metadata),
          };

          result = await apiFetch<SlideWithCount>(
            `/api/v1/slides/${actualSlide.slideUid}`,
            {
              method: "PUT",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify(updateData),
            }
          );
        } else {
          // Create slide - this might not be commonly used since slides are usually uploaded
          console.warn(
            "Creating slides via form is not typical - slides are usually uploaded"
          );
          return;
        }

        onSuccess(result);
      } catch (error) {
        console.error("Failed to save slide:", error);
        // TODO: Show error message to user
      }
    },
  });

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-muted-800">
          {isEdit
            ? `Edit Slide: ${actualSlide?.slideName}`
            : "Create New Slide"}
        </h2>
        <p className="text-sm text-muted-600 mt-1">
          {isEdit
            ? "Update slide information below"
            : "Slides are typically uploaded rather than created manually"}
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
        className="space-y-6"
      >
        {/* Slide Name */}
        <form.Field
          name="name"
          validators={{
            onChange: ({ value }) =>
              !value
                ? "Slide name is required"
                : value.length < 3
                ? "Slide name must be at least 3 characters"
                : undefined,
          }}
        >
          {(field) => (
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium text-muted-700 mb-2"
              >
                Slide Name <span className="text-red-500">*</span>
              </label>
              <input
                id="name"
                type="text"
                disabled={isLoading}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 ${
                  field.state.meta.errors.length > 0
                    ? "border-red-500"
                    : "border-gray-300"
                }`}
                placeholder="Enter slide name"
              />
              {field.state.meta.errors.length > 0 && (
                <p className="mt-1 text-sm text-red-600">
                  {field.state.meta.errors[0]}
                </p>
              )}
            </div>
          )}
        </form.Field>

        {/* Description */}
        <form.Field name="description">
          {(field) => (
            <div>
              <label
                htmlFor="description"
                className="block text-sm font-medium text-muted-700 mb-2"
              >
                Description
              </label>
              <textarea
                id="description"
                disabled={isLoading}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="Enter slide description"
              />
            </div>
          )}
        </form.Field>

        {/* Metadata */}
        <form.Field
          name="metadata"
          validators={{
            onChange: ({ value }) => {
              if (!value) return undefined;
              try {
                JSON.parse(value);
                return undefined;
              } catch {
                return "Metadata must be valid JSON";
              }
            },
          }}
        >
          {(field) => (
            <div>
              <label
                htmlFor="metadata"
                className="block text-sm font-medium text-muted-700 mb-2"
              >
                Metadata (JSON)
              </label>
              <textarea
                id="metadata"
                disabled={isLoading}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                rows={4}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm ${
                  field.state.meta.errors.length > 0
                    ? "border-red-500"
                    : "border-gray-300"
                }`}
                placeholder='{"key": "value"}'
              />
              {field.state.meta.errors.length > 0 && (
                <p className="mt-1 text-sm text-red-600">
                  {field.state.meta.errors[0]}
                </p>
              )}
              <p className="mt-1 text-xs text-muted-500">
                Enter valid JSON metadata for the slide
              </p>
            </div>
          )}
        </form.Field>

        {/* Form Actions */}
        <div className="flex items-center justify-end space-x-4 pt-6 border-t">
          <Button
            type="button"
            onClick={onCancel}
            disabled={isLoading}
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
                disabled={!canSubmit || isLoading || isSubmitting}
                variant="default"
              >
                {isSubmitting || isLoading
                  ? isEdit
                    ? "Updating..."
                    : "Creating..."
                  : isEdit
                  ? "Update Slide"
                  : "Create Slide"}
              </Button>
            )}
          </form.Subscribe>
        </div>
      </form>
    </div>
  );
};

export default SlideForm;
