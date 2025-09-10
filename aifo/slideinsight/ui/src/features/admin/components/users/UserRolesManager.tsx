// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useMemo, useEffect } from "react";
import { User } from "../../../../api";
import {
  useAssignUsersToRole,
  useRemoveUserFromRole,
} from "../../hooks/useRoles";
import { useRoles } from "../../../../api";
import { apiFetch } from "../../../../utils/fetchUtils";
import { useQuery } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "../../../../components/ui/dialog";
import { Button } from "../../../../components/ui/button";
import { Badge } from "../../../../components/ui/badge";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Separator } from "../../../../components/ui/separator";
import { Checkbox } from "../../../../components/ui/checkbox";
import {
  SecurityIcon,
  TrashIcon,
  PlusIcon,
  UserIcon,
} from "../../../../components/icons";
import { toast } from "sonner";

interface UserRolesManagerProps {
  isOpen: boolean;
  onClose: () => void;
  user: User;
}

// Custom hook to fetch all role-user mappings
function useAllRoleUsers(allRoles: any[]) {
  return useQuery({
    queryKey: ["all-role-users", allRoles.map((r) => r.name).sort()],
    queryFn: async () => {
      if (allRoles.length === 0) return {};

      // Fetch users for all roles in parallel
      const promises = allRoles.map(async (role) => {
        try {
          const userUIDs = await apiFetch<string[]>(
            `/api/v1/roles/${role.name}/users`
          );
          return { roleName: role.name, userUIDs: userUIDs || [] };
        } catch (error) {
          console.error(`Failed to fetch users for role ${role.name}:`, error);
          return { roleName: role.name, userUIDs: [] };
        }
      });

      const results = await Promise.all(promises);

      // Convert to a map: roleName -> userUIDs[]
      const roleUserMap: Record<string, string[]> = {};
      results.forEach(({ roleName, userUIDs }) => {
        roleUserMap[roleName] = userUIDs;
      });

      return roleUserMap;
    },
    enabled: allRoles.length > 0,
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

export default function UserRolesManager({
  isOpen,
  onClose,
  user,
}: UserRolesManagerProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRoleNames, setSelectedRoleNames] = useState<string[]>([]);

  // Fetch all available roles using the API hook
  const { data: allRoles = [], loading: rolesLoading } = useRoles();

  // Fetch all role-user mappings
  const {
    data: roleUserMap = {},
    isLoading: roleUsersLoading,
    refetch: refetchRoleUsers,
  } = useAllRoleUsers(allRoles);

  // Mutations
  const assignUsers = useAssignUsersToRole();
  const removeUser = useRemoveUserFromRole();

  // Determine which roles the user currently has
  const userCurrentRoles = useMemo(() => {
    return allRoles.filter((role) => {
      const userUIDs = roleUserMap[role.name] || [];
      return userUIDs.includes(user.userUid);
    });
  }, [allRoles, roleUserMap, user.userUid]);

  // Get roles that are available to assign (not currently assigned)
  const availableRoles = useMemo(() => {
    const currentRoleNames = new Set(userCurrentRoles.map((r) => r.name));
    return allRoles.filter((role) => !currentRoleNames.has(role.name));
  }, [allRoles, userCurrentRoles]);

  // Filter available roles based on search term
  const filteredAvailableRoles = useMemo(() => {
    if (!searchTerm) return availableRoles;
    const term = searchTerm.toLowerCase();
    return availableRoles.filter(
      (role) =>
        role.name.toLowerCase().includes(term) ||
        role.description?.toLowerCase().includes(term)
    );
  }, [availableRoles, searchTerm]);

  const handleRoleToggle = (roleName: string, checked: boolean) => {
    setSelectedRoleNames((prev) => {
      if (checked) {
        return [...prev, roleName];
      } else {
        return prev.filter((name) => name !== roleName);
      }
    });
  };

  const handleAssignSelected = async () => {
    if (selectedRoleNames.length === 0) {
      toast.error("No roles selected");
      return;
    }

    // Assign user to each selected role
    const promises = selectedRoleNames.map((roleName) =>
      assignUsers.mutateAsync({
        roleName,
        userUIDs: [user.userUid],
      })
    );

    try {
      await Promise.all(promises);
      setSelectedRoleNames([]);
      setSearchTerm("");
      // Refetch all role-user mappings
      await refetchRoleUsers();
      toast.success(
        `Assigned ${selectedRoleNames.length} role${
          selectedRoleNames.length !== 1 ? "s" : ""
        } to user`
      );
    } catch (error) {
      // Error handled by mutation hooks
    }
  };

  const handleRemoveRole = async (roleName: string) => {
    try {
      await removeUser.mutateAsync({
        roleName,
        userUID: user.userUid,
      });
      // Refetch all role-user mappings
      await refetchRoleUsers();
    } catch (error) {
      // Error handled by mutation hook
    }
  };

  const isLoading = rolesLoading || roleUsersLoading;
  const isMutating = assignUsers.isPending || removeUser.isPending;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <SecurityIcon className="h-5 w-5 text-indigo-500" />
            <span>Manage Roles for {user.email}</span>
          </DialogTitle>
          <DialogDescription>
            Assign and remove roles for this user. The user will inherit all
            permissions from assigned roles.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden flex flex-col space-y-6">
          {/* Current Roles Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">
                Current Roles ({userCurrentRoles.length})
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
            ) : userCurrentRoles.length === 0 ? (
              <div className="text-center py-6 text-muted-foreground bg-muted/50 rounded-lg">
                <SecurityIcon className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
                <p>No roles assigned to this user</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                {userCurrentRoles.map((role) => (
                  <div
                    key={role.name}
                    className="flex items-center justify-between p-2 bg-background border rounded-md"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2">
                        <Badge variant="secondary" className="text-xs">
                          {role.name}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        {role.description}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemoveRole(role.name)}
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

          {/* Add Roles Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">
                Add Roles ({selectedRoleNames.length} selected)
              </h4>
              {selectedRoleNames.length > 0 && (
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
                <Label htmlFor="role-search">Search Available Roles</Label>
                <Input
                  id="role-search"
                  type="text"
                  placeholder="Search roles..."
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
              ) : filteredAvailableRoles.length === 0 ? (
                <div className="text-center py-6 text-muted-foreground bg-muted/50 rounded-lg">
                  {searchTerm ? (
                    <p>No roles found matching "{searchTerm}"</p>
                  ) : (
                    <p>All roles are already assigned to this user</p>
                  )}
                </div>
              ) : (
                <div className="space-y-2 max-h-64 overflow-y-auto border rounded-md p-3">
                  {filteredAvailableRoles.map((role) => {
                    const isSelected = selectedRoleNames.includes(role.name);
                    return (
                      <label
                        key={role.name}
                        className={`flex items-start space-x-3 p-2 rounded-md cursor-pointer transition-colors ${
                          isSelected
                            ? "bg-primary/10 border border-primary/20"
                            : "hover:bg-muted/50"
                        }`}
                      >
                        <Checkbox
                          checked={isSelected}
                          onCheckedChange={(checked) =>
                            handleRoleToggle(role.name, checked as boolean)
                          }
                          className="mt-1"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center space-x-2">
                            <span className="text-sm font-medium">
                              {role.name}
                            </span>
                            <Badge variant="outline" className="text-xs">
                              {role.short_uid}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground mt-1">
                            {role.description}
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
