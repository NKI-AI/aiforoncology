// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import * as React from "react";
import {
  ColumnDef,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  VisibilityState,
  OnChangeFn,
} from "@tanstack/react-table";
import { DataTableHeader } from "./DataTableHeader";
import { DataTableControls } from "./DataTableControls";
import { DataTableBody } from "./DataTableBody";
import { DataTableFooter } from "./DataTableFooter";
import { TableLayout } from "./TableLayout";

interface DataTableProps<TData> {
  data: TData[];
  columns: ColumnDef<TData>[];
  loading?: boolean;
  emptyMessage?: string;
  getRowId?: (row: TData, index: number) => string;

  // Header configuration
  title?: string;
  description?: string;
  addButtonText?: string;
  onAdd?: () => void;
  onRefresh?: () => void;

  // Controls configuration
  showControls?: boolean;
  showPageSizeSelector?: boolean;
  showColumnVisibility?: boolean;

  // Pagination props
  pagination?: boolean;
  pageSize?: number;
  currentPage?: number;
  totalPages?: number;
  totalItems?: number;
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];

  // Sorting props
  sorting?: SortingState;
  onSortingChange?: OnChangeFn<SortingState>;
  manualSorting?: boolean;

  // Row selection (optional)
  enableRowSelection?: boolean;

  // Row click handler (optional)
  onRowClick?: (row: TData) => void;

  // Column visibility (optional)
  enableColumnVisibility?: boolean;
  columnVisibility?: VisibilityState;
  onColumnVisibilityChange?: OnChangeFn<VisibilityState>;

  // Custom class names
  className?: string;
}

function DataTable<TData>({
  data,
  columns,
  loading = false,
  emptyMessage = "No results found.",
  getRowId,
  title,
  description,
  addButtonText,
  onAdd,
  onRefresh,
  showControls = false,
  showPageSizeSelector = true,
  showColumnVisibility = true,
  pagination = false,
  pageSize = 10,
  currentPage = 0,
  totalPages = 0,
  totalItems = 0,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50, 100],
  sorting,
  onSortingChange,
  manualSorting = false,
  enableRowSelection = false,
  onRowClick,
  enableColumnVisibility = false,
  columnVisibility,
  onColumnVisibilityChange,
  className = "",
}: DataTableProps<TData>) {
  const [rowSelection, setRowSelection] = React.useState({});
  const [internalColumnVisibility, setInternalColumnVisibility] =
    React.useState<VisibilityState>({});

  // Use external column visibility if provided, otherwise use internal
  const currentColumnVisibility = columnVisibility ?? internalColumnVisibility;
  const handleColumnVisibilityChange =
    onColumnVisibilityChange ?? setInternalColumnVisibility;

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility: enableColumnVisibility ? currentColumnVisibility : {},
      rowSelection: enableRowSelection ? rowSelection : {},
    },
    getRowId,
    enableRowSelection,
    onRowSelectionChange: enableRowSelection ? setRowSelection : undefined,
    onSortingChange,
    onColumnVisibilityChange: enableColumnVisibility
      ? handleColumnVisibilityChange
      : undefined,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: manualSorting ? undefined : getSortedRowModel(),
    manualPagination: true,
    pageCount: totalPages,
    manualSorting,
  });

  const header =
    title || description || onAdd ? (
      <DataTableHeader
        title={title}
        description={description}
        addButtonText={addButtonText}
        onAdd={onAdd}
        loading={loading}
      />
    ) : undefined;

  const controls = (
    <DataTableControls
      showControls={showControls}
      showPageSizeSelector={showPageSizeSelector}
      showColumnVisibility={showColumnVisibility}
      pagination={pagination}
      totalItems={totalItems}
      pageSize={pageSize}
      onPageSizeChange={onPageSizeChange}
      pageSizeOptions={pageSizeOptions}
      enableColumnVisibility={enableColumnVisibility}
      columns={columns}
      columnVisibility={currentColumnVisibility}
      onColumnVisibilityChange={handleColumnVisibilityChange}
    />
  );

  const body = (
    <DataTableBody
      table={table}
      columns={columns}
      emptyMessage={emptyMessage}
      onRowClick={onRowClick}
    />
  );

  const footer = (
    <DataTableFooter
      table={table}
      enableRowSelection={enableRowSelection}
      pagination={pagination}
      onPageChange={onPageChange}
      totalPages={totalPages}
      currentPage={currentPage}
    />
  );

  return (
    <TableLayout
      header={header}
      controls={controls}
      body={body}
      footer={footer}
      loading={loading}
      className={className}
    />
  );
}

// Legacy export for backward compatibility
type BaseDataTableProps<TData> = DataTableProps<TData>;

export default DataTable;
