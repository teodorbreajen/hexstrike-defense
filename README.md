# hexstrike-defense

![CI Tests](https://github.com/teodorbreajen/hexstrike-defense/workflows/CI%20Tests/badge.svg)
![SDD Validate](https://github.com/teodorbreajen/hexstrike-defense/workflows/SDD%20Validate/badge.svg)
![Security](https://github.com/teodorbreajen/hexstrike-defense/workflows/Security%20Hardening/badge.svg)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Defense-in-Depth architecture for securing autonomous AI agents (hexstrike-ai).

## Version

**Current**: 2.0.0 (Security-Hardened Release)

See `CHANGELOG.md` for full version history.

## Overview

This project implements a multi-layer security architecture to protect autonomous AI agents from malicious inputs, runtime threats, and network-based attacks.

## Architecture Layers

1. **SDD Governance** - Spec-driven development ensures security requirements are captured first
2. **Semantic Firewall** - Lakera Guard / NeMo Guardrails for input validation
3. **Runtime Detection** - Falco + eBPF + Talon for behavioral monitoring
4. **Network Containment** - Cilium CNI for zero-trust networking

## Quick Start

```bash
# 1. Build the proxy
make build

# 2. Run tests
make test

# 3. Deploy to Kubernetes
./scripts/deploy.sh
```

## Tech Stack

- **Orchestration**: LangGraph
- **Protocol**: MCP (Model Context Protocol)
- **Runtime Security**: Falco, eBPF, Talon
- **Network Policies**: Cilium CNI
- **Observability**: Atlassian MCP, Sentry MCP
- **Infrastructure**: Kubernetes (kind/minikube for dev)

## Project Structure

```
hexstrike-defense/
├── src/mcp-policy-proxy/    # Go semantic firewall proxy
│   ├── proxy.go            # Main proxy implementation
│   ├── logger.go          # Structured JSON logging
│   ├── Dockerfile        # Multi-stage Docker build
│   └── *_test.go          # Unit tests (39 tests)
├── manifests/             # Kubernetes manifests
│   ├── cilium/           # Network policies
│   ├── falco/            # Runtime detection rules
│   ├── mcp-proxy/        # Proxy deployment
│   └── langgraph/         # Agent configs
├── scripts/               # Automation
├── docs/                 # Documentation
│   ├── runbook.md        # Incident response
│   ├── PHASE3-HARDENING-REPORT.md
│   └── TEST-RESULTS.md
├── openspec/              # SDD governance
├── Makefile              # Build automation
└── CHANGELOG.md          # Version history
```

## Configuration

### Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `LISTEN_ADDR` | `:8080` | Listen address |
| `MCP_BACKEND_URL` | - | Yes | MCP backend URL |
| `LAKERA_API_URL` | - | Yes | Lakera API URL |
| `LAKERA_API_KEY` | - | Yes | Lakera API key |
| `RATE_LIMIT_PER_MINUTE` | 60 | No | Rate limit |
| `LAKERA_FAIL_MODE` | `closed` | No | Fail mode (closed/open) |
| `MAX_BODY_SIZE` | 1048576 | No | Max body bytes |
| `JWT_SECRET` | - | No | JWT validation secret |

### Fail Mode

- `closed` (default): Block requests when Lakera fails (SECURE)
- `open`: Allow requests when Lakera fails (BACKWARD COMPATIBLE)

### Protected Endpoints

| Endpoint | Auth Required |
|----------|--------------|
| `/health` | No |
| `/metrics` | No |
| `/ready` | No |
| `/mcp/*` | Yes (Bearer JWT) |

## Running

### Local Development

```bash
# Build
make build

# Run tests
make test

# Run the proxy
MCP_BACKEND_URL=http://localhost:9090 \
LAKERA_API_URL=https://api.lakera.ai \
LAKERA_API_KEY=your-key \
./mcp-policy-proxy
```

### Docker

```bash
# Build image
make docker-build

# Run container
make docker-run
```

### Documentation Generator

```bash
# Generate all documentation
python docs/tools/doc_engine/main.py

# With verbose output
python docs/tools/doc_engine/main.py --verbose

# Custom output directory
python docs/tools/doc_engine/main.py --output ./my-docs
```

### Kubernetes

```bash
# Deploy
./scripts/deploy.sh

# Validate
./scripts/validate.sh
```

## Testing

### Unit Tests (39+ tests)

```bash
cd src/mcp-policy-proxy && go test -v ./...
```

### Security Testing

The project includes automated security scanning via GitHub Actions:

- **CodeQL**: Static analysis
- **Gosec**: Go security scanner
- **Semgrep**: SAST scanning
- **Trivy**: Container vulnerability scanning
- **Govulncheck**: Dependency vulnerability scanning

Results documented in `docs/TEST-RESULTS.md`.

### E2E Tests

Requires Kubernetes cluster:

```bash
cd tests/e2e && go test -v ./...
```

## Security

### Security Features (v2.0.0)

#### Authentication & Access Control
- **JWT Authentication**: Bearer token validation required (HS256/384/512 only)
- **Algorithm Confusion Protection**: Blocks alg:none and asymmetric algorithm attacks

#### Input Validation
- **Fail-Closed**: Block requests when Lakera is unavailable
- **Body Size Limit**: 1MB max (configurable)
- **Input Sanitization**: SSRF, SQL injection, command injection detection
- **Path Traversal Protection**: Blocks ../ and encoded variants

#### Rate Limiting & DoS Protection
- **Per-Client Rate Limiting**: Token bucket per client IP
- **Concurrent Request Limiting**: Max 100 concurrent requests
- **Batch Request Limits**: Max 10 requests per batch

#### Resilience
- **Circuit Breaker**: Prevents cascade failures
- **Retry with Exponential Backoff**: 1s, 2s, 4s retry strategy
- **Connection Pooling**: Reusable HTTP connections
- **Dead Letter Queue (DLQ)**: Failed requests stored for replay

#### Observability
- **Structured Logging**: JSON for SIEM integration
- **Correlation IDs**: Request tracing with UUID v4
- **Prometheus Metrics**: Endpoint at /metrics

### Error Codes

| Code | Meaning |
|------|----------|
| 200 | Success |
| 400 | Bad request |
| 401 | Unauthorized (no/invalid JWT) |
| 403 | Blocked by Lakera |
| 413 | Request body too large |
| 429 | Rate limited |
| 502 | Backend unavailable |
| 503 | Security service unavailable |

## Documentation

### Auto-Generated (doc_engine)

Documentation is automatically generated from code analysis:

```bash
# Generate documentation
python docs/tools/doc_engine/main.py

# Output: docs/generated/*.md
```

Generated sections include:
- **Component Catalog** - All Go files, functions, types
- **Architecture Diagrams** - Mermaid flowcharts
- **Configuration Reference** - Environment variables
- **Interfaces** - HTTP endpoints, component interfaces
- **Security Model** - Headers, auth, rate limits

### Manual Documentation

- `CHANGELOG.md` - Version history
- `docs/PHASE3-HARDENING-REPORT.md` - Technical report
- `docs/TEST-RESULTS.md` - Test execution results
- `docs/runbook.md` - Incident response

## Requirements

- Go 1.21+
- Docker (optional)
- Kubernetes (for deployment)
- kubectl
- helm

## License

Proprietary - HexStrike Defense