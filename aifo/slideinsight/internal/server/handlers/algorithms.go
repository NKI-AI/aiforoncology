// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"errors"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Algorithm input/output types for API handlers

// AlgorithmInput represents the input for creating an algorithm
type AlgorithmInput struct {
	TenantUID         string `json:"tenant_uid,omitempty"` // Optional - if not provided, use user's tenant
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description,omitempty"`
	Version           string `json:"version" validate:"required"`
	EndpointURL       string `json:"endpoint_url" validate:"required,url"`
	HTTPMethod        string `json:"http_method,omitempty"`
	ExecutionMode     string `json:"execution_mode" validate:"required,oneof=BATCH STREAM"`
	ProgressTransport string `json:"progress_transport,omitempty"`
	Metadata          string `json:"metadata,omitempty"`
}

// AlgorithmRunInput represents the input for creating an algorithm run
type AlgorithmRunInput struct {
	AlgorithmID   string                 `json:"algorithm_id" validate:"required"`
	CaseID        *string                `json:"case_id,omitempty"`
	ImageIDs      []string               `json:"image_ids,omitempty"`
	Regions       string                 `json:"regions,omitempty"`
	Parameters    string                 `json:"parameters,omitempty"`
	ExecutionMode *string                `json:"execution_mode,omitempty"`
	ResultSinks   []domain.NewResultSink `json:"result_sinks,omitempty"`
	ProgressSinks []domain.NewResultSink `json:"progress_sinks,omitempty"`
	PostHooks     []domain.NewPostHook   `json:"post_hooks,omitempty"`
}

