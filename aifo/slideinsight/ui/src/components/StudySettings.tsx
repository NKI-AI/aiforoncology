// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate } from "@tanstack/react-router";
import { ColorResult } from "@uiw/react-color";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { StudyStatusCell } from "@/components/ui/StudyStatusCell";
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
import {
  ArrowLeftIcon,
  PlusIcon,
  TrashIcon,
  PaintBrushIcon,
} from "@heroicons/react/24/outline";
import { toast } from "sonner";
import {
  useStudyMetadataField,
  useUpdateStudyMetadataField,
} from "@/api/hooks";
import { useStudy, useUpdateStudy, useGroups } from "@/api/hooks";
// Permissions logic moved to dedicated component
// import { useUsers, usePermissions as useAllPermissions } from "@/api/hooks";
// import { useObjectGrants, useCreateObjectGrant, useDeleteObjectGrant } from "@/features/admin/hooks/useObjectGrants";
import TabbedPage, { TabbedPagePage } from "@/components/TabbedPage";
import UserStudySettingsPermissionPage from "@/features/admin/components/studies/UserStudySettingsPermissionPage";
import UserStudySettingsGeneralPage from "@/features/admin/components/studies/UserStudySettingsGeneralPage";
import UserStudySettingsAnnotationsPage from "@/features/admin/components/studies/UserStudySettingsAnnotationsPage";
import UserStudySettingsModulesPage from "@/features/admin/components/studies/UserStudySettingsModulesPage";

interface IndexMapEntry {
  index: number;
  label: string;
}

interface ColorMapEntry {
  label: string;
  color: string;
}

type AnnotationType = "point" | "box" | "polygon";
interface AnnotationLabelEntry {
  label: string;
  type: AnnotationType;
  color: string;
}

interface PluginSettings {
  viewerSettings: boolean;
  brightnessControl: boolean;
  annotationControl: boolean;
  maskControl: boolean;
  regionControl: boolean;
}

interface StudyMetadata {
  index_map?: Record<number, string>;
  color_map?: Record<string, string>;
  allow_annotation?: boolean;
  annotations?: AnnotationLabelEntry[];
  plugin_settings?: {
    viewer_settings?: boolean;
    brightness_control?: boolean;
    annotation_control?: boolean;
    mask_control?: boolean;
    region_control?: boolean;
  };
}

interface StudySettingsProps {
  initialTab?: string;
}

