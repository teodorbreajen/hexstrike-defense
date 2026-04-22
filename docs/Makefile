# =============================================================================
# HexStrike Defense - Documentation Makefile
# =============================================================================
# Targets for documentation generation, validation, and maintenance
#
# Usage:
#   make docs              - Generate all documentation
#   make docs-validate    - Validate documentation links
#   make docs-check       - Check documentation structure
#   make docs-clean       - Clean generated docs
#   make docs-diagrams    - Generate diagrams only
#   make docs-full        - Full docs: generate + validate + check
# =============================================================================

# Configuration
PYTHON := python3
PYTHON_BIN := $(shell command -v python3 2> /dev/null || command -v python 2> /dev/null || echo "python")
DOCS_DIR := docs
GENERATED_DIR := $(DOCS_DIR)/generated
DOC_ENGINE := $(DOCS_DIR)/tools/doc_engine

# Colors
GREEN := $(shell tput setaf 2 2>/dev/null || echo "")
YELLOW := $(shell tput setaf 3 2>/dev/null || echo "")
RED := $(shell tput setaf 1 2>/dev/null || echo "")
RESET := $(shell tput sgr0 2>/dev/null || echo "")

# Default target
.PHONY: help
help:
	@echo "$(GREEN)HexStrike Defense - Documentation$(RESET)"
	@echo "================================"
	@echo ""
	@echo "$(YELLOW)Documentation Targets:$(RESET)"
	@echo "  make docs            - Generate documentation"
	@echo "  make docs-validate  - Validate internal links"
	@echo "  make docs-check     - Check documentation structure"
	@echo "  make docs-diagrams  - Generate diagrams only"
	@echo "  make docs-full      - Full docs: generate + validate + check"
	@echo "  make docs-clean     - Clean generated docs"
	@echo "  make docs-serve     - Serve docs locally (if mkdocs available)"
	@echo ""
	@echo "$(YELLOW)Info Targets:$(RESET)"
	@echo "  make docs-stats     - Show documentation statistics"
	@echo "  make docs-list      - List generated files"
	@echo ""

# =============================================================================
# Main Documentation Targets
# =============================================================================

# Generate documentation
.PHONY: docs
docs:
	@echo "$(GREEN)Generating documentation...$(RESET)"
	$(PYTHON_BIN) $(DOC_ENGINE)/main.py --root . --output $(GENERATED_DIR)
	@echo ""
	@echo "$(GREEN)✓ Documentation generated successfully$(RESET)"
	@echo "  Output: $(GENERATED_DIR)/"

# Validate documentation links
.PHONY: docs-validate
docs-validate:
	@echo "$(GREEN)Validating documentation links...$(RESET)"
	@$(PYTHON_BIN) -c "
import sys
sys.path.insert(0, '$(DOC_ENGINE)')
from validator import validate_links
success, issues = validate_links('$(GENERATED_DIR)')
if issues:
    print('Found {} issues:'.format(len(issues)))
    for i in issues[:10]:
        print('  {}: {} - {}'.format(i.severity, i.file, i.link))
    if len(issues) > 10:
        print('  ... and {} more'.format(len(issues) - 10))
    sys.exit(0 if success else 1)
else:
    print('✓ All links valid!')
" || true

# Check documentation structure
.PHONY: docs-check
docs-check:
	@echo "$(GREEN)Checking documentation structure...$(RESET)"
	@test -d $(GENERATED_DIR) && echo "$(GREEN)✓$(RESET) Generated dir exists" || (echo "$(RED)✗$(RESET) Generated dir missing" && exit 1)
	@test -f $(GENERATED_DIR)/index.md && echo "$(GREEN)✓$(RESET) Index exists" || (echo "$(RED)✗$(RESET) Index missing" && exit 1)
	@test -f $(GENERATED_DIR)/high_level_architecture.md && echo "$(GREEN)✓$(RESET) Architecture doc exists" || echo "$(YELLOW)!$(RESET) Architecture doc missing"
	@test -f $(GENERATED_DIR)/component_catalog.md && echo "$(GREEN)✓$(RESET) Components doc exists" || echo "$(YELLOW)!$(RESET) Components doc missing"
	@test -f $(GENERATED_DIR)/security_model.md && echo "$(GREEN)✓$(RESET) Security doc exists" || echo "$(YELLOW)!$(RESET) Security doc missing"
	@echo ""
	@echo "$(GREEN)Structure check complete!$(RESET)"

