// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useState } from "react";
import { User } from "../../hooks/useAdminData";
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
import { toast } from "sonner";
import {
  useApiMutation,
  createApiMutation,
} from "../../../../utils/apiQueries";
import { ApiError } from "../../../../utils/fetchUtils";

interface SendEmailFormProps {
  user: User;
  onSuccess: () => void;
  onCancel: () => void;
}

interface SendEmailRequest {
  template: string;
  subject?: string;
}

export function SendEmailForm({
  user,
  onSuccess,
  onCancel,
}: SendEmailFormProps) {
  const [template, setTemplate] = useState<string>("");
  const [subject, setSubject] = useState<string>("");

  const sendEmailMutation = useApiMutation(
    createApiMutation.post<void, SendEmailRequest>(
      `/api/v1/users/${user.userUid}/send-email`
    ),
    {
      onSuccess: () => {
        toast.success(`Email sent successfully to ${user.email}`);
        onSuccess();
      },
      onError: (error) => {
        console.error("Failed to send email:", error);
        if (error instanceof ApiError) {
          toast.error(error.message);
        } else {
          toast.error("Failed to send email");
        }
      },
    }
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!template) {
      toast.error("Please select an email template");
      return;
    }

    sendEmailMutation.mutate({
      template,
      subject: subject || undefined,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Recipient</Label>
        <div className="p-2 bg-muted rounded">
          <div className="font-medium">
            {user.firstName} {user.lastName}
          </div>
          <div className="text-sm text-muted-foreground">{user.email}</div>
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="template">Email Template *</Label>
        <Select value={template} onValueChange={setTemplate}>
          <SelectTrigger>
            <SelectValue placeholder="Select an email template" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="welcome">Welcome Email</SelectItem>
            <SelectItem value="password_reset">Password Reset</SelectItem>
            <SelectItem value="email_verification">
              Email Verification
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="subject">Custom Subject (Optional)</Label>
        <Input
          id="subject"
          type="text"
          value={subject}
          onChange={(e) => setSubject(e.target.value)}
          placeholder="Leave empty to use default subject"
        />
      </div>

      <div className="flex justify-end space-x-2 pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={sendEmailMutation.isPending}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={sendEmailMutation.isPending || !template}
        >
          {sendEmailMutation.isPending ? "Sending..." : "Send Email"}
        </Button>
      </div>
    </form>
  );
}
