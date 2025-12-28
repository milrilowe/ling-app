"""
Text-to-speech synthesis using Chatterbox.

Uses Resemble AI's Chatterbox TTS for high-quality voice synthesis.
"""

import io
from dataclasses import dataclass
from typing import Optional

import torch
import torchaudio


@dataclass
class SynthesisResult:
    """Result of TTS synthesis."""
    audio_bytes: bytes
    sample_rate: int
    duration_seconds: float


class ChatterboxSynthesizer:
    """
    Synthesizes speech from text using Chatterbox TTS.

    Uses Resemble AI's Chatterbox model which produces natural,
    expressive speech with emotion control.
    """

    def __init__(
        self,
        device: Optional[str] = None,
        use_turbo: bool = False
    ):
        """
        Initialize the synthesizer.

        Args:
            device: Device to run on ('cuda', 'cpu', or None for auto-detect)
            use_turbo: Use Chatterbox-Turbo model (faster, slightly lower quality)
        """
        # Auto-detect device
        if device is None:
            device = "cuda" if torch.cuda.is_available() else "cpu"

        self.device = device
        self.use_turbo = use_turbo

        print(f"Loading Chatterbox TTS model...")
        print(f"Device: {device}, Turbo: {use_turbo}")

        if use_turbo:
            from chatterbox.tts_turbo import ChatterboxTurboTTS
            self.model = ChatterboxTurboTTS.from_pretrained(device=device)
        else:
            from chatterbox.tts import ChatterboxTTS
            self.model = ChatterboxTTS.from_pretrained(device=device)

        self.sample_rate = self.model.sr
        print(f"Chatterbox TTS loaded successfully! Sample rate: {self.sample_rate}")

    def synthesize(
        self,
        text: str,
        exaggeration: float = 0.5,
        cfg_weight: float = 0.5,
        audio_prompt_path: Optional[str] = None
    ) -> SynthesisResult:
        """
        Synthesize speech from text.

        Args:
            text: Text to synthesize
            exaggeration: Emotion exaggeration (0.0 = monotone, 1.0 = very expressive)
            cfg_weight: Classifier-free guidance weight
            audio_prompt_path: Optional path to reference audio for voice cloning

        Returns:
            SynthesisResult with audio bytes, sample rate, and duration
        """
        # Generate audio
        if self.use_turbo:
            # Turbo model has different API
            if audio_prompt_path:
                wav = self.model.generate(text, audio_prompt_path=audio_prompt_path)
            else:
                wav = self.model.generate(text)
        else:
            # Standard model with emotion control
            wav = self.model.generate(
                text,
                exaggeration=exaggeration,
                cfg_weight=cfg_weight
            )

        # Calculate duration
        duration_seconds = wav.shape[1] / self.sample_rate

        # Convert to bytes (WAV format)
        buffer = io.BytesIO()
        torchaudio.save(buffer, wav, self.sample_rate, format="wav")
        audio_bytes = buffer.getvalue()

        return SynthesisResult(
            audio_bytes=audio_bytes,
            sample_rate=self.sample_rate,
            duration_seconds=duration_seconds
        )

    def synthesize_to_mp3(
        self,
        text: str,
        exaggeration: float = 0.5,
        cfg_weight: float = 0.5
    ) -> SynthesisResult:
        """
        Synthesize speech and convert to MP3.

        Args:
            text: Text to synthesize
            exaggeration: Emotion exaggeration (0.0 = monotone, 1.0 = very expressive)
            cfg_weight: Classifier-free guidance weight

        Returns:
            SynthesisResult with MP3 audio bytes
        """
        from pydub import AudioSegment

        # First generate WAV
        result = self.synthesize(text, exaggeration, cfg_weight)

        # Convert to MP3 using pydub
        audio = AudioSegment.from_wav(io.BytesIO(result.audio_bytes))
        mp3_buffer = io.BytesIO()
        audio.export(mp3_buffer, format="mp3", bitrate="128k")
        mp3_bytes = mp3_buffer.getvalue()

        return SynthesisResult(
            audio_bytes=mp3_bytes,
            sample_rate=self.sample_rate,
            duration_seconds=result.duration_seconds
        )
