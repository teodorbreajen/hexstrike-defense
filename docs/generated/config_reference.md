# Configuration Reference

## Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `JWT_SECRET` |  | Yes | Environment variable JWT_SECRET |
| `DLQ_Enqueue_SavesFile` |  | Yes | Environment variable DLQ_Enqueue_SavesFile |
| `DLQ_Replay_ExecutesHandler` |  | Yes | Environment variable DLQ_Replay_ExecutesHandler |
| `DLQ_Replay_FIFOOrder` |  | Yes | Environment variable DLQ_Replay_FIFOOrder |
| `DLQ_Size_CountsFiles` |  | Yes | Environment variable DLQ_Size_CountsFiles |
| `DLQ_Cleanup_EliminatesExpiredMessages` |  | Yes | Environment variable DLQ_Cleanup_EliminatesExpiredMessages |
| `DLQ_CleanupWithTTL` |  | Yes | Environment variable DLQ_CleanupWithTTL |
| `DLQ_Enqueue_GeneratesID` |  | Yes | Environment variable DLQ_Enqueue_GeneratesID |
| `DLQ_Enqueue_SetsTimestamp` |  | Yes | Environment variable DLQ_Enqueue_SetsTimestamp |
| `DLQ_Remove_DeletesMessage` |  | Yes | Environment variable DLQ_Remove_DeletesMessage |
| `DLQ_Remove_NotFound` |  | Yes | Environment variable DLQ_Remove_NotFound |
| `DLQ_Peek_ReturnsOldestMessage` |  | Yes | Environment variable DLQ_Peek_ReturnsOldestMessage |
| `DLQ_Peek_EmptyQueue` |  | Yes | Environment variable DLQ_Peek_EmptyQueue |
| `DLQ_GetMessages_ReturnsAllMessages` |  | Yes | Environment variable DLQ_GetMessages_ReturnsAllMessages |
| `DLQ_Enqueue_NilRequest` |  | Yes | Environment variable DLQ_Enqueue_NilRequest |
| `DLQ_Replay_NilHandler` |  | Yes | Environment variable DLQ_Replay_NilHandler |
| `DLQ_NewDLQ_CreatesDirectory` |  | Yes | Environment variable DLQ_NewDLQ_CreatesDirectory |
| `DLQ_Replay_ContextCancellation` |  | Yes | Environment variable DLQ_Replay_ContextCancellation |
| `DLQ_SetLogger` |  | Yes | Environment variable DLQ_SetLogger |
| `MCP_PROXY_URL` |  | Yes | Environment variable MCP_PROXY_URL |
| `LAKERA_API_KEY` |  | Yes | Environment variable LAKERA_API_KEY |
| `LISTEN_ADDR` | 127.0.0.1:8080 | No | Listen address |
| `MCP_BACKEND_URL` | http://localhost:9090 | Yes | MCP backend URL |
| `LAKERA_API_URL` | https://api.lakera.ai | Yes | Lakera API URL |
| `LAKERA_FAIL_MODE` | closed | No | Fail mode (closed/open) |
| `CORS_ALLOWED_ORIGINS` | * | No | Allowed CORS origins |
| `RATE_LIMIT_REQUESTS` | 60 | No | Requests per minute |
| `RATE_LIMIT_BURST` | 10 | No | Burst allowance |
| `TLS_ENABLED` | false | No | Enable TLS |
| `DLQ_PATH` | data/dlq | No | Dead letter queue path |

## Fail Mode

- `closed` (default): Block requests when external service fails (SECURE)
- `open`: Allow requests when external service fails (BACKWARD COMPATIBLE)

## Kubernetes Configuration

| ConfigMap Key | Purpose |
|-------------|---------|
| `config.yaml` | Main proxy configuration |
| `policy.yaml` | Security policies |