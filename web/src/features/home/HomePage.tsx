import { useEffect, useState } from 'react'
import { useCreateThread } from '@/hooks/use-thread'
import { AIAvatar } from '@/features/conversation/components/Avatar/AIAvatar'
import { PushToTalkButton } from '@/features/conversation/components/PushToTalk/PushToTalkButton'
import { StatusText } from '@/features/conversation/components/StatusText'
import { useAudioPipeline } from '@/features/conversation/hooks/use-audio-pipeline'

export function Home() {
  const [threadId, setThreadId] = useState<string | null>(null)
  const createThreadMutation = useCreateThread()

  // Create thread on mount
  useEffect(() => {
    const createThread = async () => {
      try {
        const thread = await createThreadMutation.mutateAsync({
          initialPrompt: 'Let\'s have a conversation!',
        })
        setThreadId(thread.id)
      } catch (error) {
        console.error('Failed to create thread:', error)
      }
    }

    createThread()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Show loading state while creating thread
  if (!threadId) {
    return (
      <div className="relative flex h-screen flex-col items-center justify-center bg-gradient-to-br from-background via-muted/20 to-background overflow-hidden">
        <div className="flex flex-col items-center gap-8">
          <AIAvatar
            isThinking={createThreadMutation.isPending}
            isSpeaking={false}
            audioLevel={0.5}
          />

          <div className="text-center">
            {createThreadMutation.isPending ? (
              <p className="text-lg text-muted-foreground">
                Starting conversation...
              </p>
            ) : createThreadMutation.isError ? (
              <>
                <h1 className="text-2xl font-semibold text-foreground mb-2">
                  Oops!
                </h1>
                <p className="text-destructive">
                  Failed to start conversation. Please refresh to try again.
                </p>
              </>
            ) : (
              <p className="text-lg text-muted-foreground">
                Loading...
              </p>
            )}
          </div>
        </div>
      </div>
    )
  }

  // Once thread is created, render the conversation UI
  return <ConversationUI threadId={threadId} />
}

function ConversationUI({ threadId }: { threadId: string }) {
  const { state, handleAudioRecorded, isProcessing } = useAudioPipeline({
    threadId,
  })

  return (
    <div className="relative flex h-screen flex-col items-center justify-center bg-gradient-to-br from-background via-muted/20 to-background overflow-hidden">
      {/* Centered Avatar Section */}
      <div className="flex flex-col items-center gap-8">
        <AIAvatar
          isThinking={state === 'ai-thinking'}
          isSpeaking={state === 'ai-speaking'}
          audioLevel={0.5}
        />

        {/* Status text */}
        <StatusText state={state} />
      </div>

      {/* Push-to-Talk Button (fixed bottom) */}
      <div className="fixed bottom-12 left-1/2 -translate-x-1/2">
        <PushToTalkButton
          disabled={isProcessing}
          onRecordingComplete={handleAudioRecorded}
        />
      </div>

      {/* Screen reader status */}
      <div className="sr-only" role="status" aria-live="polite">
        {state === 'ai-thinking' && 'AI is thinking...'}
        {state === 'ai-speaking' && 'AI is speaking'}
        {state === 'recording' && 'Recording your message'}
      </div>
    </div>
  )
}
