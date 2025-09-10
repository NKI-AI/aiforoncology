// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { ColumnDef } from "@tanstack/react-table";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
  createIdColumn,
  createActionsColumn,
} from "../../../components/table-helpers/columnHelpers";
import { TrashIcon } from "../../../components/icons";

interface BaseEntity {
  [key: string]: any;
}

interface AdminColumnConfig<T extends BaseEntity> {
  // Entity identification
  entityName: string;

  // Title/Description column
  titleConfig: {
    accessor: keyof T;
    header: string;
    getTitle: (entity: T) => string;
    getDescription?: (entity: T) => string | null;
  };

  // Optional columns
  includeSelection?: boolean;
  includeId?: boolean;
  includeCreatedDate?: boolean;
  includeUpdatedDate?: boolean;

  // ID column config (if enabled)
  idConfig?: {
    accessor: keyof T;
    header?: string;
    maxLength?: number;
  };

  // Date column configs
  createdDateConfig?: {
    accessor: keyof T;
    header?: string;
  };
  updatedDateConfig?: {
    accessor: keyof T;
    header?: string;
  };

  // Actions column
  actionsConfig: {
    onView?: (entity: T) => void;
    onEdit: (entity: T) => void;
    onDelete: (entity: T) => void;
    customActions?: Array<{
      label: string;
      onClick: (entity: T) => void;
      variant?: "default" | "destructive";
      icon?: React.ComponentType<any>;
    }>;
  };

  // Custom columns to insert between standard columns
  customColumns?: ColumnDef<T>[];
  customColumnsPosition?: "after-title" | "before-actions";
}

/**
 * Generates a standard set of columns for admin entity tables
 */
export function createStandardAdminColumns<T extends BaseEntity>(
  config: AdminColumnConfig<T>
): ColumnDef<T>[] {
  const columns: ColumnDef<T>[] = [];

  // Selection column
  if (config.includeSelection !== false) {
    columns.push(createSelectionColumn<T>());
  }

  // Title/Description column
  columns.push(
    createTitleDescriptionColumn<T>({
      accessor: config.titleConfig.accessor as string,
      header: config.titleConfig.header,
      getTitle: config.titleConfig.getTitle,
      getDescription: config.titleConfig.getDescription || (() => null),
      sortable: true,
    })
  );

  // Custom columns after title
  if (config.customColumns && config.customColumnsPosition === "after-title") {
    columns.push(...config.customColumns);
  }

  // ID column
  if (config.includeId && config.idConfig) {
    columns.push(
      createIdColumn<T>({
        accessor: config.idConfig.accessor as string,
        header: config.idConfig.header || "ID",
        maxLength: config.idConfig.maxLength || 8,
      })
    );
  }

  // Date columns
  if (config.includeCreatedDate && config.createdDateConfig) {
    columns.push(
      createDateColumn<T>({
        accessor: config.createdDateConfig.accessor as string,
        header: config.createdDateConfig.header || "Created",
        sortable: true,
      })
    );
  }

  if (config.includeUpdatedDate && config.updatedDateConfig) {
    columns.push(
      createDateColumn<T>({
        accessor: config.updatedDateConfig.accessor as string,
        header: config.updatedDateConfig.header || "Updated",
        sortable: true,
      })
    );
  }

  // Custom columns before actions
  if (
    config.customColumns &&
    config.customColumnsPosition === "before-actions"
  ) {
    columns.push(...config.customColumns);
  }

  // Actions column
  const actionsCustomActions = [
    ...(config.actionsConfig.customActions || []),
    {
      label: "Delete",
      onClick: config.actionsConfig.onDelete,
      variant: "destructive" as const,
      icon: TrashIcon,
    },
  ];

  columns.push(
    createActionsColumn<T>({
      onView: config.actionsConfig.onView,
      onEdit: config.actionsConfig.onEdit,
      entityName: config.entityName,
      customActions: actionsCustomActions,
    })
  );

  return columns;
}

/**
 * Creates columns specifically for user-like entities
 */
