// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package server

import (
	"context"
	"embed"
	"log/slog"

	"aifo.dev/aifo/slidescope/internal/config"
	"aifo.dev/aifo/slidescope/internal/datasources"
	"aifo.dev/aifo/slidescope/internal/server/handlers"
	"aifo.dev/aifo/slidescope/internal/server/middleware"
	"aifo.dev/aifo/slidescope/internal/server/services"
	"aifo.dev/aifo/slidescope/ui"

	"github.com/gofiber/fiber/v2"
	// "github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

//go:embed public/openapi.yaml
var openapiYAML embed.FS

const openapiYAMLPath = "public/openapi.yaml"

func NewServer(ctx context.Context, config config.Config, dataSources *datasources.DataSources) *fiber.App {
	// Configure Fiber with custom settings
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// Add a request ID to each request
	// TODO: Add it to the logger
	app.Use(requestid.New())

	// Configure CORS if enabled
	// if cfg.Server.CORS.Enabled {
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins:     strings.Join(cfg.Server.CORS.AllowOrigins, ","),
	// 	AllowMethods:     strings.Join(cfg.Server.CORS.AllowMethods, ","),
	// 	AllowHeaders:     strings.Join(cfg.Server.CORS.AllowHeaders, ","),
	// 	ExposeHeaders:    strings.Join(cfg.Server.CORS.ExposeHeaders, ","),
	// 	AllowCredentials: cfg.Server.CORS.AllowCredentials,
	// 	MaxAge:           cfg.Server.CORS.MaxAge,
	// }))
	// }

	// Set base path if configured
	var apiRoutes fiber.Router
	if config.Server.BasePath != "" {
		apiRoutes = app.Group(config.Server.BasePath + "/api")
	} else {
		apiRoutes = app.Group("/api")
	}

	apiRoutes.Get("/status", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Create services
	// TODO: Inject specific config into the service

	slog.Info("Creating services")
	slidesService := services.NewSlidesService(dataSources.DB)
	slog.Info("Creating masks service")
	masksService := services.NewMasksService(dataSources.DB)
	slog.Info("Creating user service")
	userService := services.NewUserService(dataSources.DB)
	slog.Info("Creating auth service")
	authService := services.NewAuthService(dataSources.DB, config.Auth)

	// Log auth configuration for debugging
	slog.Info("Auth configuration loaded",
		"jwt_algorithm", config.Auth.JWTAlgorithm,
		"jwt_expiration", config.Auth.JWTExpirationMinutes,
		"jwt_refresh_expiration", config.Auth.JWTRefreshExpirationMinutes,
		"cookie_name", config.Auth.Cookie.Name,
		"cookie_path", config.Auth.Cookie.Path,
		"cookie_httponly", config.Auth.Cookie.HTTPOnly,
		"cookie_secure", config.Auth.Cookie.Secure,
		"cookie_samesite", config.Auth.Cookie.SameSite,
		"jwt_secret_set", config.Auth.JWTSecret != "",
	)
	slog.Debug("Full auth config", "config", config.Auth)

	// Auth routes group
	authRoutes := apiRoutes.Group("/v1/auth")

	// Public authentication routes
	authRoutes.Post("/login", handlers.Login(authService))
	authRoutes.Post("/logout", handlers.Logout(authService))
	authRoutes.Post("/refresh", handlers.Refresh(authService))

	// Protected auth routes
	authRoutes.Get("/me", middleware.Protected(config.Auth), handlers.GetCurrentUser(authService))

	// Public user routes
	apiRoutes.Post("/v1/users", handlers.CreateUser(userService)) // Registration is public

	// Protected routes
	protectedRoutes := apiRoutes.Group("/v1")
	protectedRoutes.Use(middleware.Protected(config.Auth))

	// Protected user routes
	protectedRoutes.Get("/users/:username", handlers.GetUserByUsername(userService))

	// Protected slides routes
	protectedRoutes.Get("/slides", handlers.GetSlides(slidesService))
	protectedRoutes.Get("/slides/:slide_id", handlers.GetSlideByID(slidesService))
	protectedRoutes.Get("/slides/:slide_id/metadata", handlers.GetSlideMetadata(slidesService))
	protectedRoutes.Get("/slides/:slide_id/tiles/:z/:x/:y.:format", handlers.GetSlideTile(slidesService))
	protectedRoutes.Post("/slides", handlers.AddSlide(slidesService))

	// Protected masks routes
	protectedRoutes.Get("/slides/:slide_id/annotations/raster", handlers.GetMasks(masksService))
	protectedRoutes.Get("/slides/:slide_id/annotations/raster/default/tiles/:z/:x/:y.:format", handlers.GetMaskTile(masksService))
	protectedRoutes.Get("/slides/:slide_id/annotations/raster/:mask_id/tiles/:z/:x/:y.:format", handlers.GetMaskTile(masksService))
	protectedRoutes.Post("/slides/:slide_id/annotations/raster", handlers.AddMask(masksService))

	// Handle 404 for unmatched API routes
	apiRoutes.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API endpoint not found",
			"path":  c.Path(),
		})
	})

	// Serve the OpenAPI YAML file
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		content, err := openapiYAML.ReadFile(openapiYAMLPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error reading openapi.yaml")
		}
		// Set content type to text/plain so it displays in browser instead of downloading
		return c.Type("text/yaml").Send(content)
	})

	// Serve the UI as catch-all (this will only match non-API routes due to route order)
	app.Use("/", ui.DistDir)

	return app
}
