# API Service

Go backend for the Ling language learning app. Handles authentication, conversation management, and orchestrates the ML service for speech processing.

## Tech Stack

- Go 1.25
- Gin web framework
- PostgreSQL with GORM
- MinIO (S3-compatible storage)

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
│   ├── middleware/       # HTTP middleware (CORS, auth)
│   ├── models/           # Data models (Thread, Message, User)
│   └── services/         # External service clients
│       ├── ml_client.go          # ML service (pronunciation analysis)
│       ├── tts_client.go         # TTS via ML service
│       ├── whisper_client.go     # STT via ML service
│       ├── openai_client.go      # OpenAI for chat responses
│       ├── storage.go            # S3/MinIO storage
│       └── credits.go            # User credits management
└── pkg/                  # Public packages
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL (via `docker-compose up -d` from repo root)
- MinIO (via `docker-compose up -d` from repo root)

### Setup

1. Copy environment file:
   ```bash
   cp example.env .env
   ```

2. Configure `.env` with your API keys (see Environment Variables below)

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

Server runs on http://localhost:8080

## Database

### Migrations

Migrations run automatically on startup via GORM's `AutoMigrate`. When you add or modify models in `internal/models/`, the database schema updates on next server start.

### Connection

Uses PostgreSQL via the `DATABASE_URL` env var. Docker Compose provides a dev database:
```
postgresql://lingapp:lingapp@localhost:5432/lingapp_dev
```

## Storage (S3/MinIO)

Audio files are stored in S3-compatible storage. Docker Compose provides MinIO for development.

Access MinIO console at http://localhost:9001 (user: `minioadmin`, pass: `minioadmin`).

The service auto-creates required buckets on startup.

## External Services

| Service | Purpose | Required |
|---------|---------|----------|
| **ML Service** | TTS, STT, pronunciation analysis | Yes |
| **OpenAI** | Chat responses | Yes |
| **Stripe** | Payments/subscriptions | For billing features |
| **Google/GitHub OAuth** | Social login | For OAuth features |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/threads` | Create new conversation thread |
| GET | `/api/threads/:id` | Get thread with messages |
| POST | `/api/audio/message` | Send audio message to thread |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/register` | Register |
| GET | `/api/user/me` | Get current user |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `ML_SERVICE_URL` | ML service URL | `http://localhost:8000` |
| `OPENAI_API_KEY` | OpenAI API key for chat | - |
| `SESSION_SECRET` | Session encryption key | - |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | `http://localhost:3000` |
| `AWS_*` / `MINIO_*` | S3/MinIO configuration | - |
| `STRIPE_*` | Stripe keys (optional) | - |
| `GOOGLE_*` / `GITHUB_*` | OAuth credentials (optional) | - |

## Development

### No Hot Reload

Go doesn't have built-in hot reload. Restart the server after code changes:
```bash
# Stop with Ctrl+C, then:
go run cmd/server/main.go
```

For auto-restart, use [air](https://github.com/cosmtrek/air):
```bash
go install github.com/cosmtrek/air@latest
air
```

### Testing

No tests yet. Run with:
```bash
go test ./...
```
