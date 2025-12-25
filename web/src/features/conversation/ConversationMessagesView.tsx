import { useThread } from '@/hooks/use-thread'
import { MessageBubble } from '@/features/conversation/components/MessageBubble'
import { useAutoScroll } from '@/hooks/use-auto-scroll'

interface ConversationMessagesViewProps {
  threadId: string
}

export function ConversationMessagesView({ threadId }: ConversationMessagesViewProps) {
  const { data: thread, isLoading, isError } = useThread(threadId)

  const scrollRef = useAutoScroll<HTMLDivElement>([thread?.messages.length])

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground">Loading conversation...</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-destructive">Failed to load conversation.</p>
      </div>
    )
  }

  if (!thread?.messages.length) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground">No messages yet. Start a conversation using the audio view!</p>
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col bg-background">
      {/* Messages Area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4">
        {thread?.messages.map((message) => (
          <MessageBubble
            key={message.id}
            role={message.role as 'user' | 'assistant'}
            content={message.content}
            timestamp={message.timestamp}
            audioUrl={message.audioUrl}
            hasAudio={message.hasAudio}
            pronunciationStatus={message.pronunciationStatus}
            pronunciationAnalysis={message.pronunciationAnalysis}
            pronunciationError={message.pronunciationError}
          />
        ))}
      </div>
    </div>
  )
}
