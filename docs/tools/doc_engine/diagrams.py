"""
Diagram Generator - Enhanced Edition
============================

Advanced Mermaid diagram generation:
- Dynamic from real data
- Multiple diagram types
- Security visualization
- API documentation
- Component relationships
- Data flow diagrams
- Sequence diagrams
- State diagrams

Author: HexStrike Documentation Team
Version: 2.1.0
"""

import os
import re
import hashlib
import logging
from pathlib import Path
from typing import Dict, List, Set, Optional, Any, Tuple
from dataclasses import dataclass
from enum import Enum

logger = logging.getLogger(__name__)


class DiagramType(Enum):
    """Mermaid diagram types."""
    FLOWCHART = "flowchart"
    SEQUENCE = "sequence"
    CLASS = "classDiagram"
    STATE = "stateDiagram"
    ENTITY_RELATIONSHIP = "er"
    JOURNEY = "journey"
    GANT = "gantt"
    PIE = "pie"
    MINDMAP = "mindmap"
    TIMELINE = "timeline"


class DiagramStyle(Enum):
    """Diagram visual styles."""
    DEFAULT = "default"
    SECURITY = "security"
    DEPENDENCY = "dependency"
    DATA_FLOW = "data-flow"
    COMPONENT = "component"
    ARCHITECTURE = "architecture"


@dataclass
class Diagram:
    """A Mermaid diagram."""
    title: str
    diagram_type: DiagramType
    style: DiagramStyle
    content: str
    filename: str
    description: str = ""
    tags: List[str] = None

    def __post_init__(self):
        if self.tags is None:
            self.tags = []


@dataclass
class NodeData:
    """Data for a node in diagram."""
    id: str
    label: str
    node_type: str = "process"  # process, service, database, external
    group: str = ""
    security_level: str = "public"


@dataclass
class EdgeData:
    """Data for an edge in diagram."""
    source: str
    target: str
    label: str = ""
    style: str = ""  # solid, dotted, thick


