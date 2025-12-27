"""
MFA alignment wrapper.

Handles downloading audio, running MFA, and parsing results.
"""

import asyncio
import logging
import tempfile
from pathlib import Path

import httpx

from .textgrid_parser import parse_textgrid

logger = logging.getLogger(__name__)


class MFAAligner:
    """
    Wrapper for Montreal Forced Aligner.

    Runs MFA as a subprocess and parses the TextGrid output.
    """

    def __init__(
        self,
        acoustic_model: str = "english_us_arpa",
        dictionary: str = "english_us_arpa",
    ):
        """
        Initialize aligner with model settings.

        Args:
            acoustic_model: Name of MFA acoustic model
            dictionary: Name of MFA pronunciation dictionary
        """
        self.acoustic_model = acoustic_model
        self.dictionary = dictionary

    async def align(
        self,
        audio_url: str,
        transcript: str,
        language: str = "english_us_arpa",
    ) -> dict:
        """
        Align audio to transcript.

        Args:
            audio_url: URL to download audio from
            transcript: Text to align
            language: Language/model name (currently ignored, uses init values)

        Returns:
            Dict with 'words' and 'phones' timing lists, plus 'duration'
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            tmpdir = Path(tmpdir)
            corpus_dir = tmpdir / "corpus"
            output_dir = tmpdir / "output"
            corpus_dir.mkdir()
            output_dir.mkdir()

            # 1. Download audio and convert to WAV
            audio_path = corpus_dir / "audio.wav"
            await self._download_audio(audio_url, audio_path)

            # 2. Write transcript file (must match audio filename)
            txt_path = corpus_dir / "audio.txt"
            txt_path.write_text(transcript)

            logger.info(f"Aligning audio: {len(transcript)} chars, file: {audio_path}")

            # 3. Run MFA alignment
            await self._run_mfa(corpus_dir, output_dir)

            # 4. Parse TextGrid output
            textgrid_path = output_dir / "audio.TextGrid"
            if not textgrid_path.exists():
                # Check for common issues
                log_path = tmpdir / "mfa.log"
                if log_path.exists():
                    logger.error(f"MFA log: {log_path.read_text()}")
                raise RuntimeError("MFA did not produce output TextGrid")

            result = parse_textgrid(str(textgrid_path))
            logger.info(
                f"Alignment complete: {len(result['words'])} words, "
                f"{len(result['phones'])} phones"
            )
            return result

    async def _download_audio(self, url: str, dest: Path):
        """
        Download audio and convert to WAV 16kHz mono.

        MFA requires WAV format, 16kHz sample rate, mono channel.
        """
        logger.info(f"Downloading audio from: {url}")

        async with httpx.AsyncClient(timeout=60.0) as client:
            response = await client.get(url, follow_redirects=True)
            response.raise_for_status()

            # Save original format first
            temp_audio = dest.with_suffix(".tmp")
            temp_audio.write_bytes(response.content)

            logger.info(f"Downloaded {len(response.content)} bytes")

        # Convert to WAV 16kHz mono using ffmpeg
        proc = await asyncio.create_subprocess_exec(
            "ffmpeg",
            "-y",  # Overwrite output
            "-i", str(temp_audio),
            "-ar", "16000",  # 16kHz sample rate
            "-ac", "1",  # Mono
            "-acodec", "pcm_s16le",  # 16-bit PCM
            str(dest),
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await proc.communicate()

        if proc.returncode != 0:
            logger.error(f"ffmpeg failed: {stderr.decode()}")
            raise RuntimeError(f"Audio conversion failed: {stderr.decode()}")

        # Clean up temp file
        temp_audio.unlink(missing_ok=True)
        logger.info(f"Converted audio to WAV: {dest}")

    async def _run_mfa(self, corpus_dir: Path, output_dir: Path):
        """
        Run MFA alignment command.

        Uses subprocess to call the MFA CLI.
        """
        cmd = [
            "mfa",
            "align",
            str(corpus_dir),
            self.dictionary,
            self.acoustic_model,
            str(output_dir),
            "--clean",  # Clean up temp files
            "--single_speaker",  # Optimize for single speaker
            "--quiet",  # Reduce output
        ]

        logger.info(f"Running MFA: {' '.join(cmd)}")

        proc = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await proc.communicate()

        if proc.returncode != 0:
            error_msg = stderr.decode() if stderr else stdout.decode()
            logger.error(f"MFA failed (exit {proc.returncode}): {error_msg}")
            raise RuntimeError(f"MFA alignment failed: {error_msg}")

        logger.info("MFA alignment completed successfully")
