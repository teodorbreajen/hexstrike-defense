# Repository Structure

## Directory Tree

```
deploy/
  prometheus.yml
manifests/
      Chart.yaml
      values.yaml
    00-default-deny.yaml
    01-dns-whitelist.yaml
    02-llm-endpoints.yaml
    03-target-domains.yaml
    04-hubble-enable.yaml
    01-execve-rules.yaml
    02-etc-write-rules.yaml
    annotations.yaml
    kustomization.yaml
    talon.yaml
    agent-config.yaml
    mcp-atlassian-config.yaml
    mcp-sentry-config.yaml
    configmap.yaml
    deployment.yaml
    network-policy.yaml
    prometheus-servicemonitor.yaml
    service.yaml
openspec/
      design.md
      proposal.md
      spec.md
      tasks.md
      design.md
      proposal.md
      tasks.md
      verify-report.md
      design.md
      proposal.md
      tasks.md
  config.yaml
      spec.md
      spec.md
      spec.md
      spec.md
      spec.md
scripts/
  check-updates.sh
  deploy.sh
  test-attacks.sh
  validate.sh
sdd/
src/
    Dockerfile
    Makefile
    config.go
    cors.go
    cors_test.go
      cleanup.go
      dlq.go
      dlq_test.go
      README.md
    fuzz_test.go
    go.mod
    go.sum
    integration_test.go
    jsonrpc.go
    lakera.go
    logger.go
    logger_test.go
    main.go
    mcp-policy-proxy-obs.exe
    mcp-policy-proxy.exe
    metrics_test.go
    prometheus.go
    prometheus_test.go
    proxy.go
    proxy_handler_test.go
    race_test.go
    rate_limiter_test.go
    retry_client.go
    retry_client_test.go
    security_comprehensive_test.go
    security_test.go
tests/
    README.md
      README.md
      go.mod
      go.sum
      governance_test.go
      network_security_test.go
      runtime_security_test.go
      semantic_proxy_test.go
      cluster.go
      go.mod
      go.sum
      utils.go
    go.mod
    go.sum
    test_cilium_policies.go
    test_falco_detection.go
    test_semantic_firewall.go
```

## Key Files

| Path | Type | Purpose |
|------|------|---------|
| src/mcp-policy-proxy/main.go | source | Entry point |
| src/mcp-policy-proxy/proxy.go | source | Proxy logic |
| src/mcp-policy-proxy/lakera.go | source | Lakera client |
| src/mcp-policy-proxy/Dockerfile | dockerfile | Container definition |
| Makefile | makefile | Build automation |
| manifests/ | manifests | Kubernetes configs |
| scripts/deploy.sh | script | Deployment script |