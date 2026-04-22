"""
Go Documentation Generator
=========================

Specialized extractor for Go source code.
Extracts detailed information including:
- Functions with docstrings
- Methods on types
- Struct fields
- Interface definitions
- Error handling patterns
- Test functions
- Example functions

Author: HexStrike Documentation Team
"""

import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field


@dataclass
class GoFunction:
    """Go function or method."""
    name: str
    receiver: str = ""  # For methods
    params: List[Tuple[str, str]] = field(default_factory=list)  # (name, type)
    results: List[Tuple[str, str]] = field(default_factory=list)  # (name, type)
    is_exported: bool = False
    is_method: bool = False
    is_test: bool = False
    is_example: bool = False
    line_number: int = 0
    docstring: str = ""
    comments: List[str] = field(default_factory=list)


@dataclass
class GoType:
    """Go type definition."""
    name: str
    kind: str = "struct"  # struct, interface, enum
    fields: List[Tuple[str, str, str]] = field(default_factory=list)  # (name, type, tag)
    methods: List[str] = field(default_factory=list)
    is_exported: bool = False
    line_number: int = 0
    docstring: str = ""
    implements: List[str] = field(default_factory=list)


@dataclass
class GoConst:
    """Go constant or variable."""
    name: str
    value: str = ""
    type_hint: str = ""
    is_exported: bool = False
    docstring: str = ""


@dataclass
class GoImport:
    """Go import."""
    path: str
    alias: str = ""
    is_stdlib: bool = False
    is_external: bool = False


