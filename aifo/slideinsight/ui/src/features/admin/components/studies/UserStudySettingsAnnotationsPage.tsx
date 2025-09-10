// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
// Copyright 2025 Jonas Teuwen. All rights reserved.
// This file is part of SlideInsight.
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { PlusIcon, TrashIcon } from "@heroicons/react/24/outline";
import { ColorResult } from "@uiw/react-color";
import ColorPicker from "@/components/ui/ColorPicker";

type AnnotationType = "point" | "box" | "polygon";

export interface AnnotationLabelEntry {
  label: string;
  type: AnnotationType;
  color: string;
}

export interface IndexMapEntry {
  index: number;
  label: string;
}

export interface ColorMapEntry {
  label: string;
  color: string;
}

export interface UserStudySettingsAnnotationsPageProps {
  // availability + labels
  allowAnnotation: boolean;
  setAllowAnnotation: (v: boolean) => void;
  annotationLabels: AnnotationLabelEntry[];
  addAnnotationLabel: () => void;
  updateAnnotationLabel: (
    index: number,
    field: keyof AnnotationLabelEntry,
    value: string
  ) => void;
  removeAnnotationLabel: (index: number) => void;
  handleAnnotationColorChange: (index: number, result: ColorResult) => void;

  // color popover state (shared)
  openColorPicker: string | null;
  setOpenColorPicker: (id: string | null) => void;

  // index map
  indexMap: IndexMapEntry[];
  addIndexMapEntry: () => void;
  updateIndexMapEntry: (
    entryIndex: number,
    field: keyof IndexMapEntry,
    value: string | number
  ) => void;
  removeIndexMapEntry: (entryIndex: number) => void;

  // color map
  colorMap: ColorMapEntry[];
  addColorMapEntry: () => void;
  updateColorMapEntry: (
    row: number,
    field: keyof ColorMapEntry,
    value: string
  ) => void;
  removeColorMapEntry: (row: number) => void;
  handleColorChange: (row: number, result: ColorResult) => void;
}

