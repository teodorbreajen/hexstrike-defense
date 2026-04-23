# High-Level Architecture

## Defense-in-Depth Layers

```mermaid
graph TD
    subgraph "Layer 7: SDD Governance"
        SDD[Spec-Driven Development]
        Policy[Security Policies]
    end

    subgraph "Layer 6: Observability"
        OBS1[Sentry MCP]
        OBS2[Prometheus]
        OBS3[Hubble]
    end

    subgraph "Layer 5: Semantic Firewall"
        SF[Lakera Guard]
        RL[Rate Limiter]
    end

    subgraph "Layer 4: Runtime Detection"
        RT[Falco + eBPF]
        Tal[Talon]
    end

    subgraph "Layer 3: Network Containment"
        NC[Cilium CNI]
        NP[Network Policies]
    end

    subgraph "Layer 2: Agent Isolation"
        ISO[Kubernetes Namespaces]
        Quota[Resource Quotas]
    end

    subgraph "Layer 1: Infrastructure"
        INF1[Node Hardening]
        INF2[RBAC]
    end

    User --> SDD
    SDD --> OBS1
    OBS1 --> SF
    SF --> RT
    RT --> NC
    NC --> ISO
    ISO --> INF1
```

## Component Architecture

```mermaid
flowchart LR
    subgraph Proxy
        P[MCP Policy Proxy]
        MW[Middleware Chain]
    end

    subgraph Backend
        MCP[MCP Server]
        LK[Lakera Guard]
        RED[(Redis)]
    end

    Client -.-> P
    P -.-> MCP
    P -.-> LK
    P -.-> RED

```

## Request Processing Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant L as Lakera
    participant M as MCP Backend
    participant R as Redis

    C->>P: HTTP Request
    P->>P: Security Headers
    P->>P: Rate Limit Check
    alt Rate Limited
        P-->>C: HTTP 429 Too Many Requests
    else
        P->>P: JWT Validation
        alt Invalid Token
            P-->>C: HTTP 401 Unauthorized
        else
            P->>L: Semantic Check
            alt Blocked
                L-->>P: Score >= Threshold
                P-->>C: HTTP 403 Forbidden
            else Allowed
                L-->>P: Score < Threshold
                P->>M: Forward Request
                M-->>P: Response
                P-->>C: HTTP 200 OK
            end
        end
    end
```

## Security Layers

| Layer | Component | Function | Status |
|-------|-----------|----------|--------|
| 7 | SDD Governance | Security requirements captured first | ✓ Active |
| 6 | Observability | Monitoring and alerting | ✓ Active |
| 5 | Semantic Firewall | Input validation | ✓ Active |
| 4 | Runtime Detection | Behavioral monitoring | ✓ Active |
| 3 | Network Containment | Zero-trust networking | ✓ Active |
| 2 | Agent Isolation | Namespace isolation | ✓ Active |
| 1 | Infrastructure | Node hardening | ✓ Active |

## Design Principles

1. **Defense in Depth** - Multiple security layers
2. **Fail Secure** - Fail-closed by default
3. **Least Privilege** - Minimal permissions
4. **Zero Trust** - Never trust, always verify
5. **Observable** - Full visibility into system state

---

*Generated from code analysis*
