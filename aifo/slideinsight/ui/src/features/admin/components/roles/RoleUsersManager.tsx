// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useMemo } from "react";
import { Role } from "../../../../api/models";
import { User } from "../../../../api/models";
import { useUsers } from "../../hooks/useUsers";
import {
  useRoleUsers,
  useAssignUsersToRole,
  useRemoveUserFromRole,
} from "../../hooks/useRoles";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../../components/ui/select";
import {
  UsersIcon,
  TrashIcon,
  PlusIcon,
  UserIcon,
} from "../../../../components/icons";
import { toast } from "sonner";

interface RoleUsersManagerProps {
  isOpen: boolean;
  onClose: () => void;
  role: Role;
}

export default function RoleUsersManager({
  isOpen,
  onClose,
  role,
}: RoleUsersManagerProps) {
  const [selectedUserUid, setSelectedUserUid] = useState<string>("");
  const [selectedUsersToAdd, setSelectedUsersToAdd] = useState<User[]>([]);
  const [searchQuery, setSearchQuery] = useState("");

  // Fetch role users (returns userUIDs now)
  const {
    data: roleUserUIDs = [],
    isLoading: roleUsersLoading,
    refetch: refetchRoleUsers,
  } = useRoleUsers(role.name);

  // Fetch all users for the picker
  const { users, loading: usersLoading } = useUsers({
    limit: 100, // Reasonable limit for user picker
    q: searchQuery,
  });

  // Mutations
  const assignUsers = useAssignUsersToRole();
  const removeUser = useRemoveUserFromRole();

  // Filter out users who are already in the role
  const availableUsers = useMemo(() => {
    return users.filter((user) => {
      return !roleUserUIDs.includes(user.userUid);
    });
  }, [users, roleUserUIDs]);

  // Get user details for current role members
  const roleUsers = useMemo(() => {
    return users.filter((user) => {
      return roleUserUIDs.includes(user.userUid);
    });
  }, [users, roleUserUIDs]);

  const handleUserSelect = (userUid: string) => {
    const user = availableUsers.find((u) => u.userUid === userUid);
    if (user && !selectedUsersToAdd.find((u) => u.userUid === userUid)) {
      setSelectedUsersToAdd((prev) => [...prev, user]);
      setSelectedUserUid(""); // Reset select
    }
  };

  const removeSelectedUser = (userUid: string) => {
    setSelectedUsersToAdd((prev) => prev.filter((u) => u.userUid !== userUid));
  };

  const handleAddUsers = async () => {
    if (selectedUsersToAdd.length === 0) {
      toast.error("Please select users to add");
      return;
    }

    // Use userUIDs directly instead of converting to numbers
    const userUIDs = selectedUsersToAdd.map((user) => user.userUid);

    try {
      await assignUsers.mutateAsync({
        roleName: role.name,
        userUIDs: userUIDs,
      });
      setSelectedUsersToAdd([]);
      refetchRoleUsers();
    } catch (error) {
      // Error handled by mutation hook
    }
  };

  const handleRemoveUser = async (userUid: string) => {
    try {
      await removeUser.mutateAsync({
        roleName: role.name,
        userUID: userUid,
      });
      refetchRoleUsers();
    } catch (error) {
      // Error handled by mutation hook
    }
  };

  const isLoading = roleUsersLoading || usersLoading;
  const isMutating = assignUsers.isPending || removeUser.isPending;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col">
        <DialogHeader className="flex-shrink-0">
          <DialogTitle className="flex items-center space-x-2">
            <UsersIcon className="h-5 w-5 text-blue-500" />
            <span>Manage Users in "{role.name}"</span>
          </DialogTitle>
          <DialogDescription>
            Add and remove users from this role. Users inherit role-based
            permissions and access.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto min-h-0">
          <div className="space-y-6 p-1">
            {/* Current Users Section */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium">
                  Current Users ({roleUserUIDs.length})
                </h4>
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
              ) : roleUsers.length === 0 ? (
                <div className="text-center py-6 text-muted-foreground bg-muted/50 rounded-lg">
                  <UsersIcon className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
                  <p>No users assigned to this role</p>
                </div>
              ) : (
                <div className="space-y-2 max-h-48 overflow-y-auto border rounded-md p-3">
                  {roleUsers.map((user) => (
                    <div
                      key={user.userUid}
                      className="flex items-center justify-between p-3 bg-background border rounded-md"
                    >
                      <div className="flex items-center space-x-3">
                        <UserIcon className="h-4 w-4 text-muted-foreground" />
                        <div className="flex-1">
                          <div className="flex items-center space-x-2">
                            <span className="font-medium">{user.email}</span>
                            <Badge
                              variant={user.isActive ? "default" : "secondary"}
                              className="text-xs"
                            >
                              {user.isActive ? "Active" : "Inactive"}
                            </Badge>
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {user.email} • {user.firstName} {user.lastName}
                          </div>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveUser(user.userUid)}
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

            {/* Add Users Section */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium">Add Users</h4>
              </div>

              {/* Search for users */}
              <div className="space-y-2">
                <Label htmlFor="user-search">Search Users</Label>
                <Input
                  id="user-search"
                  type="text"
                  placeholder="Search users by name or email..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  disabled={isMutating}
                />
              </div>

              {/* User selection */}
              <div className="space-y-2">
                <Label htmlFor="user-select">Select User to Add</Label>
                <Select
                  value={selectedUserUid}
                  onValueChange={handleUserSelect}
                  disabled={isMutating || usersLoading}
                >
                  <SelectTrigger>
                    <SelectValue
                      placeholder={
                        usersLoading
                          ? "Loading users..."
                          : "Choose a user to add"
                      }
                    />
                  </SelectTrigger>
                  <SelectContent>
                    {availableUsers.length === 0 ? (
                      <div className="p-2 text-sm text-muted-foreground text-center">
                        {usersLoading
                          ? "Loading..."
                          : "No available users found"}
                      </div>
                    ) : (
                      availableUsers.map((user) => (
                        <SelectItem key={user.userUid} value={user.userUid}>
                          <div className="flex items-center space-x-2">
                            <UserIcon className="h-4 w-4" />
                            <div>
                              <div className="font-medium">{user.email}</div>
                              <div className="text-xs text-muted-foreground">
                                {user.email} • {user.firstName} {user.lastName}
                              </div>
                            </div>
                          </div>
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>

              {/* Selected users to add */}
              {selectedUsersToAdd.length > 0 && (
                <div className="space-y-2">
                  <Label>
                    Selected Users to Add ({selectedUsersToAdd.length})
                  </Label>
                  <div className="space-y-1 max-h-24 overflow-y-auto border rounded-md p-2">
                    {selectedUsersToAdd.map((user) => (
                      <div
                        key={user.userUid}
                        className="flex items-center justify-between p-2 bg-blue-50 border border-blue-200 rounded-md"
                      >
                        <div className="flex items-center space-x-2">
                          <UserIcon className="h-4 w-4 text-blue-600" />
                          <div>
                            <span className="text-sm font-medium">
                              {user.firstName} {user.lastName}
                            </span>
                            <div className="text-xs text-muted-foreground">
                              {user.email}
                            </div>
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeSelectedUser(user.userUid)}
                          disabled={isMutating}
                          className="text-blue-600 hover:text-blue-700"
                        >
                          <TrashIcon className="h-3 w-3" />
                        </Button>
                      </div>
                    ))}
                  </div>
                  <Button
                    onClick={handleAddUsers}
                    disabled={isMutating || selectedUsersToAdd.length === 0}
                    className="w-full"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    Add {selectedUsersToAdd.length} User
                    {selectedUsersToAdd.length !== 1 ? "s" : ""}
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end space-x-2 pt-4 border-t bg-background flex-shrink-0">
          <Button variant="outline" onClick={onClose} disabled={isMutating}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
