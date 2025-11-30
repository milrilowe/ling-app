from elevenlabs import ElevenLabs
from app.config import settings


class ElevenLabsService:
    """Service for text-to-speech using ElevenLabs."""

    def __init__(self):
        self.client = ElevenLabs(api_key=settings.elevenlabs_api_key) if settings.elevenlabs_api_key else None

    async def text_to_speech(self, text: str, voice_id: str = "21m00Tcm4TlvDq8ikWAM") -> bytes:
        """
        Convert text to speech audio.

        Args:
            text: Text to convert
            voice_id: ElevenLabs voice ID (default is Rachel voice)

        Returns:
            Audio bytes
        """
        if not self.client:
            raise ValueError("ElevenLabs API key not configured")

        # TODO: Implement TTS
        # This is a placeholder
        audio = self.client.generate(
            text=text,
            voice=voice_id,
            model="eleven_multilingual_v2"
        )
        return audio


# Singleton instance
elevenlabs_service = ElevenLabsService()
