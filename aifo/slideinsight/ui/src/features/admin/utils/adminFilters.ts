// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { type FilterField } from "../../../types/search";
import { type AdminEntityFilters } from "../hooks/useAdminEntityPage";

// =============================================================================
// Common Filter Interfaces
// =============================================================================

interface BaseFilters extends AdminEntityFilters {
  searchQuery: string;
}

export interface NameSearchFilters extends BaseFilters {
  searchName: string;
}

interface EmailSearchFilters extends BaseFilters {
  searchEmail: string;
}

interface DescriptionSearchFilters extends BaseFilters {
  searchDescription: string;
}

export interface StatusSearchFilters extends BaseFilters {
  searchStatus: string;
}

interface TypeSearchFilters extends BaseFilters {
  searchType: string;
}

// Combined common filter interfaces
export interface CommonUserFilters
  extends NameSearchFilters,
    EmailSearchFilters,
    StatusSearchFilters {}
export interface CommonEntityFilters
  extends NameSearchFilters,
    DescriptionSearchFilters,
    StatusSearchFilters {
  filterAccessibleStudies?: boolean;
}
export interface CommonTemplateFilters
  extends NameSearchFilters,
    TypeSearchFilters,
    StatusSearchFilters {}

// =============================================================================
// Initial Filter Objects
// =============================================================================

export const createInitialFilters = <T extends AdminEntityFilters>(
  additionalFields: Omit<T, "searchQuery"> = {} as Omit<T, "searchQuery">
): T =>
  ({
    searchQuery: "",
    ...additionalFields,
  } as T);

// Pre-defined initial filter objects for common patterns
const initialBaseFilters: BaseFilters = createInitialFilters();

export const initialNameSearchFilters: NameSearchFilters = createInitialFilters(
  {
    searchName: "",
  }
);

const initialEmailSearchFilters: EmailSearchFilters = createInitialFilters({
  searchEmail: "",
});

const initialDescriptionSearchFilters: DescriptionSearchFilters =
  createInitialFilters({
    searchDescription: "",
  });

const initialStatusSearchFilters: StatusSearchFilters = createInitialFilters({
  searchStatus: "",
});

export const initialCommonUserFilters: CommonUserFilters = createInitialFilters(
  {
    searchName: "",
    searchEmail: "",
    searchStatus: "",
  }
);

export const initialCommonEntityFilters: CommonEntityFilters =
  createInitialFilters({
    searchName: "",
    searchDescription: "",
    searchStatus: "",
    filterAccessibleStudies: false,
  });

export const initialCommonTemplateFilters: CommonTemplateFilters =
  createInitialFilters({
    searchName: "",
    searchType: "",
    searchStatus: "",
  });

// =============================================================================
// Status Option Constants
// =============================================================================

const STATUS_OPTIONS = {
  ACTIVE_INACTIVE: [
    { label: "Active", value: "active" },
    { label: "Inactive", value: "inactive" },
  ],
  PUBLISHED_DRAFT: [
    { label: "Published", value: "published" },
    { label: "Draft", value: "draft" },
  ],
  TENANT_STATUS: [
    { label: "Active", value: "active" },
    { label: "Inactive", value: "inactive" },
    { label: "Suspended", value: "suspended" },
  ],
} as const;

const EMAIL_TEMPLATE_TYPES = [
  { label: "Password Reset", value: "password_reset" },
  { label: "Email Verification", value: "email_verification" },
  { label: "Welcome", value: "welcome" },
] as const;

// =============================================================================
// Filter Field Builders
// =============================================================================

interface FilterFieldBuilderOptions<T extends AdminEntityFilters> {
  filters: T;
  updateFilter: (key: keyof T, value: string | boolean) => void;
}

const createTextFilterField = <T extends AdminEntityFilters>(
  key: keyof T,
  label: string,
  placeholder: string,
  options: FilterFieldBuilderOptions<T>
): FilterField => ({
  type: "text",
  key: key as string,
  label,
  placeholder,
  value: options.filters[key],
  onChange: (value: string | boolean) =>
    options.updateFilter(key, value as string),
});

const createSelectFilterField = <T extends AdminEntityFilters>(
  key: keyof T,
  label: string,
  placeholder: string,
  selectOptions: readonly { label: string; value: string }[],
  options: FilterFieldBuilderOptions<T>
): FilterField => ({
  type: "select",
  key: key as string,
  label,
  placeholder,
  value: options.filters[key],
  onChange: (value: string | boolean) =>
    options.updateFilter(key, value as string),
  options: [...selectOptions],
});

const createCheckboxFilterField = <T extends AdminEntityFilters>(
  key: keyof T,
  label: string,
  description: string,
  options: FilterFieldBuilderOptions<T>
): FilterField => ({
  type: "checkbox",
  key: key as string,
  label,
  description,
  value: options.filters[key],
  onChange: (value: string | boolean) =>
    options.updateFilter(key, value as boolean),
});

