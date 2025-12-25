# Web Frontend

React frontend for the Ling language learning app. Features a conversational UI with push-to-talk audio recording and real-time audio playback.

## Tech Stack

- React 19
- Vite 7
- TanStack Router (file-based routing)
- TanStack Query (data fetching)
- Tailwind CSS + shadcn/ui
- TypeScript

## Project Structure

```
src/
├── routes/              # TanStack Router file-based routes
│   ├── __root.tsx       # Root layout
│   ├── index.tsx        # Home page
│   └── c.tsx            # Conversation routes
├── features/            # Feature-based modules
│   ├── conversation/    # Main conversation UI
│   │   ├── components/  # Avatar, PushToTalk, Subtitles
│   │   └── hooks/       # Feature-specific hooks
│   ├── chat/            # Chat components
│   ├── home/            # Home page components
│   └── sidebar/         # Navigation sidebar
├── components/          # Shared components
│   ├── audio/           # AudioPlayer, AudioRecorder
│   ├── layouts/         # Layout components
│   └── ui/              # shadcn/ui components
├── hooks/               # Custom React hooks
│   ├── use-audio-player.ts
│   ├── use-audio-recorder.ts
│   ├── use-thread.ts
│   └── use-streaming.ts
├── contexts/            # React contexts
├── lib/                 # Utilities (api, audio-utils, formatting)
└── integrations/        # Third-party integrations
```

## Getting Started

### Prerequisites

- Node.js 18+
- pnpm

### Setup

1. Install dependencies:
   ```bash
   pnpm install
   ```

2. Copy environment file:
   ```bash
   cp example.env.local .env.local
   ```

3. Start development server:
   ```bash
   pnpm dev
   ```

App runs on http://localhost:3000

## Scripts

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start development server |
| `pnpm build` | Build for production |
| `pnpm test` | Run tests |
| `pnpm lint` | Run ESLint |
| `pnpm format` | Run Prettier |
| `pnpm check` | Format and lint |

## Key Features

- **Push-to-Talk**: Hold spacebar to record audio
- **Audio Playback**: Stream AI responses with playback controls
- **Conversation Threads**: Persistent conversation history
- **Real-time Streaming**: Streaming responses from API
