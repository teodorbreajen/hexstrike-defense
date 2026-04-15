# HexStrike Defense Operations Guide

This guide provides deployment procedures, operational guidance, and troubleshooting for the HexStrike Defense architecture.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Health Checks](#health-checks)
4. [Common Issues](#common-issues)
5. [Rollback Procedures](#rollback-procedures)
6. [Monitoring](#monitoring)

---

## Prerequisites

### Infrastructure Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| Kubernetes | 1.28+ | 1.29+ |
| Cilium CNI | 1.14+ | 1.15+ |
| Helm | 3.12+ | 3.14+ |
| kubectl | 1.28+ | 1.29+ |

### Required Infrastructure

- [ ] Kubernetes cluster with eBPF support
- [ ] Cilium CNI installed and configured
- [ ] Cluster admin access for RBAC deployment
- [ ] Helm repository access

### Pre-Installation Checklist

```bash
# 1. Verify Kubernetes version
kubectl version --short

# 2. Verify Cilium is installed
cilium status

# 3. Verify Helm is available
helm version

# 4. Check cluster capacity
kubectl get nodes -o jsonpath='{.items[*].status.capacity}'
```

---

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/hexstrike/hexstrike-defense.git
cd hexstrike-defense
```

### Step 2: Configure Environment

```bash
# Set your namespace context
export HEXSTRIKE_NAMESPACE=hexstrike-agents
export KUBECONFIG=/path/to/cluster/kubeconfig

# Create required namespaces
kubectl create namespace hexstrike-system
kubectl create namespace hexstrike-monitoring
kubectl create namespace $HEXSTRIKE_NAMESPACE
```

### Step 3: Apply Layer 3 - Network Policies (Cilium)

**Important**: Apply network policies BEFORE other components to ensure default-deny is in place.

```bash
# Apply Cilium policies in order
kubectl apply -f manifests/cilium/00-default-deny.yaml
kubectl apply -f manifests/cilium/01-dns-whitelist.yaml
kubectl apply -f manifests/cilium/02-llm-endpoints.yaml
kubectl apply -f manifests/cilium/03-target-domains.yaml
kubectl apply -f manifests/cilium/04-hubble-enable.yaml

# Verify policies are applied
kubectl get ciliumnetworkpolicies -A
```

### Step 4: Apply Layer 4 - Runtime Detection (Falco + Talon)

```bash
# Apply Falco rules
kubectl apply -f manifests/falco/

# Verify Falco is running
kubectl get pods -n hexstrike-monitoring -l app=falco

# Verify Talon is running
kubectl get pods -n hexstrike-monitoring -l app=talon
```

### Step 5: Apply Layer 5 - Semantic Firewall (MCP Policy Proxy)

```bash
# Build the MCP Policy Proxy image
cd src/mcp-policy-proxy
docker build -t hexstrike/mcp-policy-proxy:latest .

# Apply MCP Proxy manifests
kubectl apply -f manifests/mcp-proxy/

# Verify proxy is running
kubectl get pods -n hexstrike-system -l app=mcp-proxy
```

### Step 6: Apply Observability Integration

```bash
# Apply LangGraph agent config
kubectl apply -f manifests/langgraph/

# Apply Prometheus ServiceMonitor (if Prometheus Operator is installed)
kubectl apply -f manifests/mcp-proxy/prometheus-servicemonitor.yaml
```

### Step 7: Verify Installation

```bash
# Run the validation script
./scripts/validate.sh

# Expected output:
# [PASS] Cilium policies applied
# [PASS] Falco rules loaded
# [PASS] Talon webhook configured
# [PASS] MCP Policy Proxy running
# [PASS] Observability endpoints configured
```

---

## Health Checks

### MCP Policy Proxy Health

```bash
# Check proxy health
curl -s http://mcp-proxy.hexstrike-system.svc.cluster.local:8080/health | jq

# Expected response:
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "lakera": "ok"
  }
}

# Check proxy metrics
curl -s http://mcp-proxy.hexstrike-system.svc.cluster.local:8080/metrics | jq
```

### Falco Health

```bash
# Check Falco pod status
kubectl get pods -n hexstrike-monitoring -l app=falco

# Check Falco logs
kubectl logs -n hexstrike-monitoring -l app=falco --tail=100

# Verify rules are loaded
kubectl exec -n hexstrike-monitoring -l app=falco -- falco --list
```

### Cilium Health

```bash
# Check Cilium status
cilium status

# Verify Hubble is enabled
cilium hubble status

# Check network policies
cilium policy get
```

### Talon Health

```bash
# Check Talon pod status
kubectl get pods -n hexstrike-monitoring -l app=talon

# Check Talon logs
kubectl logs -n hexstrike-monitoring -l app=talon --tail=100

# Verify webhook is registered
kubectl get mutatingwebhookconfigurations
```

---

## Common Issues

### Issue: MCP Policy Proxy Returns 502 Bad Gateway

**Symptoms**:
```json
{"error": {"code": -32603, "message": "MCP backend unavailable"}}
```

**Diagnosis**:
```bash
# Check if MCP backend is running
kubectl get pods -n hexstrike-system -l app=mcp-backend

# Check proxy logs
kubectl logs -n hexstrike-system -l app=mcp-proxy --tail=50

# Verify service endpoints
kubectl get endpoints -n hexstrike-system -l app=mcp-backend
```

**Resolution**:
1. Ensure MCP backend is running
2. Check network connectivity
3. Verify `MCP_BACKEND_URL` environment variable

---

### Issue: Falco Not Detecting Shell Spawns

**Symptoms**: Shell commands execute without Falco alerts

**Diagnosis**:
```bash
# Check Falco is running with correct rules
kubectl exec -n hexstrike-monitoring -l app=falco -- falco --list | grep shell

# Check Falco pod has correct permissions
kubectl auth can-i exec pods -n hexstrike-agents --as=system:serviceaccount:hexstrike-monitoring:falco
```

**Resolution**:
1. Ensure Falco rules are loaded: `manifests/falco/01-execve-rules.yaml`
2. Verify Falco has access to target namespaces
3. Check Falco logs for rule loading errors

---

### Issue: Cilium Policies Not Blocking Traffic

**Symptoms**: Pods can reach unauthorized endpoints

**Diagnosis**:
```bash
# Check if policies are applied
kubectl get ciliumnetworkpolicies -A

# Check Cilium endpoint status
kubectl get cep -A

# Verify Hubble sees the traffic
hubble observe --type drop
```

**Resolution**:
1. Ensure Cilium is the active CNI
2. Verify policies are applied to correct endpoints
3. Check for conflicting policies

---

### Issue: Talon Not Responding to Alerts

**Symptoms**: CRITICAL Falco alerts don't trigger pod termination

**Diagnosis**:
```bash
# Check Talon pod status
kubectl get pods -n hexstrike-monitoring -l app=talon

# Verify webhook is configured
kubectl get mutatingwebhookconfigurations talon-webhook

# Check Talon logs for webhook events
kubectl logs -n hexstrike-monitoring -l app=talon | grep -i webhook
```

**Resolution**:
1. Ensure Talon webhook is registered
2. Check Falco is sending events to Talon
3. Verify Talon has RBAC permissions for pod deletion

---

### Issue: Lakera API Timeout Errors

**Symptoms**: Proxy logs show Lakera timeout warnings

**Diagnosis**:
```bash
# Check proxy logs
kubectl logs -n hexstrike-system -l app=mcp-proxy | grep -i lakera

# Test Lakera connectivity manually
curl -X POST https://api.lakera.ai/v1/guard/check \
  -H "Authorization: Bearer $LAKERA_API_KEY" \
  -d '{"tool_name": "test", "arguments": "test"}'
```

**Resolution**:
1. Verify `LAKERA_API_KEY` is set in ConfigMap
2. Check network policy allows outbound to `api.lakera.ai`
3. Increase `LAKERA_TIMEOUT` if needed (default: 5s)

---

### Issue: Rate Limiting Blocking Legitimate Requests

**Symptoms**: MCP client receives 429 Too Many Requests

**Diagnosis**:
```bash
# Check current rate limit configuration
kubectl get configmap -n hexstrike-system mcp-proxy-config -o yaml

# Check metrics for rate limit hits
curl -s http://mcp-proxy:8080/metrics | grep rate_limit
```

**Resolution**:
1. Increase `RATE_LIMIT_PER_MINUTE` in ConfigMap
2. Implement token bucket on client side
3. Use separate rate limits per client ID

---

## Rollback Procedures

### Quick Rollback (Feature Flags)

Each layer can be disabled independently:

```bash
# Disable Lakera semantic check (Layer 5)
kubectl set env deployment/mcp-proxy -n hexstrike-system LAKERA_API_KEY=""

# Disable Talon automated response (Layer 4)
kubectl scale deployment/talon -n hexstrike-monitoring --replicas=0

# Disable specific Cilium policy
kubectl delete -f manifests/cilium/02-llm-endpoints.yaml
```

### Full Rollback to Previous Version

```bash
# List recent deployments
kubectl rollout history deployment/mcp-proxy -n hexstrike-system

# Rollback MCP Proxy
kubectl rollout undo deployment/mcp-proxy -n hexstrike-system

# Rollback Falco
kubectl rollout undo daemonset/falco -n hexstrike-monitoring

# Verify rollback
kubectl rollout status deployment/mcp-proxy -n hexstrike-system
```

### Complete Component Removal

```bash
# Remove MCP Proxy
kubectl delete -f manifests/mcp-proxy/

# Remove Falco and Talon
kubectl delete -f manifests/falco/

# Remove Cilium policies (WARNING: May affect other workloads)
kubectl delete -f manifests/cilium/

# Remove namespaces (optional)
kubectl delete namespace hexstrike-system hexstrike-monitoring hexstrike-agents
```

---

## Monitoring

### Prometheus Metrics

Access via Prometheus or Grafana:

| Metric | Description |
|--------|-------------|
| `mcp_proxy_requests_total` | Total requests processed |
| `mcp_proxy_blocked_total` | Requests blocked by semantic firewall |
| `mcp_proxy_lakera_blocks` | Lakera blocked requests |
| `mcp_proxy_rate_limit_hits` | Rate limit rejections |
| `mcp_proxy_latency_seconds` | Request latency histogram |

### Alerting Rules

Recommended Prometheus alerts:

```yaml
groups:
  - name: hexstrike-alerts
    rules:
      - alert: MCPProxyDown
        expr: up{job="mcp-proxy"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "MCP Policy Proxy is down"
      
      - alert: HighBlockRate
        expr: rate(mcp_proxy_blocked_total[5m]) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of blocked requests"
      
      - alert: FalcoCriticalAlert
        expr: falco_events_total{priority="CRITICAL"} > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Falco CRITICAL alert detected"
```

### Hubble Flow Monitoring

```bash
# Watch all flows from hexstrike namespace
hubble observe -n hexstrike-agents --type drop

# Watch successful connections
hubble observe -n hexstrike-agents --type trace

# Export flow logs for analysis
hubble observe -n hexstrike-agents --output json > flows.json
```

---

## Support

For issues or questions:

1. Check this guide for common solutions
2. Review [ARCHITECTURE.md](./ARCHITECTURE.md) for design details
3. Review [SECURITY.md](./SECURITY.md) for hardening guidance
4. Open an issue at: https://github.com/hexstrike/hexstrike-defense/issues
