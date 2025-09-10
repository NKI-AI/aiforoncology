// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// AlgorithmsService defines the interface for algorithm-related operations
type AlgorithmsService interface {
	// Algorithm operations
	CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error)
	GetAlgorithms(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Algorithm, domain.PaginationInfo, error)
	GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error)
	UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error
	DeleteAlgorithm(ctx context.Context, algorithmID string) error
	GetAlgorithmsCount(ctx context.Context) (int, error)

	// Algorithm run operations
	CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.RunStartResponse, error)
	GetAlgorithmRuns(ctx context.Context, algorithmID string, params utils.PaginationAndSearchParams) ([]domain.AlgorithmRun, domain.PaginationInfo, error)
	GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error)
	UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error
	CancelAlgorithmRun(ctx context.Context, runID string) error
	GetRunsCount(ctx context.Context, algorithmID string) (int, error)

	// Output operations
	CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error)
	GetOutputs(ctx context.Context, runID string) ([]domain.Output, error)
	GetOutput(ctx context.Context, outputID string) (domain.Output, error)
	DeleteOutput(ctx context.Context, outputID string) error

	Close()
}

type algorithmsService struct {
	*BaseService
	db             ports.Database
	tenantsService TenantsService // Add tenants service for UID conversion
	// Generic pagination and search services
	algorithmsPaginatedService *PaginatedSearchService[domain.Algorithm, domain.Algorithm]
	runsPaginatedService       *PaginatedSearchService[domain.AlgorithmRun, domain.AlgorithmRun]
}

// algorithmConversionHelpers provides conversion helpers configured for algorithms
var algorithmConversionHelpers = DefaultConversionHelpers()

// convertAlgorithmBase handles basic algorithm conversion (no conversion needed for domain to domain)
func convertAlgorithmBase(record domain.Algorithm, helpers *ConversionHelpers) domain.Algorithm {
	return record
}

// convertAlgorithmRunBase handles basic run conversion (no conversion needed for domain to domain)
func convertAlgorithmRunBase(record domain.AlgorithmRun, helpers *ConversionHelpers) domain.AlgorithmRun {
	return record
}

// NewAlgorithmsService creates a new AlgorithmsService
func NewAlgorithmsService(db ports.Database, tenantsService TenantsService) AlgorithmsService {
	baseService := NewBaseService(db)

	service := &algorithmsService{
		BaseService:    baseService,
		db:             db,
		tenantsService: tenantsService,
	}

	// We'll set up the paginated services later since they depend on the service instance
	return service
}

// CreateAlgorithm creates a new algorithm
func (s *algorithmsService) CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error) {
	// Get authentication context to determine tenant ID
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInternal, "failed to get auth context: %v", err)
	}

	// Handle tenant ID determination
	if newAlgorithm.TenantUID != "" {
		// If tenant UID is provided, convert it to tenant ID
		// For superadmins, allow creating algorithms for other tenants
		// For regular users, verify they can only create for their own tenant

		// TODO: Add tenant name to the domain.Tenant struct to get the ID
		// For now, we'll use a workaround to get the internal ID
		// Since domain.Tenant doesn't have ID field, we need to call the database directly
		dbTenant, err := s.db.GetTenantByUID(ctx, newAlgorithm.TenantUID)
		if err != nil {
			return domain.Algorithm{}, errors.WithDetails(errors.ErrInvalidInput, "failed to get tenant ID: %v", err)
		}

		// For regular users, ensure they can only create algorithms for their own tenant
		if !authCtx.IsSuperAdmin && dbTenant.ID != authCtx.TenantID {
			return domain.Algorithm{}, errors.WithDetails(errors.ErrInvalidInput, "cannot create algorithm for different tenant")
		}

		newAlgorithm.TenantID = dbTenant.ID
	} else {
		// Use the authenticated user's tenant ID
		newAlgorithm.TenantID = authCtx.TenantID
	}

	// Validate input
	if newAlgorithm.Name == "" {
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInvalidInput, "algorithm name cannot be empty")
	}
	if newAlgorithm.TenantID == 0 {
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInvalidInput, "tenant ID cannot be empty")
	}

	algorithm, err := s.db.CreateAlgorithm(ctx, newAlgorithm)
	if err != nil {
		log.Error("Failed to create algorithm", "error", err, "name", newAlgorithm.Name)
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInternal, "failed to create algorithm: %v", err)
	}

	log.Info("Algorithm created", "algorithmID", algorithm.ID, "name", algorithm.Name)
	return algorithm, nil
}

