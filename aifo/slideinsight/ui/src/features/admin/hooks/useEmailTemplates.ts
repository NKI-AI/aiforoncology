// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useEffect } from "react";
import { apiFetch } from "../../../utils/fetchUtils";
import { toast } from "sonner";

interface EmailTemplate {
  id: number;
  tenantId: number;
  tenantName?: string;
  templateType: string;
  name: string;
  subject: string;
  bodyText: string;
  bodyHtml: string;
  variables: Record<string, any>;
  isActive: boolean;
  isSystem: boolean;
  createdBy: number;
  updatedBy?: number;
  createdAt: string;
  updatedAt: string;
}

interface TemplateVariable {
  name: string;
  description: string;
  required: boolean;
  type: string;
}

interface PaginationInfo {
  total: number;
  totalPages: number;
  currentPage: number;
  limit: number;
  offset: number;
}

interface UseEmailTemplatesParams {
  page?: number;
  limit?: number;
  q?: string;
}

interface UseEmailTemplatesResult {
  templates: EmailTemplate[];
  variables: Record<string, TemplateVariable[]>;
  pagination: PaginationInfo | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useEmailTemplates(
  params: UseEmailTemplatesParams = {}
): UseEmailTemplatesResult {
  const [templates, setTemplates] = useState<EmailTemplate[]>([]);
  const [variables, setVariables] = useState<
    Record<string, TemplateVariable[]>
  >({});
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { page = 1, limit = 20, q = "" } = params;

  const loadTemplates = async () => {
    try {
      setLoading(true);
      setError(null);

      const offset = (page - 1) * limit;
      const queryParams = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString(),
      });

      if (q) {
        queryParams.append("q", q);
      }

      const response = await apiFetch<{
        status: string;
        data: { templates: EmailTemplate[]; pagination: PaginationInfo };
      }>(`/api/v1/admin/system/email-templates?${queryParams.toString()}`);

      if (response.status === "success") {
        setTemplates(response.data.templates || []);
        setPagination(response.data.pagination);
      } else {
        setError("Failed to load email templates");
        setTemplates([]);
        setPagination(null);
      }
    } catch (error) {
      console.error("Failed to load templates:", error);
      setError(
        error instanceof Error
          ? error.message
          : "Failed to load email templates"
      );
      setTemplates([]);
      setPagination(null);
    } finally {
      setLoading(false);
    }
  };

  const loadVariables = async () => {
    try {
      const response = await apiFetch<{
        status: string;
        data: Record<string, TemplateVariable[]>;
      }>("/api/v1/admin/system/email-templates/variables");

      if (response.status === "success") {
        setVariables(response.data || {});
      } else {
        setVariables({});
      }
    } catch (error) {
      console.error("Failed to load variables:", error);
      setVariables({});
    }
  };

  useEffect(() => {
    loadTemplates();
    loadVariables();
  }, [page, limit, q]);

  const refetch = () => {
    loadTemplates();
    loadVariables();
  };

  return {
    templates,
    variables,
    pagination,
    loading,
    error,
    refetch,
  };
}

export type { EmailTemplate, TemplateVariable };
