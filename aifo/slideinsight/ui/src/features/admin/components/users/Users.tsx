// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useMemo, useState } from "react";
import {
  useUsers,
  useDeleteUser,
  type User,
  type UserQuery,
} from "../../../../api";
import { TrashIcon, SecurityIcon } from "../../../../components/icons";
import {
  createSelectionColumn,
  createTitleDescriptionColumn,
  createDateColumn,
  createActionsColumn,
} from "../../../../components/table-helpers/columnHelpers";
import { UserForm } from "../forms";
import { SendEmailForm } from "./SendEmailForm";
import { SendNotificationForm } from "./SendNotificationForm";
import UserRolesManager from "./UserRolesManager";
import StatusToggleCell from "../../../../components/ui/StatusToggleCell";
import {
  AdminEntityPage,
  type AdminEntityPageConfig,
} from "../AdminEntityPage";
import {
  useAdminEntityPage,
  type AdminEntityFilters,
} from "../../hooks/useAdminEntityPage";
import {
  type CommonUserFilters,
  initialCommonUserFilters,
  createUserFilterFields,
} from "../../utils/adminFilters";
import { Button } from "../../../../components/ui/button";
import { useNavigate } from "@tanstack/react-router";
import { FilterField } from "../../../../types/search";

// Enhanced filters interface that extends the base filters with API types
interface EnhancedUserFilters extends CommonUserFilters {
  q?: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  isActive?: boolean;
  mustResetPassword?: boolean;
  tenantUid?: string;
}

