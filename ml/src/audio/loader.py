"""
Audio loading and conversion utilities.

Handles WebM files and converts them to format suitable for Whisper model.
"""

import os
import tempfile
import warnings
from pathlib import Path
from typing import Dict, Tuple

import librosa
import numpy as np
import soundfile as sf
from pydub import AudioSegment


class AudioLoader:
    """Handles audio file loading and conversion to Whisper-compatible format."""

    WHISPER_SAMPLE_RATE = 16000  # Whisper requires 16kHz

    def __init__(self, temp_dir: str = None):
        """
        Initialize AudioLoader.

        Args:
            temp_dir: Directory for temporary file conversions.
                     If None, uses system temp directory.
        """
        self.temp_dir = temp_dir or tempfile.gettempdir()
        os.makedirs(self.temp_dir, exist_ok=True)

    def load_audio(
        self,
        file_path: str,
        apply_vad: bool = True,
        normalize: bool = False,
        vad_top_db: int = 30,
        reduce_noise: bool = False,
        apply_preemphasis: bool = False
    ) -> Tuple[np.ndarray, int]:
        """
        Load audio file, converting from WebM if needed.

        Args:
            file_path: Path to audio file (WebM, WAV, MP3, etc.)
            apply_vad: If True, trim silence from beginning and end (Voice Activity Detection)
            normalize: If True, normalize audio volume
            vad_top_db: Threshold for VAD in dB (default: 20)
            reduce_noise: If True, apply noise reduction (requires noisereduce package)
            apply_preemphasis: If True, apply pre-emphasis filter to boost high frequencies

        Returns:
            Tuple of (audio_array, sample_rate)
            - audio_array: numpy array of audio samples (mono)
            - sample_rate: sampling rate (16000 for Whisper)

        Raises:
            FileNotFoundError: If audio file doesn't exist
            ValueError: If audio file format is unsupported
        """
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"Audio file not found: {file_path}")

        file_path = Path(file_path)
        file_ext = file_path.suffix.lower()

        # Convert WebM to WAV first (librosa doesn't natively support WebM)
        if file_ext == '.webm':
            wav_path = self._convert_webm_to_wav(str(file_path))
            audio_array, sample_rate = self._load_wav(wav_path)
            # Clean up temp file
            try:
                os.remove(wav_path)
            except Exception:
                pass  # Ignore cleanup errors
        else:
            # Load directly with librosa
            audio_array, sample_rate = librosa.load(
                str(file_path),
                sr=self.WHISPER_SAMPLE_RATE,
                mono=True
            )

        # Apply noise reduction (optional, requires noisereduce package)
        if reduce_noise:
            try:
                import noisereduce as nr
                audio_array = nr.reduce_noise(y=audio_array, sr=sample_rate)
            except ImportError:
                warnings.warn(
                    "noisereduce package not installed. "
                    "Install with: pip install noisereduce"
                )

        # Apply VAD (Voice Activity Detection) - trim silence
        if apply_vad:
            audio_array, _ = librosa.effects.trim(
                audio_array,
                top_db=vad_top_db,
                frame_length=2048,
                hop_length=512
            )

        # Apply pre-emphasis filter (boost high frequencies)
        if apply_preemphasis:
            audio_array = librosa.effects.preemphasis(audio_array, coef=0.97)

        # Normalize audio volume (peak normalization, not L2)
        if normalize:
            max_val = np.max(np.abs(audio_array))
            if max_val > 0:
                audio_array = audio_array / max_val

        return audio_array, sample_rate

    def assess_audio_quality(
        self,
        audio_array: np.ndarray,
        sample_rate: int
    ) -> Dict[str, any]:
        """
        Assess audio quality and detect potential issues.

        Args:
            audio_array: Audio samples as numpy array
            sample_rate: Sample rate of audio

        Returns:
            Dictionary containing:
                - 'snr_db': Signal-to-noise ratio estimate in dB
                - 'clipping_ratio': Ratio of clipped samples (0-1)
                - 'silence_ratio': Ratio of silent frames (0-1)
                - 'duration_seconds': Duration in seconds
                - 'warnings': List of warning messages
                - 'quality_score': Overall quality score (0-100)
        """
        warnings_list = []

        # Calculate duration
        duration = len(audio_array) / sample_rate

        # Check for clipping (samples at or near ±1.0)
        clipping_threshold = 0.99
        clipped_samples = np.sum(np.abs(audio_array) >= clipping_threshold)
        clipping_ratio = clipped_samples / len(audio_array)

        if clipping_ratio > 0.01:  # More than 1% clipped
            warnings_list.append(
                f"Audio clipping detected ({clipping_ratio*100:.1f}% of samples). "
                "This may degrade transcription quality."
            )

        # Estimate SNR (simple method: ratio of signal power to noise floor)
        # Split into frames and find noise floor
        frame_length = 2048
        hop_length = 512
        frames = librosa.util.frame(audio_array, frame_length=frame_length, hop_length=hop_length)
        frame_powers = np.mean(frames ** 2, axis=0)

        # Assume bottom 10% of frames are noise
        noise_floor = np.percentile(frame_powers, 10)
        signal_power = np.mean(frame_powers)

        if noise_floor > 0:
            snr_db = 10 * np.log10(signal_power / noise_floor)
        else:
            snr_db = float('inf')

        if snr_db < 20:
            warnings_list.append(
                f"Low signal-to-noise ratio ({snr_db:.1f} dB). "
                "Background noise may affect accuracy."
            )

        # Calculate silence ratio
        silence_threshold = 0.01
        silent_frames = np.sum(frame_powers < silence_threshold)
        silence_ratio = silent_frames / len(frame_powers) if len(frame_powers) > 0 else 0

        if silence_ratio > 0.5:
            warnings_list.append(
                f"High silence ratio ({silence_ratio*100:.1f}%). "
                "Audio may be too quiet or mostly silent."
            )

        # Check duration
        if duration < 0.1:
            warnings_list.append("Audio is very short (<0.1s). May not contain enough speech.")
        elif duration > 30:
            warnings_list.append(
                f"Audio is long ({duration:.1f}s). Consider chunking for better accuracy."
            )

        # Calculate overall quality score (0-100)
        quality_score = 100.0

        # Penalize for clipping
        quality_score -= min(clipping_ratio * 100, 30)

        # Penalize for low SNR
        if snr_db < 30:
            quality_score -= (30 - snr_db)

        # Penalize for high silence
        if silence_ratio > 0.3:
            quality_score -= (silence_ratio - 0.3) * 50

        quality_score = max(0, min(100, quality_score))

        return {
            'snr_db': snr_db,
            'clipping_ratio': clipping_ratio,
            'silence_ratio': silence_ratio,
            'duration_seconds': duration,
            'warnings': warnings_list,
            'quality_score': quality_score
        }

    def load_audio_with_quality_check(
        self,
        file_path: str,
        apply_vad: bool = True,
        normalize: bool = False,
        vad_top_db: int = 30,
        reduce_noise: bool = False,
        apply_preemphasis: bool = False,
        warn_on_quality: bool = True
    ) -> Tuple[np.ndarray, int, Dict]:
        """
        Load audio and assess its quality.

        Args:
            file_path: Path to audio file
            apply_vad: If True, trim silence from beginning and end
            normalize: If True, normalize audio volume
            vad_top_db: Threshold for VAD in dB
            reduce_noise: If True, apply noise reduction
            apply_preemphasis: If True, apply pre-emphasis filter
            warn_on_quality: If True, print warnings for quality issues

        Returns:
            Tuple of (audio_array, sample_rate, quality_report)
        """
        audio_array, sample_rate = self.load_audio(
            file_path, apply_vad, normalize, vad_top_db,
            reduce_noise, apply_preemphasis
        )

        quality_report = self.assess_audio_quality(audio_array, sample_rate)

        if warn_on_quality and quality_report['warnings']:
            for warning in quality_report['warnings']:
                warnings.warn(warning)

        return audio_array, sample_rate, quality_report

    def _convert_webm_to_wav(self, webm_path: str) -> str:
        """
        Convert WebM audio to WAV format using pydub + FFmpeg.

        Args:
            webm_path: Path to WebM file

        Returns:
            Path to converted WAV file

        Raises:
            RuntimeError: If conversion fails (FFmpeg not installed, etc.)
        """
        try:
            # Load WebM with pydub (uses FFmpeg backend)
            audio = AudioSegment.from_file(webm_path, format="webm")

            # Convert to mono and set sample rate
            audio = audio.set_channels(1)  # Mono
            audio = audio.set_frame_rate(self.WHISPER_SAMPLE_RATE)  # 16kHz

            # Create temp WAV file
            temp_wav = os.path.join(
                self.temp_dir,
                f"temp_{os.path.basename(webm_path)}.wav"
            )

            # Export as WAV (16-bit PCM)
            audio.export(
                temp_wav,
                format="wav",
                parameters=["-acodec", "pcm_s16le"]
            )

            return temp_wav

        except Exception as e:
            raise RuntimeError(
                f"Failed to convert WebM to WAV. "
                f"Is FFmpeg installed? Error: {str(e)}"
            ) from e

    def _load_wav(self, wav_path: str) -> Tuple[np.ndarray, int]:
        """
        Load WAV file with librosa.

        Args:
            wav_path: Path to WAV file

        Returns:
            Tuple of (audio_array, sample_rate)
        """
        audio_array, sample_rate = librosa.load(
            wav_path,
            sr=self.WHISPER_SAMPLE_RATE,
            mono=True
        )
        return audio_array, sample_rate


def load_audio_file(
    file_path: str,
    apply_vad: bool = True,
    normalize: bool = False,
    vad_top_db: int = 30,
    reduce_noise: bool = False,
    apply_preemphasis: bool = False
) -> Tuple[np.ndarray, int]:
    """
    Convenience function to load an audio file.

    Args:
        file_path: Path to audio file
        apply_vad: If True, trim silence from beginning and end
        normalize: If True, normalize audio volume
        vad_top_db: Threshold for VAD in dB (default: 20)
        reduce_noise: If True, apply noise reduction
        apply_preemphasis: If True, apply pre-emphasis filter

    Returns:
        Tuple of (audio_array, sample_rate)
    """
    loader = AudioLoader()
    return loader.load_audio(
        file_path, apply_vad, normalize, vad_top_db,
        reduce_noise, apply_preemphasis
    )
