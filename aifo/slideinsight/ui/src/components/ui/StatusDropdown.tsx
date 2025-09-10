// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import { Button } from "./button";
import { ChevronDownIcon } from "../icons";

export type StatusType = "active" | "inactive" | "suspended" | "pending";

export interface StatusOption {
  value: StatusType;
  label: string;
  color: string;
  bgColor: string;
  description?: string;
}

const DEFAULT_STATUS_OPTIONS: StatusOption[] = [
  {
    value: "active",
    label: "Active",
    color: "text-green-800",
    bgColor: "bg-green-100",
    description: "Entity is active and operational",
  },
  {
    value: "inactive",
    label: "Inactive",
    color: "text-muted-800",
    bgColor: "bg-gray-100",
    description: "Entity is inactive but can be reactivated",
  },
  {
    value: "suspended",
    label: "Suspended",
    color: "text-red-800",
    bgColor: "bg-red-100",
    description: "Entity is suspended due to policy violations",
  },
  {
    value: "pending",
    label: "Pending",
    color: "text-yellow-800",
    bgColor: "bg-yellow-100",
    description: "Entity is pending approval or setup",
  },
];

interface StatusDropdownProps {
  currentStatus: StatusType;
  onStatusChange: (newStatus: StatusType) => Promise<void> | void;
  disabled?: boolean;
  loading?: boolean;
  options?: StatusOption[];
  size?: "sm" | "md" | "lg";
  variant?: "badge" | "button";
  showDescription?: boolean;
}

const StatusDropdown: React.FC<StatusDropdownProps> = ({
  currentStatus,
  onStatusChange,
  disabled = false,
  loading = false,
  options = DEFAULT_STATUS_OPTIONS,
  size = "sm",
  variant = "badge",
  showDescription = true,
}) => {
  const currentOption =
    options.find((opt) => opt.value === currentStatus) || options[0];

  const handleStatusChange = async (newStatus: StatusType) => {
    if (disabled || loading || newStatus === currentStatus) return;
    await onStatusChange(newStatus);
  };

  const handleTriggerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };

  const handleMenuItemClick = (
    event: React.MouseEvent,
    newStatus: StatusType
  ) => {
    event.stopPropagation();
    handleStatusChange(newStatus);
  };

  const sizeClasses = {
    sm: "text-xs px-2 py-1",
    md: "text-sm px-3 py-1.5",
    lg: "text-base px-4 py-2",
  };

  if (variant === "badge") {
    return (
      <DropdownMenu>
        <DropdownMenuTrigger asChild disabled={disabled || loading}>
          <button
            onClick={handleTriggerClick}
            className={`inline-flex items-center rounded-full font-medium transition-colors hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed ${currentOption.color} ${currentOption.bgColor} ${sizeClasses[size]}`}
            disabled={disabled || loading}
          >
            {loading ? (
              <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-current mr-1" />
            ) : null}
            {currentOption.label}
            <ChevronDownIcon className="ml-1 h-3 w-3" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-48">
          {options.map((option) => (
            <DropdownMenuItem
              key={option.value}
              onClick={(event) => handleMenuItemClick(event, option.value)}
              className={`${
                option.value === currentStatus ? "bg-purple-50" : ""
              }`}
            >
              <div className="flex items-center justify-between w-full">
                <div className="flex items-center space-x-2">
                  <span
                    className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${option.color} ${option.bgColor}`}
                  >
                    {option.label}
                  </span>
                </div>
                {option.value === currentStatus && (
                  <div className="w-2 h-2 bg-purple-600 rounded-full" />
                )}
              </div>
              {showDescription && option.description && (
                <p className="text-xs text-muted-500 mt-1">
                  {option.description}
                </p>
              )}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }

  // Button variant
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild disabled={disabled || loading}>
        <button
          className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2 gap-1"
          disabled={disabled || loading}
          onClick={handleTriggerClick}
        >
          {loading ? (
            <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-current" />
          ) : null}
          <span
            className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${currentOption.color} ${currentOption.bgColor}`}
          >
            {currentOption.label}
          </span>
          <ChevronDownIcon className="h-4 w-4" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-48">
        {options.map((option) => (
          <DropdownMenuItem
            key={option.value}
            onClick={(event) => handleMenuItemClick(event, option.value)}
            className={`${
              option.value === currentStatus ? "bg-purple-50" : ""
            }`}
          >
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center space-x-2">
                <span
                  className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${option.color} ${option.bgColor}`}
                >
                  {option.label}
                </span>
              </div>
              {option.value === currentStatus && (
                <div className="w-2 h-2 bg-purple-600 rounded-full" />
              )}
            </div>
            {showDescription && option.description && (
              <p className="text-xs text-muted-500 mt-1">
                {option.description}
              </p>
            )}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default StatusDropdown;
