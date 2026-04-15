# hexstrike-defense

Defense-in-Depth architecture for securing autonomous AI agents (hexstrike-ai).

## Overview

This project implements a multi-layer security architecture to protect autonomous AI agents from malicious inputs, runtime threats, and network-based attacks.

## Architecture Layers

1. **SDD Governance** - Spec-driven development ensures security requirements are captured first
2. **Semantic Firewall** - Lakera Guard / NeMo Guardrails for input validation
3. **Runtime Detection** - Falco + eBPF + Talon for behavioral monitoring
4. **Network Containment** - Cilium CNI for zero-trust networking

## Tech Stack

- **Orchestration**: LangGraph
- **Protocol**: MCP (Model Context Protocol)
- **Runtime Security**: Falco, eBPF, Talon
- **Network Policies**: Cilium CNI
- **Observability**: Atlassian MCP, Sentry MCP
- **Infrastructure**: Kubernetes (kind/minikube for dev)

## Project Structure

```
hexstrike-defense/
├── openspec/           # SDD governance
│   ├── config.yaml
│   ├── specs/          # Source of truth
│   └── changes/        # Active changes
├── manifests/         # Kubernetes manifests
│   ├── cilium/
│   ├── falco/
│   └── mcp-proxy/
├── scripts/           # Automation scripts
├── docs/              # Documentation
└── .atl/              # Skill registry
```

## SDD Workflow

This project follows Spec-Driven Development:

1. **Explore** - Investigate requirements (sdd-explore)
2. **Propose** - Create change proposal (sdd-propose)
3. **Spec** - Write detailed specs (sdd-spec)
4. **Design** - Technical design (sdd-design)
5. **Tasks** - Break into tasks (sdd-tasks)
6. **Apply** - Implement (sdd-apply)
7. **Verify** - Validate against specs (sdd-verify)
8. **Archive** - Sync to main specs (sdd-archive)

## Getting Started

```bash
# Set up kind cluster for local testing
./scripts/setup-kind.sh

# Apply base manifests
kubectl apply -k manifests/
```

## Documentation

See `docs/` for detailed architecture documentation.