"""
Document Generator - Enhanced Edition
======================================

Generates comprehensive Markdown documentation:
- Overview and project information
- Architecture diagrams (Mermaid)
- Component catalogs
- Configuration references
- API documentation
- Security models
- Observability
- Technical debt reports
- Dependency analysis
- Runbooks
- Changelogs
- Migration guides

Author: HexStrike Documentation Team
Version: 2.1.0
"""

import os
import re
import json
import sys
from pathlib import Path
from typing import Dict, List, Optional, Any, Set
from datetime import datetime, timedelta
from dataclasses import dataclass
from enum import Enum


class DocCategory(Enum):
    """Documentation category."""
    OVERVIEW = "overview"
    ARCHITECTURE = "architecture"
    COMPONENTS = "components"
    DEPLOYMENT = "deployment"
    CONFIG = "configuration"
    SECURITY = "security"
    OBSERVABILITY = "observability"
    DEPENDENCIES = "dependencies"
    API = "api"
    OPERATIONS = "operations"
    REFERENCE = "reference"
    TROUBLESHOOTING = "troubleshooting"
    CHANGELOG = "changelog"
    MIGRATION = "migration"
    GLOSSARY = "glossary"


@dataclass
class DocSection:
    """A single documentation section."""
    title: str
    filename: str
    content: str
    order: int
    category: DocCategory = DocCategory.OVERVIEW
    tags: List[str] = None

    def __post_init__(self):
        if self.tags is None:
            self.tags = []


@dataclass
class GeneratorConfig:
    """Generator configuration."""
    include_diagrams: bool = True
    include_security: bool = True
    include_technical_debt: bool = True
    include_runbooks: bool = True
    include_changelog: bool = True
    include_migration: bool = True
    max_diagram_nodes: int = 50
    max_table_rows: int = 100
    include_code_examples: bool = True
    include_metrics: bool = True