# Generate diagrams only
.PHONY: docs-diagrams
docs-diagrams:
	@echo "$(GREEN)Generating diagrams...$(RESET)"
	@$(PYTHON_BIN) -c "
import sys
sys.path.insert(0, '$(DOC_ENGINE)')
from analyzer import RepoAnalyzer
from diagrams import generate_diagrams
from extractor import MetadataExtractor

analyzer = RepoAnalyzer('.')
data = analyzer.analyze()
extractor = MetadataExtractor('.')
metadata = extractor.extract_all(data.files)

diagrams = generate_diagrams(data, metadata)
for d in diagrams:
    print('  Generated: {}'.format(d.filename))
print('✓ Diagrams generated')
"

# Full documentation workflow
.PHONY: docs-full
docs-full: docs docs-validate docs-check
	@echo ""
	@echo "$(GREEN)✓ Full documentation workflow complete!$(RESET)"

# Clean generated documentation
.PHONY: docs-clean
docs-clean:
	@echo "$(GREEN)Cleaning generated documentation...$(RESET)"
	@rm -f $(GENERATED_DIR)/*.md
	@echo "$(GREEN)✓ Documentation cleaned$(RESET)"

# =============================================================================
# Info Targets
# =============================================================================

# Show documentation statistics
.PHONY: docs-stats
docs-stats:
	@echo "$(GREEN)Documentation Statistics$(RESET)"
	@echo "================================"
	@echo ""
	@echo "Generated files:"
	@ls -la $(GENERATED_DIR)/*.md 2>/dev/null | wc -l | xargs echo "  Total markdown files:"
	@echo ""
	@echo "File sizes:"
	@du -sh $(GENERATED_DIR) 2>/dev/null | xargs echo "  Total size:"
	@echo ""
	@echo "Sections:"
	@ls $(GENERATED_DIR)/*.md 2>/dev/null | xargs -I{} basename {} .md | sed 's/^/  - /'

# List generated files
.PHONY: docs-list
docs-list:
	@echo "$(GREEN)Generated Documentation Files$(RESET)"
	@echo "================================"
	@ls -1 $(GENERATED_DIR)/*.md 2>/dev/null || echo "No files found"

# Serve documentation locally
.PHONY: docs-serve
docs-serve:
	@if command -v mkdocs &> /dev/null; then \
		mkdocs serve; \
	else \
		echo "$(YELLOW)Warning: mkdocs not installed. Installing...$(RESET)"; \
		pip install mkdocs; \
		mkdocs serve; \
	fi

# =============================================================================
# Git Integration
# =============================================================================

# Add docs to git
.PHONY: docs-git-add
docs-git-add:
	@echo "$(GREEN)Adding generated docs to git...$(RESET)"
	@git add $(GENERATED_DIR)/
	@echo "$(GREEN)✓ Documentation files staged$(RESET)"

# Check docs git status
.PHONY: docs-git-status
docs-git-status:
	@echo "$(GREEN)Documentation git status$(RESET)"
	@git status --short $(GENERATED_DIR)/

# =============================================================================
# CI/CD Helpers
# =============================================================================

# Check if docs need regeneration
.PHONY: docs-changed
docs-changed:
	@git diff --quiet $(GENERATED_DIR)/ && echo "unchanged" || echo "changed"

# Export for CI
.PHONY: docs-ci
docs-ci: docs docs-validate docs-check
	@echo ""
	@echo "$(GREEN)CI documentation check complete!$(RESET)"
