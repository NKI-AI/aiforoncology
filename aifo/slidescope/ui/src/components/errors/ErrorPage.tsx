// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import { useRouteError, isRouteErrorResponse } from "react-router-dom";
import { AlertIcon } from "../icons";
import FullPageMessage from "../layouts/FullPageMessage";
import { Button, LinkButton } from "../common/Button";

export default function ErrorPage() {
  const error = useRouteError();

  // Format the error message appropriately
  let errorMessage = "An unexpected error has occurred";
  if (isRouteErrorResponse(error)) {
    errorMessage = error.statusText || "Unknown error";
  } else if (error instanceof Error) {
    errorMessage = error.message;
  } else if (typeof error === "string") {
    errorMessage = error;
  }

  // For development only
  console.error("Error details:", error);

  const actions = (
    <>
      <Button onClick={() => window.location.reload()}>Try Again</Button>
      <LinkButton to="/" variant="secondary">
        Return to Home
      </LinkButton>
    </>
  );

  return (
    <FullPageMessage
      icon={<AlertIcon />}
      title="Something went wrong"
      message={errorMessage}
      actions={actions}
      iconBgColor="bg-red-100"
    />
  );
}
