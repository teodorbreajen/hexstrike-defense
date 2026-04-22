"""
HexStrike Defense - Documentation Engine
========================================

Automated documentation generator for hexstrike-defense project.

This module provides tools for:
- Repository structure analysis
- Code metadata extraction
- Document generation from templates
- Component registry management

Usage:
    from doc_engine import analyze_repo, generate_docs

    # Analyze repository
    repo_data = analyze_repo("/path/to/repo")

    # Generate documentation
    generate_docs(repo_data, output_dir="/path/to/docs")
"""

__version__ = "1.0.0"
__author__ = "HexStrike Documentation Team"

from .analyzer import RepoAnalyzer
from .extractor import MetadataExtractor
from .generator import DocGenerator
from .registry import ComponentRegistry

__all__ = [
    "RepoAnalyzer",
    "MetadataExtractor",
    "DocGenerator",
    "ComponentRegistry",
]
