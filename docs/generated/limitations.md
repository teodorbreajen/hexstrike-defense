# Risks & Limitations

## Known Limitations

1. **External Dependencies**
   - Requires Lakera API connectivity
   - Cannot operate in fail-closed without API

2. **Deployment Constraints**
   - Requires Kubernetes cluster
   - Requires external secrets

3. **Security Trade-offs**
   - Fail-open mode available but not recommended

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Lakera downtime | Fail-closed blocks requests |
| JWT misconfiguration | Production detection |
| SSRF attacks | IP whitelist validation |
| DoS attacks | Rate limiting per client |

---

*Generated from security code*
