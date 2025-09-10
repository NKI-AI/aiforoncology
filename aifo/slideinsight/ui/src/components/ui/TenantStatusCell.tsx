// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import StatusToggleCell from "./StatusToggleCell";
import { StatusType } from "./StatusDropdown";

interface Tenant {
  tenantUid: string;
  name: string;
  status: string;
}

interface TenantStatusCellProps {
  tenant: Tenant;
  onStatusUpdate?: (tenantUid: string, newStatus: StatusType) => void;
}

export const TenantStatusCell: React.FC<TenantStatusCellProps> = ({
  tenant,
  onStatusUpdate,
}) => {
  // Wrapper to ensure type safety
  const handleUpdate = (entityId: string, newValue: boolean | StatusType) => {
    if (onStatusUpdate && typeof newValue === "string") {
      onStatusUpdate(entityId, newValue as StatusType);
    }
  };

  return (
    <StatusToggleCell
      entity={tenant}
      config={{
        type: "status",
        apiEndpoint: (tenantUid) => `/api/v1/tenants/${tenantUid}`,
        apiField: "status",
        successMessage: (tenant, newStatus) => `Tenant status updated!`,
        errorMessage: "Failed to update tenant status",
      }}
      getEntityId={(tenant) => tenant.tenantUid}
      getEntityName={(tenant) => tenant.name}
      getCurrentValue={(tenant) => tenant.status as StatusType}
      onUpdate={handleUpdate}
    />
  );
};
