"""
Phoneme-level alignment between audio IPA and text IPA.

Uses sequence alignment algorithms to match phonemes and identify substitutions.
"""

from typing import List, Tuple

from .normalizer import IPANormalizer


class PhonemeAligner:
    """
    Aligns audio IPA with expected text IPA at the phoneme level.

    Uses dynamic programming (Needleman-Wunsch) to find the best alignment
    between two IPA sequences.
    """

    def __init__(self, keep_stress: bool = True):
        """
        Initialize phoneme aligner.

        Args:
            keep_stress: If True (default), include stress markers in alignment.
                        If False, stress is ignored for comparison.
        """
        self.normalizer = IPANormalizer(keep_stress=keep_stress, keep_length=True)

    def extract_phonemes(self, ipa_string: str) -> List[str]:
        """
        Extract individual phonemes from an IPA string.

        Args:
            ipa_string: IPA transcription string

        Returns:
            List of phoneme strings
        """
        return self.normalizer.extract_phonemes(ipa_string)

    def phoneme_similarity(self, p1: str, p2: str) -> float:
        """
        Calculate similarity between two phonemes.

        Args:
            p1: First phoneme
            p2: Second phoneme

        Returns:
            Similarity score (0.0 to 1.0)
            - 1.0 = exact match
            - 0.8 = very similar (same base, different stress/length)
            - 0.5 = phonetically similar (voicing pairs, etc.)
            - 0.0 = completely different
        """
        if p1 == p2:
            return 1.0

        # Strip stress markers for comparison
        p1_base = p1.lstrip('ˈˌ')
        p2_base = p2.lstrip('ˈˌ')

        if p1_base == p2_base:
            return 0.8  # Same phoneme, different stress

        # Strip length markers too
        p1_core = p1_base.rstrip(':ː')
        p2_core = p2_base.rstrip(':ː')

        if p1_core == p2_core:
            return 0.8  # Same phoneme, different length

        # Check for phonetically similar pairs
        similar_pairs = [
            # Voicing pairs (very common confusion)
            ('p', 'b'), ('t', 'd'), ('k', 'g'),
            ('f', 'v'), ('s', 'z'), ('θ', 'ð'), ('ʃ', 'ʒ'),
            # Nasals
            ('m', 'n'), ('n', 'ŋ'),
            # Liquids
            ('l', 'ɹ'), ('r', 'ɹ'), ('l', 'r'),
            # Vowel pairs (common confusions)
            ('i', 'ɪ'), ('u', 'ʊ'), ('e', 'ɛ'), ('o', 'ɔ'),
            ('æ', 'ɛ'), ('ɑ', 'ɔ'), ('ʌ', 'ə'),
            # Diphthong components
            ('eɪ', 'e'), ('aɪ', 'a'), ('ɔɪ', 'ɔ'),
            ('oʊ', 'o'), ('aʊ', 'a'),
        ]

        for pair in similar_pairs:
            if (p1_core in pair and p2_core in pair) or \
               (p1_core == pair[0] and p2_core == pair[1]) or \
               (p1_core == pair[1] and p2_core == pair[0]):
                return 0.5  # Phonetically similar

        return 0.0  # Completely different

    def align(
        self,
        audio_ipa: str,
        text_ipa: str
    ) -> List[Tuple[str, str, str]]:
        """
        Align audio IPA with text IPA at phoneme level.

        Uses dynamic programming to find optimal alignment.

        Args:
            audio_ipa: IPA from audio (Whisper)
            text_ipa: IPA from text (gruut)

        Returns:
            List of (expected_phoneme, actual_phoneme, match_type) tuples
            match_type can be:
                - 'match': Phonemes match exactly or nearly (similarity > 0.7)
                - 'substitute': Phoneme substituted
                - 'delete': Expected phoneme missing (not said)
                - 'insert': Extra phoneme said (not expected)
        """
        audio_phonemes = self.extract_phonemes(audio_ipa)
        text_phonemes = self.extract_phonemes(text_ipa)

        return self._needleman_wunsch(text_phonemes, audio_phonemes)

    def _needleman_wunsch(
        self,
        seq1: List[str],
        seq2: List[str]
    ) -> List[Tuple[str, str, str]]:
        """
        Needleman-Wunsch global alignment algorithm.

        Args:
            seq1: Expected phoneme sequence (text)
            seq2: Actual phoneme sequence (audio)

        Returns:
            List of aligned (expected, actual, type) tuples
        """
        m, n = len(seq1), len(seq2)

        # Scoring parameters
        MATCH_SCORE = 2
        SIMILAR_SCORE = 1
        MISMATCH_PENALTY = -1
        GAP_PENALTY = -2

        # Initialize DP table
        dp = [[0] * (n + 1) for _ in range(m + 1)]

        # Initialize first row and column (gaps)
        for i in range(m + 1):
            dp[i][0] = i * GAP_PENALTY
        for j in range(n + 1):
            dp[0][j] = j * GAP_PENALTY

        # Fill DP table
        for i in range(1, m + 1):
            for j in range(1, n + 1):
                similarity = self.phoneme_similarity(seq1[i-1], seq2[j-1])

                if similarity >= 0.8:
                    score = MATCH_SCORE
                elif similarity >= 0.5:
                    score = SIMILAR_SCORE
                else:
                    score = MISMATCH_PENALTY

                scores = [
                    dp[i-1][j-1] + score,      # Match/mismatch
                    dp[i-1][j] + GAP_PENALTY,  # Delete from seq1
                    dp[i][j-1] + GAP_PENALTY   # Insert to seq1
                ]
                dp[i][j] = max(scores)

        # Traceback to find alignment
        alignment = []
        i, j = m, n

        while i > 0 or j > 0:
            if i > 0 and j > 0:
                similarity = self.phoneme_similarity(seq1[i-1], seq2[j-1])

                if similarity >= 0.8:
                    score = MATCH_SCORE
                elif similarity >= 0.5:
                    score = SIMILAR_SCORE
                else:
                    score = MISMATCH_PENALTY

                if dp[i][j] == dp[i-1][j-1] + score:
                    # Match or substitution
                    expected = seq1[i-1]
                    actual = seq2[j-1]

                    if similarity >= 0.8:
                        match_type = 'match'
                    else:
                        match_type = 'substitute'

                    alignment.append((expected, actual, match_type))
                    i -= 1
                    j -= 1
                    continue

            if i > 0 and dp[i][j] == dp[i-1][j] + GAP_PENALTY:
                # Delete from seq1 (expected phoneme not said)
                alignment.append((seq1[i-1], '-', 'delete'))
                i -= 1
            elif j > 0 and dp[i][j] == dp[i][j-1] + GAP_PENALTY:
                # Insert to seq1 (extra phoneme said)
                alignment.append(('-', seq2[j-1], 'insert'))
                j -= 1
            else:
                # Fallback
                if i > 0:
                    alignment.append((seq1[i-1], '-', 'delete'))
                    i -= 1
                if j > 0:
                    alignment.append(('-', seq2[j-1], 'insert'))
                    j -= 1

        # Reverse to get correct order
        alignment.reverse()

        return alignment


def align_phonemes(audio_ipa: str, text_ipa: str) -> List[Tuple[str, str, str]]:
    """
    Convenience function to align two IPA strings.

    Args:
        audio_ipa: IPA from audio
        text_ipa: IPA from text

    Returns:
        List of (expected, actual, type) tuples
    """
    aligner = PhonemeAligner()
    return aligner.align(audio_ipa, text_ipa)
