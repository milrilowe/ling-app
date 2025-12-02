import { Button } from '@/components/ui/button'
import { AudioRecorder } from '@/components/audio/AudioRecorder'
import { Send, Mic } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'

interface ChatInputProps {
  onSubmit: (message: string) => void
  onAudioSubmit?: (audioBlob: Blob) => void
  disabled?: boolean
  placeholder?: string
}

export function ChatInput({
  onSubmit,
  onAudioSubmit,
  disabled = false,
  placeholder = 'Type your response...',
}: ChatInputProps) {
  const [message, setMessage] = useState('')
  const [isRecordingMode, setIsRecordingMode] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSubmit = () => {
    if (message.trim() && !disabled) {
      onSubmit(message.trim())
      setMessage('')
      // Reset textarea height
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const handleAudioComplete = (audioBlob: Blob) => {
    if (onAudioSubmit && !disabled) {
      onAudioSubmit(audioBlob)
      setIsRecordingMode(false)
    }
  }

  const handleAudioCancel = () => {
    setIsRecordingMode(false)
  }

  // Auto-grow textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [message])

  // Show audio recorder if in recording mode
  if (isRecordingMode) {
    return (
      <AudioRecorder
        onRecordingComplete={handleAudioComplete}
        onRecordingCancel={handleAudioCancel}
        maxDuration={300}
      />
    )
  }

  return (
    <div className="flex gap-2">
      <textarea
        ref={textareaRef}
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        rows={1}
        className="flex-1 resize-none rounded-lg border border-input bg-background px-4 py-3 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
        style={{ maxHeight: '200px' }}
      />

      {onAudioSubmit && (
        <Button
          onClick={() => setIsRecordingMode(true)}
          disabled={disabled}
          size="icon"
          variant="outline"
          className="shrink-0"
        >
          <Mic className="h-4 w-4" />
        </Button>
      )}

      <Button
        onClick={handleSubmit}
        disabled={!message.trim() || disabled}
        size="icon"
        className="shrink-0"
      >
        <Send className="h-4 w-4" />
      </Button>
    </div>
  )
}
