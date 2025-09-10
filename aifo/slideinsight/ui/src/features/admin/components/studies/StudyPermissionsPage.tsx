// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useCallback, useMemo } from "react";
import { useParams, useNavigate } from "@tanstack/react-router";
import { Button } from "../../../../components/ui/button";
import { Badge } from "../../../../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../../components/ui/card";
import { Separator } from "../../../../components/ui/separator";
import {
  ArrowLeft,
  UserPlus,
  Shield,
  Trash2,
  Users,
  Edit,
  X,
} from "lucide-react";
import { toast } from "sonner";
import { useStudyByUID } from "../../../../hooks/useStudies";
import {
  useObjectGrants,
  useDeleteObjectGrant,
  useBulkDeleteObjectGrants,
} from "../../hooks/useObjectGrants";
import { SecurityIcon } from "../../../../components/icons";
import AdminModal from "../AdminModal";
import StudyPermissionForm from "./StudyPermissionForm";
import type { ObjectGrant } from "../../hooks/useObjectGrants";

// Study-specific permissions
const STUDY_PERMISSIONS = [
  { name: "studies.view", description: "View study details and content" },
  { name: "studies.add_case", description: "Add cases to the study" },
  {
    name: "studies.modify_case",
    description: "Modify or remove cases in the study",
  },
  {
    name: "studies.annotate_case",
    description: "Create and edit annotations on study cases",
  },
];

interface UserPermissionGroup {
  user: {
    id: number;
    name: string;
    email: string;
    granteeType: "user" | "group" | "role";
  };
  permissions: Array<{
    name: string;
    description?: string;
    grantedAt: string;
    grantId: number;
  }>;
  earliestGrant: string;
  latestGrant: string;
}

