// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createFileRoute } from "@tanstack/react-router";
import GroupsAdminComponent from "@/features/admin/components/groups/Groups";

export const Route = createFileRoute("/_authenticated/admin/groups/")({
  component: GroupsAdminComponent,
});
