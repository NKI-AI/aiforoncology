// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { FormConfig, FieldConfig, CustomContentSection } from "./FormFactory";
import { Tenant } from "../../hooks/useAdminData";
import TenantDomainsSection, {
  TenantDomainsSectionRef,
} from "../tenants/TenantDomainsSection";
import { apiFetch } from "../../../../utils/fetchUtils";

const tenantFields: FieldConfig[] = [
  {
    name: "name",
    label: "Tenant Name",
    type: "text",
    required: true,
    placeholder: "Enter tenant name",
    validation: ({ value }) =>
      !value?.trim() ? "Tenant name is required" : undefined,
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    placeholder: "Enter tenant description (optional)",
    rows: 3,
  },
];

const tenantCustomContent: CustomContentSection[] = [
  {
    name: "domains",
    component: ({
      entity,
      isEdit,
      isLoading,
      ref,
    }: {
      entity?: Tenant;
      isEdit: boolean;
      isLoading: boolean;
      ref?: React.RefObject<TenantDomainsSectionRef>;
    }) => (
      <TenantDomainsSection
        tenantUid={entity?.tenantUid}
        isCreating={!isEdit}
        className="border-t pt-6"
        ref={ref}
      />
    ),
    position: "after",
    section: "main",
  },
];

export const tenantFormConfig: FormConfig<Tenant> = {
  fields: tenantFields,
  customContent: tenantCustomContent,
  apiEndpoints: {
    create: "/api/v1/tenants",
    update: (tenant: Tenant) => `/api/v1/tenants/${tenant.tenantUid}`,
  },
  entityName: "Tenant",
  themeColor: "purple",
  getDefaultValues: (tenant?: Tenant | null) => ({
    name: tenant?.name ?? "",
    description: tenant?.description ?? "",
  }),
  prepareCreateData: (values) => ({
    name: values.name,
    description: values.description,
  }),
  prepareUpdateData: (values) => ({
    name: values.name,
    description: values.description,
  }),
  getEntityTitle: (tenant) => tenant?.name || "",
  postCreateHook: async (
    tenant: Tenant,
    customContentRefs: Record<string, React.RefObject<any>>
  ) => {
    const domainsRef = customContentRefs[
      "domains"
    ] as React.RefObject<TenantDomainsSectionRef>;

    if (!domainsRef?.current) {
      return;
    }

    const localDomains = domainsRef.current.getLocalDomains();

    if (!tenant.tenantUid) {
      throw new Error("Tenant UID is missing from created tenant");
    }

    // Track successful and failed domain creations
    const results = {
      successful: [] as string[],
      failed: [] as { domain: string; error: string }[],
    };

    // Create each domain for the new tenant
    for (const domain of localDomains) {
      try {
        const response = await apiFetch(
          `/api/v1/tenants/${tenant.tenantUid}/domains`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              domain: domain.domain,
              isPrimary: domain.isPrimary,
            }),
          }
        );

        results.successful.push(domain.domain);
      } catch (error) {
        console.error(`❌ Failed to create domain ${domain.domain}:`, error);
        results.failed.push({
          domain: domain.domain,
          error: error instanceof Error ? error.message : String(error),
        });
      }
    }

    // Clear local domains after processing (even if some failed)
    domainsRef.current.clearLocalDomains();

    // If any domains failed, show a warning but don't fail the entire process
    if (results.failed.length > 0) {
      console.warn(
        `⚠️ ${results.failed.length} domains failed to create:`,
        results.failed
      );
      // Could potentially show a toast notification here
    }
  },
};
