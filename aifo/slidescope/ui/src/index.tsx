// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React, { Suspense } from "react";
import ReactDOM from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import App from "./App";
import "./styles.css";
import { ProtectedRoute, AuthProvider } from "./auth";

// Lazy load heavy components
// const SwaggerPage = React.lazy(() => import("./components/SwaggerPage"));
const LoginPage = React.lazy(() => import("./components/LoginPage"));
const ResetPasswordPage = React.lazy(
  () => import("./components/ResetPasswordPage")
);
const AccountPage = React.lazy(() => import("./components/AccountPage"));
const SlideList = React.lazy(() => import("./components/SlideList"));
const NotFoundPage = React.lazy(
  () => import("./components/errors/NotFoundPage")
);
const ErrorPage = React.lazy(() => import("./components/errors/ErrorPage"));
const RootRedirect = React.lazy(() => import("./components/RootRedirect"));

// Generic loading fallback component
const LoadingFallback = ({ message }: { message: string }) => (
  <div className="flex items-center justify-center min-h-screen bg-gray-100">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
      <div className="text-lg text-gray-700">{message}</div>
    </div>
  </div>
);

// Create router configuration
const router = createBrowserRouter([
  {
    path: "/v/:slideId?",
    element: (
      <ProtectedRoute>
        <App />
      </ProtectedRoute>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  {
    path: "/slides",
    element: (
      <ProtectedRoute>
        <Suspense fallback={<LoadingFallback message="Loading slides..." />}>
          <SlideList />
        </Suspense>
      </ProtectedRoute>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  {
    path: "/",
    element: (
      <Suspense fallback={<LoadingFallback message="Loading..." />}>
        <RootRedirect />
      </Suspense>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  {
    path: "/login",
    element: (
      <Suspense fallback={<LoadingFallback message="Loading login..." />}>
        <LoginPage />
      </Suspense>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  {
    path: "/account",
    element: (
      <ProtectedRoute>
        <Suspense fallback={<LoadingFallback message="Loading account..." />}>
          <AccountPage />
        </Suspense>
      </ProtectedRoute>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  {
    path: "/account/reset_password",
    element: (
      <Suspense
        fallback={<LoadingFallback message="Loading password reset..." />}
      >
        <ResetPasswordPage />
      </Suspense>
    ),
    errorElement: (
      <Suspense fallback={<LoadingFallback message="Loading error page..." />}>
        <ErrorPage />
      </Suspense>
    ),
  },
  // {
  //   path: "/swagger",
  //   element: (
  //     <Suspense fallback={<LoadingFallback message="Loading API documentation..." />}>
  //       <SwaggerPage />
  //     </Suspense>
  //   ),
  // },
  // Catch-all route for 404s
  {
    path: "*",
    element: (
      <Suspense fallback={<LoadingFallback message="Loading..." />}>
        <NotFoundPage />
      </Suspense>
    ),
  },
]);

console.log("Initializing SlideScope app with React Router 🔍");

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  </React.StrictMode>
);
