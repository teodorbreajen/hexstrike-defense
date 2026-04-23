"""
Repository Analyzer - Enhanced Edition
===================================

Scans and analyzes repository structure with:
- Robust error handling
- File type detection via content analysis
- Language detection
- Module/package detection
- Configuration files
- Scripts
- Security validation

Author: HexStrike Documentation Team
Version: 2.1.0
"""

import os
import re
import hashlib
import logging
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple
from dataclasses import dataclass, field
from enum import Enum
from concurrent.futures import ThreadPoolExecutor, as_completed

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class Language(Enum):
    """Detected programming languages."""
    GO = "Go"
    PYTHON = "Python"
    TYPESCRIPT = "TypeScript"
    JAVASCRIPT = "JavaScript"
    RUST = "Rust"
    JAVA = "Java"
    C = "C"
    CPP = "C++"
    CSHARP = "C#"
    YAML = "YAML"
    TOML = "TOML"
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


class SecurityLevel(Enum):
    """Security classification for files."""
    PUBLIC = "public"
    INTERNAL = "internal"
    CONFIDENTIAL = "confidential"
    RESTRICTED = "restricted"


@dataclass
class FileSignature:
    """File signature data for content-based detection."""
    magic_bytes: bytes
    encoding: str = "utf-8"
    language_hints: List[str] = field(default_factory=list)


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
    lines_of_comments: int = 0
    lines_of_blank: int = 0
    has_tests: bool = False
    imports: List[str] = field(default_factory=list)
    exports: List[str] = field(default_factory=list)
    security_level: SecurityLevel = SecurityLevel.PUBLIC
    complexity_score: int = 0
    technical_debt_minutes: int = 0
    hash_md5: str = ""
    hash_sha256: str = ""
    last_modified: float = 0.0
    error_message: Optional[str] = None


@dataclass
class ModuleInfo:
    """Information about a code module/package."""
    name: str
    path: str
    language: Language = Language.UNKNOWN
    files: List[FileInfo] = field(default_factory=list)
    functions: List[str] = field(default_factory=list)
    types: List[str] = field(default_factory=list)
    interfaces: List[str] = field(default_factory=list)
    constants: List[str] = field(default_factory=list)
    dependencies: List[str] = field(default_factory=list)
    is_main_module: bool = False
    test_coverage_percent: float = 0.0


@dataclass
class RepoData:
    """Complete repository analysis data."""
    root_path: str
    project_name: str
    version: str = "1.0.0"
    description: str = ""
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
    vulnerabilities: List[Dict] = field(default_factory=list)
    technical_debt_minutes: int = 0
    analysis_errors: List[Dict] = field(default_factory=list)
    analysis_time_seconds: float = 0.0


class ValidationError(Exception):
    """Raised when path validation fails."""
    pass


class AnalysisError(Exception):
    """Raised when analysis fails."""
    pass


