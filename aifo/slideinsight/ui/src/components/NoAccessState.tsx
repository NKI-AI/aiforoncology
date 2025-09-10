// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { PermissionExplanationCard } from "@/components/PermissionExplanation";
import { ShieldX, Info } from "lucide-react";
import type { PermissionExplanation } from "@/hooks/usePermissions";

interface NoAccessStateProps {
  title?: string;
  description?: string;
  resourceId?: string;
  resourceType?: string;
  permissionExplanation?: PermissionExplanation | null;
  showBackToStudies?: boolean;
  backToStudiesPath?: string;
  className?: string;
}

const NoAccessState: React.FC<NoAccessStateProps> = ({
  title = "Access Denied",
  description = "You don't have permission to view this resource. Please contact an administrator if you believe you should have access to this content.",
  resourceId,
  resourceType = "resource",
  permissionExplanation,
  showBackToStudies = true,
  backToStudiesPath = "/studies",
  className = "",
}) => {
  const navigate = useNavigate();
  const [showPermissionDetails, setShowPermissionDetails] =
    React.useState(false);

  return (
    <div className={`text-center py-12 ${className}`}>
      <div className="mb-6">
        <ShieldX className="mx-auto h-16 w-16 text-red-400" />
      </div>

      <h3 className="text-lg font-semibold text-muted-900 mb-2">{title}</h3>

      <p className="text-muted-600 mb-6 max-w-md mx-auto">{description}</p>

      {/* Permission explanation for denied access */}
      {permissionExplanation && !permissionExplanation.hasAccess && (
        <div className="mb-6">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowPermissionDetails(!showPermissionDetails)}
            className="mb-4"
          >
            <Info className="w-4 h-4 mr-2" />
            {showPermissionDetails ? "Hide" : "Show"} Permission Details
          </Button>

          {showPermissionDetails && (
            <div className="max-w-2xl mx-auto">
              <PermissionExplanationCard
                explanation={permissionExplanation}
                compact={false}
              />
            </div>
          )}
        </div>
      )}

      <div className="flex flex-col sm:flex-row gap-3 justify-center">
        {showBackToStudies && (
          <Button
            onClick={() => navigate({ to: backToStudiesPath })}
            className="bg-blue-600 text-white hover:bg-blue-700"
          >
            Go to Studies
          </Button>
        )}
        <Button onClick={() => window.location.reload()} variant="outline">
          Try Again
        </Button>
      </div>

      {resourceId && (
        <div className="mt-6 text-sm text-muted-500">
          <p>
            {resourceType} ID: {resourceId}
          </p>
          <p className="mt-1">
            Need access? Contact your administrator and reference this{" "}
            {resourceType} ID.
          </p>
        </div>
      )}
    </div>
  );
};

export default NoAccessState;
