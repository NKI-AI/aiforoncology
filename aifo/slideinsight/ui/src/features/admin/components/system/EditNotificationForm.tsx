// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, { useState } from "react";
import { apiFetch } from "@/utils/fetchUtils";

interface SystemNotification {
  id: number;
  title: string;
  body: string;
  link?: string;
  type: "info" | "success" | "warning" | "error";
  priority: "low" | "normal" | "high" | "urgent";
  isRead: boolean;
  isDismissed: boolean;
  expiresAt?: string;
  createdAt: string;
  readAt?: string;
  dismissedAt?: string;
}

interface EditNotificationFormProps {
  notification: SystemNotification;
  onSuccess: () => void;
  onCancel: () => void;
}

interface UpdateNotificationRequest {
  title: string;
  body: string;
  link?: string;
  type: string;
  priority: string;
  expiresAt?: string;
}

interface UpdateNotificationResponse {
  status: string;
}

export function EditNotificationForm({
  notification,
  onSuccess,
  onCancel,
}: EditNotificationFormProps) {
  const [formData, setFormData] = useState<UpdateNotificationRequest>({
    title: notification.title,
    body: notification.body,
    link: notification.link || "",
    type: notification.type,
    priority: notification.priority,
    expiresAt: notification.expiresAt || "",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const requestData: UpdateNotificationRequest = {
        title: formData.title,
        body: formData.body,
        type: formData.type,
        priority: formData.priority,
      };

      // Only include link if not empty
      if (formData.link && formData.link.trim()) {
        requestData.link = formData.link.trim();
      }

      // Only include expiresAt if not empty
      if (formData.expiresAt && formData.expiresAt.trim()) {
        requestData.expiresAt = formData.expiresAt.trim();
      }

      const response = await apiFetch<UpdateNotificationResponse>(
        `/api/v1/admin/notifications/${notification.id}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(requestData),
        }
      );

      if (response.status === "success") {
        onSuccess();
      } else {
        throw new Error("Failed to update notification");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to update notification"
      );
      console.error("Error updating notification:", err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleInputChange = (
    field: keyof UpdateNotificationRequest,
    value: string
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <div>
        <label
          htmlFor="title"
          className="block text-sm font-medium text-muted-700 mb-1"
        >
          Title *
        </label>
        <input
          type="text"
          id="title"
          value={formData.title}
          onChange={(e) => handleInputChange("title", e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      <div>
        <label
          htmlFor="body"
          className="block text-sm font-medium text-muted-700 mb-1"
        >
          Body *
        </label>
        <textarea
          id="body"
          value={formData.body}
          onChange={(e) => handleInputChange("body", e.target.value)}
          required
          rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      <div>
        <label
          htmlFor="link"
          className="block text-sm font-medium text-muted-700 mb-1"
        >
          Link (optional)
        </label>
        <input
          type="url"
          id="link"
          value={formData.link}
          onChange={(e) => handleInputChange("link", e.target.value)}
          placeholder="https://example.com"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label
            htmlFor="type"
            className="block text-sm font-medium text-muted-700 mb-1"
          >
            Type
          </label>
          <select
            id="type"
            value={formData.type}
            onChange={(e) => handleInputChange("type", e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="info">Info</option>
            <option value="success">Success</option>
            <option value="warning">Warning</option>
            <option value="error">Error</option>
          </select>
        </div>

        <div>
          <label
            htmlFor="priority"
            className="block text-sm font-medium text-muted-700 mb-1"
          >
            Priority
          </label>
          <select
            id="priority"
            value={formData.priority}
            onChange={(e) => handleInputChange("priority", e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="low">Low</option>
            <option value="normal">Normal</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>
        </div>
      </div>

      <div>
        <label
          htmlFor="expiresAt"
          className="block text-sm font-medium text-muted-700 mb-1"
        >
          Expires At (optional)
        </label>
        <input
          type="datetime-local"
          id="expiresAt"
          value={formData.expiresAt ? formData.expiresAt.slice(0, 16) : ""}
          onChange={(e) => {
            const value = e.target.value;
            if (value) {
              // Convert to ISO string
              const isoString = new Date(value).toISOString();
              handleInputChange("expiresAt", isoString);
            } else {
              handleInputChange("expiresAt", "");
            }
          }}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
        <p className="text-xs text-muted-500 mt-1">
          Leave empty for no expiration
        </p>
      </div>

      <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 text-sm font-medium text-muted-700 bg-background border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 disabled:opacity-50"
        >
          {isSubmitting ? "Updating..." : "Update Notification"}
        </button>
      </div>
    </form>
  );
}
