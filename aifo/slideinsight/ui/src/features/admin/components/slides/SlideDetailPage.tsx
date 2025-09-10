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
import { apiFetch } from "@/utils/fetchUtils";
import { queryKeys } from "@/utils/apiQueries";
import { formatDate } from "@/utils/format";
import { copyToClipboard } from "@/utils/clipboardUtils";
import {
  ImageIcon,
  ImageIcon as PhotoIcon,
  EyeIcon,
  CheckIcon,
  CloseIcon,
  EditIcon,
  PlusIcon,
  TrashIcon,
  CpuChipIcon,
} from "@/components/icons";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SlideWithCount } from "@/hooks/useSlides";
import AdminPageLayout from "@/features/admin/components/AdminPageLayout";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/AlertDialog";

interface SlideDetailPageProps {
  slideUid: string;
}

interface SlideMetadata {
  slideUid: string;
  slideName: string;
  minLevel: number;
  maxLevel: number;
  tileSize: number;
  format: string;
  slideMpp: number;
  slideWidth: number;
  slideHeight: number;
  vendor: string;
  magnification: string;
}

interface Mask {
  maskUid: string;
  maskName: string;
  tilesUrl: string;
  slideUid: string;
  labels?: any;
  maskWidth: number;
  maskHeight: number;
  maskMpp: number;
  createdAt: string;
}

interface MaskList {
  slideUid: string;
  masks: Mask[];
}

interface VectorAnnotation {
  vectorUid: string;
  vectorName: string;
  slideUid: string;
  fileUri: string;
  labels?: Array<{
    name: string;
    color: string;
  }>;
  deletedAt?: string;
  deletedBy?: number;
  createdAt: string;
}

interface VectorAnnotationList {
  slideUid: string;
  annotations: VectorAnnotation[];
}

interface SlideAnnotationMetadata {
  slideUid: string;
  rasterUrl: string;
  vectorUrl: string;
  rasterCount: number;
  vectorCount: number;
}

interface AddMaskFormData {
  maskName: string;
  maskUri: string;
}

interface AddVectorFormData {
  vectorName: string;
  fileUri: string;
}

interface EditableSlideField {
  isEditing: boolean;
  value: string;
  originalValue: string;
}

