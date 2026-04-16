# =============================================================================
# Hexstrike Defense - MCP Policy Proxy Makefile
# =============================================================================
# Targets for building, testing, and Docker operations
#
# Usage:
#   make build        - Compile Go binary
#   make test        - Run unit tests
#   make docker-build - Build Docker image
#   make docker-run  - Run container locally
#   make clean       - Remove binaries
#   make lint        - Run go vet
#   make vet         - Run go vet (alias)
# =============================================================================

# Project configuration
BINARY_NAME := mcp-policy-proxy
DOCKER_IMAGE := hexstrike/mcp-policy-proxy
DOCKER_TAG := latest
GO_VERSION := 1.21

# Source and output directories
SRC_DIR := src/mcp-policy-proxy
BUILD_DIR := $(SRC_DIR)
TEST_PACKAGES := $(SRC_DIR)/...

# Docker build arguments
DOCKER_BUILD_ARGS := --build-arg VERSION=1.0.0 --build-arg COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Default target - show help
.PHONY: help
help:
	@echo "Hexstrike Defense - MCP Policy Proxy"
	@echo "================================="
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Compile Go binary"
	@echo "  make test          - Run unit tests"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run container locally"
	@echo "  make clean        - Remove binaries"
	@echo "  make lint        - Run go vet"
	@echo "  make vet         - Run go vet (alias)"
	@echo ""

# =============================================================================
# Build Target - Compile Go binary
# =============================================================================
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	cd $(SRC_DIR) && go build -o $(BINARY_NAME) .
	@echo "Build complete: $(SRC_DIR)/$(BINARY_NAME)"

# =============================================================================
# Test Target - Run unit tests
# =============================================================================
.PHONY: test
test:
	@echo "Running tests..."
	cd $(SRC_DIR) && go test -v -race -coverprofile=coverage.out $(TEST_PACKAGES)
	@echo "Tests complete"

# =============================================================================
# Docker Build Target - Build Docker image
# =============================================================================
.PHONY: docker-build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build $(DOCKER_BUILD_ARGS) -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest $(SRC_DIR)
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# =============================================================================
# Docker Run Target - Run container locally
# =============================================================================
.PHONY: docker-run
docker-run:
	@echo "Starting $(DOCKER_IMAGE) container..."
	docker run -d \
		--name $(BINARY_NAME) \
		-p 8080:8080 \
		-e LISTEN_ADDR=0.0.0.0:8080 \
		-e MCP_BACKEND_URL=http://host.docker.internal:9090 \
		-e LAKERA_API_KEY=$$LAKERA_API_KEY \
		-e LAKERA_FAIL_MODE=closed \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started. Visit http://localhost:8080/health"

# =============================================================================
# Clean Target - Remove binaries
# =============================================================================
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	cd $(SRC_DIR) && rm -f $(BINARY_NAME) coverage.out
	rm -f $(SRC_DIR)/$(BINARY_NAME).exe 2>/dev/null || true
	@echo "Clean complete"

# =============================================================================
# Lint Target - Run go vet
# =============================================================================
.PHONY: lint
lint:
	@echo "Running go vet..."
	cd $(SRC_DIR) && go vet ./...
	@echo "Lint complete"

# =============================================================================
# Vet Target - Alias for lint
# =============================================================================
.PHONY: vet
vet: lint

# =============================================================================
# All Target - Build, test, and lint
# =============================================================================
.PHONY: all
all: build test lint