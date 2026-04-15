# HexStrike Defense Architecture

This document provides a comprehensive overview of the HexStrike Defense-in-Depth architecture, detailing the multi-layered security approach implemented to secure autonomous AI agents.

## Table of Contents

1. [Overview](#overview)
2. [Defense-in-Depth Layers](#defense-in-depth-layers)
3. [Architecture Diagrams](#architecture-diagrams)
4. [Layer-by-Layer Explanation](#layer-by-layer-explanation)
5. [Data Flow](#data-flow)
6. [Component Interactions](#component-interactions)

---

## Overview

HexStrike Defense implements a **7-layer defense-in-depth architecture** to protect autonomous AI agents with access to 150+ cybersecurity tools. Each layer provides independent security controls that inspect, validate, and contain agent behavior.

### Design Philosophy

```
┌─────────────────────────────────────────────────────────────────┐
│                     DEFENSE-IN-DEPTH PHILOSOPHY                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   "Security is not a product, but a process"                      │
│   - No single layer is foolproof                                │
│   - Each layer assumes the previous layers may fail              │
│   - Defense must be comprehensive and layered                    │
│   - Zero Trust: Never trust, always verify                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Defense-in-Depth Layers

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        HEXSTRIKE DEFENSE ARCHITECTURE                    │
│                         7-Layer Defense-in-Depth                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Layer 7: SDD Governance                                               │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Spec-Driven Development: Security requirements captured first  │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 6: Observability Integration                                    │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Sentry MCP | Atlassian MCP | Prometheus | LangGraph Agent     │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 5: Semantic Firewall (MCP Policy Proxy)                         │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Lakera Guard | Prompt Injection Detection | Rate Limiting      │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 4: Runtime Detection (Falco + Talon)                            │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  eBPF Syscall Monitoring | Shell Spawn Detection | Talon       │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 3: Network Containment (Cilium CNI)                             │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Default-Deny Egress | L7 Zero Trust | Hubble Logging         │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 2: Agent Isolation                                             │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Kubernetes Namespaces | Resource Limits | Pod Security         │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                    ↓                                    │
│   Layer 1: Infrastructure Security                                      │
│   ┌───────────────────────────────────────────────────────────────┐     │
│   │  Node Hardening | RBAC | Secrets Management | Network Policies │     │
│   └───────────────────────────────────────────────────────────────┘     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Architecture Diagrams

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL WORLD                                │
│                                                                          │
│    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                 │
│    │   User     │    │   LLM API   │    │  Target     │                 │
│    │  Request   │    │  Providers  │    │   Systems   │                 │
│    └──────┬─────┘    └──────┬─────┘    └──────▲─────┘                 │
│           │                  │                  │                      │
└───────────┼──────────────────┼──────────────────┼───────────────────────┘
            │                  │                  │
            ▼                  ▼                  │
┌─────────────────────────────────────────────────┴───────────────────────┐
│                        KUBERNETES CLUSTER                               │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      hexstrike-agents Namespace                   │   │
│  │                                                                   │   │
│  │   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐       │   │
│  │   │   HexStrike  │    │    MCP       │    │   LangGraph  │       │   │
│  │   │     AI       │───▶│   Policy     │◀──▶│    Agent     │       │   │
│  │   │   Agent      │    │   Proxy      │    │              │       │   │
│  │   └──────────────┘    └──────┬───────┘    └──────────────┘       │   │
│  │                              │                                     │   │
│  │   ┌──────────────────────────┼──────────────────────────┐        │   │
│  │   │                          │                          │        │   │
│  │   ▼                          ▼                          ▼        │   │
│  │   │ Layer 5                  │ Layer 4                   │ Layer 3  │   │
│  │   │ Semantic                │ Runtime                   │ Network  │   │
│  │   │ Firewall                │ Detection                 │ Contain  │   │
│  │   │                         │                          │          │   │
│  │   │ ┌─────────┐            │ ┌─────────┐              │ ┌─────┐  │   │
│  │   │ │ Lakera  │            │ │ Falco   │              │ │Cilium│  │   │
│  │   │ │ Guard   │            │ │ + Talon │              │ │     │  │   │
│  │   │ └─────────┘            │ └─────────┘              │ └─────┘  │   │
│  │   │                         │                          │          │   │
│  │   └─────────────────────────┼──────────────────────────┘          │   │
│  └────────────────────────────┼───────────────────────────────────────┘   │
│                                │                                          │
│  ┌────────────────────────────┼───────────────────────────────────────┐ │
│  │         hexstrike-monitoring Namespace                               │ │
│  │                                                                       │ │
│  │   ┌─────────┐    ┌─────────┐    ┌─────────┐                        │ │
│  │   │  Sentry │    │  Hubble │    │   etc   │                        │ │
│  │   │   MCP   │    │  Relay  │    │         │                        │ │
│  │   └─────────┘    └─────────┘    └─────────┘                        │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      hexstrike-system Namespace                    │   │
│  │                                                                   │   │
│  │   MCP Backend Services | ConfigMaps | Secrets                      │   │
│  │                                                                   │   │
│  └───────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          REQUEST FLOW                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   1. User Request                                                        │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 1: Infrastructure Security                               │   │
│   │  - RBAC validation                                              │   │
│   │  - TLS termination                                              │   │
│   │  - DDoS protection                                              │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 2: Agent Isolation                                        │   │
│   │  - Namespace isolation                                          │   │
│   │  - Resource quotas                                              │   │
│   │  - Pod Security Standards                                        │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 3: Network Containment (Cilium)                           │   │
│   │  - Default-deny egress                                          │   │
│   │  - DNS whitelisting                                              │   │
│   │  - Allowed endpoints only                                        │   │
│   │  - Hubble flow logging                                           │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 4: Runtime Detection (Falco + Talon)                     │   │
│   │  - eBPF syscall monitoring                                       │   │
│   │  - Shell spawn detection                                         │   │
│   │  - /etc write detection                                          │   │
│   │  - Automated response via Talon                                  │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 5: Semantic Firewall (MCP Policy Proxy)                   │   │
│   │  - JSON-RPC validation                                           │   │
│   │  - Lakera prompt injection detection                             │   │
│   │  - Rate limiting                                                 │   │
│   │  - Tool call filtering                                           │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Layer 6: Observability Integration                              │   │
│   │  - Metrics to Prometheus                                         │   │
│   │  - Errors to Sentry                                             │   │
│   │  - LangGraph state tracking                                      │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                                                                   │
│      ▼                                                                   │
│   7. MCP Backend (Tool Execution)                                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Security Event Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      SECURITY EVENT RESPONSE FLOW                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                     Malicious Activity Detected                   │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                    │
│                                    ▼                                    │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                    Falco eBPF Probe                               │   │
│   │   Detects: execve("/bin/bash"), write("/etc/passwd")            │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                    │
│                                    ▼                                    │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                    Event Classification                            │   │
│   │   - CRITICAL: Reverse shell, /etc write                         │   │
│   │   - WARNING: Shell spawn outside maintenance window              │   │
│   │   - INFO: Normal operations                                      │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                    │
│                   ┌────────────────┼────────────────┐                   │
│                   │                │                │                   │
│                   ▼                ▼                ▼                   │
│   ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐  │
│   │    CRITICAL       │  │     WARNING       │  │       INFO        │  │
│   │                   │  │                   │  │                   │  │
│   │ 1. Talon webhook │  │ 1. Add quarantine │  │ 1. Log event     │  │
│   │ 2. Terminate pod │  │    labels        │  │ 2. Update metrics│  │
│   │ 3. Create event  │  │ 2. Scale to 0    │  │                  │  │
│   │ 4. Capture logs  │  │ 3. Create event  │  │                  │  │
│   │ 5. Notify        │  │ 4. Notify        │  │                  │  │
│   └───────────────────┘  └───────────────────┘  └───────────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Layer-by-Layer Explanation

### Layer 1: Infrastructure Security

**Purpose**: Establish the foundation of cluster security

**Components**:
- Node hardening and OS-level security
- Role-Based Access Control (RBAC)
- Kubernetes secrets management
- Network policies baseline

**Key Configuration**:
```yaml
# RBAC - Principle of Least Privilege
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: hexstrike-agent
rules:
  - apiGroups: [""]
    resources: ["pods", "configmaps"]
    verbs: ["get", "list"]
  # No write permissions by default
```

---

### Layer 2: Agent Isolation

**Purpose**: Contain agent within Kubernetes namespaces

**Components**:
- Namespace isolation (hexstrike-agents, hexstrike-monitoring, hexstrike-system)
- Resource quotas and limits
- Pod Security Standards (PSS)
- Network policies

**Namespace Structure**:
```
hexstrike-agents      → Agent workloads
hexstrike-monitoring  → Falco, Talon, Hubble, Sentry
hexstrike-system      → MCP backend, ConfigMaps, Secrets
```

---

### Layer 3: Network Containment (Cilium CNI)

**Purpose**: Zero Trust networking with L7 policy enforcement

**Features**:
- Default-deny egress
- DNS whitelisting
- LLM API endpoint access control
- Target domain whitelisting
- Hubble flow logging and visibility

**Policy Example**:
```yaml
# Default deny all egress
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: hexstrike-default-deny
spec:
  endpointSelector:
    matchLabels:
      app: hexstrike-agent
  egressDeny:
    - toPorts:
        - port: "0"
          protocol: TCP
```

---

### Layer 4: Runtime Detection (Falco + Talon)

**Purpose**: Detect and respond to malicious runtime behavior

**Falco Rules**:
- Shell spawn detection (bash, sh, zsh, etc.)
- Reverse shell pattern detection
- /etc write detection
- Unauthorized network connections

**Talon Automated Response**:
| Priority | Action | Response Time |
|----------|--------|--------------|
| CRITICAL | Pod termination | < 200ms |
| WARNING | Pod quarantine (scale to 0) | < 500ms |
| INFO | Log and alert | Real-time |

---

### Layer 5: Semantic Firewall (MCP Policy Proxy)

**Purpose**: Validate MCP tool calls before execution

**Features**:
- JSON-RPC 2.0 validation
- Lakera Guard integration for prompt injection detection
- Rate limiting (60 req/min default)
- Graceful degradation (allows when Lakera unavailable)

**Middleware Chain**:
```
Request → Logging → Rate Limit → Auth → Semantic Check → MCP Backend
              ↑           ↑          ↑          ↑
         Log every   Enforce    Validate   Lakera Guard
         request     rate limit auth token  prompt injection
```

---

### Layer 6: Observability Integration

**Purpose**: Monitor, track, and alert on agent behavior

**Components**:
- Sentry MCP: Error tracking and alerting
- Atlassian MCP: Change management integration
- Prometheus: Metrics collection
- LangGraph Agent: State tracking and workflow monitoring

---

### Layer 7: SDD Governance

**Purpose**: Ensure security requirements are captured first

**SDD Workflow**:
```
Explore → Propose → Spec → Design → Tasks → Apply → Verify → Archive
    │                    │            │                    │
    ▼                    ▼            ▼                    ▼
Investigate        Write specs   Technical       Validate against
requirements       with Gherkin design         specs + run tests
```

---

## Data Flow

### Request Lifecycle

```
User Request
     │
     ▼
┌─────────────────────────────────────────┐
│ 1. Authentication (RBAC/Kubernetes)    │
└─────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────┐
│ 2. TLS Termination                      │
└─────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────┐
│ 3. Cilium Network Policy Check           │
│    - DNS resolution (kube-dns only)     │
│    - Allowed endpoint verification       │
└─────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────┐
│ 4. MCP Policy Proxy                      │
│    - JSON-RPC validation                 │
│    - Rate limit check                    │
│    - Lakera semantic check               │
└─────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────┐
│ 5. Tool Execution (MCP Backend)          │
└─────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────┐
│ 6. Runtime Monitoring (Falco)            │
│    - Syscall monitoring                  │
│    - Anomaly detection                   │
└─────────────────────────────────────────┘
     │
     ▼
Response to User
```

---

## Component Interactions

### Security Event Correlation

```
┌─────────────────────────────────────────────────────────────────┐
│                    EVENT CORRELATION MATRIX                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Event Source          │   Layer   │   Action   │  Response   │
│   ──────────────────────┼───────────┼────────────┼────────────  │
│   Lakera Guard          │   Layer 5 │   Block     │  403 Error   │
│   Rate Limiter          │   Layer 5 │   Block     │  429 Error   │
│   Cilium Policy         │   Layer 3 │   Block     │  Dropped     │
│   Falco (CRITICAL)      │   Layer 4 │   Terminate │  Pod killed │
│   Falco (WARNING)       │   Layer 4 │   Quarantine│  Scale 0     │
│   Hubble Flow Log       │   Layer 3 │   Log       │  Audit trail │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Security Properties

| Property | Layer | Implementation |
|----------|-------|----------------|
| Confidentiality | 1, 3, 5 | Secrets encryption, network policies, semantic filtering |
| Integrity | 2, 4, 5 | PSS, Falco detection, prompt validation |
| Availability | 1, 6 | RBAC, observability, graceful degradation |
| Non-repudiation | 3, 4, 6 | Hubble logs, Falco audit, Sentry |
| Defense in Depth | 1-7 | All layers working together |

---

## Risk Mitigation Matrix

| Threat | Layer(s) | Mitigation |
|--------|----------|------------|
| Prompt Injection | 5 | Lakera Guard semantic analysis |
| Command Injection | 4, 5 | Falco detection, Lakera filtering |
| Data Exfiltration | 3 | Cilium default-deny egress |
| Privilege Escalation | 2, 4 | Namespace isolation, Falco monitoring |
| Denial of Service | 1, 5 | Rate limiting, resource quotas |
| Supply Chain Attack | 1 | Image scanning, signed images |
| Insider Threat | 2, 4 | Least privilege, audit logging |
| Zero-day Exploit | 1-7 | Layered defense, rapid response |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024 | HexStrike Team | Initial architecture document |

---

## See Also

- [Operations Guide](./OPERATIONS.md) - Deployment and troubleshooting
- [Security Hardening](./SECURITY.md) - Security best practices
- [Proposal](../openspec/changes/hexstrike-defense-architecture/proposal.md) - Original design proposal
