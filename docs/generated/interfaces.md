# Interfaces & Integrations

## Component Interfaces

| Module | Interface | Methods | Purpose |
|--------|-----------|---------|---------|
| lakera | LakeraChecker | 3 methods | Interface |

## HTTP Endpoints

| Path | Method | Handler | Auth | Description |
|------|--------|---------|------|-------------|
| /health | GET | healthHandler | No | API endpoint |
| /ready | GET | readyHandler | No | API endpoint |
| /metrics | GET | promHandler | No | API endpoint |
| / | GET | proxy | No | API endpoint |
| /* | POST | proxy.Handler() | Yes | MCP proxy |