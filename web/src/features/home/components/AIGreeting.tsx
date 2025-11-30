import { Avatar, AvatarFallback } from '@/components/ui/avatar'

interface AIGreetingProps {
  message: string
}

export function AIGreeting({ message }: AIGreetingProps) {
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 duration-700 flex justify-center">
      <div className="flex max-w-md gap-3">
        <Avatar className="h-10 w-10 shrink-0">
          <AvatarFallback className="bg-bubble-ai text-bubble-ai-foreground text-lg">
            AI
          </AvatarFallback>
        </Avatar>
        <div className="rounded-2xl bg-bubble-ai px-5 py-3 text-bubble-ai-foreground">
          <p className="text-base leading-relaxed">{message}</p>
        </div>
      </div>
    </div>
  )
}
