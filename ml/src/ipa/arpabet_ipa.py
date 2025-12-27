"""
ARPAbet to IPA conversion for MFA output normalization.

Montreal Forced Aligner (MFA) outputs phonemes in ARPAbet format,
but our existing ML service uses IPA. This module provides conversion
between the two formats.

ARPAbet uses ASCII characters (e.g., AA, AE, TH) while IPA uses
Unicode phonetic symbols (e.g., ɑ, æ, θ).
"""

from typing import Optional


# Complete English ARPAbet to IPA mapping
# Based on CMU Pronouncing Dictionary / MFA english_us_arpa model
ARPABET_TO_IPA: dict[str, str] = {
    # Vowels (monophthongs)
    'AA': 'ɑ',   # father, hot
    'AE': 'æ',   # cat, bat
    'AH': 'ʌ',   # but, cup (stressed) / schwa when unstressed
    'AO': 'ɔ',   # thought, caught
    'EH': 'ɛ',   # bed, head
    'ER': 'ɝ',   # bird, her (r-colored)
    'IH': 'ɪ',   # bit, sit
    'IY': 'i',   # bee, see
    'UH': 'ʊ',   # book, put
    'UW': 'u',   # blue, too

    # Diphthongs
    'AW': 'aʊ',  # cow, how
    'AY': 'aɪ',  # buy, my
    'EY': 'eɪ',  # say, day
    'OW': 'oʊ',  # go, show
    'OY': 'ɔɪ',  # boy, toy

    # Consonants - stops
    'B': 'b',    # boy
    'D': 'd',    # dog
    'G': 'g',    # go
    'K': 'k',    # cat
    'P': 'p',    # pot
    'T': 't',    # top

    # Consonants - fricatives
    'DH': 'ð',   # this, that (voiced dental)
    'F': 'f',    # fish
    'S': 's',    # sun
    'SH': 'ʃ',   # she, ship
    'TH': 'θ',   # think, thing (voiceless dental)
    'V': 'v',    # voice
    'Z': 'z',    # zoo
    'ZH': 'ʒ',   # measure, vision

    # Consonants - affricates
    'CH': 'tʃ',  # church, check
    'JH': 'dʒ',  # judge, joy

    # Consonants - nasals
    'M': 'm',    # man
    'N': 'n',    # no
    'NG': 'ŋ',   # sing, ring

    # Consonants - liquids
    'L': 'l',    # love
    'R': 'ɹ',    # red (American English approximant)

    # Consonants - glides/semivowels
    'W': 'w',    # water
    'Y': 'j',    # yes

    # Consonants - other
    'HH': 'h',   # hello
}


# Reverse mapping: IPA to ARPAbet
IPA_TO_ARPABET: dict[str, str] = {v: k for k, v in ARPABET_TO_IPA.items()}


def strip_stress(arpabet: str) -> str:
    """
    Remove stress marker from ARPAbet symbol.

    MFA adds stress markers (0, 1, 2) after vowels:
    - 0 = no stress
    - 1 = primary stress
    - 2 = secondary stress

    Examples:
        'AA1' -> 'AA'
        'AH0' -> 'AH'
        'ER2' -> 'ER'
        'B' -> 'B' (consonants have no stress)

    Args:
        arpabet: ARPAbet symbol, possibly with stress marker

    Returns:
        Base ARPAbet symbol without stress
    """
    return arpabet.rstrip('012')


def get_stress(arpabet: str) -> Optional[int]:
    """
    Extract stress level from ARPAbet symbol.

    Args:
        arpabet: ARPAbet symbol, possibly with stress marker

    Returns:
        Stress level (0, 1, or 2) or None if no stress marker
    """
    if arpabet and arpabet[-1] in '012':
        return int(arpabet[-1])
    return None


def arpabet_to_ipa(arpabet: str) -> str:
    """
    Convert ARPAbet symbol to IPA.

    Args:
        arpabet: ARPAbet symbol (e.g., 'AA1', 'TH', 'NG')

    Returns:
        IPA equivalent (e.g., 'ɑ', 'θ', 'ŋ')
        Returns original symbol if no mapping exists
    """
    base = strip_stress(arpabet)
    return ARPABET_TO_IPA.get(base, arpabet)


def ipa_to_arpabet(ipa: str) -> Optional[str]:
    """
    Convert IPA symbol to ARPAbet.

    Args:
        ipa: IPA symbol (e.g., 'ɑ', 'θ', 'ŋ')

    Returns:
        ARPAbet equivalent or None if no mapping exists
    """
    return IPA_TO_ARPABET.get(ipa)


def arpabet_sequence_to_ipa(phones: list[str]) -> list[str]:
    """
    Convert a sequence of ARPAbet symbols to IPA.

    Args:
        phones: List of ARPAbet symbols

    Returns:
        List of IPA symbols
    """
    return [arpabet_to_ipa(p) for p in phones]


def is_vowel(arpabet: str) -> bool:
    """
    Check if ARPAbet symbol represents a vowel.

    Vowels in ARPAbet:
    - AA, AE, AH, AO, AW, AY, EH, ER, EY, IH, IY, OW, OY, UH, UW

    Args:
        arpabet: ARPAbet symbol (with or without stress)

    Returns:
        True if vowel, False otherwise
    """
    vowels = {'AA', 'AE', 'AH', 'AO', 'AW', 'AY', 'EH', 'ER', 'EY',
              'IH', 'IY', 'OW', 'OY', 'UH', 'UW'}
    return strip_stress(arpabet) in vowels


def get_phoneme_category(arpabet: str) -> str:
    """
    Get the category of an ARPAbet phoneme.

    Categories:
    - 'vowel': AA, AE, AH, AO, EH, ER, IH, IY, UH, UW
    - 'diphthong': AW, AY, EY, OW, OY
    - 'stop': B, D, G, K, P, T
    - 'fricative': DH, F, S, SH, TH, V, Z, ZH, HH
    - 'affricate': CH, JH
    - 'nasal': M, N, NG
    - 'liquid': L, R
    - 'glide': W, Y

    Args:
        arpabet: ARPAbet symbol

    Returns:
        Category string
    """
    base = strip_stress(arpabet)

    categories = {
        # Vowels
        'AA': 'vowel', 'AE': 'vowel', 'AH': 'vowel', 'AO': 'vowel',
        'EH': 'vowel', 'ER': 'vowel', 'IH': 'vowel', 'IY': 'vowel',
        'UH': 'vowel', 'UW': 'vowel',
        # Diphthongs
        'AW': 'diphthong', 'AY': 'diphthong', 'EY': 'diphthong',
        'OW': 'diphthong', 'OY': 'diphthong',
        # Stops
        'B': 'stop', 'D': 'stop', 'G': 'stop',
        'K': 'stop', 'P': 'stop', 'T': 'stop',
        # Fricatives
        'DH': 'fricative', 'F': 'fricative', 'S': 'fricative',
        'SH': 'fricative', 'TH': 'fricative', 'V': 'fricative',
        'Z': 'fricative', 'ZH': 'fricative', 'HH': 'fricative',
        # Affricates
        'CH': 'affricate', 'JH': 'affricate',
        # Nasals
        'M': 'nasal', 'N': 'nasal', 'NG': 'nasal',
        # Liquids
        'L': 'liquid', 'R': 'liquid',
        # Glides
        'W': 'glide', 'Y': 'glide',
    }

    return categories.get(base, 'unknown')
