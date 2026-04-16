# HexStrike Defense MCP Policy Proxy - Runbook

This runbook provides incident response procedures and troubleshooting guidance for the HexStrike Defense MCP Policy Proxy.

**Document Version**: 1.0.0  
**Last Updated**: 2026-04-16  
**Namespace**: `hexstrike-system`  
**Service Name**: `mcp-policy-proxy`

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Lakera Outage Response](#1-lakera-outage-response)
3. [Falco Alert Response](#2-falco-alert-response)
4. [Rate Limit Exceeded Response](#3-rate-limit-exceeded-response)
5. [General Troubleshooting](#4-general-troubleshooting)
6. [Configuration Reference](#5-configuration-reference)

---

## Quick Reference

### Endpoints

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/health` | GET | Health check with Lakera status | No |
| `/ready` | GET | Readiness probe | No |
| `/metrics` | GET | Prometheus metrics | No |
| `/mcp/*` | POST | MCP tool call proxy | Yes (JWT) |

### Error Codes

| HTTP Status | Code | Meaning |
|-------------|------|---------|
| `200` | - | Request allowed |
| `400` | - | Bad request / Invalid JSON-RPC |
| `401` | - | Missing or invalid JWT token |
| `403` | - | Request blocked by semantic firewall |
| `429` | - | Rate limit exceeded |
| `502` | - | MCP backend unavailable |
| `503` | - | Lakera service unavailable (fail-closed mode) |
| `413` | - | Request body too large |

### Service Discovery

```bash
# Internal cluster access
http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080

# From same namespace
http://mcp-policy-proxy:8080
```

---

## 1. Lakera Outage Response

### Scenario 1.1: Lakera API Unavailable

**Symptom**: HTTP 503 errors, logs show "Semantic check unavailable"

**Detection**:

```
# Check health endpoint
curl -s http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080/health | jq

# Expected degraded response when Lakera is down:
{
  "status": "degraded",
  "timestamp": "2026-04-16T10:30:00Z",
  "checks": {
    "lakera": "unavailable: context deadline exceeded"
  }
}
```

**Response Procedure**:

```gherkin
GIVEN a 503 error from MCP Policy Proxy
WHEN "Semantic check unavailable - request blocked for security" message appears
THEN follow this procedure:

  1. VERIFY the outage:
     - Check Lakera status page: https://status.lakera.ai
     - Check internal monitoring dashboards

  2. CHECK network connectivity:
     kubectl exec -it -n hexstrike-system deploy/mcp-policy-proxy -- \
       curl -v https://api.lakera.ai/v1/health

  3. DETERMINE fail mode:
     kubectl get configmap -n hexstrike-system mcp-proxy-config -o jsonpath='{.data.LAKERA_FAIL_MODE}'

     - If "closed": Requests are BLOCKED until Lakera recovers
     - If "open": Requests are ALLOWED (reduced security)

  4. DECIDE action based on fail mode and business requirements:

     OPTION A: Switch to fail-open (temporarily reduce security)
       kubectl set env deployment/mcp-policy-proxy -n hexstrike-system LAKERA_FAIL_MODE=open
       kubectl rollout status deployment/mcp-policy-proxy -n hexstrike-system

     OPTION B: Wait for recovery (maintain security posture)
       Monitor until Lakera recovers
       Logs: kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --tail=100 | grep Lakera

  5. VERIFY resolution:
     curl -s http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080/health | jq
     # Expected: "lakera": "ok"
```

### Scenario 1.2: Lakera API Timeout

**Symptom**: High latency, logs show "context deadline exceeded"

**Detection**:

```bash
# Check proxy logs for timeouts
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --since=5m | grep -E "(timeout|Lakera)"

# Check metrics for increased latency
curl -s http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080/metrics | jq
```

**Response Procedure**:

```gherkin
GIVEN elevated latency or timeout errors in proxy logs
WHEN Lakera is responding but slowly
THEN follow this procedure:

  1. CHECK current timeout configuration:
     kubectl get configmap -n hexstrike-system mcp-proxy-config -o jsonpath='{.data.LAKERA_TIMEOUT}'
     # Default: 5 seconds

  2. INCREASE timeout if business requirements allow:
     kubectl patch configmap -n hexstrike-system mcp-proxy-config \
       -p '{"data":{"LAKERA_TIMEOUT":"10"}}'

  3. RESTART proxy to apply changes:
     kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system

  4. MONITOR for improvement:
     watch -n 5 'curl -s http://mcp-policy-proxy:8080/metrics | jq .avg_latency_ms'
```

---

## 2. Falco Alert Response

### Scenario 2.1: CRITICAL Alert - Shell Spawn Detected

**Symptom**: CRITICAL priority Falco alert, Talon may terminate pod

**Detection**:

```bash
# View recent Falco alerts
kubectl logs -n hexstrike-monitoring -l app=falco --since=5m | grep -i critical

# Check for pod termination events
kubectl get events -n hexstrike-agents --sort-by='.lastTimestamp' | tail -20
```

**Response Procedure**:

```gherkin
GIVEN a CRITICAL Falco alert (e.g., "Terminal shell spawn from container")
WHEN Talon has terminated the pod
THEN follow this procedure:

  1. IDENTIFY affected pod:
     kubectl get events -n hexstrike-agents --field-selector reason=TalonAction | tail -5

  2. CHECK what triggered the alert:
     kubectl logs -n hexstrike-monitoring -l app=falco --since=5m | grep -A5 "shell spawn"

  3. REVIEW pod logs before termination:
     # Talon captures logs before deletion - check Talon logs
     kubectl logs -n hexstrike-monitoring -l app=talon --since=5m | grep -A20 "captured"

  4. DETERMINE if legitimate or security incident:
     - Check if shell spawn was expected (maintenance window)
     - Review parent process in Falco output
     - Check command line arguments for malicious patterns

  5. IF legitimate activity:
     - Add run_as_user annotation to prevent future alerts
     - Document exception in change management

  6. IF security incident:
     - Preserve forensic evidence
     - Escalate to security team
     - Review audit logs for scope of compromise
```

### Scenario 2.2: WARNING Alert - Unauthorized Shell Spawn

**Symptom**: WARNING priority alert, pod has been quarantined (scaled to 0)

**Detection**:

```bash
# Check for WARNING alerts
kubectl logs -n hexstrike-monitoring -l app=falco --since=5m | grep -i warning

# Check for quarantined pods
kubectl get pods -n hexstrike-agents -l security.hexstrike.io/quarantined=true
```

**Response Procedure**:

```gherkin
GIVEN a WARNING Falco alert
WHEN pod has been quarantined (scaled to 0)
THEN follow this procedure:

  1. IDENTIFY the quarantined pod:
     kubectl get pods -n hexstrike-agents -o json | jq '.items[] | select(.metadata.labels."security.hexstrike.io/quarantined"=="true") | {name: .metadata.name, reason: .metadata.labels."security.hexstrike.io/quarantine-reason"}'

  2. REVIEW alert details:
     kubectl logs -n hexstrike-monitoring -l app=falco --since=5m | grep -B2 -A5 "Unauthorized shell spawn"

  3. INVESTIGATE the cause:
     - Was this an authorized activity?
     - Is the agent functioning correctly?
     - Any signs of compromise?

  4. DECIDE on action:

     OPTION A: Restore pod (if safe)
       # Remove quarantine label
       kubectl label pod -n hexstrike-agents <pod-name> security.hexstrike.io/quarantined-
       # Scale deployment back up
       kubectl scale deployment -n hexstrike-agents <deployment-name> --replicas=1

     OPTION B: Keep quarantined (if suspicious)
       Preserve evidence
       Escalate to security team

  5. PREVENT future false positives:
     # Add maintenance window annotation
     kubectl annotate pod -n hexstrike-agents <pod-name> \
       hexstrike.io/maintenance-window="2026-04-16T02:00-04:00"
```

### Scenario 2.3: Falco Not Detecting Alerts

**Symptom**: Suspicious activity not being detected by Falco

**Detection**:

```bash
# Check Falco is running
kubectl get pods -n hexstrike-monitoring -l app=falco

# Check Falco rules are loaded
kubectl exec -n hexstrike-monitoring -l app=falco -- falco --list | grep hexstrike
```

**Response Procedure**:

```gherkin
GIVEN suspicious activity not being detected
WHEN Falco rules may not be working correctly
THEN follow this procedure:

  1. VERIFY Falco is running:
     kubectl get pods -n hexstrike-monitoring -l app=falco
     # All should be Running

  2. CHECK Falco logs for errors:
     kubectl logs -n hexstrike-monitoring -l app=falco --tail=100 | grep -iE "(error|failed|rule)"

  3. VERIFY rules are loaded:
     kubectl exec -n hexstrike-monitoring -l app=falco -- falco -L | grep -E "(shell|execve)"

  4. RELOAD rules if needed:
     kubectl exec -n hexstrike-monitoring -l app=falco -- kill -1 1

  5. CHECK Falco has correct permissions:
     kubectl auth can-i get pods -n hexstrike-agents --as=system:serviceaccount:hexstrike-monitoring:falco
     # Should return: yes
```

---

## 3. Rate Limit Exceeded Response

### Scenario 3.1: Rate Limit Errors (HTTP 429)

**Symptom**: MCP client receiving 429 Too Many Requests errors

**Detection**:

```bash
# Check rate limit metrics
curl -s http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080/metrics | jq '.status_codes["429"]'

# View proxy logs for rate limit hits
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --since=5m | grep -i "rate limit"
```

**Response Procedure**:

```gherkin
GIVEN HTTP 429 errors from MCP Policy Proxy
WHEN rate limit has been exceeded
THEN follow this procedure:

  1. IDENTIFY the source:
     kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --since=5m | \
       grep -i "rate limit" | head -10

  2. CHECK current rate limit configuration:
     kubectl get configmap -n hexstrike-system mcp-proxy-config -o jsonpath='{.data.RATE_LIMIT_PER_MINUTE}'
     # Default: 60 requests per minute

  3. DETERMINE if adjustment is needed:

     OPTION A: Increase rate limit (if legitimate high usage)
       kubectl patch configmap -n hexstrike-system mcp-proxy-config \
         -p '{"data":{"RATE_LIMIT_PER_MINUTE":"120"}}'
       kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system

     OPTION B: Investigate source (if unexpected high usage)
       Check logs for client identification
       Verify if traffic is from authorized clients
       Check for potential abuse or misconfiguration

  4. VERIFY fix:
     # Send test request after cooldown
     curl -X POST http://mcp-policy-proxy:8080/mcp/v1/tools \
       -H "Authorization: Bearer <token>" \
       -H "Content-Type: application/json" \
       -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

### Scenario 3.2: Legitimate Traffic Blocked

**Symptom**: Valid requests being rate limited during peak legitimate usage

**Response Procedure**:

```gherkin
GIVEN legitimate traffic being rate limited
WHEN business operations are impacted
THEN follow this procedure:

  1. ASSESS current limits vs actual usage:
     # Get metrics for the last hour
     curl -s http://mcp-policy-proxy:8080/metrics | jq

  2. CALCULATE appropriate new limit:
     # Consider: peak legitimate usage + 20% buffer

  3. UPDATE configuration:
     kubectl patch configmap -n hexstrike-system mcp-proxy-config \
       -p "{\"data\":{\"RATE_LIMIT_PER_MINUTE\":\"<calculated-limit>\"}}"

  4. RESTART proxy:
     kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system

  5. MONITOR for 24 hours:
     watch -n 30 'curl -s http://mcp-policy-proxy:8080/metrics | jq .status_codes'
```

---

## 4. General Troubleshooting

### 4.1 How to View Logs

**MCP Policy Proxy**:

```bash
# All proxy logs
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --tail=100

# Follow logs in real-time
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy -f

# Filter by correlation ID
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --since=5m | grep "X-Correlation-ID: <id>"

# Filter by error level
kubectl logs -n hexstrike-system -l app=mcp-policy-proxy --since=1h | grep -iE "(error|warn)"
```

**Falco**:

```bash
# All Falco events
kubectl logs -n hexstrike-monitoring -l app=falco --tail=100

# Filter by priority
kubectl logs -n hexstrike-monitoring -l app=falco --since=1h | grep "priority: CRITICAL"

# JSON format for analysis
kubectl logs -n hexstrike-monitoring -l app=falco --since=5m -o json | jq '.output_fields'
```

**Talon**:

```bash
# Talon action logs
kubectl logs -n hexstrike-monitoring -l app=talon --tail=100

# Filter by action type
kubectl logs -n hexstrike-monitoring -l app=talon --since=1h | grep -E "(terminate|quarantine)"
```

### 4.2 How to Check Metrics

**Proxy Metrics**:

```bash
# Get all metrics
curl -s http://mcp-policy-proxy.hexstrike-system.svc.cluster.local:8080/metrics | jq

# Key metrics to monitor:
# - total_requests: Overall traffic
# - blocked_requests: Security blocks
# - allowed_requests: Successful requests
# - avg_latency_ms: Response time
# - status_codes: Breakdown by HTTP code
```

**Prometheus Queries**:

```promql
# Request rate
rate(mcp_proxy_requests_total[5m])

# Block rate
rate(mcp_proxy_blocked_total[5m])

# Error rate
rate(mcp_proxy_requests_total{status_code=~"5.."}[5m])

# Latency percentiles
histogram_quantile(0.95, rate(mcp_proxy_latency_seconds_bucket[5m]))
```

### 4.3 How to Restart Components

**MCP Policy Proxy**:

```bash
# Rolling restart (zero downtime)
kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system

# Check rollout status
kubectl rollout status deployment/mcp-policy-proxy -n hexstrike-system

# Force restart (if stuck)
kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system --timeout=5m
```

**Falco**:

```bash
# Restart Falco daemonset
kubectl rollout restart daemonset/falco -n hexstrike-monitoring

# Check status
kubectl rollout status daemonset/falco -n hexstrike-monitoring
```

**Talon**:

```bash
# Restart Talon deployment
kubectl rollout restart deployment/talon -n hexstrike-monitoring

# Check status
kubectl rollout status deployment/talon -n hexstrike-monitoring
```

### 4.4 Network Connectivity Issues

```bash
# Test MCP backend connectivity from proxy
kubectl exec -it -n hexstrike-system deploy/mcp-policy-proxy -- \
  curl -v http://mcp-server:9090/health

# Test Lakera connectivity
kubectl exec -it -n hexstrike-system deploy/mcp-policy-proxy -- \
  curl -v https://api.lakera.ai/v1/health

# Check Cilium network policies
cilium policy get

# Check Hubble flows
hubble observe -n hexstrike-system --type drop
```

### 4.5 Common Error Resolution

**502 Bad Gateway**:

```gherkin
GIVEN HTTP 502 from MCP Policy Proxy
WHEN "MCP backend unavailable" message appears
THEN follow this procedure:

  1. CHECK MCP backend status:
     kubectl get pods -n hexstrike-system -l app=mcp-server

  2. CHECK service endpoints:
     kubectl get endpoints -n hexstrike-system -l app=mcp-server

  3. VERIFY backend URL configuration:
     kubectl get configmap -n hexstrike-system mcp-proxy-config -o jsonpath='{.data.MCP_BACKEND_URL}'

  4. RESTART MCP backend if needed:
     kubectl rollout restart deployment/mcp-server -n hexstrike-system
```

**401 Unauthorized**:

```gherkin
GIVEN HTTP 401 from MCP Policy Proxy
WHEN "Missing Authorization header" or "Invalid token" message appears
THEN follow this procedure:

  1. VERIFY Authorization header is present in request:
     curl -v -X POST http://mcp-proxy:8080/mcp/v1/tools \
       -H "Authorization: Bearer <your-jwt-token>"

  2. CHECK JWT secret configuration:
     kubectl get deployment -n hexstrike-system mcp-policy-proxy -o json | \
       jq '.spec.template.spec.containers[0].env[] | select(.name=="JWT_SECRET")'

  3. VALIDATE JWT token (decode without verification):
     # Use jwt.io to inspect token claims
```

---

## 5. Configuration Reference

### 5.1 Environment Variables

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `LISTEN_ADDR` | `0.0.0.0:8080` | Proxy listen address | `0.0.0.0:8080` |
| `MCP_BACKEND_URL` | `http://localhost:9090` | MCP backend endpoint | `http://mcp-server:9090` |
| `LAKERA_API_KEY` | (empty) | Lakera Guard API key | `sk-...` |
| `LAKERA_URL` | `https://api.lakera.ai` | Lakera API endpoint | `https://api.lakera.ai` |
| `LAKERA_TIMEOUT` | `5` | Lakera request timeout (seconds) | `10` |
| `LAKERA_THRESHOLD` | `70` | Risk score threshold (0-100) | `70` |
| `LAKERA_FAIL_MODE` | `closed` | Fail mode: `closed` (block) or `open` (allow) | `closed` |
| `RATE_LIMIT_PER_MINUTE` | `60` | Requests allowed per minute | `120` |
| `PROXY_TIMEOUT` | `30` | Backend request timeout (seconds) | `30` |
| `MAX_BODY_SIZE` | `1048576` | Max request body size (bytes) | `1048576` |
| `JWT_SECRET` | (empty) | JWT validation secret | `your-secret-key` |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` | `debug` |

### 5.2 ConfigMap Example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-proxy-config
  namespace: hexstrike-system
data:
  MCP_BACKEND_URL: "http://mcp-server:9090"
  LAKERA_API_KEY: ""  # Set via Secret
  LAKERA_URL: "https://api.lakera.ai"
  LAKERA_TIMEOUT: "10"
  LAKERA_THRESHOLD: "70"
  LAKERA_FAIL_MODE: "closed"
  RATE_LIMIT_PER_MINUTE: "60"
  LOG_LEVEL: "info"
  PROXY_TIMEOUT: "30"
```

### 5.3 Changing Settings

**Via ConfigMap (permanent changes)**:

```bash
# Edit ConfigMap
kubectl edit configmap -n hexstrike-system mcp-proxy-config

# Add or update a value
kubectl patch configmap -n hexstrike-system mcp-proxy-config \
  -p '{"data":{"LOG_LEVEL":"debug"}}'

# Restart to apply
kubectl rollout restart deployment/mcp-policy-proxy -n hexstrike-system
```

**Via Environment Variable (temporary)**:

```bash
# Set for single deployment
kubectl set env deployment/mcp-policy-proxy -n hexstrike-system \
  LAKERA_FAIL_MODE=open

# Set for all pods (selector)
kubectl set env deployment/mcp-policy-proxy -n hexstrike-system \
  LAKERA_FAIL_MODE=open --selector app=mcp-policy-proxy
```

**Via Secret (for sensitive values)**:

```bash
# Create secret
kubectl create secret generic mcp-proxy-secrets \
  -n hexstrike-system \
  --from-literal=lakera-api-key='sk-your-key' \
  --from-literal=jwt-secret='your-jwt-secret'

# Update deployment to use secret
kubectl patch deployment mcp-policy-proxy -n hexstrike-system \
  --type=json \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/env","value":[{"name":"LAKERA_API_KEY","valueFrom":{"secretKeyRef":{"name":"mcp-proxy-secrets","key":"lakera-api-key"}}}]}]'
```

### 5.4 Fail Mode Behavior

| Mode | Lakera Available | Lakera Unavailable | Lakera Timeout |
|------|------------------|-------------------|----------------|
| `closed` | Check request | **Block** (503) | **Block** (503) |
| `open` | Check request | **Allow** | **Allow** |

**Recommendation**: Use `closed` mode in production for maximum security. Use `open` mode only during Lakera outages as a temporary measure.

### 5.5 Rate Limiting Behavior

- **Algorithm**: Token bucket
- **Scope**: Global (all requests share the same bucket)
- **Reset**: Every 60 seconds
- **Response**: HTTP 429 with `Retry-After` header

---

## Escalation Contacts

| Role | Contact | When to Escalate |
|------|---------|------------------|
| On-Call SRE | PagerDuty: hexstrike-oncall | Component down > 5 min |
| Security Team | Slack: #security-incidents | Any CRITICAL Falco alert |
| Platform Team | Slack: #platform-support | Architectural issues |
| Lakera Support | support@lakera.ai | API issues > 15 min |

---

## See Also

- [Operations Guide](./OPERATIONS.md) - Deployment and setup procedures
- [Architecture](./ARCHITECTURE.md) - System design documentation
- [Security Policies](../HEXSTRIKE-DEFENSE-ESA-SECURITY-POLICIES.md) - Security configuration
