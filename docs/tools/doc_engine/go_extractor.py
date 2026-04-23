"""
Go Documentation Generator - ADDON
==================================

NOTE: Main Go extraction is now handled by extractor.py.
This module is kept for backward compatibility and utility functions only.

This module provides specialized Go utility functions that complement
the main extractor.py. The core extraction logic is in MetadataExtractor
which handles Go, Python, TypeScript, and Rust extraction uniformly.

Author: HexStrike Documentation Team
Version: 2.1.0 (addon only)
"""

# Re-export from extractor for backward compatibility
# The actual extraction happens in extractor.py

__all__ = ['extract_go_documentation']

# Import the main extractor functions for compatibility
def extract_go_documentation(file_path: str):
    """
    DEPRECATED: Use extractor.MetadataExtractor instead.
    
    This function is kept for backward compatibility only.
    """
    import warnings
    warnings.warn(
        "extract_go_documentation is deprecated. Use "
        "extractor.MetadataExtractor instead.",
        DeprecationWarning,
        stacklevel=2
    )
    return {}