import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { useStreaming } from '@/hooks/use-streaming'
import { cn } from '@/lib/utils'

interface StreamingMessageProps {
  content: string
  speed?: number
}

export function StreamingMessage({
  content,
  speed = 40,
}: StreamingMessageProps) {
  const { displayedText, isStreaming, skipStreaming } = useStreaming(content, {
    speed,
    skipOnClick: true,
  })

  return (
    <div className="mb-4 flex animate-in fade-in slide-in-from-bottom-2 duration-300 justify-start">
      <div className="flex max-w-[80%] gap-3">
        {/* Avatar */}
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback className="bg-bubble-ai text-bubble-ai-foreground">
            AI
          </AvatarFallback>
        </Avatar>

        {/* Message bubble */}
        <div className="flex flex-col gap-1">
          <div
            className={cn(
              'rounded-2xl px-4 py-2 bg-bubble-ai text-bubble-ai-foreground',
              isStreaming && 'cursor-pointer'
            )}
            onClick={skipStreaming}
            role={isStreaming ? 'button' : undefined}
            tabIndex={isStreaming ? 0 : undefined}
            onKeyDown={(e) => {
              if (isStreaming && (e.key === 'Enter' || e.key === ' ')) {
                skipStreaming()
              }
            }}
          >
            <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
              {displayedText}
              {isStreaming && (
                <span className="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-current" />
              )}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
