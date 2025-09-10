// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { createFileRoute } from "@tanstack/react-router";
import { ActivateAccountPage } from "../../features/account";

// Define the search params interface
export interface ActivateSearchParams {
  token?: string;
  activation_code?: string;
}

export const Route = createFileRoute("/account/activate")({
  component: ActivateAccountPage,
  validateSearch: (search: Record<string, unknown>): ActivateSearchParams => ({
    token: typeof search.token === "string" ? search.token : undefined,
    activation_code:
      typeof search.activation_code === "string"
        ? search.activation_code
        : undefined,
  }),
});
