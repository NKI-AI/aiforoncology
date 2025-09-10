// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { Outlet } from "@tanstack/react-router";
import Navbar from "./Navbar";
import Footer from "./Footer";

interface MainLayoutProps {
  showFooter?: boolean;
  showNavbar?: boolean;
  backgroundColor?: string;
  isViewerRoute?: boolean;
}

export function MainLayout({
  showFooter = true,
  showNavbar = true,
  backgroundColor = "bg-background",
  isViewerRoute = false,
}: MainLayoutProps) {
  return (
    <div
      className={`${backgroundColor} text-foreground ${
        isViewerRoute ? "h-screen" : "min-h-screen"
      } flex flex-col antialiased no-elastic-scroll`}
    >
      {showNavbar && <Navbar />}
      <main
        className={`flex-1 ${showNavbar ? "pt-11" : ""} ${
          isViewerRoute ? "overflow-hidden min-h-0" : "overflow-auto"
        } no-elastic-scroll`}
      >
        <Outlet />
      </main>
      {showFooter && <Footer showDescription={false} />}
    </div>
  );
}
