import { useCallback, useEffect } from 'react'
import { useSendAudioMessage } from '@/hooks/use-send-audio-message'
import { useAudioPlayerContext } from '@/contexts/AudioPlayerContext'
import { handleError } from '@/lib/error-handler'

interface AudioPipelineOptions {
  threadId: string
  onAudioEnded?: () => void
  onError?: (error: Error) => void
}

export function useAudioPipeline({ threadId, onAudioEnded, onError }: AudioPipelineOptions) {
  const sendAudioMutation = useSendAudioMessage(threadId)
  const audioPlayer = useAudioPlayerContext()

  const handleAudioRecorded = useCallback(async (audioBlob: Blob) => {
    audioPlayer.setConversationState('ai-thinking')

    try {
      const response = await sendAudioMutation.mutateAsync(audioBlob)

      if (response.assistantMessage.audioUrl) {
        // Load the audio
        audioPlayer.load(response.assistantMessage.audioUrl)

        // Wait a bit for metadata to load, then play
        setTimeout(() => {
          audioPlayer.setConversationState('ai-speaking')
          audioPlayer.play()
        }, 500)
      } else {
        audioPlayer.setConversationState('idle')
      }
    } catch (error) {
      handleError(error, 'Audio pipeline error')
      audioPlayer.setConversationState('idle')
      onError?.(error instanceof Error ? error : new Error('Unknown error'))
    }
  }, [sendAudioMutation, audioPlayer, onError])

  // Listen for audio ending - the audio player will set state to idle when ended
  useEffect(() => {
    if (audioPlayer.conversationState === 'ai-speaking' && !audioPlayer.isPlaying && audioPlayer.currentTime === 0) {
      onAudioEnded?.()
    }
  }, [audioPlayer.conversationState, audioPlayer.isPlaying, audioPlayer.currentTime, onAudioEnded])

  const reset = useCallback(() => {
    audioPlayer.setConversationState('idle')
    audioPlayer.pause()
  }, [audioPlayer])

  return {
    handleAudioRecorded,
    reset,
  }
}
