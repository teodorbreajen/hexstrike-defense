# Migration Guide

## Upgrading from v1.x to v2.0

### Breaking Changes

1. **Configuration**: Some env vars renamed
2. **API**: Response format changed
3. **Metrics**: New metric names

### Migration Steps

1. Review new configuration options
2. Update environment variables
3. Test with fail-open mode first
4. Monitor metrics during rollout

### Rollback Plan

```bash
# Rollback to previous version
kubectl rollout undo deployment/mcp-proxy -n hexstrike-system
```

---

*Generated from version analysis*
