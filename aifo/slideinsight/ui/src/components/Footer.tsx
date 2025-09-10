// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";

interface FooterProps {
  showBranding?: boolean;
  showDescription?: boolean;
}

const Footer: React.FC<FooterProps> = ({
  showBranding = true,
  showDescription = true,
}) => {
  return (
    <footer className="bg-indigo-700 dark:bg-indigo-900 text-white">
      <div className="container mx-auto px-4 py-8">
        <div className="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0">
          {showBranding && (
            <div className="text-center md:text-left">
              <h3 className="text-lg font-semibold mb-1">SlideInsight</h3>
              {showDescription && (
                <p className="text-indigo-100 text-sm">
                  Computational Pathology Platform
                </p>
              )}
            </div>
          )}

          <div className="text-center">
            {/* <p className="text-indigo-100 text-sm">For AI in Oncology Research</p> */}
            <p className="text-xs text-indigo-200 mt-1">
              © {new Date().getFullYear()} SlideInsight. All rights reserved.
            </p>
          </div>

          <div className="text-center md:text-right">
            <p className="text-xs text-indigo-200">Version 0.1.0</p>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
