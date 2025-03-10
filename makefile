# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Application parameters
BINARY_NAME=simple-go-api
MAIN_PATH=./cmd/server/main.go
BUILD_DIR=./build

# Docker parameters
DOCKER_IMAGE=simple-go-api
DOCKER_TAG=latest
DOCKER_BUILD=docker build

# Set environment variables
export GO111MODULE=on

.PHONY: all build clean run test cover lint vet tidy docker-build docker-run air-run

all: test build

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run the application
run:
	$(GORUN) $(MAIN_PATH)

# Run with air for live reloading
air-run:
	air

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Lint the code
lint:
	$(GOLINT) run ./...

# Run go vet
vet:
	$(GOVET) ./...

# Update dependencies
tidy:
	$(GOMOD) tidy

# Build docker image
docker-build:
	$(DOCKER_BUILD) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run docker container
docker-run:
	docker run --rm -p 8080:8080 \
		-e PORT=8080 \
		-e AUTH_SERVICE_URL=https://api.product.dev.alertlogic.com \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Install dependencies for development
setup-dev:
	$(GOGET) github.com/air-verse/air
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate mock files (requires mockgen)
# TODO: Implement mockgen or alternative
# mocks:
	# mockgen -source=internal/auth/service.go -destination=internal/auth/mocks/service_mock.go -package=mocks

# Show help
help:
	@echo "make - Runs tests then builds the application"
	@echo "make build - Build the application"
	@echo "make clean - Remove build artifacts"
	@echo "make run - Run the application"
	@echo "make air-run - Run the application with live reloading"
	@echo "make test - Run tests"
	@echo "make cover - Run tests with coverage report"
	@echo "make lint - Run linter"
	@echo "make vet - Run go vet"
	@echo "make tidy - Tidy dependencies"
	@echo "make docker-build - Build docker image"
	@echo "make docker-run - Run docker container"
	@echo "make setup-dev - Install development dependencies"
	@echo "make mocks - Generate mock files"