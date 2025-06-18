// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../auth";
import Navbar from "./Navbar";
import Footer from "./Footer";
import { UserIcon, KeyIcon, SettingsIcon, ArrowLeftIcon } from "./icons";

export default function AccountPage() {
  const { user } = useAuth();

  return (
    <div className="bg-gray-100 text-gray-800 font-sans min-h-screen flex flex-col">
      {/* Use the shared Navbar component */}
      <Navbar
        onToggleMaskControl={() => {}}
        onToggleSlideInfo={() => {}}
        onToggleHelp={() => {}}
        onToggleCrosshair={() => {}}
        showCrosshair={false}
      />

      {/* Main content container */}
      <div className="container mx-auto px-4 py-8 mt-16 max-w-4xl flex-grow">
        {/* Header section */}
        <div className="mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between bg-white rounded-xl shadow-sm p-6 border border-indigo-100">
          <div>
            <h1 className="text-2xl font-bold text-indigo-800 mb-1">
              Account Settings
            </h1>
            <p className="text-indigo-500">
              Manage your account preferences and security
            </p>
          </div>

          <div className="mt-4 sm:mt-0">
            <Link
              to="/slides"
              className="inline-flex items-center text-indigo-600 hover:text-indigo-800 transition-colors"
            >
              <ArrowLeftIcon />
              Back to Slides
            </Link>
          </div>
        </div>

        {/* Account sections */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Profile Information */}
          <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-center mb-4">
              <div className="bg-indigo-100 rounded-full p-3 mr-4">
                <UserIcon className="h-6 w-6 text-indigo-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-800">
                  Profile Information
                </h2>
                <p className="text-sm text-gray-600">
                  View and manage your profile details
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Username
                </label>
                <p className="text-sm text-gray-900 bg-gray-50 px-3 py-2 rounded border">
                  {user?.username || "Not available"}
                </p>
              </div>

              {user?.email && (
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Email
                  </label>
                  <p className="text-sm text-gray-900 bg-gray-50 px-3 py-2 rounded border">
                    {user.email}
                  </p>
                </div>
              )}

              {user?.roles && user.roles.length > 0 && (
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Roles
                  </label>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {user.roles.map((role, index) => (
                      <span
                        key={index}
                        className="px-2 py-1 text-xs font-medium bg-indigo-100 text-indigo-800 rounded"
                      >
                        {role}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Security Settings */}
          <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-center mb-4">
              <div className="bg-green-100 rounded-full p-3 mr-4">
                <KeyIcon className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-800">
                  Security
                </h2>
                <p className="text-sm text-gray-600">
                  Manage your password and security settings
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <Link
                to="/account/reset_password"
                className="block w-full text-left px-4 py-3 bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-800">Reset Password</p>
                    <p className="text-sm text-gray-600">
                      Change your account password
                    </p>
                  </div>
                  <svg
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </Link>

              <div className="block w-full text-left px-4 py-3 bg-gray-50 rounded border border-gray-200 opacity-50">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-800">
                      Two-Factor Authentication
                    </p>
                    <p className="text-sm text-gray-600">Coming soon</p>
                  </div>
                  <svg
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Preferences */}
          <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-center mb-4">
              <div className="bg-purple-100 rounded-full p-3 mr-4">
                <SettingsIcon className="h-6 w-6 text-purple-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-800">
                  Preferences
                </h2>
                <p className="text-sm text-gray-600">
                  Customize your application experience
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="block w-full text-left px-4 py-3 bg-gray-50 rounded border border-gray-200 opacity-50">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-800">
                      Display Settings
                    </p>
                    <p className="text-sm text-gray-600">Coming soon</p>
                  </div>
                  <svg
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>

              <div className="block w-full text-left px-4 py-3 bg-gray-50 rounded border border-gray-200 opacity-50">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-800">
                      Notification Settings
                    </p>
                    <p className="text-sm text-gray-600">Coming soon</p>
                  </div>
                  <svg
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Account Actions */}
          <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-center mb-4">
              <div className="bg-red-100 rounded-full p-3 mr-4">
                <svg
                  className="h-6 w-6 text-red-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                  />
                </svg>
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-800">
                  Account Actions
                </h2>
                <p className="text-sm text-gray-600">
                  Manage your account status
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="block w-full text-left px-4 py-3 bg-gray-50 rounded border border-gray-200 opacity-50">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-800">
                      Export Account Data
                    </p>
                    <p className="text-sm text-gray-600">Coming soon</p>
                  </div>
                  <svg
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>

              <div className="block w-full text-left px-4 py-3 bg-red-50 rounded border border-red-200 opacity-50">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-red-800">Delete Account</p>
                    <p className="text-sm text-red-600">Coming soon</p>
                  </div>
                  <svg
                    className="h-5 w-5 text-red-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <Footer showDescription={false} />
    </div>
  );
}
