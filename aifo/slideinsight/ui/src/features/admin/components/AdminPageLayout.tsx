// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import AdminSidebar from "./AdminSidebar";
import AdminHeader from "./AdminHeader";
import { SidebarInset, SidebarProvider } from "../../../components/ui/sidebar";

interface AdminPageLayoutProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
  maxWidth?: string;
  className?: string;
}

const AdminPageLayout: React.FC<AdminPageLayoutProps> = ({
  title,
  description,
  actions,
  children,
  maxWidth = "max-w-7xl",
  className = "",
}) => {
  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title={title}
          description={description}
          actions={actions}
        />
        <div className={`container mx-auto px-4 py-4 ${maxWidth} ${className}`}>
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
};

export default AdminPageLayout;
