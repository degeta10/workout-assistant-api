.PHONY: swag-init run build clean deploy check-supabase db-migration-new db-push

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
	go fmt ./...
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bootstrap ./cmd/api/main.go

# Clean up build artifacts
clean:
	rm -f bootstrap

# Deploy to AWS Lambda using Serverless Framework
deploy: clean build
	npx serverless deploy