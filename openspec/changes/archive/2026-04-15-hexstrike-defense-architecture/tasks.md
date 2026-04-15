# Tasks: hexstrike-defense-architecture

## Phase 1: Infrastructure Foundation

- [x] 1.1 Create directory structure under `manifests/{cilium,falco,mcp-proxy,langgraph}/`
- [x] 1.2 Create base Helm chart `manifests/charts/hexstrike-defense/Chart.yaml` with dependencies
- [x] 1.3 Create Kubernetes namespaces: `hexstrike-system`, `hexstrike-monitoring`, `hexstrike-agents`
- [x] 1.4 Create `manifests/charts/hexstrike-defense/values.yaml` with feature flags per layer

## Phase 2: Network Security (Cilium)

- [x] 2.1 Create `manifests/cilium/00-default-deny.yaml` with CiliumClusterwideNetworkPolicy default-deny egress
- [x] 2.2 Create `manifests/cilium/01-dns-whitelist.yaml` allowing kube-dns only
- [x] 2.3 Create `manifests/cilium/02-llm-endpoints.yaml` whitelist for LLM API endpoints
- [x] 2.4 Create `manifests/cilium/03-target-domains.yaml` with allowed target domains list
- [x] 2.5 Create `manifests/cilium/04-hubble-enable.yaml` to enable Hubble logging

## Phase 3: Runtime Security (Falco + Talon)

- [x] 3.1 Create `manifests/falco/01-execve-rules.yaml` Falco rules for execve syscall detection
- [x] 3.2 Create `manifests/falco/02-etc-write-rules.yaml` Falco rules for /etc writes
- [x] 3.3 Create `manifests/falco/talon.yaml` with Talon webhook configuration
- [x] 3.4 Create `manifests/falco/annotations.yaml` with hexstrike container annotations
- [x] 3.5 Create `manifests/falco/kustomization.yaml` for kustomize overlay

## Phase 4: Semantic Firewall (MCP Policy Proxy)

- [x] 4.1 Initialize Go module `src/mcp-policy-proxy/go.mod` with net/http, lakera SDK deps
- [x] 4.2 Create `src/mcp-policy-proxy/jsonrpc.go` with JSON-RPC 2.0 message handling
- [x] 4.3 Create `src/mcp-policy-proxy/lakera.go` with Lakera API client (content moderation)
- [x] 4.4 Create `src/mcp-policy-proxy/proxy.go` with middleware chain (auth, rate-limit, validation)
- [x] 4.5 Create `src/mcp-policy-proxy/main.go` with health check endpoint and server bootstrap
- [x] 4.6 Create `manifests/mcp-proxy/configmap.yaml` for Lakera API keys and config
- [x] 4.7 Create `manifests/mcp-proxy/deployment.yaml` and `manifests/mcp-proxy/service.yaml`
- [x] 4.8 Create Dockerfile for MCP Policy Proxy image build

## Phase 5: Observability Integration

- [x] 5.1 Create `manifests/langgraph/agent-config.yaml` with LangGraph agent security constraints
- [x] 5.2 Create `manifests/langgraph/mcp-sentry-config.yaml` Sentry MCP integration template
- [x] 5.3 Create `manifests/langgraph/mcp-atlassian-config.yaml` Atlassian MCP config template
- [x] 5.4 Create `manifests/mcp-proxy/prometheus-servicemonitor.yaml` for metrics scraping

## Phase 6: SDD Governance Setup

- [x] 6.1 Create `scripts/validate.sh` to run spec validation against running cluster
- [x] 6.2 Create `scripts/deploy.sh` with kubectl apply and health check verification
- [x] 6.3 Create `scripts/test-attacks.sh` with ethical red-team test scenarios
- [x] 6.4 Create `.github/workflows/sdd-validate.yaml` CI/CD hook for OpenSpec validation

## Phase 7: Testing and Documentation

- [x] 7.1 Create `tests/e2e/test_semantic_firewall.go` for MCP proxy validation tests
- [x] 7.2 Create `tests/e2e/test_falco_detection.go` for runtime detection tests
- [x] 7.3 Create `tests/e2e/test_cilium_policies.go` for network policy enforcement tests
- [x] 7.4 Create `docs/ARCHITECTURE.md` with defense-in-depth layer diagrams
- [x] 7.5 Create `docs/OPERATIONS.md` with deployment runbook and troubleshooting
- [x] 7.6 Create `docs/SECURITY.md` with hardening guide and incident response
