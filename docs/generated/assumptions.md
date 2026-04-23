# Design Decisions

## Architecture Decisions

| Decision | Rationale | Impact |
|----------|----------|---------|
| Go for proxy | Performance, single binary | High |
| Kubernetes | Container orchestration | High |
| Cilium CNI | eBPF-based networking | Medium |
| Lakera Guard | Semantic security | Medium |
| JWT Bearer | Standard auth | Low |

## Implementation Decisions

| Decision | Rationale |
|----------|----------|
| Fail-closed by default | Security first |
| JSON logging | SIEM integration |
| Prometheus metrics | Standard monitoring |
| Middleware chain | Extensibility |

---

*Generated from architecture analysis*
