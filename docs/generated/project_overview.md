# Project Overview

## Project Information

| Property | Value |
|----------|-------|
| **Name** | hexstrike-defense |
| **Type** | Multi-layer security architecture for AI agents |
| **Version** | 2.0.0 |
| **Primary Language** | Go 1.21+ |
| **License** | Proprietary |

## Purpose

HexStrike Defense implements a **defense-in-depth architecture** to protect autonomous AI agents from:
- Malicious inputs and prompt injection
- Runtime threats and behavioral attacks
- Network-based attacks and data exfiltration

## Key Features

- **7-layer security architecture**
- **MCP Policy Proxy** (Go) - Semantic firewall for tool calls
- **Kubernetes-native** deployment with Cilium, Falco, Talon
- **SDD Governance** - Spec-Driven Development
- **Observability** - Prometheus, Sentry, Hubble

## Technology Stack

| Layer | Technology |
|------|-----------|
| Language | Go 1.21+ |
| Orchestration | Kubernetes |
| Network Policy | Cilium CNI |
| Runtime Security | Falco + eBPF |
| Semantic Firewall | Lakera Guard |
| Observability | Prometheus, Sentry |
| Protocol | MCP (Model Context Protocol) |

## Quick Start

```bash
# Build the proxy
make build

# Run tests
make test

# Deploy to Kubernetes
./scripts/deploy.sh
```

---

*Generated from repository analysis*
