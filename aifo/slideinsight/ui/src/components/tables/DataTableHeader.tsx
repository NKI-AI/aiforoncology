// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import * as React from "react";
import { Button } from "../ui/button";
import { PlusIcon } from "../icons";

interface DataTableHeaderProps {
  title?: string;
  description?: string;
  addButtonText?: string;
  onAdd?: () => void;
  loading?: boolean;
}

export function DataTableHeader({
  title,
  description,
  addButtonText,
  onAdd,
  loading = false,
}: DataTableHeaderProps) {
  if (!title && !description && !onAdd) {
    return null;
  }

  return (
    <div className="flex items-center justify-between">
      {(title || description) && (
        <div>
          {title && <h2 className="text-lg font-medium">{title}</h2>}
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
      )}
      {onAdd && (
        <Button onClick={onAdd} size="sm" disabled={loading}>
          <PlusIcon className="h-4 w-4 mr-2" />
          {addButtonText || "Add"}
        </Button>
      )}
    </div>
  );
}
