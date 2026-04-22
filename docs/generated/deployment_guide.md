# Deployment Guide

## Prerequisites

| Requirement | Version |
|-------------|----------|
| Go | 1.21+ |
| Docker | Latest |
| Kubernetes | 1.24+ |
| kubectl | Latest |
| helm | 3.10+ |

## Build

```bash
# Build the proxy binary
make build

# Output: src/mcp-policy-proxy/mcp-policy-proxy
```

## Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## Kubernetes

```bash
# Deploy to Kubernetes
./scripts/deploy.sh

# Validate deployment
./scripts/validate.sh
```

### Namespace Structure

- `hexstrike-system` - MCP proxy and configs
- `hexstrike-agents` - Agent workloads
- `hexstrike-monitoring` - Falco, Hubble, Sentry

### Resource Requirements

| Component | CPU | Memory |
|-----------|-----|-------|
| MCP Proxy | 100m-500m | 128Mi-512Mi |

### Health Checks

- Liveness: `/health` endpoint
- Readiness: `/ready` endpoint
- Metrics: `/metrics` (Prometheus)

---

*Generated from scripts and manifests*
