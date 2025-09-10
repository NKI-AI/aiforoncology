// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Input } from "@/components/ui/input";
import { PlusIcon } from "@heroicons/react/24/outline";

export interface CategoryPickerProps {
  /** The label for the add button (e.g., "Add Category", "Add Disease") */
  addLabel: string;
  /** Array of currently selected items */
  selected: string[];
  /** Function to update the selected items */
  setSelected: React.Dispatch<React.SetStateAction<string[]>>;
  /** Array of available suggestions */
  suggestions: string[];
  /** Current search query */
  query: string;
  /** Function to update the search query */
  setQuery: (value: string) => void;
  /** Whether the picker popover is open */
  isOpen: boolean;
  /** Function to control the picker popover state */
  setIsOpen: (open: boolean) => void;
  /** Function to generate CSS classes for pills (for styling different categories) */
  pillClassFor: (item: string) => string;
  /** Placeholder text when no items are selected */
  emptyText?: string;
}

/**
 * CategoryPicker
 * A reusable component that displays selected items as pills with remove buttons
 * and provides an inline "Add" button that opens a searchable picker popover.
 */
export function CategoryPicker({
  addLabel,
  selected,
  setSelected,
  suggestions,
  query,
  setQuery,
  isOpen,
  setIsOpen,
  pillClassFor,
  emptyText = "No items selected.",
}: CategoryPickerProps) {
  const filtered = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    return suggestions
      .filter((o) => !selected.includes(o))
      .filter((o) => (q ? o.toLowerCase().includes(q) : true))
      .sort((a, b) => a.localeCompare(b)); // Sort alphabetically
  }, [suggestions, selected, query]);

  const handleRemove = (item: string) => {
    setSelected((prev) => prev.filter((i) => i !== item));
  };

  const handleAdd = (item: string) => {
    setSelected((prev) => [...prev, item].sort((a, b) => a.localeCompare(b))); // Keep selected items sorted
    setIsOpen(false);
    setQuery("");
  };

  return (
    <div className="flex flex-wrap gap-2">
      {selected
        .sort((a, b) => a.localeCompare(b))
        .map((item) => (
          <span
            key={item}
            className={`inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-1 text-xs font-medium text-foreground hover:bg-accent hover:text-accent-foreground transition-colors h-[26px] ${pillClassFor(
              item
            )}`}
          >
            <span>{item}</span>
            <button
              type="button"
              className="text-muted-foreground hover:text-foreground transition-colors"
              onClick={() => handleRemove(item)}
              aria-label={`Remove ${item}`}
            >
              ×
            </button>
          </span>
        ))}
      <div className="relative inline-block">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-1 text-xs font-medium text-foreground hover:bg-accent hover:text-accent-foreground transition-colors h-[26px]"
        >
          <PlusIcon className="h-3 w-3 flex-shrink-0" />
          {addLabel}
        </button>
        {isOpen && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setIsOpen(false)}
            />
            <div className="absolute z-50 mt-2 w-72 rounded-md border border-border bg-popover text-popover-foreground shadow-lg">
              <div className="p-2">
                <Input
                  placeholder="Search..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="mb-2"
                  autoFocus
                />
                <div className="max-h-64 overflow-auto">
                  {filtered.map((opt) => (
                    <button
                      key={opt}
                      type="button"
                      onClick={() => handleAdd(opt)}
                      className="w-full text-left px-2 py-1.5 rounded text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
                    >
                      {opt}
                    </button>
                  ))}
                  {filtered.length === 0 && (
                    <div className="text-sm text-muted-foreground px-2 py-1.5">
                      No results
                    </div>
                  )}
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default CategoryPicker;
