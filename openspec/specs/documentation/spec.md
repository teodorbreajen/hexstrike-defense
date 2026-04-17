# Documentation Specification (Delta)

## Purpose

Define incident response runbook requirements for production operations.

## ADDED Requirements

### Requirement: Incident Runbook Structure

The runbook SHALL be located at `docs/runbook.md` and follow Given/When/Then scenario format.

**Required Sections:**
1. Lakera Outage Response
2. Falco Alert Response
3. Rate Limit Exceeded Response
4. General Troubleshooting

### Requirement: Lakera Outage Response Runbook

The runbook SHALL document procedures for Lakera service unavailability.

#### Scenario: Lakera API returns 503

- GIVEN Lakera API is returning HTTP 503 errors
- WHEN requests are being blocked with 503
- THEN verify fail mode setting:
  - If `LAKERA_FAIL_MODE=closed`: This is expected behavior (security)
  - If `LAKERA_FAIL_MODE=open`: Proceed to step 3
- THEN check Lakera service status at `status.lakera.ai`
- THEN if Lakera is down, consider temporary switch to `LAKERA_FAIL_MODE=open`
- THEN after Lakera recovers, switch back to `LAKERA_FAIL_MODE=closed`

#### Scenario: Lakera timeout errors

- GIVEN Lakera API requests are timing out
- THEN verify network connectivity to Lakera endpoints
- THEN check if timeout threshold is too aggressive: `LAKERA_TIMEOUT`
- THEN consider increasing timeout from 5s to 15s temporarily
- THEN monitor error rate after adjustment

#### Scenario: Enable fail-open during extended outage

- GIVEN Lakera is confirmed down for >15 minutes
- WHEN business requires continued operation
- THEN set environment variable `LAKERA_FAIL_MODE=open`
- THEN restart proxy service
- THEN document decision and timeline in incident ticket
- THEN monitor for malicious tool calls in logs
- THEN revert to fail-closed immediately when Lakera recovers

### Requirement: Falco Alert Response Runbook

The runbook SHALL document procedures for Falco security alerts.

#### Scenario: Falco alerts on suspicious MCP tool call

- GIVEN Falco generates alert for `mcp_policy_proxy` container
- THEN identify the correlation_id from alert
- THEN query proxy logs: `correlation_id=<id>`
- THEN extract tool name, arguments, and source IP
- THEN block source IP if malicious: update firewall rules
- THEN notify security team with incident details
- THEN if false positive, update Falco rules to exclude

#### Scenario: Falco container escape alert

- GIVEN Falco alerts on namespace change or suspicious process
- THEN isolate container immediately
- THEN escalate to security incident response
- THEN preserve container logs and forensics
- THEN do NOT restart container before investigation

### Requirement: Rate Limit Exceeded Response Runbook

The runbook SHALL document procedures for rate limit issues.

#### Scenario: Legitimate traffic exceeding rate limit

- GIVEN monitoring shows rate limit 429 errors for valid users
- THEN verify rate limit is appropriate for use case
- THEN if adjustment needed, update `RATE_LIMIT_PER_MINUTE` environment variable
- THEN restart proxy to apply new limit
- THEN monitor for improvement

#### Scenario: Attack traffic causing rate limit

- GIVEN sudden spike in 429 errors from single IP
- THEN check access logs for source IP
- THEN if IP is not whitelisted, block at firewall level
- THEN add IP to blocklist in configuration
- THEN verify blocking effectiveness

### Requirement: General Troubleshooting Procedures

#### Scenario: High latency on all requests

- GIVEN requests are slow (>1s average)
- THEN check metrics endpoint for backend latency
- THEN verify MCP backend is responding normally
- THEN check proxy CPU/memory usage
- THEN review network connectivity

#### Scenario: Memory usage growing continuously

- GIVEN container memory usage is increasing
- THEN check for body size limit violations (should be limited)
- THEN verify no goroutine leaks (check with `/metrics`)
- THEN restart container as temporary fix
- THEN investigate root cause

### Requirement: Configuration Reference

The runbook SHALL include complete configuration reference.

**Required Environment Variables:**
| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `0.0.0.0:8080` | Listen address |
| `MCP_BACKEND_URL` | `http://localhost:9090` | MCP backend URL |
| `LAKERA_API_KEY` | (empty) | Lakera API key |
| `LAKERA_URL` | `https://api.lakera.ai` | Lakera API URL |
| `LAKERA_TIMEOUT` | `5` | Lakera timeout (seconds) |
| `LAKERA_FAIL_MODE` | `closed` | Fail behavior: closed/open |
| `RATE_LIMIT_PER_MINUTE` | `60` | Rate limit per minute |
| `LOG_LEVEL` | `info` | Log level: debug/info/warn/error |
| `MAX_BODY_SIZE` | `1048576` | Max body size (bytes) |
| `AUTH_ENABLED` | `true` | Enable authentication |
| `JWT_SECRET` | (empty) | **REQUIRED in production** - JWT signing secret |
| `GIN_MODE` | `debug` | Set to `release` for production (enables JWT fail-hard) |
| `CORS_ALLOWED_ORIGINS` | (empty) | Comma-separated allowed origins, or "disabled" |
| `CORS_ALLOW_CREDENTIALS` | `false` | Allow credentials in CORS responses |
