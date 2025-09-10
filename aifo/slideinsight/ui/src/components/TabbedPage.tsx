// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { Children, isValidElement } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export interface TabbedPagePageProps {
  value: string;
  label: string;
  children: React.ReactNode;
}

export const TabbedPagePage: React.FC<TabbedPagePageProps> = ({ children }) => {
  return <>{children}</>;
};
TabbedPagePage.displayName = "TabbedPagePage";

export interface TabbedPageProps {
  title?: string;
  subtitle?: string;
  activeValue: string;
  onValueChange: (value: string) => void;
  leftActions?: React.ReactNode;
  rightActions?: React.ReactNode;
  /** keeps the title + tabs visible while scrolling */
  stickyHeader?: boolean;
  /** optional extra classes for the content wrapper */
  contentClassName?: string;
  /** children should be <TabbedPagePage value="..." label="...">...</TabbedPagePage> */
  children: React.ReactNode;
}

/** A reusable, professional tabbed page layout with optional sticky header. */
export const TabbedPage: React.FC<TabbedPageProps> = ({
  title,
  subtitle,
  activeValue,
  onValueChange,
  leftActions,
  rightActions,
  stickyHeader = true,
  contentClassName,
  children,
}) => {
  // Be permissive: only keep children that have value + label props.
  type PageEl = React.ReactElement<{
    value: string;
    label: string;
    children: React.ReactNode;
  }>;
  const pages: PageEl[] = Children.toArray(children)
    .filter(isValidElement)
    .filter(
      (el: any) =>
        el?.props &&
        typeof el.props.value === "string" &&
        typeof el.props.label === "string"
    ) as PageEl[];

  return (
    <Tabs value={activeValue} onValueChange={onValueChange} className="block">
      {/* Header */}
      <div
        className={[
          stickyHeader
            ? "sticky top-0 z-30 border-b bg-background/80 backdrop-blur supports-[backdrop-filter]:bg-background/60"
            : "",
        ].join(" ")}
      >
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-3">
          <div className="flex min-w-0 items-center gap-4">
            {leftActions}
            {(title || subtitle) && (
              <div className="min-w-0">
                {title && (
                  <h1 className="truncate text-xl font-semibold tracking-tight">
                    {title}
                  </h1>
                )}
                {subtitle && (
                  <p className="truncate text-sm text-muted-foreground">
                    {subtitle}
                  </p>
                )}
              </div>
            )}
          </div>
          {rightActions && (
            <div className="flex items-center gap-2">{rightActions}</div>
          )}
        </div>

        {/* Underline tabs */}
        <div className="mx-auto max-w-5xl px-6">
          <TabsList className="h-auto w-full justify-start gap-6 rounded-none border-b bg-transparent p-0">
            {pages.map((p) => (
              <TabsTrigger
                key={p.props.value}
                value={p.props.value}
                className={[
                  "relative h-10 rounded-none bg-transparent px-1 text-sm font-medium",
                  "text-muted-foreground shadow-none data-[state=active]:text-foreground",
                  "data-[state=active]:shadow-none",
                  // underline
                  "data-[state=active]:before:absolute data-[state=active]:before:inset-x-0 data-[state=active]:before:-bottom-[1px]",
                  "data-[state=active]:before:h-0.5 data-[state=active]:before:bg-primary/90",
                  "transition-colors hover:text-foreground",
                ].join(" ")}
              >
                {p.props.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
      </div>

      {/* Content */}
      <div className={`mx-auto max-w-5xl px-6 pt-6 ${contentClassName ?? ""}`}>
        {pages.map((p) => (
          <TabsContent
            key={p.props.value}
            value={p.props.value}
            className="space-y-6"
          >
            {p.props.children}
          </TabsContent>
        ))}
      </div>
    </Tabs>
  );
};

export default TabbedPage;
