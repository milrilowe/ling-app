"""Tests for ARPAbet to IPA conversion."""

import pytest
from src.ipa.arpabet_ipa import (
    arpabet_to_ipa,
    strip_stress,
    get_stress,
    is_vowel,
    arpabet_sequence_to_ipa,
    get_phoneme_category,
)


class TestArpabetToIpa:
    """Test ARPAbet to IPA conversion."""

    def test_vowels(self):
        """Test vowel conversions."""
        assert arpabet_to_ipa("AA") == "ɑ"  # father
        assert arpabet_to_ipa("AE") == "æ"  # cat
        assert arpabet_to_ipa("AH") == "ʌ"  # but
        assert arpabet_to_ipa("AO") == "ɔ"  # thought
        assert arpabet_to_ipa("EH") == "ɛ"  # bed
        assert arpabet_to_ipa("ER") == "ɝ"  # bird
        assert arpabet_to_ipa("IH") == "ɪ"  # bit
        assert arpabet_to_ipa("IY") == "i"  # bee
        assert arpabet_to_ipa("UH") == "ʊ"  # book
        assert arpabet_to_ipa("UW") == "u"  # blue

    def test_diphthongs(self):
        """Test diphthong conversions."""
        assert arpabet_to_ipa("AW") == "aʊ"  # cow
        assert arpabet_to_ipa("AY") == "aɪ"  # buy
        assert arpabet_to_ipa("EY") == "eɪ"  # say
        assert arpabet_to_ipa("OW") == "oʊ"  # go
        assert arpabet_to_ipa("OY") == "ɔɪ"  # boy

    def test_consonants_fricatives(self):
        """Test fricative consonant conversions."""
        assert arpabet_to_ipa("TH") == "θ"  # think
        assert arpabet_to_ipa("DH") == "ð"  # this
        assert arpabet_to_ipa("SH") == "ʃ"  # she
        assert arpabet_to_ipa("ZH") == "ʒ"  # measure
        assert arpabet_to_ipa("F") == "f"
        assert arpabet_to_ipa("V") == "v"
        assert arpabet_to_ipa("S") == "s"
        assert arpabet_to_ipa("Z") == "z"
        assert arpabet_to_ipa("HH") == "h"

    def test_consonants_stops(self):
        """Test stop consonant conversions."""
        assert arpabet_to_ipa("P") == "p"
        assert arpabet_to_ipa("B") == "b"
        assert arpabet_to_ipa("T") == "t"
        assert arpabet_to_ipa("D") == "d"
        assert arpabet_to_ipa("K") == "k"
        assert arpabet_to_ipa("G") == "g"

    def test_consonants_affricates(self):
        """Test affricate conversions."""
        assert arpabet_to_ipa("CH") == "tʃ"  # church
        assert arpabet_to_ipa("JH") == "dʒ"  # judge

    def test_consonants_nasals(self):
        """Test nasal conversions."""
        assert arpabet_to_ipa("M") == "m"
        assert arpabet_to_ipa("N") == "n"
        assert arpabet_to_ipa("NG") == "ŋ"  # sing

    def test_consonants_liquids_glides(self):
        """Test liquid and glide conversions."""
        assert arpabet_to_ipa("L") == "l"
        assert arpabet_to_ipa("R") == "ɹ"
        assert arpabet_to_ipa("W") == "w"
        assert arpabet_to_ipa("Y") == "j"

    def test_with_stress_markers(self):
        """Test that stress markers are stripped."""
        assert arpabet_to_ipa("AA1") == "ɑ"
        assert arpabet_to_ipa("AE0") == "æ"
        assert arpabet_to_ipa("IY2") == "i"
        assert arpabet_to_ipa("ER1") == "ɝ"

    def test_unknown_symbol(self):
        """Test that unknown symbols return as-is."""
        assert arpabet_to_ipa("XX") == "XX"
        assert arpabet_to_ipa("???") == "???"


