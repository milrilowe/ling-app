"""
API route handlers for pronunciation analysis.
"""

import time
from typing import Optional

from fastapi import APIRouter
import httpx

import base64

from .schemas import (
    PronunciationRequest,
    PronunciationResponse,
    PronunciationAnalysis,
    PronunciationError,
    PhonemeDetail,
    AudioQuality,
    TranscribeRequest,
    TranscribeResponse,
    SynthesizeRequest,
    SynthesizeResponse,
)
from .audio_fetcher import AudioFetcher
from src.ipa.audio_to_ipa import WhisperIPAConverter
from src.ipa.text_to_ipa import GruutIPAConverter
from src.ipa.aligner import PhonemeAligner
from src.stt.transcriber import FasterWhisperTranscriber
from src.tts.synthesizer import ChatterboxSynthesizer

router = APIRouter()

# Global model instances (loaded once at startup)
whisper_converter: Optional[WhisperIPAConverter] = None
gruut_converter: Optional[GruutIPAConverter] = None
aligner: Optional[PhonemeAligner] = None
audio_fetcher: Optional[AudioFetcher] = None
stt_transcriber: Optional[FasterWhisperTranscriber] = None
tts_synthesizer: Optional[ChatterboxSynthesizer] = None


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
    global whisper_converter, gruut_converter, aligner, audio_fetcher, stt_transcriber, tts_synthesizer

    print("Loading pronunciation analysis models...")

    whisper_converter = WhisperIPAConverter(device=device)
    gruut_converter = GruutIPAConverter(language=language)
    aligner = PhonemeAligner()
    audio_fetcher = AudioFetcher()

    print("Loading STT transcriber (faster-whisper)...")
    stt_transcriber = FasterWhisperTranscriber(model_size="medium", device=device)

    # TTS disabled - using OpenAI TTS API instead (Chatterbox needs ~10GB RAM)
    # print("Loading TTS synthesizer (Chatterbox)...")
    # tts_synthesizer = ChatterboxSynthesizer(device=device)

    print("All models loaded successfully!")


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

        # 4. Align phonemes (normalizer handles tie bars, prosodic markers, etc.)
        # Debug: show extracted phonemes after normalization
        audio_phonemes = aligner.extract_phonemes(audio_ipa)
        expected_phonemes = aligner.extract_phonemes(expected_ipa)
        print(f"[DEBUG] Audio phonemes ({len(audio_phonemes)}): {audio_phonemes}")
        print(f"[DEBUG] Expected phonemes ({len(expected_phonemes)}): {expected_phonemes}")

        alignment = aligner.align(audio_ipa, expected_ipa)

        # Pretty print alignment
        matches = sum(1 for _, _, t in alignment if t == "match")
        print(f"[DEBUG] Alignment ({matches}/{len(alignment)} matches):")
        for exp, act, typ in alignment:
            symbol = "✓" if typ == "match" else "✗" if typ == "substitute" else "−" if typ == "delete" else "+"
            print(f"  {symbol} expected='{exp}' actual='{act}' ({typ})")

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


@router.post("/transcribe", response_model=TranscribeResponse)
async def transcribe(request: TranscribeRequest) -> TranscribeResponse:
    """
    Transcribe audio to text using faster-whisper.

    This endpoint:
    1. Downloads audio from the presigned URL
    2. Transcribes audio to text using faster-whisper
    3. Returns the transcription with language and duration
    """
    if stt_transcriber is None:
        return TranscribeResponse(
            status="error",
            error=PronunciationError(
                code="MODELS_NOT_LOADED",
                message="STT model is not loaded. Server may still be starting.",
                retryable=True
            )
        )

    try:
        # 1. Fetch and load audio
        try:
            audio_array, sample_rate, _ = await audio_fetcher.fetch_and_load(
                request.audio_url,
                apply_vad=False,  # Let faster-whisper handle VAD
                normalize=False
            )
        except httpx.HTTPStatusError as e:
            return TranscribeResponse(
                status="error",
                error=PronunciationError(
                    code="AUDIO_DOWNLOAD_FAILED",
                    message=f"Failed to download audio: HTTP {e.response.status_code}",
                    retryable=True
                )
            )
        except httpx.RequestError as e:
            return TranscribeResponse(
                status="error",
                error=PronunciationError(
                    code="AUDIO_DOWNLOAD_FAILED",
                    message=f"Failed to download audio: {str(e)}",
                    retryable=True
                )
            )

        # 2. Transcribe audio
        result = stt_transcriber.transcribe(
            audio_array,
            sample_rate,
            language=request.language
        )

        return TranscribeResponse(
            status="success",
            text=result.text,
            language=result.language,
            duration=result.duration
        )

    except Exception as e:
        return TranscribeResponse(
            status="error",
            error=PronunciationError(
                code="TRANSCRIPTION_ERROR",
                message=f"Failed to transcribe audio: {str(e)}",
                retryable=True
            )
        )


@router.post("/synthesize", response_model=SynthesizeResponse)
async def synthesize(request: SynthesizeRequest) -> SynthesizeResponse:
    """
    Synthesize speech from text using Chatterbox TTS.

    This endpoint:
    1. Generates audio from the provided text
    2. Returns base64-encoded audio data
    """
    if tts_synthesizer is None:
        return SynthesizeResponse(
            status="error",
            error=PronunciationError(
                code="MODELS_NOT_LOADED",
                message="TTS model is not loaded. Server may still be starting.",
                retryable=True
            )
        )

    try:
        # Generate audio
        if request.format == "mp3":
            result = tts_synthesizer.synthesize_to_mp3(
                request.text,
                exaggeration=request.exaggeration
            )
        else:
            result = tts_synthesizer.synthesize(
                request.text,
                exaggeration=request.exaggeration
            )

        # Encode audio as base64
        audio_base64 = base64.b64encode(result.audio_bytes).decode("utf-8")

        return SynthesizeResponse(
            status="success",
            audio_base64=audio_base64,
            duration=result.duration_seconds,
            format=request.format
        )

    except Exception as e:
        return SynthesizeResponse(
            status="error",
            error=PronunciationError(
                code="SYNTHESIS_ERROR",
                message=f"Failed to synthesize speech: {str(e)}",
                retryable=True
            )
        )
