// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState, useCallback, useMemo } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EyeIcon } from "@/components/icons";
import { toast } from "sonner";
import { apiFetch } from "@/utils/fetchUtils";
import {
  useEmailTemplates,
  useDeleteEmailTemplate,
  type EmailTemplateQuery,
} from "../../../../api";
import type { EmailTemplate as NewEmailTemplate } from "../../../../api";
import { EmailTemplateForm } from "../forms/EmailTemplateForm";
import { EmailTemplatePreviewModal } from "./EmailTemplatePreviewModal";
import { useAuth } from "@/auth";
import { isSuperAdmin } from "@/auth";
import {
  AdminEntityPage,
  type AdminEntityPageConfig,
} from "../AdminEntityPage";
import {
  useAdminEntityPage,
  type AdminEntityFilters,
} from "../../hooks/useAdminEntityPage";
import { createStandardAdminColumns } from "../../utils/adminColumnUtils";
import {
  type CommonTemplateFilters,
  initialCommonTemplateFilters,
  createEmailTemplateFilterFields,
} from "../../utils/adminFilters";
// Import the old EmailTemplate type with a different name for the form components
import type { EmailTemplate as OldEmailTemplate } from "../../hooks/useEmailTemplates";
import { FilterField } from "../../../../types/search";

const EmailTemplatesAdminPage: React.FC = () => {
  // Get user context to check if user is superadmin
  const { user } = useAuth();

  // Separate state for template-specific modals
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [selectedTemplateForPreview, setSelectedTemplateForPreview] =
    useState<OldEmailTemplate | null>(null);

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Email Templates",
      entityNameSingular: "Email Template",
      deleteEndpoint: (template: NewEmailTemplate) =>
        `/api/v1/admin/system/email-templates/${template.id}`,
      getEntityDisplayName: (template: NewEmailTemplate) => template.name,
      getEntityId: (template: NewEmailTemplate) => template.id.toString(),
    }),
    []
  );

  // Set up the admin entity page state
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: initialCommonTemplateFilters,
      enableOptimisticUpdates: false,
    },
    () => refetch()
  );

  // Prepare query for the new API hooks
  const templateQuery: EmailTemplateQuery = useMemo(() => {
    const { searchQuery, ...otherFilters } = pageState.filters;
    return {
      page: pageState.pagination.currentPage,
      limit: pageState.pagination.pageSize,
      q: searchQuery,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters, pageState.pagination]);

  // Fetch templates data using the new typed API
  const {
    data: templates,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useEmailTemplates(templateQuery);

  // Delete mutation
  const deleteEmailTemplate = useDeleteEmailTemplate({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete email template:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Convert new EmailTemplate to old format for preview modal compatibility
  const convertToOldFormat = useCallback(
    (template: NewEmailTemplate): OldEmailTemplate => {
      return {
        id: template.id,
        tenantId: template.tenantId,
        tenantName: template.tenantName,
        templateType: template.templateType,
        name: template.name,
        subject: template.subject,
        bodyText: template.textContent || "",
        bodyHtml: template.htmlContent,
        variables: {}, // Empty object as variables are not in new format
        isActive: template.isActive,
        isSystem: template.isSystem,
        createdBy: 0, // Default value as not available in new format
        updatedBy: undefined,
        createdAt: template.createdAt,
        updatedAt: template.updatedAt,
      };
    },
    []
  );

  const handlePreviewTemplate = useCallback(
    (template: NewEmailTemplate) => {
      setSelectedTemplateForPreview(convertToOldFormat(template));
      setIsPreviewModalOpen(true);
    },
    [convertToOldFormat]
  );

  const handleViewTemplate = useCallback(
    (template: NewEmailTemplate) => {
      setSelectedTemplateForPreview(convertToOldFormat(template));
      setIsPreviewModalOpen(true);
    },
    [convertToOldFormat]
  );

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteTemplate = useCallback(
    (template: NewEmailTemplate) => {
      deleteEmailTemplate.mutate(template.id.toString());
    },
    [deleteEmailTemplate]
  );

  const createDefaultTemplates = useCallback(async () => {
    try {
      const response = await apiFetch<{ status: string; message: string }>(
        "/api/v1/admin/system/email-templates/defaults",
        { method: "POST" }
      );

      if (response.status === "success") {
        toast.success("Default templates created successfully");
        refetch();
      }
    } catch (error) {
      console.error("Failed to create default templates:", error);
      toast.error("Failed to create default templates");
    }
  }, [refetch]);

  // Create email template-specific columns using the utility function
  const columns = useMemo(
    () =>
      createStandardAdminColumns<NewEmailTemplate>({
        entityName: "Template",
        titleConfig: {
          accessor: "name",
          header: "Template Name",
          getTitle: (template) => template.name || "Unnamed Template",
          getDescription: (template) => template.subject || null,
        },
        includeCreatedDate: true,
        createdDateConfig: {
          accessor: "createdAt",
        },
        includeId: false,
        customColumns: [
          {
            id: "templateType",
            header: "Template Type",
            cell: ({ row }: { row: { original: NewEmailTemplate } }) => {
              const template = row.original;
              const typeLabel =
                template.templateType
                  ?.replace("_", " ")
                  ?.replace(/\b\w/g, (l) => l.toUpperCase()) || "Unknown";

              return (
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-muted-900">{typeLabel}</span>
                  {template.isSystem && (
                    <Badge variant="secondary" className="text-xs">
                      System
                    </Badge>
                  )}
                </div>
              );
            },
            enableSorting: true,
          },
          {
            id: "tenant",
            header: "Tenant",
            cell: ({ row }: { row: { original: NewEmailTemplate } }) => {
              const template = row.original;
              return (
                <div className="text-sm">
                  <div className="text-muted-900">
                    {template.tenantName || "Unknown Tenant"}
                  </div>
                  <div className="text-muted-500 text-xs">
                    ID: {template.tenantId}
                  </div>
                </div>
              );
            },
            enableSorting: false,
          },
          {
            id: "status",
            header: "Status",
            cell: ({ row }: { row: { original: NewEmailTemplate } }) => {
              const template = row.original;
              const isActive = template.isActive;
              return (
                <Badge
                  variant={isActive ? "default" : "secondary"}
                  className={
                    isActive
                      ? "text-green-700 bg-green-50 border-green-200"
                      : "text-muted-600 bg-gray-50 border-gray-200"
                  }
                >
                  {isActive ? "Active" : "Inactive"}
                </Badge>
              );
            },
            enableSorting: true,
          },
          {
            id: "id",
            header: "ID",
            cell: ({ row }) => {
              const template = row.original;
              const idString = template.id?.toString() || "N/A";
              const displayValue =
                idString.length > 6 ? idString.substring(0, 6) : idString;
              return (
                <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded">
                  {displayValue}
                </span>
              );
            },
            enableSorting: false,
          },
        ],
        customColumnsPosition: "before-actions",
        actionsConfig: {
          onView: handleViewTemplate,
          onEdit: pageState.adminState.handleEditEntity,
          onDelete: handleDeleteTemplate, // Use the enhanced handler
          customActions: [
            {
              label: "Preview",
              onClick: handlePreviewTemplate,
              icon: EyeIcon,
            },
            {
              label: "Copy template ID",
              onClick: (template) =>
                navigator.clipboard.writeText(template.id.toString()),
            },
          ],
        },
      }),
    [
      handleViewTemplate,
      pageState.adminState.handleEditEntity,
      handlePreviewTemplate,
      handleDeleteTemplate,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createEmailTemplateFilterFields({
        filters: pageState.filters,
        updateFilter: pageState.updateFilter,
      }),
    [pageState.filters, pageState.updateFilter]
  );

  // Form components for modals - these still use the old format
  const CreateFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (template: NewEmailTemplate) => void;
      onCancel: () => void;
    }) => (
      <EmailTemplateForm
        entity={null}
        onSuccess={(oldTemplate: OldEmailTemplate) => {
          // Convert from old format back to new format for the success callback
          const newTemplate: NewEmailTemplate = {
            id: oldTemplate.id,
            name: oldTemplate.name,
            subject: oldTemplate.subject,
            templateType: oldTemplate.templateType,
            htmlContent: oldTemplate.bodyHtml,
            textContent: oldTemplate.bodyText || undefined,
            isActive: oldTemplate.isActive,
            isSystem: oldTemplate.isSystem,
            tenantId: oldTemplate.tenantId,
            tenantName: oldTemplate.tenantName,
            createdAt: oldTemplate.createdAt,
            updatedAt: oldTemplate.updatedAt,
          };
          onSuccess(newTemplate);
        }}
        onCancel={onCancel}
      />
    ),
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (template: NewEmailTemplate) => void;
      onCancel: () => void;
    }) => (
      <EmailTemplateForm
        entity={
          pageState.adminState.selectedEntity
            ? convertToOldFormat(pageState.adminState.selectedEntity)
            : null
        }
        onSuccess={(oldTemplate: OldEmailTemplate) => {
          // Convert from old format back to new format for the success callback
          const newTemplate: NewEmailTemplate = {
            id: oldTemplate.id,
            name: oldTemplate.name,
            subject: oldTemplate.subject,
            templateType: oldTemplate.templateType,
            htmlContent: oldTemplate.bodyHtml,
            textContent: oldTemplate.bodyText || undefined,
            isActive: oldTemplate.isActive,
            isSystem: oldTemplate.isSystem,
            tenantId: oldTemplate.tenantId,
            tenantName: oldTemplate.tenantName,
            createdAt: oldTemplate.createdAt,
            updatedAt: oldTemplate.updatedAt,
          };
          onSuccess(newTemplate);
        }}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity, convertToOldFormat]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<
    NewEmailTemplate,
    CommonTemplateFilters
  > = {
    title: "Email Templates",
    description: "Manage system email templates and notifications",
    searchPlaceholder: "Search templates by name, subject, or type...",
    emptyMessage: "No email templates found.",

    entities: templates,
    loading: loading || deleteEmailTemplate.isPending,
    error: error || (deleteEmailTemplate.error?.message ?? null),
    refetch,
    entityConfig: crudConfig,

    pagination: serverPagination
      ? {
          totalPages: serverPagination.totalPages,
          totalItems: serverPagination.total,
        }
      : undefined,

    columns,
    filterFields,

    CreateFormComponent,
    EditFormComponent,

    modalMaxWidth: "2xl",
    deleteDescription: (template) =>
      `Are you sure you want to delete the email template "${template.name}"? This action cannot be undone.`,

    customActions: (
      <>
        {isSuperAdmin(user) && (
          <Button
            onClick={createDefaultTemplates}
            variant="outline"
            className="mr-2"
          >
            Create Default Templates
          </Button>
        )}
      </>
    ),

    enablePagination: true,
  };

  return (
    <>
      <AdminEntityPage config={pageConfig} state={pageState} />

      {/* Template Preview Modal */}
      {selectedTemplateForPreview && (
        <EmailTemplatePreviewModal
          isOpen={isPreviewModalOpen}
          onClose={() => {
            setIsPreviewModalOpen(false);
            setSelectedTemplateForPreview(null);
          }}
          template={selectedTemplateForPreview}
        />
      )}
    </>
  );
};

export default EmailTemplatesAdminPage;
