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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  CheckCircle,
  XCircle,
  Shield,
  User,
  Database,
  Info,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import { useState } from "react";
import type { PermissionExplanation } from "@/hooks/usePermissions";

interface PermissionExplanationProps {
  explanation: PermissionExplanation;
  compact?: boolean;
  className?: string;
}

interface PermissionBadgeProps {
  explanation: PermissionExplanation;
  size?: "sm" | "default";
}

/**
 * Compact badge showing access status with tooltip explanation
 */
export function PermissionBadge({
  explanation,
  size = "default",
}: PermissionBadgeProps) {
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
          return "Access Granted";
      }
    }
    return "No Access";
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
      <TooltipContent side="bottom" className="max-w-sm">
        <div className="space-y-2">
          <div className="font-medium">
            {explanation.hasAccess ? "Access Granted" : "Access Denied"}
          </div>
          <div className="text-sm opacity-90">{explanation.message}</div>
          {explanation.grantingResource && (
            <div className="text-xs opacity-75 border-t pt-2 mt-2">
              Granted by:{" "}
              {explanation.grantingResource.resourceName ||
                explanation.grantingResource.resourceType}
            </div>
          )}
        </div>
      </TooltipContent>
    </Tooltip>
  );
}

/**
 * Detailed permission explanation component
 */
export function PermissionExplanationCard({
  explanation,
  compact = false,
  className = "",
}: PermissionExplanationProps) {
  const [isExpanded, setIsExpanded] = useState(!compact);

  const getStatusIcon = () => {
    return explanation.hasAccess ? (
      <CheckCircle className="w-5 h-5 text-green-600" />
    ) : (
      <XCircle className="w-5 h-5 text-red-600" />
    );
  };

  const getGrantTypeIcon = (grantType?: string) => {
    switch (grantType) {
      case "role_based_grant":
        return <Shield className="w-4 h-4 text-blue-500" />;
      case "direct_object_grant":
        return <Database className="w-4 h-4 text-green-500" />;
      case "inherited_grant":
        return <User className="w-4 h-4 text-purple-500" />;
      default:
        return <XCircle className="w-4 h-4 text-muted-400" />;
    }
  };

  const getCheckIcon = (result: boolean) => {
    return result ? (
      <CheckCircle className="w-4 h-4 text-green-500" />
    ) : (
      <XCircle className="w-4 h-4 text-muted-400" />
    );
  };

  return (
    <Card className={`${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            {getStatusIcon()}
            Permission Analysis
          </CardTitle>
          {compact && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="p-1 hover:bg-gray-100 rounded transition-colors"
            >
              {isExpanded ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
            </button>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Summary */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-start gap-3">
            {getGrantTypeIcon(explanation.grantType)}
            <div className="flex-1">
              <div className="font-medium text-sm mb-1">
                {explanation.hasAccess ? "Access Granted" : "Access Denied"}
              </div>
              <div className="text-sm text-muted-600">
                {explanation.message}
              </div>
            </div>
          </div>
        </div>

        {/* Details */}
        {(!compact || isExpanded) && (
          <div className="space-y-3">
            {/* Resource Info */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium text-muted-500">Resource:</span>
                <div className="font-mono text-xs bg-gray-100 px-2 py-1 rounded mt-1">
                  {explanation.resourceUid}
                </div>
              </div>
              <div>
                <span className="font-medium text-muted-500">Permission:</span>
                <div className="font-mono text-xs bg-gray-100 px-2 py-1 rounded mt-1">
                  {explanation.permission}
                </div>
              </div>
            </div>

            {/* Granting Resource */}
            {explanation.grantingResource && (
              <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                <div className="flex items-center gap-2 mb-2">
                  <Info className="w-4 h-4 text-green-600" />
                  <span className="font-medium text-green-800">Granted By</span>
                </div>
                <div className="text-sm text-green-700">
                  <div>
                    <strong>Type:</strong>{" "}
                    {explanation.grantingResource.resourceType}
                  </div>
                  {explanation.grantingResource.resourceName && (
                    <div>
                      <strong>Name:</strong>{" "}
                      {explanation.grantingResource.resourceName}
                    </div>
                  )}
                  {explanation.inheritancePath && (
                    <div>
                      <strong>Path:</strong> {explanation.inheritancePath}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Checks Performed */}
            <div>
              <h4 className="font-medium text-muted-700 mb-2 flex items-center gap-2">
                <span>Checks Performed</span>
                <Badge variant="outline" className="text-xs">
                  {explanation.checksPerformed.length}
                </Badge>
              </h4>
              <div className="space-y-2">
                {explanation.checksPerformed.map((check, index) => (
                  <div
                    key={index}
                    className={`flex items-start gap-3 p-3 rounded-lg border ${
                      check.result
                        ? "bg-green-50 border-green-200"
                        : "bg-gray-50 border-gray-200"
                    }`}
                  >
                    {getCheckIcon(check.result)}
                    <div className="flex-1">
                      <div className="text-sm font-medium mb-1">
                        {check.description}
                      </div>
                      {check.resourceName && (
                        <div className="text-xs text-muted-600">
                          Resource: {check.resourceName}
                        </div>
                      )}
                      {check.grantingEntity && check.result && (
                        <div className="text-xs text-green-700 mt-1">
                          ✓ Granted by{" "}
                          {check.grantingEntity.resourceName ||
                            check.grantingEntity.resourceType}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default PermissionExplanationCard;
