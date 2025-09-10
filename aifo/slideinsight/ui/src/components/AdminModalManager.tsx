// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import AdminModal from "../features/admin/components/AdminModal";
import DeleteConfirmationDialog from "./DeleteConfirmationDialog";

interface AdminModalManagerProps<T> {
  // Create Modal
  isCreateModalOpen: boolean;
  onCreateModalClose: () => void;
  CreateFormComponent: React.ComponentType<{
    onSuccess: (entity: T) => void;
    onCancel: () => void;
  }>;

  // Edit Modal
  isEditModalOpen: boolean;
  onEditModalClose: () => void;
  selectedEntity: T | null;
  EditFormComponent: React.ComponentType<{
    onSuccess: (entity: T) => void;
    onCancel: () => void;
  }>;

  // Delete Dialog
  isDeleteDialogOpen: boolean;
  onDeleteDialogOpenChange: (open: boolean) => void;
  entityToDelete: T | null;
  isDeleting: boolean;
  onConfirmDelete: () => Promise<void>;
  onCancelDelete: () => void;

  // Form handlers
  onFormSuccess: () => void;
  onFormCancel: () => void;

  // Configuration
  entityName: string;
  entityNameSingular: string;
  getEntityDisplayName: (entity: T) => string;
  createModalTitle?: string;
  editModalTitle?: string;
  deleteTitle?: string;
  deleteDescription?: (entity: T) => React.ReactNode;
  modalMaxWidth?: "sm" | "md" | "lg" | "xl" | "2xl";

  // Custom delete confirmation content
  customDeleteContent?: React.ReactNode;
}

function AdminModalManager<T>({
  // Create Modal
  isCreateModalOpen,
  onCreateModalClose,
  CreateFormComponent,

  // Edit Modal
  isEditModalOpen,
  onEditModalClose,
  selectedEntity,
  EditFormComponent,

  // Delete Dialog
  isDeleteDialogOpen,
  onDeleteDialogOpenChange,
  entityToDelete,
  isDeleting,
  onConfirmDelete,
  onCancelDelete,

  // Form handlers
  onFormSuccess,
  onFormCancel,

  // Configuration
  entityName,
  entityNameSingular,
  getEntityDisplayName,
  createModalTitle,
  editModalTitle,
  deleteTitle,
  deleteDescription,
  modalMaxWidth = "md",
  customDeleteContent,
}: AdminModalManagerProps<T>) {
  const defaultCreateTitle = createModalTitle || `Create ${entityNameSingular}`;
  const defaultEditTitle = editModalTitle || `Edit ${entityNameSingular}`;
  const defaultDeleteTitle = deleteTitle || `Delete ${entityNameSingular}`;

  const defaultDeleteDescription = (entity: T) => (
    <>
      Are you sure you want to permanently delete "
      {getEntityDisplayName(entity)}"? This action cannot be undone.
    </>
  );

  return (
    <>
      {/* Create Modal */}
      <AdminModal
        isOpen={isCreateModalOpen}
        onClose={onCreateModalClose}
        onSuccess={onFormSuccess}
        FormComponent={CreateFormComponent}
        entityName={entityNameSingular}
        title={defaultCreateTitle}
        maxWidth={modalMaxWidth}
      />

      {/* Edit Modal */}
      <AdminModal
        isOpen={isEditModalOpen}
        onClose={onEditModalClose}
        onSuccess={onFormSuccess}
        FormComponent={EditFormComponent}
        entityName={entityNameSingular}
        title={defaultEditTitle}
        maxWidth={modalMaxWidth}
      />

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmationDialog
        isOpen={isDeleteDialogOpen}
        onOpenChange={onDeleteDialogOpenChange}
        title={defaultDeleteTitle}
        entityName={entityNameSingular}
        description={
          entityToDelete
            ? customDeleteContent ||
              (deleteDescription
                ? deleteDescription(entityToDelete)
                : defaultDeleteDescription(entityToDelete))
            : "Are you sure you want to delete this item?"
        }
        isDeleting={isDeleting}
        onConfirm={onConfirmDelete}
        onCancel={onCancelDelete}
      />
    </>
  );
}

export default AdminModalManager;
