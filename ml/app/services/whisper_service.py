import whisper
from app.config import settings


class WhisperService:
    """Service for speech-to-text using Whisper."""

    def __init__(self):
        self.model = None
        self.model_name = settings.whisper_model

    def load_model(self):
        """Load the Whisper model."""
        if self.model is None:
            self.model = whisper.load_model(self.model_name)
        return self.model

    async def transcribe(self, audio_path: str) -> dict:
        """
        Transcribe audio file to text.

        Args:
            audio_path: Path to audio file

        Returns:
            Dictionary with transcription results
        """
        model = self.load_model()
        result = model.transcribe(audio_path)
        return result


# Singleton instance
whisper_service = WhisperService()