class GoExtractor:
    """Extracts documentation from Go source code."""

    # Standard library packages
    STDLIB = {
        'fmt', 'os', 'io', 'net', 'http', 'json', 'time', 'context',
        'strings', 'bytes', 'errors', 'log', 'sync', 'math', 'rand',
        'crypto', 'hash', 'encoding', 'regexp', 'path', 'flag', 'bufio',
        'encoding/json', 'io/ioutil', 'os/signal', 'syscall', 'sort',
        'unicode', 'strconv', 'reflect', 'unsafe', 'runtime', 'testing',
    }

    def __init__(self):
        self.functions: List[GoFunction] = []
        self.types: List[GoType] = []
        self.consts: List[GoConst] = []
        self.imports: List[GoImport] = []
        self.packages: List[str] = []

    def extract(self, file_path: str) -> Dict:
        """Extract all documentation from a Go file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except (OSError, UnicodeDecodeError):
            return {}

        # Extract package
        pkg_match = re.search(r'^package\s+(\w+)', content, re.MULTILINE)
        package = pkg_match.group(1) if pkg_match else ""

        # Extract all components
        self._extract_imports(content)
        self._extract_consts(content)
        self._extract_types(content)
        self._extract_functions(content, package)

        return {
            'package': package,
            'imports': self.imports,
            'consts': self.consts,
            'types': self.types,
            'functions': self.functions,
        }

    def _extract_imports(self, content: str):
        """Extract import statements."""
        self.imports = []

        # Simple imports: import "path"
        for match in re.finditer(r'import\s+"([^"]+)"', content):
            path = match.group(1)
            self.imports.append(GoImport(
                path=path,
                is_stdlib=self._is_stdlib(path),
                is_external=not self._is_stdlib(path)
            ))

        # Grouped imports
        import_block = re.search(r'import\s*\((.*?)\)', content, re.DOTALL)
        if import_block:
            for line in import_block.group(1).split('\n'):
                line = line.strip().strip('"')
                if not line or line.startswith('//'):
                    continue

                # Aliased imports: name "path"
                alias_match = re.match(r'(\w+)\s+"([^"]+)"', line)
                if alias_match:
                    self.imports.append(GoImport(
                        path=alias_match.group(2),
                        alias=alias_match.group(1),
                        is_stdlib=self._is_stdlib(alias_match.group(2)),
                        is_external=not self._is_stdlib(alias_match.group(2))
                    ))
                else:
                    self.imports.append(GoImport(
                        path=line,
                        is_stdlib=self._is_stdlib(line),
                        is_external=not self._is_stdlib(line)
                    ))

    def _is_stdlib(self, path: str) -> bool:
        """Check if import is stdlib."""
        base = path.split('/')[0]
        return base in self.STDLIB

    def _extract_consts(self, content: str):
        """Extract constant declarations."""
        self.consts = []

        # Simple const: const Name = value
        for match in re.finditer(r'const\s+(\w+)\s*=\s*"([^"]+)"', content):
            name = match.group(1)
            self.consts.append(GoConst(
                name=name,
                value=f'"{match.group(2)}"',
                is_exported=name[0].isupper()
            ))

        # iota constants
        for match in re.finditer(r'(\w+)\s*=\s*iota', content):
            name = match.group(1)
            self.consts.append(GoConst(
                name=name,
                value='iota',
                is_exported=name[0].isupper()
            ))

    def _extract_types(self, content: str):
        """Extract type definitions."""
        self.types = []

        # Struct definitions
        struct_pattern = r'type\s+(\w+)\s+struct\s*\{((?:[^\{\}]|\{[^\}]*\})*)\}'
        for match in re.finditer(struct_pattern, content, re.DOTALL):
            name = match.group(1)
            body = match.group(2)

            fields = []
            for field_match in re.finditer(r'(\w+)\s+(\*?\w+(?:\[\])?)\s*(?:`([^`]+)`)?', body):
                fields.append((
                    field_match.group(1),
                    field_match.group(2),
                    field_match.group(3) or ''
                ))

            self.types.append(GoType(
                name=name,
                kind='struct',
                fields=fields,
                is_exported=name[0].isupper()
            ))

        # Interface definitions
        interface_pattern = r'type\s+(\w+)\s+interface\s*\{(.*?)\}'
        for match in re.finditer(interface_pattern, content, re.DOTALL):
            name = match.group(1)
            body = match.group(2)

            methods = []
            for method_match in re.finditer(r'(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?', body):
                params = method_match.group(2) or ""
                results = method_match.group(3) or ""
                methods.append(f"{method_match.group(1)}({params})")

            self.types.append(GoType(
                name=name,
                kind='interface',
                methods=methods,
                is_exported=name[0].isupper()
            ))

    def _extract_functions(self, content: str, package: str):
        """Extract function and method definitions."""
        self.functions = []

        # Pattern for standalone functions: func Name(params) (results) {
        standalone_pattern = r'func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?\s*\{'

        # Pattern for methods: func (receiver) Name(params) (results) {
        method_pattern = r'func\s*\(([^)]+)\)\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?\s*\{'

        # Track functions to avoid duplicates
        seen_functions = set()

        # Extract methods first (more specific pattern)
        for match in re.finditer(method_pattern, content):
            receiver = match.group(1)
            name = match.group(2)
            params_str = match.group(3) or ""
            results_str = match.group(4) or ""

            params = self._parse_params(params_str)
            results = self._parse_params(results_str)
            line_number = content[:match.start()].count('\n') + 1

            # Skip if already seen
            if name in seen_functions and not name.startswith('Test'):
                continue
            seen_functions.add(name)

            # Determine type
            is_test = name.startswith('Test') and params and 'testing.T' in str(params)
            is_example = name.startswith('Example')
            is_exported = name[0].isupper() if name else False

            self.functions.append(GoFunction(
                name=name,
                receiver=receiver,
                params=params,
                results=results,
                is_exported=is_exported,
                is_method=True,
                is_test=is_test,
                is_example=is_example,
                line_number=line_number
            ))

    def _parse_params(self, params_str: str) -> List[Tuple[str, str]]:
        """Parse Go function parameters."""
        params = []
        if not params_str.strip():
            return params

        for param in params_str.split(','):
            param = param.strip()
            if not param:
                continue

            # Handle named params: name type
            parts = param.split()
            if len(parts) >= 2:
                name = parts[0]
                type_str = ' '.join(parts[1:])
            else:
                name = ""
                type_str = parts[0] if parts else ""

            params.append((name, type_str))

        return params


def extract_go_documentation(file_path: str) -> Dict:
    """
    Extract comprehensive Go documentation from a file.

    Args:
        file_path: Path to Go source file

    Returns:
        Dictionary with package, imports, consts, types, functions
    """
    extractor = GoExtractor()
    return extractor.extract(file_path)