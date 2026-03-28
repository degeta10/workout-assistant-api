# Workout Assistant API

Workout Assistant API is a Go-based backend service built with Gin and deployable to AWS Lambda using the Serverless Framework.

The current version includes a versioned API group and a health check endpoint. The codebase is structured to make it easy to add future workout-related routes.

## Tech Stack

- Go 1.26+
- Gin (`github.com/gin-gonic/gin`)
- AWS Lambda Go runtime (`github.com/aws/aws-lambda-go`)
- API Gateway v2 adapter (`github.com/awslabs/aws-lambda-go-api-proxy/gin`)
- Serverless Framework

## Current API

Base route group:

- `/v1`

Available endpoints:

- `GET /v1/health`
- `POST /v1/register`
- `POST /v1/login`

### Example response

```json
{
  "status": "pass",
  "version": "1.0.0",
  "release_id": "2026-03-27",
  "description": "Workout Assistant API",
  "checks": {
    "database": {
      "status": "connected",
      "component_type": "datastore",
      "time": "2026-03-27T10:00:00Z"
    }
  }
}
```

Notes:

- The timestamp fields are dynamic.
- Database health is validated using a real ping when DB is initialized.

## Project Structure

```text
.
├── cmd/api/main.go          # Composition root (config, DB init, DI wiring)
├── internal/auth/           # Feature module
│   ├── handler.go           # Delivery layer (Gin)
│   ├── service.go           # Business logic layer
│   ├── repository.go        # Persistence layer (database/sql)
│   └── models.go            # Feature models/interfaces
├── internal/health/         # Feature module
│   ├── handler.go           # Delivery layer (Gin)
│   ├── service.go           # Business logic layer
│   ├── repository.go        # Persistence layer (database/sql)
│   └── models.go            # Feature models/interfaces
├── internal/config/         # Environment config loading
├── internal/database/       # Database initialization
├── serverless.yml           # Serverless deployment config
├── Makefile                 # Build/clean/deploy commands
└── go.mod                   # Go module and dependencies
```

## Prerequisites

Install the following tools:

- Go 1.26 or newer
- Node.js + npm (required for `npx serverless`)
- AWS CLI configured with credentials and region access

Optional for local development:

- `.env` file in project root (loaded automatically if present)

## Run Locally

Start the API locally:

```bash
make run
```

Or directly:

```bash
go run ./cmd/api/main.go
```

Local server URL:

- `http://localhost:8080`

Test health endpoint:

```bash
curl http://localhost:8080/v1/health
```

Stop the server with `Ctrl+C` — it shuts down gracefully with a 5-second timeout for in-flight requests.

To run in release mode (disables Gin debug output):

```bash
APP_ENV=release make run
```

## Build for AWS Lambda

Generate the Lambda binary (`bootstrap`) for Linux ARM64:

```bash
make build
```

Clean generated binary:

```bash
make clean
```

## Deploy with Serverless

Deploy using the Makefile target:

```bash
make deploy
```

What this does:

1. Removes old `bootstrap`
2. Builds a fresh Linux ARM64 `bootstrap`
3. Runs `npx serverless deploy`

After deployment, Serverless prints the HTTP API endpoint URL.

## Make Targets

- `make run` - Run the API locally
- `make build` - Build Lambda binary (Linux ARM64)
- `make clean` - Remove Lambda binary
- `make deploy` - Clean, build, and deploy
- `make swag-init` - Regenerate Swagger docs
- `make fmt` - Format all Go source files

## Next Steps

- Add workout session routes under `/v1`
- Add JWT middleware for protected routes
- Add unit and integration tests for handlers and services
