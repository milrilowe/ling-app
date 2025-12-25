"""
API route handlers for pronunciation analysis.
"""

import time
from typing import Optional

from fastapi import APIRouter, HTTPException
import httpx

from .schemas import (
    PronunciationRequest,
    PronunciationResponse,
    PronunciationAnalysis,
    PronunciationError,
    PhonemeDetail,
    AudioQuality,
)
from .audio_fetcher import AudioFetcher
from src.ipa.audio_to_ipa import WhisperIPAConverter
from src.ipa.text_to_ipa import GruutIPAConverter
from src.ipa.aligner import PhonemeAligner

router = APIRouter()

# Global model instances (loaded once at startup)
whisper_converter: Optional[WhisperIPAConverter] = None
gruut_converter: Optional[GruutIPAConverter] = None
aligner: Optional[PhonemeAligner] = None
audio_fetcher: Optional[AudioFetcher] = None


def get_models_loaded() -> bool:
    """Check if models are loaded."""
    return whisper_converter is not None


def load_models(device: Optional[str] = None, language: str = "en-us"):
    """
    Load ML models. Called at startup.

    Args:
        device: Device for Whisper model ('cuda', 'cpu', or None for auto)
        language: Default language for text-to-IPA
    """
    global whisper_converter, gruut_converter, aligner, audio_fetcher

    print("Loading pronunciation analysis models...")

    whisper_converter = WhisperIPAConverter(device=device)
    gruut_converter = GruutIPAConverter(language=language)
    aligner = PhonemeAligner()
    audio_fetcher = AudioFetcher()

    print("Models loaded successfully!")


@router.post("/analyze-pronunciation", response_model=PronunciationResponse)
async def analyze_pronunciation(request: PronunciationRequest) -> PronunciationResponse:
    """
    Analyze pronunciation by comparing audio to expected text.

    This endpoint:
    1. Downloads audio from the presigned URL
    2. Converts audio to IPA using Whisper
    3. Converts expected text to IPA using gruut
    4. Aligns and compares the phonemes
    5. Returns detailed analysis results
    """
    if not get_models_loaded():
        return PronunciationResponse(
            status="error",
            error=PronunciationError(
                code="MODELS_NOT_LOADED",
                message="ML models are not loaded. Server may still be starting.",
                retryable=True
            )
        )

    start_time = time.time()

    try:
        # 1. Fetch and load audio
        try:
            audio_array, sample_rate, quality_report = await audio_fetcher.fetch_and_load(
                request.audio_url,
                apply_vad=True,
                normalize=False
            )
        except httpx.HTTPStatusError as e:
            return PronunciationResponse(
                status="error",
                error=PronunciationError(
                    code="AUDIO_DOWNLOAD_FAILED",
                    message=f"Failed to download audio: HTTP {e.response.status_code}",
                    retryable=True
                )
            )
        except httpx.RequestError as e:
            return PronunciationResponse(
                status="error",
                error=PronunciationError(
                    code="AUDIO_DOWNLOAD_FAILED",
                    message=f"Failed to download audio: {str(e)}",
                    retryable=True
                )
            )

        # 2. Convert audio to IPA
        # Extract language code for Whisper (e.g., 'en' from 'en-us')
        whisper_lang = request.language.split("-")[0] if "-" in request.language else request.language
        print(f"[DEBUG] Audio array shape: {audio_array.shape}, sample_rate: {sample_rate}")
        print(f"[DEBUG] Audio duration: {len(audio_array) / sample_rate:.2f}s")
        print(f"[DEBUG] Expected text: '{request.expected_text}'")

        audio_ipa = whisper_converter.audio_to_ipa(
            audio_array,
            sample_rate,
            language=whisper_lang
        )
        print(f"[DEBUG] Whisper audio_ipa: '{audio_ipa}'")

        # 3. Convert expected text to IPA
        # Create a new converter if language differs from default
        if request.language != gruut_converter.language:
            text_converter = GruutIPAConverter(language=request.language)
        else:
            text_converter = gruut_converter

        expected_ipa = text_converter.text_to_ipa(request.expected_text)
        print(f"[DEBUG] Gruut expected_ipa: '{expected_ipa}'")

        # 4. Align phonemes
        # Debug: show extracted phonemes
        audio_phonemes = aligner.extract_phonemes(audio_ipa)
        expected_phonemes = aligner.extract_phonemes(expected_ipa)
        print(f"[DEBUG] Audio phonemes ({len(audio_phonemes)}): {audio_phonemes}")
        print(f"[DEBUG] Expected phonemes ({len(expected_phonemes)}): {expected_phonemes}")

        alignment = aligner.align(audio_ipa, expected_ipa)
        print(f"[DEBUG] Alignment: {alignment}")

        # 5. Count statistics
        match_count = sum(1 for _, _, t in alignment if t == "match")
        substitution_count = sum(1 for _, _, t in alignment if t == "substitute")
        deletion_count = sum(1 for _, _, t in alignment if t == "delete")
        insertion_count = sum(1 for _, _, t in alignment if t == "insert")
        phoneme_count = len(alignment)

        # 6. Build phoneme details
        phoneme_details = [
            PhonemeDetail(
                expected=expected,
                actual=actual,
                type=match_type,
                position=i
            )
            for i, (expected, actual, match_type) in enumerate(alignment)
        ]

        # 7. Build audio quality report
        audio_quality = AudioQuality(
            quality_score=quality_report.get("quality_score", 100.0),
            snr_db=quality_report.get("snr_db", 0.0),
            duration_seconds=quality_report.get("duration_seconds", 0.0),
            warnings=quality_report.get("warnings", [])
        )

        # Calculate processing time
        processing_time_ms = int((time.time() - start_time) * 1000)

        # Build response
        analysis = PronunciationAnalysis(
            audio_ipa=audio_ipa,
            expected_ipa=expected_ipa,
            phoneme_count=phoneme_count,
            match_count=match_count,
            substitution_count=substitution_count,
            deletion_count=deletion_count,
            insertion_count=insertion_count,
            phoneme_details=phoneme_details,
            audio_quality=audio_quality,
            processing_time_ms=processing_time_ms
        )

        return PronunciationResponse(
            status="success",
            analysis=analysis
        )

    except ValueError as e:
        return PronunciationResponse(
            status="error",
            error=PronunciationError(
                code="PROCESSING_ERROR",
                message=str(e),
                retryable=False
            )
        )
    except Exception as e:
        return PronunciationResponse(
            status="error",
            error=PronunciationError(
                code="INTERNAL_ERROR",
                message=f"Unexpected error during analysis: {str(e)}",
                retryable=True
            )
        )
