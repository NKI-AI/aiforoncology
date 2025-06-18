// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { useState, useRef, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router-dom";
import { useAuth } from "../auth";
import {
  SlideScopeIcon,
  MaskControlsIcon,
  SlideInfoIcon,
  CrosshairIcon,
  HelpIcon,
  LogoutIcon,
  UserIcon,
} from "./icons";

interface NavbarProps {
  onToggleMaskControl?: () => void;
  onToggleSlideInfo?: () => void;
  onToggleHelp?: () => void;
  onToggleCrosshair?: () => void;
  showCrosshair?: boolean;
}

// For type safety with user data
interface User {
  username: string;
  roles?: string[];
  email?: string;
}

export default function Navbar({
  onToggleMaskControl,
  onToggleSlideInfo,
  onToggleHelp,
  onToggleCrosshair,
  showCrosshair,
}: NavbarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Check if we're on the viewer page - i.e., paths starting with /view
  const isViewerPage = location.pathname.startsWith("/v");

  const handleLogout = async () => {
    await logout();
    navigate("/login");
    setIsDropdownOpen(false);
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsDropdownOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <nav className="bg-indigo-700 text-white shadow-md flex items-center h-12 w-full fixed top-0 left-0 right-0 z-20">
      <div className="flex-1 flex items-center px-4">
        {isViewerPage ? (
          <Link
            to="/slides"
            className="flex items-center hover:text-indigo-200 transition-colors"
          >
            <SlideScopeIcon />
            <span className="text-xl font-bold ml-2">SlideScope</span>
          </Link>
        ) : (
          <>
            <SlideScopeIcon />
            <span className="text-xl font-bold ml-2">SlideScope</span>
          </>
        )}
      </div>

      {/* Toolbar buttons */}
      <div className="flex items-center gap-3 pr-4">
        {/* Only show these controls on the viewer page */}
        {isViewerPage && (
          <>
            {onToggleMaskControl && (
              <button
                className="toolbar-button"
                title="Mask Controls"
                onClick={onToggleMaskControl}
              >
                <MaskControlsIcon />
              </button>
            )}

            {onToggleSlideInfo && (
              <button
                className="toolbar-button"
                title="Slide Information"
                onClick={onToggleSlideInfo}
              >
                <SlideInfoIcon />
              </button>
            )}

            {onToggleCrosshair && (
              <button
                className={`toolbar-button ${
                  showCrosshair ? "bg-indigo-500" : ""
                }`}
                title="Toggle Crosshair"
                onClick={onToggleCrosshair}
              >
                <CrosshairIcon />
              </button>
            )}

            {onToggleHelp && (
              <button
                className="toolbar-button"
                title="Keyboard Shortcuts"
                onClick={onToggleHelp}
              >
                <HelpIcon />
              </button>
            )}
          </>
        )}

        {/* User Dropdown Menu */}
        {user && (
          <div className="relative inline-flex" ref={dropdownRef}>
            <span className="inline-flex divide-x divide-indigo-600 overflow-hidden rounded border border-indigo-600 bg-indigo-600 shadow-sm">
              <button
                type="button"
                className="px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-indigo-500 focus:relative"
              >
                {user.username}
              </button>

              <button
                type="button"
                className="px-2 py-1.5 text-sm font-medium text-white transition-colors hover:bg-indigo-500 focus:relative"
                aria-label="User Menu"
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth="1.5"
                  stroke="currentColor"
                  className="h-4 w-4"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="m19.5 8.25-7.5 7.5-7.5-7.5"
                  />
                </svg>
              </button>
            </span>

            {isDropdownOpen && (
              <div
                role="menu"
                className="absolute right-0 top-12 z-50 w-56 divide-y divide-gray-200 overflow-hidden rounded border border-gray-300 bg-white shadow-lg"
              >
                <div>
                  <Link
                    to="/account"
                    className="flex items-center px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 hover:text-gray-900"
                    role="menuitem"
                    onClick={() => setIsDropdownOpen(false)}
                  >
                    <UserIcon className="h-4 w-4 mr-2 text-gray-500" />
                    Account Settings
                  </Link>
                </div>

                <button
                  type="button"
                  className="flex items-center w-full px-3 py-2 text-left text-sm font-medium text-red-700 transition-colors hover:bg-red-50"
                  onClick={handleLogout}
                >
                  <LogoutIcon className="h-4 w-4 mr-2 text-red-500" />
                  Logout
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </nav>
  );
}
