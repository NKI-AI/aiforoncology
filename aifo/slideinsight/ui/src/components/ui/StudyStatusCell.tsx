// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import StatusToggleCell from "./StatusToggleCell";
import { StatusType } from "./StatusDropdown";

interface Study {
  studyUid: string;
  name: string;
  isPublished: boolean;
}

interface StudyStatusCellProps {
  study: Study;
  onStatusUpdate?: (studyUid: string, newStatus: boolean) => void;
}

export const StudyStatusCell: React.FC<StudyStatusCellProps> = ({
  study,
  onStatusUpdate,
}) => {
  // Wrapper to ensure type safety
  const handleUpdate = (entityId: string, newValue: boolean | StatusType) => {
    if (onStatusUpdate && typeof newValue === "boolean") {
      onStatusUpdate(entityId, newValue);
    }
  };

  return (
    <StatusToggleCell
      entity={study}
      config={{
        type: "boolean",
        apiEndpoint: (studyUid) => `/api/v1/studies/${studyUid}`,
        apiField: "isPublished",
        trueLabel: "Published",
        falseLabel: "Draft",
        trueColor: "text-green-800 bg-green-100 border-green-200",
        falseColor: "text-yellow-800 bg-yellow-100 border-yellow-200",
        successMessage: (study, newStatus) =>
          `Study ${newStatus ? "published" : "unpublished"}!`,
        errorMessage: "Failed to update study status",
        size: "sm",
        variant: "badge",
      }}
      getEntityId={(study) => study.studyUid}
      getEntityName={(study) => study.name}
      getCurrentValue={(study) => study.isPublished}
      onUpdate={handleUpdate}
    />
  );
};
