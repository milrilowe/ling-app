import { Avatar, AvatarFallback } from '@/components/ui/avatar'

export function TypingIndicator() {
  return (
    <div className="mb-4 flex animate-in fade-in duration-300 justify-start">
      <div className="flex gap-3">
        {/* Avatar */}
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback className="bg-bubble-ai text-bubble-ai-foreground">
            AI
          </AvatarFallback>
        </Avatar>

        {/* Typing dots */}
        <div className="flex items-center gap-1 rounded-2xl bg-bubble-ai px-4 py-3">
          <span className="h-2 w-2 animate-bounce rounded-full bg-bubble-ai-foreground [animation-delay:-0.3s]" />
          <span className="h-2 w-2 animate-bounce rounded-full bg-bubble-ai-foreground [animation-delay:-0.15s]" />
          <span className="h-2 w-2 animate-bounce rounded-full bg-bubble-ai-foreground" />
        </div>
      </div>
    </div>
  )
}
