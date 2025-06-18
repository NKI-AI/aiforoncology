// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
import React from "react";
import { WarningIcon } from "../icons";
import FullPageMessage from "../layouts/FullPageMessage";
import { LinkButton } from "../common/Button";

export default function NotFoundPage() {
  const actions = (
    <>
      <LinkButton to="/">Go to Home</LinkButton>
      <LinkButton to="/slides" variant="secondary">
        View Slides
      </LinkButton>
    </>
  );

  return (
    <FullPageMessage
      icon={<WarningIcon />}
      title="Page Not Found"
      message="The page you're looking for doesn't exist or has been moved."
      actions={actions}
    />
  );
}
