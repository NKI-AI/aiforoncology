// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useUserByUID, getFullName, type User } from "@/hooks/useUser";
import { Badge } from "@/components/ui/badge";
import { UserIcon } from "lucide-react";

interface UserCellProps {
  userUid: string;
  className?: string;
  showIcon?: boolean;
}

const UserCell: React.FC<UserCellProps> = ({
  userUid,
  className = "",
  showIcon = false,
}) => {
  const { data: user, isLoading, error } = useUserByUID(userUid);

  if (isLoading) {
    return (
      <div className={`space-y-1 ${className}`}>
        <div className="flex items-center space-x-2">
          {showIcon && (
            <div className="w-4 h-4 bg-muted rounded animate-pulse"></div>
          )}
          <div className="h-4 bg-muted rounded animate-pulse w-24"></div>
        </div>
        <div className="h-3 bg-muted/50 rounded animate-pulse w-20"></div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className={`space-y-1 ${className}`}>
        <div className="flex items-center space-x-2">
          {showIcon && <UserIcon className="w-4 h-4 text-muted-foreground" />}
          <div className="text-sm font-medium text-foreground">
            Unknown User
          </div>
        </div>
        <div className="text-xs text-muted-foreground">No email available</div>
      </div>
    );
  }

  return (
    <div className={`space-y-1 ${className}`}>
      <div className="flex items-center space-x-2">
        {showIcon && <UserIcon className="w-4 h-4 text-muted-foreground" />}
        <div className="flex items-center space-x-2">
          <span className="text-sm font-medium text-foreground">
            {getFullName(user)}
          </span>
          {!user.isActive && (
            <Badge variant="destructive" className="text-xs px-1.5 py-0">
              Inactive
            </Badge>
          )}
        </div>
      </div>
      <div className="text-xs text-muted-foreground">{user.email}</div>
    </div>
  );
};

export default UserCell;
