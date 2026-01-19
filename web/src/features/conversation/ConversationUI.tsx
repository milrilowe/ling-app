import { useNavigate, useLocation } from '@tanstack/react-router'
import { AIAvatar } from '@/features/conversation/components/Avatar/AIAvatar'
import { PushToTalkButton } from '@/features/conversation/components/PushToTalk/PushToTalkButton'
import { StatusText } from '@/features/conversation/components/StatusText'
import { useAudioPipeline } from '@/features/conversation/hooks/use-audio-pipeline'
import { useCreateThread } from '@/hooks/use-thread'
import { useEffect, useRef } from 'react'
import { sendAudioMessage } from '@/lib/api'
import { useAudioPlayerContext } from '@/contexts/AudioPlayerContext'
import { handleError } from '@/lib/error-handler'

interface ConversationUIProps {
  threadId?: string
}

interface LocationState {
  audioUrl?: string
}

export function ConversationUI({ threadId }: ConversationUIProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const createThreadMutation = useCreateThread()
  const audioPlayer = useAudioPlayerContext()

  // Track if we've already played audio for this thread (per-mount, not global)
  const playedAudioRef = useRef<string | null>(null)

  // Only use audio pipeline if we have a threadId
  const audioPipelineResult = threadId
    ? useAudioPipeline({ threadId })
    : { handleAudioRecorded: async () => {} }

  const { handleAudioRecorded } = audioPipelineResult

  // Use conversation state from context
  const state = audioPlayer.conversationState
  const isProcessing = state === 'ai-thinking' || state === 'ai-speaking'

  // Handle audio playback from navigation state
  // Only play audio when we first navigate to a thread with audio
  useEffect(() => {
    const locationState = location.state as LocationState | undefined
    if (threadId && locationState?.audioUrl && playedAudioRef.current !== threadId) {
      // We just navigated here with audio to play
      playedAudioRef.current = threadId
      audioPlayer.load(locationState.audioUrl)
      setTimeout(() => {
        audioPlayer.play()
      }, 500)
    }
  }, [threadId, audioPlayer])

  const handleRecordingComplete = async (audioBlob: Blob) => {
    // If we don't have a thread yet, create one first
    if (!threadId) {
      try {
        audioPlayer.setConversationState('ai-thinking')

        // Create the thread (no initial prompt for voice conversations)
        const newThread = await createThreadMutation.mutateAsync({})

        // Send the audio message with the new thread ID
        const response = await sendAudioMessage(newThread.id, audioBlob)

        // Set state to ai-speaking before navigation
        if (response.assistantMessage.audioUrl) {
          audioPlayer.setConversationState('ai-speaking')
        }

        // Navigate to the new thread immediately, passing audio URL if available
        navigate({
          to: '/c/$threadId',
          params: { threadId: newThread.id },
          state: response.assistantMessage.audioUrl
            ? ({ audioUrl: response.assistantMessage.audioUrl } as any)
            : undefined,
        })
      } catch (error) {
        handleError(error, 'Failed to create thread or send message')
        audioPlayer.setConversationState('idle')
      }
    } else {
      // We already have a thread, use the audio pipeline
      handleAudioRecorded(audioBlob)
    }
  }


  // Stop audio when the threadId changes (navigating between threads)
  const prevThreadIdRef = useRef(threadId)
  useEffect(() => {
    const prevThreadId = prevThreadIdRef.current
    prevThreadIdRef.current = threadId

    // If we had a thread and now have a different thread, pause audio
    if (prevThreadId !== undefined && prevThreadId !== threadId) {
      audioPlayer.pause()
    }
  }, [threadId])

  return (
    <div className="relative flex h-full flex-col items-center justify-center bg-gradient-to-br from-background via-muted/20 to-background overflow-hidden">
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

      {/* Push-to-Talk Button (absolute bottom centered in this container) */}
      <div className="absolute bottom-12 left-1/2 -translate-x-1/2">
        <PushToTalkButton
          disabled={isProcessing || createThreadMutation.isPending}
          onRecordingComplete={handleRecordingComplete}
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
