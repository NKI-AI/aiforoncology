// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React from "react";
import { toast } from "sonner";
import {
  NotificationIcon,
  ErrorIcon,
  WarningIcon,
  CheckIcon,
  AlertIcon,
} from "@/components/icons";

interface NotificationToastData {
  id: number;
  title: string;
  body: string;
  link?: string;
  type: "info" | "success" | "warning" | "error";
  priority: "low" | "normal" | "high" | "urgent";
  user?: {
    fullName: string;
    email: string;
  };
}

const NotificationToastContent: React.FC<{
  notification: NotificationToastData;
}> = ({ notification }) => {
  const getIcon = () => {
    switch (notification.type) {
      case "error":
        return <ErrorIcon className="h-5 w-5 text-red-500" />;
      case "warning":
        return <WarningIcon className="h-5 w-5 text-yellow-500" />;
      case "success":
        return <CheckIcon className="h-5 w-5 text-green-500" />;
      default:
        return <AlertIcon className="h-5 w-5 text-blue-500" />;
    }
  };

  const getPriorityBorder = () => {
    switch (notification.priority) {
      case "urgent":
        return "border-l-4 border-red-500";
      case "high":
        return "border-l-4 border-orange-500";
      case "normal":
        return "border-l-4 border-blue-500";
      default:
        return "border-l-4 border-gray-400";
    }
  };

  const handleClick = () => {
    if (notification.link) {
      window.open(notification.link, "_blank");
    }
  };

  return (
    <div
      className={`flex items-start space-x-3 p-1 ${getPriorityBorder()} ${
        notification.link ? "cursor-pointer hover:bg-gray-50" : ""
      }`}
      onClick={handleClick}
    >
      <div className="flex-shrink-0 mt-0.5">{getIcon()}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-muted-900 truncate">
            {notification.title}
          </p>
          {notification.priority === "urgent" && (
            <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
              Urgent
            </span>
          )}
        </div>
        <p className="text-sm text-muted-600 line-clamp-2">
          {notification.body}
        </p>
        {notification.user && (
          <p className="text-xs text-muted-500 mt-1">
            From: {notification.user.fullName} ({notification.user.email})
          </p>
        )}
        {notification.link && (
          <p className="text-xs text-blue-600 mt-1 hover:underline">
            Click to view →
          </p>
        )}
      </div>
    </div>
  );
};

// Enhanced Facebook-style notification with better visual prominence
export const showUserNotificationToast = (
  notification: NotificationToastData,
  userAvatar?: string
) => {
  // Play notification sound (optional)
  const playNotificationSound = () => {
    try {
      // Create a subtle notification sound using Web Audio API
      const audioContext = new (window.AudioContext ||
        (window as any).webkitAudioContext)();
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();

      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);

      oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
      oscillator.frequency.setValueAtTime(600, audioContext.currentTime + 0.1);

      gainNode.gain.setValueAtTime(0, audioContext.currentTime);
      gainNode.gain.linearRampToValueAtTime(
        0.1,
        audioContext.currentTime + 0.01
      );
      gainNode.gain.exponentialRampToValueAtTime(
        0.01,
        audioContext.currentTime + 0.2
      );

      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.2);
    } catch (err) {
      // Silently fail if audio context isn't available
    }
  };

  const getNotificationIcon = () => {
    switch (notification.type) {
      case "error":
        return <ErrorIcon className="h-4 w-4 text-red-500" />;
      case "warning":
        return <WarningIcon className="h-4 w-4 text-yellow-500" />;
      case "success":
        return <CheckIcon className="h-4 w-4 text-green-500" />;
      default:
        return <NotificationIcon className="h-4 w-4 text-blue-500" />;
    }
  };

  const getPriorityIndicator = () => {
    if (notification.priority === "urgent") {
      return (
        <div className="absolute -top-1 -right-1 h-3 w-3 bg-red-500 rounded-full animate-pulse"></div>
      );
    }
    if (notification.priority === "high") {
      return (
        <div className="absolute -top-1 -right-1 h-3 w-3 bg-orange-500 rounded-full"></div>
      );
    }
    return null;
  };

  // Play sound notification
  playNotificationSound();

  const duration =
    notification.priority === "urgent"
      ? 12000
      : notification.priority === "high"
      ? 8000
      : 6000;

  return toast.custom(
    (id) => (
      <div
        className="relative flex items-start space-x-3 p-4 min-w-[360px] max-w-[420px] cursor-pointer hover:bg-gray-50 transition-colors rounded-2xl"
        onClick={() => {
          if (notification.link) {
            window.open(notification.link, "_blank");
          }
        }}
      >
        {/* Priority indicator */}
        {getPriorityIndicator()}

        {/* User Avatar */}
        <div className="relative flex-shrink-0">
          {userAvatar ? (
            <img
              src={userAvatar}
              alt={notification.user?.email || "User"}
              className="h-10 w-10 rounded-full border-2 border-blue-100"
            />
          ) : (
            <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center border-2 border-blue-100 shadow-sm">
              <span className="text-white text-sm font-bold">
                {notification.user?.fullName?.[0]?.toUpperCase() || "S"}
              </span>
            </div>
          )}

          {/* Notification type icon overlay */}
          <div className="absolute -bottom-1 -right-1 h-5 w-5 bg-background rounded-full flex items-center justify-center border border-gray-200 shadow-sm">
            {getNotificationIcon()}
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header with sender info */}
          <div className="flex items-center space-x-2 mb-1">
            <p className="text-sm font-semibold text-muted-900 truncate">
              {notification.user?.fullName || "System"}
            </p>
            <span className="text-xs text-muted-500 font-medium">
              sent you a notification
            </span>
          </div>

          {/* Notification title */}
          <p className="text-sm text-muted-800 font-medium mb-1 leading-tight">
            {notification.title}
          </p>

          {/* Notification body */}
          <p className="text-sm text-muted-600 line-clamp-2 leading-relaxed">
            {notification.body}
          </p>

          {/* Footer with timestamp and action */}
          <div className="flex items-center justify-between mt-3">
            <span className="text-xs text-muted-500 font-medium">Just now</span>
            <div className="flex items-center space-x-2">
              {notification.link && (
                <button
                  className="text-xs text-blue-600 hover:text-blue-700 font-medium hover:underline transition-colors"
                  onClick={(e) => {
                    e.stopPropagation();
                    window.open(notification.link, "_blank");
                  }}
                >
                  View →
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    ),
    {
      duration,
      position: "top-right",
      style: {
        backgroundColor: "#ffffff",
        border: "1px solid #e5e7eb",
        borderRadius: "16px",
        boxShadow:
          "0 25px 50px -12px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(0, 0, 0, 0.05)",
        padding: "0",
        minWidth: "360px",
        maxWidth: "420px",
        marginTop: "80px", // Position below header
        marginRight: "20px",
        // Enhanced backdrop blur effect
        backdropFilter: "blur(8px)",
        WebkitBackdropFilter: "blur(8px)",
        // Subtle gradient border
        background: "linear-gradient(145deg, #ffffff 0%, #fafafa 100%)",
      },
      className: "facebook-notification-toast",
      onDismiss: () => {},
    }
  );
};