// GetAlgorithms retrieves algorithms with search and pagination
func (s *algorithmsService) GetAlgorithms(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Algorithm, domain.PaginationInfo, error) {
	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get auth context: %v", err)
	}

	var algorithms []domain.Algorithm
	var totalCount int

	if authCtx.IsSuperAdmin {
		// Superadmin can see all algorithms across all tenants
		algorithms, err = s.db.GetAllAlgorithms(ctx, params.SearchParams, params.PaginationParams)
		if err != nil {
			return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get all algorithms: %v", err)
		}

		// Get count for pagination
		if params.SearchParams.Query != "" || len(params.SearchParams.Filters) > 0 {
			totalCount, err = s.db.GetAllAlgorithmsCountWithSearch(ctx, params.SearchParams)
		} else {
			totalCount, err = s.db.GetAllAlgorithmsCount(ctx)
		}
	} else {
		// Regular users see only algorithms for their tenant
		algorithms, err = s.db.GetAlgorithms(ctx, authCtx.TenantID, params.SearchParams, params.PaginationParams)
		if err != nil {
			return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get algorithms: %v", err)
		}

		// Get count for pagination
		if params.SearchParams.Query != "" || len(params.SearchParams.Filters) > 0 {
			totalCount, err = s.db.GetAlgorithmsCountWithSearch(ctx, authCtx.TenantID, params.SearchParams)
		} else {
			totalCount, err = s.db.GetAlgorithmsCount(ctx, authCtx.TenantID)
		}
	}

	if err != nil {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get algorithms count: %v", err)
	}

	// Ensure we return an empty slice instead of nil to avoid null in JSON
	if algorithms == nil {
		algorithms = []domain.Algorithm{}
	}

	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return algorithms, paginationInfo, nil
}

// GetAlgorithm retrieves a single algorithm by ID
func (s *algorithmsService) GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error) {
	if algorithmID == "" {
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	algorithm, err := s.db.GetAlgorithm(ctx, algorithmID)
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.Algorithm{}, errors.WithDetails(errors.ErrAlgorithmNotFound, "algorithm with ID '%s'", algorithmID)
		}
		return domain.Algorithm{}, errors.WithDetails(errors.ErrInternal, "failed to get algorithm: %v", err)
	}

	return algorithm, nil
}

// UpdateAlgorithm updates an existing algorithm
func (s *algorithmsService) UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error {
	if algorithmID == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	// Verify algorithm exists
	_, err := s.db.GetAlgorithm(ctx, algorithmID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.WithDetails(errors.ErrAlgorithmNotFound, "algorithm with ID '%s'", algorithmID)
		}
		return errors.WithDetails(errors.ErrInternal, "failed to verify algorithm existence: %v", err)
	}

	err = s.db.UpdateAlgorithm(ctx, algorithmID, updates)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to update algorithm: %v", err)
	}

	log.Info("Algorithm updated", "algorithmID", algorithmID)
	return nil
}

// DeleteAlgorithm deletes an algorithm
func (s *algorithmsService) DeleteAlgorithm(ctx context.Context, algorithmID string) error {
	if algorithmID == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	err := s.db.DeleteAlgorithm(ctx, algorithmID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.WithDetails(errors.ErrAlgorithmNotFound, "algorithm with ID '%s'", algorithmID)
		}
		return errors.WithDetails(errors.ErrInternal, "failed to delete algorithm: %v", err)
	}

	log.Info("Algorithm deleted", "algorithmID", algorithmID)
	return nil
}

// GetAlgorithmsCount returns the total count of algorithms
func (s *algorithmsService) GetAlgorithmsCount(ctx context.Context) (int, error) {
	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return 0, errors.WithDetails(errors.ErrInternal, "failed to get auth context: %v", err)
	}

	var count int
	if authCtx.IsSuperAdmin {
		// Superadmin can see count for all algorithms across all tenants
		count, err = s.db.GetAllAlgorithmsCount(ctx)
	} else {
		// Regular users see count only for their tenant
		count, err = s.db.GetAlgorithmsCount(ctx, authCtx.TenantID)
	}

	if err != nil {
		return 0, errors.WithDetails(errors.ErrInternal, "failed to get algorithms count: %v", err)
	}

	return count, nil
}

