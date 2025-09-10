// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect } from "react";
import { useForm } from "@tanstack/react-form";
import { toast } from "sonner";
import { apiFetch } from "../../../../utils/fetchUtils";
import { Tenant } from "../../hooks/useAdminData";
import { PlusIcon, CheckIcon, CloseIcon } from "@/components/icons";
import AdminModal from "../AdminModal";
import ModalHeader from "../ModalHeader";
import ModalSection from "../ModalSection";
import { formatDateShort } from "@/utils/format";
import { Button } from "@/components/ui/button";

interface TenantDomain {
  id: number;
  domain: string;
  isVerified: boolean;
  isPrimary: boolean;
  createdAt: string;
  updatedAt: string;
}

interface TenantDomainsModalProps {
  isOpen: boolean;
  tenant: Tenant | null;
  onClose: () => void;
  onSuccess: () => void;
}

interface AddDomainFormData {
  domain: string;
  isPrimary: boolean;
}

const TenantDomainsModal: React.FC<TenantDomainsModalProps> = ({
  isOpen,
  tenant,
  onClose,
  onSuccess,
}) => {
  const [domains, setDomains] = useState<TenantDomain[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm({
    defaultValues: {
      domain: "",
      isPrimary: false,
    } as AddDomainFormData,
    onSubmit: async ({ value }) => {
      setIsSubmitting(true);
      setError(null);

      try {
        await apiFetch(`/api/v1/tenants/${tenant?.tenantUid}/domains`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            domain: value.domain,
            isPrimary: value.isPrimary,
          }),
        });

        form.reset();
        await loadDomains();
        onSuccess();
      } catch (err) {
        console.error("Failed to add domain:", err);
        setError(
          err instanceof Error
            ? err.message
            : "Failed to add domain. Please try again."
        );
      } finally {
        setIsSubmitting(false);
      }
    },
  });

  // Load domains when modal opens
  useEffect(() => {
    if (isOpen && tenant) {
      loadDomains();
    }
  }, [isOpen, tenant]);

  const loadDomains = async () => {
    if (!tenant) return;

    try {
      setLoading(true);
      setError(null);

      const response = await apiFetch<{ domains: TenantDomain[] }>(
        `/api/v1/tenants/${tenant.tenantUid}/domains`
      );
      setDomains(response.domains);
    } catch (err) {
      console.error("Failed to load domains:", err);
      setError("Failed to load domains. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const toggleVerification = async (
    domain: TenantDomain,
    domainIndex: number
  ) => {
    if (!tenant) return;

    try {
      setError(null);

      // Call the API to toggle verification status
      await apiFetch(
        `/api/v1/tenants/${tenant.tenantUid}/domains/${domain.id}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            isVerified: !domain.isVerified,
          }),
        }
      );

      // Update local state optimistically
      setDomains((prev) =>
        prev.map((d, i) =>
          i === domainIndex ? { ...d, isVerified: !d.isVerified } : d
        )
      );

      toast.success(
        `Domain ${domain.isVerified ? "unverified" : "verified"}!`,
        {
          description: `${domain.domain} has been ${
            domain.isVerified ? "marked as unverified" : "verified successfully"
          }.`,
        }
      );
    } catch (err) {
      console.error("Failed to toggle verification:", err);
      setError("Failed to update domain verification.");
      toast.error("Failed to update domain verification", {
        description:
          err instanceof Error ? err.message : "An unexpected error occurred",
      });
    }
  };

  const removeDomain = async (domain: TenantDomain, domainIndex: number) => {
    if (!tenant) return;

    if (
      !confirm(`Are you sure you want to remove domain "${domain.domain}"?`)
    ) {
      return;
    }

    try {
      setError(null);

      // Call the API to remove the domain
      await apiFetch(
        `/api/v1/tenants/${tenant.tenantUid}/domains/${domain.id}`,
        {
          method: "DELETE",
        }
      );

      // Update local state by removing the domain
      setDomains((prev) => prev.filter((_, i) => i !== domainIndex));

      toast.success("Domain removed!", {
        description: `${domain.domain} has been successfully removed from ${tenant.name}.`,
      });
    } catch (err) {
      console.error("Failed to remove domain:", err);
      setError("Failed to remove domain.");
      toast.error("Failed to remove domain", {
        description:
          err instanceof Error ? err.message : "An unexpected error occurred",
      });
    }
  };

  return (
    <AdminModal
      isOpen={isOpen}
      onClose={onClose}
      maxWidth="2xl"
      showHeader={false}
    >
      <div className="space-y-6">
        {/* Modal Header */}
        <ModalHeader
          title={`Manage Domains for ${tenant?.name}`}
          description="Domains allow users to register automatically. When someone registers with an email like user@acme.com, they'll be assigned to this tenant if acme.com is a verified domain."
        />

        {/* Error Message */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-3">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Add New Domain Section */}
        <ModalSection title="Add New Domain" background>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
          >
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Domain Input */}
              <form.Field
                name="domain"
                validators={{
                  onChange: ({ value }) => {
                    if (!value?.trim()) return "Domain is required";
                    if (!/^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(value)) {
                      return "Please enter a valid domain (e.g., acme.com)";
                    }
                    return undefined;
                  },
                }}
              >
                {(field) => (
                  <div className="md:col-span-2">
                    <label
                      htmlFor={field.name}
                      className="block text-sm font-medium text-muted-700 mb-1"
                    >
                      Domain <span className="text-red-500">*</span>
                    </label>
                    <input
                      id={field.name}
                      name={field.name}
                      type="text"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      placeholder="acme.com"
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500 ${
                        field.state.meta.errors.length > 0
                          ? "border-red-300"
                          : "border-gray-300"
                      }`}
                      disabled={isSubmitting}
                    />
                    {field.state.meta.errors.length > 0 && (
                      <p className="mt-1 text-sm text-red-600">
                        {field.state.meta.errors[0]}
                      </p>
                    )}
                  </div>
                )}
              </form.Field>

              {/* Primary Checkbox */}
              <div className="flex items-center space-x-2">
                <form.Field name="isPrimary">
                  {(field) => (
                    <>
                      <input
                        id={field.name}
                        name={field.name}
                        type="checkbox"
                        checked={field.state.value}
                        onChange={(e) => field.handleChange(e.target.checked)}
                        className="h-4 w-4 text-purple-600 focus:ring-purple-500 border-gray-300 rounded"
                        disabled={isSubmitting}
                      />
                      <label
                        htmlFor={field.name}
                        className="text-sm text-muted-700"
                      >
                        Primary domain
                      </label>
                    </>
                  )}
                </form.Field>
              </div>
            </div>

            {/* Submit Button */}
            <div className="mt-4 flex justify-end">
              <Button
                type="submit"
                disabled={isSubmitting || !form.state.canSubmit}
                variant="default"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                {isSubmitting ? "Adding..." : "Add Domain"}
              </Button>
            </div>
          </form>
        </ModalSection>

        {/* Existing Domains Section */}
        <ModalSection title="Existing Domains">
          {loading ? (
            <div className="flex justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-500"></div>
            </div>
          ) : domains.length === 0 ? (
            <div className="text-center py-8 text-muted-500">
              No domains configured for this tenant.
            </div>
          ) : (
            <div className="space-y-3">
              {domains.map((domain, index) => (
                <div
                  key={`${domain.domain}-${index}`}
                  className="bg-background border border-gray-200 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3">
                        <span className="font-medium text-muted-900">
                          {domain.domain}
                        </span>
                        {domain.isPrimary && (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                            Primary
                          </span>
                        )}
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            domain.isVerified
                              ? "bg-green-100 text-green-800"
                              : "bg-yellow-100 text-yellow-800"
                          }`}
                        >
                          {domain.isVerified ? "Verified" : "Unverified"}
                        </span>
                      </div>
                      <p className="text-sm text-muted-500 mt-1">
                        Added on {formatDateShort(domain.createdAt)}
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        onClick={() => toggleVerification(domain, index)}
                        variant="ghost"
                        size="icon"
                        className={
                          domain.isVerified
                            ? "text-yellow-600 hover:text-yellow-900 hover:bg-yellow-50"
                            : "text-green-600 hover:text-green-900 hover:bg-green-50"
                        }
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
                      </Button>
                      <Button
                        onClick={() => removeDomain(domain, index)}
                        variant="ghost"
                        size="icon"
                        className="text-red-600 hover:text-red-900 hover:bg-red-50"
                        title="Remove domain"
                      >
                        <CloseIcon className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ModalSection>

        {/* Footer Actions */}
        <div className="flex justify-end space-x-3 pt-6 border-t">
          <Button type="button" onClick={onClose} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </AdminModal>
  );
};

export default TenantDomainsModal;
