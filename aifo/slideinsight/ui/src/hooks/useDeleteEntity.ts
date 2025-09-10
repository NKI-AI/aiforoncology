// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useState, useCallback } from "react";
import { apiFetch } from "../utils/fetchUtils";
import { toast } from "sonner";

export interface DeleteEntityConfig<T> {
  entityNameSingular: string;
  deleteEndpoint: (entity: T) => string;
  getEntityDisplayName: (entity: T) => string;
}

export interface DeleteEntityState<T> {
  // Delete states
  isDeleteDialogOpen: boolean;
  entityToDelete: T | null;
  isDeleting: boolean;

  // Handlers
  handleDeleteEntity: (entity: T) => void;
  confirmDelete: () => Promise<void>;
  cancelDelete: () => void;

  // Setters for manual control
  setIsDeleteDialogOpen: (open: boolean) => void;
  setEntityToDelete: (entity: T | null) => void;
}

export function useDeleteEntity<T>(
  config: DeleteEntityConfig<T>,
  refetch?: () => void
): DeleteEntityState<T> {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [entityToDelete, setEntityToDelete] = useState<T | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDeleteEntity = useCallback((entity: T) => {
    setEntityToDelete(entity);
    setIsDeleteDialogOpen(true);
  }, []);

  const confirmDelete = useCallback(async () => {
    if (!entityToDelete) return;

    setIsDeleting(true);
    try {
      // Use shorter timeout for delete operations to prevent UI freezing
      await apiFetch(
        config.deleteEndpoint(entityToDelete),
        {
          method: "DELETE",
        },
        3000
      ); // 3 second timeout for delete operations

      toast.success(`${config.entityNameSingular} deleted!`, {
        description: `${config.getEntityDisplayName(
          entityToDelete
        )} has been permanently deleted.`,
      });

      setIsDeleteDialogOpen(false);
      setEntityToDelete(null);
      refetch?.();
    } catch (error) {
      console.error(
        `Failed to delete ${config.entityNameSingular.toLowerCase()}:`,
        error
      );

      let errorMessage = "An unexpected error occurred";
      if (error instanceof Error) {
        if (error.message.includes("timed out")) {
          errorMessage = "Request timed out. Please try again.";
        } else if (error.message.includes("Failed to fetch")) {
          errorMessage = "Network error. Please check your connection.";
        } else {
          errorMessage = error.message;
        }
      }

      toast.error(
        `Failed to delete ${config.entityNameSingular.toLowerCase()}`,
        {
          description: errorMessage,
        }
      );
    } finally {
      // Ensure we always reset the deleting state
      setIsDeleting(false);
    }
  }, [entityToDelete, config, refetch]);

  const cancelDelete = useCallback(() => {
    setIsDeleteDialogOpen(false);
    setEntityToDelete(null);
  }, []);

  return {
    // States
    isDeleteDialogOpen,
    entityToDelete,
    isDeleting,

    // Handlers
    handleDeleteEntity,
    confirmDelete,
    cancelDelete,

    // Setters for manual control
    setIsDeleteDialogOpen,
    setEntityToDelete,
  };
}
