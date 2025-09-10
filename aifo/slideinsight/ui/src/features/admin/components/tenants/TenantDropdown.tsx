// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useRef } from "react";
import { useTenants } from "../../hooks/useTenants";
import { Tenant } from "../../../../api/models";
import { ChevronDownIcon } from "../../../../components/icons";

interface TenantDropdownProps {
  value: string;
  onChange: (tenantUid: string) => void;
  onBlur?: () => void;
  disabled?: boolean;
  placeholder?: string;
  error?: string;
  required?: boolean;
}

const TenantDropdown: React.FC<TenantDropdownProps> = ({
  value,
  onChange,
  onBlur,
  disabled = false,
  placeholder = "Select a tenant...",
  error,
  required = false,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Fetch tenants with search
  const {
    tenants,
    loading,
    error: fetchError,
  } = useTenants({
    q: searchQuery,
    limit: 50, // Reasonable limit for dropdown
  });

  // Find selected tenant when value changes
  useEffect(() => {
    if (value && tenants.length > 0) {
      const tenant = tenants.find((t) => t.tenantUid === value);
      if (tenant) {
        setSelectedTenant(tenant);
      } else {
        // If current value not in current list, keep the existing selectedTenant for display
        // or fetch specifically for this tenant
        if (!selectedTenant || selectedTenant.tenantUid !== value) {
          setSelectedTenant({
            tenantUid: value,
            name: `Tenant ${value}`,
            description: "",
            status: "active",
            createdAt: "",
            updatedAt: "",
          });
        }
      }
    } else if (!value) {
      setSelectedTenant(null);
    }
  }, [value, tenants]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
        setSearchQuery("");
        onBlur?.();
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [isOpen, onBlur]);

  // Focus search input when dropdown opens
  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isOpen]);

  const handleTenantSelect = (tenant: Tenant) => {
    setSelectedTenant(tenant);
    onChange(tenant.tenantUid);
    setIsOpen(false);
    setSearchQuery("");
    onBlur?.();
  };

  const handleToggle = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
      if (!isOpen) {
        setSearchQuery("");
      }
    }
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Dropdown trigger */}
      <button
        type="button"
        onClick={handleToggle}
        disabled={disabled}
        className={`w-full px-3 py-2 text-left border rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring ${
          disabled
            ? "bg-muted text-muted-foreground cursor-not-allowed"
            : "bg-background hover:bg-muted cursor-pointer"
        } ${error ? "border-destructive" : "border-border"} ${
          isOpen ? "ring-2 ring-ring border-ring" : ""
        }`}
      >
        <div className="flex items-center justify-between">
          <span
            className={selectedTenant ? "text-muted-900" : "text-muted-500"}
          >
            {selectedTenant ? selectedTenant.name : placeholder}
            {required && !selectedTenant && (
              <span className="text-red-500 ml-1">*</span>
            )}
          </span>
          <ChevronDownIcon
            className={`h-4 w-4 text-muted-400 transition-transform ${
              isOpen ? "transform rotate-180" : ""
            }`}
          />
        </div>
      </button>

      {/* Dropdown content */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-popover border border-border rounded-md shadow-lg max-h-60 overflow-hidden">
          {/* Search input */}
          <div className="p-2 border-b border-gray-200">
            <div className="relative">
              <svg
                className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"
                />
              </svg>
              <input
                ref={searchInputRef}
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search tenants..."
                className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          {/* Results */}
          <div className="max-h-48 overflow-y-auto">
            {loading ? (
              <div className="px-3 py-2 text-sm text-muted-500">
                <div className="animate-pulse">Loading tenants...</div>
              </div>
            ) : fetchError ? (
              <div className="px-3 py-2 text-sm text-red-600">
                Error loading tenants
              </div>
            ) : tenants.length === 0 ? (
              <div className="px-3 py-2 text-sm text-muted-500">
                {searchQuery ? "No tenants found" : "No tenants available"}
              </div>
            ) : (
              tenants.map((tenant) => (
                <button
                  key={tenant.tenantUid}
                  type="button"
                  onClick={() => handleTenantSelect(tenant)}
                  className={`w-full px-3 py-2 text-left text-sm hover:bg-blue-50 focus:outline-none focus:bg-blue-50 ${
                    selectedTenant?.tenantUid === tenant.tenantUid
                      ? "bg-blue-100 text-blue-900"
                      : "text-muted-900"
                  }`}
                >
                  <div className="font-medium">{tenant.name}</div>
                  <div className="text-xs text-muted-500 font-mono">
                    {tenant.tenantUid}
                  </div>
                </button>
              ))
            )}
          </div>
        </div>
      )}

      {/* Error message */}
      {error && <p className="mt-1 text-sm text-red-600">{error}</p>}
    </div>
  );
};

export default TenantDropdown;
