"""Test Go Extractor"""
import sys
sys.path.insert(0, 'docs/tools/doc_engine')
from go_extractor import GoExtractor

extractor = GoExtractor()
result = extractor.extract('src/mcp-policy-proxy/proxy.go')

print('Go Extractor Test Results')
print('=' * 60)
print(f'Package: {result["package"]}')
print(f'Imports: {len(result["imports"])}')
print(f'Constants: {len(result["consts"])}')
print(f'Types: {len(result["types"])}')
print(f'Functions: {len(result["functions"])}')

print('\n--- Imports (first 5) ---')
for imp in result['imports'][:5]:
    print(f'  - {imp.path} (stdlib: {imp.is_stdlib})')

print('\n--- Types (first 5) ---')
for t in result['types'][:5]:
    print(f'  - {t.name} ({t.kind}) exported={t.is_exported}')

print('\n--- Functions (first 5) ---')
for f in result['functions'][:5]:
    print(f'  - {f.name} exported={f.is_exported} method={f.is_method}')

# Test lakera.go
print('\n--- Testing lakera.go ---')
result2 = extractor.extract('src/mcp-policy-proxy/lakera.go')
print(f'Package: {result2["package"]}')
print(f'Types: {len(result2["types"])}')
print(f'Functions: {len(result2["functions"])}')