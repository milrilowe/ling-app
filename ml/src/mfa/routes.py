"""
MFA API routes.
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from .aligner import MFAAligner

router = APIRouter()

# Global aligner instance (reused across requests)
aligner = MFAAligner()


class AlignRequest(BaseModel):
    """Request body for alignment."""

    audio_url: str = Field(..., description="URL to download audio file from")
    transcript: str = Field(..., description="Text transcript to align")
    language: str = Field(
        default="english_us_arpa",
        description="MFA language/model name",
    )


class WordTiming(BaseModel):
    """Word-level timing information."""

    word: str
    start: float
    end: float


class PhoneTiming(BaseModel):
    """Phone-level timing information with dual format."""

    arpabet: str = Field(..., description="ARPAbet symbol (MFA native)")
    ipa: str = Field(..., description="IPA symbol (for display)")
    start: float
    end: float


class AlignResponse(BaseModel):
    """Response from alignment."""

    words: list[WordTiming]
    phones: list[PhoneTiming]
    duration: float


@router.post("/align", response_model=AlignResponse)
async def align_audio(request: AlignRequest):
    """
    Align audio to transcript using Montreal Forced Aligner.

    Returns word and phoneme-level timing information.
    Phonemes are returned in both ARPAbet (MFA native) and IPA formats.
    """
    import traceback
    print(f"[MFA] Align request: url={request.audio_url[:100]}..., transcript={request.transcript[:50]}...")
    try:
        result = await aligner.align(
            audio_url=request.audio_url,
            transcript=request.transcript,
            language=request.language,
        )
        print(f"[MFA] Alignment complete: {len(result['words'])} words")
        return result
    except FileNotFoundError as e:
        print(f"[MFA] File error: {e}")
        raise HTTPException(status_code=400, detail=f"Audio file error: {e}")
    except RuntimeError as e:
        print(f"[MFA] Runtime error: {e}")
        raise HTTPException(status_code=500, detail=f"MFA error: {e}")
    except Exception as e:
        print(f"[MFA] Exception: {e}")
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=f"Alignment failed: {e}")
