// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCreateSetting } from "@/api/hooks";
import { CreateSettingRequest } from "@/api/models";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";

interface CreateSettingFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

export const CreateSettingForm: React.FC<CreateSettingFormProps> = ({
  onSuccess,
  onCancel,
}) => {
  const [formData, setFormData] = useState<CreateSettingRequest>({
    tenantId: 0,
    key: "",
    valueType: "string",
    value: "",
  });

  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});

  const createSettingMutation = useCreateSetting({
    onSuccess: () => {
      toast.success("Setting created successfully");
      onSuccess();
    },
    onError: (error: any) => {
      toast.error(error?.message || "Failed to create setting");
    },
  });

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.key.trim()) {
      errors.key = "Key is required";
    } else if (!/^[a-zA-Z0-9_.]+$/.test(formData.key)) {
      errors.key =
        "Key can only contain letters, numbers, dots, and underscores";
    }

    if (!formData.value.trim()) {
      errors.value = "Value is required";
    } else if (formData.valueType === "json") {
      try {
        JSON.parse(formData.value);
      } catch {
        errors.value = "Invalid JSON format";
      }
    } else if (formData.valueType === "number") {
      if (isNaN(Number(formData.value))) {
        errors.value = "Value must be a valid number";
      }
    } else if (formData.valueType === "boolean") {
      if (!["true", "false"].includes(formData.value.toLowerCase())) {
        errors.value = "Value must be 'true' or 'false'";
      }
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    // Normalize boolean values
    const finalValue =
      formData.valueType === "boolean"
        ? formData.value.toLowerCase()
        : formData.value;

    createSettingMutation.mutate({
      ...formData,
      value: finalValue,
    });
  };

  const handleValueTypeChange = (valueType: string) => {
    let newValue = formData.value;

    // Convert value when changing types
    if (
      valueType === "boolean" &&
      !["true", "false"].includes(formData.value.toLowerCase())
    ) {
      newValue = "false";
    } else if (valueType === "number" && isNaN(Number(formData.value))) {
      newValue = "0";
    } else if (
      valueType === "json" &&
      formData.value &&
      formData.valueType !== "json"
    ) {
      try {
        newValue = JSON.stringify(formData.value);
      } catch {
        newValue = '""';
      }
    }

    setFormData({
      ...formData,
      valueType: valueType as "boolean" | "number" | "string" | "json",
      value: newValue,
    });
  };

  const renderValueInput = () => {
    switch (formData.valueType) {
      case "boolean":
        return (
          <div className="flex items-center space-x-2">
            <Checkbox
              id="booleanValue"
              checked={formData.value === "true"}
              onCheckedChange={(checked) =>
                setFormData({ ...formData, value: checked ? "true" : "false" })
              }
            />
            <Label htmlFor="booleanValue">
              {formData.value === "true" ? "True" : "False"}
            </Label>
          </div>
        );
      case "json":
        return (
          <Textarea
            placeholder="Enter valid JSON..."
            value={formData.value}
            onChange={(e) =>
              setFormData({ ...formData, value: e.target.value })
            }
            rows={6}
            className="font-mono text-sm"
          />
        );
      default:
        return (
          <Input
            type={formData.valueType === "number" ? "number" : "text"}
            placeholder={`Enter ${formData.valueType} value...`}
            value={formData.value}
            onChange={(e) =>
              setFormData({ ...formData, value: e.target.value })
            }
          />
        );
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="tenantId">Tenant ID</Label>
        <Input
          id="tenantId"
          type="number"
          placeholder="0 for global settings"
          value={formData.tenantId}
          onChange={(e) =>
            setFormData({
              ...formData,
              tenantId: parseInt(e.target.value) || 0,
            })
          }
        />
        <p className="text-sm text-muted-foreground">
          Use 0 for global settings that apply to all tenants
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="key">Key *</Label>
        <Input
          id="key"
          placeholder="e.g., enable_registration, max_upload_size"
          value={formData.key}
          onChange={(e) => setFormData({ ...formData, key: e.target.value })}
          className={validationErrors.key ? "border-red-500" : ""}
        />
        {validationErrors.key && (
          <p className="text-sm text-red-500">{validationErrors.key}</p>
        )}
        <p className="text-sm text-muted-foreground">
          Only letters, numbers, dots, and underscores allowed
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="valueType">Value Type *</Label>
        <Select
          value={formData.valueType}
          onValueChange={handleValueTypeChange}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="string">String</SelectItem>
            <SelectItem value="number">Number</SelectItem>
            <SelectItem value="boolean">Boolean</SelectItem>
            <SelectItem value="json">JSON</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="value">Value *</Label>
        {renderValueInput()}
        {validationErrors.value && (
          <p className="text-sm text-red-500">{validationErrors.value}</p>
        )}
      </div>

      <div className="flex justify-end gap-2 pt-4">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={createSettingMutation.isPending}>
          {createSettingMutation.isPending ? "Creating..." : "Create Setting"}
        </Button>
      </div>
    </form>
  );
};
