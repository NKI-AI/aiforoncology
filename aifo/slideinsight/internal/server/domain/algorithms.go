// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

import "time"

// Algorithm represents an algorithm in the database
type Algorithm struct {
	ID                string    `json:"id"`
	TenantID          int       `json:"tenant_id"`
	TenantName        string    `json:"tenant_name,omitempty"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	Version           string    `json:"version"`
	EndpointURL       string    `json:"endpoint_url"`
	HTTPMethod        string    `json:"http_method"`
	ExecutionMode     string    `json:"execution_mode"`
	ProgressTransport string    `json:"progress_transport"`
	Metadata          string    `json:"metadata,omitempty"` // JSON string
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NewAlgorithm represents a new algorithm to be created
type NewAlgorithm struct {
	TenantID          int    `json:"tenant_id" validate:"required"`
	TenantUID         string `json:"tenant_uid,omitempty"` // Optional - for tenant conversion
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description,omitempty"`
	Version           string `json:"version" validate:"required"`
	EndpointURL       string `json:"endpoint_url" validate:"required,url"`
	HTTPMethod        string `json:"http_method,omitempty"` // defaults to POST
	ExecutionMode     string `json:"execution_mode" validate:"required,oneof=BATCH STREAM"`
	ProgressTransport string `json:"progress_transport,omitempty"` // defaults to WEBSOCKET
	Metadata          string `json:"metadata,omitempty"`
}

// AlgorithmUpdates represents fields that can be updated for an existing algorithm
type AlgorithmUpdates struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Version           *string `json:"version,omitempty"`
	EndpointURL       *string `json:"endpoint_url,omitempty"`
	HTTPMethod        *string `json:"http_method,omitempty"`
	ExecutionMode     *string `json:"execution_mode,omitempty"`
	ProgressTransport *string `json:"progress_transport,omitempty"`
	Metadata          *string `json:"metadata,omitempty"`
}

// ResultSink represents a result sink configuration
type ResultSink struct {
	ID          string    `json:"id"`
	AlgorithmID string    `json:"algorithm_id"`
	Type        string    `json:"type"`
	Config      string    `json:"config"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
}

// NewResultSink represents a new result sink to be created
type NewResultSink struct {
	Type   string `json:"type" validate:"required"`
	Config string `json:"config" validate:"required"` // JSON string
}

// PostHook represents a post-execution hook configuration
type PostHook struct {
	ID          string    `json:"id"`
	AlgorithmID string    `json:"algorithm_id"`
	Type        string    `json:"type"`
	Config      string    `json:"config"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
}

// NewPostHook represents a new post hook to be created
type NewPostHook struct {
	Type   string `json:"type" validate:"required"`
	Config string `json:"config" validate:"required"` // JSON string
}

// AlgorithmRun represents an algorithm execution run
type AlgorithmRun struct {
	ID            string     `json:"run_id"`
	AlgorithmID   string     `json:"algorithm_id"`
	CaseID        *string    `json:"case_id,omitempty"`
	ImageIDs      string     `json:"image_ids,omitempty"`  // JSON string array
	Regions       string     `json:"regions,omitempty"`    // JSON string
	Parameters    string     `json:"parameters,omitempty"` // JSON string
	ExecutionMode string     `json:"execution_mode"`
	Status        string     `json:"status"`
	Progress      int        `json:"progress"`
	ResultURI     *string    `json:"result_uri,omitempty"`
	ErrorInfo     *string    `json:"error_info,omitempty"` // JSON string
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewAlgorithmRun represents a new algorithm run to be created
type NewAlgorithmRun struct {
	AlgorithmID   string          `json:"algorithm_id" validate:"required"`
	CaseID        *string         `json:"case_id,omitempty"`
	ImageIDs      []string        `json:"image_ids,omitempty"`
	Regions       string          `json:"regions,omitempty"`    // JSON string
	Parameters    string          `json:"parameters,omitempty"` // JSON string
	ExecutionMode *string         `json:"execution_mode,omitempty"`
	ResultSinks   []NewResultSink `json:"result_sinks,omitempty"`
	ProgressSinks []NewResultSink `json:"progress_sinks,omitempty"`
	PostHooks     []NewPostHook   `json:"post_hooks,omitempty"`
}

// AlgorithmRunUpdates represents fields that can be updated for an existing run
type AlgorithmRunUpdates struct {
	Status     *string    `json:"status,omitempty"`
	Progress   *int       `json:"progress,omitempty"`
	ResultURI  *string    `json:"result_uri,omitempty"`
	ErrorInfo  *string    `json:"error_info,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// Output represents an algorithm run output
type Output struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	Name      string    `json:"name"`
	URI       string    `json:"uri"`
	Metadata  string    `json:"metadata,omitempty"` // JSON string
	CreatedAt time.Time `json:"created_at"`
}

// NewOutput represents a new output to be created
type NewOutput struct {
	RunID    string `json:"run_id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	URI      string `json:"uri" validate:"required"`
	Metadata string `json:"metadata,omitempty"`
}

// AlgorithmsResponse represents the response for listing algorithms
type AlgorithmsResponse struct {
	Algorithms []Algorithm    `json:"algorithms"`
	Pagination PaginationInfo `json:"pagination"`
}

// RunsResponse represents the response for listing runs
type RunsResponse struct {
	Runs       []AlgorithmRun `json:"runs"`
	Pagination PaginationInfo `json:"pagination"`
}

// RunStartResponse represents the response when starting a run
type RunStartResponse struct {
	RunID     string `json:"run_id"`
	Status    string `json:"status"`
	StatusURL string `json:"status_url"`
	EventsURL string `json:"events_url"`
}

// RunStatusResponse represents the response for run status
type RunStatusResponse struct {
	RunID      string     `json:"run_id"`
	Status     string     `json:"status"`
	Progress   int        `json:"progress"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	ResultURI  *string    `json:"result_uri,omitempty"`
	ErrorInfo  *string    `json:"error_info,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// OutputsResponse represents the response for listing outputs
type OutputsResponse struct {
	Outputs []Output `json:"outputs"`
}
