// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import CasesAdmin from "@/features/admin/components/cases/Cases";

export const Route = createFileRoute("/_authenticated/admin/cases/")({
  component: CasesAdmin,
});
