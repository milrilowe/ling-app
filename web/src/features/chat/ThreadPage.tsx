import { useSendMessage, useThread } from '@/hooks/use-thread'
import { useSendAudioMessage } from '@/hooks/use-send-audio-message'
import { useParams } from '@tanstack/react-router'
import { ChatMessage } from './components/ChatMessage'
import { ChatInput } from './components/ChatInput'
import { TypingIndicator } from './components/TypingIndicator'
import { StreamingMessage } from './components/StreamingMessage'
import { useAutoScroll } from '@/hooks/use-auto-scroll'
import { useState } from 'react'

export function ChatThread() {
  const { threadId } = useParams({ from: '/c/$threadId' })

  const { data: thread, isLoading, isError } = useThread(threadId)
  const sendMessageMutation = useSendMessage(threadId)
  const sendAudioMutation = useSendAudioMessage(threadId)
  const [latestMessageId, setLatestMessageId] = useState<string | null>(null)
  const [pendingMessage, setPendingMessage] = useState<string | null>(null)

  const isAnyPending = sendMessageMutation.isPending || sendAudioMutation.isPending

  const scrollRef = useAutoScroll<HTMLDivElement>([
    thread?.messages.length,
    isAnyPending,
    pendingMessage,
  ])

  const handleSendMessage = (message: string) => {
    setPendingMessage(message)
    sendMessageMutation.mutate(message, {
      onSuccess: (newMessage) => {
        setLatestMessageId(newMessage.id)
        setPendingMessage(null)
      },
      onError: () => {
        setPendingMessage(null)
      },
    })
  }

  const handleSendAudio = (audioBlob: Blob) => {
    setPendingMessage('🎤 Sending voice message...')
    sendAudioMutation.mutate(audioBlob, {
      onSuccess: (data) => {
        setLatestMessageId(data.assistantMessage.id)
        setPendingMessage(null)
      },
      onError: () => {
        setPendingMessage(null)
      },
    })
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading conversation...</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-destructive">Failed to load conversation.</p>
      </div>
    )
  }

  return (
    <div className="flex h-screen flex-col bg-background">
      {/* Messages Area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4">
        {thread?.messages.map((message) =>
          message.id === latestMessageId && message.role === 'assistant' ? (
            <StreamingMessage
              key={message.id}
              content={message.content}
              speed={40}
            />
          ) : (
            <ChatMessage
              key={message.id}
              role={message.role as 'user' | 'assistant'}
              content={message.content}
              timestamp={message.timestamp}
              audioUrl={message.audioUrl}
              hasAudio={message.hasAudio}
            />
          )
        )}
        {pendingMessage && (
          <ChatMessage
            key="pending"
            role="user"
            content={pendingMessage}
            timestamp={new Date().toISOString()}
          />
        )}
        {isAnyPending && <TypingIndicator />}
      </div>

      {/* Input Area */}
      <div className="border-t bg-card p-4">
        <ChatInput
          onSubmit={handleSendMessage}
          onAudioSubmit={handleSendAudio}
          disabled={isAnyPending}
        />
      </div>
    </div>
  )
}
