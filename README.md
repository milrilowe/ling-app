# Ling App

A conversational language learning app where users practice speaking through realistic dialogues. The system analyzes pronunciation at the phoneme level and provides intelligent feedback.

## Quick Start

### Prerequisites

- Node.js 18+ and pnpm
- Go 1.25+
- Python 3.11+ (3.12 not fully tested with gruut)
- Docker (for dev infrastructure)

### 1. Start Infrastructure

Docker Compose runs PostgreSQL and MinIO (S3-compatible storage) for development:

```bash
docker-compose up -d
```

### 2. Environment Setup

```bash
cp web/example.env.local web/.env.local
cp api/example.env api/.env
cp ml/example.env ml/.env
```

Fill in your API keys (see each package's README for details).

### 3. Run Services

Run each service in a separate terminal:

**Terminal 1 - Web** (http://localhost:3000)

```bash
cd web
pnpm install
pnpm dev
```

**Terminal 2 - API** (http://localhost:8080)

```bash
cd api
go run cmd/server/main.go
```

**Terminal 3 - ML** (http://localhost:8000)

```bash
cd ml
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt
pip install gruut[en]
uvicorn src.api.main:app --reload
```

## Architecture

```
Web (React + Vite) → API (Go) → ML (Python/FastAPI)
                         ↓
                    PostgreSQL
                      MinIO
                  3rd-party APIs
```

- **Web**: Frontend with push-to-talk audio recording and conversation UI
- **API**: Go backend handling auth, conversations, and orchestrating services
- **ML**: Python service providing TTS, STT, and pronunciation analysis

## Project Structure

```
ling-app/
├── web/                    # React frontend (Vite, TypeScript, TanStack Router)
├── api/                    # Go backend (Gin, GORM, PostgreSQL)
├── ml/                     # Python ML service (FastAPI, Whisper, gruut)
├── scripts/                # Utility scripts
└── docker-compose.yml      # Dev infrastructure (PostgreSQL, MinIO)
```

## Service Documentation

- [Web README](web/README.md) - React frontend
- [API README](api/README.md) - Go backend
- [ML README](ml/README.md) - Python speech service

## License

TBD