export default function SlideDetailPage({ slideUid }: SlideDetailPageProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [maskToDelete, setMaskToDelete] = useState<Mask | null>(null);
  const [vectorDeleteDialogOpen, setVectorDeleteDialogOpen] = useState(false);
  const [vectorToDelete, setVectorToDelete] = useState<VectorAnnotation | null>(
    null
  );

  // Inline editing state
  const [editingName, setEditingName] = useState<EditableSlideField>({
    isEditing: false,
    value: "",
    originalValue: "",
  });
  const [isSavingSlide, setIsSavingSlide] = useState(false);

  // Fetch slide details
  const {
    data: slide,
    isLoading: slideLoading,
    error: slideError,
    refetch: refetchSlide,
  } = useQuery({
    queryKey: queryKeys.slides.detail(slideUid),
    queryFn: async () => {
      const response = await apiFetch<SlideWithCount>(
        `/api/v1/slides/${slideUid}`
      );
      return response;
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  // Fetch slide metadata (technical details)
  const {
    data: slideMetadata,
    isLoading: metadataLoading,
    error: metadataError,
  } = useQuery({
    queryKey: ["slide", "metadata", slideUid],
    queryFn: async () => {
      const response = await apiFetch<SlideMetadata>(
        `/api/v1/slides/${slideUid}/metadata`
      );
      return response;
    },
    staleTime: 60 * 1000, // Metadata changes rarely
    gcTime: 10 * 60 * 1000,
  });

  // Fetch slide annotations overview
  const {
    data: annotationsOverview,
    isLoading: annotationsLoading,
    error: annotationsError,
    refetch: refetchAnnotations,
  } = useQuery({
    queryKey: ["slide", "annotations", slideUid],
    queryFn: async () => {
      const response = await apiFetch<SlideAnnotationMetadata>(
        `/api/v1/slides/${slideUid}/annotations`
      );
      return response;
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  // Fetch slide masks
  const {
    data: masksData,
    isLoading: masksLoading,
    error: masksError,
    refetch: refetchMasks,
  } = useQuery({
    queryKey: ["slide", "masks", slideUid],
    queryFn: async () => {
      const response = await apiFetch<MaskList>(
        `/api/v1/slides/${slideUid}/annotations/raster`
      );
      return response.masks || [];
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  // Fetch slide vector annotations
  const {
    data: vectorsData,
    isLoading: vectorsLoading,
    error: vectorsError,
    refetch: refetchVectors,
  } = useQuery({
    queryKey: ["slide", "vectors", slideUid],
    queryFn: async () => {
      const response = await apiFetch<VectorAnnotationList>(
        `/api/v1/slides/${slideUid}/annotations/vector`
      );
      return response.annotations || [];
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  const masks = masksData || [];
  const vectors = vectorsData || [];

  // Form for adding new masks
  const form = useForm({
    defaultValues: {
      maskName: "",
      maskUri: "",
    } as AddMaskFormData,
    onSubmit: async ({ value }) => {
      setIsSubmitting(true);
      setFormError(null);

      try {
        await apiFetch(`/api/v1/slides/${slideUid}/annotations/raster`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            maskName: value.maskName,
            maskUri: value.maskUri,
            slideUid: slideUid,
          }),
        });

        form.reset();
        await refetchMasks();
        await refetchAnnotations();
        toast.success("Mask added!", {
          description: `${value.maskName} has been added to the slide.`,
        });
      } catch (err) {
        console.error("Failed to add mask:", err);
        const errorMessage =
          err instanceof Error
            ? err.message
            : "Failed to add mask. Please try again.";
        setFormError(errorMessage);
        toast.error("Failed to add mask", {
          description: errorMessage,
        });
      } finally {
        setIsSubmitting(false);
      }
    },
  });

  // Form for adding new vector annotations
  const vectorForm = useForm({
    defaultValues: {
      vectorName: "",
      fileUri: "",
    } as AddVectorFormData,
    onSubmit: async ({ value }) => {
      setIsSubmitting(true);
      setFormError(null);

      try {
        await apiFetch(`/api/v1/slides/${slideUid}/annotations/vector`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            vectorName: value.vectorName,
            fileUri: value.fileUri,
            slideUid: slideUid,
          }),
        });

        vectorForm.reset();
        await refetchVectors();
        await refetchAnnotations();
        toast.success("Vector annotation added!", {
          description: `${value.vectorName} has been added to the slide.`,
        });
      } catch (err) {
        console.error("Failed to add vector annotation:", err);
        const errorMessage =
          err instanceof Error
            ? err.message
            : "Failed to add vector annotation. Please try again.";
        setFormError(errorMessage);
        toast.error("Failed to add vector annotation", {
          description: errorMessage,
        });
      } finally {
        setIsSubmitting(false);
      }
    },
  });

  const removeMask = async (mask: Mask) => {
    if (!slide) return;

    try {
      setFormError(null);

      await apiFetch(
        `/api/v1/slides/${slideUid}/annotations/raster/${mask.maskUid}`,
        {
          method: "DELETE",
        }
      );

      await refetchMasks();
      await refetchAnnotations();
      toast.success("Mask removed!", {
        description: `${mask.maskName} has been successfully removed.`,
      });
    } catch (err) {
      console.error("Failed to remove mask:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to remove mask.";
      setFormError(errorMessage);
      toast.error("Failed to remove mask", {
        description: errorMessage,
      });
    } finally {
      setDeleteDialogOpen(false);
      setMaskToDelete(null);
    }
  };

  const removeVector = async (vector: VectorAnnotation) => {
    if (!slide) return;

    try {
      setFormError(null);

      await apiFetch(
        `/api/v1/slides/${slideUid}/annotations/vector/${vector.vectorUid}`,
        {
          method: "DELETE",
        }
      );

      await refetchVectors();
      await refetchAnnotations();
      toast.success("Vector annotation removed!", {
        description: `${vector.vectorName} has been successfully removed.`,
      });
    } catch (err) {
      console.error("Failed to remove vector annotation:", err);
      const errorMessage =
        err instanceof Error
          ? err.message
          : "Failed to remove vector annotation.";
      setFormError(errorMessage);
      toast.error("Failed to remove vector annotation", {
        description: errorMessage,
      });
    } finally {
      setVectorDeleteDialogOpen(false);
      setVectorToDelete(null);
    }
  };

  const handleDeleteClick = (mask: Mask) => {
    setMaskToDelete(mask);
    setDeleteDialogOpen(true);
  };

  const handleVectorDeleteClick = (vector: VectorAnnotation) => {
    setVectorToDelete(vector);
    setVectorDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = () => {
    if (maskToDelete) {
      removeMask(maskToDelete);
    }
  };

  const handleVectorDeleteConfirm = () => {
    if (vectorToDelete) {
      removeVector(vectorToDelete);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    setMaskToDelete(null);
  };

  const handleVectorDeleteCancel = () => {
    setVectorDeleteDialogOpen(false);
    setVectorToDelete(null);
  };

  const handleViewSlide = () => {
    // Navigate to slide viewer
    window.open(`/i/${slideUid}`, "_blank");
  };

  // Inline editing functions
  const startEditing = (field: "name", currentValue: string) => {
    const editState = {
      isEditing: true,
      value: currentValue,
      originalValue: currentValue,
    };

    if (field === "name") {
      setEditingName(editState);
    }
  };

  const cancelEditing = (field: "name") => {
    if (field === "name") {
      setEditingName((prev) => ({
        ...prev,
        isEditing: false,
        value: prev.originalValue,
      }));
    }
  };

  const saveSlideField = async (field: "name") => {
    if (!slide) return;

    const fieldValue = editingName.value.trim();

    if (field === "name" && !fieldValue) {
      toast.error("Slide name cannot be empty");
      return;
    }

    setIsSavingSlide(true);
    try {
      const updateData = {
        [field === "name" ? "slideName" : field]: fieldValue,
      };

      await apiFetch(`/api/v1/slides/${slide.slideUid}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateData),
      });

      await refetchSlide();

      // Reset editing state
      if (field === "name") {
        setEditingName({
          isEditing: false,
          value: fieldValue,
          originalValue: fieldValue,
        });
      }

      toast.success(`Slide ${field} updated successfully!`);
    } catch (err) {
      console.error(`Failed to update slide ${field}:`, err);
      const errorMessage =
        err instanceof Error ? err.message : `Failed to update slide ${field}.`;
      toast.error(`Failed to update slide ${field}`, {
        description: errorMessage,
      });
    } finally {
      setIsSavingSlide(false);
    }
  };

  const updateFieldValue = (field: "name", value: string) => {
    if (field === "name") {
      setEditingName((prev) => ({ ...prev, value }));
    }
  };

  // Set initial values when slide data loads
  React.useEffect(() => {
    if (slide && !editingName.originalValue) {
      setEditingName({
        isEditing: false,
        value: slide.slideName || "",
        originalValue: slide.slideName || "",
      });
    }
  }, [slide, editingName.originalValue]);

  // Loading state
  if (slideLoading) {
    return (
      <AdminPageLayout
        title="Slide Details"
        description="Loading slide information"
        actions={
          <Link
            to="/admin/slides"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Slides
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

  // Error state for slide loading
  if (slideError || !slide) {
    return (
      <AdminPageLayout
        title="Slide Details"
        description="Error loading slide information"
        actions={
          <Link
            to="/admin/slides"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Slides
          </Link>
        }
      >
        <ErrorStateAlert
          error={slideError || "Failed to load slide details"}
          title="Error Loading Slide"
          onRetry={refetchSlide}
          variant="detailed"
        />
      </AdminPageLayout>
    );
  }

  const rasterCount = annotationsOverview?.rasterCount || 0;
  const vectorCount = annotationsOverview?.vectorCount || 0;
  const totalAnnotations = rasterCount + vectorCount;

  // Format resolution display
  const formatResolution = (width?: number, height?: number, mpp?: number) => {
    if (!width || !height) return "Unknown";
    const megapixels = (width * height) / 1000000;
    const mppText = mpp ? ` (${mpp.toFixed(2)} μm/px)` : "";
    return `${width} × ${height} (${megapixels.toFixed(1)}MP)${mppText}`;
  };

  return (
    <AdminPageLayout
      title={slide.slideName || "Unnamed Slide"}
      description={`Slide UID: ${slide.slideUid}`}
      actions={
        <div className="flex items-center space-x-2">
          <Button
            onClick={handleViewSlide}
            className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-md transition"
          >
            <EyeIcon className="h-4 w-4 mr-2" />
            View Slide
          </Button>
          <Link
            to="/admin/slides"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            ← Back to Slides
          </Link>
        </div>
      }
    >
      <div className="bg-background rounded-lg shadow-sm border border-gray-200 p-6">
        {/* Header */}
        <div className="flex items-center space-x-3 mb-6">
          <ImageIcon className="h-8 w-8 text-green-500" />
          <div className="flex-1">
            {/* Editable Slide Name */}
            {editingName.isEditing ? (
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  value={editingName.value}
                  onChange={(e) => updateFieldValue("name", e.target.value)}
                  className="text-2xl font-bold text-muted-900 bg-background border border-gray-300 rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500"
                  placeholder="Slide name"
                  disabled={isSavingSlide}
                />
                <Button
                  onClick={() => saveSlideField("name")}
                  size="sm"
                  disabled={isSavingSlide || !editingName.value.trim()}
                  className="h-8 px-3"
                >
                  <CheckIcon className="h-4 w-4" />
                </Button>
                <Button
                  onClick={() => cancelEditing("name")}
                  variant="outline"
                  size="sm"
                  disabled={isSavingSlide}
                  className="h-8 px-3"
                >
                  <CloseIcon className="h-4 w-4" />
                </Button>
              </div>
            ) : (
              <div className="flex items-center space-x-2 group">
                <h1 className="text-2xl font-bold text-muted-900">
                  {slide.slideName || "Unnamed Slide"}
                </h1>
                <Button
                  onClick={() => startEditing("name", slide.slideName || "")}
                  variant="ghost"
                  size="sm"
                  className="opacity-0 group-hover:opacity-100 transition-opacity h-8 w-8 p-0"
                >
                  <EditIcon className="h-4 w-4" />
                </Button>
              </div>
            )}
            <p className="text-sm text-muted-600 mt-1">
              Slide UID: {slide.slideUid}
            </p>
          </div>
        </div>

        {/* Slide Information and Statistics */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Slide Information
              </h3>
              <div className="space-y-3">
                <div>
                  <dt className="text-sm font-medium text-muted-500">Name</dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {slide.slideName || "Unnamed Slide"}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Resolution
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {formatResolution(
                      slide.slideWidth,
                      slide.slideHeight,
                      slide.slideMpp
                    )}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Created
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {(slide as any).createdAt
                      ? formatDate((slide as any).createdAt)
                      : "Unknown"}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Updated
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {(slide as any).updatedAt
                      ? formatDate((slide as any).updatedAt)
                      : "Unknown"}
                  </dd>
                </div>
                {slideMetadata && !metadataLoading && (
                  <>
                    <div>
                      <dt className="text-sm font-medium text-muted-500">
                        Format
                      </dt>
                      <dd className="text-sm text-muted-900 mt-1">
                        <Badge variant="outline">{slideMetadata.format}</Badge>
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-500">
                        Vendor
                      </dt>
                      <dd className="text-sm text-muted-900 mt-1">
                        {slideMetadata.vendor || "Unknown"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-500">
                        Magnification
                      </dt>
                      <dd className="text-sm text-muted-900 mt-1">
                        {slideMetadata.magnification || "Unknown"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-500">
                        Pyramid Levels
                      </dt>
                      <dd className="text-sm text-muted-900 mt-1">
                        {slideMetadata.minLevel} - {slideMetadata.maxLevel}
                      </dd>
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Annotation Statistics
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <PhotoIcon className="h-5 w-5 text-purple-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-purple-900">
                        {rasterCount}
                      </p>
                      <p className="text-sm text-purple-700">Raster Masks</p>
                    </div>
                  </div>
                </div>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <CpuChipIcon className="h-5 w-5 text-blue-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-blue-900">
                        {vectorCount}
                      </p>
                      <p className="text-sm text-blue-700">
                        Vector Annotations
                      </p>
                    </div>
                  </div>
                </div>
                <div className="bg-green-50 border border-green-200 rounded-lg p-4 col-span-2">
                  <div className="flex items-center justify-center">
                    <div className="text-center">
                      <p className="text-3xl font-bold text-green-900">
                        {totalAnnotations}
                      </p>
                      <p className="text-sm text-green-700">
                        Total Annotations
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Add Mask Form */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Add New Mask
          </h3>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <p className="text-sm text-muted-600 mb-4">
              Add a new raster mask annotation to this slide. Provide the mask
              file URI and a descriptive name.
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
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <form.Field
                  name="maskName"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value?.trim()) return "Mask name is required";
                      return undefined;
                    },
                  }}
                >
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        Mask Name <span className="text-red-500">*</span>
                      </label>
                      <input
                        id={field.name}
                        name={field.name}
                        type="text"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        placeholder="Tumor segmentation"
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500 ${
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

                <form.Field
                  name="maskUri"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value?.trim()) return "Mask URI is required";
                      return undefined;
                    },
                  }}
                >
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        Mask URI <span className="text-red-500">*</span>
                      </label>
                      <input
                        id={field.name}
                        name={field.name}
                        type="text"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        placeholder="/path/to/mask.tiff"
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500 ${
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
              </div>

              <div className="mt-4 flex justify-end">
                <Button
                  type="submit"
                  disabled={isSubmitting || !form.state.canSubmit}
                  className="inline-flex items-center"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  {isSubmitting ? "Adding..." : "Add Mask"}
                </Button>
              </div>
            </form>
          </div>
        </div>

        {/* Add Vector Annotation Form */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Add New Vector Annotation
          </h3>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <p className="text-sm text-muted-600 mb-4">
              Add a new vector annotation to this slide. Provide the GeoJSON
              file path and a descriptive name.
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
                vectorForm.handleSubmit();
              }}
            >
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <vectorForm.Field
                  name="vectorName"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value?.trim())
                        return "Vector annotation name is required";
                      return undefined;
                    },
                  }}
                >
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        Vector Name <span className="text-red-500">*</span>
                      </label>
                      <input
                        id={field.name}
                        name={field.name}
                        type="text"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        placeholder="Cell boundaries"
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
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
                </vectorForm.Field>

                <vectorForm.Field
                  name="fileUri"
                  validators={{
                    onChange: ({ value }) => {
                      if (!value?.trim()) return "File URI is required";
                      if (
                        !value.endsWith(".geojson") &&
                        !value.endsWith(".json")
                      ) {
                        return "File must be a GeoJSON file (.geojson or .json)";
                      }
                      return undefined;
                    },
                  }}
                >
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        File URI <span className="text-red-500">*</span>
                      </label>
                      <input
                        id={field.name}
                        name={field.name}
                        type="text"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        placeholder="/path/to/annotations.geojson"
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
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
                </vectorForm.Field>
              </div>

              <div className="mt-4 flex justify-end">
                <Button
                  type="submit"
                  disabled={isSubmitting || !vectorForm.state.canSubmit}
                  className="inline-flex items-center bg-blue-600 hover:bg-blue-700"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  {isSubmitting ? "Adding..." : "Add Vector Annotation"}
                </Button>
              </div>
            </form>
          </div>
        </div>

        {/* Masks List */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Raster Masks
          </h3>

          {masksLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="animate-pulse bg-gray-100 h-16 rounded-lg"
                ></div>
              ))}
            </div>
          ) : masksError ? (
            <ErrorStateAlert
              error={masksError}
              title="Failed to load masks"
              onRetry={refetchMasks}
              variant="inline"
            />
          ) : masks.length === 0 ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
              <PhotoIcon className="h-12 w-12 text-muted-400 mx-auto mb-4" />
              <h4 className="text-lg font-medium text-muted-900 mb-2">
                No masks found
              </h4>
              <p className="text-muted-500">
                Use the form above to add your first mask annotation.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {masks.map((mask) => (
                <div
                  key={mask.maskUid}
                  className="bg-background border border-gray-200 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <PhotoIcon className="h-5 w-5 text-purple-500" />
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-muted-900">
                            {mask.maskName}
                          </span>
                          <Badge
                            variant="outline"
                            className="font-mono text-xs"
                          >
                            {mask.maskUid.substring(0, 8)}...
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-500 mt-1">
                          {mask.maskWidth} × {mask.maskHeight} pixels
                          {mask.maskMpp && (
                            <span> • {mask.maskMpp.toFixed(2)} μm/px</span>
                          )}
                          <span> • Added {formatDate(mask.createdAt)}</span>
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        onClick={() => copyToClipboard(mask.maskUid)}
                        variant="outline"
                        size="sm"
                      >
                        Copy ID
                      </Button>
                      <Button
                        onClick={() => handleDeleteClick(mask)}
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

        {/* Vector Annotations List */}
        <div>
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Vector Annotations
          </h3>

          {vectorsLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="animate-pulse bg-gray-100 h-16 rounded-lg"
                ></div>
              ))}
            </div>
          ) : vectorsError ? (
            <ErrorStateAlert
              error={vectorsError}
              title="Failed to load vector annotations"
              onRetry={refetchVectors}
              variant="inline"
            />
          ) : vectors.length === 0 ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
              <CpuChipIcon className="h-12 w-12 text-muted-400 mx-auto mb-4" />
              <h4 className="text-lg font-medium text-muted-900 mb-2">
                No vector annotations found
              </h4>
              <p className="text-muted-500">
                Use the form above to add your first vector annotation.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {vectors.map((vector) => (
                <div
                  key={vector.vectorUid}
                  className="bg-background border border-gray-200 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <CpuChipIcon className="h-5 w-5 text-blue-500" />
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-muted-900">
                            {vector.vectorName}
                          </span>
                          <Badge
                            variant="outline"
                            className="font-mono text-xs"
                          >
                            {vector.vectorUid.substring(0, 8)}...
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-500 mt-1">
                          GeoJSON file: {vector.fileUri}
                          <span> • Added {formatDate(vector.createdAt)}</span>
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        onClick={() => copyToClipboard(vector.vectorUid)}
                        variant="outline"
                        size="sm"
                      >
                        Copy ID
                      </Button>
                      <Button
                        onClick={() => handleVectorDeleteClick(vector)}
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

      {/* Delete Mask Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Mask</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove the mask "{maskToDelete?.maskName}
              "? This action cannot be undone and will remove all associated
              annotation data.
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
              Remove Mask
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Vector Annotation Confirmation Dialog */}
      <AlertDialog
        open={vectorDeleteDialogOpen}
        onOpenChange={setVectorDeleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Vector Annotation</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove the vector annotation "
              {vectorToDelete?.vectorName}"? This action cannot be undone and
              will remove all associated annotation data.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleVectorDeleteCancel}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleVectorDeleteConfirm}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
            >
              Remove Vector Annotation
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AdminPageLayout>
  );
}
