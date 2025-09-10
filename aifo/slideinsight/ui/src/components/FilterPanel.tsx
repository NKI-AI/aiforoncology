// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Filter } from "lucide-react";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { FilterField } from "../types/search";

interface FilterPanelProps {
  fields: FilterField[];
  hasActiveFilters?: boolean;
  onClearFilters?: () => void;
}

const FilterPanel: React.FC<FilterPanelProps> = ({
  fields,
  hasActiveFilters = false,
  onClearFilters,
}) => {
  const [showFilters, setShowFilters] = useState(false);

  if (fields.length === 0) return null;

  const activeFilterCount = fields.filter((field) => field.value).length;

  return (
    <div className="space-y-4">
      {/* Filter Toggle and Clear Button */}
      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          size="default"
          onClick={() => setShowFilters(!showFilters)}
          className="gap-2 h-12 px-4"
        >
          <Filter className="h-4 w-4" />
          Filters
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
              {activeFilterCount}
            </Badge>
          )}
        </Button>

        {hasActiveFilters && onClearFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onClearFilters}
            className="text-muted-foreground hover:text-foreground"
          >
            Clear all filters
          </Button>
        )}
      </div>

      {/* Advanced Filters Panel */}
      {showFilters && (
        <div className="border rounded-lg p-4 bg-muted/50">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {fields.map((field) => (
              <div key={field.key} className="space-y-2">
                <label className="text-sm font-medium text-foreground">
                  {field.label}
                </label>
                {field.type === "text" ? (
                  <Input
                    placeholder={field.placeholder}
                    value={field.value}
                    onChange={(e) => field.onChange(e.target.value)}
                    className="h-9"
                  />
                ) : field.type === "checkbox" ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id={field.key}
                      checked={field.value || false}
                      onChange={(e) => field.onChange(e.target.checked)}
                      className="h-4 w-4 text-primary border rounded focus:ring-primary"
                    />
                    {field.description && (
                      <label
                        htmlFor={field.key}
                        className="text-sm text-muted-foreground"
                      >
                        {field.description}
                      </label>
                    )}
                  </div>
                ) : (
                  <select
                    value={field.value}
                    onChange={(e) => field.onChange(e.target.value)}
                    className="w-full h-9 px-3 text-sm border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                  >
                    <option value="">{field.placeholder}</option>
                    {field.options?.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default FilterPanel;
