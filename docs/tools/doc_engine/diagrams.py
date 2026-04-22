"""
Diagram Generator
================

Generates Mermaid diagrams automatically from code analysis.
Creates:
- Architecture diagrams
- Component relationships
- Data flow diagrams
- Dependency graphs
- Sequence diagrams

Author: HexStrike Documentation Team
"""

import os
from pathlib import Path
from typing import Dict, List, Set, Optional
from dataclasses import dataclass


@dataclass
class Diagram:
    """A Mermaid diagram."""
    title: str
    diagram_type: str  # flowchart, sequence, class, state
    content: str
    filename: str


class DiagramGenerator:
    """Generates Mermaid diagrams from code analysis."""

    def __init__(self, repo_data, metadata: Dict):
        self.repo_data = repo_data
        self.metadata = metadata

    def generate_all(self) -> List[Diagram]:
        """Generate all applicable diagrams."""
        diagrams = []

        diagrams.append(self._generate_architecture_diagram())
        diagrams.append(self._generate_component_flow())
        diagrams.append(self._generate_dependency_graph())
        diagrams.append(self._generate_request_flow())
        diagrams.append(self._generate_test_diagram())

        return [d for d in diagrams if d.content]

    def _generate_architecture_diagram(self) -> Diagram:
        """Generate high-level architecture diagram."""
        content = """```mermaid
flowchart TB
    subgraph "Layer 7: SDD Governance"
        SDD[Spec-Driven Development]
    end

    subgraph "Layer 6: Observability"
        OBS[Sentry | Prometheus]
    end

    subgraph "Layer 5: Semantic Firewall"
        SF[Lakera Guard | Rate Limiter]
    end

    subgraph "Layer 4: Runtime Detection"
        RT[Falco | Talon]
    end

    subgraph "Layer 3: Network"
        NET[Cilium CNI]
    end

    subgraph "Layer 2: Isolation"
        ISO[Kubernetes NS]
    end

    subgraph "Layer 1: Infra"
        INF[RBAC | Hardening]
    end

    User --> SF
    SF --> RT
    RT --> NET
    NET --> ISO
    ISO --> INF
```"""
        return Diagram(
            title="Defense-in-Depth Architecture",
            diagram_type="flowchart",
            content=content,
            filename="architecture.mmd"
        )

    def _generate_component_flow(self) -> Diagram:
        """Generate component flow diagram."""
        # Extract main components
        components = set()
        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_external:
                    components.add(imp.path.split('/')[-1])

        # Generate flowchart
        content = "```mermaid\nflowchart LR\n"

        # Add MCP Proxy
        content += "    subgraph Proxy\n"
        content += "        P[MCP Policy Proxy]\n"
        content += "    end\n\n"

        # Add backend services
        content += "    subgraph Backend\n"
        content += "        MCP[MCP Server]\n"
        content += "        LK[Lakera Guard]\n"
        content += "    end\n\n"

        # Add connections
        content += "    Client --> P\n"
        content += "    P --> MCP\n"
        content += "    P --> LK\n"

        content += "\n```"

        return Diagram(
            title="Component Flow",
            diagram_type="flowchart",
            content=content,
            filename="component_flow.mmd"
        )

    def _generate_dependency_graph(self) -> Diagram:
        """Generate dependency graph."""
        deps = set()

        # Collect all imports
        for path, module in self.metadata.items():
            for imp in module.imports:
                if imp.is_external:
                    deps.add(imp.path)

        if not deps:
            deps = {'golang-jwt/jwt', 'google/uuid', 'prometheus/client_golang'}

        content = "```mermaid\nflowchart TB\n"
        content += "    subgraph Dependencies\n"

        for dep in sorted(deps):
            dep_name = dep.split('/')[-1]
            content += f'        {dep_name}["{dep}"]\n'

        content += "    end\n"
        content += "```"

        return Diagram(
            title="Dependency Graph",
            diagram_type="flowchart",
            content=content,
            filename="dependencies.mmd"
        )

    def _generate_request_flow(self) -> Diagram:
        """Generate request flow sequence diagram."""
        content = """```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant L as Lakera
    participant M as MCP Backend

    C->>P: HTTP Request
    P->>P: Validate JWT
    P->>P: Rate Limit Check
    P->>L: Check Tool Call
    alt Allowed
        L->>P: Score < Threshold
        P->>M: Forward Request
        M->>P: Response
        P->>C: HTTP 200
    else Blocked
        L->>P: Score >= Threshold
        P->>C: HTTP 403
    end
```"""

        return Diagram(
            title="Request Flow Sequence",
            diagram_type="sequence",
            content=content,
            filename="request_flow.mmd"
        )

    def _generate_test_diagram(self) -> Diagram:
        """Generate test coverage diagram."""
        test_files = [f for f in self.repo_data.files if '_test.go' in f.name]

        if not test_files:
            return Diagram(
                title="Test Coverage",
                diagram_type="class",
                content="",
                filename="tests.mmd"
            )

        content = "```mermaid\ngraph TD\n"
        content += "    subgraph Test Files\n"

        for tf in test_files[:10]:  # Limit to 10
            name = tf.name.replace('_test.go', '')
            content += f'        T{name}["{tf.name}"]\n'

        content += "    end\n"
        content += "```"

        return Diagram(
            title="Test Coverage",
            diagram_type="graph",
            content=content,
            filename="tests.mmd"
        )

    def generate_class_diagram(self, module_name: str) -> Optional[Diagram]:
        """Generate class diagram for a module."""
        module = None
        for path, mod in self.metadata.items():
            if module_name in path:
                module = mod
                break

        if not module:
            return None

        content = "```mermaid\nclassDiagram\n"

        # Add types
        for t in module.types:
            visibility = "+" if t.is_exported else "-"
            content += f"    class {t.name} {{\n"

            # Add fields
            for field in t.fields:
                content += f"        {visibility}{field[0]}: {field[1]}\n"

            content += "    }\n\n"

            # Add methods
            for method in t.methods:
                content += f"    {t.name} : +{method}()\n"

        content += "```"

        return Diagram(
            title=f"Class Diagram: {module_name}",
            diagram_type="class",
            content=content,
            filename=f"{module_name}_class.mmd"
        )


def generate_diagrams(repo_data, metadata: Dict) -> List[Diagram]:
    """
    Generate all diagrams.

    Args:
        repo_data: Repository analysis data
        metadata: Module metadata

    Returns:
        List of Diagram objects
    """
    generator = DiagramGenerator(repo_data, metadata)
    return generator.generate_all()