# Workout Assistant API

Workout Assistant API is a Go-based backend service built with Gin and deployable to AWS Lambda using the Serverless Framework.

The current version exposes a versioned API with health and auth endpoints. The codebase follows a feature-based modular structure so new domains can be added cleanly.

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
- `GET /v1/me` (requires `Authorization: Bearer <token>`)

## Response Format

All APIs use a unified response envelope:

```json
{
  "success": true,
  "message": "...",
  "data": {},
  "errors": {}
}
```

Notes:

- `data` is optional and omitted when empty.
- `errors` is populated for validation and failure responses.

### Health Response Example

```json
{
  "success": true,
  "message": "Health check successful",
  "data": {
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
}
```

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
├── internal/pkg/responses/  # Shared API response helpers
├── serverless.yml           # Serverless deployment config
├── Makefile                 # Build/clean/deploy commands
└── go.mod                   # Go module and dependencies
```

## Runtime Modes

- Local mode: starts Gin HTTP server on `APP_PORT`.
- Lambda mode: auto-activated when `AWS_LAMBDA_FUNCTION_NAME` is present.

## Prerequisites

Install the following tools:

- Go 1.26 or newer
- Node.js + npm (required for `npx serverless`)
- AWS CLI configured with credentials and region access

Optional for local development:

- `.env` file in project root (loaded automatically if present)

## Environment Variables

Core variables:

- `APP_NAME`
- `APP_VERSION`
- `APP_PORT`
- `APP_ENV` (`prod` or `debug`)
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `JWT_SECRET`

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

To run in release mode (disables Gin debug logs):

```bash
APP_ENV=prod make run
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

## Supabase Notes

- For local direct connection, Supabase commonly uses host `db.<project-ref>.supabase.co` on port `5432`.
- For Lambda, use the connection details appropriate for your project/network setup.
- If Lambda cannot reach direct host networking, use the Supabase pooler endpoint and credentials.

## Database Migrations (Supabase CLI)

This project now keeps SQL migrations in `supabase/migrations/`.

First-time setup:

```bash
supabase login
supabase link --project-ref <your-project-ref>
```

Common workflow:

```bash
# create an empty migration file
make db-migration-new name=add_workouts_table

# generate a diff migration file
make db-diff name=sync_schema

# apply pending migrations to linked project
make db-push
```

## Make Targets

- `make run` - Run the API locally
- `make build` - Build Lambda binary (Linux ARM64)
- `make clean` - Remove Lambda binary
- `make deploy` - Clean, build, and deploy
- `make swag-init` - Regenerate Swagger docs
- `make fmt` - Format all Go source files
- `make db-migration-new name=<migration_name>` - Create a migration file
- `make db-diff name=<migration_name>` - Generate migration from schema diff
- `make db-push` - Apply pending migrations to linked Supabase project

## Next Steps

- Add workout session routes under `/v1`
- Expand and document JWT-protected routes using the existing middleware
- Add unit and integration tests for handlers and services
