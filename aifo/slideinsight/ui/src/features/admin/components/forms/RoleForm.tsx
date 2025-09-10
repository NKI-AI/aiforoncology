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
import { Role, CreateRoleRequest } from "../../../../types/roles";
import { useCreateRole } from "../../hooks/useRoles";

interface RoleFormProps {
  role?: Role;
  onSubmit?: (role: Role) => void;
  onCancel?: () => void;
  isEditing?: boolean;
}

// Predefined common roles that users can quickly add
const COMMON_ROLES = [
  { name: "viewer", description: "Can view content but cannot make changes" },
  { name: "editor", description: "Can view and edit content" },
  {
    name: "moderator",
    description: "Can moderate content and manage basic settings",
  },
  { name: "admin", description: "Full administrative access to the system" },
  {
    name: "tenant_admin",
    description: "Administrative access within a specific tenant",
  },
  { name: "analyst", description: "Can view and analyze data" },
  { name: "contributor", description: "Can contribute content to the system" },
  { name: "reviewer", description: "Can review and approve content" },
];

export function RoleForm({
  role,
  onSubmit,
  onCancel,
  isEditing = false,
}: RoleFormProps) {
  const [formData, setFormData] = useState<CreateRoleRequest>({
    name: role?.name || "",
    description: role?.description || "",
  });
  const [selectedCommon, setSelectedCommon] = useState<CreateRoleRequest[]>([]);

  const createRole = useCreateRole();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (selectedCommon.length > 0) {
        // For now, create roles one by one (we could implement bulk creation later)
        for (const roleData of selectedCommon) {
          const newRole = await createRole.mutateAsync(roleData);
          if (selectedCommon.indexOf(roleData) === selectedCommon.length - 1) {
            // Call onSubmit with the last created role
            onSubmit?.(newRole);
          }
        }
      } else if (formData.name && formData.description) {
        // Create single role
        const newRole = await createRole.mutateAsync(formData);
        onSubmit?.(newRole);
      }
    } catch (error) {
      // Error handling is done in the mutation hooks
    }
  };

  const handleInputChange = (field: keyof CreateRoleRequest, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const toggleCommonRole = (roleData: CreateRoleRequest) => {
    setSelectedCommon((prev) => {
      const exists = prev.find((r) => r.name === roleData.name);
      if (exists) {
        return prev.filter((r) => r.name !== roleData.name);
      } else {
        return [...prev, roleData];
      }
    });
  };

  const isLoading = createRole.isPending;
  const canSubmit =
    (formData.name && formData.description) || selectedCommon.length > 0;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-medium">
            {isEditing ? "Edit Role" : "Create Role"}
          </h3>
          <p className="text-sm text-muted-foreground">
            {isEditing
              ? "Modify the role details below."
              : "Create a new role or select from common roles."}
          </p>
        </div>

        {!isEditing && (
          <>
            <div className="space-y-3">
              <Label className="text-sm font-medium">
                Quick Add Common Roles
              </Label>
              <p className="text-xs text-muted-foreground">
                Select one or more common roles to create them:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {COMMON_ROLES.map((roleData) => {
                  const isSelected = selectedCommon.find(
                    (r) => r.name === roleData.name
                  );
                  return (
                    <label
                      key={roleData.name}
                      className={`flex items-start space-x-2 p-2 rounded-md cursor-pointer transition-colors ${
                        isSelected
                          ? "bg-primary/10 border border-primary/20"
                          : "hover:bg-muted/50"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={!!isSelected}
                        onChange={() => toggleCommonRole(roleData)}
                        className="mt-1"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium">
                          {roleData.name}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {roleData.description}
                        </div>
                      </div>
                    </label>
                  );
                })}
              </div>
              {selectedCommon.length > 0 && (
                <p className="text-xs text-primary">
                  {selectedCommon.length} role(s) selected for creation
                </p>
              )}
            </div>

            <Separator />

            <div>
              <Label className="text-sm font-medium">
                Or Create Custom Role
              </Label>
              <p className="text-xs text-muted-foreground mb-3">
                Create a custom role with specific name and description:
              </p>
            </div>
          </>
        )}

        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Role Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              placeholder="e.g., admin, editor, viewer"
              disabled={isEditing || isLoading}
              required={selectedCommon.length === 0}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Use clear, descriptive names like "admin", "editor", or "viewer"
            </p>
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange("description", e.target.value)}
              placeholder="Describe what this role allows users to do..."
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
            ? `Create ${selectedCommon.length} Roles`
            : isEditing
            ? "Update Role"
            : "Create Role"}
        </Button>
      </div>
    </form>
  );
}