function UsersAdmin() {
  const navigate = useNavigate();

  // Configure CRUD operations
  const crudConfig = useMemo(
    () => ({
      entityName: "Users",
      entityNameSingular: "User",
      deleteEndpoint: (user: User) => `/api/v1/users/${user.userUid}`,
      getEntityDisplayName: (user: User) => user.email || "Unknown User",
      getEntityId: (user: User) => user.userUid || "",
    }),
    []
  );

  // Send email modal state
  const [isSendEmailModalOpen, setIsSendEmailModalOpen] = useState(false);
  const [selectedUserForEmail, setSelectedUserForEmail] = useState<User | null>(
    null
  );

  // Send notification modal state
  const [isSendNotificationModalOpen, setIsSendNotificationModalOpen] =
    useState(false);
  const [selectedUserForNotification, setSelectedUserForNotification] =
    useState<User | null>(null);

  // User roles modal state
  const [isUserRolesModalOpen, setIsUserRolesModalOpen] = useState(false);
  const [selectedUserForRoles, setSelectedUserForRoles] = useState<User | null>(
    null
  );

  // Set up the admin entity page state with optimistic updates
  const pageState = useAdminEntityPage(
    {
      crudConfig,
      initialFilters: {
        ...initialCommonUserFilters,
        q: "",
        email: "",
        firstName: "",
        lastName: "",
        isActive: undefined,
        mustResetPassword: undefined,
        tenantUid: "",
      } as EnhancedUserFilters,
      enableOptimisticUpdates: true,
      optimisticUpdateFields: ["isActive", "mustResetPassword"],
    },
    () => {} // Refetch will be handled by the new API hooks
  );

  // Prepare query for the new API hooks
  const userQuery: UserQuery = useMemo(() => {
    const {
      q,
      email,
      firstName,
      lastName,
      isActive,
      mustResetPassword,
      tenantUid,
      status,
      ...otherFilters
    } = pageState.filters;

    return {
      page: 1,
      limit: 20,
      q: q || "",
      email: email || undefined,
      firstName: firstName || undefined,
      lastName: lastName || undefined,
      isActive: isActive,
      mustResetPassword: mustResetPassword,
      tenantUid: tenantUid || undefined,
      // Include any other filters from pageState
      ...Object.fromEntries(
        Object.entries(otherFilters).filter(([, value]) => value)
      ),
    };
  }, [pageState.filters]);

  // Fetch users data using the new typed API
  const {
    data: users,
    pagination: serverPagination,
    loading,
    error,
    refetch,
  } = useUsers(userQuery);

  // Delete mutation
  const deleteUser = useDeleteUser({
    onSuccess: () => {
      // Close any open delete dialogs
      pageState.adminState.setIsDeleteDialogOpen(false);
    },
    onError: (error) => {
      console.error("Failed to delete user:", error.message);
      // You can add toast notifications or other error handling here
    },
  });

  // Enhanced delete handler that uses the typed mutation
  const handleDeleteUser = useCallback(
    (user: User) => {
      deleteUser.mutate(user.userUid);
    },
    [deleteUser]
  );

  // Send email handlers
  const handleSendEmail = useCallback((user: User) => {
    setSelectedUserForEmail(user);
    setIsSendEmailModalOpen(true);
  }, []);

  const handleSendEmailSuccess = useCallback(() => {
    setIsSendEmailModalOpen(false);
    setSelectedUserForEmail(null);
  }, []);

  const handleSendEmailCancel = useCallback(() => {
    setIsSendEmailModalOpen(false);
    setSelectedUserForEmail(null);
  }, []);

  // Send notification handlers
  const handleSendNotification = useCallback((user: User) => {
    setSelectedUserForNotification(user);
    setIsSendNotificationModalOpen(true);
  }, []);

  const handleSendNotificationSuccess = useCallback(() => {
    setIsSendNotificationModalOpen(false);
    setSelectedUserForNotification(null);
  }, []);

  const handleSendNotificationCancel = useCallback(() => {
    setIsSendNotificationModalOpen(false);
    setSelectedUserForNotification(null);
  }, []);

  // User roles handlers
  const handleManageRoles = useCallback((user: User) => {
    setSelectedUserForRoles(user);
    setIsUserRolesModalOpen(true);
  }, []);

  const handleUserRolesSuccess = useCallback(() => {
    setIsUserRolesModalOpen(false);
    setSelectedUserForRoles(null);
  }, []);

  const handleUserRolesCancel = useCallback(() => {
    setIsUserRolesModalOpen(false);
    setSelectedUserForRoles(null);
  }, []);

  // Table columns configuration using the new helpers
  const columns = useMemo(
    () => [
      createSelectionColumn<User>(),

      createTitleDescriptionColumn<User>({
        accessor: "firstName",
        header: "User",
        getTitle: (user) => {
          const fullName = [user.firstName, user.lastName]
            .filter(Boolean)
            .join(" ");
          return fullName || "Unknown User";
        },
        getDescription: (user) => user.email || null,
        sortable: true,
      }),

      {
        id: "isActive",
        header: "Active",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <StatusToggleCell
              entity={user}
              config={{
                type: "boolean",
                apiEndpoint: (userUid) => `/api/v1/users/${userUid}`,
                apiField: "isActive",
                variant: "toggle",
                successMessage: (user, newValue) =>
                  `User ${newValue ? "activated" : "deactivated"}!`,
                errorMessage: "Failed to update user status",
              }}
              getEntityId={(user) => user.userUid}
              getEntityName={(user) => user.email}
              getCurrentValue={(user) => user.isActive}
              onUpdate={(userUid, newValue) =>
                pageState.updateOptimistic(userUid, "isActive", newValue)
              }
            />
          );
        },
        enableSorting: false,
      },

      {
        id: "mustResetPassword",
        header: "Reset",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <StatusToggleCell
              entity={user}
              config={{
                type: "boolean",
                apiEndpoint: (userUid) => `/api/v1/users/${userUid}`,
                apiField: "mustResetPassword",
                variant: "toggle",
                trueColor: "bg-yellow-100 hover:bg-yellow-200 text-yellow-700",
                falseColor: "bg-green-100 hover:bg-green-200 text-green-700",
                successMessage: (user, newValue) =>
                  `Password reset ${newValue ? "required" : "not required"}!`,
                errorMessage: "Failed to update password reset requirement",
              }}
              getEntityId={(user) => user.userUid}
              getEntityName={(user) => user.email}
              getCurrentValue={(user) => user.mustResetPassword}
              onUpdate={(userUid, newValue) =>
                pageState.updateOptimistic(
                  userUid,
                  "mustResetPassword",
                  newValue
                )
              }
            />
          );
        },
        enableSorting: false,
      },

      {
        id: "roles",
        header: "Roles",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <Button
              variant="outline"
              size="sm"
              onClick={(event) => {
                event.stopPropagation();
                handleManageRoles(user);
              }}
              className="h-8 px-2 py-1"
              title="Manage user roles"
            >
              <SecurityIcon className="h-4 w-4" />
            </Button>
          );
        },
        enableSorting: false,
      },

      createDateColumn<User>({
        accessor: "createdAt",
        header: "Created",
        sortable: true,
      }),

      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
          const user = row.original;
          return (
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={(event) => {
                  event.stopPropagation();
                  navigate({ to: `/admin/users/${user.userUid}` });
                }}
                className="h-8 px-3 py-1"
                title="View user details"
              >
                View Details
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={(event) => {
                  event.stopPropagation();
                  handleDeleteUser(user);
                }}
                className="h-8 px-3 py-1"
                title="Delete user"
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            </div>
          );
        },
        enableSorting: false,
      },
    ],
    [
      handleDeleteUser,
      handleManageRoles,
      handleSendEmail,
      handleSendNotification,
      pageState.updateOptimistic,
      navigate,
    ]
  );

  // Filter fields configuration using centralized helper
  const filterFields: FilterField[] = useMemo(
    () =>
      createUserFilterFields({
        filters: pageState.filters,
        updateFilter: pageState.updateFilter,
      }),
    [pageState.filters, pageState.updateFilter]
  );

  // Form components for modals
  const CreateFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (user: User) => void;
      onCancel: () => void;
    }) => <UserForm entity={null} onSuccess={onSuccess} onCancel={onCancel} />,
    []
  );

  const EditFormComponent = useCallback(
    ({
      onSuccess,
      onCancel,
    }: {
      onSuccess: (user: User) => void;
      onCancel: () => void;
    }) => (
      <UserForm
        entity={pageState.adminState.selectedEntity}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    ),
    [pageState.adminState.selectedEntity]
  );

  // Configure the AdminEntityPage
  const pageConfig: AdminEntityPageConfig<User, EnhancedUserFilters> = {
    title: "Users",
    description: "Manage system users and permissions",
    searchPlaceholder: "Search users by email or name...",
    emptyMessage: "No users found.",

    entities: users, // Use the typed data directly
    loading: loading || deleteUser.isPending,
    error: error || (deleteUser.error?.message ?? null),
    refetch,
    entityConfig: crudConfig,

    pagination: serverPagination
      ? {
          totalPages: serverPagination.totalPages,
          totalItems: serverPagination.total,
        }
      : undefined,

    columns,
    filterFields,

    CreateFormComponent,
    EditFormComponent,

    modalMaxWidth: "lg",
    deleteDescription: (user) => (
      <>
        Are you sure you want to permanently delete "{user.email}"?
        <br />
        <br />
        <span className="font-semibold text-red-600">
          This action is permanent and cannot be undone.
        </span>
        <br />
        <br />
        <span className="text-sm text-muted-foreground">
          Note: This will only work if the user has no ownership of resources
          (studies, annotations, etc.). If the user owns any resources, you must
          transfer or delete those first.
        </span>
      </>
    ),

    onRowClick: (user) => {
      navigate({ to: `/admin/users/${user.userUid}` });
    },

    enablePagination: true,
  };

  return (
    <>
      <AdminEntityPage config={pageConfig} state={pageState} />

      {/* Custom modals not covered by AdminEntityPage */}
      {/* Send Email Modal */}
      {isSendEmailModalOpen && selectedUserForEmail && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
          <div className="bg-background rounded-lg max-w-md w-full p-6">
            <h2 className="text-lg font-semibold mb-4">Send Email</h2>
            <SendEmailForm
              user={selectedUserForEmail}
              onSuccess={handleSendEmailSuccess}
              onCancel={handleSendEmailCancel}
            />
          </div>
        </div>
      )}

      {/* Send Notification Modal */}
      {isSendNotificationModalOpen && selectedUserForNotification && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
          <div className="bg-background rounded-lg max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
            <h2 className="text-lg font-semibold mb-4">Send Notification</h2>
            <SendNotificationForm
              user={selectedUserForNotification}
              onSuccess={handleSendNotificationSuccess}
              onCancel={handleSendNotificationCancel}
            />
          </div>
        </div>
      )}

      {/* User Roles Modal */}
      {isUserRolesModalOpen && selectedUserForRoles && (
        <UserRolesManager
          isOpen={isUserRolesModalOpen}
          onClose={handleUserRolesCancel}
          user={selectedUserForRoles}
        />
      )}
    </>
  );
}

export default UsersAdmin;
