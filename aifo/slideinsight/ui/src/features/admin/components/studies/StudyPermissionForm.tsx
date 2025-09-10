// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useCallback, useMemo, useEffect } from "react";
import { Button } from "../../../../components/ui/button";
import { Label } from "../../../../components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../../components/ui/select";
import { Checkbox } from "../../../../components/ui/checkbox";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../../components/ui/card";
import { useUsers } from "../../hooks/useUsers";
import {
  useCreateObjectGrant,
  useDeleteObjectGrant,
  useObjectGrants,
  useBulkCreateObjectGrants,
  useBulkDeleteObjectGrants,
} from "../../hooks/useObjectGrants";
import { usePermissions } from "../../hooks/usePermissions";
import { Badge } from "../../../../components/ui/badge";
import { User, Shield, CheckCircle, XCircle } from "lucide-react";
import { toast } from "sonner";

interface StudyPermissionFormProps {
  studyUid: string;
  onSubmit?: () => void;
  onCancel?: () => void;
  onSuccess?: () => void;
}

export default function StudyPermissionForm({
  studyUid,
  onSubmit,
  onCancel,
  onSuccess,
}: StudyPermissionFormProps) {
  const [selectedUserId, setSelectedUserId] = useState<string>("");
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(
    new Set()
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { users = [], loading: usersLoading } = useUsers({
    limit: 100,
    page: 1,
  });

  // Fetch all permissions from API and filter for study-related ones
  const { data: allPermissions = [], isLoading: permissionsLoading } =
    usePermissions();

  // Filter permissions to only show study-related ones
  const studyPermissions = useMemo(() => {
    return allPermissions
      .filter((permission) => permission.name.startsWith("studies."))
      .map((permission) => ({
        name: permission.name,
        description: permission.description || `Permission: ${permission.name}`,
      }));
  }, [allPermissions]);

  // Get existing grants for this study to show current user permissions
  const { data: existingGrants = [] } = useObjectGrants("study", studyUid);

  const createObjectGrant = useCreateObjectGrant();
  const deleteObjectGrant = useDeleteObjectGrant();
  const bulkCreateObjectGrants = useBulkCreateObjectGrants();
  const bulkDeleteObjectGrants = useBulkDeleteObjectGrants();

  // Get current permissions for the selected user
  const currentUserPermissions = useMemo(() => {
    if (!selectedUserId) return new Set<string>();

    const selectedUser = users.find((u) => u.userUid === selectedUserId);
    if (!selectedUser) return new Set<string>();

    // Find grants for this user using userUid
    const userGrants = existingGrants.filter(
      (grant) =>
        grant.grantee_type === "user" &&
        grant.grantee_info?.email === selectedUser.email
    );

    return new Set(userGrants.map((grant) => grant.permission));
  }, [selectedUserId, existingGrants, users]);

  // Update selected permissions when user changes or current permissions load
  useEffect(() => {
    setSelectedPermissions(new Set(currentUserPermissions));
  }, [currentUserPermissions]);

  const handleUserChange = useCallback((userId: string) => {
    setSelectedUserId(userId);
    // selectedPermissions will be updated by the useEffect above
  }, []);

  const handlePermissionToggle = useCallback((permission: string) => {
    setSelectedPermissions((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(permission)) {
        newSet.delete(permission);
      } else {
        newSet.add(permission);
      }
      return newSet;
    });
  }, []);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      if (!selectedUserId) {
        toast.error("Please select a user");
        return;
      }

      setIsSubmitting(true);

      try {
        const permissionsToAdd = Array.from(selectedPermissions).filter(
          (perm) => !currentUserPermissions.has(perm)
        );
        const permissionsToRemove = Array.from(currentUserPermissions).filter(
          (perm) => !selectedPermissions.has(perm)
        );

        // Handle additions using bulk create if multiple permissions
        const addPromise =
          permissionsToAdd.length > 0
            ? permissionsToAdd.length === 1
              ? createObjectGrant.mutateAsync({
                  grantee_type: "user",
                  grantee_uid: selectedUserId,
                  permission: permissionsToAdd[0],
                  resource_type: "study",
                  resource_uid: studyUid,
                })
              : bulkCreateObjectGrants.mutateAsync({
                  grantee_type: "user",
                  grantee_uid: selectedUserId,
                  permissions: permissionsToAdd,
                  resource_type: "study",
                  resource_uid: studyUid,
                })
            : Promise.resolve();

        // Handle removals using bulk delete if multiple permissions
        const selectedUser = users.find((u) => u.userUid === selectedUserId);
        const grantsToRemove = permissionsToRemove
          .map((permission) => {
            const grant = existingGrants.find(
              (g) =>
                g.grantee_type === "user" &&
                g.grantee_info?.email === selectedUser?.email &&
                g.permission === permission
            );
            return grant;
          })
          .filter(Boolean);

        const removePromise =
          grantsToRemove.length > 0
            ? grantsToRemove.length === 1
              ? deleteObjectGrant.mutateAsync({
                  resourceType: "study",
                  resourceId: studyUid,
                  data: {
                    grantee_type: grantsToRemove[0]!.grantee_type,
                    grantee_id: grantsToRemove[0]!.grantee_id,
                    permission: grantsToRemove[0]!.permission,
                  },
                })
              : bulkDeleteObjectGrants.mutateAsync({
                  resourceType: "study",
                  resourceId: studyUid,
                  grants: grantsToRemove.map((grant) => ({
                    grantee_type: grant!.grantee_type,
                    grantee_id: grant!.grantee_id,
                    permission: grant!.permission,
                  })),
                })
            : Promise.resolve();

        // Execute both operations
        await Promise.all([addPromise, removePromise]);

        const totalChanges =
          permissionsToAdd.length + permissionsToRemove.length;
        if (totalChanges > 0) {
          toast.success("Permissions updated!", {
            description: `${permissionsToAdd.length} granted, ${permissionsToRemove.length} revoked.`,
          });
        } else {
          toast.info("No changes made", {
            description: "The selected permissions were already up to date.",
          });
        }

        onSubmit?.();
        onSuccess?.();
      } catch (error) {
        console.error("Failed to update permissions:", error);
        toast.error("Failed to update permissions", {
          description:
            "Some permission changes may have failed. Please try again.",
        });
      } finally {
        setIsSubmitting(false);
      }
    },
    [
      selectedUserId,
      selectedPermissions,
      currentUserPermissions,
      studyUid,
      createObjectGrant,
      deleteObjectGrant,
      bulkCreateObjectGrants,
      bulkDeleteObjectGrants,
      existingGrants,
      users,
      onSubmit,
      onSuccess,
    ]
  );

  const selectedUser = useMemo(
    () => users.find((u) => u.userUid === selectedUserId),
    [users, selectedUserId]
  );

  const hasChanges = useMemo(() => {
    const current = Array.from(currentUserPermissions).sort();
    const selected = Array.from(selectedPermissions).sort();
    return JSON.stringify(current) !== JSON.stringify(selected);
  }, [currentUserPermissions, selectedPermissions]);

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-medium">Manage Study Permissions</h3>
          <p className="text-sm text-muted-foreground">
            Select a user and configure their permissions for this study.
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="user">Select User</Label>
          <Select
            value={selectedUserId}
            onValueChange={handleUserChange}
            disabled={isSubmitting || usersLoading}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={
                  usersLoading ? "Loading users..." : "Choose a user"
                }
              />
            </SelectTrigger>
            <SelectContent>
              {users.map((user) => (
                <SelectItem key={user.userUid} value={user.userUid}>
                  <div className="flex items-center space-x-2">
                    <User className="h-4 w-4" />
                    <div>
                      <div className="font-medium">
                        {user.firstName} {user.lastName}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {user.email}
                      </div>
                    </div>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {selectedUser && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Shield className="h-5 w-5" />
                <span>Permissions for {selectedUser.email}</span>
              </CardTitle>
              <CardDescription>
                Toggle the permissions this user should have for this study
                {permissionsLoading && " (Loading permissions...)"}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {permissionsLoading ? (
                <div className="text-center py-8">
                  <div className="animate-spin h-6 w-6 border-b-2 border-primary mx-auto mb-2"></div>
                  <p className="text-sm text-muted-foreground">
                    Loading available permissions...
                  </p>
                </div>
              ) : studyPermissions.length === 0 ? (
                <div className="text-center py-8">
                  <Shield className="h-12 w-12 text-muted-foreground mx-auto mb-2" />
                  <p className="text-sm text-muted-foreground">
                    No study permissions found.
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Contact an administrator to create study permissions.
                  </p>
                </div>
              ) : (
                studyPermissions.map((permission) => {
                  const isSelected = selectedPermissions.has(permission.name);
                  const wasOriginallySelected = currentUserPermissions.has(
                    permission.name
                  );
                  const isChanged = isSelected !== wasOriginallySelected;

                  return (
                    <div
                      key={permission.name}
                      className="flex items-start space-x-3 p-3 rounded-md border"
                    >
                      <Checkbox
                        id={permission.name}
                        checked={isSelected}
                        onCheckedChange={() =>
                          handlePermissionToggle(permission.name)
                        }
                        disabled={isSubmitting}
                      />
                      <div className="flex-1">
                        <label
                          htmlFor={permission.name}
                          className="text-sm font-medium cursor-pointer flex items-center space-x-2"
                        >
                          <span>{permission.name}</span>
                          {isChanged && (
                            <Badge
                              variant={isSelected ? "default" : "destructive"}
                              className="text-xs"
                            >
                              {isSelected ? (
                                <>
                                  <CheckCircle className="h-3 w-3 mr-1" />
                                  Will Grant
                                </>
                              ) : (
                                <>
                                  <XCircle className="h-3 w-3 mr-1" />
                                  Will Revoke
                                </>
                              )}
                            </Badge>
                          )}
                        </label>
                        <p className="text-xs text-muted-foreground mt-1">
                          {permission.description}
                        </p>
                      </div>
                    </div>
                  );
                })
              )}
            </CardContent>
          </Card>
        )}
      </div>

      <div className="flex justify-end space-x-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={
            !selectedUserId ||
            isSubmitting ||
            !hasChanges ||
            studyPermissions.length === 0
          }
        >
          {isSubmitting
            ? "Updating Permissions..."
            : hasChanges
            ? "Update Permissions"
            : "No Changes"}
        </Button>
      </div>
    </form>
  );
}
