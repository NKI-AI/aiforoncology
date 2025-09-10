// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { SidebarTrigger } from "@/components/ui/sidebar";

interface AdminHeaderProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
}

const AdminHeader: React.FC<AdminHeaderProps> = ({
  title = "Admin Dashboard",
  description = "Manage users, tenants, studies, slides, and system configuration",
  actions,
}) => {
  return (
    <header className="flex h-16 shrink-0 items-center gap-2 px-4 border-b border-sidebar-border bg-background">
      <SidebarTrigger className="-ml-1" />
      <div className="mx-2 h-4 w-px bg-sidebar-border" />
      <div className="flex items-center flex-1">
        <div className="flex-1">
          <h1 className="text-lg font-semibold text-foreground">{title}</h1>
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  );
};

export default AdminHeader;
