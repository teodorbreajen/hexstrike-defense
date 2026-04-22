# Configuration Reference

## Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `EXPOSE_HEALTH` | false | No | exposeHealth |
| `DEV_MODE` | false | No | devMode |
| `GIN_MODE` | debug | No | ginMode |
| `CORS_ALLOWED_ORIGINS` |  | Yes | corsOriginsStr |
| `LISTEN_ADDR` | 127.0.0.1:8080 | No | Listen address |
| `MCP_BACKEND_URL` | http://localhost:9090 | Yes | MCP backend URL |
| `LAKERA_API_URL` | https://api.lakera.ai | Yes | Lakera API URL |
| `LAKERA_API_KEY` | - | Yes | Lakera API key |
| `LAKERA_FAIL_MODE` | closed | No | Fail mode (closed/open) |
| `JWT_SECRET` | - | No | JWT validation secret |
| `CORS_ALLOW_CREDENTIALS` | false | No | Allow credentials in CORS |
| `TLS_ENABLED` | false | No | Enable TLS/HTTPS |
| `DLQ_PATH` | data/dlq | No | Dead letter queue path |
| `TRUSTED_PROXIES` | - | No | Trusted proxy IPs/CIDR |

## Fail Mode

- `closed` (default): Block requests when Lakera fails (SECURE)
- `open`: Allow requests when Lakera fails (BACKWARD COMPATIBLE)