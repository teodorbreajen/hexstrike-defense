#!/usr/bin/env python3
"""
HexStrike Documentation Engine v2.1.0
=====================================

Main entry point for the documentation generator.
This script can be run from the repository root.

Usage:
    python -m docs.tools.doc_engine.run --help
    python docs/tools/doc_engine/run.py --root . --output ../generated
"""

import argparse
import os
import sys
from pathlib import Path

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
        default="../generated",
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

    # Import modules after path setup
    sys.path.insert(0, str(root_path / "docs" / "tools" / "doc_engine"))

    try:
        from analyzer import RepoAnalyzer
        from extractor import MetadataExtractor
        from generator import DocGenerator
        from diagrams import DiagramGenerator
    except ImportError as e:
        print(f"Error importing modules: {e}")
        print("Make sure you're running from the repository root")
        sys.exit(1)

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

    # Step 2.5: Build component registry
    print("Building component registry...")
    try:
        from registry import ComponentRegistry, create_default_registry, Component, ComponentType
        registry = create_default_registry()

        # Register extracted modules
        for path, module in metadata.items():
            if module.types:
                for t in module.types:
                    if t.is_exported:
                        registry.register_component(Component(
                            name=t.name,
                            component_type=ComponentType.SOURCE,
                            path=path,
                            description=f"{t.kind} definition",
                            language=module.language.value,
                            interfaces=t.methods if t.kind == 'interface' else []
                        ))

        print(f"Registered {len(registry.components)} components")
    except ImportError as e:
        print(f"Warning: Could not import registry module: {e}")
        registry = None

    # Step 3: Generate documentation
    print("Generating documentation...")
    generator = DocGenerator(repo_data, metadata, str(output_path))
    sections = generator.generate_all()

    print(f"Generated {len(sections)} documentation sections")

    # Save sections
    generator.save_sections()

    # Step 4: Generate diagrams
    print("Generating diagrams...")
    diagram_gen = DiagramGenerator(repo_data, metadata)
    diagrams = diagram_gen.generate_all()
    print(f"Generated {len(diagrams)} diagrams")

    # Save diagrams
    diagrams_dir = output_path / "diagrams"
    diagrams_dir.mkdir(parents=True, exist_ok=True)
    for diag in diagrams:
        if diag.content:
            diagram_file = diagrams_dir / f"{diag.title.lower().replace(' ', '_')}.md"
            diagram_file.write_text(diag.content, encoding='utf-8')
            print(f"  - {diag.title}: {diagram_file.name}")

    print(f"Documentation saved to: {output_path}")
    print("\nGenerated sections:")
    for section in sections:
        print(f"  - {section.title}: {section.filename}")

    return 0


if __name__ == "__main__":
    sys.exit(main())