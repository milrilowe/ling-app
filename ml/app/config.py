from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Server
    port: int = 8000
    host: str = "0.0.0.0"

    # OpenAI
    openai_api_key: str = ""

    # ElevenLabs
    elevenlabs_api_key: str = ""

    # Whisper
    whisper_model: str = "base"

    # CORS
    cors_allowed_origins: str = "http://localhost:8080,http://localhost:3000"

    # Environment
    environment: str = "development"

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )


settings = Settings()
