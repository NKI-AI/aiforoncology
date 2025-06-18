// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  refresh_expires_in?: number;
}

export interface UserData {
  username: string;
  email?: string;
  roles?: string[];
  exp?: number;
  refresh_exp?: number;
  [key: string]: any; // For any additional user data fields
}

// Helper function to format expiry duration
export const formatAuthExpiry = (expiresInSeconds: number): number => {
  // Convert seconds to minutes for react-auth-kit
  return Math.floor(expiresInSeconds / 60);
};
