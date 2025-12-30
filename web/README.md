# Web Frontend

React frontend for the Ling language learning app. Features a conversational UI with push-to-talk audio recording and real-time audio playback.

## Tech Stack

- React 19
- Vite 7
- TypeScript
- TanStack Router (file-based routing)
- TanStack Query (data fetching)
- Tailwind CSS + shadcn/ui

## Project Structure

```
src/
├── routes/              # TanStack Router file-based routes
│   ├── __root.tsx       # Root layout with providers
│   ├── index.tsx        # Home page (/)
│   ├── c.tsx            # Conversation layout (/c)
│   └── c.$threadId.tsx  # Conversation page (/c/:threadId)
├── features/            # Feature-based modules
│   ├── conversation/    # Main conversation UI
│   │   ├── components/  # Avatar, PushToTalk, Subtitles
│   │   └── hooks/       # Feature-specific hooks
│   ├── chat/            # Chat message components
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
├── lib/                 # Utilities
│   ├── api.ts           # API client
│   ├── audio-utils.ts   # Audio processing helpers
│   └── utils.ts         # General utilities
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

## Routing

Uses [TanStack Router](https://tanstack.com/router) with file-based routing.

### Adding a New Route

1. Create a file in `src/routes/`:
   - `foo.tsx` → `/foo`
   - `foo.$id.tsx` → `/foo/:id` (dynamic segment)
   - `foo.index.tsx` → `/foo` (index route)
   - `foo.lazy.tsx` → lazy-loaded route

2. Export a `Route` from the file:
   ```tsx
   import { createFileRoute } from '@tanstack/react-router'

   export const Route = createFileRoute('/foo')({
     component: FooPage,
   })

   function FooPage() {
     return <div>Foo</div>
   }
   ```

3. Run `pnpm dev` - routes auto-generate in `routeTree.gen.ts`

## Data Fetching

Uses [TanStack Query](https://tanstack.com/query) for server state.

### Pattern

```tsx
import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@/lib/api'

// Fetching data
const { data, isLoading } = useQuery({
  queryKey: ['threads', threadId],
  queryFn: () => api.getThread(threadId),
})

// Mutations
const mutation = useMutation({
  mutationFn: api.createThread,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['threads'] })
  },
})
```

## Audio Recording

Push-to-talk audio recording using the Web Audio API.

### How It Works

1. User holds spacebar (or clicks record button)
2. `use-audio-recorder` hook captures audio via MediaRecorder
3. Audio is encoded as WebM
4. On release, audio is sent to API
5. API returns transcription and AI response with audio

### Key Files

- `src/hooks/use-audio-recorder.ts` - Recording logic
- `src/hooks/use-audio-player.ts` - Playback logic
- `src/features/conversation/components/PushToTalk.tsx` - UI component

## Styling

### Tailwind CSS

Utility-first CSS. Configure in `tailwind.config.js`.

### shadcn/ui

Pre-built components in `src/components/ui/`. Add new components:
```bash
pnpm dlx shadcn@latest add button
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `VITE_API_URL` | API server URL (default: `http://localhost:8080`) |

## Testing

No tests yet. Run with:
```bash
pnpm test
```
