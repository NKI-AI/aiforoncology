// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import * as React from "react";
import { Table } from "@tanstack/react-table";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "../pagination";

interface DataTableFooterProps<TData> {
  table: Table<TData>;
  enableRowSelection?: boolean;
  pagination?: boolean;
  onPageChange?: (page: number) => void;
  totalPages?: number;
  currentPage?: number;
}

export function DataTableFooter<TData>({
  table,
  enableRowSelection = false,
  pagination = false,
  onPageChange,
  totalPages = 0,
  currentPage = 0,
}: DataTableFooterProps<TData>) {
  // Generate page numbers for pagination
  const getVisiblePages = () => {
    if (!pagination || !onPageChange || totalPages <= 1) return [];

    const delta = 1;
    const rangeWithDots: (number | "ellipsis-start" | "ellipsis-end")[] = [];

    if (currentPage > delta + 1) {
      rangeWithDots.push(0);
      if (currentPage > delta + 2) {
        rangeWithDots.push("ellipsis-start");
      }
    }

    for (
      let i = Math.max(0, currentPage - delta);
      i <= Math.min(totalPages - 1, currentPage + delta);
      i++
    ) {
      rangeWithDots.push(i);
    }

    if (currentPage < totalPages - delta - 2) {
      if (currentPage < totalPages - delta - 3) {
        rangeWithDots.push("ellipsis-end");
      }
      rangeWithDots.push(totalPages - 1);
    }

    return rangeWithDots;
  };

  const visiblePages = getVisiblePages();
  const hasSelectionInfo =
    enableRowSelection && table.getFilteredSelectedRowModel().rows.length > 0;
  const hasPagination = pagination && onPageChange && totalPages > 1;

  if (!hasSelectionInfo && !hasPagination) {
    return null;
  }

  return (
    <div className="flex items-center justify-between space-x-2 py-4">
      <div className="flex-1 text-sm text-muted-foreground">
        {hasSelectionInfo && (
          <>
            {table.getFilteredSelectedRowModel().rows.length} of{" "}
            {table.getRowModel().rows.length} row(s) selected.
          </>
        )}
      </div>

      {/* Pagination controls */}
      {hasPagination && (
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                size="default"
                onClick={() => onPageChange(currentPage - 1)}
                className={
                  currentPage === 0
                    ? "pointer-events-none opacity-50"
                    : "cursor-pointer"
                }
              />
            </PaginationItem>

            {visiblePages.map((page, index) => {
              if (page === "ellipsis-start" || page === "ellipsis-end") {
                return (
                  <PaginationItem key={`ellipsis-${index}`}>
                    <PaginationEllipsis />
                  </PaginationItem>
                );
              }

              const pageNumber = page as number;
              return (
                <PaginationItem key={pageNumber}>
                  <PaginationLink
                    size="default"
                    onClick={() => onPageChange(pageNumber)}
                    isActive={currentPage === pageNumber}
                    className="cursor-pointer"
                  >
                    {pageNumber + 1}
                  </PaginationLink>
                </PaginationItem>
              );
            })}

            <PaginationItem>
              <PaginationNext
                size="default"
                onClick={() => onPageChange(currentPage + 1)}
                className={
                  currentPage >= totalPages - 1
                    ? "pointer-events-none opacity-50"
                    : "cursor-pointer"
                }
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      )}
    </div>
  );
}
