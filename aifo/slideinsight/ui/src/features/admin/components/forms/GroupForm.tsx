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
import { Group } from "../../../../api/models";
import { CreateGroupRequest } from "../../../../types/groups";
import { useCreateGroup } from "../../hooks/useGroups";

interface GroupFormProps {
  group?: Group;
  onSubmit?: (group: Group) => void;
  onCancel?: () => void;
  isEditing?: boolean;
}

// Predefined common groups that users can quickly add
const COMMON_GROUPS = [
  {
    name: "administrators",
    description: "System administrators with full access",
  },
  { name: "editors", description: "Content editors and moderators" },
  {
    name: "reviewers",
    description: "Content reviewers and quality assurance team",
  },
  { name: "analysts", description: "Data analysts and researchers" },
  { name: "viewers", description: "Read-only users with viewing permissions" },
  { name: "contributors", description: "Users who can contribute content" },
  {
    name: "external_partners",
    description: "External partners and collaborators",
  },
  { name: "interns", description: "Interns and temporary users" },
];

export function GroupForm({
  group,
  onSubmit,
  onCancel,
  isEditing = false,
}: GroupFormProps) {
  const [formData, setFormData] = useState<CreateGroupRequest>({
    name: group?.name || "",
    description: group?.description || "",
  });
  const [selectedCommon, setSelectedCommon] = useState<CreateGroupRequest[]>(
    []
  );

  const createGroup = useCreateGroup();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (selectedCommon.length > 0) {
        // For now, create groups one by one (we could implement bulk creation later)
        for (const groupData of selectedCommon) {
          const newGroup = await createGroup.mutateAsync(groupData);
          if (selectedCommon.indexOf(groupData) === selectedCommon.length - 1) {
            // Call onSubmit with the last created group
            onSubmit?.(newGroup);
          }
        }
      } else if (formData.name && formData.description) {
        // Create single group
        const newGroup = await createGroup.mutateAsync(formData);
        onSubmit?.(newGroup);
      }
    } catch (error) {
      // Error handling is done in the mutation hooks
    }
  };

  const handleInputChange = (
    field: keyof CreateGroupRequest,
    value: string
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const toggleCommonGroup = (groupData: CreateGroupRequest) => {
    setSelectedCommon((prev) => {
      const exists = prev.find((g) => g.name === groupData.name);
      if (exists) {
        return prev.filter((g) => g.name !== groupData.name);
      } else {
        return [...prev, groupData];
      }
    });
  };

  const isLoading = createGroup.isPending;
  const canSubmit =
    (formData.name && formData.description) || selectedCommon.length > 0;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-medium">
            {isEditing ? "Edit Group" : "Create Group"}
          </h3>
          <p className="text-sm text-muted-foreground">
            {isEditing
              ? "Modify the group details below."
              : "Create a new group or select from common groups."}
          </p>
        </div>

        {!isEditing && (
          <>
            <div className="space-y-3">
              <Label className="text-sm font-medium">
                Quick Add Common Groups
              </Label>
              <p className="text-xs text-muted-foreground">
                Select one or more common groups to create them:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {COMMON_GROUPS.map((groupData) => {
                  const isSelected = selectedCommon.find(
                    (g) => g.name === groupData.name
                  );
                  return (
                    <label
                      key={groupData.name}
                      className={`flex items-start space-x-2 p-2 rounded-md cursor-pointer transition-colors ${
                        isSelected
                          ? "bg-primary/10 border border-primary/20"
                          : "hover:bg-muted/50"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={!!isSelected}
                        onChange={() => toggleCommonGroup(groupData)}
                        className="mt-1"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium">
                          {groupData.name}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {groupData.description}
                        </div>
                      </div>
                    </label>
                  );
                })}
              </div>
              {selectedCommon.length > 0 && (
                <p className="text-xs text-primary">
                  {selectedCommon.length} group(s) selected for creation
                </p>
              )}
            </div>

            <Separator />

            <div>
              <Label className="text-sm font-medium">
                Or Create Custom Group
              </Label>
              <p className="text-xs text-muted-foreground mb-3">
                Create a custom group with specific name and description:
              </p>
            </div>
          </>
        )}

        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Group Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              placeholder="e.g., administrators, editors, reviewers"
              disabled={isEditing || isLoading}
              required={selectedCommon.length === 0}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Use descriptive names like "administrators", "editors", or
              "reviewers"
            </p>
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange("description", e.target.value)}
              placeholder="Describe the purpose of this group and what kind of users belong to it..."
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
            ? `Create ${selectedCommon.length} Groups`
            : isEditing
            ? "Update Group"
            : "Create Group"}
        </Button>
      </div>
    </form>
  );
}
