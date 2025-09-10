// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { apiFetch } from "@/utils/fetchUtils";
import { useAuth } from "@/auth";
import { isSuperAdmin } from "@/auth";
import { EmailTemplate, TemplateVariable } from "../../hooks/useEmailTemplates";

interface EmailTemplateFormProps {
  entity?: EmailTemplate | null;
  onSuccess: (template: EmailTemplate) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export const EmailTemplateForm: React.FC<EmailTemplateFormProps> = ({
  entity,
  onSuccess,
  onCancel,
}) => {
  const { user } = useAuth();
  const [variables, setVariables] = useState<
    Record<string, TemplateVariable[]>
  >({});
  const [activeTab, setActiveTab] = useState<"text" | "html">("text");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    templateType: entity?.templateType || "",
    name: entity?.name || "",
    subject: entity?.subject || "",
    bodyText: entity?.bodyText || "",
    bodyHtml: entity?.bodyHtml || "",
    isActive: entity?.isActive ?? true,
  });

  const templateTypes = [
    { value: "password_reset", label: "Password Reset" },
    { value: "email_verification", label: "Email Verification" },
    { value: "welcome", label: "Welcome" },
  ];

  // Load variables on mount
  useEffect(() => {
    const loadVariables = async () => {
      try {
        const response = await apiFetch<{
          status: string;
          data: Record<string, TemplateVariable[]>;
        }>("/api/v1/admin/system/email-templates/variables");

        if (response.status === "success") {
          setVariables(response.data || {});
        }
      } catch (error) {
        console.error("Failed to load variables:", error);
        setVariables({});
      }
    };

    loadVariables();
  }, []);

  const selectedVariables =
    formData.templateType && variables && variables[formData.templateType]
      ? Array.isArray(variables[formData.templateType])
        ? variables[formData.templateType]
        : []
      : [];

  const insertVariable = (variableName: string) => {
    if (!textAreaRef.current) return;

    const cursorPosition = textAreaRef.current.selectionStart;
    const fieldName = activeTab === "text" ? "bodyText" : "bodyHtml";
    const currentValue = formData[fieldName] || "";

    const newValue =
      currentValue.slice(0, cursorPosition) +
      variableName +
      currentValue.slice(cursorPosition);

    setFormData({ ...formData, [fieldName]: newValue });

    // Focus back to textarea and set cursor position
    setTimeout(() => {
      if (textAreaRef.current) {
        textAreaRef.current.focus();
        textAreaRef.current.setSelectionRange(
          cursorPosition + variableName.length,
          cursorPosition + variableName.length
        );
      }
    }, 0);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.templateType || !formData.name || !formData.subject) {
      toast.error("Please fill in all required fields");
      return;
    }

    setIsSubmitting(true);

    try {
      const templateData = {
        templateType: formData.templateType,
        name: formData.name,
        subject: formData.subject,
        bodyText: formData.bodyText,
        bodyHtml: formData.bodyHtml,
        isActive: formData.isActive,
      };

      const response = await apiFetch<{ status: string; data: EmailTemplate }>(
        entity
          ? `/api/v1/admin/system/email-templates/${entity.id}`
          : "/api/v1/admin/system/email-templates",
        {
          method: entity ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(
            entity
              ? {
                  name: templateData.name,
                  subject: templateData.subject,
                  bodyText: templateData.bodyText,
                  bodyHtml: templateData.bodyHtml,
                  isActive: templateData.isActive,
                }
              : templateData
          ),
        }
      );

      if (response.status === "success") {
        toast.success(
          entity
            ? "Template updated successfully"
            : "Template created successfully"
        );
        onSuccess(response.data);
      }
    } catch (error) {
      console.error("Failed to save template:", error);
      toast.error("Failed to save template");
    } finally {
      setIsSubmitting(false);
    }
  };

  // Check if editing is disabled - system templates can only be edited by superadmins
  const isSystem = entity?.isSystem;
  const canEdit = !isSystem || isSuperAdmin(user);

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Main Form */}
        <div className="lg:col-span-3 space-y-4">
          {isSystem && !isSuperAdmin(user) && (
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
              <p className="text-sm text-amber-800">
                <strong>Note:</strong> This is a system template and is
                read-only. Only superadmins can modify system templates.
              </p>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="templateType">Template Type *</Label>
              <Select
                value={formData.templateType}
                onValueChange={(value) =>
                  setFormData({ ...formData, templateType: value })
                }
                disabled={!!entity}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select template type" />
                </SelectTrigger>
                <SelectContent>
                  {templateTypes.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="name">Template Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="Enter template name"
                disabled={!canEdit}
              />
            </div>
          </div>

          <div>
            <Label htmlFor="subject">Subject *</Label>
            <Input
              id="subject"
              value={formData.subject}
              onChange={(e) =>
                setFormData({ ...formData, subject: e.target.value })
              }
              placeholder="Enter email subject"
              disabled={!canEdit}
            />
          </div>

          {/* Content Tabs */}
          <div>
            <Label className="text-base font-medium">Email Content *</Label>
            {isSystem && !canEdit && (
              <span className="text-sm text-muted-500 ml-2">
                (Read-only for system templates)
              </span>
            )}
            {isSystem && canEdit && (
              <span className="text-sm text-blue-600 ml-2">
                (Superadmin: can edit system template)
              </span>
            )}
            <Tabs
              value={activeTab}
              onValueChange={(value) => setActiveTab(value as "text" | "html")}
              className="mt-2"
            >
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="text">Plain Text</TabsTrigger>
                <TabsTrigger value="html">HTML</TabsTrigger>
              </TabsList>
              <TabsContent value="text" className="mt-4">
                <Textarea
                  ref={(ref) => {
                    if (activeTab === "text") textAreaRef.current = ref;
                  }}
                  value={formData.bodyText}
                  onChange={(e) =>
                    setFormData({ ...formData, bodyText: e.target.value })
                  }
                  placeholder="Enter plain text email body"
                  rows={12}
                  disabled={!canEdit}
                  className="font-mono text-sm"
                />
              </TabsContent>
              <TabsContent value="html" className="mt-4">
                <Textarea
                  ref={(ref) => {
                    if (activeTab === "html") textAreaRef.current = ref;
                  }}
                  value={formData.bodyHtml}
                  onChange={(e) =>
                    setFormData({ ...formData, bodyHtml: e.target.value })
                  }
                  placeholder="Enter HTML email body"
                  rows={12}
                  disabled={!canEdit}
                  className="font-mono text-sm"
                />
              </TabsContent>
            </Tabs>
          </div>

          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="isActive"
              checked={formData.isActive}
              onChange={(e) =>
                setFormData({ ...formData, isActive: e.target.checked })
              }
              disabled={!canEdit}
              className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <Label htmlFor="isActive">Template is active</Label>
          </div>

          {/* Action buttons */}
          <div className="flex justify-end space-x-3 pt-4">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || !canEdit}>
              {isSubmitting
                ? "Saving..."
                : entity
                ? "Update Template"
                : "Create Template"}
            </Button>
          </div>
        </div>

        {/* Variables Sidebar */}
        <div className="lg:col-span-1">
          <Card className="sticky top-4">
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">
                Available Variables
              </CardTitle>
              <CardDescription className="text-xs">
                Click to insert into{" "}
                {activeTab === "text" ? "plain text" : "HTML"} content
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {selectedVariables && selectedVariables.length > 0 ? (
                selectedVariables.map((variable) => (
                  <button
                    key={variable.name}
                    type="button"
                    onClick={() => insertVariable(variable.name)}
                    disabled={!canEdit}
                    className="w-full text-left p-2 rounded border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors text-sm group disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <div className="flex items-center justify-between">
                      <code className="text-blue-600 text-xs group-hover:text-blue-700">
                        {variable.name}
                      </code>
                      {variable.required && (
                        <Badge
                          variant="destructive"
                          className="text-xs px-1 py-0"
                        >
                          Required
                        </Badge>
                      )}
                    </div>
                    <p className="text-muted-500 text-xs mt-1 group-hover:text-muted-600">
                      {variable.description}
                    </p>
                  </button>
                ))
              ) : formData.templateType ? (
                <p className="text-muted-500 text-xs">
                  No variables available for this template type
                </p>
              ) : (
                <p className="text-muted-500 text-xs">
                  Select a template type to see available variables
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </form>
  );
};
