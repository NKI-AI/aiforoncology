// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package server

import (
	"context"
	"encoding/json"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/datasources"
	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/email"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/routes"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	authService "aifo.dev/aifo/slideinsight/internal/server/services/auth"
	"aifo.dev/aifo/slideinsight/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewServer(ctx context.Context, config config.Config, dataSources *datasources.DataSources) *fiber.App {
	// Configure Fiber
	app := fiber.New(fiber.Config{
		AppName:               "SlideInsight",
		CaseSensitive:         true,
		EnablePrintRoutes:     false,
		DisableStartupMessage: false,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		// Prefork:               false, // Disable threading/multiprocessing for streaming stability
		// DisableKeepalive:      false, // Keep connections alive for better streaming performance
		// ReadTimeout:           0,     // Disable read timeout for long streaming operations
		// WriteTimeout:          0,     // Disable write timeout for long streaming operations
		// IdleTimeout:           0,     // Disable idle timeout for persistent connections
	})

	// Add a request ID to each request
	RequestIDConfig := requestid.Config{
		Next:       nil,
		Header:     fiber.HeaderXRequestID,
		Generator:  uuid.NewString,
		ContextKey: "request_id",
	}

	app.Use(requestid.New(RequestIDConfig))

	// Configure zap logger with production-ready settings
	var zapConfig zap.Config
	if config.Logging.Level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = "json"
	}

	zapConfig.EncoderConfig.TimeKey = "time"
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.MessageKey = "msg"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Set log level from config
	switch config.Logging.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	// Set the global logger
	zap.ReplaceGlobals(zapLogger)

	// Use standard fiber logger middleware for HTTP request logging
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
		Next: func(c *fiber.Ctx) bool {
			// Skip logging for certain endpoints and status codes
			status := c.Response().StatusCode()
			path := c.Path()

			// Skip logging for successful static file requests (200 on non-API paths)
			if status == 200 && !strings.HasPrefix(path, "/api/") {
				return true
			}

			// Skip logging for WebSocket upgrade requests (101)
			if status == 101 {
				return true
			}

			// Skip logging for health check endpoints
			if strings.Contains(path, "/api/status") {
				return true
			}

			// Skip logging for tile requests and other static content
			if strings.Contains(path, "/tiles/") || strings.Contains(path, "/thumbnails/") {
				return true
			}

			// Skip successful GET requests to reduce noise
			if c.Method() == "GET" && status >= 200 && status < 300 {
				return true
			}

			// Log all other requests (POST, PUT, DELETE, errors, etc.)
			return false
		},
	}))

	// Custom middleware to add structured logger helpers to context
	app.Use(func(c *fiber.Ctx) error {
		// Get the request ID from the header (set by requestid middleware)
		requestID := c.Get(fiber.HeaderXRequestID)

		// Create a logger with request ID for this request
		reqLogger := zapLogger.With(zap.String("request_id", requestID))
		sugar := reqLogger.Sugar()

		// Set structured logger and sugar logger in context
		ctx := context.WithValue(c.UserContext(), "logger", reqLogger)
		ctx = context.WithValue(ctx, "sugar", sugar)
		ctx = context.WithValue(ctx, "request_id", requestID)
		c.SetUserContext(ctx)

		// Also set as locals for direct access
		c.Locals("logger", reqLogger)
		c.Locals("sugar", sugar)
		c.Locals("request_id", requestID)

		return c.Next()
	})

	// Add HTTP monitoring middleware for system metrics
	app.Use(handlers.HTTPMonitoringMiddleware())

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
	serviceContainer := createServices(dataSources, config)

	// Log auth configuration for debugging
	logAuthConfiguration(config, zapLogger)

	// Setup all routes using the modular approach
	routes.SetupAllRoutes(apiRoutes, serviceContainer, config)

	// Handle 404 for unmatched API routes
	apiRoutes.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API endpoint not found",
			"path":  c.Path(),
		})
	})

	// Serve the UI as catch-all (this will only match non-API routes due to route order)
	app.Use("/", ui.DistDir)

	return app
}

