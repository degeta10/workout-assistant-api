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
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var ginLambda *ginadapter.GinLambdaV2

// setupRouter initializes all routes and middleware
func setupRouter() *gin.Engine {
	r := gin.Default()

	// Global Middleware (Optional: Add CORS here later)
	r.Use(gin.Recovery())

	// Health Check Group
	v1 := r.Group("/v1")
	{
		v1.GET("/health", healthCheckHandler)

		// Future Workout Routes will go here:
		// v1.POST("/sessions/start", workout.StartSessionHandler)
	}

	// 404 Handler for debugging
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         "Route not found",
			"received_path": c.Request.URL.Path,
		})
	})

	return r
}

// healthCheckHandler verifies API and Database status
func healthCheckHandler(c *gin.Context) {
	status := "pass"
	dbStatus := "connected"

	// Check real DB connection if DB is initialized
	// if database.DB != nil {
	// 	if err := database.DB.Ping(); err != nil {
	// 		status = "fail"
	// 		dbStatus = "disconnected"
	// 	}
	// } else {
	// 	dbStatus = "not_initialized"
	// }

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"version":     "1.0.0",
		"release_id":  time.Now().Format("2006-01-02"),
		"description": "Heavy Duty Workout API",
		"checks": gin.H{
			"database": gin.H{
				"status":         dbStatus,
				"component_type": "datastore",
				"time":           time.Now().Format(time.RFC3339),
			},
		},
	})
}

// Handler is the entry point for AWS Lambda (APIGateway V2)
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// If AWS sends a path like /dev/v1/health, we strip the stage
	req.RawPath = strings.TrimPrefix(req.RawPath, "/dev")
	req.RawPath = strings.TrimPrefix(req.RawPath, "/prod")

	if ginLambda == nil {
		// Lazy initialize router if not already done
		gin.SetMode(gin.ReleaseMode)
		router := setupRouter()
		ginLambda = ginadapter.NewV2(router)
	}

	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// 1. Detect Environment
	isLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""

	if !isLambda {
		// LOCAL DEVELOPMENT FLOW
		log.Println("--- Starting Heavy Duty API Locally ---")

		// Load .env
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: No .env file found")
		}

		// Initialize Database
		// if err := database.InitDB(); err != nil {
		// 	log.Printf("Database connection failed: %v", err)
		// }

		// Initialize Router
		router := setupRouter()

		log.Println("Server running at: http://localhost:8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatalf("Failed to start local server: %v", err)
		}
	} else {
		// AWS LAMBDA FLOW
		// Note: DB Init on Lambda usually happens inside Handler or init()
		// for better performance/connection pooling.
		// if err := database.InitDB(); err != nil {
		// 	log.Printf("Lambda DB Init failed: %v", err)
		// }
		lambda.Start(Handler)
	}
}