// CreateAlgorithmRun creates a new algorithm run
func (s *algorithmsService) CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.RunStartResponse, error) {
	// Validate input
	if newRun.AlgorithmID == "" {
		return domain.RunStartResponse{}, errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	// Verify algorithm exists
	algorithm, err := s.db.GetAlgorithm(ctx, newRun.AlgorithmID)
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.RunStartResponse{}, errors.WithDetails(errors.ErrAlgorithmNotFound, "algorithm with ID '%s'", newRun.AlgorithmID)
		}
		return domain.RunStartResponse{}, errors.WithDetails(errors.ErrInternal, "failed to verify algorithm existence: %v", err)
	}

	// Validate that either case_id or image_ids is provided
	if newRun.CaseID == nil && len(newRun.ImageIDs) == 0 {
		return domain.RunStartResponse{}, errors.WithDetails(errors.ErrInvalidInput, "either case_id or image_ids must be provided")
	}

	run, err := s.db.CreateAlgorithmRun(ctx, newRun)
	if err != nil {
		log.Error("Failed to create algorithm run", "error", err, "algorithmID", newRun.AlgorithmID)
		return domain.RunStartResponse{}, errors.WithDetails(errors.ErrInternal, "failed to create algorithm run: %v", err)
	}

	// Determine the events URL based on progress transport
	eventsURL := fmt.Sprintf("/v1/runs/%s/events", run.ID)
	if algorithm.ProgressTransport == "WEBSOCKET" {
		eventsURL = fmt.Sprintf("/v1/runs/%s/ws", run.ID)
	}

	response := domain.RunStartResponse{
		RunID:     run.ID,
		Status:    run.Status,
		StatusURL: fmt.Sprintf("/v1/runs/%s", run.ID),
		EventsURL: eventsURL,
	}

	log.Info("Algorithm run created", "runID", run.ID, "algorithmID", newRun.AlgorithmID)
	return response, nil
}

// GetAlgorithmRuns retrieves algorithm runs with search and pagination
func (s *algorithmsService) GetAlgorithmRuns(ctx context.Context, algorithmID string, params utils.PaginationAndSearchParams) ([]domain.AlgorithmRun, domain.PaginationInfo, error) {
	if algorithmID == "" {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	// Verify algorithm exists
	_, err := s.db.GetAlgorithm(ctx, algorithmID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrAlgorithmNotFound, "algorithm with ID '%s'", algorithmID)
		}
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to verify algorithm existence: %v", err)
	}

	runs, err := s.db.GetAlgorithmRuns(ctx, algorithmID, params.SearchParams, params.PaginationParams)
	if err != nil {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get algorithm runs: %v", err)
	}

	// Get count for pagination
	var totalCount int
	if len(params.SearchParams.Filters) > 0 {
		totalCount, err = s.db.GetRunsCountWithSearch(ctx, algorithmID, params.SearchParams)
	} else {
		totalCount, err = s.db.GetRunsCount(ctx, algorithmID)
	}
	if err != nil {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrInternal, "failed to get runs count: %v", err)
	}

	// Ensure we return an empty slice instead of nil to avoid null in JSON
	if runs == nil {
		runs = []domain.AlgorithmRun{}
	}

	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return runs, paginationInfo, nil
}

// GetAlgorithmRun retrieves a single algorithm run by ID
func (s *algorithmsService) GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error) {
	if runID == "" {
		return domain.AlgorithmRun{}, errors.WithDetails(errors.ErrInvalidInput, "run ID cannot be empty")
	}

	run, err := s.db.GetAlgorithmRun(ctx, runID)
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.AlgorithmRun{}, errors.WithDetails(errors.ErrRunNotFound, "run with ID '%s'", runID)
		}
		return domain.AlgorithmRun{}, errors.WithDetails(errors.ErrInternal, "failed to get algorithm run: %v", err)
	}

	return run, nil
}

// UpdateAlgorithmRun updates an existing algorithm run
func (s *algorithmsService) UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error {
	if runID == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "run ID cannot be empty")
	}

	// Verify run exists
	_, err := s.db.GetAlgorithmRun(ctx, runID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.WithDetails(errors.ErrRunNotFound, "run with ID '%s'", runID)
		}
		return errors.WithDetails(errors.ErrInternal, "failed to verify run existence: %v", err)
	}

	err = s.db.UpdateAlgorithmRun(ctx, runID, updates)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to update algorithm run: %v", err)
	}

	log.Info("Algorithm run updated", "runID", runID)
	return nil
}

