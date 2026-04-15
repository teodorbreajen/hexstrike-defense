# HexStrike Defense E2E Tests

This directory contains end-to-end tests for the HexStrike Defense architecture.

## Structure

```
tests/e2e/
├── test_semantic_firewall.go  # MCP Policy Proxy tests
├── test_falco_detection.go     # Runtime detection tests
├── test_cilium_policies.go    # Network policy tests
├── framework/
│   ├── cluster.go              # Kubernetes client utilities
│   └── utils.go               # HTTP client utilities
└── go.mod                      # Test dependencies
```

## Prerequisites

1. A running Kubernetes cluster with Cilium CNI
2. HexStrike Defense components deployed
3. `kubectl` configured with cluster access

## Running Tests

### Run All Tests

```bash
cd tests/e2e
go mod tidy
go test -v ./...
```

### Run Specific Test Suites

```bash
# Semantic Firewall tests
go test -v -run TestSemanticFirewall

# Falco Detection tests
go test -v -run TestFalcoDetection

# Cilium Policy tests
go test -v -run TestCiliumPolicies
```

### Run Against Specific Cluster

```bash
export KUBECONFIG=/path/to/kubeconfig
export MCP_PROXY_URL=http://mcp-proxy.hexstrike-system.svc.cluster.local:8080
export HEXSTRIKE_NAMESPACE=hexstrike-agents

go test -v ./...
```

### Skip Long-Running Tests

```bash
go test -v -short ./...
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KUBECONFIG` | `~/.kube/config` | Path to kubeconfig |
| `KUBE_CONTEXT` | (none) | Kubernetes context to use |
| `MCP_PROXY_URL` | `http://localhost:8080` | MCP Policy Proxy URL |
| `HEXSTRIKE_NAMESPACE` | `hexstrike-agents` | Agent namespace |

## Test Categories

### Semantic Firewall Tests

- Valid JSON-RPC request passes through
- Malformed JSON-RPC is rejected
- Malicious prompt injection is blocked
- Rate limiting works
- Lakera timeout handling

### Falco Detection Tests

- Shell spawn triggers alert
- /etc write triggers alert
- Talon terminates pod (requires controlled environment)
- False positives don't occur

### Cilium Policy Tests

- Default-deny blocks unauthorized traffic
- DNS whitelist allows DNS
- LLM endpoints are accessible
- Target domains are accessible
- Hubble logs show drops

## Writing New Tests

Follow the existing patterns in `test_*.go` files:

```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/hexstrike/hexstrike-defense/tests/e2e/framework"
)

func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping e2e test in short mode")
    }

    // Get Kubernetes client
    k8sClient, err := framework.NewClient(nil)
    if err != nil {
        t.Skipf("Skipping test - no Kubernetes cluster: %v", err)
    }

    ctx := context.Background()
    
    // Your test code here
}
```

## Troubleshooting

### "Skipping test - no Kubernetes cluster"

Ensure:
1. `kubectl` is configured with cluster access
2. Kubeconfig path is set via `KUBECONFIG` env var

### "Failed to create REST config"

Ensure:
1. Cluster is accessible
2. Credentials are valid
3. `kubectl get nodes` works

### Tests pass locally but fail in CI

- Check environment variables are set in CI
- Ensure cluster has required components deployed
- Verify network connectivity from CI runner to cluster
