// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { CloseIcon } from "../../../components/icons";
import { Button } from "../../../components/ui/button";

// Basic modal props (for simple usage)
interface BaseModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  maxWidth?: "sm" | "md" | "lg" | "xl" | "2xl";
  title?: string;
  showHeader?: boolean;
}

// Form modal props (for admin forms with entity handling)
interface FormModalProps<TEntity> {
  isOpen: boolean;
  entity?: TEntity | null; // null for create, TEntity for edit
  onClose: () => void;
  onSuccess: (entity: TEntity) => void;
  FormComponent: React.ComponentType<{
    entity?: TEntity | null;
    onSuccess: (entity: TEntity) => void;
    onCancel: () => void;
    isLoading?: boolean;
  }>;
  maxWidth?: "sm" | "md" | "lg" | "xl" | "2xl";
  title?: string;
  entityName?: string;
  children?: never; // Ensure children isn't used when FormComponent is provided
}

// Union type for all possible props
type AdminModalProps<TEntity = any> = BaseModalProps | FormModalProps<TEntity>;

const maxWidthClasses = {
  sm: "max-w-sm",
  md: "max-w-md",
  lg: "max-w-lg",
  xl: "max-w-xl",
  "2xl": "max-w-2xl",
};

// Type guard to check if props are for form modal
function isFormModal<T>(props: AdminModalProps<T>): props is FormModalProps<T> {
  return "FormComponent" in props;
}

function AdminModal<TEntity = any>(props: AdminModalProps<TEntity>) {
  const { isOpen, onClose, maxWidth = "lg" } = props;

  if (!isOpen) return null;

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose();
    }
  };

  React.useEffect(() => {
    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "hidden";

      return () => {
        document.removeEventListener("keydown", handleKeyDown);
        document.body.style.overflow = "unset";
      };
    }
  }, [isOpen]);

  // Determine content and header based on modal type
  let content: React.ReactNode;
  let shouldShowHeader = false;
  let title = "";

  if (isFormModal(props)) {
    // Form modal logic (from BaseAdminModal)
    const { entity, onSuccess, FormComponent, entityName } = props;

    const handleSuccess = (savedEntity: TEntity) => {
      onSuccess(savedEntity);
      onClose();
    };

    title =
      props.title ||
      (entity
        ? `Edit ${entityName || "Item"}`
        : `Create New ${entityName || "Item"}`);
    shouldShowHeader = false; // Forms handle their own headers

    content = (
      <FormComponent
        entity={entity}
        onSuccess={handleSuccess}
        onCancel={onClose}
      />
    );
  } else {
    // Basic modal logic (original AdminModal)
    const { children, showHeader = true } = props;
    title = props.title || "";
    shouldShowHeader = showHeader && Boolean(title);
    content = children;
  }

  return (
    <div
      className="fixed inset-0 z-50 overflow-y-auto"
      onClick={handleBackdropClick}
    >
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black bg-opacity-50 transition-opacity" />

      {/* Modal container */}
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Modal panel */}
        <div
          className={`relative w-full ${maxWidthClasses[maxWidth]} transform rounded-lg bg-background shadow-2xl border border-gray-200 transition-all`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Close button */}
          <Button
            type="button"
            onClick={onClose}
            variant="ghost"
            size="icon"
            className="absolute right-4 top-4 z-10 text-muted-400 hover:text-muted-600"
          >
            <span className="sr-only">Close</span>
            <CloseIcon className="h-5 w-5" />
          </Button>

          {/* Optional title */}
          {shouldShowHeader && (
            <div className="px-6 pt-6 pb-2">
              <h2 className="text-xl font-semibold text-muted-900 pr-8">
                {title}
              </h2>
            </div>
          )}

          {/* Modal content */}
          <div className={shouldShowHeader ? "px-6 pb-6" : "p-6"}>
            {content}
          </div>
        </div>
      </div>
    </div>
  );
}

export default AdminModal;
