import type { ConversationState } from '../types'

interface StatusTextProps {
  state: ConversationState
}

const statusMessages: Record<ConversationState, string> = {
  idle: 'Ready to listen',
  recording: 'Listening...',
  'ai-thinking': 'Thinking...',
  'ai-speaking': '',
  archived: '',
}

export function StatusText({ state }: StatusTextProps) {
  const message = statusMessages[state]

  // Always render with fixed height to prevent layout shift
  return (
    <div className="h-7 flex items-center justify-center">
      {message && (
        <p className="text-lg text-muted-foreground animate-pulse">
          {message}
        </p>
      )}
    </div>
  )
}