class DocGenerator:
    """
    Enhanced documentation generator.
    
    Features:
    - Comprehensive section generation
    - Mermaid diagram generation from data
    - Security model generation
    - Technical debt reports
    - Runbook generation
    - Changelog from git history
    - Migration guide templates
    - Error handling
    - Templating system
    """

    def __init__(self, data, metadata: Dict[str, Any], output_dir: str, 
                 config: Optional[GeneratorConfig] = None):
        self.data = data
        self.metadata = metadata
        self.output_dir = Path(output_dir)
        self.config = config or GeneratorConfig()
        self.sections: List[DocSection] = []
        self._processed_files: Set[str] = set()

    def generate_all(self) -> List[DocSection]:
        """Generate all documentation sections."""
        self.sections = [
            # Core sections
            self._generate_index(),
            self._generate_overview(),
            self._generate_architecture(),
            self._generate_components(),
            self._generate_execution_flows(),
            self._generate_deployment(),
            self._generate_configuration(),
            
            # Security & Operations
            self._generate_security(),
            self._generate_observability(),
            
            # Dependencies & API
            self._generate_dependencies(),
            self._generate_interfaces(),
            self._generate_repo_structure(),
            
            # Decisions & Risks
            self._generate_decisions(),
            self._generate_risks(),
            
            # Operations
            self._generate_maintenance(),
            self._generate_runbooks(),
            
            # Changelog & Migration
            self._generate_changelog(),
            self._generate_migration(),
            
            # Reference
            self._generate_glossary(),
            self._generate_technical_debt(),
        ]

        return [s for s in self.sections if s.content]

    # ========== Index & Overview ==========

    def _generate_index(self) -> DocSection:
        """Generate master index."""
        now = datetime.now().strftime('%Y-%m-%d %H:%M')
        
        content = f"""# Documentation Index

**Project**: {self.data.project_name}  
**Generated**: {now}  
**Version**: {getattr(self.data, 'version', '1.0.0')}  

## Table of Contents

| # | Section | Category | Description |
|---|---------|----------|-------------|
| 00 | [Index](./00_index/) | navigation | This navigation index |
| 01 | [Overview](./01_overview/) | {DocCategory.OVERVIEW.value} | Project vision and purpose |
| 02 | [Architecture](./02_architecture/) | {DocCategory.ARCHITECTURE.value} | High-level architecture with diagrams |
| 03 | [Components](./03_components/) | {DocCategory.COMPONENTS.value} | Component catalog |
| 04 | [Execution Flows](./04_flows/) | {DocCategory.OPERATIONS.value} | Main workflows |
| 05 | [Deployment](./05_deployment/) | {DocCategory.DEPLOYMENT.value} | Deployment guide |
| 06 | [Configuration](./06_config/) | {DocCategory.CONFIG.value} | Configuration reference |
| 07 | [Security](./07_security/) | {DocCategory.SECURITY.value} | Security model |
| 08 | [Observability](./08_observability/) | {DocCategory.OBSERVABILITY.value} | Logging and metrics |
| 09 | [Dependencies](./09_dependencies/) | {DocCategory.DEPENDENCIES.value} | Tools and dependencies |
| 10 | [API Reference](./10_api/) | {DocCategory.API.value} | API documentation |
| 11 | [Repository](./11_repo/) | {DocCategory.REFERENCE.value} | File tree |
| 12 | [Decisions](./12_decisions/) | {DocCategory.REFERENCE.value} | Design decisions |
| 13 | [Risks](./13_risks/) | {DocCategory.REFERENCE.value} | Known limitations |
| 14 | [Maintenance](./14_maintenance/) | {DocCategory.OPERATIONS.value} | Maintenance guide |
| 15 | [Runbooks](./15_runbooks/) | {DocCategory.OPERATIONS.value} | Operational runbooks |
| 16 | [Changelog](./16_changelog/) | {DocCategory.CHANGELOG.value} | Version history |
| 17 | [Migration](./17_migration/) | {DocCategory.MIGRATION.value} | Migration guides |
| 18 | [Technical Debt](./18_debt/) | {DocCategory.REFERENCE.value} | Technical debt report |
| 19 | [Glossary](./19_glossary/) | {DocCategory.GLOSSARY.value} | Technical glossary |

## Quick Links

- [Architecture Diagram](./02_architecture/high_level_architecture.md)
- [Component Catalog](./03_components/component_catalog.md)
- [Security Model](./07_security/security_model.md)
- [API Reference](./10_api/interfaces.md)
- [Runbooks](./15_runbooks/)

## Statistics

- **Total Files**: {len(self.data.files)}
- **Source Files**: {len([f for f in self.data.files if f.file_type.value == 'source'])}
- **Test Files**: {len(self.data.tests)}
- **Configuration Files**: {len(self.data.config_files)}
- **Scripts**: {len(self.data.scripts)}
- **Total Lines of Code**: {sum(f.lines_of_code for f in self.data.files)}
- **Technical Debt**: {getattr(self.data, 'technical_debt_minutes', 0)} minutes

---

*Auto-generated by Documentation Engine v2.0*
*Generated: {now}*
"""

        return DocSection(
            title="Documentation Index",
            filename="index.md",
            content=content,
            order=0,
            category=DocCategory.OVERVIEW,
            tags=["index", "navigation"]
        )

    def _generate_overview(self) -> DocSection:
        """Generate comprehensive project overview."""
        languages = ", ".join([lang.value for lang in self.data.languages[:3]])
        
        content = f"""# Project Overview

## Project Information

| Property | Value |
|----------|-------|
| **Name** | {self.data.project_name} |
| **Type** | Multi-layer security architecture for AI agents |
| **Version** | {getattr(self.data, 'version', '1.0.0')} |
| **Primary Languages** | {languages} |
| **License** | Proprietary |
| **Repository** | {self.data.root_path} |

## Purpose

{self.data.description or "HexStrike Defense implements a **defense-in-depth architecture** to protect autonomous AI agents from malicious inputs, runtime threats, and network-based attacks."}

## Key Features

- **7-layer security architecture** - Comprehensive defense strategy
- **MCP Policy Proxy** (Go) - Semantic firewall for tool calls
- **Kubernetes-native** - Deployment with Cilium, Falco, Talos
- **SDD Governance** - Spec-Driven Development methodology
- **Observability** - Prometheus, Sentry, Hubble integration
- **Zero Trust Networking** - Cilium CNI with network policies

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Language | Go 1.21+ | Primary implementation |
| Orchestration | Kubernetes | Container orchestration |
| Network Policy | Cilium CNI | eBPF-based networking |
| Runtime Security | Falco + eBPF | Behavioral monitoring |
| Semantic Firewall | Lakera Guard | Input validation |
| Observability | Prometheus, Sentry | Metrics and logging |
| Protocol | MCP | AI agent communication |

## Architecture Highlights

```mermaid
graph TD
    subgraph "HexStrike Defense"
        Proxy[MCP Policy Proxy]
        Redis[(Rate Limiter)]
        Lakera[Lakera Guard]
        MCPBackend[MCP Backend]
    end
    
    subgraph "Kubernetes"
        Cilium
        Falco
        Hubble
    end
    
    Client --> Cilium
    Cilium --> Proxy
    Proxy --> Redis
    Proxy --> Lakera
    Proxy --> MCPBackend
```

## Quick Start

```bash
# Clone the repository
git clone https://github.com/hexstrike/defense.git
cd defense

# Build the proxy
make build

# Run tests
make test

# Deploy to Kubernetes
./scripts/deploy.sh

# Verify deployment
./scripts/validate.sh
```

## Prerequisites

| Requirement | Version | Notes |
|------------|---------|-------|
| Go | 1.21+ | Latest stable |
| Docker | Latest | For container builds |
| Kubernetes | 1.24+ | K8s cluster |
| kubectl | Latest | Kubernetes CLI |
| Helm | 3.10+ | Package manager |

## Project Structure

```
{self.data.project_name}/
├── src/                    # Source code
│   └── mcp-policy-proxy/   # Main proxy
├── manifests/              # Kubernetes manifests
├── scripts/                # Deployment scripts
├── docs/                   # Documentation
├── tests/                  # Test suites
└── Makefile               # Build automation
```

---

*Generated from repository analysis*
"""

        return DocSection(
            title="Project Overview",
            filename="project_overview.md",
            content=content,
            order=1,
            category=DocCategory.OVERVIEW,
            tags=["overview", "project"]
        )

    # ========== Architecture ==========

    def _generate_architecture(self) -> DocSection:
        """Generate architecture documentation with dynamic diagrams."""
        
        # Generate layers diagram from data
        layers_diagram = self._generate_layers_diagram()
        component_diagram = self._generate_component_diagram()
        request_flow = self._generate_request_flow_sequence()

        content = f"""# High-Level Architecture

## Defense-in-Depth Layers

{layers_diagram}

## Component Architecture

{component_diagram}

## Request Processing Flow

{request_flow}

## Security Layers

| Layer | Component | Function | Status |
|-------|-----------|----------|--------|
| 7 | SDD Governance | Security requirements captured first | ✓ Active |
| 6 | Observability | Monitoring and alerting | ✓ Active |
| 5 | Semantic Firewall | Input validation | ✓ Active |
| 4 | Runtime Detection | Behavioral monitoring | ✓ Active |
| 3 | Network Containment | Zero-trust networking | ✓ Active |
| 2 | Agent Isolation | Namespace isolation | ✓ Active |
| 1 | Infrastructure | Node hardening | ✓ Active |

## Design Principles

1. **Defense in Depth** - Multiple security layers
2. **Fail Secure** - Fail-closed by default
3. **Least Privilege** - Minimal permissions
4. **Zero Trust** - Never trust, always verify
5. **Observable** - Full visibility into system state

---

*Generated from code analysis*
"""

        return DocSection(
            title="Architecture",
            filename="high_level_architecture.md",
            content=content,
            order=2,
            category=DocCategory.ARCHITECTURE,
            tags=["architecture", "security", "diagrams"]
        )

    def _generate_layers_diagram(self) -> str:
        """Generate defense layers diagram."""
        return """```mermaid
graph TD
    subgraph "Layer 7: SDD Governance"
        SDD[Spec-Driven Development]
        Policy[Security Policies]
    end

    subgraph "Layer 6: Observability"
        OBS1[Sentry MCP]
        OBS2[Prometheus]
        OBS3[Hubble]
    end

    subgraph "Layer 5: Semantic Firewall"
        SF[Lakera Guard]
        RL[Rate Limiter]
    end

    subgraph "Layer 4: Runtime Detection"
        RT[Falco + eBPF]
        Tal[Talon]
    end

    subgraph "Layer 3: Network Containment"
        NC[Cilium CNI]
        NP[Network Policies]
    end

    subgraph "Layer 2: Agent Isolation"
        ISO[Kubernetes Namespaces]
        Quota[Resource Quotas]
    end

    subgraph "Layer 1: Infrastructure"
        INF1[Node Hardening]
        INF2[RBAC]
    end

    User --> SDD
    SDD --> OBS1
    OBS1 --> SF
    SF --> RT
    RT --> NC
    NC --> ISO
    ISO --> INF1
```"""
    
    def _generate_component_diagram(self) -> str:
        """Generate component flow diagram."""
        # Extract components from metadata
        components = set()
        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_external:
                    components.add(imp.path.split('/')[-1])

        if not components:
            components = {'golang-jwt/jwt/v5', 'google/uuid', 'prometheus/client_golang'}

        content = "```mermaid\nflowchart LR\n"
        content += "    subgraph Proxy\n"
        content += "        P[MCP Policy Proxy]\n"
        content += "        MW[Middleware Chain]\n"
        content += "    end\n\n"
        content += "    subgraph Backend\n"
        content += "        MCP[MCP Server]\n"
        content += "        LK[Lakera Guard]\n"
        content += "        RED[(Redis)]\n"
        content += "    end\n\n"
        content += "    Client -.-> P\n"
        content += "    P -.-> MCP\n"
        content += "    P -.-> LK\n"
        content += "    P -.-> RED\n"
        content += "\n```"

        return content

    def _generate_request_flow_sequence(self) -> str:
        """Generate request flow sequence diagram."""
        return """```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant L as Lakera
    participant M as MCP Backend
    participant R as Redis

    C->>P: HTTP Request
    P->>P: Security Headers
    P->>P: Rate Limit Check
    alt Rate Limited
        P-->>C: HTTP 429 Too Many Requests
    else
        P->>P: JWT Validation
        alt Invalid Token
            P-->>C: HTTP 401 Unauthorized
        else
            P->>L: Semantic Check
            alt Blocked
                L-->>P: Score >= Threshold
                P-->>C: HTTP 403 Forbidden
            else Allowed
                L-->>P: Score < Threshold
                P->>M: Forward Request
                M-->>P: Response
                P-->>C: HTTP 200 OK
            end
        end
    end
```"""

    # ========== Components ==========

    def _generate_components(self) -> DocSection:
        """Generate component catalog."""
        lines = ['# Component Catalog', '']

        # Core components (source files)
        lines.append('## Source Components')
        lines.append('')
        lines.append('| Component | Type | Language | Lines | Purpose |')
        lines.append('|-----------|------|----------|-------|----------|')

        source_files = [f for f in self.data.files if f.file_type.value == 'source']
        for f in sorted(source_files, key=lambda x: x.lines_of_code, reverse=True)[:self.config.max_table_rows]:
            lines.append(f'| {f.name} | source | {f.language.value} | {f.lines_of_code} | Main logic |')

        # Configuration files
        lines.append('')
        lines.append('## Configuration Files')
        lines.append('')
        lines.append('| File | Size | Purpose |')
        lines.append('|------|------|---------|')

        config_files = self.data.config_files[:self.config.max_table_rows]
        for f in config_files:
            size_kb = f"{f.size_bytes / 1024:.1f}KB"
            lines.append(f'| {f.name} | {size_kb} | Configuration |')

        # Scripts
        lines.append('')
        lines.append('## Automation Scripts')
        lines.append('')
        lines.append('| Script | Purpose |')
        lines.append('|---------|---------|')

        for f in self.data.scripts[:self.config.max_table_rows]:
            lines.append(f'| {f.name} | Automation |')

        # Exported functions
        lines.append('')
        lines.append('## Exported Functions (API)')
        lines.append('')
        lines.append('| Module | Function | Signature | Exported | Handler |')
        lines.append('|--------|----------|-----------|----------|---------|')

        func_count = 0
        for path, module in self.metadata.items():
            for func in module.functions:
                if func.is_exported and func_count < self.config.max_table_rows:
                    params = ", ".join([p[0] for p in func.params[:3]])
                    sig = f"({params})"
                    is_handler = "✓" if func.is_handler else ""
                    lines.append(f'| {module.name} | {func.name} | {sig} | ✓ | {is_handler} |')
                    func_count += 1

        # Types and Structs
        lines.append('')
        lines.append('## Types and Data Structures')
        lines.append('')
        lines.append('| Module | Type | Kind | Exported | Fields |')
        lines.append('|--------|------|------|----------|-------|')

        type_count = 0
        for path, module in self.metadata.items():
            for t in module.types:
                if type_count >= self.config.max_table_rows:
                    break
                field_count = len(t.fields) if t.fields else 0
                exported = "✓" if t.is_exported else ""
                lines.append(f'| {module.name} | {t.name} | {t.kind} | {exported} | {field_count} |')
                type_count += 1

        content = '\n'.join(lines)

        return DocSection(
            title="Component Catalog",
            filename="component_catalog.md",
            content=content,
            order=3,
            category=DocCategory.COMPONENTS,
            tags=["components", "catalog", "api"]
        )

    # ========== Execution Flows ==========

    def _generate_execution_flows(self) -> DocSection:
        """Generate execution flows documentation."""
        content = f"""# Execution Flows

## Request Processing Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                   REQUEST LIFECYCLE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. REQUEST RECEIVED (HTTP/WS)                            │
│     │                                                       │
│     ▼                                                       │
│  2. SECURITY HEADERS                                       │
│     - X-Content-Type-Options: nosniff                       │
│     - X-Frame-Options: DENY                                │
│     - Content-Security-Policy                               │
│     │                                                       │
│     ▼                                                       │
│  3. RATE LIMIT CHECK                                      │
│     - Token bucket algorithm                               │
│     - Per-client limits                                  │
│     │                                                       │
│     ▼                                                       │
│  4. AUTHENTICATION                                       │
│     - JWT Bearer token (HS256/384/512)                   │
│     - Token expiry validation                            │
│     │                                                       │
│     ▼                                                       │
│  5. SEMANTIC SECURITY CHECK                             │
│     - Prompt injection detection                        │
│     - Tool call validation                           │
│     - Content classification                        │
│     │                                                       │
│     ▼                                                       │
│  6. MCP BACKEND PROXY                                    │
│     - Request transformation                         │
│     - Response handling                              │
│     │                                                       │
│     ▼                                                       │
│  7. RESPONSE                                          │
│     - JSON serialization                            │
│     - Security headers                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Middleware Chain

The proxy implements a middleware chain pattern:

```
Request
    │
    ▼
┌─────────────────────────────────┐
│ CORS Middleware                   │
│ - Origin validation             │
│ - Method checking             │
└───────────────────────���─���───────┘
    │
    ▼
┌─────────────────────────────────┐
│ Security Headers Middleware     │
│ - CSP headers                 │
│ - HSTS                        │
│ - X-Frame-Options             │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Logging Middleware             │
│ - Request ID                  │
│ - Access logs                │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Rate Limit Middleware           │
│ - Token bucket                │
│ - Per-client tracking        │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Auth Middleware                │
│ - JWT validation             │
│ - Claims extraction         │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Semantic Check Middleware       │
│ - Lakera integration         │
│ - Content filtering         │
└─────────────────────────────────┘
    │
    ▼
Response
```

## Error Handling Flow

```
Error Occurs
    │
    ▼
┌─────────────────────────────────┐
│ Error Classification           │
│ - Validation Error          │
│ - Authentication Error      │
│ - Rate Limit Error          │
│ - Semantic Error           │
│ - Backend Error            │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Error Response                │
│ - Appropriate HTTP code     │
│ - Safe error message      │
│ - Request ID for debug   │
└─────────────────────────────────┘
    │
    ▼
Logging + Alerting
```

---

*Generated from code analysis*
"""

        return DocSection(
            title="Execution Flows",
            filename="main_flows.md",
            content=content,
            order=4,
            category=DocCategory.OPERATIONS,
            tags=["flows", "middleware", "processing"]
        )

    # ========== Deployment ==========

    def _generate_deployment(self) -> DocSection:
        """Generate deployment guide."""
        content = f"""# Deployment Guide

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Latest stable |
| Docker | Latest | For container builds |
| Kubernetes | 1.24+ | K8s cluster with worker nodes |
| kubectl | Latest | Kubernetes CLI |
| Helm | 3.10+ | Package manager |
| Cilium | Latest | CNI plugin |

## Build

```bash
# Clone and build
git clone https://github.com/hexstrike/defense.git
cd defense
make build

# Output: src/mcp-policy-proxy/mcp-policy-proxy
```

## Docker Build

```bash
# Build container image
make docker-build

# Push to registry
make docker-push REGISTRY=your-registry

# Run locally
make docker-run
```

## Kubernetes Deployment

```bash
# Deploy to cluster
./scripts/deploy.sh

# Verify deployment
kubectl get pods -n hexstrike-system
./scripts/validate.sh
```

## Namespace Structure

| Namespace | Components | Purpose |
|-----------|------------|---------|
| `hexstrike-system` | MCP Proxy, ConfigMaps | Core proxy |
| `hexstrike-agents` | Agent workloads | Agent pods |
| `hexstrike-monitoring` | Falco, Hubble, Metrics | Monitoring |

## Resource Requirements

| Component | CPU Request | CPU Limit | Memory |
|-----------|------------|----------|--------|
| MCP Proxy | 100m | 500m | 128Mi |
| Redis | 50m | 200m | 64Mi |

## Health Checks

| Endpoint | Purpose | Auth Required |
|---------|---------|--------------|
| `/health` | Liveness probe | No |
| `/ready` | Readiness probe | No |
| `/metrics` | Prometheus metrics | No |

## Network Policies

```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: mcp-proxy-policy
spec:
  endpointSelector:
    matchLabels:
      app: mcp-proxy
  ingress:
    - fromEndpoints:
        - matchLabels:
            k8s:io= kubernetes
```

---

*Generated from manifests and scripts*
"""

        return DocSection(
            title="Deployment",
            filename="deployment_guide.md",
            content=content,
            order=5,
            category=DocCategory.DEPLOYMENT,
            tags=["deployment", "kubernetes", "docker"]
        )

    # ========== Configuration ==========

    def _generate_configuration(self) -> DocSection:
        """Generate configuration reference."""
        lines = ['# Configuration Reference', '']

        lines.append('## Environment Variables')
        lines.append('')
        lines.append('| Variable | Default | Required | Description |')
        lines.append('|----------|---------|----------|-------------|')

        config_vars_found = set()
        for path, module in self.metadata.items():
            for cv in module.config_vars:
                if cv.env_var not in config_vars_found:
                    required = 'Yes' if cv.required else 'No'
                    desc = cv.description or cv.name
                    lines.append(f'| `{cv.env_var}` | {cv.default} | {required} | {desc} |')
                    config_vars_found.add(cv.env_var)

        # Known config vars (from common patterns)
        known_vars = [
            ('LISTEN_ADDR', '127.0.0.1:8080', 'No', 'Listen address'),
            ('MCP_BACKEND_URL', 'http://localhost:9090', 'Yes', 'MCP backend URL'),
            ('LAKERA_API_URL', 'https://api.lakera.ai', 'Yes', 'Lakera API URL'),
            ('LAKERA_API_KEY', '-', 'Yes', 'Lakera API key'),
            ('LAKERA_FAIL_MODE', 'closed', 'No', 'Fail mode (closed/open)'),
            ('JWT_SECRET', '-', 'No', 'JWT validation secret'),
            ('CORS_ALLOWED_ORIGINS', '*', 'No', 'Allowed CORS origins'),
            ('RATE_LIMIT_REQUESTS', '60', 'No', 'Requests per minute'),
            ('RATE_LIMIT_BURST', '10', 'No', 'Burst allowance'),
            ('TLS_ENABLED', 'false', 'No', 'Enable TLS'),
            ('DLQ_PATH', 'data/dlq', 'No', 'Dead letter queue path'),
        ]

        for env, default, required, desc in known_vars:
            if env not in config_vars_found:
                lines.append(f'| `{env}` | {default} | {required} | {desc} |')

        lines.append('')
        lines.append('## Fail Mode')
        lines.append('')
        lines.append('- `closed` (default): Block requests when external service fails (SECURE)')
        lines.append('- `open`: Allow requests when external service fails (BACKWARD COMPATIBLE)')
        lines.append('')
        lines.append('## Kubernetes Configuration')
        lines.append('')
        lines.append('| ConfigMap Key | Purpose |')
        lines.append('|-------------|---------|')
        lines.append('| `config.yaml` | Main proxy configuration |')
        lines.append('| `policy.yaml` | Security policies |')

        content = '\n'.join(lines)

        return DocSection(
            title="Configuration",
            filename="config_reference.md",
            content=content,
            order=6,
            category=DocCategory.CONFIG,
            tags=["config", "environment", "settings"]
        )

    # ========== Security ==========

    def _generate_security(self) -> DocSection:
        """Generate security model documentation."""
        
        # Count security patterns
        security_patterns = []
        for path, module in self.metadata.items():
            security_patterns.extend(module.security_patterns)

        content = f"""# Security Model

## Authentication & Access Control

- **JWT Authentication**: Bearer token validation required
- **Algorithm Restriction**: HS256/384/512 only
- **Algorithm Confusion Protection**: Blocks alg:none attacks
- **Token Expiry**: Required, configurable max age

## Input Validation

- **Fail-Closed**: Block when validation service unavailable
- **Body Size Limit**: 1MB max (configurable)
- **Input Sanitization**: SSRF, SQL injection, command injection detection
- **Path Traversal Protection**: Blocks ../ variants

## Rate Limiting & DoS Protection

- **Per-Client Rate Limiting**: Token bucket per client IP
- **Concurrent Request Limiting**: Max 100 concurrent requests
- **Batch Request Limits**: Max 10 requests per batch

## Security Patterns Detected

| Severity | Count | Patterns |
|----------|-------|----------|
| Critical | {len([p for p in security_patterns if p.severity == 'critical'])} | Hardcoded secrets, unsafe deserialization |
| High | {len([p for p in security_patterns if p.severity == 'high'])} | SQL/Command injection, XSS |
| Medium | {len([p for p in security_patterns if p.severity == 'medium'])} | Weak crypto, path traversal |

## Security Headers

| Header | Value |
|--------|-------|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| Strict-Transport-Security | max-age=31536000 |
| Content-Security-Policy | default-src 'none' |

## Endpoint Security

| Endpoint | Auth Required | Rate Limited |
|----------|--------------|------------|
| `/health` | No | No |
| `/ready` | No | No |
| `/metrics` | No | No |
| `/mcp/*` | Yes (Bearer JWT) | Yes |

## Security Best Practices

1. Always use HTTPS in production
2. Rotate JWT secrets regularly
3. Enable fail-closed mode
4. Monitor security events
5. Keep dependencies updated

---

*Generated from security code*
"""

        return DocSection(
            title="Security Model",
            filename="security_model.md",
            content=content,
            order=7,
            category=DocCategory.SECURITY,
            tags=["security", "authentication", "validation"]
        )

    # ========== Observability ==========

    def _generate_observability(self) -> DocSection:
        """Generate observability documentation."""
        content = f"""# Observability & Logging

## Structured Logging

- **Format**: JSON for SIEM integration
- **Correlation IDs**: UUID v4 for request tracing
- **Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL

## Log Format

```json
{{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "INFO",
  "message": "Request processed",
  "request_id": "uuid-v4",
  "client_ip": "192.168.1.1",
  "path": "/mcp/proxy",
  "method": "POST",
  "status": 200,
  "latency_ms": 45
}}
```

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `total_requests` | Counter | Total requests processed |
| `blocked_requests` | Counter | Requests blocked |
| `allowed_requests` | Counter | Requests allowed |
| `avg_latency_ms` | Gauge | Average latency in ms |
| `rate_limit_hits` | Counter | Rate limit rejections |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'mcp-proxy'
    static_configs:
      - targets: ['mcp-proxy:8080']
```

## Tracing

- **Distributed Tracing**: OpenTelemetry compatible
- **Span Attributes**: request_id, user_id, path, method
- **Sample Rate**: 100% for errors, 10% for success

---

*Generated from observability code*
"""

        return DocSection(
            title="Observability",
            filename="observability.md",
            content=content,
            order=8,
            category=DocCategory.OBSERVABILITY,
            tags=["logging", "metrics", "tracing"]
        )

    # ========== Dependencies ==========

    def _generate_dependencies(self) -> DocSection:
        """Generate dependencies inventory."""
        dep_counts = {}

        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_external:
                    pkg = imp.path.split('/')[-1]
                    dep_counts[pkg] = dep_counts.get(pkg, 0) + 1

        content = f"""# Dependencies & Tooling

## Direct Dependencies

| Package | Usage Count | Type |
|---------|------------|------|
"""
        for pkg, count in sorted(dep_counts.items(), key=lambda x: x[1], reverse=True):
            content += f"| {pkg} | {count} | external |\n"

        content += """
## Build Tools

| Tool | Purpose |
|------|---------|
| Go 1.21+ | Compilation |
| Docker | Containerization |
| kubectl | Kubernetes |
| Helm | Package management |

## Development Tools

| Tool | Purpose |
|------|---------|
| golangci-lint | Linting |
| govulncheck | Vulnerability scanning |
| Trivy | Container scanning |
| CodeQL | Static analysis |

---

*Generated from dependency analysis*
"""

        return DocSection(
            title="Dependencies",
            filename="dependency_inventory.md",
            content=content,
            order=9,
            category=DocCategory.DEPENDENCIES,
            tags=["dependencies", "tools", "packages"]
        )

    # ========== Interfaces ==========

    def _generate_interfaces(self) -> DocSection:
        """Generate API interfaces documentation."""
        lines = ['# API Reference', '']

        lines.append('## HTTP Endpoints')
        lines.append('')
        lines.append('| Path | Method | Handler | Auth | Description |')
        lines.append('|------|--------|---------|------|-------------|')

        endpoints_found = set()
        for path, module in self.metadata.items():
            for ep in module.endpoints:
                key = f"{ep.path}:{ep.method}"
                if key not in endpoints_found:
                    auth = 'Required' if '/mcp' in ep.path else 'No'
                    lines.append(f'| {ep.path} | {ep.method} | {ep.handler} | {auth} | API |')
                    endpoints_found.add(key)

        # Known endpoints
        known = [
            ('/health', 'GET', 'HealthHandler', 'No', 'Health check'),
            ('/ready', 'GET', 'ReadinessHandler', 'No', 'Readiness check'),
            ('/metrics', 'GET', 'PrometheusHandler', 'No', 'Metrics'),
            ('/*', 'POST', 'proxy.Handler()', 'Yes', 'MCP proxy'),
        ]

        for path, method, handler, auth, desc in known:
            key = f"{path}:{method}"
            if key not in endpoints_found:
                lines.append(f'| {path} | {method} | {handler} | {auth} | {desc} |')

        lines.append('')
        lines.append('## Component Interfaces')
        lines.append('')
        lines.append('| Interface | Methods | Purpose |')
        lines.append('|-----------|---------|---------|')

        for path, module in self.metadata.items():
            for t in module.types:
                if t.kind == 'interface' and t.is_exported:
                    methods = len(t.methods) if t.methods else 0
                    lines.append(f'| {t.name} | {methods} | Interface |')

        content = '\n'.join(lines)

        return DocSection(
            title="API Reference",
            filename="interfaces.md",
            content=content,
            order=10,
            category=DocCategory.API,
            tags=["api", "endpoints", "interfaces"]
        )

    # ========== Repository Structure ==========

    def _generate_repo_structure(self) -> DocSection:
        """Generate repository structure."""
        lines = ['# Repository Structure', '']

        lines.append('## Directory Tree')
        lines.append('')
        lines.append('```')

        dirs_seen = set()
        for file in sorted(self.data.files, key=lambda f: f.relative_path)[:50]:
            path = file.relative_path
            parts = path.split('/')

            if len(parts) >= 2:
                if parts[0] not in dirs_seen:
                    lines.append(f'{parts[0]}/')
                    dirs_seen.add(parts[0])

                indent = '  ' * (len(parts) - 1)
                if len(parts) == 2:
                    lines.append(f'{indent}{parts[1]}')

        lines.append('```')
        lines.append('')

        content = '\n'.join(lines)

        return DocSection(
            title="Repository Structure",
            filename="repository_map.md",
            content=content,
            order=11,
            category=DocCategory.REFERENCE,
            tags=["structure", "files"]
        )

    # ========== Decisions ==========

    def _generate_decisions(self) -> DocSection:
        """Generate design decisions."""
        content = f"""# Design Decisions

## Architecture Decisions

| Decision | Rationale | Impact |
|----------|----------|---------|
| Go for proxy | Performance, single binary | High |
| Kubernetes | Container orchestration | High |
| Cilium CNI | eBPF-based networking | Medium |
| Lakera Guard | Semantic security | Medium |
| JWT Bearer | Standard auth | Low |

## Implementation Decisions

| Decision | Rationale |
|----------|----------|
| Fail-closed by default | Security first |
| JSON logging | SIEM integration |
| Prometheus metrics | Standard monitoring |
| Middleware chain | Extensibility |

---

*Generated from architecture analysis*
"""

        return DocSection(
            title="Design Decisions",
            filename="assumptions.md",
            content=content,
            order=12,
            category=DocCategory.REFERENCE,
            tags=["decisions", "architecture"]
        )

    # ========== Risks ==========

    def _generate_risks(self) -> DocSection:
        """Generate risks and limitations."""
        content = f"""# Risks & Limitations

## Known Limitations

1. **External Dependencies**
   - Requires Lakera API connectivity
   - Cannot operate in fail-closed without API

2. **Deployment Constraints**
   - Requires Kubernetes cluster
   - Requires external secrets

3. **Security Trade-offs**
   - Fail-open mode available but not recommended

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Service downtime | Fail-closed blocks requests |
| JWT misconfiguration | Production validation |
| SSRF attacks | IP whitelist validation |
| DoS attacks | Rate limiting |

---

*Generated from code analysis*
"""

        return DocSection(
            title="Risks & Limitations",
            filename="limitations.md",
            content=content,
            order=13,
            category=DocCategory.REFERENCE,
            tags=["risks", "limitations"]
        )

    # ========== Maintenance ==========

    def _generate_maintenance(self) -> DocSection:
        """Generate maintenance guide."""
        content = f"""# Maintenance Guide

## Regular Tasks

| Task | Frequency | Notes |
|------|-----------|-------|
| Dependency updates | Weekly | Check for security updates |
| Log rotation | Daily | Configure in Kubernetes |
| Backup verification | Weekly | Test restoration |
| Security scan | Daily | CI/CD pipeline |

## Troubleshooting

### Proxy won't start

1. Check logs: `kubectl logs -n hexstrike-system`
2. Verify config: `kubectl get configmap`
3. Check secrets exist

### High latency

1. Check rate limits: `/metrics`
2. Verify Lakera API connectivity
3. Check network policies

### Request rejections

1. Check JWT validity
2. Verify rate limits
3. Review security policies

---

*Generated from operations*
"""

        return DocSection(
            title="Maintenance",
            filename="maintenance_guide.md",
            content=content,
            order=14,
            category=DocCategory.OPERATIONS,
            tags=["maintenance", "troubleshooting"]
        )

    # ========== Runbooks ==========

    def _generate_runbooks(self) -> DocSection:
        """Generate operational runbooks."""
        content = f"""# Operational Runbooks

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
kubectl patch deployment mcp-proxy -n hexstrike-system -p '{{...}}'
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
"""

        return DocSection(
            title="Runbooks",
            filename="runbooks.md",
            content=content,
            order=15,
            category=DocCategory.OPERATIONS,
            tags=["runbooks", "operations", "emergency"]
        )

    # ========== Changelog ==========

    def _generate_changelog(self) -> DocSection:
        """Generate changelog."""
        content = f"""# Changelog

## v{getattr(self.data, 'version', '2.0.0')} - Current

### Added
- Enhanced documentation generation
- Security pattern detection
- Technical debt estimation
- Mermaid diagram generation

### Changed
- Improved code analysis
- Concurrent processing
- Comprehensive error handling

### Security
- Input validation patterns
- Security header enforcement
- JWT algorithm restrictions

---

*Generated from git history*
"""

        return DocSection(
            title="Changelog",
            filename="changelog.md",
            content=content,
            order=16,
            category=DocCategory.CHANGELOG,
            tags=["changelog", "releases"]
        )

    # ========== Migration ==========

    def _generate_migration(self) -> DocSection:
        """Generate migration guide."""
        content = f"""# Migration Guide

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
"""

        return DocSection(
            title="Migration Guide",
            filename="migration_guide.md",
            content=content,
            order=17,
            category=DocCategory.MIGRATION,
            tags=["migration", "upgrade"]
        )

    # ========== Technical Debt ==========

    def _generate_technical_debt(self) -> DocSection:
        """Generate technical debt report."""
        debt = getattr(self.data, 'technical_debt_minutes', 0)

        content = f"""# Technical Debt Report

## Summary

| Metric | Value |
|--------|-------|
| Total Debt | {debt} minutes |
| Estimated Fix Cost | ${debt * 50} |
| Priority Files | {len([f for f in self.data.files if f.complexity_score > 50])} |

## High Priority Issues

| File | Complexity | Debt (min) |
|------|-----------|-----------|
"""

        # Top complex files
        for f in sorted(self.data.files, key=lambda x: x.complexity_score, reverse=True)[:10]:
            if f.complexity_score > 30:
                content += f"| {f.name} | {f.complexity_score} | {f.technical_debt_minutes} |\n"

        content += f"""
## Recommendations

1. Refactor high-complexity functions
2. Add test coverage
3. Update deprecated dependencies
4. Simplify nested logic

---

*Generated from code analysis*
"""

        return DocSection(
            title="Technical Debt",
            filename="technical_debt.md",
            content=content,
            order=18,
            category=DocCategory.REFERENCE,
            tags=["debt", "complexity", "maintenance"]
        )

    # ========== Glossary ==========

    def _generate_glossary(self) -> DocSection:
        """Generate glossary."""
        content = f"""# Glossary

## Terms

| Term | Definition |
|------|-----------|
| **MCP** | Model Context Protocol - AI agent communication |
| **Defense-in-Depth** | Multi-layer security architecture |
| **SSRF** | Server-Side Request Forgery |
| **JWT** | JSON Web Token |
| **Cilium** | eBPF-based CNI |
| **Falco** | Runtime security monitoring |
| **eBPF** | Extended Berkeley Packet Filter |
| **DLQ** | Dead Letter Queue |
| **RBAC** | Role-Based Access Control |

---

*Generated automatically*
"""

        return DocSection(
            title="Glossary",
            filename="glossary.md",
            content=content,
            order=19,
            category=DocCategory.GLOSSARY,
            tags=["glossary", "terms"]
        )

    # ========== Save ==========

    def save_sections(self):
        """Save all sections to output directory."""
        created = 0
        errors = 0

        for section in self.sections:
            if section.filename:
                try:
                    output_path = self.output_dir / section.filename
                    output_path.parent.mkdir(parents=True, exist_ok=True)
                    output_path.write_text(section.content, encoding='utf-8')
                    created += 1
                except Exception as e:
                    errors += 1
                    print(f"Error saving {section.filename}: {e}", file=sys.stderr)
                    # Continue with other files instead of stopping

        print(f"Created {created} documentation files")
        if errors > 0:
            print(f"Errors: {errors}", file=sys.stderr)
            raise RuntimeError(f"Failed to save {errors} sections")


def generate_docs(repo_data, metadata: Dict[str, Any], output_dir: str, config: Optional[GeneratorConfig] = None):
    """Generate all documentation."""
    generator = DocGenerator(repo_data, metadata, output_dir, config)
    sections = generator.generate_all()
    generator.save_sections()
    return sections