"""
Pydantic models for API request/response schemas.
"""

from typing import List, Optional
from pydantic import BaseModel, Field


class PronunciationRequest(BaseModel):
    """Request body for pronunciation analysis."""

    audio_url: str = Field(
        ...,
        description="Presigned URL to fetch the audio file from MinIO/S3"
    )
    expected_text: str = Field(
        ...,
        description="The expected text that the user was supposed to say"
    )
    language: str = Field(
        default="en-us",
        description="Language code for phoneme conversion (e.g., 'en-us', 'es', 'fr')"
    )


class PhonemeDetail(BaseModel):
    """Details about a single phoneme alignment."""

    expected: str = Field(..., description="Expected phoneme from text")
    actual: str = Field(..., description="Actual phoneme from audio")
    type: str = Field(
        ...,
        description="Match type: 'match', 'substitute', 'delete', or 'insert'"
    )
    position: int = Field(..., description="Position in the alignment sequence")


class AudioQuality(BaseModel):
    """Audio quality assessment metrics."""

    quality_score: float = Field(..., description="Overall quality score (0-100)")
    snr_db: float = Field(..., description="Signal-to-noise ratio in dB")
    duration_seconds: float = Field(..., description="Audio duration in seconds")
    warnings: List[str] = Field(
        default_factory=list,
        description="List of quality warnings"
    )


class PronunciationAnalysis(BaseModel):
    """Detailed pronunciation analysis results."""

    audio_ipa: str = Field(..., description="IPA transcription from audio")
    expected_ipa: str = Field(..., description="IPA transcription from expected text")
    phoneme_count: int = Field(..., description="Total number of phonemes analyzed")
    match_count: int = Field(..., description="Number of matching phonemes")
    substitution_count: int = Field(..., description="Number of substituted phonemes")
    deletion_count: int = Field(..., description="Number of deleted (not said) phonemes")
    insertion_count: int = Field(..., description="Number of inserted (extra) phonemes")
    phoneme_details: List[PhonemeDetail] = Field(
        default_factory=list,
        description="Per-phoneme alignment details"
    )
    audio_quality: Optional[AudioQuality] = Field(
        default=None,
        description="Audio quality metrics"
    )
    processing_time_ms: int = Field(..., description="Processing time in milliseconds")


class PronunciationError(BaseModel):
    """Error details when analysis fails."""

    code: str = Field(..., description="Error code")
    message: str = Field(..., description="Human-readable error message")
    retryable: bool = Field(
        default=False,
        description="Whether the request can be retried"
    )


class PronunciationResponse(BaseModel):
    """Response body for pronunciation analysis."""

    status: str = Field(
        ...,
        description="Status of the analysis: 'success' or 'error'"
    )
    analysis: Optional[PronunciationAnalysis] = Field(
        default=None,
        description="Analysis results (present when status is 'success')"
    )
    error: Optional[PronunciationError] = Field(
        default=None,
        description="Error details (present when status is 'error')"
    )


class HealthResponse(BaseModel):
    """Health check response."""

    status: str = Field(default="healthy")
    model_loaded: bool = Field(default=False)
