// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import * as React from "react";
import { ColumnDef, VisibilityState, OnChangeFn } from "@tanstack/react-table";
import { ChevronDown } from "lucide-react";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

interface DataTableControlsProps<TData> {
  showControls?: boolean;
  showPageSizeSelector?: boolean;
  showColumnVisibility?: boolean;
  pagination?: boolean;
  totalItems?: number;
  pageSize?: number;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  enableColumnVisibility?: boolean;
  columns: ColumnDef<TData>[];
  columnVisibility?: VisibilityState;
  onColumnVisibilityChange?: OnChangeFn<VisibilityState>;
}

export function DataTableControls<TData>({
  showControls = false,
  showPageSizeSelector = true,
  showColumnVisibility = true,
  pagination = false,
  totalItems = 0,
  pageSize = 10,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50, 100],
  enableColumnVisibility = false,
  columns,
  columnVisibility,
  onColumnVisibilityChange,
}: DataTableControlsProps<TData>) {
  // Get column info for the dropdown - we need this to show the column names
  const getColumnDisplayName = (columnId: string) => {
    const column = columns.find(
      (col) => col.id === columnId || (col as any).accessorKey === columnId
    );

    if (column && typeof column.header === "string") {
      return column.header;
    }

    // Fallback to columnId with some formatting
    return columnId
      .replace(/([A-Z])/g, " $1")
      .replace(/^./, (str) => str.toUpperCase());
  };

  // Get hideable columns
  const hideableColumns = columns.filter((col) => {
    // Don't show selection column and actions column
    const columnId = col.id || ((col as any).accessorKey as string) || "";
    return (
      columnId !== "select" &&
      columnId !== "actions" &&
      col.enableHiding !== false
    );
  });

  if (!showControls) {
    return null;
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-2">
        {/* Combined showing info and page size selector */}
        {pagination && totalItems > 0 && (
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            {totalItems >= 10 && onPageSizeChange && showPageSizeSelector ? (
              <>
                <span>Showing</span>
                <Select
                  value={`${pageSize}`}
                  onValueChange={(value) => onPageSizeChange(Number(value))}
                >
                  <SelectTrigger className="h-8 w-[70px]">
                    <SelectValue placeholder={pageSize} />
                  </SelectTrigger>
                  <SelectContent side="top">
                    {pageSizeOptions.map((size) => (
                      <SelectItem key={size} value={`${size}`}>
                        {size}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <span>of {totalItems} items</span>
              </>
            ) : (
              <span>Showing {totalItems} items</span>
            )}
          </div>
        )}
      </div>

      {/* Column visibility toggle */}
      {showColumnVisibility &&
        enableColumnVisibility &&
        hideableColumns.length > 0 &&
        onColumnVisibilityChange && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3">
                Columns <ChevronDown className="ml-2 h-4 w-4" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {hideableColumns.map((column) => {
                const columnId =
                  column.id || ((column as any).accessorKey as string) || "";
                const isVisible = columnVisibility?.[columnId] !== false;

                return (
                  <DropdownMenuCheckboxItem
                    key={columnId}
                    className="capitalize"
                    checked={isVisible}
                    onCheckedChange={(value) =>
                      onColumnVisibilityChange((prev) => ({
                        ...prev,
                        [columnId]: !!value,
                      }))
                    }
                  >
                    {getColumnDisplayName(columnId)}
                  </DropdownMenuCheckboxItem>
                );
              })}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
    </div>
  );
}
