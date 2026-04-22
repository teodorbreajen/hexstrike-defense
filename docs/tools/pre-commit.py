#!/usr/bin/env python3
"""
Pre-commit hook for documentation generation
===================================
Cross-platform pre-commit hook that generates documentation before commits.

Install:
  - Unix/Linux/Mac: ln -s ../../docs/tools/pre-commit.py .git/hooks/pre-commit
  - Windows: Copy this file to .git/hooks/pre-commit.py

The hook:
1. Generates documentation using doc_engine
2. Adds any new/changed docs to the commit
3. Fails the commit if documentation generation fails (optional)
"""

import subprocess
import sys
import os
from pathlib import Path


def run_hook(fail_on_error: bool = False) -> int:
    """Run the documentation generation hook.
    
    Args:
        fail_on_error: If True, exit with error code on generation failure
        
    Returns:
        0 on success, 1 on failure
    """
    # Get repo root (parent of .git)
    repo_root = Path(__file__).resolve().parent.parent.parent
    
    print("Running documentation generation...")
    print(f"Repository root: {repo_root}")
    
    # Check if doc_engine exists
    doc_engine_dir = repo_root / "docs" / "tools" / "doc_engine"
    if not doc_engine_dir.exists():
        print("Warning: doc_engine not found, skipping documentation generation")
        return 0
    
    # Find Python
    python_cmd = "python"
    for cmd in ["python3", "python"]:
        try:
            subprocess.run([cmd, "--version"], capture_output=True, check=True)
            python_cmd = cmd
            break
        except (subprocess.CalledProcessError, FileNotFoundError):
            continue
    
    # Generate documentation
    print("Generating documentation...")
    main_script = doc_engine_dir / "main.py"
    
    try:
        result = subprocess.run(
            [python_cmd, str(main_script), "--root", str(repo_root), "--output", "docs/generated"],
            capture_output=True,
            text=True,
            cwd=repo_root
        )
        
        if result.returncode != 0:
            print(f"Error generating documentation: {result.stderr}")
            if fail_on_error:
                return 1
            return 0
        
        print(result.stdout)
        
        # Check for changes
        docs_dir = repo_root / "docs" / "generated"
        if docs_dir.exists():
            try:
                diff_result = subprocess.run(
                    ["git", "diff", "--quiet", "docs/generated/"],
                    capture_output=True,
                    cwd=repo_root
                )
                
                if diff_result.returncode != 0:  # There are changes
                    print("")
                    print("Documentation generated. Adding changes...")
                    subprocess.run(
                        ["git", "add", "docs/generated/"],
                        cwd=repo_root
                    )
                    print("Documentation changes added to commit.")
                    
            except FileNotFoundError:
                print("Warning: git not found, skipping git add")
        
    except Exception as e:
        print(f"Error: {e}")
        if fail_on_error:
            return 1
    
    print("Documentation generation complete.")
    return 0


if __name__ == "__main__":
    # Parse arguments
    fail_on_error = "--fail" in sys.argv or "-f" in sys.argv
    
    exit_code = run_hook(fail_on_error=fail_on_error)
    sys.exit(exit_code)