// CancelAlgorithmRun cancels an algorithm run
func (s *algorithmsService) CancelAlgorithmRun(ctx context.Context, runID string) error {
	if runID == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "run ID cannot be empty")
	}

	// Verify run exists and is cancellable
	run, err := s.db.GetAlgorithmRun(ctx, runID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.WithDetails(errors.ErrRunNotFound, "run with ID '%s'", runID)
		}
		return errors.WithDetails(errors.ErrInternal, "failed to verify run existence: %v", err)
	}

	if run.Status != "QUEUED" && run.Status != "RUNNING" {
		return errors.WithDetails(errors.ErrInvalidInput, "cannot cancel run with status '%s'", run.Status)
	}

	err = s.db.CancelAlgorithmRun(ctx, runID)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to cancel algorithm run: %v", err)
	}

	log.Info("Algorithm run cancelled", "runID", runID)
	return nil
}

// GetRunsCount returns the total count of runs for an algorithm
func (s *algorithmsService) GetRunsCount(ctx context.Context, algorithmID string) (int, error) {
	if algorithmID == "" {
		return 0, errors.WithDetails(errors.ErrInvalidInput, "algorithm ID cannot be empty")
	}

	count, err := s.db.GetRunsCount(ctx, algorithmID)
	if err != nil {
		return 0, errors.WithDetails(errors.ErrInternal, "failed to get runs count: %v", err)
	}

	return count, nil
}

// CreateOutput creates a new output for a run
func (s *algorithmsService) CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error) {
	// Validate input
	if newOutput.RunID == "" {
		return domain.Output{}, errors.WithDetails(errors.ErrInvalidInput, "run ID cannot be empty")
	}
	if newOutput.Name == "" {
		return domain.Output{}, errors.WithDetails(errors.ErrInvalidInput, "output name cannot be empty")
	}
	if newOutput.URI == "" {
		return domain.Output{}, errors.WithDetails(errors.ErrInvalidInput, "output URI cannot be empty")
	}

	// Verify run exists
	_, err := s.db.GetAlgorithmRun(ctx, newOutput.RunID)
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.Output{}, errors.WithDetails(errors.ErrRunNotFound, "run with ID '%s'", newOutput.RunID)
		}
		return domain.Output{}, errors.WithDetails(errors.ErrInternal, "failed to verify run existence: %v", err)
	}

	output, err := s.db.CreateOutput(ctx, newOutput)
	if err != nil {
		log.Error("Failed to create output", "error", err, "runID", newOutput.RunID)
		return domain.Output{}, errors.WithDetails(errors.ErrInternal, "failed to create output: %v", err)
	}

	log.Info("Output created", "outputID", output.ID, "runID", newOutput.RunID)
	return output, nil
}

// GetOutputs retrieves all outputs for a run
func (s *algorithmsService) GetOutputs(ctx context.Context, runID string) ([]domain.Output, error) {
	if runID == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "run ID cannot be empty")
	}

	// Verify run exists
	_, err := s.db.GetAlgorithmRun(ctx, runID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.WithDetails(errors.ErrRunNotFound, "run with ID '%s'", runID)
		}
		return nil, errors.WithDetails(errors.ErrInternal, "failed to verify run existence: %v", err)
	}

	outputs, err := s.db.GetOutputs(ctx, runID)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrInternal, "failed to get outputs: %v", err)
	}

	// Ensure we return an empty slice instead of nil to avoid null in JSON
	if outputs == nil {
		outputs = []domain.Output{}
	}

	return outputs, nil
}

// GetOutput retrieves a single output by ID
func (s *algorithmsService) GetOutput(ctx context.Context, outputID string) (domain.Output, error) {
	if outputID == "" {
		return domain.Output{}, errors.WithDetails(errors.ErrInvalidInput, "output ID cannot be empty")
	}

	output, err := s.db.GetOutput(ctx, outputID)
	if err != nil {
		if errors.IsNotFound(err) {
			return domain.Output{}, errors.WithDetails(errors.ErrOutputNotFound, "output with ID '%s'", outputID)
		}
		return domain.Output{}, errors.WithDetails(errors.ErrInternal, "failed to get output: %v", err)
	}

	return output, nil
}

// DeleteOutput deletes an output
func (s *algorithmsService) DeleteOutput(ctx context.Context, outputID string) error {
	if outputID == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "output ID cannot be empty")
	}

	err := s.db.DeleteOutput(ctx, outputID)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.WithDetails(errors.ErrOutputNotFound, "output with ID '%s'", outputID)
		}
		return errors.WithDetails(errors.ErrInternal, "failed to delete output: %v", err)
	}

	log.Info("Output deleted", "outputID", outputID)
	return nil
}

// Close releases all resources held by the service
func (s *algorithmsService) Close() {
	// no-op
}
