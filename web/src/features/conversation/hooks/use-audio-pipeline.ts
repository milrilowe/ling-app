import { useCallback, useEffect, useRef, useState } from 'react'
import { useSendAudioMessage } from '@/hooks/use-send-audio-message'
import { useAudioPlayer } from '@/hooks/use-audio-player'
import type { ConversationState } from '../types'

interface AudioPipelineOptions {
  threadId: string
  onAudioEnded?: () => void
  onError?: (error: Error) => void
}

export function useAudioPipeline({ threadId, onAudioEnded, onError }: AudioPipelineOptions) {
  const [state, setState] = useState<ConversationState>('idle')
  const sendAudioMutation = useSendAudioMessage(threadId)
  const audioPlayer = useAudioPlayer()

  const handleAudioRecorded = useCallback(async (audioBlob: Blob) => {
    setState('ai-thinking')

    try {
      const response = await sendAudioMutation.mutateAsync(audioBlob)

      if (response.assistantMessage.audioUrl) {
        // Load the audio
        audioPlayer.load(response.assistantMessage.audioUrl)

        // Wait a bit for metadata to load, then play
        setTimeout(() => {
          setState('ai-speaking')
          audioPlayer.play()
        }, 500)
      } else {
        setState('idle')
      }
    } catch (error) {
      console.error('Audio pipeline error:', error)
      setState('idle')
      onError?.(error instanceof Error ? error : new Error('Unknown error'))
    }
  }, [sendAudioMutation, audioPlayer, onError])

  // Listen for audio ending
  useEffect(() => {
    if (state === 'ai-speaking' && !audioPlayer.isPlaying && audioPlayer.currentTime === 0) {
      setState('idle')
      onAudioEnded?.()
    }
  }, [state, audioPlayer.isPlaying, audioPlayer.currentTime, onAudioEnded])

  const reset = useCallback(() => {
    setState('idle')
    audioPlayer.pause()
  }, [audioPlayer])

  return {
    state,
    handleAudioRecorded,
    audioPlayer,
    reset,
    isProcessing: state === 'ai-thinking' || state === 'ai-speaking',
  }
}
