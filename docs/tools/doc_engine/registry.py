"""
Component Registry
=================

Maintains a registry of components, relationships, and dependencies.
Provides a central catalog for tracking project components.

Author: HexStrike Documentation Team
"""

from typing import Dict, List, Optional, Set
from dataclasses import dataclass, field
from enum import Enum


class ComponentType(Enum):
    """Types of components in the system."""
    SOURCE = "source"
    CONFIG = "config"
    SERVICE = "service"
    LIBRARY = "library"
    SCRIPT = "script"
    MANIFEST = "manifest"
    TEST = "test"


@dataclass
class Component:
    """A registered component."""
    name: str
    component_type: ComponentType
    path: str
    description: str = ""
    dependencies: List[str] = field(default_factory=list)
    dependents: List[str] = field(default_factory=list)
    language: str = ""
    language_version: str = ""
    interfaces: List[str] = field(default_factory=list)


@dataclass
class Relationship:
    """A relationship between components."""
    source: str
    target: str
    relationship_type: str  # uses, depends_on, calls, etc.
    description: str = ""


class ComponentRegistry:
    """Registry for project components and their relationships."""

    def __init__(self):
        self.components: Dict[str, Component] = {}
        self.relationships: List[Relationship] = {}

    def register_component(self, component: Component):
        """Register a component."""
        self.components[component.name] = component

    def register_relationship(self, relationship: Relationship):
        """Register a relationship between components."""
        key = f"{relationship.source}:{relationship.target}"
        self.relationships[key] = relationship

    def get_component(self, name: str) -> Optional[Component]:
        """Get a component by name."""
        return self.components.get(name)

    def get_components_by_type(self, component_type: ComponentType) -> List[Component]:
        """Get all components of a specific type."""
        return [c for c in self.components.values() if c.component_type == component_type]

    def get_dependencies(self, name: str) -> List[Component]:
        """Get dependencies of a component."""
        component = self.get_component(name)
        if not component:
            return []

        return [
            self.get_component(dep)
            for dep in component.dependencies
            if self.get_component(dep)
        ]

    def to_dict(self) -> Dict:
        """Convert registry to dictionary for serialization."""
        return {
            "components": {
                name: {
                    "type": c.component_type.value,
                    "path": c.path,
                    "description": c.description,
                    "dependencies": c.dependencies,
                    "language": c.language,
                }
                for name, c in self.components.items()
            },
            "relationships": [
                {
                    "source": r.source,
                    "target": r.target,
                    "type": r.relationship_type,
                }
                for r in self.relationships.values()
            ]
        }


def create_default_registry() -> ComponentRegistry:
    """Create a default registry with known components."""
    registry = ComponentRegistry()

    # Register core components
    components = [
        Component(
            name="mcp-policy-proxy",
            component_type=ComponentType.SERVICE,
            path="src/mcp-policy-proxy/",
            description="Semantic firewall proxy for MCP tool calls",
            language="Go",
            language_version="1.21+",
            dependencies=["golang-jwt", "prometheus-client", "google-uuid"],
        ),
        Component(
            name="proxy.go",
            component_type=ComponentType.SOURCE,
            path="src/mcp-policy-proxy/proxy.go",
            description="Main proxy implementation with middleware chain",
            language="Go",
            interfaces=["HTTP Handler", "Middleware Chain"],
        ),
        Component(
            name="lakera.go",
            component_type=ComponentType.SOURCE,
            path="src/mcp-policy-proxy/lakera.go",
            description="Lakera Guard API client",
            language="Go",
            interfaces=["LakeraChecker"],
        ),
        Component(
            name="dlq",
            component_type=ComponentType.LIBRARY,
            path="src/mcp-policy-proxy/dlq/",
            description="Dead Letter Queue for failed requests",
            language="Go",
        ),
    ]

    for comp in components:
        registry.register_component(comp)

    return registry