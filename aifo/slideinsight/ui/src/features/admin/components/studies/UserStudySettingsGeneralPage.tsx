// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { StudyStatusCell } from "@/components/ui/StudyStatusCell";
import { CategoryPicker } from "@/components/ui/CategoryPicker";

interface GroupItem {
  name: string;
  displayName?: string;
}

export interface UserStudySettingsGeneralPageProps {
  studyUid: string;
  studyData?: { studyUid: string; name: string; isPublished: boolean } | null;
  studyName: string;
  setStudyName: (v: string) => void;
  studyDescription: string;
  setStudyDescription: (v: string) => void;
  publishedStatus: boolean | null;
  setPublishedStatus: (v: boolean) => void;
  // Categories
  categories: string[];
  setCategories: React.Dispatch<React.SetStateAction<string[]>>;
  categoryQuery: string;
  setCategoryQuery: (v: string) => void;
  categoryPickerOpen: boolean;
  setCategoryPickerOpen: (v: boolean) => void;
  CATEGORY_SUGGESTIONS: string[];
  // Diseases
  diseases: string[];
  setDiseases: React.Dispatch<React.SetStateAction<string[]>>;
  diseaseQuery: string;
  setDiseaseQuery: (v: string) => void;
  diseasePickerOpen: boolean;
  setDiseasePickerOpen: (v: boolean) => void;
  DISEASE_SUGGESTIONS: string[];
  // Stainings
  stainings: string[];
  setStainings: React.Dispatch<React.SetStateAction<string[]>>;
  stainingQuery: string;
  setStainingQuery: (v: string) => void;
  stainingPickerOpen: boolean;
  setStainingPickerOpen: (v: boolean) => void;
  STAINING_SUGGESTIONS: string[];
  // Groups
  groupName: string;
  setGroupName: (v: string) => void;
  groups: GroupItem[];
  // UI helpers
  pillClassFor: (label: string) => string;
}

export default function UserStudySettingsGeneralPage({
  studyUid,
  studyData,
  studyName,
  setStudyName,
  studyDescription,
  setStudyDescription,
  publishedStatus,
  setPublishedStatus,
  categories,
  setCategories,
  categoryQuery,
  setCategoryQuery,
  categoryPickerOpen,
  setCategoryPickerOpen,
  CATEGORY_SUGGESTIONS,
  diseases,
  setDiseases,
  diseaseQuery,
  setDiseaseQuery,
  diseasePickerOpen,
  setDiseasePickerOpen,
  DISEASE_SUGGESTIONS,
  stainings,
  setStainings,
  stainingQuery,
  setStainingQuery,
  stainingPickerOpen,
  setStainingPickerOpen,
  STAINING_SUGGESTIONS,
  groupName,
  setGroupName,
  groups,
  pillClassFor,
}: UserStudySettingsGeneralPageProps) {
  return (
    <Card className="border bg-card shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Study profile</CardTitle>
        <p className="text-sm text-muted-foreground">
          Define the core properties and taxonomy for this study.
        </p>
      </CardHeader>

      {/* Professional form layout: label column (left), field column (right), divided rows */}
      <CardContent className="p-0">
        <div className="divide-y">
          {/* Name */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label htmlFor="study-name" className="text-sm font-medium">
                Name
              </Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Concise and recognizable.
              </p>
            </div>
            <div className="sm:col-span-8">
              <Input
                id="study-name"
                value={studyName}
                onChange={(e) => setStudyName(e.target.value)}
                placeholder="e.g., TCGA Breast Cohort"
              />
              <p className="mt-1.5 text-xs text-muted-foreground">
                Appears in listings and share dialogs.
              </p>
            </div>
          </section>

          {/* Description */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label
                htmlFor="study-description"
                className="text-sm font-medium"
              >
                Description
              </Label>
              <p className="mt-1 text-xs text-muted-foreground">
                What does this study cover?
              </p>
            </div>
            <div className="sm:col-span-8">
              <Textarea
                id="study-description"
                value={studyDescription}
                onChange={(e) => setStudyDescription(e.target.value)}
                className="min-h-24"
                placeholder="Briefly describe the scope, cohort, and primary objectives."
              />
            </div>
          </section>

          {/* Publication */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Publication</Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Control visibility to collaborators and groups.
              </p>
            </div>
            <div className="flex items-center justify-between sm:col-span-8">
              <div className="rounded-md border bg-muted/30 px-3 py-2 text-xs">
                <span className="text-muted-foreground">Status:&nbsp;</span>
                <span className="font-medium">
                  {publishedStatus ? "Published" : "Draft"}
                </span>
              </div>
              {studyData && (
                <StudyStatusCell
                  study={{
                    studyUid: studyData.studyUid,
                    name: studyData.name,
                    isPublished: (publishedStatus ??
                      studyData.isPublished) as boolean,
                  }}
                  onStatusUpdate={(_, v) => {
                    if (typeof v === "boolean") setPublishedStatus(v);
                  }}
                />
              )}
            </div>
          </section>

          {/* Group */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label htmlFor="study-group" className="text-sm font-medium">
                Group
              </Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Used for access and filtering.
              </p>
            </div>
            <div className="sm:col-span-8">
              <Select value={groupName} onValueChange={setGroupName}>
                <SelectTrigger id="study-group">
                  <SelectValue placeholder="Select a group" />
                </SelectTrigger>
                <SelectContent>
                  {(groups.length === 0
                    ? [{ name: "Example Group", displayName: "Example Group" }]
                    : groups
                  ).map((g) => (
                    <SelectItem key={g.name} value={g.name}>
                      {g.displayName || g.name}
                    </SelectItem>
                  ))}
                  {groups.length === 0 && (
                    <SelectItem value="__create__" disabled>
                      Create new group (coming soon)
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>
          </section>

          {/* Organs */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Organs</Label>
              <p className="mt-1 text-xs text-muted-foreground">Choose 1–3.</p>
            </div>
            <div className="sm:col-span-8">
              <CategoryPicker
                addLabel="Add Category"
                selected={categories}
                setSelected={setCategories}
                suggestions={CATEGORY_SUGGESTIONS}
                query={categoryQuery}
                setQuery={setCategoryQuery}
                isOpen={categoryPickerOpen}
                setIsOpen={setCategoryPickerOpen}
                pillClassFor={pillClassFor}
              />
            </div>
          </section>

          {/* Disease */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Disease</Label>
            </div>
            <div className="sm:col-span-8">
              <CategoryPicker
                addLabel="Add Disease"
                selected={diseases}
                setSelected={setDiseases}
                suggestions={DISEASE_SUGGESTIONS}
                query={diseaseQuery}
                setQuery={setDiseaseQuery}
                isOpen={diseasePickerOpen}
                setIsOpen={setDiseasePickerOpen}
                pillClassFor={pillClassFor}
              />
            </div>
          </section>

          {/* Staining */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Staining</Label>
            </div>
            <div className="sm:col-span-8">
              <CategoryPicker
                addLabel="Add Staining"
                selected={stainings}
                setSelected={setStainings}
                suggestions={STAINING_SUGGESTIONS}
                query={stainingQuery}
                setQuery={setStainingQuery}
                isOpen={stainingPickerOpen}
                setIsOpen={setStainingPickerOpen}
                pillClassFor={pillClassFor}
              />
            </div>
          </section>
        </div>
      </CardContent>
    </Card>
  );
}
