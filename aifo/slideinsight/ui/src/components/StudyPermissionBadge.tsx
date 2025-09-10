// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Shield, Database, XCircle, AlertCircle, Loader } from "lucide-react";
import { usePermissionExplanation } from "@/hooks/usePermissions";

interface StudyPermissionBadgeProps {
  studyUid: string;
  permission?: string;
  size?: "sm" | "default";
  showLoading?: boolean;
}

/**
 * Component that shows a permission badge for a study with a tooltip explanation
 * Useful for study listings and tables
 */
export function StudyPermissionBadge({
  studyUid,
  permission = "studies.view",
  size = "sm",
  showLoading = false,
}: StudyPermissionBadgeProps) {
  const {
    data: explanation,
    isLoading,
    error,
  } = usePermissionExplanation("study", studyUid, permission, {
    enabled: !!studyUid,
  });

  // Loading state
  if (isLoading && showLoading) {
    return (
      <Badge
        variant="outline"
        className={size === "sm" ? "text-xs px-2 py-0.5" : ""}
      >
        <Loader className="w-3 h-3 mr-1 animate-spin" />
        Checking...
      </Badge>
    );
  }

  // Error state
  if (error) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge
            variant="secondary"
            className={size === "sm" ? "text-xs px-2 py-0.5" : ""}
          >
            <AlertCircle className="w-3 h-3 mr-1" />
            Unknown
          </Badge>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-sm">
          <div className="space-y-1">
            <div className="font-medium text-amber-200">
              Unable to check permissions
            </div>
            <div className="text-xs opacity-90">
              {error instanceof Error
                ? error.message
                : "Failed to load permission status"}
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    );
  }

  // No data yet (but not loading) - don't show anything
  if (!explanation) {
    return null;
  }

  const getVariant = () => {
    if (explanation.hasAccess) {
      return explanation.grantType === "role_based_grant"
        ? "default"
        : "secondary";
    }
    return "destructive";
  };

  const getIcon = () => {
    if (explanation.hasAccess) {
      return explanation.grantType === "role_based_grant" ? (
        <Shield className="w-3 h-3 mr-1" />
      ) : (
        <Database className="w-3 h-3 mr-1" />
      );
    }
    return <XCircle className="w-3 h-3 mr-1" />;
  };

  const getBadgeText = () => {
    if (explanation.hasAccess) {
      switch (explanation.grantType) {
        case "role_based_grant":
          return "Role Access";
        case "direct_object_grant":
          return "Direct Access";
        case "inherited_grant":
          return "Inherited Access";
        default:
          return "Access";
      }
    }
    return "No Access";
  };

  const getTooltipContent = () => {
    if (explanation.hasAccess) {
      return (
        <div className="space-y-2">
          <div className="font-medium text-green-200">✓ Access Granted</div>
          <div className="text-sm opacity-90">{explanation.message}</div>
          {explanation.grantingResource && (
            <div className="text-xs opacity-75 border-t border-green-500/30 pt-2 mt-2">
              <strong>Granted by:</strong>{" "}
              {explanation.grantingResource.resourceName ||
                explanation.grantingResource.resourceType}
            </div>
          )}
          <div className="text-xs opacity-75">
            <strong>Permission:</strong> {explanation.permission}
          </div>
        </div>
      );
    } else {
      return (
        <div className="space-y-2">
          <div className="font-medium text-red-200">✗ Access Denied</div>
          <div className="text-sm opacity-90">{explanation.message}</div>
          <div className="text-xs opacity-75 border-t border-red-500/30 pt-2 mt-2">
            <strong>Permission:</strong> {explanation.permission}
          </div>
          <div className="text-xs opacity-75 mt-1">
            Contact an administrator if you need access to this study.
          </div>
        </div>
      );
    }
  };

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge
          variant={getVariant()}
          className={size === "sm" ? "text-xs px-2 py-0.5" : ""}
        >
          {getIcon()}
          {getBadgeText()}
        </Badge>
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-sm">
        {getTooltipContent()}
      </TooltipContent>
    </Tooltip>
  );
}

export default StudyPermissionBadge;