function createUserEntityColumns<T extends BaseEntity>(config: {
  onEdit: (entity: T) => void;
  onDelete: (entity: T) => void;
  customActions?: Array<{
    label: string;
    onClick: (entity: T) => void;
    variant?: "default" | "destructive";
    icon?: React.ComponentType<any>;
  }>;
  statusColumns?: ColumnDef<T>[];
}): ColumnDef<T>[] {
  return createStandardAdminColumns<T>({
    entityName: "User",
    titleConfig: {
      accessor: "firstName" as keyof T,
      header: "User",
      getTitle: (entity) => {
        const firstName = entity.firstName || "";
        const lastName = entity.lastName || "";
        const fullName = [firstName, lastName].filter(Boolean).join(" ");
        return fullName || "Unknown User";
      },
      getDescription: (entity) => entity.email || null,
    },
    includeCreatedDate: true,
    createdDateConfig: {
      accessor: "createdAt" as keyof T,
    },
    customColumns: config.statusColumns,
    customColumnsPosition: "after-title",
    actionsConfig: {
      onEdit: config.onEdit,
      onDelete: config.onDelete,
      customActions: config.customActions,
    },
  });
}

/**
 * Creates columns for simple name-based entities (roles, permissions, groups)
 */
export function createNameBasedEntityColumns<T extends BaseEntity>(config: {
  entityName: string;
  onView?: (entity: T) => void;
  onEdit: (entity: T) => void;
  onDelete: (entity: T) => void;
  customActions?: Array<{
    label: string;
    onClick: (entity: T) => void;
    variant?: "default" | "destructive";
    icon?: React.ComponentType<any>;
  }>;
  includeShortUid?: boolean;
  customColumns?: ColumnDef<T>[];
}): ColumnDef<T>[] {
  return createStandardAdminColumns<T>({
    entityName: config.entityName,
    titleConfig: {
      accessor: "name" as keyof T,
      header: config.entityName,
      getTitle: (entity) => entity.name || `Unnamed ${config.entityName}`,
      getDescription: (entity) => entity.description || null,
    },
    includeId: config.includeShortUid,
    idConfig: config.includeShortUid
      ? {
          accessor: "short_uid" as keyof T,
          header: "Short UID",
          maxLength: 8,
        }
      : undefined,
    includeCreatedDate: true,
    createdDateConfig: {
      accessor: "created_at" as keyof T,
    },
    customColumns: config.customColumns,
    customColumnsPosition: "before-actions",
    actionsConfig: {
      onView: config.onView,
      onEdit: config.onEdit,
      onDelete: config.onDelete,
      customActions: config.customActions,
    },
  });
}

/**
 * Creates columns for tenant entities
 */
export function createTenantEntityColumns<T extends BaseEntity>(config: {
  onView?: (entity: T) => void;
  onEdit: (entity: T) => void;
  onDelete: (entity: T) => void;
  customActions?: Array<{
    label: string;
    onClick: (entity: T) => void;
    variant?: "default" | "destructive";
    icon?: React.ComponentType<any>;
  }>;
  statusColumn?: ColumnDef<T>;
  domainsColumn?: ColumnDef<T>;
}): ColumnDef<T>[] {
  const customColumns: ColumnDef<T>[] = [];

  if (config.statusColumn) {
    customColumns.push(config.statusColumn);
  }

  if (config.domainsColumn) {
    customColumns.push(config.domainsColumn);
  }

  return createStandardAdminColumns<T>({
    entityName: "Tenant",
    titleConfig: {
      accessor: "name" as keyof T,
      header: "Tenant",
      getTitle: (entity) => entity.name || "Unnamed Tenant",
      getDescription: (entity) => entity.description || null,
    },
    includeId: true,
    idConfig: {
      accessor: "tenantUid" as keyof T,
      header: "ID",
      maxLength: 8,
    },
    includeCreatedDate: true,
    createdDateConfig: {
      accessor: "createdAt" as keyof T,
    },
    customColumns,
    customColumnsPosition: "after-title",
    actionsConfig: {
      onView: config.onView,
      onEdit: config.onEdit,
      onDelete: config.onDelete,
      customActions: config.customActions,
    },
  });
}
