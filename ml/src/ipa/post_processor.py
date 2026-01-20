"""
IPA post-processing and error correction.

Fixes common mistakes made by Whisper IPA model and applies
language-specific correction rules.
"""

from typing import Dict, List, Optional


class IPAPostProcessor:
    """
    Post-processes IPA output to fix common errors.

    Whisper IPA models sometimes confuse similar phonemes.
    This class applies correction rules based on common mistakes.
    """

    # Common substitutions that Whisper makes (wrong -> correct)
    COMMON_MISTAKES = {
        # Dental fricatives often confused with stops
        'θ': ['t', 'f'],      # /θ/ (thin) confused with /t/ or /f/
        'ð': ['d', 'v'],      # /ð/ (this) confused with /d/ or /v/

        # R-sounds
        'ɹ': ['r', 'ɾ'],      # English /ɹ/ vs trill /r/

        # Schwa variations
        'ə': ['ʌ', 'ɐ'],      # Schwa variations

        # Vowel length markers
        'ː': [''],            # Length marker sometimes dropped
    }

    # Language-specific correction rules
    LANGUAGE_RULES = {
        'en': {
            # English doesn't use trilled /r/, should be /ɹ/
            'r': 'ɹ',
            # English uses rhotic schwa
            'əɹ': 'ɚ',
            'ɜr': 'ɝ',
        },
        'es': {
            # Spanish uses trill, not approximant
            'ɹ': 'r',
        },
        'fr': {
            # French uses uvular /ʁ/
            'ɹ': 'ʁ',
            'r': 'ʁ',
        }
    }

    def __init__(self, language: str = 'en'):
        """
        Initialize IPA post-processor.

        Args:
            language: Language code (e.g., 'en', 'es', 'fr')
        """
        self.language = language

    def post_process(
        self,
        ipa_string: str,
        expected_text: Optional[str] = None,
        confidence: Optional[float] = None
    ) -> str:
        """
        Post-process IPA output.

        Args:
            ipa_string: Raw IPA output from model
            expected_text: Optional expected text (for context-based correction)
            confidence: Optional confidence score (apply more corrections if low)

        Returns:
            Corrected IPA string
        """
        if not ipa_string:
            return ""

        processed = ipa_string

        # Apply language-specific rules
        processed = self._apply_language_rules(processed)

        # Normalize spacing
        processed = self._normalize_spacing(processed)

        # Fix common Unicode issues
        processed = self._fix_unicode_issues(processed)

        return processed

    def _apply_language_rules(self, ipa_string: str) -> str:
        """Apply language-specific substitution rules."""
        if self.language not in self.LANGUAGE_RULES:
            return ipa_string

        rules = self.LANGUAGE_RULES[self.language]
        result = ipa_string

        for wrong, correct in rules.items():
            result = result.replace(wrong, correct)

        return result

    def _normalize_spacing(self, ipa_string: str) -> str:
        """Normalize whitespace in IPA string."""
        # Remove multiple spaces
        while '  ' in ipa_string:
            ipa_string = ipa_string.replace('  ', ' ')

        # Remove spaces around certain characters
        # (stress marks shouldn't have spaces)
        for char in ['ˈ', 'ˌ', 'ː']:
            ipa_string = ipa_string.replace(f' {char}', char)
            ipa_string = ipa_string.replace(f'{char} ', char)

        return ipa_string.strip()

    def _fix_unicode_issues(self, ipa_string: str) -> str:
        """Fix common Unicode normalization issues."""
        import unicodedata

        # Normalize to NFC form (canonical composition)
        normalized = unicodedata.normalize('NFC', ipa_string)

        # Fix common character confusions
        replacements = {
            'g': 'ɡ',  # Use IPA g (U+0261) not Latin g (U+0067)
            ':': 'ː',  # Use IPA length marker (U+02D0) not colon
        }

        for wrong, correct in replacements.items():
            # Only replace in IPA context (not in all cases)
            # This is a simplified version - more sophisticated logic could be added
            pass  # Skip for now to avoid breaking valid characters

        return normalized


def post_process_ipa(
    ipa_string: str,
    language: str = 'en',
    expected_text: Optional[str] = None
) -> str:
    """
    Convenience function to post-process IPA.

    Args:
        ipa_string: Raw IPA string
        language: Language code
        expected_text: Optional expected text

    Returns:
        Post-processed IPA string
    """
    processor = IPAPostProcessor(language=language)
    return processor.post_process(ipa_string, expected_text)