// =============================================================================
// Common Filter Field Builders
// =============================================================================

const createNameFilterField = <T extends NameSearchFilters>(
  options: FilterFieldBuilderOptions<T>,
  customLabel: string = "Name",
  customPlaceholder: string = "Filter by name..."
): FilterField =>
  createTextFilterField("searchName", customLabel, customPlaceholder, options);

const createEmailFilterField = <T extends EmailSearchFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField =>
  createTextFilterField("searchEmail", "Email", "Filter by email...", options);

const createDescriptionFilterField = <T extends DescriptionSearchFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField =>
  createTextFilterField(
    "searchDescription",
    "Description",
    "Filter by description...",
    options
  );

const createStatusFilterField = <T extends StatusSearchFilters>(
  options: FilterFieldBuilderOptions<T>,
  statusOptions: readonly {
    label: string;
    value: string;
  }[] = STATUS_OPTIONS.ACTIVE_INACTIVE
): FilterField =>
  createSelectFilterField(
    "searchStatus",
    "Status",
    "All statuses",
    statusOptions,
    options
  );

const createTypeFilterField = <T extends TypeSearchFilters>(
  options: FilterFieldBuilderOptions<T>,
  typeOptions: readonly { label: string; value: string }[],
  customLabel: string = "Type"
): FilterField =>
  createSelectFilterField(
    "searchType",
    customLabel,
    "All types",
    typeOptions,
    options
  );

// =============================================================================
// Pre-built Filter Field Collections
// =============================================================================

export const createUserFilterFields = <T extends CommonUserFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [
  createEmailFilterField(options),
  createStatusFilterField(options, STATUS_OPTIONS.ACTIVE_INACTIVE),
];

export const createStudyFilterFields = <T extends CommonEntityFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [
  createNameFilterField(options, "Study Name"),
  createDescriptionFilterField(options),
  createStatusFilterField(options, STATUS_OPTIONS.PUBLISHED_DRAFT),
  createCheckboxFilterField(
    "filterAccessibleStudies" as keyof T,
    "Filter studies",
    "Only show studies that you have access to",
    options
  ),
];

export const createTenantFilterFields = <
  T extends NameSearchFilters & StatusSearchFilters
>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [
  createNameFilterField(options, "Tenant Name"),
  createStatusFilterField(options, STATUS_OPTIONS.TENANT_STATUS),
];

export const createEmailTemplateFilterFields = <
  T extends CommonTemplateFilters
>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [
  createNameFilterField(options, "Template Name"),
  createTypeFilterField(options, EMAIL_TEMPLATE_TYPES, "Template Type"),
  createStatusFilterField(options, STATUS_OPTIONS.ACTIVE_INACTIVE),
];

export const createSlideFilterFields = <T extends NameSearchFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [createNameFilterField(options, "Slide Name")];

export const createCaseFilterFields = <T extends NameSearchFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [createNameFilterField(options, "Case Name")];

// =============================================================================
// Settings Filter Support
// =============================================================================

export interface CommonSettingsFilters extends BaseFilters {
  searchKey: string;
  searchValueType: string;
  searchTenantId: string;
}

export const initialCommonSettingsFilters: CommonSettingsFilters =
  createInitialFilters({
    searchKey: "",
    searchValueType: "",
    searchTenantId: "",
  });

const SETTING_VALUE_TYPE_OPTIONS = [
  { label: "Boolean", value: "boolean" },
  { label: "Number", value: "number" },
  { label: "String", value: "string" },
  { label: "JSON", value: "json" },
] as const;

const createKeyFilterField = <
  T extends AdminEntityFilters & { searchKey: string }
>(
  options: FilterFieldBuilderOptions<T>
): FilterField =>
  createTextFilterField(
    "searchKey",
    "Key",
    "Filter by setting key...",
    options
  );

const createValueTypeFilterField = <
  T extends AdminEntityFilters & { searchValueType: string }
>(
  options: FilterFieldBuilderOptions<T>
): FilterField =>
  createSelectFilterField(
    "searchValueType",
    "Value Type",
    "All types",
    SETTING_VALUE_TYPE_OPTIONS,
    options
  );

const createTenantIdFilterField = <
  T extends AdminEntityFilters & { searchTenantId: string }
>(
  options: FilterFieldBuilderOptions<T>
): FilterField =>
  createTextFilterField(
    "searchTenantId",
    "Tenant ID",
    "Filter by tenant ID...",
    options
  );

export const createSettingsFilterFields = <T extends CommonSettingsFilters>(
  options: FilterFieldBuilderOptions<T>
): FilterField[] => [
  createKeyFilterField(options),
  createValueTypeFilterField(options),
  createTenantIdFilterField(options),
];
