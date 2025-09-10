// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useQueryClient } from "@tanstack/react-query";
import {
  useApiQuery,
  useApiMutation,
  createApiMutation,
} from "../../../utils/apiQueries";
import { queryKeys } from "../../../api/queryKeys";
import {
  // Use the new paginated hooks from api
  useGroups as useGroupsPaginated,
  useCreateGroup as useCreateGroupPaginated,
  useDeleteGroup as useDeleteGroupPaginated,
  type GroupQuery,
} from "../../../api";
import {
  Group,
  CreateGroupRequest,
  UserGroupAssignment,
} from "../../../types/groups";
import { toast } from "sonner";

// Hook to get all groups - now uses paginated API
export function useGroups(query: GroupQuery = {}) {
  return useGroupsPaginated(query);
}

// Hook to create a group - use the new paginated API
export function useCreateGroup() {
  return useCreateGroupPaginated({
    onSuccess: (newGroup) => {
      toast.success("Group created!", {
        description: `${newGroup.name} has been created successfully.`,
      });
    },
    onError: (error) => {
      console.error("Failed to create group:", error);
      toast.error("Failed to create group", {
        description: error?.message || "An unexpected error occurred",
      });
    },
  });
}

// Hook to delete a group - use the new paginated API
export function useDeleteGroup() {
  return useDeleteGroupPaginated({
    onSuccess: (_, groupName) => {
      toast.success("Group deleted!", {
        description: `${groupName} has been permanently deleted.`,
      });
    },
    onError: (error) => {
      console.error("Failed to delete group:", error);
      toast.error("Failed to delete group", {
        description: error?.message || "An unexpected error occurred",
      });
    },
  });
}

// Hook to get group users
export function useGroupUsers(groupName: string) {
  return useApiQuery<number[]>(
    queryKeys.groups.users(groupName),
    `/api/v1/groups/${groupName}/users`,
    {
      enabled: !!groupName,
      select: (data) => data || [],
      staleTime: 1000 * 60 * 5, // 5 minutes
    }
  );
}

// Hook to assign users to a group
export function useAssignUsersToGroup() {
  const queryClient = useQueryClient();

  return useApiMutation(
    ({ groupName, userUIDs }: { groupName: string; userUIDs: string[] }) =>
      createApiMutation.post<any, { user_uids: string[] }>(
        `/api/v1/groups/${groupName}/users`
      )({ user_uids: userUIDs }),
    {
      onSuccess: (response, { groupName }) => {
        // Invalidate group users
        queryClient.invalidateQueries({
          queryKey: queryKeys.groups.users(groupName),
        });

        const assignedCount = response.count || 0;
        const errorCount = response.error_count || 0;

        if (errorCount > 0) {
          toast.warning("User assignment completed with some errors", {
            description: `${assignedCount} users assigned, ${errorCount} failed.`,
          });
        } else {
          toast.success("Users assigned!", {
            description: `${assignedCount} users assigned to ${groupName}.`,
          });
        }
      },
      onError: (error) => {
        console.error("Failed to assign users to group:", error);
        toast.error("Failed to assign users", {
          description:
            error instanceof Error
              ? error.message
              : "An unexpected error occurred",
        });
      },
    }
  );
}

// Hook to remove a user from a group
export function useRemoveUserFromGroup() {
  const queryClient = useQueryClient();

  return useApiMutation(
    ({ groupName, userUID }: { groupName: string; userUID: string }) =>
      createApiMutation.delete<void>(
        `/api/v1/groups/${groupName}/users/${userUID}`
      )(),
    {
      onSuccess: (_, { groupName, userUID }) => {
        // Invalidate group users
        queryClient.invalidateQueries({
          queryKey: queryKeys.groups.users(groupName),
        });
        toast.success("User removed!", {
          description: `User removed from ${groupName}.`,
        });
      },
      onError: (error) => {
        console.error("Failed to remove user from group:", error);
        toast.error("Failed to remove user", {
          description:
            error instanceof Error
              ? error.message
              : "An unexpected error occurred",
        });
      },
    }
  );
}
