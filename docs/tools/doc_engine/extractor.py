"""
Metadata Extractor
=================

Extracts detailed metadata from code files:
- Functions and methods
- Types and structs
- Interfaces
- Constants
- Configuration variables
- API endpoints
- Dependencies
- Imports

Author: HexStrike Documentation Team
"""

import re
import os
from pathlib import Path
from typing import Dict, List, Optional, Set
from dataclasses import dataclass, field


@dataclass
class FunctionInfo:
    """Information about a function or method."""
    name: str
    params: List[str] = field(default_factory=list)
    return_type: str = ""
    is_exported: bool = False
    is_handler: bool = False
    line_number: int = 0
    docstring: str = ""


@dataclass
class TypeInfo:
    """Information about a type or struct."""
    name: str
    kind: str = ""  # struct, interface, enum
    fields: List[Dict[str, str]] = field(default_factory=list)
    methods: List[str] = field(default_factory=list)
    is_exported: bool = False
    line_number: int = 0
    docstring: str = ""


@dataclass
class ConstantInfo:
    """Information about a constant."""
    name: str
    value: str = ""
    type_hint: str = ""
    is_exported: bool = False
    docstring: str = ""


@dataclass
class ConfigVar:
    """Information about a configuration variable."""
    name: str
    env_var: str = ""
    default: str = ""
    required: bool = False
    description: str = ""


@dataclass
class ImportInfo:
    """Information about an import."""
    path: str
    alias: str = ""
    is_standard: bool = False
    is_external: bool = False


@dataclass
class EndpointInfo:
    """Information about an HTTP endpoint."""
    path: str
    method: str = "GET"
    handler: str = ""
    middleware: List[str] = field(default_factory=list)
    auth_required: bool = False
    description: str = ""


@dataclass
class ModuleData:
    """Complete metadata for a module."""
    path: str
    name: str
    package: str = ""
    imports: List[ImportInfo] = field(default_factory=list)
    functions: List[FunctionInfo] = field(default_factory=list)
    types: List[TypeInfo] = field(default_factory=list)
    constants: List[ConstantInfo] = field(default_factory=list)
    endpoints: List[EndpointInfo] = field(default_factory=list)
    config_vars: List[ConfigVar] = field(default_factory=list)


