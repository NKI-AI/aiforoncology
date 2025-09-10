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
import { useUpdateSetting } from "@/api/hooks";
import { Setting, UpdateSettingRequest } from "@/api/models";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";

interface EditSettingFormProps {
  setting: Setting;
  onSuccess: () => void;
  onCancel: () => void;
}

export const EditSettingForm: React.FC<EditSettingFormProps> = ({
  setting,
  onSuccess,
  onCancel,
}) => {
  const [formData, setFormData] = useState<UpdateSettingRequest>({
    valueType: setting.valueType,
    value: setting.value,
  });

  const [validationErrors, setValidationErrors] = useState<
    Record<string, string>
  >({});

  const updateSettingMutation = useUpdateSetting(
    setting.tenantId,
    setting.key,
    {
      onSuccess: () => {
        toast.success("Setting updated successfully");
        onSuccess();
      },
      onError: (error: any) => {
        toast.error(error?.message || "Failed to update setting");
      },
    }
  );

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.value?.trim()) {
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
        ? formData.value!.toLowerCase()
        : formData.value!;

    updateSettingMutation.mutate({
      ...formData,
      value: finalValue,
    });
  };

  const handleValueTypeChange = (valueType: string) => {
    let newValue = formData.value || "";

    // Convert value when changing types
    if (
      valueType === "boolean" &&
      !["true", "false"].includes(newValue.toLowerCase())
    ) {
      newValue = "false";
    } else if (valueType === "number" && isNaN(Number(newValue))) {
      newValue = "0";
    } else if (
      valueType === "json" &&
      newValue &&
      formData.valueType !== "json"
    ) {
      try {
        newValue = JSON.stringify(newValue);
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
            value={formData.value || ""}
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
            value={formData.value || ""}
            onChange={(e) =>
              setFormData({ ...formData, value: e.target.value })
            }
          />
        );
    }
  };

  return (
    <div className="space-y-4">
      {/* Display read-only information */}
      <div className="space-y-4 p-4 bg-muted/50 rounded-lg">
        <div className="flex items-center justify-between">
          <div>
            <Label className="text-sm font-medium">Tenant ID</Label>
            <div className="flex items-center gap-2 mt-1">
              {setting.tenantId === 0 ? (
                <Badge
                  variant="outline"
                  className="bg-purple-50 text-purple-700 border-purple-200"
                >
                  Global
                </Badge>
              ) : (
                <span className="font-mono text-sm">{setting.tenantId}</span>
              )}
            </div>
          </div>
          <div>
            <Label className="text-sm font-medium">Key</Label>
            <div className="font-mono text-sm mt-1">{setting.key}</div>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
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
          <Button type="submit" disabled={updateSettingMutation.isPending}>
            {updateSettingMutation.isPending ? "Updating..." : "Update Setting"}
          </Button>
        </div>
      </form>
    </div>
  );
};
