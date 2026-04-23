# Operational Runbooks

## Emergency Response

### High CPU Usage

```bash
# Check pod resources
kubectl top pod -n hexstrike-system

# Check for runaway processes
kubectl exec -it <pod> -- top

# Restart if needed
kubectl rollout restart deployment/mcp-proxy -n hexstrike-system
```

### Memory Exhaustion

```bash
# Check OOM kills
kubectl describe pod <pod> | grep -A 5 "Last State"

# Increase memory limit
kubectl patch deployment mcp-proxy -n hexstrike-system -p '{...}'
```

### Service Unavailable

```bash
# Check all pods
kubectl get pods -n hexstrike-system

# Check events
kubectl get events -n hexstrike-system --sort-by='.lastTimestamp'

# Restart deployment
kubectl rollout restart deployment/mcp-proxy -n hexstrike-system
```

### Rate Limit Errors

```bash
# Check current limits
curl http://mcp-proxy:8080/metrics | grep rate_limit

# Adjust if needed
kubectl patch configmap mcp-config -n hexstrike-system
```

---

*Generated from operational procedures*
