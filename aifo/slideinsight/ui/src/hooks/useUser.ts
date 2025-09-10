// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import { useApiQuery, queryKeys } from "../utils/apiQueries";

export interface User {
  tenantUid: string;
  userUid: string;
  email: string;
  firstName: string;
  lastName: string;
  mustResetPassword: boolean;
  isActive: boolean;
  emailVerified: boolean;
  createdAt: string;
  updatedAt: string;
}

// Hook to fetch a single user by UID
export function useUserByUID(userUid: string | null) {
  const queryKey = useMemo(
    () => (userUid ? queryKeys.users.detail(userUid) : null),
    [userUid]
  );

  const url = useMemo(
    () => (userUid ? `/api/v1/users/${userUid}` : null),
    [userUid]
  );

  const queryResult = useApiQuery<User>(
    queryKey || ["users", "detail", "disabled"],
    url || "",
    {
      enabled: !!userUid,
      staleTime: 1000 * 60 * 10, // 10 minutes - user data doesn't change often
    }
  );

  return {
    data: queryResult.data || null,
    isLoading: queryResult.isLoading,
    error: queryResult.error,
    refetch: queryResult.refetch,
  };
}

// Utility function to get full name from user
export function getFullName(user: User | null): string {
  if (!user) return "Unknown User";
  const fullName = [user.firstName, user.lastName].filter(Boolean).join(" ");
  return fullName || "Unknown User";
}