class RepoAnalyzer:
    """
    Enhanced repository analyzer with robust error handling.
    
    Features:
    - Content-based language detection
    - Security classification
    - Complexity scoring
    - Technical debt estimation
    - Concurrent file processing
    - File hashing
    - Comprehensive error handling
    """

    # File extension to language mapping
    EXTENSION_MAP: Dict[str, Language] = {
        ".go": Language.GO,
        ".py": Language.PYTHON,
        ".ts": Language.TYPESCRIPT,
        ".tsx": Language.TYPESCRIPT,
        ".js": Language.JAVASCRIPT,
        ".jsx": Language.JAVASCRIPT,
        ".rs": Language.RUST,
        ".java": Language.JAVA,
        ".c": Language.C,
        ".cpp": Language.CPP,
        ".cs": Language.CSHARP,
        ".yaml": Language.YAML,
        ".yml": Language.YAML,
        ".toml": Language.TOML,
        ".md": Language.MARKDOWN,
        ".sh": Language.BASH,
        ".bash": Language.BASH,
        ".dockerfile": Language.DOCKERFILE,
        ".json": Language.JSON,
    }

    # File type classification patterns
    TYPE_PATTERNS: Dict[FileType, List[str]] = {
        FileType.TEST: [
            r"_test\.go$", r"_test\.py$", r"_test\.ts$",
            r"test_.*\.sh$", r"\.test\.", r"\.spec\.",
            r"^tests?/", r"^__tests__/", r"^test_"
        ],
        FileType.CONFIG: [
            r"^config", r"\.config\.", r"settings\.",
            r"\.conf\.", r"\.ini$", r"\.toml$", r"\.yaml$"
        ],
        FileType.MANIFEST: [
            r"^manifests?/", r"^k8s?/", r"^deploy/",
            r"\.yaml$", r"\.yml$", r"^helm/"
        ],
        FileType.SCRIPT: [
            r"^scripts/", r"^tools/", r"\.sh$", r"\.bash$"
        ],
        FileType.DOCUMENTATION: [
            r"^docs?/", r"\.md$", r"README", r"CHANGELOG", r"LICENSE"
        ],
        FileType.WORKFLOW: [
            r"^\.github/workflows/", r"^ci/", r"^workflows/",
            r"\.yml$", r"\.yaml$"
        ],
        FileType.DEPENDENCY: [
            r"^go\.mod$", r"^go\.sum$",
            r"^package\.json$", r"^requirements\.txt$",
            r"^Pipfile$", r"^Cargo\.toml$", r"^pom\.xml$"
        ],
        FileType.MAKEFILE: [
            r"^Makefile$", r"^makefile$", r"^GNUmakefile$"
        ],
    }

    # Security-sensitive file patterns
    SECURITY_PATTERNS: List[str] = [
        r"secret", r"password", r"token", r"api.key",
        r"credentials", r"\.env$", r"\.key$",
        r"id_rsa", r"id_dsa", r"id_ecdsa"
    ]

    # Complexity scoring weights
    COMPLEXITY_WEIGHTS = {
        "nested_depth": 2,
        "cyclomatic": 1,
        "long_function": 3,
        "many_params": 2,
        "global_vars": 3,
    }

    def __init__(self, root_path: str, skip_patterns: Optional[List[str]] = None):
        """
        Initialize analyzer.
        
        Args:
            root_path: Path to repository root
            skip_patterns: List of directory patterns to skip
        
        Raises:
            ValidationError: If path is invalid
            AnalysisError: If analysis fails
        """
        self.root_path = self._validate_path(root_path)
        self.skip_patterns = skip_patterns or [
            '.git', 'vendor', 'node_modules', '__pycache__',
            '.venv', 'dist', 'build', '.idea', '.vscode',
            'generated', 'docs', 'tools'
        ]
        self.project_name = self.root_path.name
        self._initialize_project_metadata()

    def _validate_path(self, path: str) -> Path:
        """Validate and normalize path."""
        try:
            p = Path(path).resolve()
            if not p.exists():
                raise ValidationError(f"Path does not exist: {path}")
            if not p.is_dir():
                raise ValidationError(f"Path is not a directory: {path}")
            return p
        except Exception as e:
            logger.error(f"Path validation failed: {e}")
            raise ValidationError(f"Invalid path: {path}") from e

    def _initialize_project_metadata(self):
        """Initialize project metadata from existing files."""
        # Try to read version from common files
        version_patterns = [
            (self.root_path / "VERSION", r"(\d+\.\d+\.\d+)"),
            (self.root_path / "version.txt", r"(\d+\.\d+\.\d+)"),
            (self.root_path / "pyproject.toml", r'version\s*=\s*"(\d+\.\d+\.\d+)"'),
        ]
        
        for version_file, pattern in version_patterns:
            if version_file.exists():
                try:
                    content = version_file.read_text(encoding='utf-8')
                    match = re.search(pattern, content)
                    if match:
                        self.project_version = match.group(1)
                        break
                except Exception:
                    continue
        
        # Try to read description from README
        readme_path = self.root_path / "README.md"
        if readme_path.exists():
            try:
                content = readme_path.read_text(encoding='utf-8', errors='ignore')
                # Get first non-title paragraph
                lines = content.split('\n')
                in_description = False
                for line in lines[1:]:
                    if line.startswith('#'):
                        break
                    if line.strip() and not line.startswith('```'):
                        in_description = True
                        self.project_description = line.strip()[:200]
                        break
            except Exception:
                pass

    def analyze(self, max_workers: int = 4) -> RepoData:
        """
        Perform complete repository analysis with concurrent processing.
        
        Args:
            max_workers: Maximum number of worker threads
            
        Returns:
            RepoData with all analysis results
            
        Raises:
            AnalysisError: If analysis fails
        """
        import time
        start_time = time.time()
        
        logger.info(f"Starting analysis of {self.root_path}")
        
        try:
            data = RepoData(
                root_path=str(self.root_path),
                project_name=self.project_name,
                version=getattr(self, 'project_version', '1.0.0'),
                description=getattr(self, 'project_description', '')
            )

            # Collect all files first
            all_files = self._collect_files()
            logger.info(f"Found {len(all_files)} files to analyze")

            # Process files concurrently
            with ThreadPoolExecutor(max_workers=max_workers) as executor:
                future_to_file = {
                    executor.submit(self._analyze_file, file_path, self.root_path): file_path
                    for file_path in all_files
                }

                for future in as_completed(future_to_file):
                    file_path = future_to_file[future]
                    try:
                        file_info = future.result()
                        if file_info:
                            data.files.append(file_info)
                            self._classify_file(file_info, data)
                    except Exception as e:
                        error_msg = str(e)
                        logger.warning(f"Failed to analyze {file_path}: {error_msg}")
                        data.analysis_errors.append({
                            "file": str(file_path),
                            "error": error_msg
                        })

            # Detect languages and parse Makefile
            data.languages = self._detect_languages(data.files)
            data.makefile_targets = self._parse_makefile()
            
            # Calculate technical debt
            data.technical_debt_minutes = self._calculate_technical_debt(data)

            # Update analysis time
            data.analysis_time_seconds = time.time() - start_time
            
            logger.info(f"Analysis completed in {data.analysis_time_seconds:.2f}s")
            logger.info(f"Analysis errors: {len(data.analysis_errors)}")
            
            return data

        except Exception as e:
            logger.error(f"Analysis failed: {e}")
            raise AnalysisError(f"Repository analysis failed: {e}") from e

    def _collect_files(self) -> List[Path]:
        """Collect all files in repository."""
        all_files = []
        
        for root, dirs, files in os.walk(self.root_path):
            root_path = Path(root)

            # Filter directories in-place
            dirs[:] = [
                d for d in dirs
                if not d.startswith('.')
                and d not in self.skip_patterns
            ]

            # Add files
            for filename in files:
                if filename.startswith('.'):
                    continue
                all_files.append(root_path / filename)

        return all_files

    def _analyze_file(self, file_path: Path, root: Path) -> Optional[FileInfo]:
        """
        Analyze a single file with comprehensive detection.
        
        Args:
            file_path: Path to file
            root: Repository root
            
        Returns:
            FileInfo or None if analysis fails
        """
        try:
            stat = file_path.stat()
            relative = file_path.relative_to(root)
        except (OSError, PermissionError) as e:
            logger.debug(f"Cannot stat {file_path}: {e}")
            return None

        extension = file_path.suffix.lower()
        language = self.EXTENSION_MAP.get(extension, Language.UNKNOWN)

        # Detect language from content if needed
        if language == Language.UNKNOWN:
            language = self._detect_language_from_content(file_path)

        # Count lines with detailed metrics
        loc, loc_comments, loc_blank = self._count_lines_detailed(file_path)

        # Calculate hash
        hash_md5, hash_sha256 = self._calculate_hashes(file_path)

        # Detect security level
        security_level = self._detect_security_level(file_path, relative)

        # Estimate complexity
        complexity = self._estimate_complexity(file_path, loc)

        # Estimate technical debt
        debt = self._estimate_file_debt(file_path, complexity)

        return FileInfo(
            path=str(file_path),
            relative_path=str(relative),
            name=file_path.name,
            extension=extension,
            file_type=FileType.UNKNOWN,  # Will be classified later
            language=language,
            size_bytes=stat.st_size,
            lines_of_code=loc,
            lines_of_comments=loc_comments,
            lines_of_blank=loc_blank,
            has_tests='_test' in file_path.name,
            security_level=security_level,
            complexity_score=complexity,
            technical_debt_minutes=debt,
            hash_md5=hash_md5,
            hash_sha256=hash_sha256,
            last_modified=stat.st_mtime
        )

    def _detect_language_from_content(self, file_path: Path) -> Language:
        """Detect language from file content patterns."""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read(2048)

            # Go detection
            if re.search(r'^package\s+\w+', content, re.MULTILINE):
                if 'func ' in content and 'import (' in content:
                    return Language.GO

            # Python detection
            if re.search(r'^import\s+\w+', content, re.MULTILINE):
                if re.search(r'^def\s+\w+\s*\(', content, re.MULTILINE):
                    return Language.PYTHON

            # TypeScript/JavaScript detection
            if 'import ' in content and ' from ' in content:
                if re.search(r':\s*(string|number|boolean|any)\b', content):
                    return Language.TYPESCRIPT
                if 'const ' in content or 'let ' in content:
                    return Language.JAVASCRIPT

            # Rust detection
            if re.search(r'^fn\s+\w+', content, re.MULTILINE):
                if 'let mut' in content or 'impl ' in content:
                    return Language.RUST

            # YAML detection
            if content.startswith('---'):
                return Language.YAML
            if re.search(r'^[a-z]+:\s*$', content, re.MULTILINE):
                return Language.YAML

            # Dockerfile detection
            if re.search(r'^FROM\s+\w+', content, re.MULTILINE | re.IGNORECASE):
                return Language.DOCKERFILE

            # Bash detection
            if re.search(r'^#!.*bash', content, re.MULTILINE):
                return Language.BASH

        except (OSError, UnicodeDecodeError):
            pass

        return Language.UNKNOWN

    def _count_lines_detailed(self, file_path: Path) -> Tuple[int, int, int]:
        """Count total, comment, and blank lines."""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()

            lines = content.split('\n')
            total = len(lines)

            # Count comments and blanks
            in_block_comment = False
            comments = 0
            blank = 0

            for line in lines:
                stripped = line.strip()

                if not stripped:
                    blank += 1
                    continue

                # Block comments (Go, C, C++, Java)
                if '/*' in stripped and '*/' in stripped:
                    comments += 1
                    continue
                if '/*' in stripped:
                    in_block_comment = True
                    comments += 1
                    continue
                if in_block_comment:
                    comments += 1
                    if '*/' in stripped:
                        in_block_comment = False
                    continue

                # Single-line comments
                if stripped.startswith('//'):
                    comments += 1
                    continue
                if stripped.startswith('#'):
                    comments += 1
                    continue

            return total - blank, comments, blank

        except (OSError, UnicodeDecodeError):
            return 0, 0, 0

    def _calculate_hashes(self, file_path: Path) -> Tuple[str, str]:
        """Calculate MD5 and SHA256 hashes."""
        try:
            content = file_path.read_bytes()
            return (
                hashlib.md5(content).hexdigest(),
                hashlib.sha256(content).hexdigest()
            )
        except (OSError, IOError):
            return "", ""

    def _detect_security_level(self, file_path: Path, relative: Path) -> SecurityLevel:
        """Detect security classification."""
        path_str = str(relative).lower()
        name = file_path.name.lower()

        # Check for security-sensitive files
        for pattern in self.SECURITY_PATTERNS:
            if re.search(pattern, path_str):
                return SecurityLevel.RESTRICTED

        # Check directory
        if path_str.startswith('secret') or path_str.startswith('private'):
            return SecurityLevel.CONFIDENTIAL
        if path_str.startswith('internal'):
            return SecurityLevel.INTERNAL

        return SecurityLevel.PUBLIC

    def _estimate_complexity(self, file_path: Path, loc: int) -> int:
        """Estimate code complexity score."""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()

            complexity = 0

            # Nested depth
            max_depth = max(
                len(re.findall(r'\{', line))
                for line in content.split('\n')
            )
            complexity += max_depth * self.COMPLEXITY_WEIGHTS["nested_depth"]

            # Cyclomatic complexity (simplified)
            complexity += len(re.findall(r'\b(if|for|switch|case)\b', content))

            # Long lines
            for line in content.split('\n'):
                if len(line) > 120:
                    complexity += self.COMPLEXITY_WEIGHTS["long_function"]

            return min(complexity, 100)  # Cap at 100

        except Exception:
            return 0

    def _estimate_file_debt(self, file_path: Path, complexity: int) -> int:
        """Estimate technical debt in minutes."""
        # Simple estimation based on complexity and file size
        try:
            stat = file_path.stat()
            size_kb = stat.st_size / 1024

            # Base debt: 5 minutes per 100 lines of code
            base_debt = (size_kb / 4) * 5

            # Complexity penalty
            complexity_penalty = complexity * 2

            return int(base_debt + complexity_penalty)

        except Exception:
            return 0

    def _classify_file(self, file_info: FileInfo, data: RepoData):
        """Classify file into appropriate category."""
        path = file_info.relative_path
        name = file_info.name.lower()

        # Check by type patterns
        for ftype, patterns in self.TYPE_PATTERNS.items():
            for pattern in patterns:
                if re.search(pattern, path, re.IGNORECASE):
                    file_info.file_type = ftype
                    break

        # Fallback: classify by extension
        if file_info.file_type == FileType.UNKNOWN:
            if file_info.extension in ('.go', '.py', '.ts', '.js', '.rs'):
                file_info.file_type = FileType.SOURCE
            elif file_info.extension in ('.yaml', '.yml', '.json'):
                file_info.file_type = FileType.MANIFEST

        # Add to appropriate list
        classify_map = {
            FileType.TEST: data.tests,
            FileType.CONFIG: data.config_files,
            FileType.MANIFEST: data.manifests,
            FileType.SCRIPT: data.scripts,
            FileType.DOCUMENTATION: data.docs,
            FileType.WORKFLOW: data.workflows,
            FileType.MAKEFILE: data.config_files,
        }

        target_list = classify_map.get(file_info.file_type)
        if target_list is not None:
            target_list.append(file_info)

        # Special handling for dockerfiles
        if file_info.extension == '.dockerfile':
            data.dockerfiles.append(file_info)

        # Mark entrypoints
        if file_info.name in ('main.go', 'main.py', 'app.py', 'server.py',
                            'main.ts', 'index.js', 'main.rs', 'lib.rs'):
            data.entrypoints.append(file_info)

    def _detect_languages(self, files: List[FileInfo]) -> List[Language]:
        """Detect programming languages used in project."""
        language_counts = {}

        for file in files:
            lang = file.language
            if lang != Language.UNKNOWN:
                language_counts[lang] = language_counts.get(lang, 0) + 1

        # Sort by usage count
        sorted_langs = sorted(
            language_counts.items(),
            key=lambda x: x[1],
            reverse=True
        )
        return [lang for lang, _ in sorted_langs]

    def _parse_makefile(self) -> Dict[str, str]:
        """Parse Makefile targets with descriptions."""
        targets = {}

        makefile_path = self.root_path / 'Makefile'
        if not makefile_path.exists():
            return targets

        try:
            with open(makefile_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # Pattern: target: description
            for match in re.finditer(
                r'^([a-zA-Z_][a-zA-Z0-9_-]*):\s*(.*?)$',
                content,
                re.MULTILINE
            ):
                target = match.group(1)
                desc = match.group(2).strip()

                # Use description if meaningful
                if desc and not desc.startswith('#'):
                    targets[target] = desc
                elif not desc:
                    targets[target] = ""

        except (OSError, UnicodeDecodeError):
            pass

        return targets

    def _calculate_technical_debt(self, data: RepoData) -> int:
        """Calculate total technical debt in minutes."""
        total = 0

        for file_info in data.files:
            # Skip test files and documentation
            if file_info.file_type in (FileType.TEST, FileType.DOCUMENTATION):
                continue

            # Skip large generated files
            if file_info.size_bytes > 1_000_000:
                continue

            total += file_info.technical_debt_minutes

        return total


def analyze_repo(root_path: str, **kwargs) -> RepoData:
    """
    Convenience function to analyze a repository.
    
    Args:
        root_path: Path to repository root
        **kwargs: Additional arguments to RepoAnalyzer
        
    Returns:
        RepoData with analysis results
    """
    analyzer = RepoAnalyzer(root_path)
    return analyzer.analyze(**kwargs)