// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Input } from "@/components/ui/input";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";

export interface SlidePanelListProps<TItem> {
  /** Full collection to render inside the list */
  items: TItem[];
  /** Return a string used for client-side filtering */
  getFilterText?: (item: TItem) => string;
  /** Render function receiving the filtered items */
  children: (filteredItems: TItem[]) => React.ReactNode;
  /** Threshold above which a search input is shown */
  searchThreshold?: number;
  /** Placeholder for the search input */
  searchPlaceholder?: string;
  /** Optional label shown next to the select-all checkbox */
  selectAllLabel?: string;
  /** Provide a function to read the selected/checked state for each item */
  getItemChecked?: (item: TItem) => boolean;
  /** Toggle all filtered items to the given checked state */
  onToggleAll?: (filteredItems: TItem[], newChecked: boolean) => void;
  /** Optional className for the outer wrapper */
  className?: string;
  /** Optional: return a stable id for keyboard navigation */
  getItemId?: (item: TItem) => string;
  /** Optional: externally controlled active id */
  activeId?: string | null;
  /** Optional: notify when ArrowUp/Down changes the active item */
  onActiveIdChange?: (id: string) => void;
  /** Optional: toggle action for Space/Enter on the active item */
  onToggleActive?: (id: string) => void;
  /** Auto-focus the list on mount (default true) */
  autoFocus?: boolean;
  /** Auto-scroll the active row into view when it changes (default true) */
  autoScrollActive?: boolean;
}

/**
 * SlidePanelList renders a searchable, optionally bulk-selectable list area
 * designed to live inside SlidePanel.Body. It renders a header with a
 * conditional search field and select-all checkbox, and a scrollable content
 * area where callers render rows via a render-prop.
 */
export function SlidePanelList<TItem>({
  items,
  getFilterText,
  children,
  searchThreshold = 5,
  searchPlaceholder = "Filter...",
  selectAllLabel = "Select all",
  getItemChecked,
  onToggleAll,
  className,
  getItemId,
  activeId,
  onActiveIdChange,
  onToggleActive,
  autoFocus = true,
  autoScrollActive = true,
}: SlidePanelListProps<TItem>) {
  const [query, setQuery] = React.useState("");
  const containerRef = React.useRef<HTMLDivElement>(null);
  const scrollAreaRef = React.useRef<HTMLDivElement>(null);

  const showSearch =
    typeof getFilterText === "function" && items.length >= searchThreshold;

  const filteredItems = React.useMemo(() => {
    if (!showSearch || !query.trim()) return items;
    const q = query.toLowerCase();
    return items.filter((it) =>
      (getFilterText?.(it) || "").toLowerCase().includes(q)
    );
  }, [items, query, showSearch, getFilterText]);

  const selectAllEnabled =
    typeof getItemChecked === "function" && typeof onToggleAll === "function";
  const allVisible = React.useMemo(() => {
    if (!selectAllEnabled) return false;
    if (filteredItems.length === 0) return false;
    return filteredItems.every((it) => getItemChecked!(it));
  }, [filteredItems, selectAllEnabled, getItemChecked]);
  const someVisible = React.useMemo(() => {
    if (!selectAllEnabled) return false;
    return filteredItems.some((it) => getItemChecked!(it));
  }, [filteredItems, selectAllEnabled, getItemChecked]);

  // Auto-focus to allow Arrow navigation without extra clicks
  React.useEffect(() => {
    if (!autoFocus) return;
    const id = requestAnimationFrame(() => {
      containerRef.current?.focus();
    });
    return () => cancelAnimationFrame(id);
  }, [autoFocus]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (!getItemId || !onActiveIdChange) return;
    if (filteredItems.length === 0) return;
    const idList = filteredItems.map((it) => getItemId(it));
    const currentIndex = activeId ? idList.indexOf(activeId) : -1;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      const nextIndex =
        currentIndex < 0 ? 0 : (currentIndex + 1) % idList.length;
      onActiveIdChange(idList[nextIndex]);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      const prevIndex =
        currentIndex < 0
          ? idList.length - 1
          : (currentIndex - 1 + idList.length) % idList.length;
      onActiveIdChange(idList[prevIndex]);
    } else if (
      (e.key === " " || e.key === "Enter") &&
      onToggleActive &&
      activeId
    ) {
      e.preventDefault();
      onToggleActive(activeId);
    }
  };

  // Auto-scroll active item into view
  React.useEffect(() => {
    if (!autoScrollActive) return;
    if (!activeId) return;
    const scroller = scrollAreaRef.current;
    if (!scroller) return;
    const activeEl = scroller.querySelector(
      '[data-active="true"]'
    ) as HTMLElement | null;
    if (activeEl) {
      activeEl.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
        inline: "nearest",
      });
    }
  }, [activeId, autoScrollActive]); // Removed filteredItems dependency

  return (
    <div
      ref={containerRef}
      className={["flex-1 min-h-0 flex flex-col", className ?? ""].join(" ")}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      {/* Header */}
      <div className="px-3 pt-2 pb-1 flex-shrink-0">
        {showSearch ? (
          <div className="mb-1.5">
            <Input
              type="search"
              placeholder={searchPlaceholder}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full px-2 py-1 bg-input text-foreground placeholder-muted-foreground text-xs border-border focus:border-primary"
            />
          </div>
        ) : null}

        {/* Select all + count row */}
        {selectAllEnabled ? (
          <div className="mb-2 flex items-center justify-between px-1.5 py-1 bg-muted rounded">
            <span className="text-xs font-medium text-muted-foreground">
              {filteredItems.length} item{filteredItems.length !== 1 ? "s" : ""}
            </span>
            <div className="inline-flex items-center space-x-2">
              <span className="text-xs text-muted-foreground">
                {selectAllLabel}
              </span>
              <button
                type="button"
                className="p-1 rounded hover:bg-accent text-muted-foreground"
                title={allVisible ? "Hide all" : "Show all"}
                aria-label={allVisible ? "Hide all" : "Show all"}
                onClick={() => onToggleAll?.(filteredItems, !allVisible)}
              >
                {allVisible || someVisible ? (
                  <EyeIcon className="h-4 w-4" />
                ) : (
                  <EyeSlashIcon className="h-4 w-4" />
                )}
              </button>
            </div>
          </div>
        ) : (
          <div className="mb-2 flex items-center justify-between px-1.5 py-1 bg-muted rounded">
            <span className="text-xs font-medium text-muted-foreground">
              {filteredItems.length} item{filteredItems.length !== 1 ? "s" : ""}
            </span>
          </div>
        )}
      </div>

      {/* Scrollable content */}
      <div ref={scrollAreaRef} className="px-3 flex-1 overflow-y-auto min-h-0">
        {children(filteredItems)}
      </div>
    </div>
  );
}

