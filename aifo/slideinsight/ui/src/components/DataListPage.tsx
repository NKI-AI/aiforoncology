// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { ColumnDef } from "@tanstack/react-table";
import { FilterField } from "../types/search";
import SearchBar from "./SearchBar";
import FilterPanel from "./FilterPanel";
import DataTable from "./tables/DataTable";
import ErrorStateAlert from "./ErrorStateAlert";
import PageHeader from "./PageHeader";
import PageContainer from "./PageContainer";

interface DataListPageProps<TData> {
  // Page header
  title?: string;
  subtitle?: string;

  // Data and state
  data: TData[];
  loading: boolean;
  error?: string | Error | null;

  // Table configuration
  columns: ColumnDef<TData>[];
  getRowId: (item: TData, index: number) => string;
  onRowClick?: (item: TData) => void;
  emptyMessage?: string;

  // Search and filters
  searchQuery: string;
  onSearchQueryChange: (query: string) => void;
  searchPlaceholder?: string;
  filterFields?: FilterField[];
  hasActiveFilters?: boolean;
  onClearFilters?: () => void;

  // Pagination
  pagination?: boolean;
  pageSize?: number;
  currentPage?: number;
  totalPages?: number;
  totalItems?: number;
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;

  // Container styling
  useContainer?: boolean;
  maxWidth?: string;
  className?: string;
}

function DataListPage<TData>({
  title,
  subtitle,
  data,
  loading,
  error,
  columns,
  getRowId,
  onRowClick,
  emptyMessage = "No items found.",
  searchQuery,
  onSearchQueryChange,
  searchPlaceholder = "Search...",
  filterFields = [],
  hasActiveFilters = false,
  onClearFilters,
  pagination = true,
  pageSize = 20,
  currentPage = 0,
  totalPages = 0,
  totalItems = 0,
  onPageChange,
  onPageSizeChange,
  useContainer = true,
  maxWidth = "max-w-7xl",
  className = "",
}: DataListPageProps<TData>) {
  const content = (
    <>
      {/* Header */}
      <PageHeader title={title} subtitle={subtitle} />

      {/* Error state */}
      <ErrorStateAlert error={error || null} variant="inline" />

      {/* Search and Filters */}
      <div className="mb-6 space-y-4">
        {/* Main Search Bar */}
        <SearchBar
          placeholder={searchPlaceholder}
          value={searchQuery}
          onChange={onSearchQueryChange}
        />

        {/* Filter Panel */}
        {filterFields.length > 0 && (
          <FilterPanel
            fields={filterFields}
            hasActiveFilters={hasActiveFilters}
            onClearFilters={onClearFilters}
          />
        )}
      </div>

      {/* Data Table */}
      <DataTable
        data={data}
        columns={columns}
        loading={loading}
        title=""
        description=""
        emptyMessage={emptyMessage}
        getRowId={getRowId}
        onRowClick={onRowClick}
        showControls={true}
        enableColumnVisibility={true}
        enableRowSelection={true}
        pagination={pagination}
        pageSize={pageSize}
        currentPage={currentPage}
        totalPages={totalPages}
        totalItems={totalItems}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </>
  );

  if (useContainer) {
    return (
      <PageContainer maxWidth={maxWidth} className={className}>
        {content}
      </PageContainer>
    );
  }

  return <div className={className}>{content}</div>;
}

export default DataListPage;
