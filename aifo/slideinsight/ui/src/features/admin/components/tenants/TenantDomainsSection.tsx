// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, {
  useState,
  useEffect,
  forwardRef,
  useImperativeHandle,
} from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiFetch } from "@/utils/fetchUtils";
import { queryKeys } from "@/utils/apiQueries";
import {
  PlusIcon,
  CheckIcon,
  CloseIcon,
  TrashIcon,
} from "../../../../components/icons";

interface TenantDomain {
  id: number;
  domain: string;
  isVerified: boolean;
  isPrimary: boolean;
  createdAt: string;
  updatedAt: string;
}

interface TenantDomainsSectionProps {
  tenantUid?: string | null;
  isCreating?: boolean;
  className?: string;
}

// Define the interface for methods that can be called via ref
export interface TenantDomainsSectionRef {
  getLocalDomains: () => TenantDomain[];
  clearLocalDomains: () => void;
}

const TenantDomainsSection = forwardRef<
  TenantDomainsSectionRef,
  TenantDomainsSectionProps
>(({ tenantUid, isCreating = false, className = "" }, ref) => {
  const [newDomain, setNewDomain] = useState("");
  const [newDomainIsPrimary, setNewDomainIsPrimary] = useState(false);
  const [localDomains, setLocalDomains] = useState<TenantDomain[]>([]);
  const [isAddingDomain, setIsAddingDomain] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const queryClient = useQueryClient();

  // Expose methods via ref
  useImperativeHandle(ref, () => ({
    getLocalDomains: () => localDomains,
    clearLocalDomains: () => setLocalDomains([]),
  }));

  // Only fetch domains if we have a tenantUid (editing mode)
  const {
    data: fetchedDomains = [],
    isLoading,
    error: fetchError,
  } = useQuery({
    queryKey: tenantUid
      ? queryKeys.tenants.domains(tenantUid)
      : ["domains", "null"],
    queryFn: async () => {
      if (!tenantUid || isCreating) {
        return [];
      }

      const response = await apiFetch<{ domains: TenantDomain[] }>(
        `/api/v1/tenants/${tenantUid}/domains`
      );
      return response.domains || [];
    },
    enabled: !!tenantUid && !isCreating,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

  // Sync fetched domains with local state
  useEffect(() => {
    if (!isCreating) {
      setLocalDomains(fetchedDomains);
    }
  }, [fetchedDomains, isCreating]);

  const validateDomain = (domain: string): string | null => {
    if (!domain.trim()) {
      return "Domain is required";
    }
    if (!/^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(domain)) {
      return "Please enter a valid domain (e.g., acme.com)";
    }
    if (
      localDomains.some((d) => d.domain.toLowerCase() === domain.toLowerCase())
    ) {
      return "Domain already exists";
    }
    return null;
  };

  const handleAddDomain = async () => {
    const validationError = validateDomain(newDomain);
    if (validationError) {
      setError(validationError);
      return;
    }

    setError(null);

    const domainToAdd: TenantDomain = {
      id: -Math.floor(Math.random() * 10000), // Temporary negative ID for local domains
      domain: newDomain.toLowerCase(),
      isVerified: false,
      isPrimary: newDomainIsPrimary,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    if (isCreating) {
      // For new tenants, just add to local state
      setLocalDomains((prev) => [...prev, domainToAdd]);
    } else if (tenantUid) {
      // For existing tenants, save immediately
      setIsAddingDomain(true);
      try {
        await apiFetch(`/api/v1/tenants/${tenantUid}/domains`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            domain: domainToAdd.domain,
            isPrimary: domainToAdd.isPrimary,
          }),
        });

        // Invalidate and refetch domains
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });

        toast.success("Domain added!", {
          description: `${domainToAdd.domain} has been successfully added.`,
        });
      } catch (err) {
        console.error("Failed to add domain:", err);
        setError(err instanceof Error ? err.message : "Failed to add domain");
        toast.error("Failed to add domain", {
          description:
            err instanceof Error ? err.message : "An unexpected error occurred",
        });
        return;
      } finally {
        setIsAddingDomain(false);
      }
    }

    // Reset form
    setNewDomain("");
    setNewDomainIsPrimary(false);
  };

  const handleRemoveDomain = async (
    domainToRemove: TenantDomain,
    index: number
  ) => {
    if (isCreating) {
      // For new tenants, just remove from local state
      setLocalDomains((prev) => prev.filter((_, i) => i !== index));
    } else if (tenantUid && domainToRemove.id > 0) {
      // For existing tenants with real domain IDs, call the API
      if (
        !confirm(
          `Are you sure you want to remove domain "${domainToRemove.domain}"?`
        )
      ) {
        return;
      }

      try {
        setError(null);
        await apiFetch(
          `/api/v1/tenants/${tenantUid}/domains/${domainToRemove.id}`,
          {
            method: "DELETE",
          }
        );

        // Invalidate and refetch domains
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });

        toast.success("Domain removed!", {
          description: `${domainToRemove.domain} has been successfully removed.`,
        });
      } catch (err) {
        console.error("Failed to remove domain:", err);
        setError(
          err instanceof Error ? err.message : "Failed to remove domain"
        );
        toast.error("Failed to remove domain", {
          description:
            err instanceof Error ? err.message : "An unexpected error occurred",
        });
      }
    } else {
      // Fallback for existing tenants with missing domain IDs
      setError(
        "Cannot remove domain: Domain ID is missing. Please refresh the page and try again."
      );
    }
  };

  const handleToggleVerification = async (
    domain: TenantDomain,
    index: number
  ) => {
    if (isCreating) {
      // For new tenants, toggle locally
      setLocalDomains((prev) =>
        prev.map((d, i) =>
          i === index ? { ...d, isVerified: !d.isVerified } : d
        )
      );
    } else if (tenantUid && domain.id > 0) {
      // For existing tenants with real domain IDs, call the API
      try {
        setError(null);
        await apiFetch(`/api/v1/tenants/${tenantUid}/domains/${domain.id}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            isVerified: !domain.isVerified,
          }),
        });

        // Invalidate and refetch domains
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });

        toast.success("Domain verification updated!", {
          description: `${domain.domain} has been successfully ${
            !domain.isVerified ? "verified" : "unverified"
          }.`,
        });
      } catch (err) {
        console.error("Failed to toggle verification:", err);
        setError(
          err instanceof Error
            ? err.message
            : "Failed to update domain verification"
        );
        toast.error("Failed to update domain verification", {
          description:
            err instanceof Error ? err.message : "An unexpected error occurred",
        });
      }
    } else {
      setError(
        "Cannot update verification: Domain ID is missing. Please refresh the page and try again."
      );
    }
  };

  const handleTogglePrimary = async (domain: TenantDomain, index: number) => {
    if (isCreating) {
      // For new tenants, toggle locally (only one can be primary)
      setLocalDomains((prev) =>
        prev.map((d, i) => ({
          ...d,
          isPrimary: i === index ? !d.isPrimary : false,
        }))
      );
    } else if (tenantUid && domain.id > 0) {
      // For existing tenants with real domain IDs, call the API
      try {
        setError(null);
        await apiFetch(`/api/v1/tenants/${tenantUid}/domains/${domain.id}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            isPrimary: !domain.isPrimary,
          }),
        });

        // Invalidate and refetch domains
        queryClient.invalidateQueries({
          queryKey: queryKeys.tenants.domains(tenantUid),
        });

        toast.success("Primary domain updated!", {
          description: `${domain.domain} has been successfully ${
            !domain.isPrimary ? "set as primary" : "removed from primary"
          }.`,
        });
      } catch (err) {
        console.error("Failed to toggle primary status:", err);
        setError(
          err instanceof Error
            ? err.message
            : "Failed to update primary domain status"
        );
        toast.error("Failed to update primary domain status", {
          description:
            err instanceof Error ? err.message : "An unexpected error occurred",
        });
      }
    } else {
      setError(
        "Cannot update primary status: Domain ID is missing. Please refresh the page and try again."
      );
    }
  };

  // Show loading for existing tenants
  if (isLoading && !isCreating) {
    return (
      <div className={`${className} space-y-4`}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium text-muted-800">Domains</h3>
        </div>
        <div className="flex items-center space-x-2 text-sm text-muted-500">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400"></div>
          <span>Loading domains...</span>
        </div>
      </div>
    );
  }

  // Show error for fetch errors
  if (fetchError && !isCreating) {
    return (
      <div className={`${className} space-y-4`}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium text-muted-800">Domains</h3>
        </div>
        <div className="bg-red-50 border border-red-200 rounded-md p-3">
          <p className="text-sm text-red-600">
            Failed to load domains:{" "}
            {fetchError instanceof Error ? fetchError.message : "Unknown error"}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={`${className} space-y-4`}>
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-muted-800">Domains</h3>
        <span className="text-sm text-muted-500">
          {localDomains.length} domain{localDomains.length !== 1 ? "s" : ""}
        </span>
      </div>

      <p className="text-sm text-muted-600">
        Domains allow users to register automatically. When someone registers
        with an email like user@acme.com, they'll be assigned to this tenant if
        acme.com is a verified domain.
      </p>

      {/* Error Message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-3">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      {/* Add New Domain */}
      <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
        <h4 className="text-sm font-medium text-muted-700 mb-3">Add Domain</h4>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div className="md:col-span-2">
            <input
              type="text"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              placeholder="acme.com"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500"
              disabled={isAddingDomain}
            />
          </div>
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="new-domain-primary"
              checked={newDomainIsPrimary}
              onChange={(e) => setNewDomainIsPrimary(e.target.checked)}
              className="h-4 w-4 text-purple-600 focus:ring-purple-500 border-gray-300 rounded"
              disabled={isAddingDomain}
            />
            <label
              htmlFor="new-domain-primary"
              className="text-sm text-muted-700"
            >
              Primary
            </label>
          </div>
        </div>
        <div className="mt-3 flex justify-end">
          <button
            type="button"
            onClick={handleAddDomain}
            disabled={isAddingDomain || !newDomain.trim()}
            className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-white bg-purple-600 border border-transparent rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <PlusIcon className="h-4 w-4 mr-1" />
            {isAddingDomain ? "Adding..." : "Add"}
          </button>
        </div>
      </div>

      {/* Existing Domains */}
      {localDomains.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-muted-700">
            Current Domains
          </h4>
          {localDomains.map((domain, index) => (
            <div
              key={`${domain.domain}-${index}`}
              className="bg-background border border-gray-200 rounded-lg p-3"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span className="font-medium text-muted-900">
                    {domain.domain}
                  </span>
                  {domain.isPrimary && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                      Primary
                    </span>
                  )}
                  <span
                    className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                      domain.isVerified
                        ? "bg-green-100 text-green-800"
                        : "bg-yellow-100 text-yellow-800"
                    }`}
                  >
                    {domain.isVerified ? "Verified" : "Unverified"}
                  </span>
                </div>
                <div className="flex items-center space-x-1">
                  {(isCreating || domain.id > 0) && (
                    <>
                      <button
                        type="button"
                        onClick={() => handleToggleVerification(domain, index)}
                        className={`p-1 text-sm rounded transition-colors ${
                          domain.isVerified
                            ? "text-yellow-600 hover:bg-yellow-50"
                            : "text-green-600 hover:bg-green-50"
                        }`}
                        title={
                          domain.isVerified
                            ? "Mark as unverified"
                            : "Mark as verified"
                        }
                      >
                        {domain.isVerified ? (
                          <CloseIcon className="h-4 w-4" />
                        ) : (
                          <CheckIcon className="h-4 w-4" />
                        )}
                      </button>
                      <button
                        type="button"
                        onClick={() => handleTogglePrimary(domain, index)}
                        className="p-1 text-purple-600 hover:bg-purple-50 rounded transition-colors"
                        title={
                          domain.isPrimary
                            ? "Remove primary status"
                            : "Set as primary"
                        }
                      >
                        <span className="text-xs font-bold">P</span>
                      </button>
                    </>
                  )}
                  <button
                    type="button"
                    onClick={() => handleRemoveDomain(domain, index)}
                    className="p-1 text-red-600 hover:bg-red-50 rounded transition-colors"
                    title="Remove domain"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {localDomains.length === 0 && (
        <div className="text-center py-4 text-muted-500 text-sm">
          No domains configured. Add one above to allow automatic user
          registration.
        </div>
      )}
    </div>
  );
});

TenantDomainsSection.displayName = "TenantDomainsSection";

export default TenantDomainsSection;
