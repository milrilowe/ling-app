"""
Audio to IPA conversion using Whisper fine-tuned model.

Uses neuralang/ipa-whisper-small model to transcribe audio directly to IPA.
"""

import warnings
from typing import Optional

import numpy as np
import torch
from transformers import WhisperProcessor, WhisperForConditionalGeneration

from .post_processor import IPAPostProcessor


class WhisperIPAConverter:
    """
    Converts audio to IPA phonetic transcription using Whisper.

    Uses the neuralang/ipa-whisper-small model, which is fine-tuned
    to output IPA instead of standard text.
    """

    DEFAULT_MODEL = "neurlang/ipa-whisper-small"

    def __init__(
        self,
        model_name: str = DEFAULT_MODEL,
        device: Optional[str] = None,
        post_process_language: str = 'en'
    ):
        """
        Initialize Whisper IPA converter.

        Args:
            model_name: HuggingFace model name (default: neurlang/ipa-whisper-small)
            device: Device to run model on ('cuda', 'cpu', or None for auto-detect)
            post_process_language: Language for post-processing corrections (default: 'en')
        """
        self.model_name = model_name
        self.post_processor = IPAPostProcessor(language=post_process_language)

        # Auto-detect device if not specified
        if device is None:
            self.device = "cuda" if torch.cuda.is_available() else "cpu"
        else:
            self.device = device

        print(f"Loading Whisper IPA model: {model_name}")
        print(f"Using device: {self.device}")

        # Load processor and model
        self.processor = WhisperProcessor.from_pretrained(model_name)
        self.model = WhisperForConditionalGeneration.from_pretrained(model_name)

        # Move model to device
        self.model.to(self.device)

        # CRITICAL: Configure model to avoid errors
        # These settings are required for the fine-tuned IPA model
        self.model.config.forced_decoder_ids = None
        self.model.config.suppress_tokens = []

        # Also set on generation config
        if hasattr(self.model, 'generation_config'):
            self.model.generation_config.forced_decoder_ids = None

        print("Model loaded successfully!")

    def audio_to_ipa(
        self,
        audio_array: np.ndarray,
        sampling_rate: int = 16000,
        language: Optional[str] = None,
        num_beams: int = 5,
        return_confidence: bool = False
    ) -> str:
        """
        Convert audio to IPA transcription.

        Args:
            audio_array: Audio samples as numpy array
            sampling_rate: Sample rate of audio (should be 16000 for Whisper)
            language: Language code for language-specific hints (e.g., 'en', 'es', 'fr')
            num_beams: Number of beams for beam search (default: 5, use 1 for greedy)
            return_confidence: If True, returns tuple of (transcription, confidence_score)

        Returns:
            IPA transcription string (e.g., "hɛˈloʊ wˈɜrld")
            Or tuple of (transcription, confidence) if return_confidence=True

        Raises:
            ValueError: If sampling rate is not 16000
        """
        if sampling_rate != 16000:
            warnings.warn(
                f"Whisper expects 16kHz audio, got {sampling_rate}Hz. "
                f"Results may be degraded."
            )

        # Process audio through WhisperProcessor
        input_features = self.processor(
            audio_array,
            sampling_rate=sampling_rate,
            return_tensors="pt"
        ).input_features

        # Move to same device as model
        input_features = input_features.to(self.device)

        # Prepare generation kwargs
        generate_kwargs = {
            "num_beams": num_beams,
            "temperature": 0.0,  # Deterministic output
            "do_sample": False,  # No sampling
            "repetition_penalty": 1.2,  # Penalize repetition
            "length_penalty": 1.0,  # Control length
            "max_length": 448,  # Whisper max length
        }

        # Add language-specific hints if provided
        if language:
            try:
                forced_decoder_ids = self.processor.get_decoder_prompt_ids(
                    language=language,
                    task="transcribe"
                )
                generate_kwargs["forced_decoder_ids"] = forced_decoder_ids
            except Exception:
                # Language not supported, continue without hints
                pass

        # Add confidence tracking if requested
        if return_confidence:
            generate_kwargs["return_dict_in_generate"] = True
            generate_kwargs["output_scores"] = True

        # Generate IPA tokens
        with torch.no_grad():
            output = self.model.generate(input_features, **generate_kwargs)

        # Handle different output formats
        if return_confidence:
            predicted_ids = output.sequences
            # Calculate confidence from scores
            confidence = self._calculate_confidence(output.scores)
        else:
            predicted_ids = output

        # Decode to IPA string
        transcription = self.processor.batch_decode(
            predicted_ids,
            skip_special_tokens=True
        )[0]

        transcription = transcription.strip()

        # Apply post-processing to fix common errors
        transcription = self.post_processor.post_process(
            transcription,
            confidence=confidence if return_confidence else None
        )

        if return_confidence:
            return transcription, confidence
        return transcription

    def _calculate_confidence(self, scores) -> float:
        """
        Calculate average confidence score from generation scores.

        Args:
            scores: Tuple of score tensors from model generation

        Returns:
            Average confidence score (0-1)
        """
        if not scores:
            return 1.0

        # Convert scores to probabilities and average
        confidences = []
        for score in scores:
            probs = torch.softmax(score, dim=-1)
            max_prob = torch.max(probs).item()
            confidences.append(max_prob)

        return sum(confidences) / len(confidences) if confidences else 1.0

    def audio_to_ipa_chunked(
        self,
        audio_array: np.ndarray,
        sampling_rate: int = 16000,
        chunk_duration: float = 30.0,
        language: Optional[str] = None,
        num_beams: int = 5
    ) -> str:
        """
        Convert long audio to IPA by processing in chunks.

        For audio longer than chunk_duration, this method splits
        the audio into chunks and processes each separately, then
        concatenates the results.

        Args:
            audio_array: Audio samples as numpy array
            sampling_rate: Sample rate of audio
            chunk_duration: Duration of each chunk in seconds (default: 30)
            language: Language code for hints
            num_beams: Number of beams for beam search

        Returns:
            IPA transcription string
        """
        duration = len(audio_array) / sampling_rate

        # If audio is short enough, process directly
        if duration <= chunk_duration:
            return self.audio_to_ipa(
                audio_array, sampling_rate, language, num_beams
            )

        # Split into chunks
        chunk_samples = int(chunk_duration * sampling_rate)
        chunks = []

        for i in range(0, len(audio_array), chunk_samples):
            chunk = audio_array[i:i + chunk_samples]
            if len(chunk) > sampling_rate * 0.1:  # Skip chunks < 0.1s
                chunks.append(chunk)

        # Process each chunk
        ipa_results = []
        for i, chunk in enumerate(chunks):
            print(f"Processing chunk {i+1}/{len(chunks)}...")
            chunk_ipa = self.audio_to_ipa(
                chunk, sampling_rate, language, num_beams
            )
            ipa_results.append(chunk_ipa)

        # Join with spaces
        return ' '.join(ipa_results)

    def transcribe_file(self, audio_path: str) -> str:
        """
        Convenience method to transcribe an audio file directly.

        Args:
            audio_path: Path to audio file

        Returns:
            IPA transcription string
        """
        from src.audio.loader import load_audio_file

        audio_array, sample_rate = load_audio_file(audio_path)
        return self.audio_to_ipa(audio_array, sample_rate)


def audio_to_ipa(
    audio_array: np.ndarray,
    sampling_rate: int = 16000,
    model_name: str = WhisperIPAConverter.DEFAULT_MODEL
) -> str:
    """
    Convenience function to convert audio to IPA.

    Args:
        audio_array: Audio samples as numpy array
        sampling_rate: Sample rate of audio
        model_name: HuggingFace model name

    Returns:
        IPA transcription string
    """
    converter = WhisperIPAConverter(model_name=model_name)
    return converter.audio_to_ipa(audio_array, sampling_rate)
