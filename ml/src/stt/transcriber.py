"""
Speech-to-text transcription using faster-whisper.

Uses CTranslate2-based Whisper implementation for efficient inference.
"""

from dataclasses import dataclass
from typing import Optional

import numpy as np
from faster_whisper import WhisperModel


@dataclass
class TranscriptionResult:
    """Result of a transcription."""
    text: str
    language: str
    duration: float


class FasterWhisperTranscriber:
    """
    Transcribes audio to text using faster-whisper.

    Uses the CTranslate2 Whisper implementation which is significantly
    faster than the original OpenAI implementation.
    """

    def __init__(
        self,
        model_size: str = "medium",
        device: Optional[str] = None,
        compute_type: str = "auto"
    ):
        """
        Initialize the transcriber.

        Args:
            model_size: Whisper model size ('tiny', 'base', 'small', 'medium', 'large-v3')
            device: Device to run on ('cuda', 'cpu', or None for auto-detect)
            compute_type: Compute type ('auto', 'float16', 'int8', etc.)
        """
        self.model_size = model_size

        # Auto-detect device
        if device is None:
            import torch
            device = "cuda" if torch.cuda.is_available() else "cpu"

        self.device = device

        # Adjust compute type based on device
        if compute_type == "auto":
            compute_type = "float16" if device == "cuda" else "int8"

        print(f"Loading faster-whisper model: {model_size}")
        print(f"Device: {device}, Compute type: {compute_type}")

        self.model = WhisperModel(
            model_size,
            device=device,
            compute_type=compute_type
        )

        print("Faster-whisper model loaded successfully!")

    def transcribe(
        self,
        audio_array: np.ndarray,
        sample_rate: int = 16000,
        language: Optional[str] = None
    ) -> TranscriptionResult:
        """
        Transcribe audio to text.

        Args:
            audio_array: Audio samples as numpy array (mono, float32)
            sample_rate: Sample rate of audio (should be 16000 for Whisper)
            language: Language code (e.g., 'en', 'es'). None for auto-detect.

        Returns:
            TranscriptionResult with text, detected language, and duration
        """
        if sample_rate != 16000:
            print(f"Warning: Whisper expects 16kHz audio, got {sample_rate}Hz")

        # Ensure float32
        if audio_array.dtype != np.float32:
            audio_array = audio_array.astype(np.float32)

        # Run transcription
        segments, info = self.model.transcribe(
            audio_array,
            language=language,
            beam_size=5,
            vad_filter=True,  # Filter out silence
            vad_parameters=dict(
                min_silence_duration_ms=500,
            )
        )

        # Collect all segment text
        text_parts = []
        for segment in segments:
            text_parts.append(segment.text.strip())

        full_text = " ".join(text_parts).strip()

        return TranscriptionResult(
            text=full_text,
            language=info.language,
            duration=info.duration
        )
