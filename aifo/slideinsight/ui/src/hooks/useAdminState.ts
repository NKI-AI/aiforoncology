// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useMemo } from "react";
import { useModalState, type ModalState } from "./useModalState";
import {
  useDeleteEntity,
  type DeleteEntityState,
  type DeleteEntityConfig,
} from "./useDeleteEntity";

interface AdminStateConfig<T> extends DeleteEntityConfig<T> {
  entityName: string;
  getEntityId: (entity: T) => string;
}

interface AdminState<T> extends ModalState<T>, DeleteEntityState<T> {
  // All modal and delete functionality combined
}

export function useAdminState<T>(
  config: AdminStateConfig<T>,
  refetch?: () => void
): AdminState<T> {
  // Prepare delete config
  const deleteConfig = useMemo(
    () => ({
      entityNameSingular: config.entityNameSingular,
      deleteEndpoint: config.deleteEndpoint,
      getEntityDisplayName: config.getEntityDisplayName,
    }),
    [config]
  );

  // Use the focused hooks
  const modalState = useModalState<T>(refetch);
  const deleteState = useDeleteEntity<T>(deleteConfig, refetch);

  // Combine the states
  return {
    ...modalState,
    ...deleteState,
  };
}
