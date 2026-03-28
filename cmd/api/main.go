package main

import (
	"context"
	"database/sql"
	"log"
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

		db, err := database.InitDB(cfg.DB)
		if err != nil {
			log.Printf("Critical: Database connection failed: %v", err)
			return events.APIGatewayV2HTTPResponse{StatusCode: 500}, err
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
		log.Println("Info: .env file not found, using system environment variables")
	}

	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg.DB)
	if err != nil {
		log.Fatalf("Critical: Database connection failed: %v", err)
	}
	defer db.Close()

	router := setupRouter(cfg, db)
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Printf("Starting %s v%s on local port %s", cfg.AppName, cfg.AppVersion, cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

func setupRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	// Set APP_ENV before initialization to silence warnings
	if cfg.AppEnv == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() to have full control over middleware
	r := gin.New()

	// Manually attach standard middleware
	r.Use(gin.Recovery())
	if cfg.AppEnv != "release" {
		r.Use(gin.Logger())
	}

	// Dependency Injection
	healthRepo := health.NewRepository(db)
	healthSvc := health.NewService(healthRepo, cfg.AppName, cfg.AppVersion)
	healthHandler := health.NewHandler(healthSvc)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	v1 := r.Group("/v1")
	{
		healthHandler.RegisterRoutes(v1)
		authHandler.RegisterRoutes(v1)
	}

	return r
}
