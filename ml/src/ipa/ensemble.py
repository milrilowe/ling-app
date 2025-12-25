"""
Ensemble approach for improved IPA transcription accuracy.

Uses multiple Whisper models and voting/averaging to improve results.
"""

from collections import Counter
from typing import List, Optional, Tuple

import numpy as np

from .audio_to_ipa import WhisperIPAConverter


class EnsembleIPAConverter:
    """
    Ensemble converter that uses multiple models for better accuracy.

    Runs audio through multiple Whisper IPA models and uses voting
    or weighted averaging to select the best transcription.
    """

    DEFAULT_MODELS = [
        "neurlang/ipa-whisper-small",
    ]

    def __init__(
        self,
        model_names: Optional[List[str]] = None,
        device: Optional[str] = None,
        post_process_language: str = 'en'
    ):
        """
        Initialize ensemble IPA converter.

        Args:
            model_names: List of model names to use (default: DEFAULT_MODELS)
            device: Device to run models on
            post_process_language: Language for post-processing
        """
        if model_names is None:
            model_names = self.DEFAULT_MODELS

        self.model_names = model_names
        self.converters = []

        print(f"Loading {len(model_names)} models for ensemble...")

        for model_name in model_names:
            try:
                converter = WhisperIPAConverter(
                    model_name=model_name,
                    device=device,
                    post_process_language=post_process_language
                )
                self.converters.append(converter)
            except Exception as e:
                print(f"Warning: Could not load {model_name}: {e}")

        if not self.converters:
            raise ValueError("No models could be loaded for ensemble")

        print(f"Successfully loaded {len(self.converters)} models")

    def audio_to_ipa_ensemble(
        self,
        audio_array: np.ndarray,
        sampling_rate: int = 16000,
        language: Optional[str] = None,
        num_beams: int = 5,
        strategy: str = 'vote'
    ) -> str:
        """
        Convert audio to IPA using ensemble approach.

        Args:
            audio_array: Audio samples as numpy array
            sampling_rate: Sample rate of audio
            language: Language code for hints
            num_beams: Number of beams for beam search
            strategy: Ensemble strategy ('vote', 'confidence', or 'first')
                - 'vote': Use most common result (simple voting)
                - 'confidence': Use result with highest confidence
                - 'first': Use first model only (no ensemble)

        Returns:
            IPA transcription string
        """
        if strategy == 'first' or len(self.converters) == 1:
            # Just use the first model
            return self.converters[0].audio_to_ipa(
                audio_array, sampling_rate, language, num_beams
            )

        # Get results from all models
        results = []
        confidences = []

        for i, converter in enumerate(self.converters):
            print(f"Running model {i+1}/{len(self.converters)}...")

            if strategy == 'confidence':
                ipa, conf = converter.audio_to_ipa(
                    audio_array, sampling_rate, language, num_beams,
                    return_confidence=True
                )
                results.append(ipa)
                confidences.append(conf)
            else:
                ipa = converter.audio_to_ipa(
                    audio_array, sampling_rate, language, num_beams
                )
                results.append(ipa)

        # Apply ensemble strategy
        if strategy == 'vote':
            return self._vote_strategy(results)
        elif strategy == 'confidence':
            return self._confidence_strategy(results, confidences)
        else:
            raise ValueError(f"Unknown strategy: {strategy}")

    def _vote_strategy(self, results: List[str]) -> str:
        """
        Use simple voting to pick most common result.

        Args:
            results: List of IPA strings from different models

        Returns:
            Most common IPA string
        """
        # Count occurrences
        counter = Counter(results)

        # Return most common
        most_common = counter.most_common(1)[0][0]

        return most_common

    def _confidence_strategy(
        self,
        results: List[str],
        confidences: List[float]
    ) -> str:
        """
        Use confidence scores to pick best result.

        Args:
            results: List of IPA strings
            confidences: List of confidence scores

        Returns:
            IPA string with highest confidence
        """
        if not confidences:
            return results[0]

        # Find index of highest confidence
        max_idx = confidences.index(max(confidences))

        return results[max_idx]

    def _character_voting_strategy(
        self,
        results: List[str]
    ) -> str:
        """
        Use character-level voting for more fine-grained ensemble.

        For each position, vote on the most common character across models.
        More sophisticated than string-level voting.

        Args:
            results: List of IPA strings

        Returns:
            IPA string constructed from character-level voting
        """
        if not results:
            return ""

        # Pad all results to same length
        max_len = max(len(r) for r in results)
        padded = [r.ljust(max_len) for r in results]

        # Vote for each position
        voted = []
        for i in range(max_len):
            chars = [r[i] for r in padded]
            # Get most common character
            most_common_char = Counter(chars).most_common(1)[0][0]
            if most_common_char != ' ' or i < max_len - 1:
                voted.append(most_common_char)

        return ''.join(voted).rstrip()


def audio_to_ipa_ensemble(
    audio_array: np.ndarray,
    sampling_rate: int = 16000,
    model_names: Optional[List[str]] = None,
    language: Optional[str] = None,
    strategy: str = 'vote'
) -> str:
    """
    Convenience function for ensemble IPA conversion.

    Args:
        audio_array: Audio samples
        sampling_rate: Sample rate
        model_names: List of model names
        language: Language code
        strategy: Ensemble strategy

    Returns:
        IPA transcription string
    """
    converter = EnsembleIPAConverter(
        model_names=model_names,
        post_process_language=language or 'en'
    )

    return converter.audio_to_ipa_ensemble(
        audio_array, sampling_rate, language, strategy=strategy
    )
