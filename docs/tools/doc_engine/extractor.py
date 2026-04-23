"""
Metadata Extractor - Enhanced Edition
=====================================

Advanced metadata extraction with:
- Multi-language support (Go, Python, TypeScript, JavaScript, Rust)
- Robust regex patterns with named groups
- Docstring extraction
- API endpoint detection
- Security pattern detection
- Dependency analysis
- Complexity metrics
- Concurrent processing

Author: HexStrike Documentation Team
Version: 2.1.0
"""

import re
import os
import ast
import hashlib
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple, Any
from dataclasses import dataclass, field
from concurrent.futures import ThreadPoolExecutor, as_completed
from enum import Enum
import logging

logger = logging.getLogger(__name__)


class ExtractorError(Exception):
    """Raised when extraction fails."""
    pass


class Language(Enum):
    """Supported languages for extraction."""
    GO = "Go"
    PYTHON = "Python"
    TYPESCRIPT = "TypeScript"
    JAVASCRIPT = "JavaScript"
    RUST = "Rust"
    UNKNOWN = "Unknown"


@dataclass
class FunctionInfo:
    """Information about a function or method."""
    name: str
    params: List[Tuple[str, str]] = field(default_factory=list)  # (name, type)
    return_type: str = ""
    is_exported: bool = False
    is_handler: bool = False
    is_test: bool = False
    is_async: bool = False
    line_number: int = 0
    end_line: int = 0
    docstring: str = ""
    complexity: int = 0
    coverage_percent: float = 0.0
    calls: List[str] = field(default_factory=list)


@dataclass
class TypeInfo:
    """Information about a type or struct."""
    name: str
    kind: str = ""  # struct, interface, enum, class, trait
    fields: List[Tuple[str, str, str]] = field(default_factory=list)  # (name, type, tag)
    methods: List[str] = field(default_factory=list)
    implements: List[str] = field(default_factory=list)
    is_exported: bool = False
    is_error: bool = False  # Implements error interface
    line_number: int = 0
    end_line: int = 0
    docstring: str = ""
    complexity: int = 0


@dataclass
class ConstantInfo:
    """Information about a constant."""
    name: str
    value: str = ""
    type_hint: str = ""
    is_exported: bool = False
    is_literal: bool = False  # Literal type (iota, const)
    group: str = ""  # For iota groups
    docstring: str = ""


@dataclass
class ImportInfo:
    """Information about an import."""
    path: str
    alias: str = ""
    is_standard: bool = False
    is_external: bool = False
    is_local: bool = False


@dataclass
class EndpointInfo:
    """Information about an HTTP endpoint."""
    path: str
    method: str = "GET"
    handler: str = ""
    middleware: List[str] = field(default_factory=list)
    auth_required: bool = False
    rate_limited: bool = False
    description: str = ""
    request_body: str = ""
    response_body: str = ""
    status_codes: List[int] = field(default_factory=list)


@dataclass
class SecurityPattern:
    """Security-sensitive code pattern."""
    pattern_type: str  # SQL injection, XSS, command injection, etc.
    severity: str = "medium"  # low, medium, high, critical
    line_number: int = 0
    description: str = ""
    suggestion: str = ""


@dataclass
class ConfigVarInfo:
    """Information about a configuration variable."""
    name: str = ""
    env_var: str = ""
    default: str = ""
    required: bool = False
    description: str = ""


@dataclass
class ModuleData:
    """Complete metadata for a module."""
    path: str
    name: str
    package: str = ""
    language: Language = Language.UNKNOWN
    imports: List[ImportInfo] = field(default_factory=list)
    functions: List[FunctionInfo] = field(default_factory=list)
    types: List[TypeInfo] = field(default_factory=list)
    constants: List[ConstantInfo] = field(default_factory=list)
    endpoints: List[EndpointInfo] = field(default_factory=list)
    security_patterns: List[SecurityPattern] = field(default_factory=list)
    config_vars: List[ConfigVarInfo] = field(default_factory=list)
    complexity_score: int = 0
    maintainability_index: float = 100.0


