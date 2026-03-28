.PHONY: swag-init run build clean deploy fmt check-supabase db-migration-new db-diff db-push

SUPABASE ?= supabase

# Generate Swagger documentation
swag-init:
	swag init -g cmd/api/main.go

# Format all code and organize imports
fmt:
	go fmt ./...

# Run the API locally
run:
	go run cmd/api/main.go

# Create a new migration file: make db-migration-new name=create_users
db-migration-new: check-supabase
	@if [ -z "$(name)" ]; then echo "Usage: make db-migration-new name=<migration_name>"; exit 1; fi
	$(SUPABASE) migration new $(name)

# Generate a migration file from schema diff: make db-diff name=sync_schema
db-diff: check-supabase
	@if [ -z "$(name)" ]; then echo "Usage: make db-diff name=<migration_name>"; exit 1; fi
	$(SUPABASE) db diff --use-migra -f $(name)

# Apply pending migrations to the linked Supabase project
db-push: check-supabase
	$(SUPABASE) db push

# Ensure Supabase CLI is installed before running DB tasks
check-supabase:
	@command -v $(SUPABASE) >/dev/null 2>&1 || { \
		echo "Supabase CLI not found."; \
		echo "Install on macOS: brew install supabase/tap/supabase"; \
		echo "Then run: supabase login"; \
		exit 1; \
	}

# Build the API for AWS Lambda
build:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bootstrap ./cmd/api/main.go

# Clean up build artifacts
clean:
	rm -f bootstrap

# Deploy to AWS Lambda (Development Environment)
deploy-dev: clean build
	npx serverless deploy --stage dev

# Deploy to AWS Lambda (Production Environment)
deploy-prod: clean build
	@echo "⚠️ WARNING: Deploying to PRODUCTION..."
	npx serverless deploy --stage prod