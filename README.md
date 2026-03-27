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

Available endpoint:

- `GET /v1/health`

### Example response

```json
{
  "status": "pass",
  "version": "1.0.0",
  "release_id": "2026-03-27",
  "description": "Heavy Duty Workout API",
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
- Database health is currently mocked as connected in code.

## Project Structure

```text
.
├── main.go          # App setup, local server entrypoint, Lambda handler
├── serverless.yml   # Serverless deployment config
├── Makefile         # Build/clean/deploy commands
├── go.mod           # Go module and dependencies
└── bootstrap        # Compiled Linux ARM64 binary for Lambda custom runtime
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
go run main.go
```

Local server URL:

- `http://localhost:8080`

Test health endpoint:

```bash
curl http://localhost:8080/v1/health
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

## AWS Routing Notes

- The Lambda handler trims `/dev` and `/prod` stage prefixes from incoming paths.
- This allows routes like `/dev/v1/health` or `/prod/v1/health` to map correctly to `/v1/health` inside Gin.

## Make Targets

- `make build` - Build Lambda binary
- `make clean` - Remove Lambda binary
- `make deploy` - Clean, build, and deploy

## Next Steps

- Add workout session routes under `/v1`
- Integrate and validate real database health checks
- Add tests for handlers and Lambda adapter behavior