class TestStripStress:
    """Test stress marker stripping."""

    def test_primary_stress(self):
        assert strip_stress("AA1") == "AA"
        assert strip_stress("IY1") == "IY"

    def test_no_stress(self):
        assert strip_stress("AH0") == "AH"
        assert strip_stress("ER0") == "ER"

    def test_secondary_stress(self):
        assert strip_stress("OW2") == "OW"
        assert strip_stress("AE2") == "AE"

    def test_consonants_unchanged(self):
        """Consonants don't have stress markers."""
        assert strip_stress("TH") == "TH"
        assert strip_stress("P") == "P"
        assert strip_stress("NG") == "NG"


class TestGetStress:
    """Test stress level extraction."""

    def test_primary_stress(self):
        assert get_stress("AA1") == 1
        assert get_stress("IY1") == 1

    def test_no_stress(self):
        assert get_stress("AH0") == 0

    def test_secondary_stress(self):
        assert get_stress("OW2") == 2

    def test_no_marker(self):
        """Consonants and unmarked symbols return None."""
        assert get_stress("TH") is None
        assert get_stress("P") is None
        assert get_stress("AA") is None


class TestIsVowel:
    """Test vowel detection."""

    def test_vowels(self):
        assert is_vowel("AA") is True
        assert is_vowel("AE") is True
        assert is_vowel("IY") is True
        assert is_vowel("ER") is True

    def test_vowels_with_stress(self):
        assert is_vowel("AA1") is True
        assert is_vowel("IY0") is True
        assert is_vowel("ER2") is True

    def test_diphthongs(self):
        assert is_vowel("AW") is True
        assert is_vowel("AY") is True
        assert is_vowel("EY") is True
        assert is_vowel("OW") is True
        assert is_vowel("OY") is True

    def test_consonants(self):
        assert is_vowel("TH") is False
        assert is_vowel("P") is False
        assert is_vowel("NG") is False
        assert is_vowel("SH") is False
        assert is_vowel("CH") is False


class TestArpabetSequenceToIpa:
    """Test full sequence conversion."""

    def test_simple_word(self):
        # "cat" = K AE1 T
        result = arpabet_sequence_to_ipa(["K", "AE1", "T"])
        assert result == ["k", "æ", "t"]

    def test_word_with_diphthong(self):
        # "buy" = B AY1
        result = arpabet_sequence_to_ipa(["B", "AY1"])
        assert result == ["b", "aɪ"]

    def test_word_with_affricates(self):
        # "church" = CH ER1 CH
        result = arpabet_sequence_to_ipa(["CH", "ER1", "CH"])
        assert result == ["tʃ", "ɝ", "tʃ"]

    def test_complex_word(self):
        # "thinking" = TH IH1 NG K IH0 NG
        result = arpabet_sequence_to_ipa(["TH", "IH1", "NG", "K", "IH0", "NG"])
        assert result == ["θ", "ɪ", "ŋ", "k", "ɪ", "ŋ"]


class TestGetPhonemeCategory:
    """Test phoneme categorization."""

    def test_vowels(self):
        assert get_phoneme_category("AA") == "vowel"
        assert get_phoneme_category("IY1") == "vowel"

    def test_diphthongs(self):
        assert get_phoneme_category("AW") == "diphthong"
        assert get_phoneme_category("OY1") == "diphthong"

    def test_stops(self):
        assert get_phoneme_category("P") == "stop"
        assert get_phoneme_category("T") == "stop"
        assert get_phoneme_category("K") == "stop"

    def test_fricatives(self):
        assert get_phoneme_category("TH") == "fricative"
        assert get_phoneme_category("SH") == "fricative"
        assert get_phoneme_category("F") == "fricative"

    def test_affricates(self):
        assert get_phoneme_category("CH") == "affricate"
        assert get_phoneme_category("JH") == "affricate"

    def test_nasals(self):
        assert get_phoneme_category("M") == "nasal"
        assert get_phoneme_category("N") == "nasal"
        assert get_phoneme_category("NG") == "nasal"

    def test_liquids(self):
        assert get_phoneme_category("L") == "liquid"
        assert get_phoneme_category("R") == "liquid"

    def test_glides(self):
        assert get_phoneme_category("W") == "glide"
        assert get_phoneme_category("Y") == "glide"

    def test_unknown(self):
        assert get_phoneme_category("XX") == "unknown"
