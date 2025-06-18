// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import { apiFetch } from "../utils/fetchUtils";
import { UserData, LoginResponse } from "./authProvider";
import { jwtDecode } from "jwt-decode";
import WarningModal from "../components/common/WarningModal";

// Token duration constants (in milliseconds)
const TOKEN_REFRESH_INTERVAL = 5 * 1000; // TEMPORARY: Check every 5 seconds for testing
const INACTIVITY_THRESHOLD = 5 * 60 * 1000; // 5 minutes of inactivity
const ACCESS_TOKEN_EXPIRY_THRESHOLD = 60 * 1000; // Refresh when token is 1 minute from expiry
const REFRESH_TOKEN_WARNING_THRESHOLD = 10 * 1000; // TEMPORARY: Show warning after 10 seconds for testing

// TESTING ONLY: Remove this for production
const TESTING_MODE = false; // Set to false to disable test mode

interface AuthContextType {
  isLoggedIn: boolean;
  user: UserData | null;
  setAuth: (user: UserData | null, isLoggedIn: boolean) => void;
  logout: () => Promise<void>;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);
  const [user, setUser] = useState<UserData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [lastActivity, setLastActivity] = useState<number>(Date.now());
  const [showExpiryWarning, setShowExpiryWarning] = useState<boolean>(false);

  const setAuth = (userData: UserData | null, loggedIn: boolean) => {
    if (userData) {
      // Add login time for testing purposes
      userData.login_time = Date.now();
    }
    setUser(userData);
    setIsLoggedIn(loggedIn);
  };

  const logout = async () => {
    try {
      // Call logout endpoint
      await apiFetch("/api/v1/auth/logout", { method: "POST" });
    } catch (error) {
      console.warn("Error logging out:", error);
    } finally {
      // Clear auth state regardless of API success
      setAuth(null, false);
    }
  };

  // Check authentication status on initial load
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Fetch current user from /api/v1/auth/me endpoint
        const userData = await apiFetch<UserData>("/api/v1/auth/me");
        setAuth(userData, true);
      } catch (error) {
        // If API call fails, user is not authenticated
        setAuth(null, false);
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  // Track user activity
  useEffect(() => {
    const update = () => setLastActivity(Date.now());
    window.addEventListener("mousemove", update);
    window.addEventListener("keydown", update);
    return () => {
      window.removeEventListener("mousemove", update);
      window.removeEventListener("keydown", update);
    };
  }, []);

  const refreshSession = async () => {
    try {
      const resp = await apiFetch<LoginResponse>("/api/v1/auth/refresh", {
        method: "POST",
      });
      const decodedAccess = jwtDecode<{
        sub: string;
        roles?: string[];
        email?: string;
        exp?: number;
      }>(resp.access_token);
      let refreshExp: number | undefined = undefined;
      if (resp.refresh_token) {
        try {
          const decRef = jwtDecode<{ exp?: number }>(resp.refresh_token);
          refreshExp = decRef.exp;
        } catch {}
      }

      setAuth(
        {
          username: decodedAccess.sub,
          roles: decodedAccess.roles || [],
          email: decodedAccess.email,
          exp: decodedAccess.exp,
          refresh_exp: refreshExp,
        },
        true
      );
      setShowExpiryWarning(false);
    } catch (err) {
      console.warn("Token refresh failed", err);
      setAuth(null, false);
    }
  };

  // Automatic token refresh
  useEffect(() => {
    const interval = setInterval(async () => {
      if (!isLoggedIn || !user || !user.exp) return;
      const now = Date.now();

      // TESTING MODE: Force show the warning modal for testing
      if (TESTING_MODE) {
        // For testing purposes, show the modal after 10 seconds of logging in
        const loginTime = user.login_time || Date.now();
        if (
          now - loginTime > REFRESH_TOKEN_WARNING_THRESHOLD &&
          !showExpiryWarning
        ) {
          console.log("TESTING: Showing session expiry warning modal");
          setShowExpiryWarning(true);
          return;
        }
      }

      if (
        user.refresh_exp &&
        user.refresh_exp * 1000 - now < REFRESH_TOKEN_WARNING_THRESHOLD
      ) {
        setShowExpiryWarning(true);
      }

      if (
        user.exp * 1000 - now < ACCESS_TOKEN_EXPIRY_THRESHOLD &&
        now - lastActivity < INACTIVITY_THRESHOLD
      ) {
        await refreshSession();
      }
    }, TOKEN_REFRESH_INTERVAL);
    return () => clearInterval(interval);
  }, [isLoggedIn, user, lastActivity, showExpiryWarning]);

  return (
    <AuthContext.Provider
      value={{ isLoggedIn, user, setAuth, logout, loading }}
    >
      <WarningModal
        isOpen={showExpiryWarning}
        onClose={() => {
          setShowExpiryWarning(false);
          logout();
        }}
        onRefresh={refreshSession}
        title="Session Expiring Soon"
        message="Your session is about to expire due to inactivity. Would you like to continue your session?"
      />
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
