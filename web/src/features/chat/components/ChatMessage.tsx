import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { formatMessageTime } from '@/lib/formatting'
import { cn } from '@/lib/utils'

interface ChatMessageProps {
  role: 'user' | 'assistant'
  content: string
  timestamp: string | Date
}

export function ChatMessage({ role, content, timestamp }: ChatMessageProps) {
  const isUser = role === 'user'

  return (
    <div
      className={cn(
        'mb-4 flex animate-in fade-in slide-in-from-bottom-2 duration-300',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      <div
        className={cn(
          'flex max-w-[80%] gap-3',
          isUser ? 'flex-row-reverse' : 'flex-row'
        )}
      >
        {/* Avatar */}
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback
            className={cn(
              isUser
                ? 'bg-bubble-user text-bubble-user-foreground'
                : 'bg-bubble-ai text-bubble-ai-foreground'
            )}
          >
            {isUser ? 'Y' : 'AI'}
          </AvatarFallback>
        </Avatar>

        {/* Message bubble */}
        <div className="flex flex-col gap-1">
          <div
            className={cn(
              'rounded-2xl px-4 py-2',
              isUser
                ? 'bg-bubble-user text-bubble-user-foreground'
                : 'bg-bubble-ai text-bubble-ai-foreground'
            )}
          >
            <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
              {content}
            </p>
          </div>

          {/* Timestamp - shown on hover */}
          <span
            className={cn(
              'text-xs text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100',
              isUser ? 'text-right' : 'text-left'
            )}
          >
            {formatMessageTime(timestamp)}
          </span>
        </div>
      </div>
    </div>
  )
}
