.PHONY: swag-init run build clean deploy

# Generate Swagger documentation
swag-init:
	swag init -g cmd/api/main.go

# Format all code and organize imports
fmt:
	go fmt ./...

# Run the API locally
run:
	go run cmd/api/main.go

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