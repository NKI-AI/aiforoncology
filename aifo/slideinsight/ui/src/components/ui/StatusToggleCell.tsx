// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import StatusDropdown, { StatusType, StatusOption } from "./StatusDropdown";
import { apiFetch } from "../../utils/fetchUtils";
import { toast } from "sonner";
import { Button } from "./button";
import { CheckIcon, CloseIcon } from "../icons";

// Configuration for boolean toggles
interface BooleanToggleConfig {
  type: "boolean";
  apiEndpoint: (entityId: string) => string;
  apiField: string;
  trueLabel?: string;
  falseLabel?: string;
  trueColor?: string;
  falseColor?: string;
  successMessage?: (entity: any, newValue: boolean) => string;
  errorMessage?: string;
  size?: "sm" | "md" | "lg";
  variant?: "toggle" | "badge";
}

// Configuration for status dropdowns
interface StatusDropdownConfig {
  type: "status";
  apiEndpoint: (entityId: string) => string;
  apiField: string;
  options?: StatusOption[];
  successMessage?: (entity: any, newStatus: StatusType) => string;
  errorMessage?: string;
  size?: "sm" | "md" | "lg";
}

type ToggleConfig = BooleanToggleConfig | StatusDropdownConfig;

interface StatusToggleCellProps<T extends Record<string, any>> {
  entity: T;
  config: ToggleConfig;
  onUpdate?: (entityId: string, newValue: boolean | StatusType) => void;
  getEntityId: (entity: T) => string;
  getEntityName?: (entity: T) => string;
  getCurrentValue: (entity: T) => boolean | StatusType;
}

function StatusToggleCell<T extends Record<string, any>>({
  entity,
  config,
  onUpdate,
  getEntityId,
  getEntityName,
  getCurrentValue,
}: StatusToggleCellProps<T>) {
  const [isUpdating, setIsUpdating] = useState(false);
  const [optimisticValue, setOptimisticValue] = useState<
    boolean | StatusType | null
  >(null);

  const entityId = getEntityId(entity);
  const entityName = getEntityName ? getEntityName(entity) : entityId;
  const currentValue =
    optimisticValue !== null ? optimisticValue : getCurrentValue(entity);

  const handleUpdate = async (newValue: boolean | StatusType) => {
    // Immediately update the UI for optimistic updates
    setOptimisticValue(newValue);

    // Notify parent about the optimistic change
    if (onUpdate) {
      onUpdate(entityId, newValue);
    }

    setIsUpdating(true);

    try {
      await apiFetch(config.apiEndpoint(entityId), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          [config.apiField]: newValue,
        }),
      });

      const successMessage = config.successMessage
        ? config.type === "boolean"
          ? (config as BooleanToggleConfig).successMessage?.(
              entity,
              newValue as boolean
            )
          : (config as StatusDropdownConfig).successMessage?.(
              entity,
              newValue as StatusType
            )
        : `Updated ${config.apiField} successfully!`;

      toast.success(successMessage, {
        description: `${entityName} has been updated.`,
      });

      // Keep the optimistic value since the API call succeeded
    } catch (error) {
      console.error(`Failed to update ${config.apiField}:`, error);

      // Revert on error
      setOptimisticValue(null);

      // Notify parent to revert the optimistic change
      if (onUpdate) {
        onUpdate(entityId, getCurrentValue(entity));
      }

      const errorMessage =
        config.errorMessage || `Failed to update ${config.apiField}`;
      toast.error(errorMessage, {
        description:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  const handleClick = (
    event: React.MouseEvent,
    newValue: boolean | StatusType
  ) => {
    event.stopPropagation();
    handleUpdate(newValue);
  };

  // Render boolean toggle
  if (config.type === "boolean") {
    const booleanConfig = config as BooleanToggleConfig;
    const value = currentValue as boolean;

    if (booleanConfig.variant === "toggle") {
      // Compact toggle button variant
      return (
        <Button
          variant="ghost"
          size="sm"
          onClick={(event) => handleClick(event, !value)}
          disabled={isUpdating}
          className={`h-6 w-6 p-0 rounded-full transition-colors ${
            value
              ? booleanConfig.trueColor ||
                "bg-green-100 hover:bg-green-200 text-green-700"
              : booleanConfig.falseColor ||
                "bg-gray-100 hover:bg-gray-200 text-muted-500"
          }`}
          title={`Click to ${value ? "disable" : "enable"}`}
        >
          {isUpdating ? (
            <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-current" />
          ) : value ? (
            <CheckIcon className="h-3 w-3" />
          ) : (
            <CloseIcon className="h-3 w-3" />
          )}
        </Button>
      );
    }

    // Badge variant
    const sizeClasses = {
      sm: "text-xs px-2 py-1",
      md: "text-sm px-3 py-1.5",
      lg: "text-base px-4 py-2",
    };

    return (
      <button
        onClick={(event) => handleClick(event, !value)}
        disabled={isUpdating}
        className={`inline-flex items-center rounded-full font-medium transition-colors hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed ${
          value
            ? booleanConfig.trueColor || "text-green-800 bg-green-100"
            : booleanConfig.falseColor || "text-muted-600 bg-gray-100"
        } ${sizeClasses[booleanConfig.size || "sm"]}`}
        title={`Click to ${value ? "disable" : "enable"}`}
      >
        {isUpdating && (
          <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-current mr-1" />
        )}
        {value
          ? booleanConfig.trueLabel || "Yes"
          : booleanConfig.falseLabel || "No"}
      </button>
    );
  }

  // Render status dropdown
  if (config.type === "status") {
    const statusConfig = config as StatusDropdownConfig;

    const handleStatusChange = async (newStatus: StatusType) => {
      await handleUpdate(newStatus);
    };

    return (
      <StatusDropdown
        currentStatus={currentValue as StatusType}
        onStatusChange={handleStatusChange}
        variant="badge"
        size={statusConfig.size || "sm"}
        loading={isUpdating}
        options={statusConfig.options}
      />
    );
  }

  return null;
}

export default StatusToggleCell;
