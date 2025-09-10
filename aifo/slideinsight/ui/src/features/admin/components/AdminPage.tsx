// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { useEffect } from "react";
import { useAdminData } from "@/features/admin/hooks/useAdminData";
import AdminPageLayout from "./AdminPageLayout";
import AdminStats from "./AdminStats";
import { AdminErrorAlert } from "./AdminErrorAlert";

export default function AdminPage() {
  const {
    users,
    tenants,
    studies,
    studiesCount,
    slidesCount,
    casesCount,
    loading,
    error,
    refetch,
  } = useAdminData();

  // Set document title
  useEffect(() => {
    document.title = "SlideInsight - Admin Dashboard";
    return () => {
      document.title = "SlideInsight Viewer";
    };
  }, []);

  // Ensure we have valid arrays even if API fails
  const safeUsers = Array.isArray(users) ? users : [];
  const safeTenants = Array.isArray(tenants) ? tenants : [];
  const safeStudies = Array.isArray(studies) ? studies : [];

  // Handle user deletion
  const handleDeleteUser = () => {
    // Refresh the data after deletion
    refetch();
  };

  const headerActions = (
    <button
      onClick={refetch}
      disabled={loading}
      className="inline-flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-400 text-white text-sm font-medium rounded-lg transition shadow-sm hover:shadow focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50"
    >
      {loading ? (
        <>
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
          Refreshing...
        </>
      ) : (
        "Refresh Data"
      )}
    </button>
  );

  return (
    <AdminPageLayout actions={headerActions}>
      <div className="space-y-6">
        {/* Error state */}
        {error && (
          <AdminErrorAlert error={error} loading={loading} onRetry={refetch} />
        )}

        {/* Statistics Cards */}
        <AdminStats
          users={safeUsers}
          tenants={safeTenants}
          studies={safeStudies}
          studiesCount={studiesCount}
          slides={slidesCount}
          cases={casesCount}
          loading={loading}
        />
      </div>
    </AdminPageLayout>
  );
}
