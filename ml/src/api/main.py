"""
FastAPI application for pronunciation analysis.

Run with:
    uvicorn src.api.main:app --host 0.0.0.0 --port 8000
"""

import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routes import router, load_models, get_models_loaded
from .schemas import HealthResponse


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Lifespan context manager for startup/shutdown events.
    """
    # Startup: Load models
    device = os.getenv("ML_DEVICE", None)  # None = auto-detect
    language = os.getenv("ML_DEFAULT_LANGUAGE", "en-us")

    load_models(device=device, language=language)

    yield

    # Shutdown: Cleanup (if needed)
    print("Shutting down pronunciation analysis API...")


app = FastAPI(
    title="Pronunciation Analysis API",
    description="Analyzes pronunciation by comparing audio to expected text using IPA phoneme alignment",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=os.getenv("CORS_ORIGINS", "*").split(","),
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router, prefix="/api/v1")


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    return HealthResponse(
        status="healthy" if get_models_loaded() else "starting",
        model_loaded=get_models_loaded()
    )


@app.get("/")
async def root():
    """Root endpoint with API info."""
    return {
        "name": "Pronunciation Analysis API",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health"
    }
