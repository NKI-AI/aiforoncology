// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface FilterField {
  type: "text" | "select" | "checkbox";
  key: string;
  label: string;
  placeholder?: string;
  description?: string;
  value: any;
  onChange: (value: string | boolean) => void;
  options?: { label: string; value: string }[];
}
