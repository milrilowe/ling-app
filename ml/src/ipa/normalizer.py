"""
IPA normalization utilities.

Handles differences in IPA format between different tools and versions.
"""

import unicodedata
from typing import Dict, List, Tuple


class IPANormalizer:
    """
    Normalizes IPA strings for consistent comparison.

    Handles known differences between gruut versions and Whisper output,
    including Unicode variations, character substitutions, and formatting.
    """

    # Known IPA character substitutions
    # Some tools use different Unicode characters for the same phoneme
    CHAR_SUBSTITUTIONS = {
        '\u0261': 'g',  # ɡ (U+0261) -> g (U+0067) - both valid for voiced velar stop
        'g': '\u0261',  # Also add reverse mapping
    }

    def __init__(self):
        """Initialize IPA normalizer."""
        pass

    def normalize(self, ipa_string: str) -> str:
        """
        Apply normalization to IPA string.

        Normalization steps:
        1. Unicode normalization (NFC form)
        2. Whitespace normalization
        3. Character substitutions (if needed)

        Args:
            ipa_string: IPA string to normalize

        Returns:
            Normalized IPA string
        """
        if not ipa_string:
            return ""

        # Step 1: Unicode normalization (NFC - Canonical Composition)
        # This ensures combining diacritics are in standard form
        normalized = unicodedata.normalize('NFC', ipa_string)

        # Step 2: Whitespace normalization
        # Remove extra spaces, normalize to single spaces
        normalized = ' '.join(normalized.split())

        return normalized

    def normalize_for_comparison(self, ipa_string: str) -> str:
        """
        Normalize IPA string specifically for comparison.

        More aggressive than regular normalization:
        - Removes all whitespace
        - Converts to lowercase (for stress markers that might vary)

        Args:
            ipa_string: IPA string to normalize

        Returns:
            Normalized IPA string for comparison
        """
        normalized = self.normalize(ipa_string)

        # Remove all whitespace for character-by-character comparison
        normalized = normalized.replace(' ', '')

        return normalized

    def compare(
        self,
        ipa1: str,
        ipa2: str,
        ignore_whitespace: bool = False
    ) -> Dict:
        """
        Compare two IPA strings.

        Args:
            ipa1: First IPA string
            ipa2: Second IPA string
            ignore_whitespace: If True, ignores whitespace in comparison

        Returns:
            Dictionary containing:
                - 'normalized_1': Normalized version of ipa1
                - 'normalized_2': Normalized version of ipa2
                - 'exact_match': Whether strings match exactly
                - 'length_1': Length of first string
                - 'length_2': Length of second string
                - 'differences': List of (index, char1, char2) for mismatches
        """
        if ignore_whitespace:
            norm1 = self.normalize_for_comparison(ipa1)
            norm2 = self.normalize_for_comparison(ipa2)
        else:
            norm1 = self.normalize(ipa1)
            norm2 = self.normalize(ipa2)

        # Check exact match
        exact_match = norm1 == norm2

        # Find character-level differences
        differences = []
        max_len = max(len(norm1), len(norm2))

        for i in range(max_len):
            char1 = norm1[i] if i < len(norm1) else None
            char2 = norm2[i] if i < len(norm2) else None

            if char1 != char2:
                differences.append((i, char1, char2))

        return {
            'normalized_1': norm1,
            'normalized_2': norm2,
            'exact_match': exact_match,
            'length_1': len(norm1),
            'length_2': len(norm2),
            'differences': differences
        }

    def get_alignment(self, ipa1: str, ipa2: str) -> Tuple[str, str, str]:
        """
        Create visual alignment of two IPA strings.

        Args:
            ipa1: First IPA string
            ipa2: Second IPA string

        Returns:
            Tuple of (aligned_ipa1, match_string, aligned_ipa2)
            where match_string shows '|' for matches and ' ' for mismatches
        """
        norm1 = self.normalize(ipa1)
        norm2 = self.normalize(ipa2)

        max_len = max(len(norm1), len(norm2))

        aligned1 = norm1.ljust(max_len)
        aligned2 = norm2.ljust(max_len)

        match_string = ''
        for c1, c2 in zip(aligned1, aligned2):
            if c1 == c2:
                match_string += '|'
            else:
                match_string += ' '

        return aligned1, match_string, aligned2


def normalize_ipa(ipa_string: str) -> str:
    """
    Convenience function to normalize an IPA string.

    Args:
        ipa_string: IPA string to normalize

    Returns:
        Normalized IPA string
    """
    normalizer = IPANormalizer()
    return normalizer.normalize(ipa_string)


def compare_ipa(ipa1: str, ipa2: str) -> Dict:
    """
    Convenience function to compare two IPA strings.

    Args:
        ipa1: First IPA string
        ipa2: Second IPA string

    Returns:
        Comparison dictionary
    """
    normalizer = IPANormalizer()
    return normalizer.compare(ipa1, ipa2, ignore_whitespace=True)
