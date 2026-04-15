# HexStrike Defense Extended E2E Test Suite

Extended end-to-end tests for HexStrike Defense architecture, focusing on Red Teaming vulnerabilities.

## Overview

This suite complements the base E2E tests with comprehensive coverage of attack vectors identified through Red Teaming exercises.

## Test Categories

### 1. Semantic Proxy Tests (Lakera)

Tests the MCP Policy Proxy's semantic firewall against prompt injection and jailbreak attempts.

| Test | Description |
|------|-------------|
| TestPromptInjection_Direct | Direct prompt injection patterns |
| TestPromptInjection_Base64Encoded | Base64 encoded attempts |
| TestPromptInjection_UnicodeObfuscation | Unicode obfuscation techniques |
| TestJailbreak_IgnorePreviousInstructions | Classic jailbreak patterns |
| TestJailbreak_CharacterRolePlay | Role-play based jailbreaks |
| TestContextExhaustion_TokenPadding | Token padding attacks |
| TestRateLimiting_Enforced | Rate limiting enforcement |

### 2. Runtime Security Tests (Falco/Talon)

Tests runtime detection and response mechanisms.

| Test | Description |
|------|-------------|
| TestReverseShell_bash_i | Bash interactive reverse shell |
| TestReverseShell_netcat | Netcat reverse shell variants |
| TestReverseShell_python | Python socket-based shells |
| TestReverseShell_curl_pipe_bash | curl \| bash patterns |
| TestFileWrite_etc_passwd | /etc passwd write attempts |
| TestExec_from_unusual_directory | Execution from unusual directories |
| TestRuntimeSecurity_FalcoAlertValidation | Alert timing validation |

### 3. Network Security Tests (Cilium)

Tests network policy enforcement.

| Test | Description |
|------|-------------|
| TestEgress_BlockedNonWhitelistedDomain | Egress to non-whitelisted domains |
| TestEgress_Allowed_api_anthropic_com | LLM API endpoint access |
| TestDNS_Only_CoreDNS | DNS whitelisting |
| TestIngress_DefaultDeny | Ingress default deny |
| TestL7Protocol_DROP_on_C2 | L7 protocol enforcement |
| TestNetworkSecurity_CiliumPolicyValidation | Policy validation |

### 4. Governance Tests (SDD)

Tests SDD governance layer enforcement.

| Test | Description |
|------|-------------|
| TestSpecImmutability | Delta spec immutability |
| TestContractEnforcement | Layer contract enforcement |
| TestNoScopeExpansion | Scope validation |
| TestGovernance_SDDWorkflow | SDD workflow validation |
| TestGovernance_ChangeLifecycle | Change lifecycle validation |
| TestGovernance_ValidationEnforcement | Validation enforcement |

## Running Tests

### Run All Extended Tests

```bash
cd tests/e2e/extended_suite
go mod tidy
go test -v ./...
```

### Run Specific Category

```bash
# Semantic proxy tests
go test -v -run TestPromptInjection

# Runtime security tests
go test -v -run TestReverseShell

# Network security tests
go test -v -run TestEgress

# Governance tests
go test -v -run TestGovernance
```

### Skip Long-Running Tests

```bash
go test -v -short ./...
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KUBECONFIG` | `~/.kube/config` | Kubernetes config path |
| `MCP_PROXY_URL` | `http://localhost:8080` | MCP Policy Proxy URL |
| `HEXSTRIKE_NAMESPACE` | `hexstrike-agents` | Agent namespace |
| `OPENSPEC_DIR` | (see README) | OpenSpec directory |

## Prerequisites

1. Running Kubernetes cluster with Cilium CNI
2. HexStrike Defense components deployed
3. Lakera API key configured (for semantic tests)
4. `kubectl` configured with cluster access

## Test Execution Notes

### Semantic Proxy Tests

- Require `LAKERA_API_KEY` environment variable
- Without API key, tests pass (graceful degradation)
- Tests verify blocking behavior when configured

### Runtime Security Tests

- Require Kubernetes cluster access
- Execute commands in test pods
- Verify Falco alerts through logs
- Talon response tests are destructive (skipped in CI)

### Network Security Tests

- Require pod execution capability
- Test egress policy enforcement
- Verify Hubble logs for dropped connections

### Governance Tests

- Require OpenSpec directory access
- Verify SDD workflow compliance
- Check validation enforcement

## Attack Vectors Covered

### Prompt Injection

- Direct instructions: "Ignore all previous instructions"
- Base64 encoding attempts
- Unicode obfuscation (zero-width, homoglyphs, combining chars)
- Context exhaustion via token padding

### Jailbreak

- DAN (Do Anything Now) patterns
- Character/role-play exploitation
- "New instructions" bypass
- Fictional AI simulation

### Runtime Attacks

- Reverse shell: bash -i, nc, python, curl|bash
- File write: /etc/passwd, /etc/shadow, /etc/hosts
- Execution from unusual directories

### Network Attacks

- Egress to non-whitelisted domains
- DNS exfiltration
- C2 protocol patterns
- Unauthorized ingress

## Troubleshooting

### "Skipping e2e test in short mode"

Use `go test -v` (without `-short`) for full E2E tests.

### "Skipping test - no Kubernetes cluster"

Ensure `kubectl` is configured and cluster is accessible.

### Lakera blocks fail silently

Without `LAKERA_API_KEY`, proxy allows all requests (expected behavior).

## Contributing

When adding new tests:
1. Follow table-driven test pattern
2. Add test to appropriate category
3. Document attack vector in comments
4. Test both positive and negative cases