export default function UserStudySettingsAnnotationsPage({
  // availability + labels
  allowAnnotation,
  setAllowAnnotation,
  annotationLabels,
  addAnnotationLabel,
  updateAnnotationLabel,
  removeAnnotationLabel,
  handleAnnotationColorChange,

  // popover state
  openColorPicker,
  setOpenColorPicker,

  // index map
  indexMap,
  addIndexMapEntry,
  updateIndexMapEntry,
  removeIndexMapEntry,

  // color map
  colorMap,
  addColorMapEntry,
  updateColorMapEntry,
  removeColorMapEntry,
  handleColorChange,
}: UserStudySettingsAnnotationsPageProps) {
  const uid = React.useId();

  // Debug logging
  React.useEffect(() => {
    console.log(
      "UserStudySettingsAnnotationsPage openColorPicker state:",
      openColorPicker
    );
  }, [openColorPicker]);

  return (
    <Card className="border bg-card shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">Annotations</CardTitle>
        <p className="text-sm text-muted-foreground">
          Configure availability, label set, index mapping and colors in one
          place.
        </p>
      </CardHeader>

      <CardContent className="p-0">
        <div className="divide-y">
          {/* Availability */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label
                htmlFor={`${uid}-allow-annotation`}
                className="text-sm font-medium"
              >
                Annotation availability
              </Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Enable manual annotations in the viewer.
              </p>
            </div>
            <div className="sm:col-span-8">
              <div className="flex items-center gap-3">
                <Switch
                  id={`${uid}-allow-annotation`}
                  checked={allowAnnotation}
                  onCheckedChange={(v) => setAllowAnnotation(Boolean(v))}
                />
                <span className="text-sm text-muted-foreground">
                  {allowAnnotation ? "Enabled" : "Disabled"}
                </span>
              </div>
            </div>
          </section>

          {/* Label set */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Label set</Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Add concise labels, choose geometry, pick a color.
              </p>
            </div>

            <div className="sm:col-span-8 space-y-3">
              {!allowAnnotation ? (
                <div className="rounded-lg border bg-muted/20 p-4 text-sm text-muted-foreground">
                  Annotations are disabled. Enable them above to manage labels.
                </div>
              ) : (
                <>
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-medium">Annotation labels</div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={addAnnotationLabel}
                    >
                      <PlusIcon className="mr-1 h-4 w-4" />
                      Add label
                    </Button>
                  </div>

                  <div className="overflow-hidden rounded-lg border">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[45%]">Label</TableHead>
                          <TableHead className="w-[25%]">Type</TableHead>
                          <TableHead className="w-[25%]">Color</TableHead>
                          <TableHead className="w-[5%]" />
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {annotationLabels.length === 0 ? (
                          <TableRow>
                            <TableCell
                              colSpan={4}
                              className="py-10 text-center text-sm text-muted-foreground"
                            >
                              No labels yet. Click{" "}
                              <span className="font-medium">Add label</span>.
                            </TableCell>
                          </TableRow>
                        ) : (
                          annotationLabels.map((entry, i) => {
                            const rowId = `${uid}-ann-${i}`;
                            const isOpen = openColorPicker === rowId;
                            console.log(
                              "Annotation row",
                              i,
                              "rowId:",
                              rowId,
                              "isOpen:",
                              isOpen,
                              "openColorPicker:",
                              openColorPicker
                            );
                            return (
                              <TableRow key={rowId} className="align-middle">
                                <TableCell className="py-3">
                                  <Label
                                    htmlFor={`${rowId}-label`}
                                    className="sr-only"
                                  >
                                    Label
                                  </Label>
                                  <Input
                                    id={`${rowId}-label`}
                                    value={entry.label}
                                    onChange={(e) =>
                                      updateAnnotationLabel(
                                        i,
                                        "label",
                                        e.target.value
                                      )
                                    }
                                    placeholder="e.g., tumor, stroma, artifact"
                                  />
                                </TableCell>
                                <TableCell className="py-3">
                                  <Label
                                    htmlFor={`${rowId}-type`}
                                    className="sr-only"
                                  >
                                    Type
                                  </Label>
                                  <Select
                                    value={entry.type}
                                    onValueChange={(v) =>
                                      updateAnnotationLabel(i, "type", v)
                                    }
                                  >
                                    <SelectTrigger
                                      id={`${rowId}-type`}
                                      className="w-full"
                                    >
                                      <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                      <SelectItem value="point">
                                        Point
                                      </SelectItem>
                                      <SelectItem value="box">Box</SelectItem>
                                      <SelectItem value="polygon">
                                        Polygon
                                      </SelectItem>
                                    </SelectContent>
                                  </Select>
                                </TableCell>
                                <TableCell className="py-3">
                                  <ColorPicker
                                    color={entry.color}
                                    onChange={(res) =>
                                      handleAnnotationColorChange(i, res)
                                    }
                                    onChangeComplete={(res) =>
                                      handleAnnotationColorChange(i, res)
                                    }
                                    open={isOpen}
                                    onOpenChange={(o) => {
                                      console.log(
                                        "Annotation ColorPicker onOpenChange:",
                                        o,
                                        "rowId:",
                                        rowId
                                      );
                                      setOpenColorPicker(o ? rowId : null);
                                    }}
                                    side="right"
                                    align="start"
                                  />
                                </TableCell>
                                <TableCell className="py-3 text-right">
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => removeAnnotationLabel(i)}
                                    className="text-destructive hover:bg-destructive/10 hover:text-destructive/80"
                                    aria-label="Remove label"
                                  >
                                    <TrashIcon className="h-4 w-4" />
                                  </Button>
                                </TableCell>
                              </TableRow>
                            );
                          })
                        )}
                      </TableBody>
                    </Table>
                  </div>
                </>
              )}
            </div>
          </section>

          {/* Index mapping */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Index mapping</Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Map numeric indices to labels (for model outputs, etc.).
              </p>
            </div>
            <div className="sm:col-span-8 space-y-3">
              <div className="flex items-center justify-between">
                <div className="text-sm font-medium">Mappings</div>
                <Button variant="outline" size="sm" onClick={addIndexMapEntry}>
                  <PlusIcon className="mr-1 h-4 w-4" /> Add entry
                </Button>
              </div>

              <div className="overflow-hidden rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[30%]">Index</TableHead>
                      <TableHead className="w-[65%]">Label</TableHead>
                      <TableHead className="w-[5%]" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {indexMap.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={3}
                          className="py-8 text-center text-sm text-muted-foreground"
                        >
                          No index mappings yet.
                        </TableCell>
                      </TableRow>
                    ) : (
                      indexMap.map((m, _row) => (
                        <TableRow key={`${uid}-imap-${_row}`}>
                          <TableCell className="py-3">
                            <Input
                              type="number"
                              inputMode="numeric"
                              min={0}
                              value={m.index}
                              onChange={(e) =>
                                updateIndexMapEntry(
                                  m.index,
                                  "index",
                                  Number(e.target.value) || 0
                                )
                              }
                            />
                          </TableCell>
                          <TableCell className="py-3">
                            <Input
                              value={m.label}
                              onChange={(e) =>
                                updateIndexMapEntry(
                                  m.index,
                                  "label",
                                  e.target.value
                                )
                              }
                              placeholder="e.g., tissue, background"
                            />
                          </TableCell>
                          <TableCell className="py-3 text-right">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => removeIndexMapEntry(m.index)}
                              className="text-destructive hover:bg-destructive/10 hover:text-destructive/80"
                              aria-label="Remove mapping"
                            >
                              <TrashIcon className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
            </div>
          </section>

          {/* Color mapping */}
          <section className="grid gap-4 px-6 py-5 sm:grid-cols-12">
            <div className="sm:col-span-4">
              <Label className="text-sm font-medium">Color mapping</Label>
              <p className="mt-1 text-xs text-muted-foreground">
                Define colors per label used in visualisations.
              </p>
            </div>
            <div className="sm:col-span-8 space-y-3">
              <div className="flex items-center justify-between">
                <div className="text-sm font-medium">Colors</div>
                <Button variant="outline" size="sm" onClick={addColorMapEntry}>
                  <PlusIcon className="mr-1 h-4 w-4" /> Add color
                </Button>
              </div>

              <div className="overflow-hidden rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[55%]">Label</TableHead>
                      <TableHead className="w-[40%]">Color</TableHead>
                      <TableHead className="w-[5%]" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {colorMap.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={3}
                          className="py-8 text-center text-sm text-muted-foreground"
                        >
                          No color mappings yet.
                        </TableCell>
                      </TableRow>
                    ) : (
                      colorMap.map((c, i) => {
                        const rowId = `${uid}-cmap-${i}`;
                        const isOpen = openColorPicker === rowId;
                        console.log(
                          "Color mapping row",
                          i,
                          "rowId:",
                          rowId,
                          "isOpen:",
                          isOpen,
                          "openColorPicker:",
                          openColorPicker
                        );
                        return (
                          <TableRow key={rowId}>
                            <TableCell className="py-3">
                              <Input
                                value={c.label}
                                onChange={(e) =>
                                  updateColorMapEntry(
                                    i,
                                    "label",
                                    e.target.value
                                  )
                                }
                                placeholder="e.g., stroma"
                              />
                            </TableCell>
                            <TableCell className="py-3">
                              <ColorPicker
                                color={c.color}
                                onChange={(res) => handleColorChange(i, res)}
                                onChangeComplete={(res) =>
                                  handleColorChange(i, res)
                                }
                                open={isOpen}
                                onOpenChange={(o) => {
                                  console.log(
                                    "Color mapping ColorPicker onOpenChange:",
                                    o,
                                    "rowId:",
                                    rowId
                                  );
                                  setOpenColorPicker(o ? rowId : null);
                                }}
                                side="right"
                                align="start"
                              />
                            </TableCell>
                            <TableCell className="py-3 text-right">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => removeColorMapEntry(i)}
                                className="text-destructive hover:bg-destructive/10 hover:text-destructive/80"
                                aria-label="Remove color"
                              >
                                <TrashIcon className="h-4 w-4" />
                              </Button>
                            </TableCell>
                          </TableRow>
                        );
                      })
                    )}
                  </TableBody>
                </Table>
              </div>
              <p className="text-xs text-muted-foreground">
                Colors here also apply to labels used in the viewer overlays.
              </p>
            </div>
          </section>
        </div>
      </CardContent>
    </Card>
  );
}
