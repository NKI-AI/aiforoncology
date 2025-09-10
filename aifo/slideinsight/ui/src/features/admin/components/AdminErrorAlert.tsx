// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { ErrorIcon } from "../../../components/icons";
import { Button } from "../../../components/ui/button";

interface AdminErrorAlertProps {
  error: string;
  loading: boolean;
  onRetry: () => void;
}

export function AdminErrorAlert({
  error,
  loading,
  onRetry,
}: AdminErrorAlertProps) {
  return (
    <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-r-lg mx-4 lg:mx-6">
      <div className="flex">
        <div className="flex-shrink-0">
          <ErrorIcon className="h-5 w-5 text-red-400" />
        </div>
        <div className="ml-3">
          <p className="text-sm text-red-700">
            <strong>Unable to load admin data:</strong> {error}
          </p>
          <p className="text-xs text-red-600 mt-1">
            Please check your network connection and ensure the admin endpoints
            are accessible.
          </p>
          <Button
            onClick={onRetry}
            disabled={loading}
            variant="link"
            className="mt-2 text-sm text-red-600 hover:text-red-800 h-auto p-0"
          >
            {loading ? "Retrying..." : "Try again"}
          </Button>
        </div>
      </div>
    </div>
  );
}
