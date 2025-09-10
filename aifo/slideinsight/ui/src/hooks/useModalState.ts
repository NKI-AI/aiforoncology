// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useCallback } from "react";
import type { ModalState } from "./types";

// Re-export the type for convenience
export type { ModalState };

export function useModalState<T>(refetch?: () => void): ModalState<T> {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedEntity, setSelectedEntity] = useState<T | null>(null);

  const handleAddEntity = useCallback(() => {
    setSelectedEntity(null);
    setIsCreateModalOpen(true);
  }, []);

  const handleEditEntity = useCallback((entity: T) => {
    setSelectedEntity(entity);
    setIsEditModalOpen(true);
  }, []);

  const handleFormSuccess = useCallback(() => {
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
    setSelectedEntity(null);
    refetch?.();
  }, [refetch]);

  const handleFormCancel = useCallback(() => {
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
    setSelectedEntity(null);
  }, []);

  return {
    // States
    isCreateModalOpen,
    isEditModalOpen,
    selectedEntity,

    // Handlers
    handleAddEntity,
    handleEditEntity,
    handleFormSuccess,
    handleFormCancel,

    // Setters for manual control
    setIsCreateModalOpen,
    setIsEditModalOpen,
    setSelectedEntity,
  };
}
