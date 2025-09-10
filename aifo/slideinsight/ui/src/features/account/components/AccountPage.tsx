// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import { useAuth } from "@/auth";
import {
  UserIcon,
  KeyIcon,
  SettingsIcon,
  ArrowLeftIcon,
} from "@/components/icons";
import { LogoBranding } from "../../../components/LogoBranding";
import { DarkModeToggle } from "../../../components/DarkModeToggle";

export default function AccountPage() {
  const { user } = useAuth();

  if (!user) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 flex items-center justify-center relative">
        <div className="absolute top-4 right-4 z-10">
          <DarkModeToggle size="md" />
        </div>
        <div className="text-gray-600 dark:text-muted-400">
          Please log in to view your account.
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 relative">
      <div className="absolute top-4 right-4 z-10">
        <DarkModeToggle size="md" />
      </div>

      <div className="container mx-auto px-4 py-8">
        <LogoBranding className="mb-8" />

        <div className="max-w-2xl mx-auto">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-muted-100 mb-8">
            Account Information
          </h1>

          <div className="bg-white/95 dark:bg-gray-800/90 backdrop-blur-sm rounded-xl shadow-2xl border border-gray-200/50 dark:border-gray-700/50 overflow-hidden">
            <div className="p-6 space-y-6">
              {/* Email */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-muted-300 mb-2">
                  Email
                </label>
                <div className="text-gray-900 dark:text-muted-100 bg-gray-100 dark:bg-gray-700/50 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600/50">
                  {user?.email || "Not available"}
                </div>
              </div>

              {/* User UID */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-muted-300 mb-2">
                  User ID
                </label>
                <div className="text-gray-900 dark:text-muted-100 bg-gray-100 dark:bg-gray-700/50 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600/50 font-mono text-sm">
                  {user?.userUid || "Not available"}
                </div>
              </div>

              {/* Roles */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-muted-300 mb-2">
                  Roles
                </label>
                <div className="flex flex-wrap gap-2">
                  {user?.roles && user.roles.length > 0 ? (
                    user.roles.map((role, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-indigo-100 dark:bg-indigo-600/20 text-indigo-800 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30"
                      >
                        {role}
                      </span>
                    ))
                  ) : (
                    <span className="text-gray-600 dark:text-muted-400 text-sm">
                      No roles assigned
                    </span>
                  )}
                </div>
              </div>
            </div>

            <div className="px-6 py-4 bg-gray-100 dark:bg-gray-800/50 border-t border-gray-200 dark:border-gray-700/30">
              <p className="text-sm text-gray-600 dark:text-muted-400">
                To update your account information, please contact your
                administrator.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
