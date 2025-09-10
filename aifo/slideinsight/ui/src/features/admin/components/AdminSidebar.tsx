// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { NavAdminEnhanced } from "@/components/NavAdminEnhanced";
import { NavAdminUser } from "@/components/NavAdminUser";
import {
  IconSettings,
  IconUsers,
  IconBuilding,
  IconPhoto,
  IconCircle,
  IconShield,
  IconList,
  IconCpu,
  IconBell,
  IconMail,
} from "@tabler/icons-react";
import { SlideInsightIcon } from "@/components/icons";
import { useAuth } from "@/auth";

// Move data outside component to prevent recreation on every render
const sidebarData = {
  navMain: [
    {
      title: "Dashboard",
      url: "/admin",
      icon: IconSettings,
    },
    {
      title: "Users",
      url: "/admin/users",
      icon: IconUsers,
    },
    {
      title: "Groups",
      url: "/admin/groups",
      icon: IconUsers,
    },
    {
      title: "Roles",
      url: "/admin/roles",
      icon: IconShield,
    },
    {
      title: "Tenants",
      url: "/admin/tenants",
      icon: IconBuilding,
    },
    {
      title: "Algorithms",
      url: "/admin/algorithms",
      icon: IconCpu,
    },
    {
      title: "Permissions",
      url: "/admin/permissions",
      icon: IconShield,
    },
    {
      title: "Studies",
      url: "/admin/studies",
      icon: IconCircle,
    },
    {
      title: "Cases",
      url: "/admin/cases",
      icon: IconCircle,
    },
    {
      title: "Slides",
      url: "/admin/slides",
      icon: IconPhoto,
    },
  ],
  systemSection: {
    title: "System",
    icon: IconSettings,
    defaultOpen: false,
    items: [
      {
        title: "Settings",
        url: "/admin/system/settings",
        icon: IconSettings,
      },
      {
        title: "System Monitor",
        url: "/admin/system/monitor",
        icon: IconSettings,
      },
      {
        title: "Queue Monitor",
        url: "/admin/system/queue",
        icon: IconList,
      },
      {
        title: "Notifications",
        url: "/admin/system/notifications",
        icon: IconBell,
      },
      {
        title: "Email Templates",
        url: "/admin/system/email",
        icon: IconMail,
      },
    ],
  },
};

interface AdminSidebarProps extends React.ComponentProps<typeof Sidebar> {}

const AdminSidebar: React.FC<AdminSidebarProps> = ({ ...props }) => {
  const { user } = useAuth();

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild className="px-2">
              <Link to="/admin">
                <SlideInsightIcon className="h-5 w-5" />
                <span className="text-base font-semibold">
                  SlideInsight Admin
                </span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <NavAdminEnhanced items={sidebarData.navMain} />
        <SidebarSeparator />
        <NavAdminEnhanced
          collapsibleItems={[sidebarData.systemSection]}
          className="mt-auto"
        />
      </SidebarContent>
    </Sidebar>
  );
};

export default AdminSidebar;