export interface SlidePanelListItemProps {
  active?: boolean;
  leftColor?: string;
  primary: React.ReactNode;
  secondary?: React.ReactNode;
  onClick?: () => void;
  rightSlot?: React.ReactNode;
  dataActive?: boolean;
}

export const SlidePanelListItem = React.forwardRef<
  HTMLLIElement,
  SlidePanelListItemProps
>(
  (
    {
      active = false,
      leftColor,
      primary,
      secondary,
      onClick,
      rightSlot,
      dataActive,
    },
    ref
  ) => {
    return (
      <li
        ref={ref}
        className={[
          "flex items-center justify-between px-1.5 py-1 rounded cursor-pointer transition-colors",
          active
            ? "bg-primary text-primary-foreground hover:bg-primary/90"
            : "hover:bg-accent",
        ].join(" ")}
        data-active={dataActive ? true : undefined}
        onClick={onClick}
      >
        <div className="flex items-center space-x-2 min-w-0 flex-1">
          {leftColor ? (
            <span
              className="h-2.5 w-2.5 block rounded-sm flex-shrink-0"
              style={{ backgroundColor: leftColor }}
            />
          ) : null}
          <span
            className={[
              "text-sm truncate",
              active
                ? "text-primary-foreground font-medium"
                : "text-foreground",
            ].join(" ")}
          >
            {primary}
          </span>
          {secondary !== undefined && (
            <small
              className={
                active ? "text-primary-foreground/80" : "text-muted-foreground"
              }
            >
              {typeof secondary === "string" ? `(${secondary})` : secondary}
            </small>
          )}
        </div>
        {rightSlot}
      </li>
    );
  }
);

SlidePanelListItem.displayName = "SlidePanelListItem";

export default Object.assign(SlidePanelList, { Item: SlidePanelListItem });
