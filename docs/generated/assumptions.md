# Decisions & Assumptions

## Architecture Decisions

| Decision | Rationale |
|----------|----------|
| Go for proxy | Performance, single binary, Kubernetes support |
| Kubernetes | Container orchestration, scaling |
| Cilium CNI | eBPF-based networking, zero-trust |
| Lakera Guard | Semantic security for AI agents |
| JWT Bearer | Standard auth, stateless |

## Security Assumptions

- JWT_SECRET configured in production
- Lakera API key provided
- TLS enabled in production
- Network policies enforced
- Runtime monitoring active

## Known Limitations

- Kubernetes required for full deployment
- Lakera API connectivity required
- JWT mandatory in production
- External API keys needed

---

*Generated from code and config analysis*
