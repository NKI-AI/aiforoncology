// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useForm } from "@tanstack/react-form";
import { toast } from "sonner";
import { apiFetch } from "@/utils/fetchUtils";
import { queryKeys } from "@/utils/apiQueries";
import { formatDate } from "@/utils/format";
import { copyToClipboard } from "@/utils/clipboardUtils";
import {
  CpuChipIcon,
  PlayIcon,
  StopIcon,
  ClockIcon,
  CheckIcon,
  CloseIcon,
} from "@/components/icons";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Algorithm, AlgorithmRun } from "@/features/admin/hooks/useAdminData";
import AdminPageLayout from "@/features/admin/components/AdminPageLayout";
import ErrorStateAlert from "@/components/ErrorStateAlert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/AlertDialog";

interface AlgorithmDetailPageProps {
  algorithmId: string;
}

interface CreateRunFormData {
  caseId?: string;
  parameters?: string; // JSON string
  executionMode: "BATCH" | "STREAM";
}

interface EditableAlgorithmField {
  isEditing: boolean;
  value: string;
  originalValue: string;
}

export default function AlgorithmDetailPage({
  algorithmId,
}: AlgorithmDetailPageProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [runToCancel, setRunToCancel] = useState<AlgorithmRun | null>(null);

  // Inline editing state
  const [editingName, setEditingName] = useState<EditableAlgorithmField>({
    isEditing: false,
    value: "",
    originalValue: "",
  });
  const [editingDescription, setEditingDescription] =
    useState<EditableAlgorithmField>({
      isEditing: false,
      value: "",
      originalValue: "",
    });
  const [isSavingAlgorithm, setIsSavingAlgorithm] = useState(false);

  // Fetch algorithm details
  const {
    data: algorithm,
    isLoading: algorithmLoading,
    error: algorithmError,
    refetch: refetchAlgorithm,
  } = useQuery({
    queryKey: queryKeys.algorithms.detail(algorithmId),
    queryFn: async () => {
      const response = await apiFetch<Algorithm>(
        `/api/v1/algorithms/${algorithmId}`
      );
      return response;
    },
    staleTime: 30 * 1000,
    gcTime: 5 * 60 * 1000,
  });

  // Fetch algorithm runs
  const {
    data: runsData,
    isLoading: runsLoading,
    error: runsError,
    refetch: refetchRuns,
  } = useQuery({
    queryKey: queryKeys.algorithms.runs(algorithmId),
    queryFn: async () => {
      const response = await apiFetch<{ runs: AlgorithmRun[] }>(
        `/api/v1/algorithms/${algorithmId}/runs`
      );
      return response.runs || [];
    },
    staleTime: 10 * 1000, // More frequent updates for runs
    gcTime: 5 * 60 * 1000,
  });

  const runs = runsData || [];

  // Form for creating new runs
  const form = useForm({
    defaultValues: {
      caseId: "",
      parameters: "{}",
      executionMode: "BATCH",
    } as CreateRunFormData,
    onSubmit: async ({ value }) => {
      setIsSubmitting(true);
      setFormError(null);

      try {
        let parameters = {};
        if (value.parameters?.trim()) {
          try {
            parameters = JSON.parse(value.parameters);
          } catch (e) {
            throw new Error("Invalid JSON in parameters field");
          }
        }

        await apiFetch(`/api/v1/runs`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            algorithmId: algorithmId,
            caseId: value.caseId || null,
            parameters: parameters,
            executionMode: value.executionMode,
          }),
        });

        form.reset();
        await refetchRuns();
        toast.success("Algorithm run started!", {
          description: `New ${value.executionMode.toLowerCase()} run has been queued for ${
            algorithm?.name
          }.`,
        });
      } catch (err) {
        console.error("Failed to create run:", err);
        const errorMessage =
          err instanceof Error
            ? err.message
            : "Failed to start algorithm run. Please try again.";
        setFormError(errorMessage);
        toast.error("Failed to start run", {
          description: errorMessage,
        });
      } finally {
        setIsSubmitting(false);
      }
    },
  });

  const cancelRun = async (run: AlgorithmRun) => {
    if (!algorithm) return;

    try {
      setFormError(null);

      await apiFetch(`/api/v1/runs/${run.id}/cancel`, {
        method: "POST",
      });

      await refetchRuns();
      toast.success("Run cancelled!", {
        description: `Algorithm run ${run.id.substring(
          0,
          8
        )} has been cancelled.`,
      });
    } catch (err) {
      console.error("Failed to cancel run:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to cancel run.";
      setFormError(errorMessage);
      toast.error("Failed to cancel run", {
        description: errorMessage,
      });
    } finally {
      setCancelDialogOpen(false);
      setRunToCancel(null);
    }
  };

  const handleCancelClick = (run: AlgorithmRun) => {
    setRunToCancel(run);
    setCancelDialogOpen(true);
  };

  const handleCancelConfirm = () => {
    if (runToCancel) {
      cancelRun(runToCancel);
    }
  };

  const handleCancelCancel = () => {
    setCancelDialogOpen(false);
    setRunToCancel(null);
  };

  // Inline editing functions (similar to tenant detail page)
  const startEditing = (
    field: "name" | "description",
    currentValue: string
  ) => {
    const editState = {
      isEditing: true,
      value: currentValue,
      originalValue: currentValue,
    };

    if (field === "name") {
      setEditingName(editState);
    } else {
      setEditingDescription(editState);
    }
  };

  const cancelEditing = (field: "name" | "description") => {
    if (field === "name") {
      setEditingName((prev) => ({
        ...prev,
        isEditing: false,
        value: prev.originalValue,
      }));
    } else {
      setEditingDescription((prev) => ({
        ...prev,
        isEditing: false,
        value: prev.originalValue,
      }));
    }
  };

  const saveAlgorithmField = async (field: "name" | "description") => {
    if (!algorithm) return;

    const fieldValue =
      field === "name"
        ? editingName.value.trim()
        : editingDescription.value.trim();

    if (field === "name" && !fieldValue) {
      toast.error("Algorithm name cannot be empty");
      return;
    }

    setIsSavingAlgorithm(true);
    try {
      const updateData = {
        [field]: fieldValue || (field === "description" ? null : fieldValue),
      };

      await apiFetch(`/api/v1/algorithms/${algorithm.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateData),
      });

      await refetchAlgorithm();

      // Reset editing state
      if (field === "name") {
        setEditingName({
          isEditing: false,
          value: fieldValue,
          originalValue: fieldValue,
        });
      } else {
        setEditingDescription({
          isEditing: false,
          value: fieldValue,
          originalValue: fieldValue,
        });
      }

      toast.success(`Algorithm ${field} updated successfully!`);
    } catch (err) {
      console.error(`Failed to update algorithm ${field}:`, err);
      const errorMessage =
        err instanceof Error
          ? err.message
          : `Failed to update algorithm ${field}.`;
      toast.error(`Failed to update algorithm ${field}`, {
        description: errorMessage,
      });
    } finally {
      setIsSavingAlgorithm(false);
    }
  };

  const updateFieldValue = (field: "name" | "description", value: string) => {
    if (field === "name") {
      setEditingName((prev) => ({ ...prev, value }));
    } else {
      setEditingDescription((prev) => ({ ...prev, value }));
    }
  };

  // Set initial values when algorithm data loads
  React.useEffect(() => {
    if (algorithm && !editingName.originalValue) {
      setEditingName({
        isEditing: false,
        value: algorithm.name,
        originalValue: algorithm.name,
      });
    }
    if (algorithm && !editingDescription.originalValue) {
      setEditingDescription({
        isEditing: false,
        value: algorithm.description || "",
        originalValue: algorithm.description || "",
      });
    }
  }, [algorithm, editingName.originalValue, editingDescription.originalValue]);

  // Status color helper
  const getStatusColor = (status: string) => {
    switch (status) {
      case "QUEUED":
        return "bg-yellow-100 text-yellow-800";
      case "RUNNING":
        return "bg-blue-100 text-blue-800";
      case "SUCCEEDED":
        return "bg-green-100 text-green-800";
      case "FAILED":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-muted-800";
    }
  };

  // Loading state
  if (algorithmLoading) {
    return (
      <AdminPageLayout
        title="Algorithm Details"
        description="Loading algorithm information"
        actions={
          <Link
            to="/admin/algorithms"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            text-muted ← Back to Algorithms
          </Link>
        }
      >
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </AdminPageLayout>
    );
  }

  // Error state for algorithm loading
  if (algorithmError || !algorithm) {
    return (
      <AdminPageLayout
        title="Algorithm Details"
        description="Error loading algorithm information"
        actions={
          <Link
            to="/admin/algorithms"
            className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
          >
            text-muted ← Back to Algorithms
          </Link>
        }
      >
        <ErrorStateAlert
          error={algorithmError || "Failed to load algorithm details"}
          title="Error Loading Algorithm"
          onRetry={refetchAlgorithm}
          variant="detailed"
        />
      </AdminPageLayout>
    );
  }

  const activeRuns = runs.filter(
    (r) => r.status === "RUNNING" || r.status === "QUEUED"
  ).length;
  const completedRuns = runs.filter((r) => r.status === "SUCCEEDED").length;
  const failedRuns = runs.filter((r) => r.status === "FAILED").length;

  return (
    <AdminPageLayout
      title={algorithm.name}
      description={`Algorithm ID: ${algorithm.id}`}
      actions={
        <Link
          to="/admin/algorithms"
          className="inline-flex items-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-muted-700 text-sm font-medium rounded-md transition"
        >
          ← Back to Algorithms
        </Link>
      }
    >
      <div className="bg-background rounded-lg shadow-sm border border-gray-200 p-6">
        {/* Header */}
        <div className="flex items-center space-x-3 mb-6">
          <CpuChipIcon className="h-8 w-8 text-blue-500" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-muted-900">
              {algorithm.name}
            </h1>
            <div className="flex items-center space-x-4 mt-1">
              <p className="text-sm text-muted-600">
                Algorithm ID: {algorithm.id}
              </p>
              <Badge variant="outline" className="font-mono">
                v{algorithm.version}
              </Badge>
              <Badge
                variant={
                  algorithm.executionMode === "STREAM" ? "default" : "secondary"
                }
                className={
                  algorithm.executionMode === "STREAM"
                    ? "bg-blue-100 text-blue-800"
                    : "bg-gray-100 text-muted-800"
                }
              >
                {algorithm.executionMode}
              </Badge>
            </div>
          </div>
        </div>

        {/* Algorithm Information and Run Statistics */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Algorithm Information
              </h3>
              <div className="space-y-3">
                <div>
                  <dt className="text-sm font-medium text-muted-500">Name</dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {algorithm.name}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Description
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {algorithm.description || "No description provided"}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Version
                  </dt>
                  <dd className="text-sm text-muted-900 font-mono mt-1">
                    {algorithm.version}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Endpoint URL
                  </dt>
                  <dd className="text-sm text-muted-900 font-mono mt-1 break-all">
                    {algorithm.endpointUrl}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    HTTP Method
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    <Badge variant="outline">{algorithm.httpMethod}</Badge>
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Progress Transport
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    <Badge variant="outline">
                      {algorithm.progressTransport}
                    </Badge>
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-500">
                    Created
                  </dt>
                  <dd className="text-sm text-muted-900 mt-1">
                    {formatDate(algorithm.createdAt)}
                  </dd>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-muted-900 mb-3">
                Run Statistics
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <PlayIcon className="h-5 w-5 text-blue-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-blue-900">
                        {activeRuns}
                      </p>
                      <p className="text-sm text-blue-700">Active Runs</p>
                    </div>
                  </div>
                </div>
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <CheckIcon className="h-5 w-5 text-green-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-green-900">
                        {completedRuns}
                      </p>
                      <p className="text-sm text-green-700">Completed</p>
                    </div>
                  </div>
                </div>
                <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <CloseIcon className="h-5 w-5 text-red-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-red-900">
                        {failedRuns}
                      </p>
                      <p className="text-sm text-red-700">Failed</p>
                    </div>
                  </div>
                </div>
                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <ClockIcon className="h-5 w-5 text-muted-500 mr-2" />
                    <div>
                      <p className="text-2xl font-bold text-muted-900">
                        {runs.length}
                      </p>
                      <p className="text-sm text-muted-700">Total Runs</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Create Run Form */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Start New Run
          </h3>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <p className="text-sm text-muted-600 mb-4">
              Start a new execution of this algorithm. You can specify a case ID
              and custom parameters.
            </p>

            {formError && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3 mb-4">
                <p className="text-sm text-red-600">{formError}</p>
              </div>
            )}

            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
              }}
            >
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <form.Field name="caseId">
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        Case ID (optional)
                      </label>
                      <input
                        id={field.name}
                        name={field.name}
                        type="text"
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        placeholder="case-123"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        disabled={isSubmitting}
                      />
                    </div>
                  )}
                </form.Field>

                <form.Field name="executionMode">
                  {(field) => (
                    <div>
                      <label
                        htmlFor={field.name}
                        className="block text-sm font-medium text-muted-700 mb-1"
                      >
                        Execution Mode
                      </label>
                      <select
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onChange={(e) =>
                          field.handleChange(
                            e.target.value as "BATCH" | "STREAM"
                          )
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        disabled={isSubmitting}
                      >
                        <option value="BATCH">Batch Processing</option>
                        <option value="STREAM">Stream Processing</option>
                      </select>
                    </div>
                  )}
                </form.Field>

                <div className="flex items-end">
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full inline-flex items-center justify-center"
                  >
                    <PlayIcon className="h-4 w-4 mr-2" />
                    {isSubmitting ? "Starting..." : "Start Run"}
                  </Button>
                </div>
              </div>

              <form.Field name="parameters">
                {(field) => (
                  <div className="mt-4">
                    <label
                      htmlFor={field.name}
                      className="block text-sm font-medium text-muted-700 mb-1"
                    >
                      Parameters (JSON)
                    </label>
                    <textarea
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      placeholder='{"threshold": 0.5, "model": "default"}'
                      rows={3}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                      disabled={isSubmitting}
                    />
                    <p className="mt-1 text-xs text-muted-500">
                      Enter parameters as JSON. Leave empty for default
                      parameters.
                    </p>
                  </div>
                )}
              </form.Field>
            </form>
          </div>
        </div>

        {/* Recent Runs */}
        <div>
          <h3 className="text-lg font-semibold text-muted-900 mb-4">
            Recent Runs
          </h3>

          {runsLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="animate-pulse bg-gray-100 h-16 rounded-lg"
                ></div>
              ))}
            </div>
          ) : runsError ? (
            <ErrorStateAlert
              error={runsError}
              title="Failed to load runs"
              onRetry={refetchRuns}
              variant="inline"
            />
          ) : runs.length === 0 ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
              <PlayIcon className="h-12 w-12 text-muted-400 mx-auto mb-4" />
              <h4 className="text-lg font-medium text-muted-900 mb-2">
                No runs yet
              </h4>
              <p className="text-muted-500">
                Use the form above to start your first algorithm run.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {runs.slice(0, 10).map((run) => (
                <div
                  key={run.id}
                  className="bg-background border border-gray-200 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <PlayIcon className="h-5 w-5 text-muted-400" />
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-muted-900 font-mono">
                            {run.id.substring(0, 8)}...
                          </span>
                          <Badge className={getStatusColor(run.status)}>
                            {run.status}
                          </Badge>
                          <Badge variant="outline">{run.executionMode}</Badge>
                          {run.progress > 0 && (
                            <span className="text-sm text-muted-500">
                              {run.progress}%
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-muted-500 mt-1">
                          Started {formatDate(run.createdAt)}
                          {run.finishedAt && (
                            <span>
                              {" "}
                              • Finished {formatDate(run.finishedAt)}
                            </span>
                          )}
                          {run.caseId && <span> • Case: {run.caseId}</span>}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      {(run.status === "RUNNING" ||
                        run.status === "QUEUED") && (
                        <Button
                          onClick={() => handleCancelClick(run)}
                          variant="outline"
                          size="sm"
                          className="border-red-300 text-red-700 hover:bg-red-50"
                        >
                          <StopIcon className="h-4 w-4 mr-1" />
                          Cancel
                        </Button>
                      )}
                      <Button
                        onClick={() => copyToClipboard(run.id)}
                        variant="outline"
                        size="sm"
                      >
                        Copy ID
                      </Button>
                    </div>
                  </div>
                </div>
              ))}

              {runs.length > 10 && (
                <div className="text-center pt-4">
                  <Button variant="outline">
                    View All Runs ({runs.length})
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Cancel Run Confirmation Dialog */}
      <AlertDialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Cancel Algorithm Run</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to cancel the run "
              {runToCancel?.id.substring(0, 8)}..."? This action cannot be
              undone and may affect any processing in progress.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCancelCancel}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelConfirm}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
            >
              Cancel Run
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AdminPageLayout>
  );
}
