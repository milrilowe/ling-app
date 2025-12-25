# API Service

Go backend service for the Ling language learning app. Handles conversation management, audio processing, and integrations with ML service, OpenAI, and ElevenLabs.

## Tech Stack

- Go 1.25
- Gin web framework
- PostgreSQL database
- GORM (ORM)

## Project Structure

```
api/
├── cmd/server/           # Entry point
│   └── main.go
├── internal/
│   ├── config/           # Configuration management
│   ├── db/               # Database connection and migrations
│   ├── handlers/         # HTTP request handlers
│   │   ├── audio.go      # Audio message handling
│   │   ├── health.go     # Health check
│   │   ├── prompt.go     # Prompt generation
│   │   └── thread.go     # Conversation threads
│   ├── middleware/       # HTTP middleware (CORS)
│   ├── models/           # Data models (Thread, Message)
│   └── services/         # External service clients
│       ├── ml_client.go          # ML service integration
│       ├── openai_client.go      # OpenAI API client
│       ├── elevenlabs_client.go  # ElevenLabs TTS
│       ├── whisper_client.go     # Whisper STT
│       └── pronunciation_worker.go
└── pkg/                  # Public packages
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL (or use `docker-compose up -d` from root)

### Setup

1. Copy environment file:
   ```bash
   cp example.env .env
   ```

2. Configure `.env` with your API keys

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

Server runs on http://localhost:8080

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/threads` | Create new conversation thread |
| GET | `/api/threads/:id` | Get thread with messages |
| POST | `/api/audio/message` | Send audio message to thread |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `ML_SERVICE_URL` | ML service URL | `http://localhost:8000` |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `SESSION_SECRET` | Session encryption key | - |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | `http://localhost:3000` |
