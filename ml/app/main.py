from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

from app.config import settings
from app.routers import health

app = FastAPI(
    title="Ling App ML Service",
    description="Machine learning service for pronunciation analysis",
    version="0.1.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_allowed_origins.split(","),
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health.router)

# TODO: Add more routers
# from app.routers import whisper, conversation
# app.include_router(whisper.router, prefix="/api/speech", tags=["speech"])
# app.include_router(conversation.router, prefix="/api/conversation", tags=["conversation"])


if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.environment == "development"
    )
