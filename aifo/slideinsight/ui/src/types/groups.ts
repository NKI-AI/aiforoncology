// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface Group {
  id: number;
  short_uid: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGroupRequest {
  name: string;
  description: string;
}

interface GroupsResponse {
  groups?: Group[];
}

interface CreateGroupResponse {
  group: Group;
}

export interface UserGroupAssignment {
  user_uids: string[];
}
