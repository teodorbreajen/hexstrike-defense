```mermaid
flowchart TB
    subgraph "Layer 7: SDD Governance"
        SDD[Spec-Driven Development]
    end

    subgraph "Layer 6: Observability"
        OBS[Sentry | Prometheus]
    end

    subgraph "Layer 5: Semantic Firewall"
        SF[Lakera Guard | Rate Limiter]
    end

    subgraph "Layer 4: Runtime Detection"
        RT[Falco | Talon]
    end

    subgraph "Layer 3: Network"
        NET[Cilium CNI]
    end

    subgraph "Layer 2: Isolation"
        ISO[Kubernetes NS]
    end

    subgraph "Layer 1: Infra"
        INF[RBAC | Hardening]
    end

    User --> SF
    SF --> RT
    RT --> NET
    NET --> ISO
    ISO --> INF
```