function StudyPermissionsPage() {
  const { studyUid } = useParams({
    from: "/_authenticated/admin/studies/$studyUid/permissions",
  });
  const navigate = useNavigate();
  const [isAddPermissionModalOpen, setIsAddPermissionModalOpen] =
    useState(false);

  // Fetch study data
  const {
    data: study,
    isLoading: studyLoading,
    error: studyError,
  } = useStudyByUID(studyUid);

  // Fetch study permissions using the studyUid instead of internal ID
  const {
    data: grants = [],
    isLoading: grantsLoading,
    error: grantsError,
  } = useObjectGrants("study", studyUid);

  const deleteObjectGrant = useDeleteObjectGrant();
  const bulkDeleteObjectGrants = useBulkDeleteObjectGrants();

  // Group grants by user
  const userPermissionGroups = useMemo((): UserPermissionGroup[] => {
    const groupsMap = new Map<string, UserPermissionGroup>();

    grants.forEach((grant) => {
      const userKey = `${grant.grantee_type}-${grant.grantee_id}`;

      if (!groupsMap.has(userKey)) {
        groupsMap.set(userKey, {
          user: {
            id: grant.grantee_id,
            name:
              grant.grantee_info?.name ||
              `${grant.grantee_type} ${grant.grantee_id}`,
            email: grant.grantee_info?.email || `${grant.grantee_type}@unknown`,
            granteeType: grant.grantee_type,
          },
          permissions: [],
          earliestGrant: grant.created_at,
          latestGrant: grant.created_at,
        });
      }

      const group = groupsMap.get(userKey)!;
      const permissionInfo = STUDY_PERMISSIONS.find(
        (p) => p.name === grant.permission
      );

      group.permissions.push({
        name: grant.permission,
        description: permissionInfo?.description,
        grantedAt: grant.created_at,
        grantId: grant.id,
      });

      // Update earliest and latest grant dates
      if (new Date(grant.created_at) < new Date(group.earliestGrant)) {
        group.earliestGrant = grant.created_at;
      }
      if (new Date(grant.created_at) > new Date(group.latestGrant)) {
        group.latestGrant = grant.created_at;
      }
    });

    return Array.from(groupsMap.values()).sort((a, b) =>
      a.user.name.localeCompare(b.user.name)
    );
  }, [grants]);

  const handleBackToStudies = useCallback(() => {
    navigate({ to: "/admin/studies" });
  }, [navigate]);

  const handleAddPermission = useCallback(() => {
    setIsAddPermissionModalOpen(true);
  }, []);

  const handleCloseAddPermissionModal = useCallback(() => {
    setIsAddPermissionModalOpen(false);
  }, []);

  const handleEditUserPermissions = useCallback(
    (userGroup: UserPermissionGroup) => {
      // Open the permission form pre-populated with this user
      setIsAddPermissionModalOpen(true);
    },
    []
  );

  const handleRemoveUserPermissions = useCallback(
    async (userGroup: UserPermissionGroup) => {
      try {
        // Find all grants for this user
        const userGrants = grants.filter(
          (grant) =>
            grant.grantee_type === userGroup.user.granteeType &&
            grant.grantee_id === userGroup.user.id
        );

        if (userGrants.length === 1) {
          // Single permission - use regular delete
          await deleteObjectGrant.mutateAsync({
            resourceType: "study",
            resourceId: studyUid,
            data: {
              grantee_type: userGrants[0].grantee_type,
              grantee_id: userGrants[0].grantee_id,
              permission: userGrants[0].permission,
            },
          });
        } else if (userGrants.length > 1) {
          // Multiple permissions - use bulk delete
          await bulkDeleteObjectGrants.mutateAsync({
            resourceType: "study",
            resourceId: studyUid,
            grants: userGrants.map((grant) => ({
              grantee_type: grant.grantee_type,
              grantee_id: grant.grantee_id,
              permission: grant.permission,
            })),
          });
        }
      } catch (error) {
        console.error("Failed to remove user permissions:", error);
        toast.error("Failed to remove permissions", {
          description:
            "Please try again or contact support if the issue persists.",
        });
      }
    },
    [studyUid, grants, deleteObjectGrant, bulkDeleteObjectGrants]
  );

  const handleRemoveSinglePermission = useCallback(
    async (userGroup: UserPermissionGroup, permissionName: string) => {
      try {
        const grant = grants.find(
          (g) =>
            g.grantee_type === userGroup.user.granteeType &&
            g.grantee_id === userGroup.user.id &&
            g.permission === permissionName
        );

        if (grant) {
          await deleteObjectGrant.mutateAsync({
            resourceType: "study",
            resourceId: studyUid,
            data: {
              grantee_type: grant.grantee_type,
              grantee_id: grant.grantee_id,
              permission: grant.permission,
            },
          });
        }
      } catch (error) {
        console.error("Failed to remove permission:", error);
      }
    },
    [studyUid, grants, deleteObjectGrant]
  );

  const AddPermissionFormComponent = useCallback(
    ({
      entity,
      onSuccess,
      onCancel,
    }: {
      entity?: any;
      onSuccess: (entity: any) => void;
      onCancel: () => void;
    }) => (
      <StudyPermissionForm
        studyUid={studyUid}
        onSubmit={() => onSuccess({})}
        onCancel={onCancel}
        onSuccess={() => onSuccess({})}
      />
    ),
    [studyUid]
  );

  if (studyLoading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="h-4 bg-gray-200 rounded w-2/3"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (studyError || !study) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Error loading study
              </h3>
              <div className="mt-2 text-sm text-red-700">
                {studyError?.message ||
                  "Study not found. Please check the study ID and try again."}
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleBackToStudies}
            className="text-muted-foreground"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Studies
          </Button>
          <Separator orientation="vertical" className="h-6" />
          <div className="flex items-center space-x-3">
            <SecurityIcon className="h-8 w-8 text-green-500" />
            <div>
              <h1 className="text-2xl font-bold">Study Permissions</h1>
              <p className="text-muted-foreground">
                Manage user permissions for "{study.name}"
              </p>
            </div>
          </div>
        </div>

        <Button onClick={handleAddPermission}>
          <UserPlus className="h-4 w-4 mr-2" />
          Grant Permission
        </Button>
      </div>

      {/* Study Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Shield className="h-5 w-5 text-green-500" />
            <span>Study Information</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-muted-foreground">
                Study Name
              </label>
              <p className="font-medium">{study.name}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground">
                Study ID
              </label>
              <p className="font-mono text-sm">{study.studyUid}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground">
                Description
              </label>
              <p className="text-sm">
                {study.description || "No description provided"}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground">
                Status
              </label>
              <Badge variant={study.isPublished ? "default" : "secondary"}>
                {study.isPublished ? "Published" : "Draft"}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* User Permissions Table */}
      <Card>
        <CardHeader>
          <CardTitle>User Permissions</CardTitle>
          <CardDescription>
            Users with permissions to access this study
          </CardDescription>
        </CardHeader>
        <CardContent>
          {grantsLoading ? (
            <div className="py-8 text-center">
              <div className="animate-spin h-8 w-8 border-b-2 border-primary mx-auto"></div>
              <p className="mt-2 text-muted-foreground">
                Loading permissions...
              </p>
            </div>
          ) : grantsError ? (
            <div className="py-8 text-center text-red-600">
              <p>Failed to load permissions. Please try again.</p>
            </div>
          ) : userPermissionGroups.length === 0 ? (
            <div className="py-8 text-center">
              <Shield className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No permissions granted
              </h3>
              <p className="text-muted-foreground mb-4">
                This study has no specific user permissions yet.
              </p>
              <Button onClick={handleAddPermission} variant="outline">
                <UserPlus className="h-4 w-4 mr-2" />
                Grant First Permission
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {userPermissionGroups.map((userGroup) => (
                <div
                  key={`${userGroup.user.granteeType}-${userGroup.user.id}`}
                  className="rounded-lg border p-4 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3 flex-1">
                      <Users className="h-5 w-5 text-blue-500 mt-0.5" />
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <h4 className="font-medium">{userGroup.user.name}</h4>
                          <Badge variant="outline" className="text-xs">
                            {userGroup.user.granteeType}
                          </Badge>
                        </div>
                        {userGroup.user.email && (
                          <p className="text-sm text-muted-foreground mb-3">
                            {userGroup.user.email}
                          </p>
                        )}

                        {/* Permissions */}
                        <div className="space-y-2">
                          <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                            Permissions ({userGroup.permissions.length})
                          </label>
                          <div className="flex flex-wrap gap-2">
                            {userGroup.permissions.map((permission) => (
                              <div
                                key={permission.name}
                                className="group relative"
                              >
                                <Badge
                                  variant="secondary"
                                  className="pr-6 group-hover:pr-2 transition-all"
                                  title={permission.description}
                                >
                                  {permission.name}
                                  <button
                                    onClick={() =>
                                      handleRemoveSinglePermission(
                                        userGroup,
                                        permission.name
                                      )
                                    }
                                    className="absolute right-1 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 hover:bg-red-100 rounded-full p-0.5 transition-all"
                                    title="Remove this permission"
                                  >
                                    <X className="h-3 w-3 text-red-600" />
                                  </button>
                                </Badge>
                              </div>
                            ))}
                          </div>
                        </div>

                        {/* Grant date info */}
                        <div className="mt-3 pt-2 border-t">
                          <p className="text-xs text-muted-foreground">
                            {userGroup.permissions.length === 1
                              ? `Granted ${new Date(
                                  userGroup.permissions[0].grantedAt
                                ).toLocaleDateString()}`
                              : `First granted ${new Date(
                                  userGroup.earliestGrant
                                ).toLocaleDateString()}, last granted ${new Date(
                                  userGroup.latestGrant
                                ).toLocaleDateString()}`}
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center space-x-2 ml-4">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEditUserPermissions(userGroup)}
                        title="Edit permissions"
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveUserPermissions(userGroup)}
                        disabled={
                          deleteObjectGrant.isPending ||
                          bulkDeleteObjectGrants.isPending
                        }
                        className="text-red-600 hover:text-red-800 hover:bg-red-50"
                        title="Remove all permissions"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Add Permission Modal */}
      <AdminModal
        isOpen={isAddPermissionModalOpen}
        onClose={handleCloseAddPermissionModal}
        onSuccess={handleCloseAddPermissionModal}
        FormComponent={AddPermissionFormComponent}
        entityName="Permission"
        title="Grant Study Permission"
        maxWidth="lg"
      />
    </div>
  );
}

export default StudyPermissionsPage;
