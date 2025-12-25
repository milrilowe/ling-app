"""
Audio fetcher for downloading audio from presigned URLs.
"""

import os
import tempfile
from typing import Tuple

import httpx
import numpy as np

from src.audio.loader import AudioLoader


class AudioFetcher:
    """Fetches and processes audio from presigned URLs."""

    def __init__(self, timeout: float = 60.0):
        """
        Initialize audio fetcher.

        Args:
            timeout: HTTP request timeout in seconds
        """
        self.timeout = timeout
        self.loader = AudioLoader()

    async def fetch_and_load(
        self,
        audio_url: str,
        apply_vad: bool = True,
        normalize: bool = False
    ) -> Tuple[np.ndarray, int, dict]:
        """
        Fetch audio from URL and load it for processing.

        Args:
            audio_url: Presigned URL to fetch audio from
            apply_vad: Whether to apply voice activity detection
            normalize: Whether to normalize audio volume

        Returns:
            Tuple of (audio_array, sample_rate, quality_report)

        Raises:
            httpx.HTTPError: If download fails
            RuntimeError: If audio processing fails
        """
        # Download audio to temp file
        temp_path = None
        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.get(audio_url)
                response.raise_for_status()

                # Determine file extension from content-type or URL
                content_type = response.headers.get("content-type", "")
                if "webm" in content_type or audio_url.endswith(".webm"):
                    ext = ".webm"
                elif "wav" in content_type or audio_url.endswith(".wav"):
                    ext = ".wav"
                elif "mp3" in content_type or audio_url.endswith(".mp3"):
                    ext = ".mp3"
                else:
                    # Default to webm since that's what the app uses
                    ext = ".webm"

                # Write to temp file
                with tempfile.NamedTemporaryFile(
                    suffix=ext,
                    delete=False
                ) as tmp:
                    tmp.write(response.content)
                    temp_path = tmp.name

            # Load and process audio
            audio_array, sample_rate, quality_report = self.loader.load_audio_with_quality_check(
                temp_path,
                apply_vad=apply_vad,
                normalize=normalize,
                warn_on_quality=False  # We'll handle warnings in the response
            )

            return audio_array, sample_rate, quality_report

        finally:
            # Clean up temp file
            if temp_path and os.path.exists(temp_path):
                try:
                    os.remove(temp_path)
                except Exception:
                    pass  # Ignore cleanup errors


# Convenience function
async def fetch_audio(
    audio_url: str,
    apply_vad: bool = True,
    normalize: bool = False
) -> Tuple[np.ndarray, int, dict]:
    """
    Convenience function to fetch and load audio from URL.

    Args:
        audio_url: Presigned URL to fetch audio from
        apply_vad: Whether to apply voice activity detection
        normalize: Whether to normalize audio volume

    Returns:
        Tuple of (audio_array, sample_rate, quality_report)
    """
    fetcher = AudioFetcher()
    return await fetcher.fetch_and_load(audio_url, apply_vad, normalize)