const StudySettings: React.FC<StudySettingsProps> = ({ initialTab }) => {
  const { studyUid } = useParams({
    from: "/_authenticated/studies/$studyUid/settings",
  });
  const navigate = useNavigate();

  // Fetch current metadata
  const {
    data: metadataResponse,
    isLoading: loadingMetadata,
    error: metadataError,
  } = useStudyMetadataField(studyUid);

  // Update mutation
  const updateMetadata = useUpdateStudyMetadataField(studyUid, {
    onSuccess: () => {
      toast.success("Study settings saved successfully!");
    },
    onError: (error) => {
      toast.error("Failed to save study settings", {
        description: error.message || "Please try again.",
      });
    },
  });

  // Local state for editing
  const [indexMap, setIndexMap] = useState<IndexMapEntry[]>([]);
  const [colorMap, setColorMap] = useState<ColorMapEntry[]>([]);
  const [allowAnnotation, setAllowAnnotation] = useState(false);
  const [annotationLabels, setAnnotationLabels] = useState<
    AnnotationLabelEntry[]
  >([]);
  const [pluginSettings, setPluginSettings] = useState<PluginSettings>({
    viewerSettings: true,
    brightnessControl: true,
    annotationControl: false,
    maskControl: true,
    regionControl: false,
  });
  const [isDirty, setIsDirty] = useState(false);
  const [showUnsavedDialog, setShowUnsavedDialog] = useState(false);
  const [pendingNavigation, setPendingNavigation] = useState<string | null>(
    null
  );
  const [openColorPicker, setOpenColorPicker] = useState<string | null>(null);

  // Debug logging for openColorPicker state changes
  React.useEffect(() => {
    console.log("StudySettings openColorPicker changed to:", openColorPicker);
  }, [openColorPicker]);
  const [activeTab, setActiveTab] = useState<string>(initialTab || "general");

  // Permissions tab moved to UserStudySettingsPermissionPage

  // Study tab state
  const { data: studyData } = useStudy(studyUid);
  const updateStudy = useUpdateStudy(studyUid);
  const { data: groups = [] } = useGroups({ page: 1, limit: 1000 });
  const [studyName, setStudyName] = useState("");
  const [studyDescription, setStudyDescription] = useState("");
  const [publishedStatus, setPublishedStatus] = useState<boolean | null>(null);
  const [categories, setCategories] = useState<string[]>([]);
  const [diseases, setDiseases] = useState<string[]>([]);
  const [diseaseQuery, setDiseaseQuery] = useState("");
  const [diseasePickerOpen, setDiseasePickerOpen] = useState(false);
  const DISEASE_SUGGESTIONS: string[] = [
    "Melanoma",
    "DCIS",
    "Invasive Ductal Carcinoma",
    "Basal Cell Carcinoma",
    "Squamous Cell Carcinoma",
    "Glioblastoma",
    "Adenocarcinoma",
    "Lymphoma",
    "Leukemia",
    "Sarcoma",
    "Metastasis",
  ];
  const [stainings, setStainings] = useState<string[]>([]);
  const [groupName, setGroupName] = useState<string>("");
  const [categoryQuery, setCategoryQuery] = useState("");
  const [stainingQuery, setStainingQuery] = useState("");
  const [categoryPickerOpen, setCategoryPickerOpen] = useState(false);
  const [stainingPickerOpen, setStainingPickerOpen] = useState(false);

  const CATEGORY_SUGGESTIONS: string[] = [
    "Skin",
    "Breast",
    "Lung",
    "Colon",
    "Prostate",
    "Brain",
    "Liver",
    "Kidney",
    "Pancreas",
    "Ovary",
    "Uterus",
    "Bladder",
    "Head & Neck",
    "Esophagus",
    "Stomach",
    "Small Intestine",
    "Bone",
    "Soft Tissue",
    "Lymph Node",
    "Eye",
  ];

  const STAINING_SUGGESTIONS: string[] = [
    "H&E",
    "PD-L1",
    "CD3",
    "CD8",
    "Ki-67",
    "HER2",
    "ER",
    "PR",
    "PAX8",
    "SOX10",
    "S100",
    "PanCK",
    "CK7",
    "CK20",
    "CD68",
    "CD163",
    "CD4",
    "CD20",
    "DAPI",
  ];

  // Filtering now handled inside Picker component

  const pillClassFor = useCallback((label: string) => {
    const palette = [
      ["bg-emerald-500/15", "text-emerald-200", "border-emerald-500/30"],
      ["bg-sky-500/15", "text-sky-200", "border-sky-500/30"],
      ["bg-fuchsia-500/15", "text-fuchsia-200", "border-fuchsia-500/30"],
      ["bg-amber-500/15", "text-amber-200", "border-amber-500/30"],
      ["bg-violet-500/15", "text-violet-200", "border-violet-500/30"],
      ["bg-rose-500/15", "text-rose-200", "border-rose-500/30"],
      ["bg-cyan-500/15", "text-cyan-200", "border-cyan-500/30"],
      ["bg-lime-500/15", "text-lime-200", "border-lime-500/30"],
      ["bg-indigo-500/15", "text-indigo-200", "border-indigo-500/30"],
      ["bg-teal-500/15", "text-teal-200", "border-teal-500/30"],
    ];
    let hash = 0;
    for (let i = 0; i < label.length; i++)
      hash = (hash * 31 + label.charCodeAt(i)) >>> 0;
    const [bg, fg, bd] = palette[hash % palette.length];
    return `${bg} ${fg} ${bd}`;
  }, []);

  // Initialize local state from API data
  useEffect(() => {
    if (metadataResponse?.metadata) {
      const metadata = metadataResponse.metadata as StudyMetadata;

      // Convert index_map to array format
      if (metadata.index_map) {
        const indexEntries = Object.entries(metadata.index_map).map(
          ([index, label]) => ({
            index: parseInt(index),
            label,
          })
        );
        setIndexMap(indexEntries);
      } else {
        setIndexMap([]);
      }

      // Convert color_map to array format
      if (metadata.color_map) {
        const colorEntries = Object.entries(metadata.color_map).map(
          ([label, color]) => ({
            label,
            color,
          })
        );
        setColorMap(colorEntries);
      } else {
        setColorMap([]);
      }

      // Annotation toggle and labels
      setAllowAnnotation(Boolean(metadata.allow_annotation));
      if (metadata.annotations && Array.isArray(metadata.annotations)) {
        setAnnotationLabels(
          metadata.annotations.map((a) => ({
            label: a.label,
            type: a.type as AnnotationType,
            color: a.color || "#FF0000",
          }))
        );
      } else {
        setAnnotationLabels([]);
      }

      // Plugin settings
      if (metadata.plugin_settings) {
        setPluginSettings({
          viewerSettings: metadata.plugin_settings.viewer_settings ?? true,
          brightnessControl:
            metadata.plugin_settings.brightness_control ?? true,
          annotationControl:
            metadata.plugin_settings.annotation_control ??
            Boolean(metadata.allow_annotation),
          maskControl: metadata.plugin_settings.mask_control ?? true,
          regionControl: metadata.plugin_settings.region_control ?? false,
        });
      } else {
        // Default plugin settings if not in metadata
        setPluginSettings({
          viewerSettings: true,
          brightnessControl: true,
          annotationControl: Boolean(metadata.allow_annotation),
          maskControl: true,
          regionControl: false,
        });
      }

      // Load study_info data (organs, diseases, stainings)
      if ((metadata as any).study_info) {
        const studyInfo = (metadata as any).study_info;

        // Handle both 'organs' and 'categories' for backward compatibility
        const organsData = studyInfo.organs || studyInfo.categories || [];
        setCategories(Array.isArray(organsData) ? organsData : []);

        setDiseases(
          Array.isArray(studyInfo.diseases) ? studyInfo.diseases : []
        );
        setStainings(
          Array.isArray(studyInfo.stainings) ? studyInfo.stainings : []
        );
        setGroupName(studyInfo.group || "");
      } else {
        // Reset to defaults if no study_info
        setCategories([]);
        setDiseases([]);
        setStainings([]);
        setGroupName("");
      }

      setIsDirty(false);
    }
  }, [metadataResponse]);

  // Initialize study info from study entity and metadata
  useEffect(() => {
    if (studyData) {
      setStudyName(studyData.name || "");
      setStudyDescription(studyData.description || "");
      setPublishedStatus(
        typeof studyData.isPublished === "boolean"
          ? studyData.isPublished
          : null
      );
    }
  }, [studyData]);

  // Sync initial tab from props and keep URL in sync
  useEffect(() => {
    if (initialTab && initialTab !== activeTab) {
      setActiveTab(initialTab);
    }
  }, [initialTab]);

  // Mark as dirty when data changes
  useEffect(() => {
    if (metadataResponse?.metadata) {
      setIsDirty(true);
    }
  }, [
    indexMap,
    colorMap,
    allowAnnotation,
    annotationLabels,
    pluginSettings,
    studyName,
    studyDescription,
    categories,
    diseases,
    stainings,
    groupName,
    publishedStatus,
  ]);

  const handleBack = useCallback(() => {
    if (isDirty) {
      setPendingNavigation(`/studies/${studyUid}`);
      setShowUnsavedDialog(true);
    } else {
      navigate({ to: `/studies/${studyUid}` });
    }
  }, [isDirty, navigate, studyUid]);

  const handleSave = useCallback(async () => {
    // Convert arrays back to object format
    const index_map: Record<number, string> = {};
    indexMap.forEach(({ index, label }) => {
      if (label.trim()) {
        index_map[index] = label.trim();
      }
    });

    const color_map: Record<string, string> = {};
    colorMap.forEach(({ label, color }) => {
      if (label.trim()) {
        color_map[label.trim()] = color;
      }
    });

    // Ensure annotation label colors are reflected in global color_map
    annotationLabels.forEach(({ label, color }) => {
      if (label.trim()) {
        color_map[label.trim()] = color;
      }
    });

    const metadata: StudyMetadata = {
      index_map,
      color_map,
      allow_annotation: allowAnnotation,
      annotations: annotationLabels
        .filter((a) => a.label.trim())
        .map((a) => ({
          label: a.label.trim(),
          type: a.type,
          color: a.color,
        })),
      plugin_settings: {
        viewer_settings: pluginSettings.viewerSettings,
        brightness_control: pluginSettings.brightnessControl,
        annotation_control: pluginSettings.annotationControl,
        mask_control: pluginSettings.maskControl,
        region_control: pluginSettings.regionControl,
      },
    };

    // Add non-API study settings into metadata for now
    (metadata as any).study_info = {
      organs: categories.sort((a, b) => a.localeCompare(b)), // Sort alphabetically and rename to organs
      diseases: diseases.sort((a, b) => a.localeCompare(b)), // Sort alphabetically
      stainings: stainings.sort((a, b) => a.localeCompare(b)), // Sort alphabetically
      group: groupName || undefined,
    };

    // Log the payload for debugging/visibility
    // (Preview card removed per requirements)
    // eslint-disable-next-line no-console
    console.log("Saving study annotation settings:", metadata);

    // Persist study name/description and metadata in parallel
    await Promise.all([
      (async () => {
        const updates: any = {};
        if (studyName !== studyData?.name) {
          updates.name = studyName;
        }
        if (studyDescription !== (studyData?.description || "")) {
          updates.description = studyDescription;
        }
        if (
          publishedStatus !== null &&
          publishedStatus !== studyData?.isPublished
        ) {
          updates.isPublished = publishedStatus;
        }
        if (Object.keys(updates).length > 0) {
          await updateStudy.mutateAsync(updates);
        }
      })(),
      updateMetadata.mutateAsync(metadata),
    ]);
    setIsDirty(false);
  }, [
    indexMap,
    colorMap,
    allowAnnotation,
    annotationLabels,
    categories,
    diseases,
    stainings,
    groupName,
    studyName,
    studyDescription,
    studyData,
    updateStudy,
    updateMetadata,
  ]);

  const handleUnsavedConfirm = useCallback(() => {
    if (pendingNavigation) {
      navigate({ to: pendingNavigation });
    }
    setShowUnsavedDialog(false);
    setPendingNavigation(null);
  }, [navigate, pendingNavigation]);

  const handleUnsavedCancel = useCallback(() => {
    setShowUnsavedDialog(false);
    setPendingNavigation(null);
  }, []);

  // Index map handlers
  const addIndexMapEntry = useCallback(() => {
    const nextIndex = Math.max(0, ...indexMap.map((e) => e.index)) + 1;
    setIndexMap((prev) => [...prev, { index: nextIndex, label: "" }]);
  }, [indexMap]);

  const updateIndexMapEntry = useCallback(
    (index: number, field: keyof IndexMapEntry, value: string | number) => {
      setIndexMap((prev) =>
        prev.map((entry) =>
          entry.index === index ? { ...entry, [field]: value } : entry
        )
      );
    },
    []
  );

  const removeIndexMapEntry = useCallback((index: number) => {
    setIndexMap((prev) => prev.filter((entry) => entry.index !== index));
  }, []);

  // Color map handlers
  const addColorMapEntry = useCallback(() => {
    setColorMap((prev) => [...prev, { label: "", color: "#FF0000" }]);
  }, []);

  const updateColorMapEntry = useCallback(
    (index: number, field: keyof ColorMapEntry, value: string) => {
      setColorMap((prev) =>
        prev.map((entry, i) =>
          i === index ? { ...entry, [field]: value } : entry
        )
      );
    },
    []
  );

  const removeColorMapEntry = useCallback((index: number) => {
    setColorMap((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const handleColorChange = useCallback(
    (index: number, result: ColorResult) => {
      updateColorMapEntry(index, "color", result.hex);
    },
    [updateColorMapEntry]
  );

  // Annotation labels handlers
  const addAnnotationLabel = useCallback(() => {
    setAnnotationLabels((prev) => [
      ...prev,
      { label: "", type: "point", color: "#22c55e" },
    ]);
  }, []);

  const updateAnnotationLabel = useCallback(
    (index: number, field: keyof AnnotationLabelEntry, value: string) => {
      setAnnotationLabels((prev) =>
        prev.map((entry, i) =>
          i === index ? { ...entry, [field]: value } : entry
        )
      );
    },
    []
  );

  const removeAnnotationLabel = useCallback((index: number) => {
    setAnnotationLabels((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const handleAnnotationColorChange = useCallback(
    (index: number, result: ColorResult) => {
      updateAnnotationLabel(index, "color", result.hex);
    },
    [updateAnnotationLabel]
  );

  if (loadingMetadata) {
    return (
      <div className="mx-auto max-w-5xl px-6 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 w-1/3 rounded bg-muted" />
          <div className="h-64 rounded bg-muted" />
        </div>
      </div>
    );
  }

  if (metadataError) {
    return (
      <div className="mx-auto max-w-5xl px-6 py-8">
        <div className="py-10 text-center">
          <p className="text-destructive">Failed to load study settings</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const handleTabChange = useCallback(
    (tabValue: string) => {
      setActiveTab(tabValue);
      navigate({ to: `/studies/${studyUid}/settings/${tabValue}` });
    },
    [navigate, studyUid]
  );

  return (
    <div className="mx-auto max-w-5xl px-6 py-8">
      {/* Header — neutral, no gradient */}
      <header className="mb-6 rounded-lg border bg-card shadow-sm">
        <div className="flex flex-col gap-4 p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold tracking-tight">
              {studyName || "Study"}
            </h2>
            <p className="text-sm text-muted-foreground">
              UID: <span className="font-mono">{studyUid}</span>
            </p>
          </div>
          {studyData && (
            <div className="flex items-center gap-3">
              <span className="text-sm text-muted-foreground">Visibility</span>
              <StudyStatusCell
                study={{
                  studyUid: studyData.studyUid,
                  name: studyData.name,
                  isPublished: publishedStatus ?? studyData.isPublished,
                }}
                onStatusUpdate={(_, v) => {
                  if (typeof v === "boolean") setPublishedStatus(v);
                }}
              />
            </div>
          )}
        </div>
      </header>

      <TabbedPage
        title="Study Settings"
        subtitle="Configure study options, annotations and permissions"
        activeValue={activeTab}
        onValueChange={handleTabChange}
        leftActions={
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeftIcon className="mr-2 h-4 w-4" />
            Back to Study
          </Button>
        }
        rightActions={
          <div className="flex items-center gap-3">
            {isDirty && (
              <span className="inline-flex items-center gap-2 rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs text-amber-700 dark:border-amber-900/40 dark:bg-amber-950/40 dark:text-amber-400">
                <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                Unsaved changes
              </span>
            )}
            <Button
              onClick={handleSave}
              disabled={updateMetadata.isPending || !isDirty}
            >
              {updateMetadata.isPending ? "Saving..." : "Save Settings"}
            </Button>
          </div>
        }
      >
        {/* GENERAL */}
        <TabbedPagePage value="general" label="Study">
          <UserStudySettingsGeneralPage
            studyUid={studyUid}
            studyData={studyData as any}
            studyName={studyName}
            setStudyName={setStudyName}
            studyDescription={studyDescription}
            setStudyDescription={setStudyDescription}
            publishedStatus={publishedStatus}
            setPublishedStatus={(v) => setPublishedStatus(v)}
            categories={categories}
            setCategories={setCategories}
            categoryQuery={categoryQuery}
            setCategoryQuery={setCategoryQuery}
            categoryPickerOpen={categoryPickerOpen}
            setCategoryPickerOpen={setCategoryPickerOpen}
            CATEGORY_SUGGESTIONS={CATEGORY_SUGGESTIONS}
            diseases={diseases}
            setDiseases={setDiseases}
            diseaseQuery={diseaseQuery}
            setDiseaseQuery={setDiseaseQuery}
            diseasePickerOpen={diseasePickerOpen}
            setDiseasePickerOpen={setDiseasePickerOpen}
            DISEASE_SUGGESTIONS={DISEASE_SUGGESTIONS}
            stainings={stainings}
            setStainings={setStainings}
            stainingQuery={stainingQuery}
            setStainingQuery={setStainingQuery}
            stainingPickerOpen={stainingPickerOpen}
            setStainingPickerOpen={setStainingPickerOpen}
            STAINING_SUGGESTIONS={STAINING_SUGGESTIONS}
            groupName={groupName}
            setGroupName={setGroupName}
            groups={groups as any}
            pillClassFor={pillClassFor}
          />
        </TabbedPagePage>

        {/* MODULES */}
        <TabbedPagePage value="modules" label="Modules">
          <UserStudySettingsModulesPage
            pluginSettings={pluginSettings}
            setPluginSettings={setPluginSettings}
          />
        </TabbedPagePage>

        {/* ANNOTATIONS */}
        <TabbedPagePage value="annotations" label="Annotations">
          <UserStudySettingsAnnotationsPage
            // availability + labels
            allowAnnotation={allowAnnotation}
            setAllowAnnotation={setAllowAnnotation}
            annotationLabels={annotationLabels}
            addAnnotationLabel={addAnnotationLabel}
            updateAnnotationLabel={updateAnnotationLabel}
            removeAnnotationLabel={removeAnnotationLabel}
            handleAnnotationColorChange={handleAnnotationColorChange}
            // shared color popover state
            openColorPicker={openColorPicker}
            setOpenColorPicker={setOpenColorPicker}
            // index mapping (moved into the same card)
            indexMap={indexMap}
            addIndexMapEntry={addIndexMapEntry}
            updateIndexMapEntry={updateIndexMapEntry}
            removeIndexMapEntry={removeIndexMapEntry}
            // color mapping (moved into the same card)
            colorMap={colorMap}
            addColorMapEntry={addColorMapEntry}
            updateColorMapEntry={updateColorMapEntry}
            removeColorMapEntry={removeColorMapEntry}
            handleColorChange={handleColorChange}
          />
        </TabbedPagePage>

        {/* PERMISSIONS */}
        <TabbedPagePage value="permissions" label="Permissions">
          <UserStudySettingsPermissionPage />
        </TabbedPagePage>
      </TabbedPage>

      {/* Unsaved Changes Dialog */}
      <AlertDialog open={showUnsavedDialog} onOpenChange={setShowUnsavedDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Unsaved Changes</AlertDialogTitle>
            <AlertDialogDescription>
              You have unsaved changes to your study settings. Are you sure you
              want to leave without saving?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleUnsavedCancel}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleUnsavedConfirm}>
              Leave Without Saving
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default StudySettings;
