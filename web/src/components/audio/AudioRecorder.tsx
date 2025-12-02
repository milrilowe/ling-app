import { useEffect } from 'react'
import { Mic, Square, Pause, Play, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAudioRecorder } from '@/hooks/use-audio-recorder'
import { formatDuration } from '@/lib/audio-utils'
import { cn } from '@/lib/utils'

interface AudioRecorderProps {
  onRecordingComplete?: (audioBlob: Blob) => void
  onRecordingCancel?: () => void
  maxDuration?: number // in seconds
  className?: string
}

export function AudioRecorder({
  onRecordingComplete,
  onRecordingCancel,
  maxDuration = 300, // 5 minutes default
  className,
}: AudioRecorderProps) {
  const {
    isRecording,
    isPaused,
    recordingTime,
    audioBlob,
    error,
    startRecording,
    stopRecording,
    pauseRecording,
    resumeRecording,
    resetRecording,
  } = useAudioRecorder()

  // Auto-stop when max duration is reached
  useEffect(() => {
    if (isRecording && recordingTime >= maxDuration) {
      stopRecording()
    }
  }, [isRecording, recordingTime, maxDuration, stopRecording])

  // Call onRecordingComplete when audio blob is ready
  useEffect(() => {
    if (audioBlob && onRecordingComplete) {
      onRecordingComplete(audioBlob)
    }
  }, [audioBlob, onRecordingComplete])

  const handleStart = async () => {
    await startRecording()
  }

  const handleStop = () => {
    stopRecording()
  }

  const handleCancel = () => {
    resetRecording()
    if (onRecordingCancel) {
      onRecordingCancel()
    }
  }

  const handlePauseResume = () => {
    if (isPaused) {
      resumeRecording()
    } else {
      pauseRecording()
    }
  }

  if (error) {
    return (
      <div className={cn('rounded-lg border border-destructive bg-destructive/10 p-4', className)}>
        <p className="text-sm text-destructive">{error}</p>
        <Button
          variant="outline"
          size="sm"
          onClick={resetRecording}
          className="mt-2"
        >
          Dismiss
        </Button>
      </div>
    )
  }

  if (!isRecording && !audioBlob) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <Button
          variant="default"
          size="icon"
          onClick={handleStart}
          className="rounded-full"
        >
          <Mic className="h-4 w-4" />
        </Button>
        <span className="text-sm text-muted-foreground">
          Click to start recording
        </span>
      </div>
    )
  }

  if (isRecording) {
    return (
      <div className={cn('flex items-center gap-3 rounded-lg border bg-card p-4', className)}>
        <div className="flex items-center gap-2 flex-1">
          <div className={cn(
            'h-3 w-3 rounded-full bg-red-500',
            !isPaused && 'animate-pulse'
          )} />
          <span className="font-mono text-lg">
            {formatDuration(recordingTime)}
          </span>
          <span className="text-sm text-muted-foreground">
            / {formatDuration(maxDuration)}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={handlePauseResume}
          >
            {isPaused ? (
              <Play className="h-4 w-4" />
            ) : (
              <Pause className="h-4 w-4" />
            )}
          </Button>

          <Button
            variant="outline"
            size="icon"
            onClick={handleCancel}
          >
            <Trash2 className="h-4 w-4" />
          </Button>

          <Button
            variant="default"
            size="icon"
            onClick={handleStop}
          >
            <Square className="h-4 w-4" />
          </Button>
        </div>
      </div>
    )
  }

  // audioBlob is ready but component hasn't been unmounted yet
  return null
}
