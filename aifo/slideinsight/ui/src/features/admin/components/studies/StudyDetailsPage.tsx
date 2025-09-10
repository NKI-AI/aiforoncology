// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useCallback, useState, useEffect } from "react";
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
  Users,
  FileText,
  Calendar,
  User,
  Building,
} from "lucide-react";
import { useStudyByUID } from "../../../../hooks/useStudies";
import { SecurityIcon } from "../../../../components/icons";
import AdminSidebar from "../AdminSidebar";
import AdminHeader from "../AdminHeader";
import {
  SidebarInset,
  SidebarProvider,
} from "../../../../components/ui/sidebar";
import { StudyForm } from "../forms";
import { TrashIcon, EditIcon } from "../../../../components/icons";
import { formatDate } from "../../../../utils/adminTableUtils";
import UserCell from "../../../../components/UserCell";

function StudyDetailsPage() {
  const { studyUid } = useParams({
    from: "/_authenticated/admin/studies/$studyUid/",
  });
  const navigate = useNavigate();

  // Fetch study data
  const {
    data: study,
    isLoading: studyLoading,
    error: studyError,
  } = useStudyByUID(studyUid);

  const handleBackToStudies = useCallback(() => {
    navigate({ to: "/admin/studies" });
  }, [navigate]);

  const handleEditStudy = useCallback(() => {
    // TODO: Navigate to edit study page when implemented
  }, [studyUid]);

  const handleManagePermissions = useCallback(() => {
    navigate({ to: `/admin/studies/${studyUid}/permissions` });
  }, [navigate, studyUid]);

  const handleViewCases = useCallback(() => {
    // TODO: Navigate to study cases page when implemented
  }, [studyUid]);

  if (studyLoading) {
    return (
      <SidebarProvider>
        <AdminSidebar variant="inset" />
        <SidebarInset>
          <AdminHeader
            title="Study Details"
            description="Loading study information..."
          />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <div className="px-4 lg:px-6">
                  <div className="animate-pulse space-y-4">
                    <div className="h-8 bg-gray-200 rounded w-1/3"></div>
                    <div className="h-4 bg-gray-200 rounded w-2/3"></div>
                    <div className="h-64 bg-gray-200 rounded"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  if (studyError || !study) {
    return (
      <SidebarProvider>
        <AdminSidebar variant="inset" />
        <SidebarInset>
          <AdminHeader
            title="Study Details"
            description="Error loading study information"
          />
          <div className="flex flex-1 flex-col">
            <div className="@container/main flex flex-1 flex-col gap-2">
              <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <div className="px-4 lg:px-6">
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
              </div>
            </div>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  return (
    <SidebarProvider>
      <AdminSidebar variant="inset" />
      <SidebarInset>
        <AdminHeader
          title={study.name}
          description="Study Details"
          actions={
            <div className="flex space-x-2">
              <Button variant="outline" onClick={handleEditStudy}>
                <Edit className="h-4 w-4 mr-2" />
                Edit Study
              </Button>
              <Button onClick={handleManagePermissions}>
                <Shield className="h-4 w-4 mr-2" />
                Manage Permissions
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
                    onClick={handleBackToStudies}
                    className="text-muted-foreground"
                  >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Studies
                  </Button>
                </div>

                {/* Study Information Card */}
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <FileText className="h-5 w-5 text-blue-500" />
                      <span>Study Information</span>
                    </CardTitle>
                    <CardDescription>
                      Detailed information about this study
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <FileText className="h-4 w-4 mr-1" />
                          Study Name
                        </label>
                        <p className="font-medium text-lg">{study.name}</p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Building className="h-4 w-4 mr-1" />
                          Study ID
                        </label>
                        <p className="font-mono text-sm bg-muted px-2 py-1 rounded">
                          {study.studyUid}
                        </p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <User className="h-4 w-4 mr-1" />
                          Creator
                        </label>
                        <UserCell userUid={study.creatorUid} showIcon={true} />
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Badge className="h-4 w-4 mr-1" />
                          Status
                        </label>
                        <Badge
                          variant={study.isPublished ? "default" : "secondary"}
                        >
                          {study.isPublished ? "Published" : "Draft"}
                        </Badge>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Calendar className="h-4 w-4 mr-1" />
                          Created
                        </label>
                        <p className="text-sm">
                          {new Date(study.createdAt).toLocaleDateString()}
                        </p>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium text-muted-foreground flex items-center">
                          <Calendar className="h-4 w-4 mr-1" />
                          Last Updated
                        </label>
                        <p className="text-sm">
                          {new Date(study.updatedAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>

                    <Separator className="my-4" />

                    <div className="space-y-2">
                      <label className="text-sm font-medium text-muted-foreground">
                        Description
                      </label>
                      <p className="text-sm leading-relaxed">
                        {study.description || "No description provided"}
                      </p>
                    </div>
                  </CardContent>
                </Card>

                {/* Study Metrics Card */}
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                      <Users className="h-5 w-5 text-green-500" />
                      <span>Study Metrics</span>
                    </CardTitle>
                    <CardDescription>
                      Key metrics and statistics for this study
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                      <div className="text-center p-4 bg-blue-50 rounded-lg">
                        <div className="text-2xl font-bold text-blue-600">
                          {study.caseCount || 0}
                        </div>
                        <div className="text-sm text-blue-700">Cases</div>
                      </div>

                      <div className="text-center p-4 bg-green-50 rounded-lg">
                        <div className="text-2xl font-bold text-green-600">
                          {study.slideCount || 0}
                        </div>
                        <div className="text-sm text-green-700">Slides</div>
                      </div>

                      <div className="text-center p-4 bg-purple-50 rounded-lg">
                        <div className="text-2xl font-bold text-purple-600">
                          0
                        </div>
                        <div className="text-sm text-purple-700">
                          Annotations
                        </div>
                      </div>

                      <div className="text-center p-4 bg-orange-50 rounded-lg">
                        <div className="text-2xl font-bold text-orange-600">
                          0
                        </div>
                        <div className="text-sm text-orange-700">
                          Contributors
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Actions Card */}
                <Card>
                  <CardHeader>
                    <CardTitle>Actions</CardTitle>
                    <CardDescription>
                      Available actions for this study
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleViewCases}
                      >
                        <Users className="h-6 w-6 text-blue-500" />
                        <div className="text-center">
                          <div className="font-medium">View Cases</div>
                          <div className="text-xs text-muted-foreground">
                            Browse study cases and slides
                          </div>
                        </div>
                      </Button>

                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleManagePermissions}
                      >
                        <Shield className="h-6 w-6 text-green-500" />
                        <div className="text-center">
                          <div className="font-medium">Manage Permissions</div>
                          <div className="text-xs text-muted-foreground">
                            Control user access to this study
                          </div>
                        </div>
                      </Button>

                      <Button
                        variant="outline"
                        className="h-auto p-4 flex flex-col items-center space-y-2"
                        onClick={handleEditStudy}
                      >
                        <Edit className="h-6 w-6 text-orange-500" />
                        <div className="text-center">
                          <div className="font-medium">Edit Study</div>
                          <div className="text-xs text-muted-foreground">
                            Modify study details and settings
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
    </SidebarProvider>
  );
}

export default StudyDetailsPage;
