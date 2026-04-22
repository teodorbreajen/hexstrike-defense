# HexStrike Defense - Documentation System

## Overview

This directory contains the **Documentation Engine** for hexstrike-defense. The engine automatically analyzes the repository and generates comprehensive technical documentation.

## Documentation Structure

```
docs/
├── README.md                    # This file - Documentation system guide
├── index.md                     # Master documentation index
├── generated/                   # Auto-generated documentation
│   ├── index.md
│   ├── project_overview.md
│   ├── high_level_architecture.md
│   ├── component_catalog.md
│   ├── main_flows.md
│   ├── deployment_guide.md
│   ├── config_reference.md
│   ├── security_model.md
│   ├── observability.md
│   ├── dependency_inventory.md
│   ├── interfaces.md
│   ├── repository_map.md
│   ├── assumptions.md
│   ├── limitations.md
│   └── glossary.md
├── tools/                       # Documentation engine tools
│   └── doc_engine/
│       ├── __init__.py
│       ├── analyzer.py         # Repository analysis
│       ├── extractor.py        # Metadata extraction
│       ├── generator.py        # Document generation
│       ├── registry.py         # Component registry
│       └── main.py             # Entry point
├── templates/                   # Reusable templates (future)
└── diagrams/                    # Architecture diagrams (future)
```

## Quick Start

### Generate Documentation

```bash
# Run from repository root
python docs/tools/doc_engine/main.py --root . --output docs/generated

# Or run the runner script directly
cd docs/tools/doc_engine
python main.py --root ../../.. --output ../../generated
```

### View Documentation

Open `docs/generated/index.md` in any Markdown viewer, or use:

- VS Code with Markdown Preview
- GitHub web interface
- MkDocs or similar static site generator

## Documentation Sections

| Section | Content |
|---------|---------|
| [00_index](./00_index/) | Navigation index |
| [01_overview](./01_overview/) | Project vision and purpose |
| [02_scope_and_objectives](./02_scope_and_objectives/) | Goals and boundaries |
| [03_architecture](./03_architecture/) | High-level architecture |
| [04_components](./04_components/) | Component catalog |
| [05_execution_flows](./05_execution_flows/) | Main workflows |
| [06_deployment](./06_deployment/) | Deployment guide |
| [07_configuration](./07_configuration/) | Config reference |
| [08_security_and_constraints](./08_security_and_constraints/) | Security model |
| [09_observability_and_logging](./09_observability_and_logging/) | Logging and metrics |
| [10_dependencies_and_tooling](./10_dependencies_and_tooling/) | Tools and deps |
| [11_interfaces_and_integrations](./11_interfaces_and_integrations/) | APIs and integrations |
| [12_repo_structure](./12_repo_structure/) | File tree |
| [13_decisions_and_assumptions](./13_decisions_and_assumptions/) | Design decisions |
| [14_risks_and_limitations](./14_risks_and_limitations/) | Known limitations |
| [15_maintenance](./15_maintenance/) | Maintenance guide |
| [16_glossary](./16_glossary/) | Technical glossary |

## Documentation Engine Components

### Analyzer (`analyzer.py`)

Scans and analyzes the repository structure:
- File tree detection
- Language detection (Go, YAML, Python, etc.)
- File type classification
- Module identification
- Entry point detection

### Extractor (`extractor.py`)

Extracts detailed metadata from source code:
- Functions and methods
- Types and structs
- Interfaces
- Configuration variables
- HTTP endpoints
- Dependencies

### Generator (`generator.py`)

Generates Markdown documentation:
- Architecture diagrams (Mermaid)
- Component catalogs
- Configuration references
- API documentation
- Security models

### Registry (`registry.py`)

Maintains component registry:
- Component catalog
- Dependency relationships
- Interface definitions

## Regeneration

Documentation is regenerated when:

1. Running the generation script manually:
   ```bash
   python docs/tools/doc_engine/main.py --root . --output docs/generated
   ```

2. As part of CI/CD pipeline (see `.github/workflows/`)

3. After significant code changes

## Adding New Documentation

### Static Content

Add static documentation to the appropriate section folder:
- `docs/01_overview/` for project overview
- `docs/06_deployment/` for deployment guides
- `docs/08_security_and_constraints/` for security docs

### Generated Content

To extend the generator:

1. Edit `generator.py`
2. Add a new method `_generate_<section>()`
3. Return a `DocSection` object
4. Add to `generate_all()` list

### Templates

Place reusable templates in `docs/templates/`:
- Section templates
- Component templates
- API templates

## Limitations

The documentation engine has these limitations:

1. **Cannot extract**:
   - Business logic intent
   - Non-standard patterns
   - Complex type relationships
   - Runtime behavior

2. **Requires validation**:
   - Architecture decisions
   - Security model accuracy
   - Deployment details
   - Integration points

3. **Manual documentation needed**:
   - Use case descriptions
   - Historical context
   - Team-specific conventions
   - Operational procedures

## Maintenance

### Regular Tasks

- Regenerate after major changes
- Validate generated content accuracy
- Update static sections
- Refresh diagrams when architecture changes

### Versioning

- Generated docs are versioned with the code
- Use git to track changes
- Review generated docs in PRs

## Contributing

When contributing to the documentation:

1. Follow the SDD structure
2. Use clear, professional language
3. Include code examples where relevant
4. Validate diagrams render correctly
5. Test regeneration works

## License

Documentation follows project license. See main project LICENSE file.