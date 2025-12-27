"""
MFA Alignment Service - FastAPI application.

Wraps Montreal Forced Aligner CLI to provide HTTP API for forced alignment.
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routes import router

app = FastAPI(
    title="MFA Alignment Service",
    description="HTTP wrapper for Montreal Forced Aligner",
    version="1.0.0",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router, prefix="/api/v1")


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "service": "mfa"}


@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "MFA Alignment Service",
        "version": "1.0.0",
        "endpoints": {
            "health": "/health",
            "align": "/api/v1/align",
        },
    }
