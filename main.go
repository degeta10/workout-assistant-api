// @title Workout Assistant API
// @version 1.0
// @description This is a high-performance custom workout app.
// @host localhost:8080
// @BasePath /v1
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/degeta10/workout-assistant-api/internal/auth"
	"github.com/degeta10/workout-assistant-api/internal/config"
	"github.com/degeta10/workout-assistant-api/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var ginLambda *ginadapter.GinLambdaV2

// setupRouter initializes all routes and middleware
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())

	v1 := r.Group("/v1")
	{
		v1.GET("/health", healthCheckHandler)

		// Auth "Controller" Routes
		v1.POST("/register", auth.Register)
		v1.POST("/login", auth.Login)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         "Route not found",
			"received_path": c.Request.URL.Path,
		})
	})

	return r
}

func healthCheckHandler(c *gin.Context) {
	status := "pass"
	dbStatus := "connected"

	if database.DB != nil {
		if err := database.DB.Ping(); err != nil {
			status = "fail"
			dbStatus = "disconnected"
		}
	} else {
		dbStatus = "not_initialized"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"version":     "1.0.0",
		"release_id":  time.Now().Format("2006-01-02"),
		"description": config.LoadConfig().AppName + " API",
		"checks": gin.H{
			"database": gin.H{
				"status":         dbStatus,
				"component_type": "datastore",
				"time":           time.Now().Format(time.RFC3339),
			},
		},
	})
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	req.RawPath = strings.TrimPrefix(req.RawPath, "/dev")
	req.RawPath = strings.TrimPrefix(req.RawPath, "/prod")

	if ginLambda == nil {
		gin.SetMode(gin.ReleaseMode)

		// 1. Load config in Lambda environment
		cfg := config.LoadConfig()

		// 2. Initialize DB for Lambda
		if err := database.InitDB(cfg.DB); err != nil {
			log.Printf("Lambda DB Init failed: %v", err)
		}

		router := setupRouter()
		ginLambda = ginadapter.NewV2(router)
	}

	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	isLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""

	if !isLambda {
		log.Printf("--- Starting %s API Locally ---", config.LoadConfig().AppName)

		// Load .env only for local development
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: No .env file found, using system environment variables")
		}

		// 1. Load the Config struct using our new package
		cfg := config.LoadConfig()

		// 2. Pass only the DB portion of the config to InitDB
		if err := database.InitDB(cfg.DB); err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}

		router := setupRouter()

		log.Printf("Server running at: http://localhost:%s", cfg.AppPort)
		if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
			log.Fatalf("Failed to start local server: %v", err)
		}
	} else {
		// AWS LAMBDA FLOW
		lambda.Start(Handler)
	}
}
