# Stack Service Makefile
.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down lint format

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  test        - Run comprehensive test suite"
	@echo "  test-unit   - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  test-quick  - Run quick tests without coverage"
	@echo "  test-race   - Run tests with race detection"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run with Docker Compose"
	@echo "  migrate-up  - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  lint        - Run linter"
	@echo "  format      - Format code"
	@echo "  deps        - Install dependencies"
	@echo "  gen-swagger - Generate Swagger documentation"

# Application variables
APP_NAME = stack_service
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/main.go

# Build for production
build-prod:
	@echo "Building $(APP_NAME) for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/main.go

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run comprehensive test suite
test:
	@echo "Running comprehensive test suite..."
	./test/run_tests.sh

# Run only unit tests
test-unit:
	@echo "Running unit tests..."
	./test/run_tests.sh --unit-only

# Run only integration tests
test-integration:
	@echo "Running integration tests..."
	./test/run_tests.sh --integration-only

# Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	./test/run_tests.sh --verbose

# Run quick tests without coverage
test-quick:
	@echo "Running quick tests..."
	go test -short ./... -v

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./... -v

# Run tests in watch mode
test-watch:
	@echo "Running tests in watch mode..."
	find . -name "*.go" | entr -r go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Run linter and fix issues
lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

# Generate Swagger documentation
gen-swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/main.go -o docs/swagger

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Running database migrations down..."
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create:
	@echo "Creating new migration: $(name)"
	migrate create -ext sql -dir migrations -seq $(name)

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker build -t $(APP_NAME):latest .

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Development environment
dev-setup: deps docker-run migrate-up
	@echo "Development environment setup complete!"

# Start development environment
dev-start:
	@echo "Starting development environment..."
	docker-compose up -d db redis
	@echo "Waiting for services to be ready..."
	sleep 5
	@$(MAKE) migrate-up
	@echo "Services are ready! You can now run 'make run'"

# Stop development environment
dev-stop:
	@echo "Stopping development environment..."
	docker-compose down

# Reset development environment
dev-reset: dev-stop
	@echo "Resetting development environment..."
	docker-compose down -v
	@$(MAKE) dev-start

# Production targets
prod-build: clean build-prod docker-build
	@echo "Production build complete!"

# Security scan
security-scan:
	@echo "Running security scan..."
	gosec ./...

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Performance benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Load test (requires hey to be installed)
load-test:
	@echo "Running load test..."
	hey -n 1000 -c 10 http://localhost:8080/health

# Database operations
db-seed:
	@echo "Seeding database with sample data..."
	go run scripts/seed.go

db-reset: migrate-down migrate-up
	@echo "Database reset complete!"

# Kubernetes targets
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deployments/k8s/

k8s-delete:
	@echo "Removing from Kubernetes..."
	kubectl delete -f deployments/k8s/

# Monitoring
monitor-start:
	@echo "Starting monitoring stack..."
	docker-compose --profile monitoring up -d

monitor-stop:
	@echo "Stopping monitoring stack..."
	docker-compose --profile monitoring down

# Admin tools
admin-start:
	@echo "Starting admin tools..."
	docker-compose --profile admin up -d

admin-stop:
	@echo "Stopping admin tools..."
	docker-compose --profile admin down