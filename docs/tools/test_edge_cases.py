"""Edge Case Tests"""
import sys
import os

# Add correct path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'doc_engine'))

# Test edge cases
print('Edge Case Tests')
print('=' * 60)

# 1. Test with empty/non-existent directory
from analyzer import RepoAnalyzer
try:
    a = RepoAnalyzer('/nonexistent/path')
    r = a.analyze()
    print('[FAIL] Should have raised error for non-existent path')
except Exception as e:
    print(f'[OK] Non-existent path handled: {type(e).__name__}')

# 2. Test with missing go.mod
try:
    a = RepoAnalyzer('src/mcp-policy-proxy')
    r = a.analyze()
    print(f'[OK] go.mod not required, found {len(r.files)} files')
except Exception as e:
    print(f'[FAIL] {e}')

# 3. Test metadata extraction with binary files
from extractor import MetadataExtractor
extractor = MetadataExtractor('.')
print('\nHandling special files:')

# Test binary file handling
binary_file = 'src/mcp-policy-proxy/mcp-policy-proxy'
result = extractor._extract_go_file(binary_file)
if result is None:
    print(f'  [OK] Binary file correctly handled (returned None)')
else:
    print(f'  [WARN] Binary file returned result: {result}')

# Test non-go file
non_go = 'README.md'
result2 = extractor._extract_go_file(non_go)
if result2 is None:
    print(f'  [OK] Non-Go file correctly handled (returned None)')

print('\nAll edge case tests completed!')