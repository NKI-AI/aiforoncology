// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Button } from "../../../../components/ui/button";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Textarea } from "../../../../components/ui/textarea";
import { Separator } from "../../../../components/ui/separator";
import {
  Permission,
  CreatePermissionRequest,
} from "../../../../types/permissions";
import {
  useCreatePermission,
  useCreatePermissionsBulk,
} from "../../hooks/usePermissions";

interface PermissionFormProps {
  permission?: Permission;
  onSubmit?: (permission: Permission) => void;
  onCancel?: () => void;
  isEditing?: boolean;
}

// Predefined common permissions that users can quickly add
const COMMON_PERMISSIONS = [
  { name: "studies.view", description: "View a study" },
  { name: "studies.add_case", description: "Add an existing case to a study" },
  {
    name: "studies.modify_case",
    description: "Modify or remove cases in a study",
  },
  { name: "studies.annotate_case", description: "Annotate cases in a study" },
  { name: "cases.view", description: "View a case" },
  { name: "cases.edit", description: "Edit case details" },
  { name: "cases.delete", description: "Delete a case" },
  { name: "slides.view", description: "View slides" },
  { name: "slides.edit", description: "Edit slide details" },
  { name: "slides.delete", description: "Delete slides" },
  { name: "annotations.view", description: "View annotations" },
  { name: "annotations.create", description: "Create annotations" },
  { name: "annotations.edit", description: "Edit annotations" },
  { name: "annotations.delete", description: "Delete annotations" },
];

export function PermissionForm({
  permission,
  onSubmit,
  onCancel,
  isEditing = false,
}: PermissionFormProps) {
  const [formData, setFormData] = useState<CreatePermissionRequest>({
    name: permission?.name || "",
    description: permission?.description || "",
  });
  const [selectedCommon, setSelectedCommon] = useState<
    CreatePermissionRequest[]
  >([]);

  const createPermission = useCreatePermission();
  const createPermissionsBulk = useCreatePermissionsBulk();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (selectedCommon.length > 0) {
        // Bulk create selected common permissions
        const result = await createPermissionsBulk.mutateAsync(selectedCommon);
        onSubmit?.(result);
      } else if (formData.name && formData.description) {
        // Create single permission
        const newPermission = await createPermission.mutateAsync(formData);
        onSubmit?.(newPermission);
      }
    } catch (error) {
      // Error handling is done in the mutation hooks
    }
  };

  const handleInputChange = (
    field: keyof CreatePermissionRequest,
    value: string
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const toggleCommonPermission = (perm: CreatePermissionRequest) => {
    setSelectedCommon((prev) => {
      const exists = prev.find((p) => p.name === perm.name);
      if (exists) {
        return prev.filter((p) => p.name !== perm.name);
      } else {
        return [...prev, perm];
      }
    });
  };

  const isLoading =
    createPermission.isPending || createPermissionsBulk.isPending;
  const canSubmit =
    (formData.name && formData.description) || selectedCommon.length > 0;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-medium">
            {isEditing ? "Edit Permission" : "Create Permission"}
          </h3>
          <p className="text-sm text-muted-foreground">
            {isEditing
              ? "Modify the permission details below."
              : "Create a new permission or select from common permissions."}
          </p>
        </div>

        {!isEditing && (
          <>
            <div className="space-y-3">
              <Label className="text-sm font-medium">
                Quick Add Common Permissions
              </Label>
              <p className="text-xs text-muted-foreground">
                Select one or more common permissions to create them in bulk:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {COMMON_PERMISSIONS.map((perm) => {
                  const isSelected = selectedCommon.find(
                    (p) => p.name === perm.name
                  );
                  return (
                    <label
                      key={perm.name}
                      className={`flex items-start space-x-2 p-2 rounded-md cursor-pointer transition-colors ${
                        isSelected
                          ? "bg-primary/10 border border-primary/20"
                          : "hover:bg-muted/50"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={!!isSelected}
                        onChange={() => toggleCommonPermission(perm)}
                        className="mt-1"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium">{perm.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {perm.description}
                        </div>
                      </div>
                    </label>
                  );
                })}
              </div>
              {selectedCommon.length > 0 && (
                <p className="text-xs text-primary">
                  {selectedCommon.length} permission(s) selected for bulk
                  creation
                </p>
              )}
            </div>

            <Separator />

            <div>
              <Label className="text-sm font-medium">
                Or Create Custom Permission
              </Label>
              <p className="text-xs text-muted-foreground mb-3">
                Create a custom permission with specific name and description:
              </p>
            </div>
          </>
        )}

        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Permission Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              placeholder="e.g., studies.view or cases.edit"
              disabled={isEditing || isLoading}
              required={selectedCommon.length === 0}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Use dot notation like "resource.action" (e.g., studies.view,
              cases.edit)
            </p>
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange("description", e.target.value)}
              placeholder="Describe what this permission allows..."
              disabled={isEditing || isLoading}
              required={selectedCommon.length === 0}
              rows={3}
            />
          </div>
        </div>
      </div>

      <div className="flex justify-end space-x-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={!canSubmit || isLoading}>
          {isLoading
            ? "Creating..."
            : selectedCommon.length > 0
            ? `Create ${selectedCommon.length} Permissions`
            : isEditing
            ? "Update Permission"
            : "Create Permission"}
        </Button>
      </div>
    </form>
  );
}
