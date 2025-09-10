// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "@tanstack/react-router";
import { usePermissionExplanation } from "@/hooks/usePermissions";
import { useStudy } from "@/api/hooks";
import { useStudyCases } from "@/hooks/useStudyCases";
import { isAuthorizationError } from "@/utils/errorUtils";
import StudyCasesHeader from "@/components/StudyCasesHeader";
import StudyCasesTable from "@/components/StudyCasesTable";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import NoAccessState from "@/components/NoAccessState";
import { Card, CardContent } from "@/components/ui/card";

const StudyCases: React.FC = () => {
  const { studyUid } = useParams({
    from: "/_authenticated/studies/$studyUid/",
  });
  const navigate = useNavigate();
  const [showPermissionDetails, setShowPermissionDetails] = useState(false);

  // Fetch study details using the centralized hook
  const {
    data: study,
    isLoading: studyLoading,
    error: studyError,
    refetch: refetchStudy,
  } = useStudy(studyUid);

  // Fetch permission explanation for the study
  const {
    data: permissionExplanation,
    isLoading: permissionLoading,
    error: permissionError,
  } = usePermissionExplanation("study", studyUid, "studies.view", {
    enabled: !!studyUid,
  });

  // Use the centralized study cases hook
  const {
    cases,
    pagination,
    loading: casesLoading,
    error: casesError,
    refetch: refetchCases,
    filters,
    updateFilter,
    clearFilters,
    hasActiveFilters,
    currentPage,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
  } = useStudyCases({ studyUid });

  // Set document title
  useEffect(() => {
    if (study) {
      document.title = `SlideInsight - ${study.name}`;
    } else {
      document.title = "SlideInsight - Study Cases";
    }
    return () => {
      document.title = "SlideInsight Viewer";
    };
  }, [study]);

  // Determine title based on study loading state
  const getTitle = () => {
    if (studyLoading) return "Loading Study...";
    if (isAuthorizationError(studyError)) return "Access Denied";
    if (studyError) return "Error Loading Study";
    if (study) return `${study.name} - Cases`;
    return "Study Cases";
  };

  const getSubtitle = () => {
    if (studyLoading) return "Please wait while we load the study information";
    if (isAuthorizationError(studyError))
      return "You don't have permission to view this study";
    if (studyError) return "There was an error loading the study details";
    if (study?.description) return study.description;
    return "Browse study cases and slides";
  };

  // Check if either study or cases have authorization errors
  const hasAuthError =
    isAuthorizationError(studyError) || isAuthorizationError(casesError);
  const hasOtherError =
    (studyError && !isAuthorizationError(studyError)) ||
    (casesError && !isAuthorizationError(casesError));

  return (
    <div className="container mx-auto max-w-7xl px-4 py-6">
      <Card className="shadow-sm">
        <CardContent className="p-6">
          {hasAuthError ? (
            <NoAccessState
              title="Access Denied"
              description="You don't have permission to view this study. Please contact an administrator if you believe you should have access to this content."
              resourceId={studyUid}
              resourceType="Study"
              permissionExplanation={permissionExplanation}
              showBackToStudies={true}
              backToStudiesPath="/studies"
            />
          ) : hasOtherError ? (
            <ErrorStateAlert
              error={studyError || casesError}
              title={
                studyError
                  ? "Failed to load study details"
                  : "Failed to load cases"
              }
              variant="detailed"
              className="mb-6"
              onRetry={studyError ? refetchStudy : refetchCases}
              retryText="Retry"
            />
          ) : (
            <>
              {/* Study header component */}
              <StudyCasesHeader
                study={study ?? null}
                permissionExplanation={permissionExplanation ?? null}
                showPermissionDetails={showPermissionDetails}
                onTogglePermissionDetails={() =>
                  setShowPermissionDetails(!showPermissionDetails)
                }
              />

              {/* Study cases table component */}
              <StudyCasesTable
                studyUid={studyUid}
                cases={cases}
                loading={casesLoading || studyLoading}
                pagination={pagination}
                filters={filters}
                hasActiveFilters={hasActiveFilters}
                currentPage={currentPage}
                pageSize={pageSize}
                onFilterChange={updateFilter}
                onClearFilters={clearFilters}
                onPageChange={handlePageChange}
                onPageSizeChange={handlePageSizeChange}
                getTitle={getTitle}
                getSubtitle={getSubtitle}
              />
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default StudyCases;
