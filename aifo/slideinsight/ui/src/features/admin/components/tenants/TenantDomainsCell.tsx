// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch, ApiError } from "../../../../utils/fetchUtils";
import { queryKeys } from "../../../../utils/apiQueries";
import { CheckIcon, CloseIcon } from "../../../../components/icons";

interface TenantDomain {
  id: number;
  domain: string;
  isVerified: boolean;
  isPrimary: boolean;
  createdAt: string;
  updatedAt: string;
}

interface TenantDomainsCellProps {
  tenantUid: string;
}

const TenantDomainsCell: React.FC<TenantDomainsCellProps> = React.memo(
  ({ tenantUid }) => {
    const {
      data: domains = [],
      isLoading: loading,
      error,
    } = useQuery({
      queryKey: queryKeys.tenants.domains(tenantUid),
      queryFn: async () => {
        if (
          !tenantUid ||
          typeof tenantUid !== "string" ||
          tenantUid.trim() === ""
        ) {
          throw new Error("Invalid tenant ID");
        }

        const response = await apiFetch<{ domains: TenantDomain[] }>(
          `/api/v1/tenants/${tenantUid}/domains`
        );
        return response.domains || [];
      },
      enabled: !!tenantUid,
      staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
      gcTime: 10 * 60 * 1000, // Keep in cache for 10 minutes
    });

    if (loading) {
      return (
        <div className="flex items-center space-x-1">
          <div className="w-3 h-3 bg-gray-300 rounded animate-pulse"></div>
          <span className="text-xs text-muted-400">Loading...</span>
        </div>
      );
    }

    if (error) {
      let errorMessage = "Failed to load";
      if (error instanceof ApiError) {
        if (error.status === 404) {
          errorMessage = "Tenant not found";
        } else if (error.status === 403) {
          errorMessage = "Access denied";
        } else if (error.status === 401) {
          errorMessage = "Authentication required";
        }
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }

      return (
        <span className="text-xs text-red-500" title={errorMessage}>
          {errorMessage.includes("not found")
            ? "Tenant not found"
            : "Error loading domains"}
        </span>
      );
    }

    if (domains.length === 0) {
      return <span className="text-xs text-muted-400">No domains</span>;
    }

    const primaryDomains = domains.filter((d) => d.isPrimary);
    const verifiedCount = domains.filter((d) => d.isVerified).length;

    return (
      <div className="space-y-1 min-w-0">
        {domains.slice(0, 2).map((domain, index) => (
          <div key={index} className="flex items-center space-x-1 text-xs">
            <span
              className="font-mono text-muted-700 truncate max-w-[100px]"
              title={domain.domain}
            >
              {domain.domain}
            </span>
            <div className="flex items-center space-x-0.5">
              {domain.isPrimary && (
                <span className="inline-flex items-center px-1 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                  P
                </span>
              )}
              {domain.isVerified ? (
                <CheckIcon className="h-3 w-3 text-green-500" />
              ) : (
                <CloseIcon className="h-3 w-3 text-red-500" />
              )}
            </div>
          </div>
        ))}
        {domains.length > 2 && (
          <div className="text-xs text-muted-400">
            +{domains.length - 2} more ({verifiedCount}/{domains.length}{" "}
            verified)
          </div>
        )}
        {primaryDomains.length > 0 && domains.length <= 2 && (
          <div className="text-xs text-purple-600">
            {verifiedCount}/{domains.length} verified
          </div>
        )}
      </div>
    );
  }
);

TenantDomainsCell.displayName = "TenantDomainsCell";

export default TenantDomainsCell;
