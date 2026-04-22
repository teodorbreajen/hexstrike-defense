# Documentation Index

> **⚠️ NOTE**: This index is for navigation within the documentation section.
> For the master project index, see [docs/README.md](../README.md) or [docs/generated/index.md](../generated/index.md).

## Section Index

| Section | Description | File |
|---------|-------------|------|
| [Overview](./project_overview.md) | Project vision and purpose | 01_overview |
| [Architecture](./high_level_architecture.md) | High-level architecture | 03_architecture |
| [Components](./component_catalog.md) | Component catalog | 04_components |
| [Flows](./main_flows.md) | Main execution flows | 05_execution_flows |
| [Deployment](./deployment_guide.md) | Deployment guide | 06_deployment |
| [Configuration](./config_reference.md) | Configuration reference | 07_configuration |
| [Security](./security_model.md) | Security model | 08_security_and_constraints |
| [Observability](./observability.md) | Observability & logging | 09_observability_and_logging |
| [Dependencies](./dependency_inventory.md) | Dependencies inventory | 10_dependencies_and_tooling |
| [Interfaces](./interfaces.md) | API interfaces | 11_interfaces_and_integrations |
| [Structure](./repository_map.md) | Repository structure | 12_repo_structure |
| [Decisions](./assumptions.md) | Design decisions | 13_decisions_and_assumptions |
| [Risks](./limitations.md) | Known limitations | 14_risks_and_limitations |
| [Glossary](./glossary.md) | Technical glossary | 16_glossary |

## Quick Reference

### Key Files

| File | Purpose |
|------|---------|
| `main.go` | Proxy entry point |
| `proxy.go` | Main proxy logic |
| `lakera.go` | Lakera client |
| `dlq/` | Dead Letter Queue |

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | Listen address |
| `MCP_BACKEND_URL` | - | MCP backend URL |
| `LAKERA_API_KEY` | - | Lakera API key |
| `JWT_SECRET` | - | JWT secret |
| `LAKERA_FAIL_MODE` | `closed` | Fail mode |

### Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/health` | GET | No | Health check |
| `/ready` | GET | No | Readiness |
| `/metrics` | GET | No | Prometheus metrics |
| `/mcp/*` | POST | Yes | MCP proxy |

---

*Auto-generated documentation index*