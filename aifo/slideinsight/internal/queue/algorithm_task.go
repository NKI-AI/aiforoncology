// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package queue

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// AlgorithmTaskType represents the type identifier for algorithm tasks
const AlgorithmTaskType = "algorithm"

// AlgorithmFunction represents a function that can be executed by an algorithm task
type AlgorithmFunction func(ctx context.Context, parameters map[string]interface{}) error

// AlgorithmTask implements the Task interface for executing algorithms
type AlgorithmTask struct {
	id         string
	name       string
	function   AlgorithmFunction
	parameters map[string]interface{}
	maxRetries int
	retryDelay time.Duration
	timeout    time.Duration
}

// generateAlgorithmUUID creates a UUID v4 using crypto/rand
func generateAlgorithmUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)

	// Set version (4) and variant bits according to RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// NewAlgorithmTask creates a new algorithm task
func NewAlgorithmTask(name string, function AlgorithmFunction, parameters map[string]interface{}) *AlgorithmTask {
	return &AlgorithmTask{
		id:         generateAlgorithmUUID(),
		name:       name,
		function:   function,
		parameters: parameters,
		maxRetries: 2,                // Default to 2 retries
		retryDelay: 60 * time.Second, // Default to 60 second delay between retries
		timeout:    10 * time.Minute, // Default to 10 minute timeout
	}
}

// NewAlgorithmTaskWithConfig creates a new algorithm task with custom configuration
func NewAlgorithmTaskWithConfig(name string, function AlgorithmFunction, parameters map[string]interface{}, maxRetries int, retryDelay, timeout time.Duration) *AlgorithmTask {
	return &AlgorithmTask{
		id:         generateAlgorithmUUID(),
		name:       name,
		function:   function,
		parameters: parameters,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		timeout:    timeout,
	}
}

// ID returns the unique identifier for this task
func (a *AlgorithmTask) ID() string {
	return a.id
}

// Type returns the task type
func (a *AlgorithmTask) Type() string {
	return AlgorithmTaskType
}

// Execute runs the algorithm function with the provided parameters
func (a *AlgorithmTask) Execute(ctx context.Context) error {
	// Create a context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	return a.function(taskCtx, a.parameters)
}

// MaxRetries returns the maximum number of retry attempts
func (a *AlgorithmTask) MaxRetries() int {
	return a.maxRetries
}

// RetryDelay returns the delay between retry attempts
func (a *AlgorithmTask) RetryDelay() time.Duration {
	return a.retryDelay
}

// String returns a human-readable description of the task
func (a *AlgorithmTask) String() string {
	return fmt.Sprintf("AlgorithmTask(id=%s, name=%s, parameters=%v)",
		a.id, a.name, a.parameters)
}

// Name returns the algorithm name
func (a *AlgorithmTask) Name() string {
	return a.name
}

// Parameters returns the algorithm parameters
func (a *AlgorithmTask) Parameters() map[string]interface{} {
	return a.parameters
}

// Timeout returns the algorithm execution timeout
func (a *AlgorithmTask) Timeout() time.Duration {
	return a.timeout
}

// AlgorithmTaskBuilder provides a fluent interface for building algorithm tasks
type AlgorithmTaskBuilder struct {
	name       string
	function   AlgorithmFunction
	parameters map[string]interface{}
	maxRetries int
	retryDelay time.Duration
	timeout    time.Duration
}

// NewAlgorithmTaskBuilder creates a new algorithm task builder
func NewAlgorithmTaskBuilder(name string, function AlgorithmFunction) *AlgorithmTaskBuilder {
	return &AlgorithmTaskBuilder{
		name:       name,
		function:   function,
		parameters: make(map[string]interface{}),
		maxRetries: 2,
		retryDelay: 60 * time.Second,
		timeout:    10 * time.Minute,
	}
}

// WithParameter adds a parameter to the algorithm task
func (b *AlgorithmTaskBuilder) WithParameter(key string, value interface{}) *AlgorithmTaskBuilder {
	b.parameters[key] = value
	return b
}

// WithParameters sets all parameters for the algorithm task
func (b *AlgorithmTaskBuilder) WithParameters(parameters map[string]interface{}) *AlgorithmTaskBuilder {
	b.parameters = parameters
	return b
}

// WithMaxRetries sets the maximum number of retry attempts
func (b *AlgorithmTaskBuilder) WithMaxRetries(maxRetries int) *AlgorithmTaskBuilder {
	b.maxRetries = maxRetries
	return b
}

// WithRetryDelay sets the delay between retry attempts
func (b *AlgorithmTaskBuilder) WithRetryDelay(delay time.Duration) *AlgorithmTaskBuilder {
	b.retryDelay = delay
	return b
}

// WithTimeout sets the execution timeout
func (b *AlgorithmTaskBuilder) WithTimeout(timeout time.Duration) *AlgorithmTaskBuilder {
	b.timeout = timeout
	return b
}

// Build creates the algorithm task
func (b *AlgorithmTaskBuilder) Build() *AlgorithmTask {
	return NewAlgorithmTaskWithConfig(b.name, b.function, b.parameters, b.maxRetries, b.retryDelay, b.timeout)
}

// Example algorithm functions for demonstration

// ExampleImageProcessingAlgorithm is an example image processing algorithm
func ExampleImageProcessingAlgorithm(ctx context.Context, parameters map[string]interface{}) error {
	inputPath, ok := parameters["input_path"].(string)
	if !ok {
		return fmt.Errorf("input_path parameter is required and must be a string")
	}

	outputPath, ok := parameters["output_path"].(string)
	if !ok {
		return fmt.Errorf("output_path parameter is required and must be a string")
	}

	// Simulate image processing work
	select {
	case <-time.After(2 * time.Second):
		// Processing completed
		fmt.Printf("Processed image from %s to %s\n", inputPath, outputPath)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ExampleDataAnalysisAlgorithm is an example data analysis algorithm
func ExampleDataAnalysisAlgorithm(ctx context.Context, parameters map[string]interface{}) error {
	datasetID, ok := parameters["dataset_id"].(string)
	if !ok {
		return fmt.Errorf("dataset_id parameter is required and must be a string")
	}

	analysisType, ok := parameters["analysis_type"].(string)
	if !ok {
		return fmt.Errorf("analysis_type parameter is required and must be a string")
	}

	// Simulate data analysis work
	select {
	case <-time.After(5 * time.Second):
		// Analysis completed
		fmt.Printf("Completed %s analysis on dataset %s\n", analysisType, datasetID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
