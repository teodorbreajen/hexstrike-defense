# API Reference

## HTTP Endpoints

| Path | Method | Handler | Auth | Description |
|------|--------|---------|------|-------------|
| /health | GET | HealthHandler | No | API |
| /ready | GET | ReadinessHandler | No | API |
| /metrics | GET | proxy | No | API |
| / | GET | proxy | No | API |
| /* | POST | proxy.Handler() | Yes | MCP proxy |

## Component Interfaces

| Interface | Methods | Purpose |
|-----------|---------|---------|