// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useCallback, useState } from "react";
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
  Edit,
  Shield,
  Mail,
  Calendar,
  User,
  Building,
  Trash2,
  Settings,
} from "lucide-react";
import { useUserByUID, useDeleteUser } from "../../../../api/hooks";
import { SecurityIcon, TrashIcon } from "../../../../components/icons";
import AdminSidebar from "../AdminSidebar";
import AdminHeader from "../AdminHeader";
import {
  SidebarInset,
  SidebarProvider,
} from "../../../../components/ui/sidebar";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../../../../components/AlertDialog";
import { toast } from "sonner";

function UserDetailPage() {
  const { userUid } = useParams({
    from: "/_authenticated/admin/users/$userUid",
  });
  const navigate = useNavigate();

  // State for delete confirmation dialog
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  // Fetch user data
  const {
    data: user,
    isLoading: userLoading,
    error: userError,
  } = useUserByUID(userUid);

  // Delete user mutation
  const deleteUser = useDeleteUser({
    onSuccess: () => {
      toast.success("User account has been disabled successfully");
      navigate({ to: "/admin/users" });
    },
    onError: (error) => {
      toast.error(`Failed to disable user: ${error.message}`);
    },
  });

  const handleBackToUsers = useCallback(() => {
    navigate({ to: "/admin/users" });
  }, [navigate]);

  const handleEditUser = useCallback(() => {
    // TODO: Navigate to edit user page or open edit modal
    toast.info("Edit functionality will be implemented soon");
  }, []);

  const handleManageRoles = useCallback(() => {
    // TODO: Open roles management modal
    toast.info("Role management functionality will be implemented soon");
  }, []);

  const handleSendEmail = useCallback(() => {
    // TODO: Open send email modal
    toast.info("Send email functionality will be implemented soon");
  }, []);

  const handleDeleteUser = useCallback(() => {
    setIsDeleteDialogOpen(true);
  }, []);

  const handleConfirmDelete = useCallback(() => {
    deleteUser.mutate(userUid);
    setIsDeleteDialogOpen(false);
  }, [deleteUser, userUid]);

  const handleCancelDelete = useCallback(() => {
    setIsDeleteDialogOpen(false);
  }, []);

  if (userLoading) {
    return (
      <SidebarProvider>
        <AdminSidebar variant="inset" />
        <SidebarInset>
          <AdminHeader
            title="User Details"
            description="Loading user information..."
          />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <div className="px-4 lg:px-6">
                  <div className="animate-pulse space-y-4">
                    <div className="h-8 bg-muted rounded w-1/3"></div>
                    <div className="h-4 bg-muted rounded w-2/3"></div>
                    <div className="h-64 bg-muted rounded"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  if (userError || !user) {
    return (
      <SidebarProvider>
        <AdminSidebar variant="inset" />
        <SidebarInset>
          <AdminHeader
            title="User Details"
            description="Error loading user information"
          />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <div className="px-4 lg:px-6">
                  <div className="bg-destructive/10 border border-destructive/20 rounded-md p-4">
                    <div className="flex">
                      <div className="ml-3">
                        <h3 className="text-sm font-medium text-destructive">
                          Error loading user
                        </h3>
                        <div className="mt-2 text-sm text-destructive/80">
                          {userError?.message ||
                            "User not found. Please check the user ID and try again."}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const fullName =
    [user.firstName, user.lastName].filter(Boolean).join(" ") || "Unknown User";

  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title={fullName}
          description="User Details"
          actions={
            <div className="flex space-x-2">
              <Button variant="outline" onClick={handleEditUser}>
                <Edit className="h-4 w-4 mr-2" />
                Edit User
              </Button>
              <Button variant="outline" onClick={handleManageRoles}>
                <Shield className="h-4 w-4 mr-2" />
                Manage Roles
              </Button>
              <Button variant="destructive" onClick={handleDeleteUser}>
                <Trash2 className="h-4 w-4 mr-2" />
                Disable Account
              </Button>
            </div>
          }
        />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              <div className="px-4 lg:px-6 space-y-6">
                {/* Back Button */}
                <div className="flex items-center">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleBackToUsers}
                    className="text-muted-foreground"
                  >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Users
                  </Button>
                </div>

                {/* User Information Card */}
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <User className="h-5 w-5 text-primary" />
                      <span>User Information</span>
                    </CardTitle>
                    <CardDescription>
                      Detailed information about this user account
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <User className="h-4 w-4 mr-1" />
                          Full Name
                        </label>
                        <p className="font-medium text-lg">{fullName}</p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Mail className="h-4 w-4 mr-1" />
                          Email Address
                        </label>
                        <p className="font-mono text-sm">{user.email}</p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Building className="h-4 w-4 mr-1" />
                          User ID
                        </label>
                        <p className="font-mono text-sm bg-muted/50 px-2 py-1 rounded border">
                          {user.userUid}
                        </p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Settings className="h-4 w-4 mr-1" />
                          Account Status
                        </label>
                        <Badge
                          variant={user.isActive ? "default" : "secondary"}
                        >
                          {user.isActive ? "Active" : "Inactive"}
                        </Badge>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Settings className="h-4 w-4 mr-1" />
                          Password Reset Required
                        </label>
                        <Badge
                          variant={
                            user.mustResetPassword ? "destructive" : "secondary"
                          }
                        >
                          {user.mustResetPassword ? "Required" : "Not Required"}
                        </Badge>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Mail className="h-4 w-4 mr-1" />
                          Email Verified
                        </label>
                        <Badge
                          variant={user.emailVerified ? "default" : "outline"}
                        >
                          {user.emailVerified ? "Verified" : "Not Verified"}
                        </Badge>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Calendar className="h-4 w-4 mr-1" />
                          Created
                        </label>
                        <p className="text-sm">
                          {new Date(user.createdAt).toLocaleDateString()}
                        </p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Calendar className="h-4 w-4 mr-1" />
                          Last Updated
                        </label>
                        <p className="text-sm">
                          {new Date(user.updatedAt).toLocaleDateString()}
                        </p>
                      </div>

                      {user.tenantUid && (
                        <div className="space-y-2">
                          <label className="text-sm font-medium text-muted-foreground flex items-center">
                            <Building className="h-4 w-4 mr-1" />
                            Tenant
                          </label>
                          <p className="font-mono text-sm bg-muted/50 px-2 py-1 rounded border">
                            {user.tenantUid}
                          </p>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>

                {/* Actions Card */}
                <Card>
                  <CardHeader>
                    <CardTitle>Actions</CardTitle>
                    <CardDescription>
                      Available actions for this user account
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleEditUser}
                      >
                        <Edit className="h-6 w-6 text-primary" />
                        <div className="text-center">
                          <div className="font-medium">Edit User</div>
                          <div className="text-xs text-muted-foreground">
                            Modify user details and settings
                          </div>
                        </div>
                      </Button>

                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleManageRoles}
                      >
                        <Shield className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                        <div className="text-center">
                          <div className="font-medium">Manage Roles</div>
                          <div className="text-xs text-muted-foreground">
                            Control user permissions and access
                          </div>
                        </div>
                      </Button>

                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleSendEmail}
                      >
                        <Mail className="h-6 w-6 text-orange-600 dark:text-orange-400" />
                        <div className="text-center">
                          <div className="font-medium">Send Email</div>
                          <div className="text-xs text-muted-foreground">
                            Send a message to this user
                          </div>
                        </div>
                      </Button>

                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2 border-destructive/20 hover:bg-destructive/5"
                        onClick={handleDeleteUser}
                      >
                        <Trash2 className="h-6 w-6 text-destructive" />
                        <div className="text-center">
                          <div className="font-medium text-destructive">
                            Disable Account
                          </div>
                          <div className="text-xs text-muted-foreground">
                            Permanently disable this user account
                          </div>
                        </div>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={isDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Disable User Account</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to disable the account for "{user.email}"?
              <br />
              <br />
              <span className="font-semibold text-amber-600 dark:text-amber-400">
                We recommend disabling the account instead of permanent
                deletion.
              </span>
              <br />
              <br />
              <span className="text-sm text-muted-foreground">
                Disabling the account will prevent the user from logging in
                while preserving their data and any resources they own. The
                account can be re-enabled later if needed.
              </span>
              <br />
              <br />
              <span className="text-sm text-muted-foreground">
                Note: Permanent deletion will only work if the user has no
                ownership of resources (studies, annotations, etc.). If the user
                owns any resources, you must transfer or delete those first.
              </span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCancelDelete}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              className="bg-amber-600 hover:bg-amber-700 dark:bg-amber-700 dark:hover:bg-amber-600 text-white"
            >
              Disable Account
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </SidebarProvider>
  );
}

export default UserDetailPage;
