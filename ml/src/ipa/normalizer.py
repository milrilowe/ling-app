"""
IPA normalization utilities.

Normalizes IPA strings to canonical form for consistent comparison
between different sources (Whisper, gruut, etc.).
"""

import unicodedata
from typing import List


class IPANormalizer:
    """
    Normalizes IPA strings to a canonical form for comparison.

    Handles differences between gruut and Whisper output:
    - Tie bars in affricates (d͡ʒ → dʒ)
    - Unicode variants (ɡ U+0261 → g U+0067)
    - Prosodic markers (removes ‖ |)
    - R-coloring variants (ɚ → əɹ)
    """

    # Tie bar character used in affricates
    TIE_BAR = '\u0361'  # ͡

    # Prosodic/boundary markers to remove
    PROSODIC_MARKERS = {'‖', '|', '‿'}

    # Unicode character normalization mappings
    # Maps variant forms to canonical forms
    CHAR_MAPPINGS = {
        'ɡ': 'g',      # U+0261 → U+0067 (IPA g → ASCII g)
        'ː': ':',      # U+02D0 → U+003A (IPA length → colon) - optional
        ''': "'",      # Curly apostrophe → straight
        ''': "'",
    }

    # R-colored vowel expansions (canonical form uses separate ɹ)
    R_COLORED_VOWELS = {
        'ɚ': 'əɹ',     # Schwa + r
        'ɝ': 'ɜɹ',     # Open-mid central + r
    }

    # Characters that are NOT phonemes (modifiers, diacritics, etc.)
    # These should stay attached to the preceding phoneme
    COMBINING_MARKS = set(
        '\u0300\u0301\u0302\u0303\u0304\u0305\u0306\u0307'  # Combining accents
        '\u0308\u0309\u030a\u030b\u030c\u030d\u030e\u030f'
        '\u0310\u0311\u0312\u0313\u0314\u0315\u0316\u0317'
        '\u0318\u0319\u031a\u031b\u031c\u031d\u031e\u031f'
        '\u0320\u0321\u0322\u0323\u0324\u0325\u0326\u0327'
        '\u0328\u0329\u032a\u032b\u032c\u032d\u032e\u032f'
        '\u0330\u0331\u0332\u0333\u0334\u0335\u0336\u0337'
        '\u0338\u0339\u033a\u033b\u033c\u033d\u033e\u033f'
        '\u0340\u0341\u0342\u0343\u0344\u0345\u0346\u0347'
        '\u0348\u0349\u034a\u034b\u034c\u034d\u034e\u034f'
        '\u0350\u0351\u0352\u0353\u0354\u0355\u0356\u0357'
        '\u0358\u0359\u035a\u035b\u035c\u035d\u035e\u035f'
        '\u0360\u0361\u0362'  # Includes tie bar
    )

    # IPA modifier letters that modify the preceding sound
    MODIFIER_LETTERS = {
        'ː',   # Length mark
        'ˑ',   # Half-length
        'ʰ',   # Aspiration
        'ʷ',   # Labialization
        'ʲ',   # Palatalization
        'ˠ',   # Velarization
        'ˤ',   # Pharyngealization
        'ⁿ',   # Nasal release
        'ˡ',   # Lateral release
    }

    # Stress markers - these are kept but treated as separate "phonemes"
    # for alignment purposes
    STRESS_MARKERS = {'ˈ', 'ˌ'}

    def __init__(self, keep_stress: bool = True, keep_length: bool = True):
        """
        Initialize IPA normalizer.

        Args:
            keep_stress: If True, preserve stress markers (ˈ ˌ)
            keep_length: If True, preserve length markers (ː)
        """
        self.keep_stress = keep_stress
        self.keep_length = keep_length

    def normalize(self, ipa_string: str) -> str:
        """
        Normalize IPA string to canonical form.

        Args:
            ipa_string: Raw IPA string from any source

        Returns:
            Normalized IPA string
        """
        if not ipa_string:
            return ""

        # Step 1: Unicode NFC normalization
        result = unicodedata.normalize('NFC', ipa_string)

        # Step 2: Remove tie bars (d͡ʒ → dʒ)
        result = result.replace(self.TIE_BAR, '')

        # Step 3: Remove prosodic markers
        for marker in self.PROSODIC_MARKERS:
            result = result.replace(marker, '')

        # Step 4: Apply character mappings
        for old, new in self.CHAR_MAPPINGS.items():
            result = result.replace(old, new)

        # Step 5: Expand r-colored vowels
        for old, new in self.R_COLORED_VOWELS.items():
            result = result.replace(old, new)

        # Step 6: Optionally remove stress markers
        if not self.keep_stress:
            for marker in self.STRESS_MARKERS:
                result = result.replace(marker, '')

        # Step 7: Optionally remove length markers
        if not self.keep_length:
            result = result.replace('ː', '').replace(':', '')

        # Step 8: Normalize whitespace
        result = ' '.join(result.split())

        return result

    def extract_phonemes(self, ipa_string: str) -> List[str]:
        """
        Extract individual phonemes from normalized IPA string.

        Each phoneme includes its combining marks and modifiers.
        Stress markers are treated as separate tokens.

        Args:
            ipa_string: IPA string (will be normalized first)

        Returns:
            List of phoneme strings
        """
        normalized = self.normalize(ipa_string)
        if not normalized:
            return []

        phonemes = []
        current_phoneme = ""

        for char in normalized:
            # Skip whitespace
            if char.isspace():
                if current_phoneme:
                    phonemes.append(current_phoneme)
                    current_phoneme = ""
                continue

            # Check if this is a combining mark or modifier
            is_combining = (
                unicodedata.combining(char) != 0 or
                char in self.COMBINING_MARKS or
                char in self.MODIFIER_LETTERS
            )

            if is_combining:
                # Attach to current phoneme
                current_phoneme += char
            elif char in self.STRESS_MARKERS:
                # Stress markers: save current phoneme, add stress as its own token
                if current_phoneme:
                    phonemes.append(current_phoneme)
                    current_phoneme = ""
                # Attach stress to the NEXT phoneme by starting a new one with it
                current_phoneme = char
            else:
                # New base character
                if current_phoneme and current_phoneme not in self.STRESS_MARKERS:
                    # Save previous phoneme (unless it's just a stress marker waiting for its vowel)
                    phonemes.append(current_phoneme)
                    current_phoneme = char
                else:
                    # Either empty or stress marker - add this char to it
                    current_phoneme += char

        # Don't forget the last phoneme
        if current_phoneme:
            phonemes.append(current_phoneme)

        return phonemes


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


def extract_phonemes(ipa_string: str) -> List[str]:
    """
    Convenience function to extract phonemes from IPA string.

    Args:
        ipa_string: IPA string

    Returns:
        List of phoneme strings
    """
    normalizer = IPANormalizer()
    return normalizer.extract_phonemes(ipa_string)
