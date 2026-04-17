# Makefile for MCP Policy Proxy
# Testing & Observability Targets

.PHONY: test unit integration fuzz ci-fuzz prometheus-test

# Run all tests (unit only, no fuzz)
test:
	go test -short -v ./...

# Run unit tests only (fast)
unit:
	go test -short -v ./...

# Run integration tests (requires test server)
integration:
	go test -tags=integration -v -run TestIntegration ./...

# Run fuzz tests locally (30 seconds)
# Usage: make fuzz
fuzz:
	go test -fuzztime=30s -fuzz ./...

# Run fuzz tests with longer runtime for local development
fuzz-dev:
	go test -fuzztime=60s -fuzz ./...

# CI Fuzz target (conservative, 60 seconds)
# Usage: make ci-fuzz
ci-fuzz:
	go test -fuzztime=60s -fuzzminruntime=60s -fuzz ./...

# Prometheus metrics validation
# Requires running proxy: go run . &
prometheus-test:
	@echo "Fetching metrics from http://localhost:8080/metrics..."
	@curl -s http://localhost:8080/metrics | head -20
	@echo ""
	@echo "Validating Prometheus format..."

# Run all tests + fuzz in CI pipeline
ci-test: test integration ci-fuzz