#!/bin/bash
#
# Pre-commit hook for documentation generation
# =============================================
# This hook generates documentation before commits.
# 
# INSTALL (Unix/Linux/Mac):
#   ln -s ../../docs/tools/pre-commit.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#
# NOTE: For Windows, use pre-commit.py instead:
#   copy docs/tools/pre-commit.py .git/hooks/pre-commit
#

set -e

echo "Running documentation generation..."

# Change to repo root
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Check if Python is available
if ! command -v python3 &> /dev/null && ! command -v python &> /dev/null; then
    echo "Warning: Python not found, skipping documentation generation"
    exit 0
fi

PYTHON_CMD="python3"
if ! command -v python3 &> /dev/null; then
    PYTHON_CMD="python"
fi

# Check if doc_engine exists
if [ ! -d "docs/tools/doc_engine" ]; then
    echo "Warning: doc_engine not found, skipping documentation generation"
    exit 0
fi

# Generate documentation
echo "Generating documentation..."
$PYTHON_CMD docs/tools/doc_engine/main.py --root "$REPO_ROOT" --output docs/generated

# Check if there were changes
if ! git diff --quiet docs/generated/; then
    echo ""
    echo "Documentation generated. Adding changes..."
    git add docs/generated/
    echo "Documentation changes added to commit."
fi

echo "Documentation generation complete."

exit 0