// OutputInput represents the input for creating an output
type OutputInput struct {
	RunID    string `json:"run_id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	URI      string `json:"uri" validate:"required"`
	Metadata string `json:"metadata,omitempty"`
}

// Parameter types for path/query parameters

// AlgorithmIDParams represents algorithm ID path parameters
type AlgorithmIDParams struct {
	AlgorithmID string `params:"algorithmId" validate:"required"`
}

// RunIDParams represents run ID path parameters
type RunIDParams struct {
	RunID string `params:"runId" validate:"required"`
}

// OutputIDParams represents output ID path parameters
type OutputIDParams struct {
	OutputID string `params:"outputId" validate:"required"`
}

// CreateAlgorithm creates a new algorithm
// @Summary Create a new algorithm
// @Description Register a new algorithm in the system
// @Tags algorithms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param algorithm body AlgorithmInput true "Algorithm details"
// @Success 201 {object} domain.Algorithm "Created algorithm"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms [post]
func CreateAlgorithm(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input AlgorithmInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateAlgorithm request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Get auth context to determine tenant ID
		// The service will automatically set the tenant ID from authentication context
		// or convert the provided tenant_uid if specified
		newAlgorithm := domain.NewAlgorithm{
			TenantUID:         input.TenantUID, // Pass tenant UID if provided
			Name:              input.Name,
			Description:       input.Description,
			Version:           input.Version,
			EndpointURL:       input.EndpointURL,
			HTTPMethod:        input.HTTPMethod,
			ExecutionMode:     input.ExecutionMode,
			ProgressTransport: input.ProgressTransport,
			Metadata:          input.Metadata,
		}

		createdAlgorithm, err := service.CreateAlgorithm(c.UserContext(), newAlgorithm)
		if err != nil {
			log.Error("CreateAlgorithm failed", "error", err, "name", input.Name)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdAlgorithm)
	}
}

// GetAlgorithms retrieves algorithms with search/filter and pagination
// @Summary Get algorithms
// @Description Retrieve algorithms with pagination and search support. Superadmins see all algorithms across tenants, regular users see only their tenant's algorithms.
// @Tags algorithms
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param q query string false "General search across algorithm name and description" example("tissue")
// @Param name query string false "Filter by algorithm name" example("Tissue Segmentation")
// @Param version query string false "Filter by version" example("1.0.0")
// @Param execution_mode query string false "Filter by execution mode" example("BATCH")
// @Success 200 {object} domain.AlgorithmsResponse "List of algorithms"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms [get]
func GetAlgorithms(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse pagination and search parameters
		params, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		algorithms, paginationInfo, err := service.GetAlgorithms(c.UserContext(), params)
		if err != nil {
			log.Error("GetAlgorithms failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.AlgorithmsResponse{
			Algorithms: algorithms,
			Pagination: paginationInfo,
		})
	}
}

// GetAlgorithm retrieves a specific algorithm by ID
// @Summary Get algorithm by ID
// @Description Retrieve a specific algorithm by its unique identifier
// @Tags algorithms
// @Produce json
// @Security BearerAuth
// @Param algorithmId path string true "Algorithm ID" example("uuid-algo-1234")
// @Success 200 {object} domain.Algorithm "Algorithm details"
// @Failure 400 {object} domain.ErrorResponse "Bad request - algorithm ID required"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/{algorithmId} [get]
func GetAlgorithm(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params AlgorithmIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetAlgorithm request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		algorithm, err := service.GetAlgorithm(c.UserContext(), params.AlgorithmID)
		if err != nil {
			log.Error("GetAlgorithm failed", "error", err, "algorithmID", params.AlgorithmID)
			if errors.Is(err, apperrors.ErrAlgorithmNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "algorithm not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(algorithm)
	}
}

// UpdateAlgorithm updates an existing algorithm
// @Summary Update algorithm
// @Description Update an existing algorithm
// @Tags algorithms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param algorithmId path string true "Algorithm ID" example("uuid-algo-1234")
// @Param updates body domain.AlgorithmUpdates true "Algorithm updates"
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/{algorithmId} [put]
func UpdateAlgorithm(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params AlgorithmIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateAlgorithm request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var updates domain.AlgorithmUpdates
		if err := c.BodyParser(&updates); err != nil {
			log.Warn("UpdateAlgorithm request body parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		err := service.UpdateAlgorithm(c.UserContext(), params.AlgorithmID, updates)
		if err != nil {
			log.Error("UpdateAlgorithm failed", "error", err, "algorithmID", params.AlgorithmID)
			if errors.Is(err, apperrors.ErrAlgorithmNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "algorithm not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Algorithm updated successfully"})
	}
}

// DeleteAlgorithm deletes an algorithm
// @Summary Delete algorithm
// @Description Delete an algorithm and all associated runs
// @Tags algorithms
// @Produce json
// @Security BearerAuth
// @Param algorithmId path string true "Algorithm ID" example("uuid-algo-1234")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - algorithm ID required"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/{algorithmId} [delete]
func DeleteAlgorithm(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params AlgorithmIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteAlgorithm request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteAlgorithm(c.UserContext(), params.AlgorithmID)
		if err != nil {
			log.Error("DeleteAlgorithm failed", "error", err, "algorithmID", params.AlgorithmID)
			if errors.Is(err, apperrors.ErrAlgorithmNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "algorithm not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Algorithm deleted successfully"})
	}
}

// CreateAlgorithmRun creates a new algorithm run
// @Summary Start a new algorithm run
// @Description Start a new algorithm run (image or case level)
// @Tags runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param run body AlgorithmRunInput true "Run details"
// @Success 201 {object} domain.RunStartResponse "Created run"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/runs [post]
func CreateAlgorithmRun(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input AlgorithmRunInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateAlgorithmRun request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		newRun := domain.NewAlgorithmRun{
			AlgorithmID:   input.AlgorithmID,
			CaseID:        input.CaseID,
			ImageIDs:      input.ImageIDs,
			Regions:       input.Regions,
			Parameters:    input.Parameters,
			ExecutionMode: input.ExecutionMode,
			ResultSinks:   input.ResultSinks,
			ProgressSinks: input.ProgressSinks,
			PostHooks:     input.PostHooks,
		}

		response, err := service.CreateAlgorithmRun(c.UserContext(), newRun)
		if err != nil {
			log.Error("CreateAlgorithmRun failed", "error", err, "algorithmID", input.AlgorithmID)
			if errors.Is(err, apperrors.ErrAlgorithmNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "algorithm not found")
			}
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(response)
	}
}

// GetAlgorithmRuns retrieves runs for an algorithm with search and pagination
// @Summary Get algorithm runs
// @Description Retrieve runs for a specific algorithm with search and pagination
// @Tags runs
// @Produce json
// @Security BearerAuth
// @Param algorithmId path string true "Algorithm ID" example("uuid-algo-1234")
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param status query string false "Filter by status" example("RUNNING")
// @Param execution_mode query string false "Filter by execution mode" example("BATCH")
// @Success 200 {object} domain.RunsResponse "List of runs"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/{algorithmId}/runs [get]
func GetAlgorithmRuns(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params AlgorithmIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetAlgorithmRuns request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Parse pagination and search parameters
		paginationParams, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		runs, paginationInfo, err := service.GetAlgorithmRuns(c.UserContext(), params.AlgorithmID, paginationParams)
		if err != nil {
			log.Error("GetAlgorithmRuns failed", "error", err, "algorithmID", params.AlgorithmID)
			if errors.Is(err, apperrors.ErrAlgorithmNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "algorithm not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.RunsResponse{
			Runs:       runs,
			Pagination: paginationInfo,
		})
	}
}

// GetAlgorithmRun retrieves a specific algorithm run by ID
// @Summary Get algorithm run status
// @Description Retrieve status and details of a specific algorithm run
// @Tags runs
// @Produce json
// @Security BearerAuth
// @Param runId path string true "Run ID" example("run-abc123")
// @Success 200 {object} domain.RunStatusResponse "Run status"
// @Failure 400 {object} domain.ErrorResponse "Bad request - run ID required"
// @Failure 404 {object} domain.ErrorResponse "Run not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/runs/{runId} [get]
func GetAlgorithmRun(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RunIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetAlgorithmRun request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		run, err := service.GetAlgorithmRun(c.UserContext(), params.RunID)
		if err != nil {
			log.Error("GetAlgorithmRun failed", "error", err, "runID", params.RunID)
			if errors.Is(err, apperrors.ErrRunNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "run not found")
			}
			return middleware.HandleError(c, err)
		}

		response := domain.RunStatusResponse{
			RunID:      run.ID,
			Status:     run.Status,
			Progress:   run.Progress,
			StartedAt:  run.StartedAt,
			ResultURI:  run.ResultURI,
			ErrorInfo:  run.ErrorInfo,
			FinishedAt: run.FinishedAt,
		}

		return c.JSON(response)
	}
}

// CancelAlgorithmRun cancels an algorithm run
// @Summary Cancel algorithm run
// @Description Cancel a running or queued algorithm run
// @Tags runs
// @Produce json
// @Security BearerAuth
// @Param runId path string true "Run ID" example("run-abc123")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - run ID required or run not cancellable"
// @Failure 404 {object} domain.ErrorResponse "Run not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/runs/{runId}/cancel [post]
func CancelAlgorithmRun(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RunIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("CancelAlgorithmRun request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.CancelAlgorithmRun(c.UserContext(), params.RunID)
		if err != nil {
			log.Error("CancelAlgorithmRun failed", "error", err, "runID", params.RunID)
			if errors.Is(err, apperrors.ErrRunNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "run not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Run cancelled successfully"})
	}
}

// CreateOutput creates a new output for a run
// @Summary Create algorithm run output
// @Description Create a new output artifact for an algorithm run
// @Tags outputs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param output body OutputInput true "Output details"
// @Success 201 {object} domain.Output "Created output"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Run not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/outputs [post]
func CreateOutput(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input OutputInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateOutput request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		newOutput := domain.NewOutput{
			RunID:    input.RunID,
			Name:     input.Name,
			URI:      input.URI,
			Metadata: input.Metadata,
		}

		output, err := service.CreateOutput(c.UserContext(), newOutput)
		if err != nil {
			log.Error("CreateOutput failed", "error", err, "runID", input.RunID)
			if errors.Is(err, apperrors.ErrRunNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "run not found")
			}
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(output)
	}
}

// GetOutputs retrieves all outputs for a run
// @Summary Get algorithm run outputs
// @Description Retrieve all output artifacts for a specific algorithm run
// @Tags outputs
// @Produce json
// @Security BearerAuth
// @Param runId path string true "Run ID" example("run-abc123")
// @Success 200 {object} domain.OutputsResponse "List of outputs"
// @Failure 400 {object} domain.ErrorResponse "Bad request - run ID required"
// @Failure 404 {object} domain.ErrorResponse "Run not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/runs/{runId}/outputs [get]
func GetOutputs(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RunIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetOutputs request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		outputs, err := service.GetOutputs(c.UserContext(), params.RunID)
		if err != nil {
			log.Error("GetOutputs failed", "error", err, "runID", params.RunID)
			if errors.Is(err, apperrors.ErrRunNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "run not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.OutputsResponse{
			Outputs: outputs,
		})
	}
}

// GetOutput retrieves a specific output by ID
// @Summary Get algorithm output
// @Description Retrieve a specific output artifact by its ID
// @Tags outputs
// @Produce json
// @Security BearerAuth
// @Param outputId path string true "Output ID" example("output-xyz789")
// @Success 200 {object} domain.Output "Output details"
// @Failure 400 {object} domain.ErrorResponse "Bad request - output ID required"
// @Failure 404 {object} domain.ErrorResponse "Output not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/outputs/{outputId} [get]
func GetOutput(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params OutputIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetOutput request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		output, err := service.GetOutput(c.UserContext(), params.OutputID)
		if err != nil {
			log.Error("GetOutput failed", "error", err, "outputID", params.OutputID)
			if errors.Is(err, apperrors.ErrOutputNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "output not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(output)
	}
}

// DeleteOutput deletes an output
// @Summary Delete algorithm output
// @Description Delete an output artifact
// @Tags outputs
// @Produce json
// @Security BearerAuth
// @Param outputId path string true "Output ID" example("output-xyz789")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - output ID required"
// @Failure 404 {object} domain.ErrorResponse "Output not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/outputs/{outputId} [delete]
func DeleteOutput(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params OutputIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteOutput request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteOutput(c.UserContext(), params.OutputID)
		if err != nil {
			log.Error("DeleteOutput failed", "error", err, "outputID", params.OutputID)
			if errors.Is(err, apperrors.ErrOutputNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "output not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Output deleted successfully"})
	}
}

// GetAlgorithmsCount returns the count of algorithms
// @Summary Get algorithms count
// @Description Get the total count of algorithms. Superadmins see count across all tenants, regular users see count for their tenant only.
// @Tags algorithms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Count response"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/count [get]
func GetAlgorithmsCount(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetAlgorithmsCount(c.UserContext())
		if err != nil {
			log.Error("GetAlgorithmsCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// GetRunsCount returns the total count of runs for an algorithm
// @Summary Get algorithm runs count
// @Description Retrieve the total count of runs for a specific algorithm
// @Tags runs
// @Produce json
// @Security BearerAuth
// @Param algorithmId path string true "Algorithm ID" example("uuid-algo-1234")
// @Success 200 {object} fiber.Map "Total count of runs"
// @Failure 400 {object} domain.ErrorResponse "Bad request - algorithm ID required"
// @Failure 404 {object} domain.ErrorResponse "Algorithm not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/algorithms/{algorithmId}/runs/count [get]
func GetRunsCount(service services.AlgorithmsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params AlgorithmIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetRunsCount request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		count, err := service.GetRunsCount(c.UserContext(), params.AlgorithmID)
		if err != nil {
			log.Error("GetRunsCount failed", "error", err, "algorithmID", params.AlgorithmID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}
