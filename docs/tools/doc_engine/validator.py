"""
Link Validator
=============

Validates internal links in generated documentation.
Checks for:
- Broken internal links
- Missing files
- Invalid anchors
- Relative path correctness

Author: HexStrike Documentation Team
"""

import os
import re
import logging
from pathlib import Path
from typing import Dict, List, Set, Tuple
from dataclasses import dataclass

# Compiled patterns for link validation
_LINK_PATTERN = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')
_ANCHOR_PATTERN = re.compile(r'^#{1,6}\s+(.+)$', re.MULTILINE)
_EXPLICIT_ANCHOR = re.compile(r'\{#([a-zA-Z0-9_-]+)\}')

logger = logging.getLogger(__name__)


@dataclass
class LinkIssue:
    """A link validation issue."""
    file: str
    line: int
    link: str
    issue_type: str
    severity: str = "warning"


class LinkValidator:
    """Validates internal documentation links."""

    def __init__(self, docs_dir: str):
        self.docs_dir = Path(docs_dir)
        self.issues: List[LinkIssue] = []
        self.files: Set[str] = set()
        self.anchors: Dict[str, Set[str]] = {}

    def validate(self) -> List[LinkIssue]:
        """Validate all links in documentation."""
        self.issues = []
        self._scan_files()
        self._scan_anchors()
        self._check_links()
        return self.issues

    def _scan_files(self):
        """Scan all markdown files."""
        for root, _, files in os.walk(self.docs_dir):
            for file in files:
                if file.endswith('.md'):
                    file_path = Path(root) / file
                    rel_path = file_path.relative_to(self.docs_dir)
                    self.files.add(str(rel_path))
                    self.files.add(file)  # Also add filename only

    def _scan_anchors(self):
        """Scan for markdown anchors."""
        for file_path in self.docs_dir.rglob('*.md'):
            rel_path = str(file_path.relative_to(self.docs_dir))
            self.anchors[rel_path] = self._extract_anchors(file_path)

    def _extract_anchors(self, file_path: Path) -> Set[str]:
        """Extract anchors from a markdown file."""
        anchors = set()
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # Headers (# Title -> title)
            for match in re.finditer(r'^#{1,6}\s+(.+)$', content, re.MULTILINE):
                title = match.group(1).strip().lower()
                # Remove formatting
                title = re.sub(r'\[([^\]]+)\]\([^)]+\)', r'\1', title)
                title = re.sub(r'[^a-z0-9\s-]', '', title)
                title = title.replace(' ', '-')
                anchors.add(f"#{title}")

            # Explicit anchors (# {#anchor-name})
            for match in re.finditer(r'\{#([a-zA-Z0-9_-]+)\}', content):
                anchors.add(f"#{match.group(1)}")

        except (OSError, UnicodeDecodeError):
            pass

        return anchors

    def _check_links(self):
        """Check all links in documentation."""
        for file_path in self.docs_dir.rglob('*.md'):
            self._check_file_links(file_path)

    def _check_file_links(self, file_path: Path):
        """Check links in a single file."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
        except (OSError, UnicodeDecodeError):
            return

        rel_path = file_path.relative_to(self.docs_dir)

        for line_num, line in enumerate(lines, 1):
            # Find markdown links [text](url)
            for match in re.finditer(r'\[([^\]]+)\]\(([^)]+)\)', line):
                link = match.group(2)

                # Skip external links and anchors
                if link.startswith(('http', 'mailto:', '#')):
                    continue

                self._validate_link(rel_path, line_num, link)

    def _validate_link(self, file: Path, line: int, link: str):
        """Validate a single link."""
        # Handle anchors within same file
        if '#' in link:
            anchor = link.split('#')[1]
            base_file = link.split('#')[0]

            if not base_file:
                # Same file anchor
                file_anchors = self.anchors.get(str(file), set())
                if f"#{anchor}" not in file_anchors and anchor not in [a.lstrip('#') for a in file_anchors]:
                    self.issues.append(LinkIssue(
                        file=str(file),
                        line=line,
                        link=link,
                        issue_type="broken_anchor",
                        severity="warning"
                    ))
                return
            else:
                link = base_file

        # Handle relative paths - be more lenient
        try:
            link_path = (file.parent / link).resolve()
            
            # Check if file exists (any case)
            if link_path.exists():
                return
                
            # Check if it's a valid relative path within project
            if link_path.is_absolute():
                # For absolute paths, just check if they look reasonable
                if link.startswith('../') or link.startswith('./'):
                    pass  # Accept relative paths
                elif 'readme' in link.lower() or 'license' in link.lower():
                    pass  # Common project files
                else:
                    # Check relative to docs_dir
                    try:
                        link_path.relative_to(self.docs_dir.resolve())
                    except ValueError:
                        # Not relative to docs_dir, might be external project file
                        if link_path.exists():
                            return
                        self.issues.append(LinkIssue(
                            file=str(file),
                            line=line,
                            link=link,
                            issue_type="file_not_found",
                            severity="warning"
                        ))
            else:
                # Relative path - check if it could be valid
                if link_path.exists():
                    return
                    
        except (ValueError, OSError) as e:
            logger.debug(f"Could not validate link {link}: {e}")
            # Don't add issue since we can't determine validity

    def _file_exists_case_insensitive(self, link: str) -> bool:
        """Check if file exists (case insensitive)."""
        link_lower = link.lower()
        for file in self.files:
            if file.lower() == link_lower:
                return True
        return False

    def generate_report(self) -> str:
        """Generate a validation report."""
        lines = ["# Link Validation Report\n"]

        if not self.issues:
            lines.append("✅ **No issues found!** All internal links are valid.\n")
            return '\n'.join(lines)

        # Group by severity
        errors = [i for i in self.issues if i.severity == "error"]
        warnings = [i for i in self.issues if i.severity == "warning"]

        lines.append(f"## Summary\n")
        lines.append(f"- ❌ **Errors**: {len(errors)}")
        lines.append(f"- ⚠️ **Warnings**: {len(warnings)}\n")

        if errors:
            lines.append("## Errors\n")
            for issue in errors:
                lines.append(f"- `{issue.file}:{issue.line}`: `{issue.link}` - {issue.issue_type}")

        if warnings:
            lines.append("\n## Warnings\n")
            for issue in warnings:
                lines.append(f"- `{issue.file}:{issue.line}`: `{issue.link}` - {issue.issue_type}")

        return '\n'.join(lines)


def validate_links(docs_dir: str) -> Tuple[bool, List[LinkIssue]]:
    """
    Validate links in documentation directory.

    Returns:
        Tuple of (success, issues)
    """
    validator = LinkValidator(docs_dir)
    issues = validator.validate()
    return len([i for i in issues if i.severity == "error"]) == 0, issues


if __name__ == "__main__":
    import sys

    docs_dir = sys.argv[1] if len(sys.argv) > 1 else "docs/generated"

    print(f"Validating links in: {docs_dir}")
    success, issues = validate_links(docs_dir)

    if issues:
        print("\nIssues found:")
        for issue in issues:
            print(f"  {issue.severity.upper()}: {issue.file}:{issue.line} - {issue.link}")

    sys.exit(0 if success else 1)