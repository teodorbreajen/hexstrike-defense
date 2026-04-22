# Documentation Maintenance

## Overview

This document describes how to maintain and update the documentation system for hexstrike-defense.

## Regeneration Workflow

### When to Regenerate

Regenerate documentation after:

1. **Major Changes**
   - New components added
   - Architecture changes
   - New endpoints or APIs
   - Configuration changes

2. **Before Releases**
   - Version releases
   - Security patches
   - Deployment changes

3. **On Request**
   - Code review feedback
   - Documentation bugs

### How to Regenerate

```bash
# From repository root
python docs/tools/doc_engine/main.py --root . --output docs/generated
```

### Regeneration Options

```bash
# Verbose output
python docs/tools/doc_engine/main.py --root . --output docs/generated --verbose

# Custom output directory
python docs/tools/doc_engine/main.py --root /path/to/repo --output /path/to/docs
```

## Content Types

### Generated Content

These files are auto-generated and should NOT be edited directly:

- `generated/*.md` - Auto-generated sections
- Diagrams in Mermaid format

To modify generated content, edit the generator code in `tools/doc_engine/`.

### Static Content

These files should be edited directly:

- `docs/ARCHITECTURE.md` - Detailed architecture
- `docs/SECURITY.md` - Security hardening
- `docs/OPERATIONS.md` - Operations guide
- Section folders in `docs/XX_*/`

## Best Practices

### Writing Documentation

1. **Be Clear and Concise**
   - Use simple language
   - Avoid jargon or explain it
   - Use active voice

2. **Include Examples**
   - Code snippets
   - Configuration examples
   - Command examples

3. **Use Visual Aids**
   - Diagrams for architecture
   - Tables for comparisons
   - Lists for steps

### Maintaining Accuracy

1. **Review Generated Content**
   - Verify technical accuracy
   - Check for outdated info
   - Validate diagrams

2. **Update Manually**
   - Keep static docs current
   - Add context to generated docs
   - Document exceptions

3. **Track Changes**
   - Use git to track doc changes
   - Review docs in PRs
   - Test regeneration regularly

## Quality Checklist

Before releasing documentation, verify:

- [ ] All generated sections render correctly
- [ ] Links are valid
- [ ] Diagrams display properly
- [ ] Code examples work
- [ ] Configuration references are accurate
- [ ] Security content is reviewed
- [ ] No sensitive information exposed

## Troubleshooting

### Generator Fails

**Problem**: Generator script errors

**Solutions**:
1. Check Python version (3.8+ required)
2. Verify dependencies installed
3. Check file permissions
4. Review error messages

### Missing Content

**Problem**: Generated docs missing information

**Solutions**:
1. Check source files are in expected locations
2. Verify file extensions recognized
3. Update analyzer patterns if needed
4. Add manual content for gaps

### Outdated Content

**Problem**: Generated docs don't reflect code

**Solutions**:
1. Re-run generator
2. Check for new files added
3. Update extraction patterns
4. Add static overrides

## CI/CD Integration

### GitHub Actions

Add to workflow:

```yaml
- name: Generate Documentation
  run: |
    python docs/tools/doc_engine/main.py --root . --output docs/generated
```

### Pre-commit Hook

Add to `.pre-commit-config.yaml`:

```yaml
- repo: local
  hooks:
    - id: generate-docs
      name: Generate Documentation
      entry: python docs/tools/doc_engine/main.py
      language: system
      pass_filenames: false
```

## Future Enhancements

Planned improvements:

1. **Template System**
   - Reusable section templates
   - Component templates
   - API documentation templates

2. **Diagram Generation**
   - Auto-generate architecture diagrams
   - Dependency graphs
   - Flow charts

3. **Validation**
   - Link checking
   - Code example testing
   - Style enforcement

4. **Integration**
   - OpenAPI spec import
   - Kubernetes manifest parsing
   - CI/CD validation

## Support

For documentation issues:

1. Check this guide
2. Review generator source
3. Open an issue
4. Contact the documentation team