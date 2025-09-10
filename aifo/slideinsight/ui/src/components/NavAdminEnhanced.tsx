// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
"use client";

import * as React from "react";
import { Link, useLocation } from "@tanstack/react-router";
import type { Icon } from "@tabler/icons-react";
import { ChevronRightIcon } from "@/components/icons";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
} from "@/components/ui/sidebar";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

interface NavItem {
  title: string;
  url: string;
  icon: Icon;
}

interface CollapsibleNavItem {
  title: string;
  icon: Icon;
  items: NavItem[];
  defaultOpen?: boolean;
}

interface NavAdminEnhancedProps
  extends React.ComponentPropsWithoutRef<typeof SidebarGroup> {
  /** Regular navigation items that are not collapsible */
  items?: NavItem[];
  /** Collapsible navigation sections with sub-items */
  collapsibleItems?: CollapsibleNavItem[];
}

/**
 * Utility function to determine if a navigation item is active
 */
function isNavItemActive(currentPath: string, itemUrl: string): boolean {
  return (
    currentPath === itemUrl ||
    (itemUrl !== "/admin" && currentPath.startsWith(itemUrl))
  );
}

// Memoized collapsible section component to prevent unnecessary rerenders
const CollapsibleSection = React.memo(
  ({
    collapsibleItem,
    location,
  }: {
    collapsibleItem: CollapsibleNavItem;
    location: { pathname: string };
  }) => {
    const isSectionActive = React.useMemo(() => {
      return collapsibleItem.items.some((item) =>
        isNavItemActive(location.pathname, item.url)
      );
    }, [collapsibleItem.items, location.pathname]);

    // Use controlled state instead of defaultOpen to prevent rerenders
    const [isOpen, setIsOpen] = React.useState(
      collapsibleItem.defaultOpen || isSectionActive
    );

    // Update open state when section becomes active/inactive
    React.useEffect(() => {
      if (isSectionActive && !isOpen) {
        setIsOpen(true);
      }
    }, [isSectionActive, isOpen]);

    return (
      <Collapsible
        open={isOpen}
        onOpenChange={setIsOpen}
        className="group/collapsible"
      >
        <SidebarMenuItem>
          <CollapsibleTrigger asChild>
            <SidebarMenuButton isActive={isSectionActive}>
              <collapsibleItem.icon className="h-4 w-4 shrink-0" />
              <span>{collapsibleItem.title}</span>
              {/* This chevron rotates 90° when open - this is intentional UX behavior */}
              <ChevronRightIcon className="ml-auto h-4 w-4 shrink-0 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
            </SidebarMenuButton>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <SidebarMenuSub>
              {collapsibleItem.items.map((subItem) => {
                const isSubActive = isNavItemActive(
                  location.pathname,
                  subItem.url
                );

                return (
                  <SidebarMenuSubItem key={subItem.title}>
                    <SidebarMenuSubButton asChild isActive={isSubActive}>
                      <Link to={subItem.url}>
                        <subItem.icon className="h-4 w-4 shrink-0 transform-none" />
                        <span>{subItem.title}</span>
                      </Link>
                    </SidebarMenuSubButton>
                  </SidebarMenuSubItem>
                );
              })}
            </SidebarMenuSub>
          </CollapsibleContent>
        </SidebarMenuItem>
      </Collapsible>
    );
  }
);

CollapsibleSection.displayName = "CollapsibleSection";

/**
 * Enhanced navigation component for admin sidebar that supports both regular and collapsible navigation items.
 *
 * Features:
 * - Regular navigation items with active state detection
 * - Collapsible sections with sub-items
 * - Automatic expansion when a sub-item is active
 * - Smooth animations with chevron rotation
 * - Full TypeScript support
 * - Optimized to prevent unnecessary rerenders
 *
 * @example
 * ```tsx
 * // Basic usage with both regular and collapsible items
 * <NavAdminEnhanced
 *   items={[
 *     { title: "Dashboard", url: "/admin", icon: DashboardIcon },
 *     { title: "Users", url: "/admin/users", icon: UsersIcon }
 *   ]}
 *   collapsibleItems={[
 *     {
 *       title: "System",
 *       icon: SystemIcon,
 *       defaultOpen: false,
 *       items: [
 *         { title: "Monitor", url: "/admin/system/monitor", icon: MonitorIcon },
 *         { title: "Logs", url: "/admin/system/logs", icon: LogsIcon }
 *       ]
 *     }
 *   ]}
 * />
 * ```
 */
export const NavAdminEnhanced = React.memo<NavAdminEnhancedProps>(
  ({ items = [], collapsibleItems = [], ...props }) => {
    const location = useLocation();

    return (
      <SidebarGroup {...props}>
        <SidebarGroupContent>
          <SidebarMenu>
            {/* Regular navigation items */}
            {items.map((item) => {
              const isActive = isNavItemActive(location.pathname, item.url);

              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={isActive}>
                    <Link to={item.url}>
                      <item.icon className="h-4 w-4 shrink-0 transform-none" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })}

            {/* Collapsible navigation sections */}
            {collapsibleItems.map((collapsibleItem) => (
              <CollapsibleSection
                key={collapsibleItem.title}
                collapsibleItem={collapsibleItem}
                location={location}
              />
            ))}
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    );
  }
);

NavAdminEnhanced.displayName = "NavAdminEnhanced";
