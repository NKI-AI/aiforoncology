// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import {
  PermissionBadge,
  PermissionExplanationCard,
} from "@/components/PermissionExplanation";
import { Info, Settings } from "lucide-react";
import type { Study } from "@/api/models";
import type { PermissionExplanation } from "@/hooks/usePermissions";

interface StudyCasesHeaderProps {
  study: Study | null;
  permissionExplanation: PermissionExplanation | null;
  showPermissionDetails: boolean;
  onTogglePermissionDetails: () => void;
}

const StudyCasesHeader: React.FC<StudyCasesHeaderProps> = ({
  study,
  permissionExplanation,
  showPermissionDetails,
  onTogglePermissionDetails,
}) => {
  const navigate = useNavigate();

  if (!study) return null;

  // Add explanation property by aliasing message for compatibility
  const permissionExplanationWithCompat = permissionExplanation
    ? {
        ...permissionExplanation,
        explanation: permissionExplanation.message,
      }
    : null;

  const handleSettingsClick = () => {
    navigate({ to: `/studies/${study.studyUid}/settings` });
  };

  return (
    <div className="mb-6 p-4 bg-muted/50 rounded-lg border">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-2">
            <h2 className="text-xl font-semibold text-foreground">
              {study.name}
            </h2>
            {permissionExplanationWithCompat && (
              <PermissionBadge
                explanation={permissionExplanationWithCompat}
                size="sm"
              />
            )}
          </div>
          {study.description && (
            <p className="text-muted-foreground text-sm mb-2">
              {study.description}
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Study ID: {study.studyUid}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleSettingsClick}
            className="text-muted-foreground hover:text-foreground"
          >
            <Settings className="w-4 h-4 mr-1" />
            Settings
          </Button>

          {permissionExplanationWithCompat &&
            permissionExplanationWithCompat.hasAccess && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onTogglePermissionDetails}
                className="text-muted-foreground hover:text-foreground"
              >
                <Info className="w-4 h-4 mr-1" />
                Access Info
              </Button>
            )}
        </div>
      </div>

      {/* Detailed permission explanation (expandable) */}
      {showPermissionDetails && permissionExplanationWithCompat && (
        <div className="mt-4 pt-4 border-t">
          <PermissionExplanationCard
            explanation={permissionExplanationWithCompat}
            compact={true}
            className="bg-background"
          />
        </div>
      )}
    </div>
  );
};

export default StudyCasesHeader;
