// Package main Workout Assistant API.
// @title Workout Assistant API
// @version 1.0
// @description API for authentication and health endpoints.
// @BasePath /v1
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/degeta10/workout-assistant-api/internal/auth"
	"github.com/degeta10/workout-assistant-api/internal/config"
	"github.com/degeta10/workout-assistant-api/internal/database"
	"github.com/degeta10/workout-assistant-api/internal/health"
	"github.com/degeta10/workout-assistant-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	ginLambda *ginadapter.GinLambdaV2
	globalDB  *sql.DB
)

// Handler is the entry point for AWS Lambda
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if ginLambda == nil {
		cfg := config.LoadConfig()
		setupLogger(cfg.AppEnv)

		db, err := database.InitDBWithContext(ctx, cfg.DB)
		if err != nil {
			slog.Error("Critical: Database connection failed", "error", err.Error())
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"message":"internal server error"}`,
			}, nil
		}
		globalDB = db // Store globally so we could theoretically close it if needed

		router := setupRouter(cfg, db)
		ginLambda = ginadapter.NewV2(router)
	}

	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// 1. Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
		return
	}

	// 2. LOCAL SERVER MODE
	if err := godotenv.Load(); err != nil {
		slog.Info("Info: .env file not found, using system environment variables")
	}

	cfg := config.LoadConfig()
	setupLogger(cfg.AppEnv)
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer bootCancel()
	db, err := database.InitDBWithContext(bootCtx, cfg.DB)
	if err != nil {
		slog.Error("Critical: Database connection failed", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	router := setupRouter(cfg, db)
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		slog.Info("Starting server",
			"app_name", cfg.AppName,
			"version", cfg.AppVersion,
			"port", cfg.AppPort,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err.Error())
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Server exiting")
}

func setupRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	// Set APP_ENV before initialization to silence warnings
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() to have full control over middleware
	r := gin.New()

	// Manually attach standard middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())

	// Dependency Injection
	healthRepo := health.NewRepository(db)
	healthSvc := health.NewService(healthRepo, cfg.AppName, cfg.AppVersion)
	healthHandler := health.NewHandler(healthSvc)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	public := r.Group("/v1")
	{
		healthHandler.RegisterRoutes(public)
		authHandler.RegisterPublicRoutes(public)
	}

	protected := r.Group("/v1")
	protected.Use(auth.RequireAuth(cfg.JWTSecret))
	{
		authHandler.RegisterProtectedRoutes(protected)
	}

	return r
}

func setupLogger(env string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "local" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger) // This makes slog.Info() available globally
}
