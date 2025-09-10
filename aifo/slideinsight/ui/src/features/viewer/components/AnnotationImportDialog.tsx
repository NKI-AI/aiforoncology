// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Loader2, CheckCircle, AlertCircle, Info } from "lucide-react";
import {
  vectorAnnotationsService,
  type VectorAnnotation,
  type AnnotationImportResult,
} from "@/services/vectorAnnotationsService";

interface AnnotationImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  slideUid: string;
  onImportComplete?: (result: AnnotationImportResult) => void;
  studyAnnotationSettings?: any; // Study annotation settings with allowed labels
}

export default function AnnotationImportDialog({
  open,
  onOpenChange,
  slideUid,
  onImportComplete,
  studyAnnotationSettings,
}: AnnotationImportDialogProps) {
  const [vectorAnnotations, setVectorAnnotations] = useState<
    VectorAnnotation[]
  >([]);
  const [loading, setLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [selectedVectorUid, setSelectedVectorUid] = useState<string | null>(
    null
  );
  const [importResult, setImportResult] =
    useState<AnnotationImportResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Load vector annotations when dialog opens
  useEffect(() => {
    if (open && slideUid) {
      loadVectorAnnotations();
    }
  }, [open, slideUid]);

  const loadVectorAnnotations = async () => {
    setLoading(true);
    setError(null);
    try {
      const annotations =
        await vectorAnnotationsService.getVectorAnnotationsForSlide(slideUid);
      setVectorAnnotations(annotations);
    } catch (err) {
      console.error("Failed to load vector annotations:", err);
      setError("Failed to load vector annotations. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleImport = async () => {
    if (!selectedVectorUid) return;

    setImporting(true);
    setError(null);
    try {
      // Convert study annotation settings to the format expected by the API
      const allowedLabels = studyAnnotationSettings?.annotations?.map(
        (annotation: any) => ({
          label: annotation.label,
          type: annotation.type,
          color: annotation.color,
        })
      ) || [
        // Fallback labels if study settings are not available
        { label: "roi", type: "polygon", color: "#ff0000" },
        { label: "tumor", type: "polygon", color: "#00ff00" },
        { label: "stroma", type: "polygon", color: "#0000ff" },
        { label: "artifact", type: "polygon", color: "#ffff00" },
      ];

      const result =
        await vectorAnnotationsService.importVectorAnnotationToWorkspace(
          slideUid,
          selectedVectorUid,
          { allowedLabels }
        );
      setImportResult(result);
      onImportComplete?.(result);
    } catch (err) {
      console.error("Failed to import vector annotation:", err);
      setError("Failed to import vector annotation. Please try again.");
    } finally {
      setImporting(false);
    }
  };

  const handleClose = () => {
    setSelectedVectorUid(null);
    setImportResult(null);
    setError(null);
    onOpenChange(false);
  };

  const selectedAnnotation = vectorAnnotations.find(
    (ann) => ann.vectorUid === selectedVectorUid
  );

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Import Vector Annotations</DialogTitle>
          <DialogDescription>
            Select a vector annotation to import as workspace annotations. Only
            labels that match your study settings will be imported.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-md">
              <AlertCircle className="h-4 w-4 text-red-500" />
              <span className="text-sm text-red-700">{error}</span>
            </div>
          )}

          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin" />
              <span className="ml-2">Loading vector annotations...</span>
            </div>
          ) : vectorAnnotations.length === 0 ? (
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <Info className="h-5 w-5 mr-2" />
              No vector annotations found for this slide.
            </div>
          ) : (
            <div className="space-y-3">
              <h4 className="text-sm font-medium">
                Available Vector Annotations:
              </h4>
              <div className="h-48 border rounded-md p-3 overflow-y-auto">
                <div className="space-y-2">
                  {vectorAnnotations.map((annotation) => (
                    <div
                      key={annotation.vectorUid}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                        selectedVectorUid === annotation.vectorUid
                          ? "border-blue-500 bg-blue-50"
                          : "border-gray-200 hover:border-gray-300 hover:bg-gray-50"
                      }`}
                      onClick={() => setSelectedVectorUid(annotation.vectorUid)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="space-y-1">
                          <div className="font-medium text-sm">
                            {annotation.vectorName ||
                              `Vector ${annotation.vectorUid}`}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            ID: {annotation.vectorUid}
                          </div>
                          {annotation.actorType && (
                            <Badge variant="outline" className="text-xs">
                              {annotation.actorType}
                            </Badge>
                          )}
                        </div>
                        {annotation.labels && annotation.labels.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {annotation.labels.slice(0, 3).map((label, idx) => (
                              <Badge
                                key={idx}
                                variant="secondary"
                                className="text-xs"
                                style={{
                                  backgroundColor: `${label.color}20`,
                                  color: label.color,
                                }}
                              >
                                {label.name}
                              </Badge>
                            ))}
                            {annotation.labels.length > 3 && (
                              <Badge variant="secondary" className="text-xs">
                                +{annotation.labels.length - 3}
                              </Badge>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {importResult && (
            <div className="space-y-3 p-4 bg-green-50 border border-green-200 rounded-md">
              <div className="flex items-center gap-2">
                <CheckCircle className="h-5 w-5 text-green-500" />
                <h4 className="font-medium text-green-800">
                  Import Successful!
                </h4>
              </div>
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div>
                  <div className="text-green-700 font-medium">Imported</div>
                  <div className="text-green-600">
                    {importResult.importedCount} annotations
                  </div>
                </div>
                <div>
                  <div className="text-yellow-700 font-medium">Skipped</div>
                  <div className="text-yellow-600">
                    {importResult.skippedCount} annotations
                  </div>
                </div>
                <div>
                  <div className="text-blue-700 font-medium">Overwritten</div>
                  <div className="text-blue-600">
                    {importResult.overwrittenCount} annotations
                  </div>
                </div>
              </div>
              {importResult.skippedLabels.length > 0 && (
                <div className="text-sm">
                  <div className="text-yellow-700 font-medium">
                    Skipped labels:
                  </div>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {importResult.skippedLabels.map((label) => (
                      <Badge key={label} variant="outline" className="text-xs">
                        {label}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            {importResult ? "Close" : "Cancel"}
          </Button>
          {!importResult && (
            <Button
              onClick={handleImport}
              disabled={!selectedVectorUid || importing || loading}
            >
              {importing ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  Importing...
                </>
              ) : (
                "Import Annotations"
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
