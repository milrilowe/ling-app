import { useNavigate, useLocation } from '@tanstack/react-router'
import { AIAvatar } from '@/features/conversation/components/Avatar/AIAvatar'
import { PushToTalkButton } from '@/features/conversation/components/PushToTalk/PushToTalkButton'
import { StatusText } from '@/features/conversation/components/StatusText'
import { useAudioPipeline } from '@/features/conversation/hooks/use-audio-pipeline'
import { useCreateThread } from '@/hooks/use-thread'
import { useState, useEffect, useRef } from 'react'
import { sendAudioMessage } from '@/lib/api'
import { useAudioPlayerContext } from '@/contexts/AudioPlayerContext'
import type { ConversationState } from './types'

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
  const [localState, setLocalState] = useState<ConversationState>('idle')
  const hasPlayedAudioRef = useRef(false)
  const pendingNavigationRef = useRef<{ threadId: string } | null>(null)

  // Only use audio pipeline if we have a threadId
  const audioPipelineResult = threadId
    ? useAudioPipeline({ threadId })
    : { state: localState, handleAudioRecorded: async () => {}, isProcessing: false }

  const { state, handleAudioRecorded, isProcessing } = audioPipelineResult

  // Handle audio playback from navigation state
  useEffect(() => {
    const locationState = location.state as LocationState | undefined
    if (threadId && locationState?.audioUrl && !hasPlayedAudioRef.current) {
      // We just navigated here with audio to play
      hasPlayedAudioRef.current = true
      audioPlayer.load(locationState.audioUrl)
      setTimeout(() => {
        audioPlayer.play()
      }, 500)
    }
  }, [threadId, location.state, audioPlayer])

  // Reset the played audio flag when threadId changes
  useEffect(() => {
    hasPlayedAudioRef.current = false
  }, [threadId])

  const handleRecordingComplete = async (audioBlob: Blob) => {
    // If we don't have a thread yet, create one first
    if (!threadId) {
      try {
        setLocalState('ai-thinking')

        // Create the thread
        const newThread = await createThreadMutation.mutateAsync({
          initialPrompt: 'Voice conversation',
        })

        // Send the audio message with the new thread ID
        const response = await sendAudioMessage(newThread.id, audioBlob)

        // Load and play audio immediately without navigating
        if (response.assistantMessage.audioUrl) {
          audioPlayer.load(response.assistantMessage.audioUrl)
          setTimeout(() => {
            setLocalState('ai-speaking')
            audioPlayer.play()
          }, 500)

          // Store the thread ID for navigation after audio finishes
          pendingNavigationRef.current = { threadId: newThread.id }
        } else {
          // No audio, navigate immediately
          navigate({
            to: '/c/$threadId',
            params: { threadId: newThread.id },
          })
        }
      } catch (error) {
        console.error('Failed to create thread or send message:', error)
        setLocalState('idle')
      }
    } else {
      // We already have a thread, use the audio pipeline
      handleAudioRecorded(audioBlob)
    }
  }

  // Listen for audio ending when we're managing state locally
  useEffect(() => {
    // Detect when audio finishes: not playing AND either currentTime is 0 (ended event fired)
    // OR currentTime equals duration (audio reached the end)
    const audioFinished = !audioPlayer.isPlaying && (
      audioPlayer.currentTime === 0 ||
      (audioPlayer.duration > 0 && Math.abs(audioPlayer.currentTime - audioPlayer.duration) < 0.1)
    )

    if (!threadId && localState === 'ai-speaking' && audioFinished) {
      setLocalState('idle')

      // If we have a pending navigation, do it now
      if (pendingNavigationRef.current) {
        const { threadId: newThreadId } = pendingNavigationRef.current
        pendingNavigationRef.current = null
        navigate({
          to: '/c/$threadId',
          params: { threadId: newThreadId },
        })
      }
    }
  }, [threadId, localState, audioPlayer.isPlaying, audioPlayer.currentTime, audioPlayer.duration, audioPlayer.error, navigate])

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
