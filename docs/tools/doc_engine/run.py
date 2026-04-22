#!/usr/bin/env python3
"""
Documentation Engine Runner
==========================

Executes the documentation engine to generate project documentation.
Can be run standalone or imported as a module.

Usage:
    python run.py [--root /path/to/repo] [--output /path/to/docs]

Author: HexStrike Documentation Team
"""

import argparse
import os
import sys
from pathlib import Path

# Add current directory to path for local imports
script_dir = Path(__file__).parent
sys.path.insert(0, str(script_dir))

# Import local modules directly
from analyzer import RepoAnalyzer
from extractor import MetadataExtractor
from generator import DocGenerator


def main():
    parser = argparse.ArgumentParser(
        description="HexStrike Documentation Engine"
    )
    parser.add_argument(
        "--root",
        default=".",
        help="Path to repository root (default: current directory)"
    )
    parser.add_argument(
        "--output",
        default="../../generated",
        help="Output directory for generated docs"
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose output"
    )

    args = parser.parse_args()

    # Resolve paths
    root_path = Path(args.root).resolve()
    output_path = (root_path / args.output).resolve()

    if args.verbose:
        print(f"Repository root: {root_path}")
        print(f"Output directory: {output_path}")

    # Step 1: Analyze repository
    print("Analyzing repository structure...")
    analyzer = RepoAnalyzer(str(root_path))
    repo_data = analyzer.analyze()

    print(f"Found {len(repo_data.files)} files")
    print(f"Languages: {[lang.value for lang in repo_data.languages]}")

    # Step 2: Extract metadata
    print("Extracting metadata...")
    extractor = MetadataExtractor(str(root_path))
    metadata = extractor.extract_all(repo_data.files)

    print(f"Analyzed {len(metadata)} modules")

    # Step 3: Generate documentation
    print("Generating documentation...")
    generator = DocGenerator(repo_data, metadata, str(output_path))
    sections = generator.generate_all()

    print(f"Generated {len(sections)} documentation sections")

    # Save sections
    generator.save_sections()

    print(f"Documentation saved to: {output_path}")
    print("\nGenerated sections:")
    for section in sections:
        print(f"  - {section.title}: {section.filename}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
