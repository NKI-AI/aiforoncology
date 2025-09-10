// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Link } from "@tanstack/react-router";
import { LogoBranding } from "@/components/LogoBranding";
import { DarkModeToggle } from "@/components/DarkModeToggle";

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 py-12 px-4 relative">
      {/* Dark Mode Toggle - Top right */}
      <div className="absolute top-4 right-4 z-10">
        <DarkModeToggle size="md" />
      </div>

      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex justify-center mb-8">
            <LogoBranding size="small" linkTo="/" />
          </div>
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
            Privacy Policy
          </h2>
          <p className="text-gray-600 dark:text-gray-300">
            Last updated: June 2025
          </p>
        </div>

        {/* Content */}
        <div className="bg-white/90 dark:bg-gray-800/50 backdrop-blur-sm rounded-xl p-8 border border-gray-200/50 dark:border-gray-700/50">
          <div className="prose prose-gray max-w-none text-gray-700 dark:text-gray-200">
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              1. Information We Collect
            </h3>
            <p className="mb-6">
              We collect information you provide directly to us, such as when
              you create an account, use our services, or contact us for
              support. This may include your name, email address, and other
              information you choose to provide.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              2. How We Use Your Information
            </h3>
            <p className="mb-6">
              We use the information we collect to provide, maintain, and
              improve our services, process transactions, send you technical
              notices and support messages, and communicate with you about
              products, services, and events.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              3. Information Sharing
            </h3>
            <p className="mb-6">
              We do not sell, trade, or otherwise transfer your personal
              information to third parties without your consent, except as
              described in this policy. We may share your information in certain
              limited circumstances, such as with service providers who assist
              us in operating our platform.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              4. Data Security
            </h3>
            <p className="mb-6">
              We implement appropriate technical and organizational measures to
              protect your personal information against unauthorized access,
              alteration, disclosure, or destruction. However, no method of
              transmission over the internet is 100% secure.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              5. Data Retention
            </h3>
            <p className="mb-6">
              We retain your personal information for as long as necessary to
              provide you with our services and as described in this privacy
              policy. We may also retain and use this information as necessary
              to comply with our legal obligations, resolve disputes, and
              enforce our agreements.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              6. Your Rights
            </h3>
            <p className="mb-6">
              Depending on your location, you may have certain rights regarding
              your personal information, including the right to access, update,
              or delete your information. Please contact us if you would like to
              exercise any of these rights.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              7. Medical Data
            </h3>
            <p className="mb-6">
              SlideInsight processes medical and pathological data for research
              and diagnostic purposes. All medical data is handled in accordance
              with applicable healthcare privacy laws and regulations, including
              HIPAA where applicable.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              8. Contact Us
            </h3>
            <p className="mb-6">
              If you have any questions about this Privacy Policy, please
              contact us through our support channels.
            </p>
          </div>
        </div>

        {/* Back to Login */}
        <div className="text-center mt-8">
          <Link
            to="/login"
            className="inline-flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors duration-200"
          >
            ← Back to Login
          </Link>
        </div>
      </div>
    </div>
  );
}
