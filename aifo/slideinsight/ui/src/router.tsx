// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createRouter, ErrorComponent } from "@tanstack/react-router";
import { QueryClient } from "@tanstack/react-query";
import { routeTree } from "./routeTree.gen";
import { useAuth } from "./auth";

// Create a query client instance
const queryClient = new QueryClient();

// Create the router instance
export const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
  // defaultErrorComponent: ErrorBoundary,
  context: {
    auth: undefined!,
    queryClient,
  },
});

// Register router for type safety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

// Router Provider Hook
export function useRouterContext() {
  const auth = useAuth();
  return {
    auth: {
      isLoggedIn: auth.isLoggedIn,
      loading: auth.isLoading,
      user: auth.user,
    },
    queryClient,
  };
}
