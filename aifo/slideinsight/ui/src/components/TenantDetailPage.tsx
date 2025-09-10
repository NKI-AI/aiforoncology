// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useForm } from "@tanstack/react-form";
import { toast } from "sonner";
import {
  useUpdateTenant,
  useTenantDomains,
  useCreateTenantDomain,
  useDeleteTenantDomain,
  type TenantDomain,
} from "@/api/hooks";
import { apiFetch } from "@/utils/fetchUtils";
import { queryKeys } from "@/utils/apiQueries";
import { formatDate } from "@/utils/format";
import {
  TenantsIcon,
  GlobeIcon,
  CheckIcon,
  CloseIcon,
  EditIcon,
  PlusIcon,
  TrashIcon,
} from "./icons";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { TenantStatusCell } from "./ui/TenantStatusCell";
import { Tenant } from "../features/admin/hooks/useAdminData";
import AdminPageLayout from "../features/admin/components/AdminPageLayout";
import ErrorStateAlert from "./ErrorStateAlert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./AlertDialog";

interface TenantDetailPageProps {
  tenantUid: string;
}

interface AddDomainFormData {
  domain: string;
  isPrimary: boolean;
}

interface EditableTenantField {
  isEditing: boolean;
  value: string;
  originalValue: string;
}

export default function TenantDetailPage({ tenantUid }: TenantDetailPageProps) {
  const [formError, setFormError] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [domainToDelete, setDomainToDelete] = useState<TenantDomain | null>(
    null
  );

  // Inline editing state
  const [editingName, setEditingName] = useState<EditableTenantField>({
    isEditing: false,
    value: "",
    originalValue: "",
  });
  const [editingDescription, setEditingDescription] =
    useState<EditableTenantField>({
      isEditing: false,
      value: "",
      originalValue: "",
    });

  // Fetch tenant details
  const {
    data: tenant,
    isLoading: tenantLoading,
    error: tenantError,
    refetch: refetchTenant,
  } = useQuery({
    queryKey: queryKeys.tenants.detail(tenantUid),
    queryFn: async () => {
      const response = await apiFetch<Tenant>(`/api/v1/tenants/${tenantUid}`);
      // Ensure updatedAt is present for compatibility
      return {
        ...response,
        updatedAt: response.updatedAt || response.createdAt,
      };
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  // Fetch tenant domains using the centralized hook
  const {
    data: domains,
    isLoading: domainsLoading,
    error: domainsError,
    refetch: refetchDomains,
  } = useTenantDomains(tenantUid);

  // Domain mutation hooks
  const createDomainMutation = useCreateTenantDomain(tenantUid, {
    onSuccess: (data, variables) => {
      toast.success("Domain added!", {
        description: `${variables.domain} has been added to ${tenant?.name}.`,
      });
      form.reset();
    },
    onError: (error) => {
      console.error("Failed to add domain:", error);
      const errorMessage =
        error?.message || "Failed to add domain. Please try again.";
      setFormError(errorMessage);
      toast.error("Failed to add domain", {
        description: errorMessage,
      });
    },
  });

  const deleteDomainMutation = useDeleteTenantDomain(tenantUid, {
    onSuccess: () => {
      toast.success("Domain removed!", {
        description: `${domainToDelete?.domain} has been successfully removed from ${tenant?.name}.`,
      });
      setDeleteDialogOpen(false);
      setDomainToDelete(null);
    },
    onError: (error) => {
      console.error("Failed to remove domain:", error);
      const errorMessage = error?.message || "Failed to remove domain.";
      setFormError(errorMessage);
      toast.error("Failed to remove domain", {
        description: errorMessage,
      });
      setDeleteDialogOpen(false);
      setDomainToDelete(null);
    },
  });

  // Tenant update mutation
  const updateTenantMutation = useUpdateTenant(tenantUid, {
    onSuccess: () => {
      toast.success("Tenant updated successfully!");
    },
    onError: (error) => {
      console.error("Failed to update tenant:", error);
      const errorMessage = error?.message || "Failed to update tenant.";
      toast.error("Failed to update tenant", {
        description: errorMessage,
      });
    },
  });

  // Form for adding new domains
  const form = useForm({
    defaultValues: {
      domain: "",
      isPrimary: false,
    } as AddDomainFormData,
    onSubmit: async ({ value }) => {
      setFormError(null);
      createDomainMutation.mutate(value);
    },
  });

  const toggleVerification = async (domain: TenantDomain) => {
    if (!tenant) return;

    try {
      setFormError(null);

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

      await refetchDomains();
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
      const errorMessage =
        err instanceof Error
          ? err.message
          : "Failed to update domain verification.";
      setFormError(errorMessage);
      toast.error("Failed to update domain verification", {
        description: errorMessage,
      });
    }
  };

  const handleDeleteClick = (domain: TenantDomain) => {
    setDomainToDelete(domain);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = () => {
    if (domainToDelete) {
      deleteDomainMutation.mutate(domainToDelete.id);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    setDomainToDelete(null);
  };

  // Inline editing functions
  const startEditing = (
    field: "name" | "description",
    currentValue: string
  ) => {
    const editState = {
      isEditing: true,
      value: currentValue,
      originalValue: currentValue,
    };

    if (field === "name") {
      setEditingName(editState);
    } else {
      setEditingDescription(editState);
    }
  };

  const cancelEditing = (field: "name" | "description") => {
    if (field === "name") {
      setEditingName((prev) => ({
        ...prev,
        isEditing: false,
        value: prev.originalValue,
      }));
    } else {
      setEditingDescription((prev) => ({
        ...prev,
        isEditing: false,
        value: prev.originalValue,
      }));
    }
  };

  const saveTenantField = async (field: "name" | "description") => {
    if (!tenant) return;

    const fieldValue =
      field === "name"
        ? editingName.value.trim()
        : editingDescription.value.trim();

    if (field === "name" && !fieldValue) {
      toast.error("Tenant name cannot be empty");
      return;
    }

    const updateData = {
      [field]: fieldValue || (field === "description" ? null : fieldValue),
    };

    updateTenantMutation.mutate(updateData, {
      onSuccess: () => {
        // Reset editing state
        if (field === "name") {
          setEditingName({
            isEditing: false,
            value: fieldValue,
            originalValue: fieldValue,
          });
        } else {
          setEditingDescription({
            isEditing: false,
            value: fieldValue,
            originalValue: fieldValue,
          });
        }
      },
    });
  };

  const updateFieldValue = (field: "name" | "description", value: string) => {
    if (field === "name") {
      setEditingName((prev) => ({ ...prev, value }));
    } else {
      setEditingDescription((prev) => ({ ...prev, value }));
    }
  };

  // Set initial values when tenant data loads
  React.useEffect(() => {
    if (tenant && !editingName.originalValue) {
      setEditingName({
        isEditing: false,
        value: tenant.name,
        originalValue: tenant.name,
      });
    }
    if (tenant && !editingDescription.originalValue) {
      setEditingDescription({
        isEditing: false,
        value: tenant.description || "",
        originalValue: tenant.description || "",
      });
    }
  }, [tenant, editingName.originalValue, editingDescription.originalValue]);

  // Loading state
  if (tenantLoading) {
    return (
      <AdminPageLayout
        title="Tenant Details"
        description="Loading tenant information"
        actions={
          <Link
            to="/admin/tenants"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Tenants
          </Link>
        }
      >
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </AdminPageLayout>
    );
  }

  // Error state for tenant loading
  if (tenantError || !tenant) {
    return (
      <AdminPageLayout
        title="Tenant Details"
        description="Error loading tenant information"
        actions={
          <Link
            to="/admin/tenants"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Tenants
          </Link>
        }
      >
        <ErrorStateAlert
          error={tenantError || "Failed to load tenant details"}
          title="Error Loading Tenant"
          onRetry={refetchTenant}
          variant="detailed"
        />
      </AdminPageLayout>
    );
  }

  const verifiedDomains = domains?.filter((d) => d.isVerified).length || 0;
  const primaryDomains = domains?.filter((d) => d.isPrimary).length || 0;
  const totalDomains = domains?.length || 0;

  return (
    <AdminPageLayout
      title={tenant.name}
      description={`Tenant UID: ${tenant.tenantUid}`}
      actions={
        <Link
          to="/admin/tenants"
          className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
        >
          ← Back to Tenants
        </Link>
      }
    >
      <div className="bg-background rounded-lg shadow-sm border border-gray-200 p-6">
        {/* Header */}
        <div className="flex items-center space-x-3 mb-6">
          <TenantsIcon className="h-8 w-8 text-purple-500" />
          <div className="flex-1">
            {/* Editable Tenant Name */}
            {editingName.isEditing ? (
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  value={editingName.value}
                  onChange={(e) => updateFieldValue("name", e.target.value)}
                  className="text-2xl font-bold text-muted-900 bg-background border border-gray-300 rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500"
                  placeholder="Tenant name"
                  disabled={updateTenantMutation.isPending}
                />
                <Button
                  onClick={() => saveTenantField("name")}
                  size="sm"
                  disabled={
                    updateTenantMutation.isPending || !editingName.value.trim()
                  }
                  className="h-8 px-3"
                >
                  <CheckIcon className="h-4 w-4" />
                </Button>
                <Button
                  onClick={() => cancelEditing("name")}
                  variant="outline"
                  size="sm"
                  disabled={updateTenantMutation.isPending}
                  className="h-8 px-3"
                >
                  <CloseIcon className="h-4 w-4" />
                </Button>
              </div>
            ) : (
              <div className="flex items-center space-x-2 group">
                <h1 className="text-2xl font-bold text-muted-900">
                  {tenant.name}
                </h1>
                <Button
                  onClick={() => startEditing("name", tenant.name)}
                  variant="ghost"
                  size="sm"
                  className="opacity-0 group-hover:opacity-100 transition-opacity h-8 w-8 p-0"
                >
                  <EditIcon className="h-4 w-4" />
                </Button>
              </div>
            )}
            <p className="text-sm text-muted-600 mt-1">
              Tenant UID: {tenant.tenantUid}
            </p>
          </div>
        </div>

        {/* Tenant Information */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Tenant Information
              </h3>
              <div className="space-y-3">
                <div>
                  <dt className="text-sm font-medium text-muted-500">Name</dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {editingName.isEditing ? (
                      <div className="flex items-center space-x-2">
                        <input
                          type="text"
                          value={editingName.value}
                          onChange={(e) =>
                            updateFieldValue("name", e.target.value)
                          }
                          className="flex-1 bg-background border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500"
                          placeholder="Tenant name"
                          disabled={updateTenantMutation.isPending}
                        />
                        <Button
                          onClick={() => saveTenantField("name")}
                          size="sm"
                          disabled={
                            updateTenantMutation.isPending ||
                            !editingName.value.trim()
                          }
                          className="h-7 px-2"
                        >
                          <CheckIcon className="h-3 w-3" />
                        </Button>
                        <Button
                          onClick={() => cancelEditing("name")}
                          variant="outline"
                          size="sm"
                          disabled={updateTenantMutation.isPending}
                          className="h-7 px-2"
                        >
                          <CloseIcon className="h-3 w-3" />
                        </Button>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-2 group">
                        <span>{tenant.name}</span>
                        <Button
                          onClick={() => startEditing("name", tenant.name)}
                          variant="ghost"
                          size="sm"
                          className="opacity-0 group-hover:opacity-100 transition-opacity h-6 w-6 p-0"
                        >
                          <EditIcon className="h-3 w-3" />
                        </Button>
                      </div>
                    )}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Description
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {editingDescription.isEditing ? (
                      <div className="space-y-2">
                        <textarea
                          value={editingDescription.value}
                          onChange={(e) =>
                            updateFieldValue("description", e.target.value)
                          }
                          className="w-full bg-background border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500 resize-none"
                          placeholder="Tenant description (optional)"
                          rows={3}
                          disabled={updateTenantMutation.isPending}
                        />
                        <div className="flex items-center space-x-2">
                          <Button
                            onClick={() => saveTenantField("description")}
                            size="sm"
                            disabled={updateTenantMutation.isPending}
                            className="h-7 px-3"
                          >
                            <CheckIcon className="h-3 w-3 mr-1" />
                            Save
                          </Button>
                          <Button
                            onClick={() => cancelEditing("description")}
                            variant="outline"
                            size="sm"
                            disabled={updateTenantMutation.isPending}
                            className="h-7 px-3"
                          >
                            <CloseIcon className="h-3 w-3 mr-1" />
                            Cancel
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <div className="group">
                        <div className="flex items-start space-x-2">
                          <span className="flex-1">
                            {tenant.description || "No description provided"}
                          </span>
                          <Button
                            onClick={() =>
                              startEditing(
                                "description",
                                tenant.description || ""
                              )
                            }
                            variant="ghost"
                            size="sm"
                            className="opacity-0 group-hover:opacity-100 transition-opacity h-6 w-6 p-0 flex-shrink-0"
                          >
                            <EditIcon className="h-3 w-3" />
                          </Button>
                        </div>
                      </div>
                    )}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">Status</dt>
                  <dd className="mt-1">
                    <TenantStatusCell tenant={tenant} />
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Created
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {formatDate(tenant.createdAt)}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Tenant UID
                  </dt>
                  <dd className="text-sm text-muted-900 font-mono mt-1">
                    {tenant.tenantUid}
                  </dd>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Domain Statistics
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
                  <div className="flex items-center">
                    <GlobeIcon className="h-5 w-5 text-primary mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-primary">
                        {totalDomains}
                      </p>
                      <p className="text-sm text-primary/80">Total Domains</p>
                    </div>
                  </div>
                </div>
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <CheckIcon className="h-5 w-5 text-green-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-green-900">
                        {verifiedDomains}
                      </p>
                      <p className="text-sm text-green-700">Verified</p>
                    </div>
                  </div>
                </div>
                <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <span className="text-orange-500 mr-2 text-lg">★</span>
                    <div>
                      <p className="text-2xl font-bold text-orange-900">
                        {primaryDomains}
                      </p>
                      <p className="text-sm text-orange-700">Primary</p>
                    </div>
                  </div>
                </div>
                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <EditIcon className="h-5 w-5 text-muted-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-muted-900">
                        {totalDomains - verifiedDomains}
                      </p>
                      <p className="text-sm text-muted-700">Unverified</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Add Domain Form */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Add New Domain
          </h3>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <p className="text-sm text-muted-600 mb-4">
              Domains allow users to register automatically. When someone
              registers with an email like user@acme.com, they'll be assigned to
              this tenant if acme.com is a verified domain.
            </p>

            {formError && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3 mb-4">
                <p className="text-sm text-red-600">{formError}</p>
              </div>
            )}

            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
              }}
            >
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
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
                        disabled={createDomainMutation.isPending}
                      />
                      {field.state.meta.errors.length > 0 && (
                        <p className="mt-1 text-sm text-red-600">
                          {field.state.meta.errors[0]}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>

                <form.Field name="isPrimary">
                  {(field) => (
                    <div className="flex items-center space-x-2 mt-6">
                      <input
                        id={field.name}
                        name={field.name}
                        type="checkbox"
                        checked={field.state.value}
                        onChange={(e) => field.handleChange(e.target.checked)}
                        className="h-4 w-4 text-purple-600 focus:ring-purple-500 border-gray-300 rounded"
                        disabled={createDomainMutation.isPending}
                      />
                      <label
                        htmlFor={field.name}
                        className="text-sm font-medium text-muted-700"
                      >
                        Primary Domain
                      </label>
                    </div>
                  )}
                </form.Field>
              </div>

              <div className="mt-4 flex justify-end">
                <Button
                  type="submit"
                  disabled={
                    createDomainMutation.isPending || !form.state.canSubmit
                  }
                  className="inline-flex items-center"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  {createDomainMutation.isPending ? "Adding..." : "Add Domain"}
                </Button>
              </div>
            </form>
          </div>
        </div>

        {/* Domains List */}
        <div>
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Existing Domains
          </h3>

          {domainsLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="animate-pulse bg-gray-100 h-16 rounded-lg"
                ></div>
              ))}
            </div>
          ) : domainsError ? (
            <ErrorStateAlert
              error={domainsError}
              title="Failed to load domains"
              onRetry={refetchDomains}
              variant="inline"
            />
          ) : !domains || domains.length === 0 ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
              <GlobeIcon className="h-12 w-12 text-muted-400 mx-auto mb-4" />
              <h4 className="text-lg font-medium text-muted-900 mb-2">
                No domains configured
              </h4>
              <p className="text-muted-500">
                Use the form above to add your first domain.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {domains.map((domain, index) => (
                <div
                  key={`${domain.domain}-${index}`}
                  className="bg-background border border-gray-200 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <GlobeIcon className="h-5 w-5 text-muted-400" />
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-muted-900">
                            {domain.domain}
                          </span>
                          {domain.isPrimary && (
                            <Badge
                              variant="secondary"
                              className="bg-purple-100 text-purple-800"
                            >
                              Primary
                            </Badge>
                          )}
                          <Badge
                            variant={domain.isVerified ? "default" : "outline"}
                            className={
                              domain.isVerified
                                ? "bg-green-100 text-green-800 border-green-200"
                                : "border-yellow-300 text-yellow-700"
                            }
                          >
                            {domain.isVerified ? "Verified" : "Unverified"}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-500 mt-1">
                          Added {formatDate(domain.createdAt)}
                          {domain.updatedAt !== domain.createdAt && (
                            <span>
                              {" "}
                              • Updated {formatDate(domain.updatedAt)}
                            </span>
                          )}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        onClick={() => toggleVerification(domain)}
                        variant="outline"
                        size="sm"
                        className={
                          domain.isVerified
                            ? "border-green-300 text-green-700 hover:bg-green-50"
                            : "border-yellow-300 text-yellow-700 hover:bg-yellow-50"
                        }
                      >
                        {domain.isVerified ? (
                          <>
                            <CloseIcon className="h-4 w-4 mr-1" />
                            Unverify
                          </>
                        ) : (
                          <>
                            <CheckIcon className="h-4 w-4 mr-1" />
                            Verify
                          </>
                        )}
                      </Button>
                      <Button
                        onClick={() => handleDeleteClick(domain)}
                        variant="outline"
                        size="sm"
                        className="border-red-300 text-red-700 hover:bg-red-50"
                      >
                        <TrashIcon className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Domain</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove the domain "
              {domainToDelete?.domain}"? This action cannot be undone and will
              prevent users with email addresses from this domain from
              automatically registering to this tenant.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleDeleteCancel}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
            >
              Remove Domain
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AdminPageLayout>
  );
}
