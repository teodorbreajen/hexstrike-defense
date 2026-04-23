# Deployment Guide

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Latest stable |
| Docker | Latest | For container builds |
| Kubernetes | 1.24+ | K8s cluster with worker nodes |
| kubectl | Latest | Kubernetes CLI |
| Helm | 3.10+ | Package manager |
| Cilium | Latest | CNI plugin |

## Build

```bash
# Clone and build
git clone https://github.com/hexstrike/defense.git
cd defense
make build

# Output: src/mcp-policy-proxy/mcp-policy-proxy
```

## Docker Build

```bash
# Build container image
make docker-build

# Push to registry
make docker-push REGISTRY=your-registry

# Run locally
make docker-run
```

## Kubernetes Deployment

```bash
# Deploy to cluster
./scripts/deploy.sh

# Verify deployment
kubectl get pods -n hexstrike-system
./scripts/validate.sh
```

## Namespace Structure

| Namespace | Components | Purpose |
|-----------|------------|---------|
| `hexstrike-system` | MCP Proxy, ConfigMaps | Core proxy |
| `hexstrike-agents` | Agent workloads | Agent pods |
| `hexstrike-monitoring` | Falco, Hubble, Metrics | Monitoring |

## Resource Requirements

| Component | CPU Request | CPU Limit | Memory |
|-----------|------------|----------|--------|
| MCP Proxy | 100m | 500m | 128Mi |
| Redis | 50m | 200m | 64Mi |

## Health Checks

| Endpoint | Purpose | Auth Required |
|---------|---------|--------------|
| `/health` | Liveness probe | No |
| `/ready` | Readiness probe | No |
| `/metrics` | Prometheus metrics | No |

## Network Policies

```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: mcp-proxy-policy
spec:
  endpointSelector:
    matchLabels:
      app: mcp-proxy
  ingress:
    - fromEndpoints:
        - matchLabels:
            k8s:io= kubernetes
```

---

*Generated from manifests and scripts*
