import { useCallback, useEffect, useRef, useState } from 'react'
import { useAudioRecorder } from '@/hooks/use-audio-recorder'

interface UsePushToTalkOptions {
  onRecordingComplete?: (audioBlob: Blob) => void
  onRecordingStart?: () => void
  onRecordingEnd?: () => void
  triggerKey?: string
  disabled?: boolean
}

export function usePushToTalk({
  onRecordingComplete,
  onRecordingStart,
  onRecordingEnd,
  triggerKey = ' ',
  disabled = false,
}: UsePushToTalkOptions = {}) {
  const audioRecorder = useAudioRecorder()
  const [isHolding, setIsHolding] = useState(false)
  const isHoldingRef = useRef(false)

  // Track audioBlob and trigger callback when it changes
  useEffect(() => {
    if (audioRecorder.audioBlob && !audioRecorder.isRecording) {
      onRecordingComplete?.(audioRecorder.audioBlob)
      // Reset the recorder after handling the blob
      audioRecorder.resetRecording()
    }
  }, [audioRecorder.audioBlob, audioRecorder.isRecording, onRecordingComplete])

  const startRecording = useCallback(async () => {
    if (disabled || isHoldingRef.current) return

    setIsHolding(true)
    isHoldingRef.current = true
    onRecordingStart?.()
    await audioRecorder.startRecording()
  }, [disabled, audioRecorder, onRecordingStart])

  const stopRecording = useCallback(() => {
    if (!isHoldingRef.current) return

    setIsHolding(false)
    isHoldingRef.current = false
    onRecordingEnd?.()
    audioRecorder.stopRecording()
  }, [audioRecorder, onRecordingEnd])

  // Keyboard handler (spacebar)
  useEffect(() => {
    if (disabled) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is typing in an input/textarea
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return
      }

      if (e.key === triggerKey && !e.repeat && !isHoldingRef.current) {
        e.preventDefault()
        startRecording()
      }
    }

    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === triggerKey && isHoldingRef.current) {
        e.preventDefault()
        stopRecording()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('keyup', handleKeyUp)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('keyup', handleKeyUp)
    }
  }, [disabled, triggerKey, startRecording, stopRecording])

  return {
    isRecording: audioRecorder.isRecording,
    recordingTime: audioRecorder.recordingTime,
    error: audioRecorder.error,
    startRecording,
    stopRecording,
  }
}
