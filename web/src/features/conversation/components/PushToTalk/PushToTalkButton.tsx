import { Mic, Square } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDuration } from '@/lib/audio-utils'
import { usePushToTalk } from './use-push-to-talk'

interface PushToTalkButtonProps {
  disabled?: boolean
  onRecordingComplete?: (audioBlob: Blob) => void
  className?: string
}

export function PushToTalkButton({
  disabled = false,
  onRecordingComplete,
  className,
}: PushToTalkButtonProps) {
  const { isRecording, recordingTime, error, startRecording, stopRecording } = usePushToTalk({
    onRecordingComplete,
    disabled,
  })

  return (
    <div className={cn('flex flex-col items-center gap-4', className)}>
      {/* Recording timer */}
      {isRecording && (
        <span className="text-sm font-mono text-muted-foreground">
          {formatDuration(recordingTime)}
        </span>
      )}

      {/* Main button */}
      <button
        disabled={disabled}
        className={cn(
          'relative h-20 w-20 rounded-full transition-all duration-200',
          'bg-primary shadow-lg hover:shadow-xl',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          'focus:outline-none focus:ring-4 focus:ring-primary/20',
          isRecording && 'scale-110 bg-destructive',
        )}
        onMouseDown={!disabled ? startRecording : undefined}
        onMouseUp={!disabled ? stopRecording : undefined}
        onMouseLeave={isRecording ? stopRecording : undefined}
        onTouchStart={!disabled ? startRecording : undefined}
        onTouchEnd={!disabled ? stopRecording : undefined}
        aria-label={isRecording ? 'Release to send' : 'Hold to talk'}
        aria-pressed={isRecording}
      >
        <div className="flex items-center justify-center">
          {isRecording ? (
            <Square className="h-8 w-8 text-white fill-white" />
          ) : (
            <Mic className="h-8 w-8 text-white" />
          )}
        </div>

        {/* Pulse animation ring when recording */}
        {isRecording && (
          <div className="absolute inset-0 rounded-full border-4 border-destructive animate-ping opacity-75" />
        )}
      </button>

      {/* Hint text */}
      {!isRecording && !disabled && (
        <span className="text-xs text-muted-foreground">
          Hold to talk or press spacebar
        </span>
      )}

      {/* Error message */}
      {error && (
        <span className="text-xs text-destructive">
          {error}
        </span>
      )}
    </div>
  )
}
