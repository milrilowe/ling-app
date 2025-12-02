import { useParams } from '@tanstack/react-router'
import { ConversationUI } from '@/features/conversation/ConversationUI'
import { ConversationMessagesView } from '@/features/conversation/ConversationMessagesView'
import { ViewToggle } from '@/features/conversation/components/ViewToggle'
import { useConversationView } from '@/hooks/use-conversation-view'

export function ConversationPage() {
  const { threadId } = useParams({ from: '/c/$threadId' })
  const { viewMode, toggleView } = useConversationView()

  return (
    <div className="relative h-full">
      {/* Only show toggle if we have a threadId (not on initial conversation screen) */}
      {threadId && <ViewToggle currentView={viewMode} onToggle={toggleView} />}

      {viewMode === 'audio' ? (
        <ConversationUI threadId={threadId} />
      ) : (
        threadId ? (
          <ConversationMessagesView threadId={threadId} />
        ) : (
          <div className="flex h-full items-center justify-center">
            <p className="text-muted-foreground">Start a conversation to view messages</p>
          </div>
        )
      )}
    </div>
  )
}