class MetadataExtractor:
    """Extracts metadata from source code files."""

    # Standard library packages
    STANDARD_PACKAGES = {
        'fmt', 'os', 'io', 'net', 'http', 'json', 'time', 'context',
        'strings', 'bytes', 'errors', 'log', 'sync', 'math', 'rand',
        'crypto', 'hash', 'encoding', 'regexp', 'path', 'flag', 'os/exec',
    }

    def __init__(self, root_path: str):
        self.root_path = Path(root_path)

    def extract_all(self, files: List) -> Dict[str, ModuleData]:
        """
        Extract metadata from all source files.

        Args:
            files: List of FileInfo objects from analyzer

        Returns:
            Dictionary mapping file paths to ModuleData
        """
        modules = {}

        for file_info in files:
            if file_info.extension == '.go':
                module_data = self._extract_go_file(file_info.path)
                if module_data:
                    modules[file_info.path] = module_data

            elif file_info.extension == '.py':
                module_data = self._extract_python_file(file_info.path)
                if module_data:
                    modules[file_info.path] = module_data

        return modules

    def _extract_go_file(self, file_path: str) -> Optional[ModuleData]:
        """Extract metadata from Go source file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except (OSError, UnicodeDecodeError):
            return None

        module_data = ModuleData(
            path=file_path,
            name=Path(file_path).stem
        )

        # Extract package
        package_match = re.search(r'^package\s+(\w+)', content, re.MULTILINE)
        if package_match:
            module_data.package = package_match.group(1)

        # Extract imports
        module_data.imports = self._extract_go_imports(content)

        # Extract constants
        module_data.constants = self._extract_go_constants(content)

        # Extract functions
        module_data.functions = self._extract_go_functions(content)

        # Extract types
        module_data.types = self._extract_go_types(content)

        # Extract endpoints
        module_data.endpoints = self._extract_go_endpoints(content)

        # Extract config variables
        module_data.config_vars = self._extract_go_config(content)

        return module_data

    def _extract_go_imports(self, content: str) -> List[ImportInfo]:
        """Extract import statements from Go code."""
        imports = []

        # Find import block
        import_match = re.search(
            r'import\s*\((.*?)\)|import\s+"([^"]+)"',
            content,
            re.DOTALL
        )

        if not import_match:
            return imports

        if import_match.group(1):
            # Block import
            block = import_match.group(1)
            for line in block.split('\n'):
                line = line.strip().strip('"')
                if not line or line.startswith('//'):
                    continue

                # Handle aliased imports
                alias_match = re.match(r'(\w+)\s+"([^"]+)"', line)
                if alias_match:
                    imp = ImportInfo(
                        path=alias_match.group(2),
                        alias=alias_match.group(1)
                    )
                else:
                    imp = ImportInfo(path=line)

                # Classify import
                imp.is_standard = any(imp.path.startswith(p) for p in self.STANDARD_PACKAGES)
                imp.is_external = not imp.is_standard and not imp.path.startswith('.')
                imports.append(imp)

        return imports

    def _extract_go_constants(self, content: str) -> List[ConstantInfo]:
        """Extract constant declarations from Go code."""
        constants = []

        # Find const block
        const_match = re.search(
            r'const\s*\((.*?)\)|const\s+(\w+)\s*=',
            content,
            re.DOTALL
        )

        if not const_match:
            return constants

        if const_match.group(1):
            # Block const
            block = const_match.group(1)
            for line in block.split('\n'):
                line = line.strip()
                if not line or line.startswith('//'):
                    continue

                # Parse: name = value or name type = value
                const_match = re.match(r'(\w+)(?:\s+\w+)?\s*=\s*(.+)', line)
                if const_match:
                    name = const_match.group(1)
                    value = const_match.group(2).rstrip(',')
                    is_exported = name[0].isupper() if name else False

                    constants.append(ConstantInfo(
                        name=name,
                        value=value,
                        is_exported=is_exported
                    ))

        return constants

    def _extract_go_functions(self, content: str) -> List[FunctionInfo]:
        """Extract function declarations from Go code.
        
        Handles both:
        - Standalone functions: func name(params) return_type { ... }
        - Methods with receiver: func (r *Type) name(params) return_type { ... }
        """
        functions = []
        seen_positions = set()  # Track positions to avoid duplicates
        
        # Pattern 1: Methods with receiver: func (r *Type) name(params) return_type { ... }
        method_pattern = r'func\s+\(([^)]+)\)\s+(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?\s*\{'
        
        for match in re.finditer(method_pattern, content):
            pos = match.start()
            if pos in seen_positions:
                continue
            seen_positions.add(pos)
            
            receiver = match.group(1)
            name = match.group(2)
            params = match.group(3) if match.group(3) else ""
            return_type = match.group(4) if match.group(4) else ""
            line_number = content[:pos].count('\n') + 1

            is_exported = name[0].isupper() if name else False
            is_handler = 'Handler' in name or name.endswith('Handler') or name.endswith('Func')

            param_list = [p.strip() for p in params.split(',') if p.strip()]
            full_params = [receiver] + param_list

            functions.append(FunctionInfo(
                name=name,
                params=full_params,
                return_type=return_type,
                is_exported=is_exported,
                is_handler=is_handler,
                line_number=line_number
            ))

        # Pattern 2: Standalone functions: func name(params) return_type { ... }
        func_pattern = r'func\s+(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?\s*\{'

        for match in re.finditer(func_pattern, content):
            pos = match.start()
            if pos in seen_positions:
                continue
            seen_positions.add(pos)

            name = match.group(1)
            params = match.group(2) if match.group(2) else ""
            return_type = match.group(3) if match.group(3) else ""
            line_number = content[:pos].count('\n') + 1

            is_exported = name[0].isupper() if name else False
            is_handler = 'Handler' in name or name.endswith('Handler') or name.endswith('Func')

            param_list = [p.strip() for p in params.split(',') if p.strip()]

            functions.append(FunctionInfo(
                name=name,
                params=param_list,
                return_type=return_type,
                is_exported=is_exported,
                is_handler=is_handler,
                line_number=line_number
            ))

        return functions

    def _extract_go_types(self, content: str) -> List[TypeInfo]:
        """Extract type and struct declarations from Go code."""
        types = []

        # Match: type Name struct { ... }
        struct_pattern = r'type\s+(\w+)\s+struct\s*\{(.*?)\}'

        for match in re.finditer(struct_pattern, content, re.DOTALL):
            name = match.group(1)
            body = match.group(2)
            line_number = content[:match.start()].count('\n') + 1

            is_exported = name[0].isupper() if name else False

            # Extract fields
            fields = []
            for field_line in body.split('\n'):
                field_line = field_line.strip()
                if not field_line or field_line.startswith('//'):
                    continue

                # Parse: Name Type `json:"tag"`
                field_match = re.match(r'(\w+)\s+(\w+(?:\[\])?)\s*(?:`([^`]+)`)?', field_line)
                if field_match:
                    fields.append({
                        'name': field_match.group(1),
                        'type': field_match.group(2),
                        'tag': field_match.group(3) or ''
                    })

            types.append(TypeInfo(
                name=name,
                kind='struct',
                fields=fields,
                is_exported=is_exported,
                line_number=line_number
            ))

        # Match: type Name interface { ... }
        interface_pattern = r'type\s+(\w+)\s+interface\s*\{(.*?)\}'

        for match in re.finditer(interface_pattern, content, re.DOTALL):
            name = match.group(1)
            body = match.group(2)
            line_number = content[:match.start()].count('\n') + 1

            is_exported = name[0].isupper() if name else False

            # Extract methods
            methods = []
            for method_line in body.split('\n'):
                method_line = method_line.strip()
                if not method_line or method_line.startswith('//'):
                    continue
                # Match: MethodName(params) return_type
                method_match = re.match(r'(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?', method_line)
                if method_match:
                    methods.append(method_match.group(1))

            types.append(TypeInfo(
                name=name,
                kind='interface',
                methods=methods,
                is_exported=is_exported,
                line_number=line_number
            ))

        return types

    def _extract_go_endpoints(self, content: str) -> List[EndpointInfo]:
        """Extract HTTP endpoint handlers from Go code."""
        endpoints = []

        # Match: mux.Handle or router.Handle or http.Handle
        handle_patterns = [
            r'mux\.Handle\("([^"]+)"\s*,\s*(\w+)',
            r'router\.Handle\("([^"]+)"\s*,\s*(\w+)',
            r'\.HandleFunc\("([^"]+)"\s*,\s*(\w+)',
        ]

        for pattern in handle_patterns:
            for match in re.finditer(pattern, content):
                path = match.group(1)
                handler = match.group(2)

                # Determine HTTP method
                method = "GET"
                if "POST" in content[match.start():match.start()+200]:
                    method = "POST"

                endpoints.append(EndpointInfo(
                    path=path,
                    method=method,
                    handler=handler
                ))

        return endpoints

    def _extract_go_config(self, content: str) -> List[ConfigVar]:
        """Extract configuration variables from Go code."""
        config_vars = []

        # Match: VarName = getEnv("VAR_NAME", default)
        pattern = r'(\w+)\s*:?\s*=\s*getEnv\(["\'](\w+)["\']\s*,\s*(.+?)\)'

        for match in re.finditer(pattern, content):
            name = match.group(1)
            env_var = match.group(2)
            default = match.group(3).strip()

            # Required if: empty string default, or comment says required/mandatory
            is_required = (
                default == '""' or 
                default == "''" or
                'REQUIRED' in name.upper() or
                'MANDATORY' in name.upper()
            )

            config_vars.append(ConfigVar(
                name=name,
                env_var=env_var,
                default=default.strip('"').strip("'"),
                required=is_required
            ))

        return config_vars

    def _extract_python_file(self, file_path: str) -> Optional[ModuleData]:
        """Extract metadata from Python source file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except (OSError, UnicodeDecodeError):
            return None

        module_data = ModuleData(
            path=file_path,
            name=Path(file_path).stem
        )

        # Extract package (directory name)
        module_data.package = Path(file_path).parent.name

        # Extract imports
        for match in re.finditer(r'^\s*(?:from\s+(\S+)\s+)?import\s+([^\n#]+)', content, re.MULTILINE):
            module_data.imports.append(ImportInfo(
                path=match.group(1) or match.group(2),
                is_standard=not match.group(1) or '.' not in match.group(1)
            ))

        # Extract functions
        for match in re.finditer(r'^def\s+(\w+)\s*\(([^)]*)\):', content, re.MULTILINE):
            params = [p.strip() for p in match.group(2).split(',') if p.strip()]
            name = match.group(1)
            is_exported = not name.startswith('_')

            module_data.functions.append(FunctionInfo(
                name=name,
                params=params,
                is_exported=is_exported,
                line_number=content[:match.start()].count('\n') + 1
            ))

        # Extract classes
        for match in re.finditer(r'^class\s+(\w+)(?:\(([^)]+)\))?:', content, re.MULTILINE):
            name = match.group(1)
            is_exported = not name.startswith('_')

            module_data.types.append(TypeInfo(
                name=name,
                kind='class',
                is_exported=is_exported,
                line_number=content[:match.start()].count('\n') + 1
            ))

        return module_data

    def extract_dependencies(self, module_data: Dict[str, ModuleData]) -> Dict[str, Set[str]]:
        """Extract dependency graph from modules."""
        deps = {}

        for path, data in module_data.items():
            modules = set()

            for imp in data.imports:
                if imp.is_external:
                    # Extract module name from import path
                    module = imp.path.split('/')[-1]
                    modules.add(module)

            deps[path] = modules

        return deps


def extract_metadata(root_path: str, files: List) -> Dict[str, ModuleData]:
    """
    Convenience function to extract metadata from files.

    Args:
        root_path: Path to repository root
        files: List of FileInfo objects

    Returns:
        Dictionary mapping file paths to ModuleData
    """
    extractor = MetadataExtractor(root_path)
    return extractor.extract_all(files)