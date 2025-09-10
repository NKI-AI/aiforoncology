// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import { apiFetch } from "@/utils/fetchUtils";
import { EmailTemplate } from "../../hooks/useEmailTemplates";

interface RenderedEmail {
  subject: string;
  bodyText: string;
  bodyHtml: string;
}

interface EmailTemplatePreviewModalProps {
  isOpen: boolean;
  template: EmailTemplate | null;
  onClose: () => void;
}

export const EmailTemplatePreviewModal: React.FC<
  EmailTemplatePreviewModalProps
> = ({ isOpen, template, onClose }) => {
  const [previewData, setPreviewData] = useState<RenderedEmail | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isOpen && template) {
      loadPreview();
    } else {
      setPreviewData(null);
    }
  }, [isOpen, template]);

  const loadPreview = async () => {
    if (!template) return;

    setLoading(true);
    try {
      const response = await apiFetch<{ status: string; data: RenderedEmail }>(
        `/api/v1/admin/system/email-templates/preview?templateType=${template.templateType}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({}), // Use default sample data
        }
      );

      if (response.status === "success") {
        setPreviewData(response.data);
      }
    } catch (error) {
      console.error("Failed to preview template:", error);
      toast.error("Failed to preview template");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Preview: {template?.name}</DialogTitle>
          <DialogDescription>Email preview with sample data</DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
          </div>
        ) : previewData ? (
          <Tabs defaultValue="html" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="html">HTML Preview</TabsTrigger>
              <TabsTrigger value="text">Text Preview</TabsTrigger>
            </TabsList>
            <TabsContent value="html" className="space-y-4">
              <div>
                <Label>Subject</Label>
                <div className="bg-gray-50 p-2 rounded border text-sm">
                  {previewData.subject}
                </div>
              </div>
              <div>
                <Label>HTML Body</Label>
                <div
                  className="bg-background p-4 rounded border max-h-96 overflow-y-auto"
                  dangerouslySetInnerHTML={{ __html: previewData.bodyHtml }}
                />
              </div>
            </TabsContent>
            <TabsContent value="text" className="space-y-4">
              <div>
                <Label>Subject</Label>
                <div className="bg-gray-50 p-2 rounded border text-sm">
                  {previewData.subject}
                </div>
              </div>
              <div>
                <Label>Text Body</Label>
                <div className="bg-gray-50 p-4 rounded border max-h-96 overflow-y-auto whitespace-pre-wrap text-sm">
                  {previewData.bodyText}
                </div>
              </div>
            </TabsContent>
          </Tabs>
        ) : (
          <div className="text-center py-8 text-muted-500">
            Failed to load preview
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
