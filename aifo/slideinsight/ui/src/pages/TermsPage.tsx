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

export default function TermsPage() {
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
            Terms of Service
          </h2>
          <p className="text-gray-600 dark:text-gray-300">
            Last updated: June 2025
          </p>
        </div>

        {/* Content */}
        <div className="bg-white/90 dark:bg-gray-800/50 backdrop-blur-sm rounded-xl p-8 border border-gray-200/50 dark:border-gray-700/50">
          <div className="prose prose-gray max-w-none text-gray-700 dark:text-gray-200">
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              1. Acceptance of Terms
            </h3>
            <p className="mb-6">
              By accessing and using SlideInsight, you accept and agree to be
              bound by the terms and provision of this agreement.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              2. Use License
            </h3>
            <p className="mb-6">
              Permission is granted to temporarily download one copy of
              SlideInsight per device for personal, non-commercial transitory
              viewing only.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              3. Disclaimer
            </h3>
            <p className="mb-6">
              The materials on SlideInsight are provided on an 'as is' basis.
              SlideInsight makes no warranties, expressed or implied, and hereby
              disclaims and negates all other warranties including without
              limitation, implied warranties or conditions of merchantability,
              fitness for a particular purpose, or non-infringement of
              intellectual property or other violation of rights.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              4. Limitations
            </h3>
            <p className="mb-6">
              In no event shall SlideInsight or its suppliers be liable for any
              damages (including, without limitation, damages for loss of data
              or profit, or due to business interruption) arising out of the use
              or inability to use SlideInsight, even if SlideInsight or a
              SlideInsight authorized representative has been notified orally or
              in writing of the possibility of such damage.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              5. Privacy Policy
            </h3>
            <p className="mb-6">
              Your privacy is important to us. Please review our Privacy Policy,
              which also governs your use of the Service, to understand our
              practices.
            </p>

            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              6. Contact Information
            </h3>
            <p className="mb-6">
              If you have any questions about these Terms of Service, please
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
