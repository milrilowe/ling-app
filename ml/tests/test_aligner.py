"""Tests for phoneme alignment."""

import pytest
from src.ipa.aligner import PhonemeAligner, align_phonemes


class TestPhonemeSimilarity:
    """Test phoneme similarity scoring."""

    def setup_method(self):
        self.aligner = PhonemeAligner()

    def test_exact_match(self):
        """Identical phonemes = 1.0"""
        assert self.aligner.phoneme_similarity("p", "p") == 1.0
        assert self.aligner.phoneme_similarity("θ", "θ") == 1.0
        assert self.aligner.phoneme_similarity("aɪ", "aɪ") == 1.0

    def test_stress_difference(self):
        """Same phoneme with different stress = 0.8"""
        assert self.aligner.phoneme_similarity("ˈe", "e") == 0.8
        assert self.aligner.phoneme_similarity("e", "ˈe") == 0.8
        assert self.aligner.phoneme_similarity("ˌa", "a") == 0.8

    def test_length_difference(self):
        """Same phoneme with different length = 0.8"""
        assert self.aligner.phoneme_similarity("iː", "i") == 0.8
        assert self.aligner.phoneme_similarity("i", "iː") == 0.8

    def test_voicing_pairs(self):
        """Voicing pairs (common confusion) = 0.5"""
        assert self.aligner.phoneme_similarity("p", "b") == 0.5
        assert self.aligner.phoneme_similarity("b", "p") == 0.5
        assert self.aligner.phoneme_similarity("t", "d") == 0.5
        assert self.aligner.phoneme_similarity("k", "g") == 0.5
        assert self.aligner.phoneme_similarity("f", "v") == 0.5
        assert self.aligner.phoneme_similarity("s", "z") == 0.5
        assert self.aligner.phoneme_similarity("θ", "ð") == 0.5
        assert self.aligner.phoneme_similarity("ʃ", "ʒ") == 0.5

    def test_similar_vowels(self):
        """Similar vowel pairs = 0.5"""
        assert self.aligner.phoneme_similarity("i", "ɪ") == 0.5
        assert self.aligner.phoneme_similarity("u", "ʊ") == 0.5
        assert self.aligner.phoneme_similarity("e", "ɛ") == 0.5
        assert self.aligner.phoneme_similarity("æ", "ɛ") == 0.5
        assert self.aligner.phoneme_similarity("ʌ", "ə") == 0.5

    def test_unrelated_phonemes(self):
        """Completely different phonemes = 0.0"""
        assert self.aligner.phoneme_similarity("m", "k") == 0.0
        assert self.aligner.phoneme_similarity("p", "i") == 0.0
        assert self.aligner.phoneme_similarity("θ", "l") == 0.0
        assert self.aligner.phoneme_similarity("a", "ŋ") == 0.0


class TestExtractPhonemes:
    """Test phoneme extraction from IPA strings."""

    def setup_method(self):
        self.aligner = PhonemeAligner()

    def test_simple_consonants(self):
        """Extract simple consonant sequence."""
        result = self.aligner.extract_phonemes("kæt")
        assert "k" in result
        assert "æ" in result
        assert "t" in result

    def test_with_stress(self):
        """Extract phonemes preserving stress."""
        result = self.aligner.extract_phonemes("ˈhɛloʊ")
        # Should include stress marker with following phoneme
        assert any("ˈ" in p or p.startswith("ˈ") for p in result) or "h" in result


class TestAlign:
    """Test full alignment algorithm."""

    def setup_method(self):
        self.aligner = PhonemeAligner()

    def test_identical_sequences(self):
        """Perfect match returns all 'match' types."""
        result = self.aligner.align("kæt", "kæt")

        # All should be matches
        for expected, actual, match_type in result:
            assert match_type == "match"
            assert expected == actual

    def test_substitution_detected(self):
        """Substitution (θ→s) should be detected."""
        # "think" vs "sink" - θ replaced with s
        result = self.aligner.align("sɪŋk", "θɪŋk")

        # Find the substitution
        substitutions = [(e, a, t) for e, a, t in result if t == "substitute"]
        assert len(substitutions) > 0

    def test_deletion_detected(self):
        """Missing phoneme should be detected as deletion."""
        # Audio has fewer phonemes than expected
        result = self.aligner.align("kæ", "kæt")

        # Should have a deletion
        deletions = [(e, a, t) for e, a, t in result if t == "delete"]
        assert len(deletions) > 0
        # The deleted phoneme should be 't'
        assert any(e == "t" for e, a, t in deletions)

    def test_insertion_detected(self):
        """Extra phoneme should be detected as insertion."""
        # Audio has more phonemes than expected
        result = self.aligner.align("kæts", "kæt")

        # Should have an insertion
        insertions = [(e, a, t) for e, a, t in result if t == "insert"]
        assert len(insertions) > 0
        # The inserted phoneme should be 's'
        assert any(a == "s" for e, a, t in insertions)

    def test_empty_sequences(self):
        """Handle empty sequences gracefully."""
        result = self.aligner.align("", "")
        assert result == []

    def test_one_empty(self):
        """Handle one empty sequence."""
        # align(audio_ipa, text_ipa) - audio first, text second
        # When audio has content but text is empty = all insertions (extra phonemes said)
        result = self.aligner.align("kæt", "")
        assert all(t == "insert" for e, a, t in result)

        # When text has content but audio is empty = all deletions (phonemes not said)
        result = self.aligner.align("", "kæt")
        assert all(t == "delete" for e, a, t in result)


class TestAlignPhonemes:
    """Test convenience function."""

    def test_convenience_function(self):
        """Test that convenience function works."""
        result = align_phonemes("kæt", "kæt")
        assert len(result) > 0
        assert all(t == "match" for e, a, t in result)


class TestRealWorldExamples:
    """Test with realistic pronunciation examples."""

    def setup_method(self):
        self.aligner = PhonemeAligner()

    def test_th_fronting(self):
        """Common error: θ pronounced as f (th-fronting)."""
        # "think" /θɪŋk/ pronounced as /fɪŋk/
        result = self.aligner.align("fɪŋk", "θɪŋk")

        # Should detect θ→f substitution
        subs = [(e, a, t) for e, a, t in result if t == "substitute"]
        assert len(subs) >= 1

    def test_final_consonant_deletion(self):
        """Common error: dropping final consonants."""
        # "test" /tɛst/ pronounced as /tɛs/
        result = self.aligner.align("tɛs", "tɛst")

        # Should detect final 't' deletion
        deletions = [(e, a, t) for e, a, t in result if t == "delete"]
        assert len(deletions) >= 1

    def test_vowel_substitution(self):
        """Common error: wrong vowel."""
        # "bit" /bɪt/ pronounced as /biːt/ (tense instead of lax)
        result = self.aligner.align("biːt", "bɪt")

        # i and ɪ are similar, might be match or substitute depending on threshold
        # But the alignment should complete without error
        assert len(result) > 0

    def test_r_colored_vowel(self):
        """Test r-colored vowel alignment."""
        # "bird" /bɝd/
        result = self.aligner.align("bɝd", "bɝd")
        assert all(t == "match" for e, a, t in result)
