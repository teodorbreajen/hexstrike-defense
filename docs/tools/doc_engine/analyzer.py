"""
Repository Analyzer
================

Scans and analyzes the repository structure to detect:
- File tree
- Programming languages
- Key files (entrypoints, configs, tests)
- Modules and packages
- Configuration files
- Scripts

Author: HexStrike Documentation Team
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum


class Language(Enum):
    """Detected programming languages."""
    GO = "Go"
    PYTHON = "Python"
    YAML = "YAML"
    MARKDOWN = "Markdown"
    BASH = "Bash"
    DOCKERFILE = "Dockerfile"
    JSON = "JSON"
    UNKNOWN = "Unknown"


class FileType(Enum):
    """Classified file types."""
    SOURCE = "source"
    TEST = "test"
    CONFIG = "config"
    MANIFEST = "manifest"
    SCRIPT = "script"
    DOCUMENTATION = "documentation"
    TEMPLATE = "template"
    WORKFLOW = "workflow"
    DEPENDENCY = "dependency"
    DOCKERFILE = "dockerfile"
    MAKEFILE = "makefile"
    UNKNOWN = "unknown"


@dataclass
class FileInfo:
    """Information about a single file."""
    path: str
    relative_path: str
    name: str
    extension: str
    file_type: FileType = FileType.UNKNOWN
    language: Language = Language.UNKNOWN
    size_bytes: int = 0
    lines_of_code: int = 0
    has_tests: bool = False
    imports: List[str] = field(default_factory=list)
    exports: List[str] = field(default_factory=list)


@dataclass
class ModuleInfo:
    """Information about a code module/package."""
    name: str
    path: str
    files: List[FileInfo] = field(default_factory=list)
    functions: List[str] = field(default_factory=list)
    types: List[str] = field(default_factory=list)
    interfaces: List[str] = field(default_factory=list)
    constants: List[str] = field(default_factory=list)


@dataclass
class RepoData:
    """Complete repository analysis data."""
    root_path: str
    project_name: str
    languages: List[Language] = field(default_factory=list)
    files: List[FileInfo] = field(default_factory=list)
    modules: List[ModuleInfo] = field(default_factory=list)
    config_files: List[FileInfo] = field(default_factory=list)
    manifests: List[FileInfo] = field(default_factory=list)
    scripts: List[FileInfo] = field(default_factory=list)
    docs: List[FileInfo] = field(default_factory=list)
    tests: List[FileInfo] = field(default_factory=list)
    workflows: List[FileInfo] = field(default_factory=list)
    entrypoints: List[FileInfo] = field(default_factory=list)
    makefile_targets: Dict[str, str] = field(default_factory=dict)
    dockerfiles: List[FileInfo] = field(default_factory=list)


class RepoAnalyzer:
    """Analyzes repository structure and content."""

    # File extension to language mapping
    EXTENSION_MAP = {
        ".go": Language.GO,
        ".py": Language.PYTHON,
        ".yaml": Language.YAML,
        ".yml": Language.YAML,
        ".md": Language.MARKDOWN,
        ".sh": Language.BASH,
        ".bash": Language.BASH,
        ".dockerfile": Language.DOCKERFILE,
        ".json": Language.JSON,
    }

    # File type classification rules
    TYPE_PATTERNS = {
        FileType.TEST: [r"_test\.go$", r"_test\.py$", r"test_.*\.sh$", r"\.test\."],
        FileType.CONFIG: [r"^config", r"\.config\.", r"settings\."],
        FileType.MANIFEST: [r"^manifests/", r"\.yaml$", r"\.yml$"],
        FileType.SCRIPT: [r"^scripts/", r"\.sh$"],
        FileType.DOCUMENTATION: [r"^docs/", r"\.md$", r"README", r"CHANGELOG"],
        FileType.WORKFLOW: [r"^ci/", r"^ workflows/", r"\.yml$"],
        FileType.DEPENDENCY: [r"^go\.mod$", r"^package\.json$", r"^requirements\.txt$", r"^Pipfile$"],
        FileType.MAKEFILE: [r"^Makefile$", r"^makefile$"],
    }

    def __init__(self, root_path: str):
        self.root_path = Path(root_path)
        if not self.root_path.exists():
            raise FileNotFoundError(f"Repository path does not exist: {self.root_path}")
        if not self.root_path.is_dir():
            raise NotADirectoryError(f"Repository path is not a directory: {self.root_path}")
        self.project_name = self.root_path.name

    def analyze(self) -> RepoData:
        """
        Perform complete repository analysis.

        Returns:
            RepoData with all analysis results
        """
        data = RepoData(
            root_path=str(self.root_path),
            project_name=self.project_name
        )

        # Walk directory tree
        for root, dirs, files in os.walk(self.root_path):
            root_path = Path(root)

            # Skip hidden and ignored directories
            dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ('vendor', 'node_modules', '__pycache__', 'docs', 'generated', 'tools')]

            for filename in files:
                file_path = root_path / filename
                relative = file_path.relative_to(self.root_path)

                # Skip hidden files
                if filename.startswith('.'):
                    continue

                file_info = self._analyze_file(file_path, relative)
                if file_info:
                    data.files.append(file_info)

                    # Classify by type
                    self._classify_file(file_info, data)

        # Detect languages used
        data.languages = self._detect_languages(data.files)

        # Parse Makefile
        data.makefile_targets = self._parse_makefile()

        return data

    def _analyze_file(self, file_path: Path, relative: Path) -> Optional[FileInfo]:
        """Analyze a single file."""
        try:
            stat = file_path.stat()
        except (OSError, PermissionError):
            return None

        extension = file_path.suffix.lower()
        language = self.EXTENSION_MAP.get(extension, Language.UNKNOWN)

        # Detect language from content if unknown extension
        if language == Language.UNKNOWN:
            language = self._detect_language_from_content(file_path)

        # Count lines of code
        loc = self._count_lines(file_path)

        return FileInfo(
            path=str(file_path),
            relative_path=str(relative),
            name=file_path.name,
            extension=extension,
            file_type=FileType.UNKNOWN,
            language=language,
            size_bytes=stat.st_size,
            lines_of_code=loc
        )

    def _detect_language_from_content(self, file_path: Path) -> Language:
        """Detect language from file content patterns."""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read(1024)

            # Check for Go patterns
            if 'package main' in content or 'func ' in content and 'import (' in content:
                return Language.GO

            # Check for Python
            if 'import ' in content and 'def ' in content:
                return Language.PYTHON

            # Check for YAML
            if content.startswith('---') or re.match(r'^[a-z]+:', content):
                return Language.YAML

            # Check for Dockerfile
            if content.startswith('FROM '):
                return Language.DOCKERFILE

            # Check for Bash
            if content.startswith('#!/bin/bash') or content.startswith('#!/bin/sh'):
                return Language.BASH

        except (OSError, UnicodeDecodeError):
            pass

        return Language.UNKNOWN

    def _count_lines(self, file_path: Path) -> int:
        """Count non-empty lines in file."""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                return sum(1 for line in f if line.strip())
        except (OSError, UnicodeDecodeError):
            return 0

    def _classify_file(self, file_info: FileInfo, data: RepoData):
        """Classify file into appropriate category."""
        path = file_info.relative_path
        name = file_info.name.lower()

        # Check by patterns
        for ftype, patterns in self.TYPE_PATTERNS.items():
            for pattern in patterns:
                if re.search(pattern, path, re.IGNORECASE):
                    file_info.file_type = ftype
                    break

        # Fallback: classify by extension
        if file_info.file_type == FileType.UNKNOWN:
            if file_info.extension in ('.go', '.py'):
                file_info.file_type = FileType.SOURCE
            elif file_info.extension in ('.yaml', '.yml'):
                file_info.file_type = FileType.MANIFEST

        # Add to appropriate list
        if file_info.file_type == FileType.TEST:
            data.tests.append(file_info)
        elif file_info.file_type == FileType.CONFIG:
            data.config_files.append(file_info)
        elif file_info.file_type == FileType.MANIFEST:
            data.manifests.append(file_info)
        elif file_info.file_type == FileType.SCRIPT:
            data.scripts.append(file_info)
        elif file_info.file_type == FileType.DOCUMENTATION:
            data.docs.append(file_info)
        elif file_info.file_type == FileType.WORKFLOW:
            data.workflows.append(file_info)
        elif file_info.file_type == FileType.MAKEFILE:
            data.config_files.append(file_info)
        elif file_info.extension == '.dockerfile':
            data.dockerfiles.append(file_info)

        # Mark entrypoints
        if file_info.name in ('main.go', 'main.py', 'app.py', 'server.py'):
            data.entrypoints.append(file_info)

    def _detect_languages(self, files: List[FileInfo]) -> List[Language]:
        """Detect programming languages used in project."""
        language_counts = {}

        for file in files:
            lang = file.language
            if lang != Language.UNKNOWN:
                language_counts[lang] = language_counts.get(lang, 0) + 1

        # Sort by usage count
        sorted_langs = sorted(language_counts.items(), key=lambda x: x[1], reverse=True)
        return [lang for lang, _ in sorted_langs]

    def _parse_makefile(self) -> Dict[str, str]:
        """Parse Makefile targets."""
        targets = {}

        makefile_path = self.root_path / 'Makefile'
        if not makefile_path.exists():
            return targets

        try:
            with open(makefile_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # Find targets (lines starting with alphanumeric followed by :)
            for match in re.finditer(r'^([a-zA-Z_][a-zA-Z0-9_-]*):\s*(.*?)$', content, re.MULTILINE):
                target = match.group(1)
                desc = match.group(2).strip()
                if desc and not desc.startswith('#'):
                    targets[target] = desc
                elif not desc:
                    targets[target] = ""

        except (OSError, UnicodeDecodeError):
            pass

        return targets


def analyze_repo(root_path: str) -> RepoData:
    """
    Convenience function to analyze a repository.

    Args:
        root_path: Path to repository root

    Returns:
        RepoData with analysis results
    """
    analyzer = RepoAnalyzer(root_path)
    return analyzer.analyze()