// createServices initializes all services and returns them in a container
func createServices(dataSources *datasources.DataSources, config config.Config) *routes.ServiceContainer {
	logger := zap.L()
	sugar := logger.Sugar()

	sugar.Info("Creating services")
	sugar.Infow("Email config",
		"provider", config.Email.Provider,
		"ses_region", config.Email.SES.Region,
		"from_address", config.Email.SES.FromAddress,
	)

	// Create email service with template repository
	emailService, err := services.NewEmailService(context.Background(), config.Email, dataSources.DB)
	if err != nil {
		logger.Error("Failed to create email service", zap.Error(err))
		// Use a fallback console email service on error
		fallbackConfig := config.Email // Start with the existing config
		fallbackConfig.Provider = "console"
		fallbackConfig.SES.Region = "us-east-1"
		fallbackConfig.SES.FromAddress = "noreply@example.com"
		fallbackConfig.SES.FromName = "SlideInsight"
		emailService, _ = services.NewEmailService(context.Background(), fallbackConfig, dataSources.DB)
	}

	// Create queue manager
	queueConfig := queue.DefaultQueueConfig()
	queueManager := queue.NewQueueManager(queueConfig, emailService)

	// Start the queue manager
	if err := queueManager.Start(); err != nil {
		logger.Error("Failed to start queue manager", zap.Error(err))
	} else {
		logger.Info("Queue manager started successfully")
	}

	// Create async email service
	asyncEmailService := email.NewAsyncEmailService(emailService, queueManager)

	// Create domain services
	slidesService := services.NewSlidesService(dataSources.DB)
	masksService := services.NewRasterAnnotationsService(dataSources.DB)
	vectorAnnotationsService := services.NewVectorAnnotationsService(dataSources.DB)
	tenantService := services.NewTenantsService(dataSources.DB)
	userService := services.NewUserService(dataSources.DB)
	studiesService := services.NewStudiesService(dataSources.DB)
	authService := authService.NewAuthService(dataSources.DB, config.Auth, emailService, userService, queueManager)
	casesService := services.NewCasesService(dataSources.DB, slidesService)
	permissionService := services.NewPermissionService(dataSources.DB)
	roleService := services.NewRoleService(dataSources.DB)
	groupService := services.NewGroupService(dataSources.DB)
	objectGrantService := services.NewObjectGrantService(dataSources.DB)
	algorithmsService := services.NewAlgorithmsService(dataSources.DB, tenantService)
	settingsService := services.NewSettingsService(dataSources.DB)

	// Create image types related services
	imageTypesService := services.NewImageTypesService(dataSources.DB)
	slideHistogramsService := services.NewSlideHistogramsService(dataSources.DB)
	stainingProtocolsService := services.NewStainingProtocolsService(dataSources.DB)

	// Create notification service
	notificationService := services.NewNotificationService(dataSources.DB, userService)

	// Create email template service (user ID will be set per request via middleware)
	emailTemplateService := services.NewEmailTemplateService(dataSources.DB, 0) // 0 as placeholder, will be set per request

	// Create application service with all domain services for complete orchestration
	applicationService := services.NewApplicationService(
		dataSources.DB,
		slidesService,
		casesService,
		masksService,
		vectorAnnotationsService,
		studiesService,
		tenantService,
		userService,
	)

	// Create regions service
	regionsService := services.NewRegionsService(dataSources.DB, services.NewBaseService(dataSources.DB))

	return &routes.ServiceContainer{
		Database:                 dataSources.DB,
		SlidesService:            slidesService,
		MasksService:             masksService,
		VectorAnnotationsService: vectorAnnotationsService,
		RegionsService:           regionsService,
		TenantService:            tenantService,
		UserService:              userService,
		StudiesService:           studiesService,
		AuthService:              authService,
		CasesService:             casesService,
		ApplicationService:       applicationService,
		PermissionService:        permissionService,
		RoleService:              roleService,
		GroupService:             groupService,
		ObjectGrantService:       objectGrantService,
		NotificationService:      notificationService,
		EmailService:             emailService,
		AsyncEmailService:        asyncEmailService,
		EmailTemplateService:     emailTemplateService,
		AlgorithmsService:        algorithmsService,
		ImageTypesService:        imageTypesService,
		SlideHistogramsService:   slideHistogramsService,
		StainingProtocolsService: stainingProtocolsService,
		QueueManager:             queueManager,
		SettingsService:          settingsService,
	}
}

// logAuthConfiguration logs the auth configuration for debugging
func logAuthConfiguration(config config.Config, logger *zap.Logger) {
	// Use structured logging for auth configuration
	logger.Info("Auth configuration loaded",
		zap.String("jwt_algorithm", config.Auth.JWTAlgorithm),
		zap.Duration("jwt_expiration", config.Auth.JWTExpirationMinutes),
		zap.Duration("jwt_refresh_expiration", config.Auth.JWTRefreshExpirationMinutes),
		zap.String("cookie_name", config.Auth.Cookie.Name),
		zap.String("cookie_path", config.Auth.Cookie.Path),
		zap.Bool("cookie_httponly", config.Auth.Cookie.HTTPOnly),
		zap.Bool("cookie_secure", config.Auth.Cookie.Secure),
		zap.String("cookie_samesite", config.Auth.Cookie.SameSite),
		zap.Bool("jwt_secret_set", config.Auth.JWTSecret != ""),
	)

	// Use debug level for full auth config (contains sensitive data)
	logger.Debug("Full auth config", zap.Any("config", config.Auth))
}
