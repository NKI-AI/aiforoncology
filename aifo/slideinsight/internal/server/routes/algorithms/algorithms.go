// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package algorithms

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupAlgorithmRoutes configures all algorithm related routes
func SetupAlgorithmRoutes(protectedRoutes fiber.Router, algorithmsService services.AlgorithmsService) {
	// Algorithm management routes
	protectedRoutes.Post("/algorithms", handlers.CreateAlgorithm(algorithmsService))
	protectedRoutes.Get("/algorithms", handlers.GetAlgorithms(algorithmsService))
	protectedRoutes.Get("/algorithms/count", handlers.GetAlgorithmsCount(algorithmsService))
	protectedRoutes.Get("/algorithms/:algorithmId", handlers.GetAlgorithm(algorithmsService))
	protectedRoutes.Put("/algorithms/:algorithmId", handlers.UpdateAlgorithm(algorithmsService))
	protectedRoutes.Delete("/algorithms/:algorithmId", handlers.DeleteAlgorithm(algorithmsService))

	// Algorithm run management routes
	protectedRoutes.Get("/algorithms/:algorithmId/runs", handlers.GetAlgorithmRuns(algorithmsService))
	protectedRoutes.Get("/algorithms/:algorithmId/runs/count", handlers.GetRunsCount(algorithmsService))

	// Run creation and management (unified endpoint)
	protectedRoutes.Post("/runs", handlers.CreateAlgorithmRun(algorithmsService))
	protectedRoutes.Get("/runs/:runId", handlers.GetAlgorithmRun(algorithmsService))
	protectedRoutes.Post("/runs/:runId/cancel", handlers.CancelAlgorithmRun(algorithmsService))

	// Output management routes
	protectedRoutes.Post("/outputs", handlers.CreateOutput(algorithmsService))
	protectedRoutes.Get("/runs/:runId/outputs", handlers.GetOutputs(algorithmsService))
	protectedRoutes.Get("/outputs/:outputId", handlers.GetOutput(algorithmsService))
	protectedRoutes.Delete("/outputs/:outputId", handlers.DeleteOutput(algorithmsService))

	// TODO: WebSocket and SSE endpoints for real-time progress updates
	// These would typically be implemented as separate handlers:
	// protectedRoutes.Get("/runs/:runId/ws", handlers.HandleRunWebSocket(algorithmsService))
	// protectedRoutes.Get("/runs/:runId/events", handlers.HandleRunSSE(algorithmsService))
}
