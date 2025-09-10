// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface PickerProps {
  buttonContent: React.ReactNode;
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  query: string;
  setQuery: (value: string) => void;
  suggestions: string[];
  selected: string[];
  onPick: (option: string) => void;
}

/**
 * Picker
 * Generic searchable popover picker for selecting string options from suggestions.
 */
export function Picker({
  buttonContent,
  isOpen,
  setIsOpen,
  query,
  setQuery,
  suggestions,
  selected,
  onPick,
}: PickerProps) {
  const filtered = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    return suggestions
      .filter((o) => !selected.includes(o))
      .filter((o) => (q ? o.toLowerCase().includes(q) : true));
  }, [suggestions, selected, query]);

  return (
    <div className="relative inline-block">
      <Button variant="outline" size="sm" onClick={() => setIsOpen(!isOpen)}>
        {buttonContent}
      </Button>
      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute z-50 mt-2 w-72 rounded-md border bg-popover text-popover-foreground shadow-md">
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
                    onClick={() => {
                      onPick(opt);
                      setIsOpen(false);
                      setQuery("");
                    }}
                    className="w-full text-left px-2 py-1.5 rounded hover:bg-accent hover:text-accent-foreground text-sm"
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
  );
}

export default Picker;
