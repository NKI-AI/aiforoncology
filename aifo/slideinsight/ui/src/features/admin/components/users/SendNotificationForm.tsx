// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { User } from "../../../../api/models";
import { useUsers } from "../../hooks/useUsers";
import { Button } from "../../../../components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../../components/ui/select";
import { Input } from "../../../../components/ui/input";
import { Label } from "../../../../components/ui/label";
import { Textarea } from "../../../../components/ui/textarea";
import { toast } from "sonner";
import { useNotificationContext } from "../../../../contexts/NotificationContext";
import { UserIcon } from "../../../../components/icons";
import { showUserNotificationToast } from "../../../../components/NotificationToast";
import { apiFetch, ApiError } from "../../../../utils/fetchUtils";

interface SendNotificationFormProps {
  user?: User; // Optional - if not provided, user selection will be shown
  onSuccess: () => void;
  onCancel: () => void;
}

type NotificationType = "info" | "success" | "warning" | "error";
type NotificationPriority = "low" | "normal" | "high" | "urgent";

interface SendNotificationApiResponse {
  status: string;
}

export function SendNotificationForm({
  user,
  onSuccess,
  onCancel,
}: SendNotificationFormProps) {
  const { onNotificationUpdate, markNotificationAsSelfSent } =
    useNotificationContext();

  // State for user selection when no user is provided
  const [selectedUserUid, setSelectedUserUid] = useState<string>(
    user?.userUid || ""
  );
  const [selectedUser, setSelectedUser] = useState<User | null>(user || null);

  const [title, setTitle] = useState<string>("");
  const [body, setBody] = useState<string>("");
  const [type, setType] = useState<NotificationType>("info");
  const [priority, setPriority] = useState<NotificationPriority>("normal");
  const [link, setLink] = useState<string>("");
  const [expiresAt, setExpiresAt] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);

  // Fetch users for selection if no user is provided
  const { users, loading: usersLoading } = useUsers({
    limit: 100, // Reasonable limit for dropdown
  });

  const handleUserChange = (userUid: string) => {
    setSelectedUserUid(userUid);
    const foundUser = users.find((u) => u.userUid === userUid);
    setSelectedUser(foundUser || null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim()) {
      toast.error("Please enter a notification title");
      return;
    }

    if (!body.trim()) {
      toast.error("Please enter a notification message");
      return;
    }

    const targetUser = user || selectedUser;
    if (!targetUser) {
      toast.error("Please select a user to send the notification to");
      return;
    }

    setIsLoading(true);

    // Mark this notification as self-sent BEFORE sending to ensure proper timing
    markNotificationAsSelfSent(targetUser.userUid);

    try {
      const payload: any = {
        userUid: targetUser.userUid,
        title: title.trim(),
        body: body.trim(),
        type,
        priority,
      };

      // Add optional fields only if they have values
      if (link.trim()) {
        payload.link = link.trim();
      }

      if (expiresAt) {
        // Convert the datetime-local input to ISO string
        payload.expiresAt = new Date(expiresAt).toISOString();
      }

      const result = await apiFetch<SendNotificationApiResponse>(
        "/api/v1/notifications",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(payload),
        }
      );

      if (result.status === "success") {
        // Trigger notification context update to refresh user's notification counter
        onNotificationUpdate();

        // Show Facebook-style notification preview (what the user would see)
        showUserNotificationToast({
          id: Date.now(), // Temporary ID for preview
          title: title.trim(),
          body: body.trim(),
          type,
          priority,
          link: link.trim() || undefined,
          user: {
            fullName: "Admin Preview",
            email: "admin@slideinsight.net",
          },
        });

        // Also show a small confirmation toast
        toast.success("Notification sent! 📤", {
          description: `Preview shown - delivered to ${targetUser.email}`,
          duration: 2000,
        });

        // Small delay to allow WebSocket updates to propagate
        setTimeout(() => {
          onSuccess();
        }, 150);
      } else {
        throw new Error("Failed to send notification");
      }
    } catch (error) {
      console.error("Failed to send notification:", error);
      if (error instanceof ApiError) {
        toast.error(error.message);
      } else {
        toast.error("Failed to send notification");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const targetUser = user || selectedUser;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* User Selection (only shown if no user prop provided) */}
      {!user && (
        <div className="space-y-2">
          <Label htmlFor="user">Select Recipient *</Label>
          <Select
            value={selectedUserUid}
            onValueChange={handleUserChange}
            disabled={isLoading || usersLoading}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={
                  usersLoading ? "Loading users..." : "Choose a user to notify"
                }
              />
            </SelectTrigger>
            <SelectContent>
              {users.map((user) => (
                <SelectItem key={user.userUid} value={user.userUid}>
                  <div className="flex items-center space-x-2">
                    <UserIcon className="h-4 w-4" />
                    <div>
                      <div className="font-medium">{user.email}</div>
                      <div className="text-xs text-muted-foreground">
                        {user.email} • {user.firstName} {user.lastName}
                      </div>
                    </div>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {/* Recipient Display (when user is provided or selected) */}
      {targetUser && (
        <div className="space-y-2">
          <Label>Recipient</Label>
          <div className="p-2 bg-muted rounded">
            <div className="font-medium">{targetUser.email}</div>
            <div className="text-sm text-muted-foreground">
              {targetUser.email} • {targetUser.firstName} {targetUser.lastName}
            </div>
          </div>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="title">Title *</Label>
        <Input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Enter notification title"
          maxLength={200}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="body">Message *</Label>
        <Textarea
          id="body"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Enter notification message"
          rows={4}
          maxLength={1000}
        />
        <div className="text-xs text-muted-foreground">
          {body.length}/1000 characters
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="type">Type</Label>
          <Select
            value={type}
            onValueChange={(value: NotificationType) => setType(value)}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="info">Info</SelectItem>
              <SelectItem value="success">Success</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="error">Error</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="priority">Priority</Label>
          <Select
            value={priority}
            onValueChange={(value: NotificationPriority) => setPriority(value)}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="low">Low</SelectItem>
              <SelectItem value="normal">Normal</SelectItem>
              <SelectItem value="high">High</SelectItem>
              <SelectItem value="urgent">Urgent</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="link">Link (Optional)</Label>
        <Input
          id="link"
          type="url"
          value={link}
          onChange={(e) => setLink(e.target.value)}
          placeholder="https://example.com/link-to-relevant-page"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="expiresAt">Expires At (Optional)</Label>
        <Input
          id="expiresAt"
          type="datetime-local"
          value={expiresAt}
          onChange={(e) => setExpiresAt(e.target.value)}
          min={new Date().toISOString().slice(0, 16)} // Prevent past dates
        />
        <div className="text-xs text-muted-foreground">
          Leave empty for notifications that don't expire
        </div>
      </div>

      <div className="flex justify-end space-x-2 pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={
            isLoading ||
            !title.trim() ||
            !body.trim() ||
            (!user && !selectedUser)
          }
        >
          {isLoading ? "Sending..." : "Send Notification"}
        </Button>
      </div>
    </form>
  );
}