class MetadataExtractor:
    """
    Enhanced metadata extractor with robust regex and multi-language support.
    
    Features:
    - Comprehensive regex with error handling
    - Docstring extraction
    - Security pattern detection
    - API endpoint analysis
    - Complexity metrics
    - Concurrent processing
    """

    # Standard library packages by language
    STANDARD_PACKAGES: Dict[Language, Set[str]] = {
        Language.GO: {
            'fmt', 'os', 'io', 'net', 'http', 'json', 'time', 'context',
            'strings', 'bytes', 'errors', 'log', 'sync', 'math', 'rand',
            'crypto', 'hash', 'encoding', 'regexp', 'path', 'flag',
            'os/exec', 'io/ioutil', 'bufio', 'container/list',
            'container/ring', 'net/http', 'encoding/json', 'testing',
        },
        Language.PYTHON: {
            'os', 'sys', 'time', 'datetime', 'json', 'logging',
            'typing', 'collections', 'functools', 'itertools',
            'pathlib', 'abc', 'copy', 're', 'math', 'random',
        },
        Language.TYPESCRIPT: {
            'assert', 'buffer', 'child_process', 'cluster',
            'console', 'constants', 'crypto', 'dgram', 'dns',
            'events', 'fs', 'http', 'https', 'net', 'os',
            'path', 'process', 'punycode', 'querystring',
        },
        Language.RUST: {
            'std', 'core', 'alloc', 'rustc_deprecated_since',
            'proc_macro', 'builtin', 'self',
        },
    }

    # Security patterns to detect (language-agnostic)
    SECURITY_PATTERNS: List[Tuple[str, str, str, str]] = [
        # SQL Injection
        (r'execute\s*\(\s*["\'].*?%s', 'sql_injection', 'high',
         'Use parameterized queries instead of string formatting'),
        (r'["\'].*?\.format\s*\(.*?s', 'sql_injection', 'medium',
         'Avoid format strings in SQL queries'),
        
        # Command Injection
        (r'exec\s*\(\s*["\'].*?subprocess', 'command_injection', 'high',
         'Avoid shell execution with user input'),
        (r'system\s*\(\s*["\']', 'command_injection', 'high',
         'Avoid system() calls with user input'),
        
        # XSS
        (r'innerHTML\s*=', 'xss', 'high',
         'Use textContent instead of innerHTML'),
        (r'dangerouslySetInnerHTML', 'xss', 'critical',
         'Avoid dangerouslySetInnerHTML in React'),
        
# Path Traversal
        (r'\.\./', 'path_traversal', 'medium',
         'Validate and sanitize file paths'),
        (r'\.\.\\\\', 'path_traversal', 'medium',
         'Validate Windows file paths'),
        
        # Weak Crypto
        (r'md5', 'weak_crypto', 'medium',
         'Use SHA-256 or stronger for hashing'),
        (r'sha1', 'weak_crypto', 'medium',
         'SHA-1 is considered weak'),
        (r'DES', 'weak_crypto', 'medium',
         'DES is considered weak'),
        
        # Hardcoded Secrets
        (r'password\s*=\s*["\']', 'hardcoded_secret', 'critical',
         'Move secrets to environment variables'),
        (r'secret\s*=\s*["\']', 'hardcoded_secret', 'critical',
         'Move secrets to environment variables'),
        (r'api[_-]?key\s*=\s*["\']', 'hardcoded_secret', 'critical',
         'Use API key management service'),
        (r'token\s*=\s*["\']', 'hardcoded_secret', 'high',
         'Avoid hardcoded tokens'),
        
        # JWT issues
        (r'alg\s*:\s*["\']none["\']', 'jwt_alg_none', 'critical',
         'JWT alg:none is insecure'),
        (r'Algorithm\s*=\s*["\']none["\']', 'jwt_alg_none', 'critical',
         'JWT alg:none is insecure'),
        
        # Deserialization
        (r'pickle\.loads', 'unsafe_deserialization', 'critical',
         'Pickle is insecure, use JSON'),
        (r'yaml\.load', 'unsafe_deserialization', 'high',
         'Use yaml.safe_load'),
        
        # Random weak
        (r'random\.random', 'weak_random', 'medium',
         'Use cryptographically secure random'),
        (r'Math\.random', 'weak_random', 'medium',
         'Use Web Crypto API for security'),
    ]

    # Pre-compiled regex patterns for performance
    _COMPILED_PATTERNS = {
        'package': re.compile(r'^package\s+(\w+)', re.MULTILINE),
        'import_block': re.compile(r'import\s*\((.*?)\)', re.DOTALL),
        'const_block': re.compile(r'const\s*\((.*?)\)', re.DOTALL),
        'struct': re.compile(r'type\s+(\w+)\s+struct\s*\{'),
        'interface': re.compile(r'type\s+(\w+)\s+interface\s*\{'),
        'func_method': re.compile(
            r'func\s+\(([^)]+)\)\s+(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]+)\))?\s*\{'
        ),
        'func_standalone': re.compile(r'func\s+(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]+)\))?\s*\{'),
        'def_func': re.compile(r'^def\s+(\w+)\s*\(([^)]*)\):', re.MULTILINE),
        'class': re.compile(r'^class\s+(\w+)(?:\(([^)]+)\))?:', re.MULTILINE),
        # Extended patterns
        'import_alias': re.compile(r'(\w+)\s+"([^"]+)"'),
        'field_match': re.compile(r'(\w+)\s+(\*?[\w\[\]]+(?:\[[\d]*\])?)\s*(?:`([^`]+)`)?'),
        'method_sig': re.compile(r'(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?'),
        'enum_pattern': re.compile(r'const\s*\((.*?)type\s+(\w+)\s+', re.DOTALL),
        'python_import': re.compile(r'^\s*(?:from\s+(\S+)\s+)?import\s+([^\n#]+)', re.MULTILINE),
        'ts_import': re.compile(r"(?:import|export)\s+(?:type\s+)?(?:{([^}]+)}|(\w+))\s+from\s+['\"]([^'\"]+)['\"]"),
        'ts_class': re.compile(r'class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+([^\n{]+))?'),
        'ts_interface': re.compile(r'interface\s+(\w+)\s*\{([^}]+)\}'),
        'ts_field': re.compile(r'(\w+)(\??):\s*([^;]+)'),
        'rust_use': re.compile(r'use\s+(?:(?:pub\s+))?(\w+(?:::\w+)*)(?:\s+as\s+(\w+))?'),
        'rust_fn': re.compile(r'(?:pub\s+)?fn\s+(\w+)<([^>]*)>\(([^)]*)\)(?:\s*->\s*([^({]+))?'),
        'rust_struct': re.compile(r'(?:pub\s+)?struct\s+(\w+)(?::\s*([^\n{]+))?\{'),
        'rust_trait': re.compile(r'(?:pub\s+)?trait\s+(\w+)(?::\s*([^\n{]+))?\{'),
    }

    def __init__(self, root_path: str):
        self.root_path = Path(root_path)
        if not self.root_path.exists():
            raise ExtractorError(f"Root path does not exist: {root_path}")

    def extract_all(self, files: List, max_workers: int = 4) -> Dict[str, ModuleData]:
        """
        Extract metadata from all source files concurrently.
        
        Args:
            files: List of FileInfo objects from analyzer
            max_workers: Maximum worker threads
            
        Returns:
            Dictionary mapping file paths to ModuleData
        """
        modules = {}

        # Filter source files by extension
        source_files = [
            f for f in files
            if f.extension in ('.go', '.py', '.ts', '.tsx', '.js', '.jsx', '.rs')
        ]

        logger.info(f"Extracting metadata from {len(source_files)} files")

        # Process concurrently
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            future_to_file = {
                executor.submit(self._extract_file_safe, f): f
                for f in source_files
            }

            for future in as_completed(future_to_file):
                file_info = future_to_file[future]
                try:
                    module_data = future.result()
                    if module_data:
                        modules[file_info.path] = module_data
                except Exception as e:
                    logger.warning(f"Failed to extract {file_info.path}: {e}")

        return modules

    def _extract_file_safe(self, file_info: Any) -> Optional[ModuleData]:
        """Safely extract metadata from a file."""
        try:
            return self._extract_file(file_info.path, file_info.extension)
        except Exception as e:
            logger.error(f"Extraction error for {file_info.path}: {e}")
            return None

    def _extract_file(self, file_path: str, extension: str) -> Optional[ModuleData]:
        """Extract metadata from source file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except (OSError, UnicodeDecodeError):
            return None

        module_data = ModuleData(
            path=file_path,
            name=Path(file_path).stem
        )

        # Determine language and extract
        if extension == '.go':
            self._extract_go(content, module_data)
        elif extension in ('.py',):
            self._extract_python(content, module_data)
        elif extension in ('.ts', '.tsx', '.js', '.jsx'):
            self._extract_typescript(content, module_data, extension)
        elif extension == '.rs':
            self._extract_rust(content, module_data)

        # Extract security patterns (language-agnostic)
        module_data.security_patterns = self._extract_security_patterns(content)

        # Calculate complexity
        module_data.complexity_score = self._calculate_complexity(content, module_data)

        return module_data

    # ========== Go Extraction ==========

    def _extract_go(self, content: str, module_data: ModuleData):
        """Extract metadata from Go source."""
        module_data.language = Language.GO

        # Package
        pkg_match = re.search(r'^package\s+(\w+)', content, re.MULTILINE)
        if pkg_match:
            module_data.package = pkg_match.group(1)

        # Imports
        module_data.imports = self._extract_go_imports(content)

        # Constants
        module_data.constants = self._extract_go_constants(content)

        # Types
        module_data.types = self._extract_go_types(content)

        # Functions
        module_data.functions = self._extract_go_functions(content)

        # Endpoints
        module_data.endpoints = self._extract_go_endpoints(content)

        # Config variables
        module_data.config_vars = self._extract_go_config_vars(content)

    def _extract_go_config_vars(self, content: str) -> List[ConfigVarInfo]:
        """Extract environment variable configuration."""
        config_vars = []
        # Look for patterns like: LAKERA_API_KEY, JWT_SECRET, etc.
        patterns = [
            r'(LAKERA_\w+)',
            r'(JWT_\w+)',
            r'(RATE_LIMIT_\w+)',
            r'(MCP_\w+)',
            r'(LISTEN_\w+)',
            r'(CORS_\w+)',
            r'(TLS_\w+)',
            r'(DLQ_\w+)',
        ]
        found = set()
        for pattern in patterns:
            for match in re.finditer(pattern, content):
                var = match.group(1)
                if var not in found:
                    found.add(var)
                    config_vars.append(ConfigVarInfo(
                        name=var.lower(),
                        env_var=var,
                        default="",
                        required=True,
                        description=f"Environment variable {var}"
                    ))
        return config_vars

    def _extract_go_imports(self, content: str) -> List[ImportInfo]:
        """Extract Go imports with classification."""
        imports = []

        # Find import block
        import_block = re.search(r'import\s*\((.*?)\)', content, re.DOTALL)
        if import_block:
            block = import_block.group(1)
            for line in block.split('\n'):
                line = line.strip().strip('"')
                if not line or line.startswith('//'):
                    continue

                # Aliased import
                alias_match = re.match(r'(\w+)\s+"([^"]+)"', line)
                if alias_match:
                    imp = ImportInfo(
                        path=alias_match.group(2),
                        alias=alias_match.group(1)
                    )
                else:
                    imp = ImportInfo(path=line)

                # Classify
                imp.is_standard = imp.path.split('/')[0] in self.STANDARD_PACKAGES[Language.GO]
                imp.is_external = not imp.is_standard and not imp.path.startswith('.')
                imp.is_local = imp.path.startswith('.')

                imports.append(imp)

        return imports

    def _extract_go_constants(self, content: str) -> List[ConstantInfo]:
        """Extract Go constants with groups."""
        constants = []
        iota_group = {}

        # Find const block
        const_block = re.search(r'const\s*\((.*?)\)', content, re.DOTALL)
        if const_block:
            block = const_block.group(1)
            iota_value = 0

            for line in block.split('\n'):
                line = line.strip()
                if not line or line.startswith('//'):
                    continue

                # Parse: name = value or name type = value
                const_match = re.match(r'(\w+)(?:\s+(\w+))?\s*=\s*(.+)', line)
                if const_match:
                    name = const_match.group(1)
                    type_hint = const_match.group(2) or ""
                    value = const_match.group(3).rstrip(',')

                    is_exported = name[0].isupper() if name else False

                    # Track iota groups
                    if value == 'iota':
                        iota_group[name] = iota_value
                        iota_value += 1

                    constants.append(ConstantInfo(
                        name=name,
                        value=value,
                        type_hint=type_hint,
                        is_exported=is_exported,
                        is_literal=value in ('iota', 'true', 'false')
                    ))

        # Simple constants
        for match in re.finditer(r'const\s+(\w+)\s*=\s*(.+)', content):
            name = match.group(1)
            value = match.group(2)

            if not any(c.name == name for c in constants):
                constants.append(ConstantInfo(
                    name=name,
                    value=value,
                    is_exported=name[0].isupper()
                ))

        return constants

    def _extract_go_types(self, content: str) -> List[TypeInfo]:
        """Extract Go types, structs, interfaces."""
        types = []

        # Structs with fields
        struct_pattern = r'type\s+(\w+)\s+struct\s*\{'
        for match in re.finditer(struct_pattern, content):
            name = match.group(1)
            start = match.end()

            # Find matching brace
            end = self._find_matching_brace(content, start - 1)
            body = content[start:end]

            fields = []
            for line in body.split('\n'):
                line = line.strip()
                if not line or line.startswith('//'):
                    continue

                # Parse field: Name Type `tag`
                field_match = re.match(
                    r'(\w+)\s+(\*?[\w\[\]]+(?:\[[\d]*\])?)\s*(?:`([^`]+)`)?',
                    line
                )
                if field_match:
                    fields.append((
                        field_match.group(1),
                        field_match.group(2),
                        field_match.group(3) or ''
                    ))

            line_number = content[:match.start()].count('\n') + 1

            types.append(TypeInfo(
                name=name,
                kind='struct',
                fields=fields,
                is_exported=name[0].isupper(),
                is_error=name in ('error', 'Error'),
                line_number=line_number,
                end_line=content[:end].count('\n')
            ))

        # Interfaces
        interface_pattern = r'type\s+(\w+)\s+interface\s*\{'
        for match in re.finditer(interface_pattern, content):
            name = match.group(1)
            start = match.end()
            end = self._find_matching_brace(content, start - 1)
            body = content[start:end]

            methods = []
            for line in body.split('\n'):
                line = line.strip()
                if not line or line.startswith('//'):
                    continue

                method_match = re.match(r'(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]*)\))?', line)
                if method_match:
                    methods.append(method_match.group(1))

            line_number = content[:match.start()].count('\n') + 1

            types.append(TypeInfo(
                name=name,
                kind='interface',
                methods=methods,
                is_exported=name[0].isupper(),
                line_number=line_number,
                end_line=content[:end].count('\n')
            ))

        # Enums (iota groups)
        enum_pattern = r'const\s*\((.*?)type\s+(\w+)\s+'
        for match in re.finditer(enum_pattern, content, re.DOTALL):
            enum_type = match.group(2)
            line_number = content[:match.start()].count('\n') + 1

            types.append(TypeInfo(
                name=enum_type,
                kind='enum',
                is_exported=enum_type[0].isupper(),
                line_number=line_number
            ))

        return types

    def _extract_go_functions(self, content: str) -> List[FunctionInfo]:
        """Extract Go functions and methods."""
        functions = []

        # Methods with receiver: func (r *Type) Name(params) return_type {
        method_pattern = (
            r'func\s+\(([^)]+)\)\s+(\w+)\s*\(([^)]*)\)\s*'
            r'(?:\(([^)]+)\))?\s*\{'
        )

        for match in re.finditer(method_pattern, content):
            receiver = match.group(1)
            name = match.group(2)
            params_str = match.group(3) or ""
            return_type = match.group(4) or ""

            line_number = content[:match.start()].count('\n') + 1

            params = self._parse_go_params(params_str)
            is_exported = name[0].isupper() if name else False

            functions.append(FunctionInfo(
                name=name,
                params=params,
                return_type=return_type,
                is_exported=is_exported,
                is_handler='Handler' in name or name.endswith('Handler') or name.endswith('Func'),
                is_test=name.startswith('Test') or name.startswith('Benchmark'),
                is_method=True,
                line_number=line_number
            ))

        # Standalone functions: func Name(params) return_type {
        func_pattern = r'func\s+(\w+)\s*\(([^)]*)\)\s*(?:\(([^)]+)\))?\s*\{'

        for match in re.finditer(func_pattern, content):
            name = match.group(1)
            params_str = match.group(2) or ""
            return_type = match.group(3) or ""

            line_number = content[:match.start()].count('\n') + 1

            params = self._parse_go_params(params_str)
            is_exported = name[0].isupper() if name else False

            functions.append(FunctionInfo(
                name=name,
                params=params,
                return_type=return_type,
                is_exported=is_exported,
                is_handler='Handler' in name or name.endswith('Handler') or name.endswith('Func'),
                is_test=name.startswith('Test') or name.startswith('Benchmark'),
                line_number=line_number
            ))

        return functions

    def _parse_go_params(self, params_str: str) -> List[Tuple[str, str]]:
        """Parse Go function parameters."""
        params = []
        if not params_str.strip():
            return params

        for param in params_str.split(','):
            param = param.strip()
            if not param:
                continue

            parts = param.split()
            if len(parts) >= 2:
                name = parts[0]
                type_str = ' '.join(parts[1:])
            else:
                name = ""
                type_str = parts[0]

            params.append((name, type_str))

        return params

    def _extract_go_endpoints(self, content: str) -> List[EndpointInfo]:
        """Extract Go HTTP endpoints."""
        endpoints = []

        # mux.Handle, router.Handle, http.HandleFunc
        handle_patterns = [
            r'(?:mux|router|http)\.Handle(?:Func)?\(\s*"([^"]+)"\s*,\s*(\w+)',
            r'http\.Handle(?:Func)?\s*\(\s*"([^"]+)"\s*,\s*(\w+)',
            r'@router\s+(\S+)\s+(\w+)',
        ]

        for pattern in handle_patterns:
            for match in re.finditer(pattern, content):
                path = match.group(1)
                handler = match.group(2)

                # Determine method
                method = "GET"
                context = content[max(0, match.start()-200):match.start()+200]
                if "POST" in context:
                    method = "POST"
                elif "PUT" in context:
                    method = "PUT"
                elif "DELETE" in context:
                    method = "DELETE"
                elif "PATCH" in context:
                    method = "PATCH"

                endpoints.append(EndpointInfo(
                    path=path,
                    method=method,
                    handler=handler
                ))

        return endpoints

    # ========== Python Extraction ==========

    def _extract_python(self, content: str, module_data: ModuleData):
        """Extract metadata from Python source."""
        module_data.language = Language.PYTHON

        # Module name (filename)
        module_data.name = Path(module_data.path).stem

        # Try AST-based extraction for better accuracy
        try:
            tree = ast.parse(content)
            module_data.package = tree.body[0].names[0].name if tree.body and isinstance(tree.body[0], ast.ImportFrom) else ""
        except SyntaxError:
            pass

        # Imports
        for match in re.finditer(r'^\s*(?:from\s+(\S+)\s+)?import\s+([^\n#]+)', content, re.MULTILINE):
            module_path = match.group(1) or match.group(2)
            imports = match.group(2).split(',')

            for imp in imports:
                imp = imp.strip()
                if not imp or imp.startswith('#'):
                    continue

                alias = ""
                if ' as ' in imp:
                    parts = imp.split(' as ')
                    imp = parts[0].strip()
                    alias = parts[1].strip()

                imp_info = ImportInfo(
                    path=module_path + "." + imp if module_path else imp,
                    alias=alias
                )
                imp_info.is_standard = not '.' in (module_path or imp)
                module_data.imports.append(imp_info)

        # Functions
        for match in re.finditer(r'^def\s+(\w+)\s*\(([^)]*)\):', content, re.MULTILINE):
            name = match.group(1)
            params_str = match.group(2)
            params = [(p.strip(), "") for p in params_str.split(',') if p.strip()]

            line_number = content[:match.start()].count('\n') + 1
            is_exported = not name.startswith('_')

            module_data.functions.append(FunctionInfo(
                name=name,
                params=params,
                is_exported=is_exported,
                is_handler='Handler' in name or name.endswith('Handler'),
                is_test=name.startswith('test_') or name.startswith('Test'),
                is_async='async ' in content[match.start():match.start()+10],
                line_number=line_number
            ))

        # Classes
        class_pattern = r'^class\s+(\w+)(?:\(([^)]+)\))?:'
        for match in re.finditer(class_pattern, content, re.MULTILINE):
            name = match.group(1)
            bases = match.group(2) or ""

            line_number = content[:match.start()].count('\n') + 1
            is_exported = not name.startswith('_')
            is_error = name.endswith('Error')

            module_data.types.append(TypeInfo(
                name=name,
                kind='class',
                implements=[b.strip() for b in bases.split(',') if b.strip()],
                is_exported=is_exported,
                is_error=is_error,
                line_number=line_number
            ))

    # ========== TypeScript Extraction ==========

    def _extract_typescript(self, content: str, module_data: ModuleData, extension: str):
        """Extract metadata from TypeScript/JavaScript."""
        module_data.language = (
            Language.TYPESCRIPT if extension in ('.ts', '.tsx')
            else Language.JAVASCRIPT
        )
        module_data.name = Path(module_data.path).stem

        # Imports
        import_pattern = r"(?:import|export)\s+(?:type\s+)?(?:{([^}]+)}|(\w+))\s+from\s+['\"]([^'\"]+)['\"]"
        for match in re.finditer(import_pattern, content):
            spec = match.group(1) or match.group(2)
            path = match.group(3)

            for name in spec.split(','):
                name = name.strip()
                if name:
                    module_data.imports.append(ImportInfo(
                        path=path,
                        alias=name,
                        is_external=not path.startswith('.')
                    ))

        # Functions: const/function/arrow
        func_patterns = [
            r'(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)\s*=>',
            r'(?:export\s+)?function\s+(\w+)\s*\(([^)]*)\)',
        ]

        for pattern in func_patterns:
            for match in re.finditer(pattern, content):
                name = match.group(1)
                params_str = match.group(2) or ""

                params = [(p.strip(), "") for p in params_str.split(',') if p.strip()]
                line_number = content[:match.start()].count('\n') + 1

                module_data.functions.append(FunctionInfo(
                    name=name,
                    params=params,
                    is_exported='export' in content[max(0, match.start()-20):match.start()],
                    is_handler='Handler' in name,
                    line_number=line_number
                ))

        # Classes
        class_pattern = r'class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+([^\n{]+))?'
        for match in re.finditer(class_pattern, content):
            name = match.group(1)
            extends = match.group(2) or ""
            implements = match.group(3) or ""

            line_number = content[:match.start()].count('\n') + 1

            module_data.types.append(TypeInfo(
                name=name,
                kind='class',
                extends=extends,
                implements=[i.strip() for i in implements.split(',') if i.strip()],
                is_exported=True,
                line_number=line_number
            ))

        # Interfaces
        interface_pattern = r'interface\s+(\w+)\s*\{([^}]+)\}'
        for match in re.finditer(interface_pattern, content):
            name = match.group(1)
            body = match.group(2)

            fields = []
            for line in body.split('\n'):
                line = line.strip()
                if not line or line.startswith('//'):
                    continue

                field_match = re.match(r'(\w+)(\??):\s*([^;]+)', line)
                if field_match:
                    fields.append((field_match.group(1), field_match.group(3), ""))

            module_data.types.append(TypeInfo(
                name=name,
                kind='interface',
                fields=fields,
                is_exported=True,
                line_number=content[:match.start()].count('\n') + 1
            ))

    # ========== Rust Extraction ==========

    def _extract_rust(self, content: str, module_data: ModuleData):
        """Extract metadata from Rust source."""
        module_data.language = Language.RUST
        module_data.name = Path(module_data.path).stem

        # Module declaration
        mod_match = re.search(r'^module\s+(\w+)', content, re.MULTILINE)
        if mod_match:
            module_data.package = mod_match.group(1)

        # Imports (use statements)
        use_pattern = r'use\s+(?:(?:pub\s+))?(\w+(?:::\w+)*)(?:\s+as\s+(\w+))?'
        for match in re.finditer(use_pattern, content):
            path = match.group(1)
            alias = match.group(2) or ""

            imp = ImportInfo(path=path, alias=alias)
            imp.is_standard = path.split('::')[0] in self.STANDARD_PACKAGES[Language.RUST]
            module_data.imports.append(imp)

        # Functions
        func_pattern = r'(?:pub\s+)?fn\s+(\w+)<([^>]*)>\(([^)]*)\)(?:\s*->\s*([^({]+))?'
        for match in re.finditer(func_pattern, content):
            name = match.group(1)
            params_str = match.group(3) or ""
            return_type = match.group(4) or ""

            params = []
            for param in params_str.split(','):
                param = param.strip()
                if param:
                    parts = param.split(':')
                    if len(parts) == 2:
                        params.append((parts[0].strip(), parts[1].strip()))

            module_data.functions.append(FunctionInfo(
                name=name,
                params=params,
                return_type=return_type,
                is_exported='pub' in content[max(0, match.start()-10):match.start()],
                line_number=content[:match.start()].count('\n') + 1
            ))

        # Structs
        struct_pattern = r'(?:pub\s+)?struct\s+(\w+)(?::\s*([^\n{]+))?\{'
        for match in re.finditer(struct_pattern, content):
            name = match.group(1)
            generics = match.group(2) or ""

            module_data.types.append(TypeInfo(
                name=name,
                kind='struct',
                is_exported='pub' in content[max(0, match.start()-10):match.start()],
                line_number=content[:match.start()].count('\n') + 1
            ))

        # Traits
        trait_pattern = r'(?:pub\s+)?trait\s+(\w+)(?::\s*([^\n{]+))?\{'
        for match in re.finditer(trait_pattern, content):
            name = match.group(1)

            module_data.types.append(TypeInfo(
                name=name,
                kind='trait',
                is_exported='pub' in content[max(0, match.start()-10):match.start()],
                line_number=content[:match.start()].count('\n') + 1
            ))

    # ========== Security Patterns ==========

    def _extract_security_patterns(self, content: str) -> List[SecurityPattern]:
        """Extract security-sensitive patterns."""
        patterns = []

        for pattern, ptype, severity, suggestion in self.SECURITY_PATTERNS:
            for match in re.finditer(pattern, content, re.IGNORECASE | re.MULTILINE):
                line_number = content[:match.start()].count('\n') + 1

                patterns.append(SecurityPattern(
                    pattern_type=ptype,
                    severity=severity,
                    line_number=line_number,
                    description=f"Potential {ptype.replace('_', ' ')}",
                    suggestion=suggestion
                ))

        return patterns

    # ========== Complexity ==========

    def _calculate_complexity(self, content: str, module_data: ModuleData) -> int:
        """Calculate code complexity score."""
        complexity = 0

        # Function complexity
        complexity += len(module_data.functions) * 2

        # Type complexity
        complexity += len(module_data.types)

        # Cyclomatic complexity indicators
        complexity += len(re.findall(r'\b(if|for|switch|case|match)\b', content)) // 10

        # Nesting depth estimation
        max_nesting = max(
            line.count('{') - line.count('}')
            for line in content.split('\n')
        )
        complexity += max_nesting * 3

        return min(complexity, 100)

    # ========== Utilities ==========

    def _find_matching_brace(self, content: str, start: int) -> int:
        """Find matching closing brace."""
        depth = 1
        for i in range(start + 1, len(content)):
            if content[i] == '{':
                depth += 1
            elif content[i] == '}':
                depth -= 1
                if depth == 0:
                    return i
        return len(content)


def extract_metadata(root_path: str, files: List, **kwargs) -> Dict[str, ModuleData]:
    """
    Convenience function to extract metadata.
    
    Args:
        root_path: Path to repository root
        files: List of FileInfo objects
        **kwargs: Additional arguments
        
    Returns:
        Dictionary mapping file paths to ModuleData
    """
    extractor = MetadataExtractor(root_path)
    return extractor.extract_all(files, **kwargs)