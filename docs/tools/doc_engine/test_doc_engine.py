"""
Test Suite for Doc Engine
========================
Real pytest tests for the documentation engine.
"""

import pytest
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from analyzer import RepoAnalyzer
from extractor import MetadataExtractor, ModuleData, ConfigVarInfo, Language
from generator import DocGenerator, DocSection, GeneratorConfig
from diagrams import DiagramGenerator, Diagram, DiagramType, DiagramStyle
from registry import ComponentRegistry, Component, ComponentType


class TestAnalyzer:
    """Tests for RepoAnalyzer."""
    
    def test_analyzer_initialization(self):
        """Test analyzer can be initialized."""
        analyzer = RepoAnalyzer(".")
        assert analyzer.root_path.exists()
    
    def test_analyzer_with_invalid_path(self):
        """Test analyzer raises error for invalid path."""
        with pytest.raises(Exception):
            RepoAnalyzer("/nonexistent/path/that/does/not/exist")
    
    def test_analyzer_detects_languages(self):
        """Test analyzer detects project languages."""
        analyzer = RepoAnalyzer(".")
        data = analyzer.analyze()
        assert len(data.languages) > 0


class TestExtractor:
    """Tests for MetadataExtractor."""
    
    def test_extractor_initialization(self):
        """Test extractor can be initialized."""
        extractor = MetadataExtractor(".")
        assert extractor.root_path.exists()
    
    def test_config_var_info_dataclass(self):
        """Test ConfigVarInfo dataclass works."""
        cv = ConfigVarInfo(
            name="api_key",
            env_var="LAKERA_API_KEY",
            default="",
            required=True,
            description="Lakera API key"
        )
        assert cv.env_var == "LAKERA_API_KEY"
        assert cv.required is True
    
    def test_module_data_has_config_vars(self):
        """Test ModuleData has config_vars attribute."""
        module = ModuleData(path="test.go", name="test")
        assert hasattr(module, 'config_vars')
        assert isinstance(module.config_vars, list)


class TestGenerator:
    """Tests for DocGenerator."""
    
    def test_generator_config_defaults(self):
        """Test GeneratorConfig has correct defaults."""
        config = GeneratorConfig()
        assert config.include_diagrams is True
        assert config.include_security is True
        assert config.max_diagram_nodes == 50
    
    def test_doc_section_creation(self):
        """Test DocSection can be created."""
        section = DocSection(
            title="Test",
            filename="test.md",
            content="# Test",
            order=1
        )
        assert section.title == "Test"
        assert section.order == 1


class TestDiagrams:
    """Tests for DiagramGenerator."""
    
    def test_diagram_types_enum(self):
        """Test DiagramType enum values."""
        assert DiagramType.FLOWCHART.value == "flowchart"
        assert DiagramType.SEQUENCE.value == "sequence"
    
    def test_diagram_creation(self):
        """Test Diagram dataclass works."""
        diagram = Diagram(
            title="Test",
            diagram_type=DiagramType.FLOWCHART,
            style=DiagramStyle.DEFAULT,
            content="```mermaid\ngraph TD\n```",
            filename="test.mmd"
        )
        assert diagram.title == "Test"


class TestRegistry:
    """Tests for ComponentRegistry."""
    
    def test_registry_initialization(self):
        """Test registry initializes empty."""
        registry = ComponentRegistry()
        assert len(registry.components) == 0
    
    def test_register_component(self):
        """Test component registration."""
        registry = ComponentRegistry()
        component = Component(
            name="test",
            component_type=ComponentType.SOURCE,
            path="/test/path"
        )
        registry.register_component(component)
        assert "test" in registry.components
    
    def test_get_component(self):
        """Test retrieving registered component."""
        registry = ComponentRegistry()
        component = Component(
            name="test",
            component_type=ComponentType.SERVICE,
            path="/test/path"
        )
        registry.register_component(component)
        retrieved = registry.get_component("test")
        assert retrieved is not None
        assert retrieved.name == "test"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])