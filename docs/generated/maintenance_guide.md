# Maintenance Guide

## Regular Tasks

| Task | Frequency | Notes |
|------|-----------|-------|
| Dependency updates | Weekly | Check for security updates |
| Log rotation | Daily | Configure in Kubernetes |
| Backup verification | Weekly | Test restoration |
| Security scan | Daily | CI/CD pipeline |

## Troubleshooting

### Proxy won't start

1. Check logs: `kubectl logs -n hexstrike-system`
2. Verify config: `kubectl get configmap`
3. Check secrets exist

### High latency

1. Check rate limits: `/metrics`
2. Verify Lakera API connectivity
3. Check network policies

### Request rejections

1. Check JWT validity
2. Verify rate limits
3. Review security policies

---

*Generated from operations*
