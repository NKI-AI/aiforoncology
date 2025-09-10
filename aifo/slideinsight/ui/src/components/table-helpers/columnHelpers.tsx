// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, MoreHorizontal, Eye } from "lucide-react";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import { Checkbox } from "../ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { CircleIcon } from "../icons";
import { formatDateShort } from "@/utils/format";
import UserCell from "../UserCell";

// Selection column helper
export function createSelectionColumn<T>(): ColumnDef<T> {
  return {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  };
}

// Title with description column helper
export function createTitleDescriptionColumn<T>(config: {
  accessor: keyof T;
  descriptionAccessor?: keyof T;
  header: string;
  sortable?: boolean;
  getTitle: (item: T) => string;
  getDescription?: (item: T) => string | null;
  maxDescriptionWidth?: string;
}): ColumnDef<T> {
  return {
    id: config.accessor as string,
    header: config.sortable
      ? ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 font-medium hover:bg-transparent"
          >
            {config.header}
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      : config.header,
    cell: ({ row }) => {
      const item = row.original;
      const title = config.getTitle(item);
      const description = config.getDescription?.(item);

      return (
        <div
          className="space-y-1.5 min-h-[3rem] flex flex-col justify-center"
          style={{ width: "320px" }}
        >
          <div className="font-medium text-foreground leading-tight">
            {title}
          </div>
          {description && (
            <div
              className="text-xs text-muted-foreground leading-relaxed line-clamp-3"
              title={description}
            >
              {description}
            </div>
          )}
        </div>
      );
    },
    enableSorting: config.sortable ?? true,
  };
}

// Status badge column helper
function createStatusColumn<T>(config: {
  accessor: keyof T;
  header: string;
  getStatusConfig: (status: string) => {
    className: string;
    icon: string;
    label?: string;
  };
}): ColumnDef<T> {
  return {
    accessorKey: config.accessor,
    header: config.header,
    cell: ({ getValue }) => {
      const status = getValue() as string;
      const statusConfig = config.getStatusConfig(status);

      return (
        <Badge variant="outline" className={statusConfig.className}>
          <CircleIcon className={`w-2 h-2 mr-1 ${statusConfig.icon}`} />
          {statusConfig.label ||
            (status
              ? status.charAt(0).toUpperCase() + status.slice(1)
              : "Unknown")}
        </Badge>
      );
    },
    enableSorting: false,
  };
}

// Date column helper
export function createDateColumn<T>(config: {
  accessor: keyof T;
  header: string;
  sortable?: boolean;
}): ColumnDef<T> {
  return {
    accessorKey: config.accessor,
    header: config.sortable
      ? ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 font-medium hover:bg-transparent"
          >
            {config.header}
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      : config.header,
    cell: ({ getValue }) => (
      <span className="text-sm text-muted-foreground">
        {getValue() ? formatDateShort(getValue() as string) : "Unknown"}
      </span>
    ),
    enableSorting: config.sortable ?? true,
  };
}

// ID column helper
export function createIdColumn<T>(config: {
  accessor: keyof T;
  header: string;
  maxLength?: number;
}): ColumnDef<T> {
  return {
    accessorKey: config.accessor,
    header: config.header,
    cell: ({ getValue }) => {
      const value = getValue() as string;
      const displayValue = config.maxLength
        ? value?.substring(0, config.maxLength)
        : value;
      return (
        <span className="font-mono text-xs bg-muted px-2 py-1 rounded">
          {displayValue || "N/A"}
        </span>
      );
    },
    enableSorting: false,
  };
}

// Actions column helper
export function createActionsColumn<T>(config: {
  onView?: (item: T) => void;
  onEdit?: (item: T) => void;
  customActions?: Array<{
    label: string;
    onClick: (item: T) => void;
    icon?: React.ComponentType<{ className?: string }>;
    variant?: "default" | "destructive";
  }>;
  entityName?: string;
}): ColumnDef<T> {
  return {
    id: "actions",
    header: "Actions",
    enableHiding: false,
    cell: ({ row }) => {
      const item = row.original;

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-8 w-8 p-0"
              onClick={(e) => e.stopPropagation()}
            >
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            className="w-[160px]"
            onClick={(e) => e.stopPropagation()}
          >
            <DropdownMenuLabel>Actions</DropdownMenuLabel>

            {config.onView && (
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  config.onView!(item);
                }}
                className="cursor-pointer"
              >
                <Eye className="mr-2 h-4 w-4" />
                View {config.entityName?.toLowerCase() || "item"}
              </DropdownMenuItem>
            )}

            {config.onEdit && (
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  config.onEdit!(item);
                }}
                className="cursor-pointer"
              >
                Edit {config.entityName?.toLowerCase() || "item"}
              </DropdownMenuItem>
            )}

            {(config.onView || config.onEdit) &&
              config.customActions &&
              config.customActions.length > 0 && <DropdownMenuSeparator />}

            {config.customActions?.map((action, index) => (
              <DropdownMenuItem
                key={index}
                onClick={(e) => {
                  e.stopPropagation();
                  action.onClick(item);
                }}
                className={`cursor-pointer ${
                  action.variant === "destructive"
                    ? "text-destructive hover:text-destructive/80 hover:bg-destructive/10"
                    : ""
                }`}
              >
                {action.icon && <action.icon className="mr-2 h-4 w-4" />}
                {action.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  };
}

// User column helper
export function createUserColumn<T>(config: {
  accessor: keyof T;
  header: string;
  sortable?: boolean;
  showIcon?: boolean;
}): ColumnDef<T> {
  return {
    id: config.accessor as string,
    header: config.sortable
      ? ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="h-auto p-0 font-medium hover:bg-transparent"
          >
            {config.header}
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      : config.header,
    cell: ({ row }) => {
      const item = row.original;
      const userUid = item[config.accessor] as string;

      if (!userUid) {
        return (
          <div className="space-y-1">
            <div className="text-sm font-medium text-muted-foreground">
              Unknown User
            </div>
            <div className="text-xs text-muted-foreground">
              No user assigned
            </div>
          </div>
        );
      }

      return <UserCell userUid={userUid} showIcon={config.showIcon} />;
    },
    enableSorting: config.sortable ?? false,
  };
}
