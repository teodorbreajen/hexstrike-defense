# HexStrike Security Hardening Guide

This guide provides security hardening recommendations, best practices, and incident response procedures for the HexStrike Defense architecture.

## Table of Contents

1. [Secure Defaults](#secure-defaults)
2. [Network Policy Best Practices](#network-policy-best-practices)
3. [Secret Management](#secret-management)
4. [Incident Response Playbook](#incident-response-playbook)
5. [Compliance Considerations](#compliance-considerations)

---

## Secure Defaults

### MCP Policy Proxy Configuration

```yaml
# manifests/mcp-proxy/configmap.yaml - SECURE DEFAULTS
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-proxy-config
  namespace: hexstrike-system
data:
  config.json: |
    {
      "listen_addr": "0.0.0.0:8080",
      "rate_limit_per_minute": 60,
      "proxy_timeout": 30,
      "lakera_timeout": 5,
      "allowed_origins": [],
      "require_auth": true
    }
```

### Hardening Checklist

| Setting | Default | Recommended | Reason |
|---------|---------|-------------|--------|
| Rate Limiting | 100/min | 60/min | Prevent abuse |
| Lakera Timeout | 10s | 5s | Fail fast |
| Proxy Timeout | 60s | 30s | Prevent hanging |
| Auth Required | false | true | Prevent unauthorized access |
| Debug Mode | true | false | Information leakage |

### Falco Secure Configuration

```yaml
# manifests/falco/annotations.yaml - Security annotations
podAnnotations:
  # Enable Falco monitoring
  security.cyber.com/falco/enable: "true"
  
  # Set priority threshold
  security.cyber.com/falco/priority: "WARNING"
  
  # Configure response actions
  security.cyber.com/falco/action: "terminate"
  
  # Tag for Talon processing
  hexstrike.io/monitor: "true"
  hexstrike.io/policy-tier: "critical"
```

### Pod Security Standards

```yaml
# Apply PSS restricted policy
apiVersion: v1
kind: Namespace
metadata:
  name: hexstrike-agents
  labels:
    # Enforce restricted PSS
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: v1.29
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/audit: restricted
```

---

## Network Policy Best Practices

### Default-Deny First

**ALWAYS** apply default-deny before whitelisting:

```bash
# 1. Apply default-deny FIRST
kubectl apply -f manifests/cilium/00-default-deny.yaml

# 2. Verify it's active
cilium policy get | grep default-deny

# 3. THEN apply allowlists
kubectl apply -f manifests/cilium/01-dns-whitelist.yaml
```

### Principle of Least Privilege

```yaml
# BAD: Too permissive
egress:
  - toPorts:
      - port: "443"
        protocol: TCP

# GOOD: Specific destinations only
egress:
  - toFQDNs:
      - matchName: api.openai.com
      - matchPattern: "*.github.com"
```

### Network Policy Review Checklist

- [ ] Default-deny applied to all hexstrike namespaces
- [ ] DNS whitelisted to kube-dns only
- [ ] LLM endpoints explicitly listed
- [ ] No `toEntities: ["world"]` rules
- [ ] Hubble logging enabled for audit
- [ ] Regular policy review (quarterly)

### Segregation by Trust Level

```yaml
# High-trust namespace (hexstrike-system)
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: hexstrike-system-internal
  namespace: hexstrike-system
spec:
  endpointSelector:
    matchLabels:
      namespace: hexstrike-system
  egress:
    # Allow internal cluster communication
    - toEndpoints:
        - matchLabels:
            namespace: hexstrike-system
    # Allow DNS
    - toServices:
        - kubeDNS:
            namespace: kube-system
---
# Low-trust namespace (hexstrike-agents)
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: hexstrike-agents-strict
  namespace: hexstrike-agents
spec:
  endpointSelector:
    matchLabels:
      app: hexstrike-agent
  egressDeny:
    - toPorts:
        - port: "0"
          protocol: TCP
  egress:
    # Only whitelisted destinations
    - toFQDNs:
        - matchName: api.openai.com
```

---

## Secret Management

### Secrets Security Requirements

| Requirement | Implementation |
|-------------|----------------|
| Encryption at rest | Enable etcd encryption |
| Encryption in transit | TLS for all connections |
| Access control | RBAC + Vault policies |
| Rotation | Automated rotation (90 days) |
| Audit | All access logged |

### Kubernetes Secrets Best Practices

```yaml
# Use external secrets management
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: lakera-credentials
  namespace: hexstrike-system
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  target:
    name: lakera-credentials
    creationPolicy: Owner
  data:
    - secretKey: api-key
      remoteRef:
        key: production/lakera
        property: api-key
```

### Enable etcd Encryption

```bash
# Create encryption key
kubectl create secret generic encryption-key \
  --from-literal=key=$(head -c 32 /dev/urandom | base64) \
  -n kube-system

# Create encryption configuration
cat <<EOF | kubectl apply -f -
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
metadata:
  name: encryption-config
spec:
  resources:
    - resources:
        - secrets
      providers:
        - aescbc:
            keys:
              - name: key1
                secret: <base64-encoded-key>
        - identity: {}
EOF
```

### Rotate API Keys

```bash
# Rotate Lakera API key
kubectl create secret generic lakera-key-v2 \
  --from-literal=api-key=$NEW_LAKERA_KEY \
  -n hexstrike-system

# Update ConfigMap reference
kubectl patch configmap mcp-proxy-config \
  --patch '{"data":{"lakera_api_key":"$NEW_LAKERA_KEY"}}'

# Verify new key is working
curl -s http://mcp-proxy:8080/health | jq .checks.lakera
```

---

## Incident Response Playbook

### Incident Classification

| Level | Description | Response Time | Examples |
|-------|-------------|---------------|----------|
| P0 - Critical | Active breach, data exfiltration | Immediate | Reverse shell, Talon failed to terminate |
| P1 - High | Detection triggered, no breach | < 15 min | Falco CRITICAL alert, Lakera blocked |
| P2 - Medium | Anomaly detected | < 1 hour | Multiple rate limit hits, false positive |
| P3 - Low | Informational | Next business day | Maintenance window, testing |

### P0 Incident Response

**CRITICAL: Reverse Shell or Unauthorized Access Detected**

```bash
# Step 1: IMMEDIATE ISOLATION
# Terminate affected pods immediately
kubectl delete pod <pod-name> -n hexstrike-agents --grace-period=0 --force

# Scale deployment to 0 to prevent rescheduling
kubectl scale deployment hexstrike-agent -n hexstrike-agents --replicas=0

# Step 2: CHECK FOR LATERAL MOVEMENT
# Review Hubble logs for the past hour
hubble observe -n hexstrike-agents --since 1h --type drop

# Check other pods in namespace
kubectl get pods -n hexstrike-agents -o wide

# Step 3: COLLECT EVIDENCE
# Capture pod logs
kubectl logs <pod-name> -n hexstrike-agents --previous > pod-logs.txt

# Export Falco events
kubectl logs -l app=falco -n hexstrike-monitoring > falco-events.txt

# Export Hubble flows
hubble observe -n hexstrike-agents --output json > hubble-flows.json

# Step 4: NOTIFY
# Alert security team
# Preserve evidence for forensics
# Begin incident report

# Step 5: RESTORE
# After investigation, redeploy with fresh images
kubectl scale deployment hexstrike-agent -n hexstrike-agents --replicas=1
```

### P1 Incident Response

**Falco CRITICAL Alert Triggered**

```bash
# Step 1: VERIFY ALERT
kubectl logs -l app=falco -n hexstrike-monitoring | grep -A 5 "CRITICAL"

# Step 2: CHECK TALON RESPONSE
kubectl get events -n hexstrike-agents --sort-by='.lastTimestamp'

# Step 3: IF TALON FAILED
# Manually quarantine pod
kubectl label pod <pod-name> -n hexstrike-agents \
  security.hexstrike.io/quarantined=true \
  security.hexstrike.io/quarantine-reason="Manual quarantine"

kubectl scale deployment hexstrike-agent -n hexstrike-agents --replicas=0

# Step 4: INVESTIGATE
# Review timeline of events
# Check for patterns

# Step 5: RESTORE AND HARDEN
# Redeploy with additional monitoring
kubectl scale deployment hexstrike-agent -n hexstrike-agents --replicas=1
```

### Post-Incident Actions

```bash
# 1. Conduct root cause analysis
# 2. Update Falco rules if needed
# 3. Adjust network policies
# 4. Add monitoring/alerting
# 5. Document lessons learned
# 6. Update incident response plan
```

---

## Compliance Considerations

### SOC 2 Type II

| Trust Principle | HexStrike Implementation |
|-----------------|-------------------------|
| Security | RBAC, Network Policies, PSS |
| Availability | Health Checks, Graceful Degradation |
| Processing Integrity | Semantic Firewall, Rate Limiting |
| Confidentiality | Encryption, Secrets Management |
| Privacy | Network Isolation, Default-Deny |

### GDPR Considerations

| Requirement | Implementation |
|-------------|----------------|
| Data minimization | Network policies restrict data exfiltration |
| Right to deletion | Pod deletion = data deletion |
| Breach notification | Falco + Talon detect exfiltration |
| Data protection | Encryption at rest and in transit |

### Audit Logging Requirements

Enable comprehensive audit logging:

```yaml
# Kubernetes Audit Policy
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  # Log all requests to hexstrike namespaces
  - level: RequestResponse
    namespaces: ["hexstrike-agents", "hexstrike-system", "hexstrike-monitoring"]
  
  # Log all write operations
  - level: Metadata
    verbs: ["create", "update", "delete"]
    resources:
      - group: ""
        resources: ["pods", "services", "configmaps"]
```

### Hubble Audit Compliance

```bash
# Enable Hubble persistence
cilium config set hubble-enabled true
cilium config set hubble-flow-buffer-size 4096
cilium config set hubble-metrics "flow:sourceContext=pod;destinationContext=pod"

# Export flows to SIEM
hubble observe --type drop --output json > /var/log/hexstrike-drops.json
```

---

## Security Hardening Automation

Use this script to verify hardening compliance:

```bash
#!/bin/bash
# security-hardening-check.sh

echo "=== HexStrike Security Hardening Check ==="

PASS=0
FAIL=0

# Check 1: etcd encryption enabled
if kubectl get configmap -n kube-system encryption-config &>/dev/null; then
  echo "✅ etcd encryption configured"
  ((PASS++))
else
  echo "❌ etcd encryption NOT configured"
  ((FAIL++))
fi

# Check 2: PSS restricted on hexstrike namespaces
PSS_STATUS=$(kubectl get namespace hexstrike-agents -o jsonpath='{.metadata.labels.pod-security\.kubernetes\.io\/enforce}')
if [ "$PSS_STATUS" == "restricted" ]; then
  echo "✅ PSS restricted enforced"
  ((PASS++))
else
  echo "❌ PSS restricted NOT enforced"
  ((FAIL++))
fi

# Check 3: Network policies applied
POLICY_COUNT=$(kubectl get ciliumnetworkpolicies -A --no-headers | wc -l)
if [ $POLICY_COUNT -ge 4 ]; then
  echo "✅ Network policies applied ($POLICY_COUNT)"
  ((PASS++))
else
  echo "❌ Insufficient network policies ($POLICY_COUNT)"
  ((FAIL++))
fi

# Check 4: Falco running
FALCO_RUNNING=$(kubectl get pods -l app=falco -n hexstrike-monitoring --no-headers 2>/dev/null | grep Running | wc -l)
if [ $FALCO_RUNNING -gt 0 ]; then
  echo "✅ Falco running ($FALCO_RUNNING instances)"
  ((PASS++))
else
  echo "❌ Falco NOT running"
  ((FAIL++))
fi

# Check 5: Talon webhook registered
TALON_WEBHOOK=$(kubectl get mutatingwebhookconfigurations 2>/dev/null | grep talon | wc -l)
if [ $TALON_WEBHOOK -gt 0 ]; then
  echo "✅ Talon webhook registered"
  ((PASS++))
else
  echo "❌ Talon webhook NOT registered"
  ((FAIL++))
fi

echo ""
echo "=== Summary ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"

if [ $FAIL -eq 0 ]; then
  echo "✅ All hardening checks passed!"
  exit 0
else
  echo "❌ Some hardening checks failed"
  exit 1
fi
```

---

## Security Updates

### Regular Security Maintenance

| Task | Frequency | Owner |
|------|-----------|-------|
| Update Falco rules | Weekly | Security Team |
| Review network policies | Monthly | Network Team |
| Rotate secrets | Quarterly | DevOps |
| Pen testing | Annually | External |
| Update Kubernetes | Per release | Platform Team |
| Review incidents | Monthly | Security Team |

### Dependency Updates

```bash
# Update MCP Policy Proxy dependencies
cd src/mcp-policy-proxy
go get -u
go mod tidy
go build

# Update Helm charts
helm repo update
helm upgrade hexstrike-defense ./manifests/charts/hexstrike-defense

# Update Cilium
cilium upgrade
```

---

## Support

For security-related questions:

- Security issues: security@hexstrike.ai
- General support: support@hexstrike.ai
- Documentation: https://docs.hexstrike.ai

**Note**: For critical vulnerabilities, follow responsible disclosure practices.
