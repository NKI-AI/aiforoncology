// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useMemo } from "react";
import { Role } from "../../../../types/roles";
import { Permission } from "../../../../types/permissions";
import {
  useRolePermissions,
  useAssignPermissionsToRole,
  useRemovePermissionFromRole,
} from "../../hooks/useRoles";
import { usePermissions } from "../../hooks/usePermissions";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "../../../../components/ui/dialog";
import { Button } from "../../../../components/ui/button";
import { Badge } from "../../../../components/ui/badge";
import { Checkbox } from "../../../../components/ui/checkbox";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Separator } from "../../../../components/ui/separator";
import {
  SecurityIcon,
  TrashIcon,
  PlusIcon,
} from "../../../../components/icons";
import { toast } from "sonner";

interface RolePermissionsManagerProps {
  isOpen: boolean;
  onClose: () => void;
  role: Role;
}

export default function RolePermissionsManager({
  isOpen,
  onClose,
  role,
}: RolePermissionsManagerProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedPermissionIds, setSelectedPermissionIds] = useState<number[]>(
    []
  );

  // Fetch role permissions and all available permissions
  const {
    data: rolePermissions = [],
    isLoading: rolePermissionsLoading,
    refetch: refetchRolePermissions,
  } = useRolePermissions(role.name);

  const { data: allPermissions = [], isLoading: allPermissionsLoading } =
    usePermissions();

  // Mutations
  const assignPermissions = useAssignPermissionsToRole();
  const removePermission = useRemovePermissionFromRole();

  // Get permissions that are not yet assigned to this role
  const availablePermissions = useMemo(() => {
    const assignedPermissionIds = new Set(rolePermissions.map((p) => p.id));
    return allPermissions.filter((p) => !assignedPermissionIds.has(p.id));
  }, [allPermissions, rolePermissions]);

  // Filter available permissions based on search term
  const filteredAvailablePermissions = useMemo(() => {
    if (!searchTerm) return availablePermissions;
    const term = searchTerm.toLowerCase();
    return availablePermissions.filter(
      (p) =>
        p.name.toLowerCase().includes(term) ||
        p.description.toLowerCase().includes(term)
    );
  }, [availablePermissions, searchTerm]);

  const handlePermissionToggle = (permissionId: number, checked: boolean) => {
    setSelectedPermissionIds((prev) => {
      if (checked) {
        return [...prev, permissionId];
      } else {
        return prev.filter((id) => id !== permissionId);
      }
    });
  };

  const handleAssignSelected = async () => {
    if (selectedPermissionIds.length === 0) {
      toast.error("No permissions selected");
      return;
    }

    try {
      await assignPermissions.mutateAsync({
        roleName: role.name,
        permissionIDs: selectedPermissionIds,
      });
      setSelectedPermissionIds([]);
      setSearchTerm("");
      refetchRolePermissions();
    } catch (error) {
      // Error handled by mutation hook
    }
  };

  const handleRemovePermission = async (permission: Permission) => {
    try {
      await removePermission.mutateAsync({
        roleName: role.name,
        permissionID: permission.id,
      });
      refetchRolePermissions();
    } catch (error) {
      // Error handled by mutation hook
    }
  };

  const isLoading = rolePermissionsLoading || allPermissionsLoading;
  const isMutating = assignPermissions.isPending || removePermission.isPending;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <SecurityIcon className="h-5 w-5 text-purple-500" />
            <span>Manage Permissions for "{role.name}"</span>
          </DialogTitle>
          <DialogDescription>
            Assign and remove permissions for this role. Users with this role
            will inherit all assigned permissions.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden flex flex-col space-y-6">
          {/* Current Permissions Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">
                Current Permissions ({rolePermissions.length})
              </h4>
            </div>

            {isLoading ? (
              <div className="space-y-2">
                {[1, 2, 3].map((i) => (
                  <div
                    key={i}
                    className="animate-pulse bg-gray-100 h-10 rounded"
                  ></div>
                ))}
              </div>
            ) : rolePermissions.length === 0 ? (
              <div className="text-center py-6 text-muted-foreground bg-muted/50 rounded-lg">
                <SecurityIcon className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
                <p>No permissions assigned to this role</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {rolePermissions.map((permission) => (
                  <div
                    key={permission.id}
                    className="flex items-center justify-between p-2 bg-background border rounded-md"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2">
                        <Badge variant="secondary" className="text-xs">
                          {permission.name}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        {permission.description}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemovePermission(permission)}
                      disabled={isMutating}
                      className="text-red-600 hover:text-red-700 hover:bg-red-50"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <Separator />

          {/* Add Permissions Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">
                Add Permissions ({selectedPermissionIds.length} selected)
              </h4>
              {selectedPermissionIds.length > 0 && (
                <Button
                  onClick={handleAssignSelected}
                  disabled={isMutating}
                  size="sm"
                >
                  <PlusIcon className="h-4 w-4 mr-1" />
                  Assign Selected
                </Button>
              )}
            </div>

            <div className="space-y-3">
              <div>
                <Label htmlFor="permission-search">
                  Search Available Permissions
                </Label>
                <Input
                  id="permission-search"
                  type="text"
                  placeholder="Search permissions..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="mt-1"
                />
              </div>

              {isLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3].map((i) => (
                    <div
                      key={i}
                      className="animate-pulse bg-gray-100 h-12 rounded"
                    ></div>
                  ))}
                </div>
              ) : filteredAvailablePermissions.length === 0 ? (
                <div className="text-center py-6 text-muted-foreground bg-muted/50 rounded-lg">
                  {searchTerm ? (
                    <p>No permissions found matching "{searchTerm}"</p>
                  ) : (
                    <p>All permissions are already assigned to this role</p>
                  )}
                </div>
              ) : (
                <div className="space-y-2 max-h-64 overflow-y-auto border rounded-md p-3">
                  {filteredAvailablePermissions.map((permission) => {
                    const isSelected = selectedPermissionIds.includes(
                      permission.id
                    );
                    return (
                      <label
                        key={permission.id}
                        className={`flex items-start space-x-3 p-2 rounded-md cursor-pointer transition-colors ${
                          isSelected
                            ? "bg-primary/10 border border-primary/20"
                            : "hover:bg-muted/50"
                        }`}
                      >
                        <Checkbox
                          checked={isSelected}
                          onCheckedChange={(checked) =>
                            handlePermissionToggle(
                              permission.id,
                              checked as boolean
                            )
                          }
                          className="mt-1"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center space-x-2">
                            <span className="text-sm font-medium">
                              {permission.name}
                            </span>
                            <Badge variant="outline" className="text-xs">
                              {permission.short_uid}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground mt-1">
                            {permission.description}
                          </p>
                        </div>
                      </label>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end space-x-2 pt-4 border-t">
          <Button variant="outline" onClick={onClose} disabled={isMutating}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