class DiagramGenerator:
    """
    Enhanced diagram generator with dynamic data.
    
    Features:
    - Dynamic from extracted data
    - Multiple diagram types
    - Security visualization
    - API documentation
    - Component relationships
    - Concurrent request flows
    """

    def __init__(self, repo_data, metadata: Dict[str, Any], 
                 config: Optional[Dict] = None):
        self.repo_data = repo_data
        self.metadata = metadata
        self.config = config or self._default_config()
        self.nodes: Dict[str, NodeData] = {}
        self.edges: List[EdgeData] = []

    def _default_config(self) -> Dict:
        """Default configuration."""
        return {
            "max_nodes": 50,
            "include_tests": False,
            "show_complexity": True,
            "include_security": True,
            "show_dependencies": True,
        }

    def generate_all(self) -> List[Diagram]:
        """Generate all applicable diagrams."""
        diagrams = []

        # Core architectural diagrams
        diagrams.append(self._generate_architecture_diagram())
        diagrams.append(self._generate_component_flow())
        
        # Dependency diagrams
        if self.config.get("show_dependencies", True):
            diagrams.append(self._generate_dependency_graph())
            diagrams.append(self._generate_import_graph())
        
        # Flow diagrams
        diagrams.append(self._generate_request_flow())
        diagrams.append(self._generate_middleware_flow())
        
        # Security diagrams
        if self.config.get("show_security", True):
            diagrams.append(self._generate_security_layers())
        
        # Component diagrams
        diagrams.append(self._generate_class_diagram())
        
        # API diagrams
        diagrams.append(self._generate_api_flow())
        
        # Test coverage
        if self.config.get("include_tests", False):
            diagrams.append(self._generate_test_diagram())

        return [d for d in diagrams if d.content]

    # ========== Architecture Diagrams ==========

    def _generate_architecture_diagram(self) -> Diagram:
        """Generate high-level architecture diagram."""
        content = """```mermaid
---
title: Defense-in-Depth Architecture
---
graph TD
    subgraph "Layer 7: Governance"
        SDD["SDD Governance"]
        Policy["Security Policies"]
    end

    subgraph "Layer 6: Observability"
        OBS1["Sentry"]
        OBS2["Prometheus"]
        OBS3["Hubble"]
    end

    subgraph "Layer 5: Semantic Firewall"
        SF["Lakera Guard"]
        RL["Rate Limiter"]
    end

    subgraph "Layer 4: Runtime"
        RT["Falco + eBPF"]
        Tal["Talon"]
    end

    subgraph "Layer 3: Network"
        NC["Cilium CNI"]
        NP["Network Policies"]
    end

    subgraph "Layer 2: Isolation"
        ISO["K8s Namespaces"]
        Quota["Resource Quotas"]
    end

    subgraph "Layer 1: Infrastructure"
        RBAC["RBAC"]
        Hard["Node Hardening"]
    end

    Client --> SDD
    SDD --> OBS1
    OBS1 --> SF
    SF --> RL
    RL --> RT
    RT --> NC
    NC --> NP
    NP --> ISO
    ISO --> Quota
    Quota --> RBAC
    RBAC --> Hard
```"""

        return Diagram(
            title="Defense-in-Depth Architecture",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.ARCHITECTURE,
            content=content,
            filename="architecture.mmd",
            description="High-level security architecture",
            tags=["architecture", "security", "layers"]
        )

    def _generate_component_flow(self) -> Diagram:
        """Generate component flow diagram."""
        # Collect components from metadata
        components = set()
        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_external:
                    components.add(imp.path.split('/')[-1])

        if not components:
            logger.warning("No components found - using minimal diagram")
            components = set()  # Empty is fine, diagram will be minimal

        content = "```mermaid\n"
        content += "flowchart LR\n"

        # External dependencies
        content += "    subgraph External\n"
        for comp in list(components)[:10]:
            content += f'        {comp.upper()}["{comp}"]\n'
        content += "    end\n\n"

        # Core components
        content += "    subgraph Core\n"
        content += '        PROXY["MCP Proxy"]\n'
        content += '        MW["Middleware"]\n'
        content += "    end\n\n"

        # Backend
        content += "    subgraph Backend\n"
        content += '        LAKERA["Lakera"]\n'
        content += '        REDIS["Redis"]\n'
        content += '        MCP["MCP Server"]\n'
        content += "    end\n\n"

        # Connections
        content += "    CLIENT -.-> PROXY\n"
        content += "    PROXY -.-> MW\n"
        content += "    PROXY -.-> LAKERA\n"
        content += "    PROXY -.-> REDIS\n"
        content += "    PROXY -.-> MCP\n"

        content += "\n```"

        return Diagram(
            title="Component Flow",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.COMPONENT,
            content=content,
            filename="component_flow.mmd",
            description="Component relationships",
            tags=["components", "flow"]
        )

    # ========== Dependency Diagrams ==========

    def _generate_dependency_graph(self) -> Diagram:
        """Generate dependency graph."""
        deps: Dict[str, Set[str]] = {}

        for path, module in self.metadata.items():
            module_deps = set()
            for imp in module.imports:
                if imp.is_external:
                    module_deps.add(imp.path.split('/')[-1])
            if module_deps:
                deps[module.name] = module_deps

        if not deps:
            logger.warning("No dependencies found - using minimal diagram")
            deps = {'proxy': set()}

        content = "```mermaid\n"
        content += "flowchart BT\n"
        content += "    style PROXY fill:#f9f,stroke:#333,stroke-width:2px\n"

        content += "    subgraph Dependencies\n"
        for module, module_deps in deps.items():
            for dep in module_deps:
                content += f'        {module} --> {dep.upper()}\n'
        content += "    end\n"
        content += "\n```"

        return Diagram(
            title="Dependency Graph",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.DEPENDENCY,
            content=content,
            filename="dependency_graph.mmd",
            description="Dependency relationships",
            tags=["dependencies", "imports"]
        )

    def _generate_import_graph(self) -> Diagram:
        """Generate import hierarchy graph."""
        
        content = "```mermaid\n"
        content += "mindmap\n"
        content += "  root((HexStrike Defense))\n"

        # Group imports by type
        stdlib = set()
        external = set()

        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_standard:
                    stdlib.add(imp.path.split('/')[0])
                elif imp.is_external:
                    external.add(imp.path.split('/')[-1])

        if stdlib:
            content += "    Standard Library\n"
            for s in sorted(list(stdlib)[:10]):
                content += f"      - {s}\n"

        if external:
            content += "    External\n"
            for e in sorted(list(external)[:10]):
                content += f"      - {e}\n"

        content += "```"

        return Diagram(
            title="Import Hierarchy",
            diagram_type=DiagramType.MINDMAP,
            style=DiagramStyle.DEPENDENCY,
            content=content,
            filename="import_hierarchy.mmd",
            description="Import organization",
            tags=["imports", "structure"]
        )

    # ========== Flow Diagrams ==========

    def _generate_request_flow(self) -> Diagram:
        """Generate request processing sequence."""
        content = """```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant RL as Rate Limiter
    participant A as Auth
    participant L as Lakera
    participant M as MCP Backend
    participant R as Redis

    C->>P: POST /mcp/proxy
    P->>P: Security Headers
    P->>RL: Check Rate Limit
    alt Limited
        RL-->>P: 429 Too Many
        P-->>C: HTTP 429
    else
        P->>A: Validate JWT
        alt Invalid
            A-->>P: 401 Unauthorized
            P-->>C: HTTP 401
        else
            P->>L: Check Semantic
            alt Blocked
                L-->>P: Score >= Threshold
                P-->>C: HTTP 403 Forbidden
            else Allowed
                L-->>P: Score < Threshold
                P->>M: Forward Request
                M-->>P: Response
                P->>R: Cache Response
                P-->>C: HTTP 200 OK
            end
        end
    end
```"""

        return Diagram(
            title="Request Processing Flow",
            diagram_type=DiagramType.SEQUENCE,
            style=DiagramStyle.DATA_FLOW,
            content=content,
            filename="request_flow.mmd",
            description="Request lifecycle",
            tags=["request", "processing", "flow"]
        )

    def _generate_middleware_flow(self) -> Diagram:
        """Generate middleware chain diagram."""
        content = """```mermaid
flowchart TB
    subgraph "Middleware Chain"
        M1[CORS]
        M2[Security Headers]
        M3[Logging]
        M4[Rate Limit]
        M5[Auth]
        M6[Semantic Check]
        M7[Metrics]
    end

    Request --> M1
    M1 --> M2
    M2 --> M3
    M3 --> M4
    M4 --> M5
    M5 --> M6
    M6 --> M7
    M7 --> Response

    style M1 fill:#ff9,stroke:#333
    style M6 fill:#f96,stroke:#333,color:#fff
```"""

        return Diagram(
            title="Middleware Chain",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.DATA_FLOW,
            content=content,
            filename="middleware_flow.mmd",
            description="Middleware processing",
            tags=["middleware", "chain"]
        )

    # ========== Security Diagrams ==========

    def _generate_security_layers(self) -> Diagram:
        """Generate security layer visualization."""
        content = """```mermaid
graph TD
    subgraph "Security Layers"
        L1[{"layer": "1", "name": "Infrastructure"}]
        L2[{"layer": "2", "name": "Isolation"}]
        L3[{"layer": "3", "name": "Network"}]
        L4[{"layer": "4", "name": "Runtime"}]
        L5[{"layer": "5", "name": "Semantic"}]
        L6[{"layer": "6", "name": "Observability"}]
    end

    Attack --> L1
    Attack --> L2
    Attack --> L3
    Attack --> L4
    Attack --> L5
    Attack --> L6

    L1 -->|"blocked"| Blocked
    L2 -->|"blocked"| Blocked
    L3 -->|"blocked"| Blocked
    L4 -->|"detected"| Blocked
    L5 -->|"blocked"| Blocked
    L6 -->|"alerted"| Monitored

    style Attack fill:#f00,color:#fff
    style Blocked fill:#f00,color:#fff
    style Monitored fill:#ff0
```"""

        return Diagram(
            title="Security Layers",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.SECURITY,
            content=content,
            filename="security_layers.mmd",
            description="Security visualization",
            tags=["security", "layers"]
        )

    # ========== Class Diagrams ==========

    def _generate_class_diagram(self) -> Diagram:
        """Generate class diagram for main types."""
        content = "```mermaid\n"
        content += "classDiagram\n"

        # Get exported types
        for path, module in self.metadata.items():
            for t in module.types:
                if t.is_exported:
                    content += f'    class {t.name} {{\n'

                    # Fields
                    for field in t.fields[:5]:
                        name, ftype, tag = field
                        content += f'        +{name}: {ftype}\n'

                    content += '    }\n\n'

                    # Methods
                    for method in (t.methods or [])[:5]:
                        content += f'    {t.name} : +{method}()\n'

        if not any(t.is_exported for m in self.metadata.values() for t in m.types):
            content += "    class ProxyConfig\n"
            content += "        +backendURL: string\n"
            content += "        +rateLimit: int\n"

        content += "```"

        return Diagram(
            title="Class Diagram",
            diagram_type=DiagramType.CLASS,
            style=DiagramStyle.COMPONENT,
            content=content,
            filename="class_diagram.mmd",
            description="Type definitions",
            tags=["classes", "types", "api"]
        )

    # ========== API Diagrams ==========

    def _generate_api_flow(self) -> Diagram:
        """Generate API endpoint flow."""
        content = "```mermaid\n"
        content += "flowchart TD\n"

        # Collect endpoints
        endpoints = []
        for path, module in self.metadata.items():
            endpoints.extend(module.endpoints)

        if not endpoints:
            endpoints = [
                {"path": "/health", "method": "GET"},
                {"path": "/mcp/*", "method": "POST"},
            ]

        for ep in endpoints[:10]:
            node_id = ep.path.replace('/', '_').replace('*', 'ALL')
            node_id = node_id.strip('_')
            method = ep.method
            
            if method == "GET":
                style = "fill:#9f9"
            elif method == "POST":
                style = "fill:#99f"
            elif method == "PUT":
                style = "fill:#ff9"
            elif method == "DELETE":
                style = "fill:#f99"
            else:
                style = "fill:#999"

            content += f'    {node_id}["{ep.path}"]\n'
            content += f'    style {node_id} {style},stroke:#333\n'

        content += "\n```"

        return Diagram(
            title="API Endpoints",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.DATA_FLOW,
            content=content,
            filename="api_flow.mmd",
            description="API endpoints",
            tags=["api", "endpoints"]
        )

    # ========== Test Diagrams ==========

    def _generate_test_diagram(self) -> Diagram:
        """Generate test coverage diagram."""
        test_files = [f for f in self.repo_data.files if '_test' in f.name]

        if not test_files:
            return Diagram(
                title="Test Coverage",
                diagram_type=DiagramType.FLOWCHART,
                style=DiagramStyle.DEFAULT,
                content="",
                filename="test_coverage.mmd",
                description="No tests found",
                tags=["tests"]
            )

        content = "```mermaid\n"
        content += "pie\n"
        content += "    title Test Coverage\n"

        # Count test types
        test_types = {}
        for tf in test_files:
            name = tf.name
            if '_test.go' in name:
                test_types['Go Tests'] = test_types.get('Go Tests', 0) + 1
            elif '_test.py' in name:
                test_types['Python Tests'] = test_types.get('Python Tests', 0) + 1
            elif '_test.ts' in name:
                test_types['TypeScript Tests'] = test_types.get('TypeScript Tests', 0) + 1
            else:
                test_types['Other'] = test_types.get('Other', 0) + 1

        for test_type, count in test_types.items():
            content += f'    "{test_type}": {count}\n'

        content += "```"

        return Diagram(
            title="Test Coverage",
            diagram_type=DiagramType.PIE,
            style=DiagramStyle.DEFAULT,
            content=content,
            filename="test_coverage.mmd",
            description="Test file distribution",
            tags=["tests", "coverage"]
        )

    # ========== Export ==========

    def save_diagrams(self, output_dir: str):
        """Save all diagrams to output directory."""
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)

        created = 0
        for diagram in self.generate_all():
            if diagram.content:
                file_path = output_path / diagram.filename
                file_path.write_text(diagram.content, encoding='utf-8')
                created += 1

        return created


def generate_diagrams(repo_data, metadata: Dict[str, Any], output_dir: str, config: Optional[Dict] = None) -> List[Diagram]:
    """Generate all diagrams."""
    generator = DiagramGenerator(repo_data, metadata, config)
    diagrams = generator.generate_all()
    generated = generator.save_diagrams(output_dir)
    print(f"Generated {generated} diagram files")
    return diagrams