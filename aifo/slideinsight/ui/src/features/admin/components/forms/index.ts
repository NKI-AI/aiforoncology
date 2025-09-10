// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import { createForm } from "./FormFactory";
import { userFormConfig } from "./userFormConfig";
import { studyFormConfig } from "./studyFormConfig";
import { tenantFormConfig } from "./tenantFormConfig";
import { algorithmFormConfig } from "./algorithmFormConfig";

// Import standalone form components
export { PermissionForm } from "./PermissionForm";
// Generate form components from configurations
const UserFormGenerated = createForm(userFormConfig);
const StudyFormGenerated = createForm(studyFormConfig);
const TenantFormGenerated = createForm(tenantFormConfig);
const AlgorithmFormGenerated = createForm(algorithmFormConfig);

// Export as main form components (these replace the wrapper components)
export const UserForm = UserFormGenerated;
export const StudyForm = StudyFormGenerated;
export const TenantForm = TenantFormGenerated;
export const AlgorithmForm = AlgorithmFormGenerated;

// Export configs for potential reuse
