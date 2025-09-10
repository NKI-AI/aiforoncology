// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useRef } from "react";
import { useLocation, Link, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@/auth";
import {
  SlideInsightIcon,
  SlideInfoIcon,
  HelpIcon,
  SunIcon,
  MoonIcon,
  UserIcon,
  AdminIcon,
  LogoutIcon,
} from "@/components/icons";
import NotificationDropdown from "@/components/NotificationDropdown";

interface NavbarProps {
  onToggleSlideInfo?: () => void;
  onToggleHelp?: () => void;
}

// For type safety with user data
interface User {
  firstName: string;
  lastName: string;
  roles?: string[];
  email?: string;
  originalEmail?: string;
}

export default function Navbar({
  onToggleSlideInfo,
  onToggleHelp,
}: NavbarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement | null>(null);
  const [isDark, setIsDark] = useState<boolean>(() => {
    try {
      const storedTheme = localStorage.getItem("si-theme");
      if (storedTheme === "dark") return true;
      if (storedTheme === "light") return false;
    } catch (_) {
      // Ignore access errors (e.g., privacy mode)
    }
    if (typeof window !== "undefined" && window.matchMedia) {
      return window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    return false;
  });

  // Check if we're on the viewer page - i.e., paths starting with /view
  const isViewerPage = location.pathname.startsWith("/v");

  // Check if user is admin
  const isAdmin =
    user?.roles?.includes("superadmin") ||
    user?.roles?.some((role) => role.includes("platform.admin"));

  // Handle click outside user menu
  useEffect(() => {
    const onDocClick = (e: MouseEvent) => {
      if (
        userMenuRef.current &&
        !userMenuRef.current.contains(e.target as Node)
      ) {
        setIsUserMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  // Sync theme to document and persist in localStorage
  useEffect(() => {
    const root = document.documentElement;
    if (isDark) {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
    try {
      localStorage.setItem("si-theme", isDark ? "dark" : "light");
    } catch (_) {
      // Ignore storage errors
    }
  }, [isDark]);

  const handleLogout = async () => {
    await logout();
    navigate({ to: "/login" });
    setIsUserMenuOpen(false);
  };

  return (
    <nav className="fixed inset-x-0 top-0 z-30 bg-indigo-700 dark:bg-indigo-900 text-white shadow">
      <div className="mx-auto max-w-screen-2xl px-3 sm:px-4">
        <div className="flex h-11 items-center justify-between gap-2">
          {/* Left: Brand + primary nav */}
          <div className="flex items-center gap-2 min-w-0">
            {isViewerPage ? (
              <Link
                to="/studies"
                className="flex items-center gap-2 hover:text-indigo-200 transition-colors"
              >
                <SlideInsightIcon />
                <span className="text-base sm:text-lg font-semibold truncate">
                  SlideInsight
                </span>
              </Link>
            ) : (
              <div className="flex items-center gap-2">
                <SlideInsightIcon />
                <span className="text-base sm:text-lg font-semibold">
                  SlideInsight
                </span>
              </div>
            )}
            {!isViewerPage && (
              <div className="hidden md:flex items-center gap-2 ml-2 sm:ml-4">
                <Link
                  to="/studies"
                  className={`rounded-md px-3 py-1 text-sm font-medium transition-colors ${
                    location.pathname === "/studies" ||
                    location.pathname.startsWith("/studies/")
                      ? "bg-indigo-600 dark:bg-indigo-800 text-white"
                      : "text-indigo-100 hover:text-white hover:bg-indigo-600 dark:hover:bg-indigo-800"
                  }`}
                >
                  Studies
                </Link>
              </div>
            )}
          </div>

          {/* Right: Viewer tools, notifications, user, hamburger */}
          <div className="flex items-center gap-2 sm:gap-3">
            {/* Dark mode toggle */}
            <button
              type="button"
              onClick={() => setIsDark((v) => !v)}
              title={isDark ? "Switch to light mode" : "Switch to dark mode"}
              className="hidden md:inline-flex h-8 w-8 items-center justify-center rounded-md text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/50"
            >
              {isDark ? (
                <SunIcon className="h-4 w-4" />
              ) : (
                <MoonIcon className="h-4 w-4" />
              )}
            </button>
            {/* Viewer toolbar icons */}
            {isViewerPage && (
              <div className="flex items-center gap-1.5">
                {onToggleSlideInfo && (
                  <button
                    type="button"
                    title="Slide Information"
                    onClick={onToggleSlideInfo}
                    className="inline-flex h-8 w-8 items-center justify-center rounded-md text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/50"
                  >
                    <SlideInfoIcon />
                  </button>
                )}
                {onToggleHelp && (
                  <button
                    type="button"
                    title="Keyboard Shortcuts"
                    onClick={onToggleHelp}
                    className="inline-flex h-8 w-8 items-center justify-center rounded-md text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/50"
                  >
                    <HelpIcon />
                  </button>
                )}
              </div>
            )}

            {/* Notifications */}
            {user && <NotificationDropdown />}

            {/* User dropdown */}
            {user && (
              <div className="relative" ref={userMenuRef}>
                <button
                  type="button"
                  onClick={() => setIsUserMenuOpen((v) => !v)}
                  className="inline-flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/50 transition-colors"
                >
                  <UserIcon className="h-4 w-4" />
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="h-3 w-3"
                  >
                    <path d="M7 10l5 5 5-5z" />
                  </svg>
                </button>
                {isUserMenuOpen && (
                  <div className="absolute right-0 top-full mt-1 w-56 rounded-lg border border-white/10 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 shadow-lg ring-1 ring-black/5 dark:ring-white/10 z-50">
                    <div className="px-3 py-2 border-b border-gray-200 dark:border-gray-700 flex items-center gap-2">
                      <UserIcon className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                      <div className="truncate font-medium text-sm">
                        {user.email}
                      </div>
                    </div>
                    <div className="py-1">
                      <Link
                        to="/account"
                        className="flex items-center px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                        onClick={() => setIsUserMenuOpen(false)}
                      >
                        <UserIcon className="h-4 w-4 mr-2 text-gray-500 dark:text-gray-400" />
                        Account Settings
                      </Link>
                      {isAdmin && (
                        <Link
                          to="/admin"
                          className="flex items-center px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                          onClick={() => setIsUserMenuOpen(false)}
                        >
                          <AdminIcon className="h-4 w-4 mr-2 text-gray-500 dark:text-gray-400" />
                          Admin Dashboard
                        </Link>
                      )}
                      <button
                        type="button"
                        className="flex items-center w-full px-3 py-2 text-left text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                        onClick={handleLogout}
                      >
                        <LogoutIcon className="h-4 w-4 mr-2 text-red-600 dark:text-red-400" />
                        Logout
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Mobile hamburger (only when there are nav items to show) */}
            {!isViewerPage && (
              <button
                type="button"
                aria-label="Open menu"
                aria-controls="mobile-nav"
                aria-expanded={isMobileMenuOpen}
                onClick={() => setIsMobileMenuOpen((v) => !v)}
                className="inline-flex items-center justify-center rounded-md p-2 text-indigo-100 hover:bg-indigo-600 dark:hover:bg-indigo-800 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/50 md:hidden"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  className="h-6 w-6"
                >
                  {isMobileMenuOpen ? (
                    <path
                      fillRule="evenodd"
                      d="M6.225 4.811a.75.75 0 011.06 0L12 9.525l4.715-4.714a.75.75 0 111.06 1.06L13.06 10.586l4.715 4.714a.75.75 0 11-1.06 1.06L12 11.646l-4.715 4.714a.75.75 0 11-1.06-1.06l4.714-4.714-4.714-4.715a.75.75 0 010-1.06z"
                      clipRule="evenodd"
                    />
                  ) : (
                    <path d="M3.75 6.75h16.5v1.5H3.75v-1.5zM3.75 11.25h16.5v1.5H3.75v-1.5zM3.75 15.75h16.5v1.5H3.75v-1.5z" />
                  )}
                </svg>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Mobile panel */}
      {!isViewerPage && isMobileMenuOpen && (
        <div id="mobile-nav" className="md:hidden">
          <div className="space-y-1 px-3 pb-3 pt-2">
            <Link
              to="/studies"
              onClick={() => setIsMobileMenuOpen(false)}
              className={`block rounded-md px-3 py-2 text-base font-medium ${
                location.pathname === "/studies" ||
                location.pathname.startsWith("/studies/")
                  ? "bg-indigo-600 dark:bg-indigo-800 text-white"
                  : "text-indigo-100 hover:text-white hover:bg-indigo-600 dark:hover:bg-indigo-800"
              }`}
            >
              Studies
            </Link>
          </div>
        </div>
      )}
    </nav>
  );
}
