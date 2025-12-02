import { useState, useEffect, useRef } from 'react'
import { Play, Pause, Volume2, VolumeX } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { useAudioPlayer } from '@/hooks/use-audio-player'
import { formatDuration } from '@/lib/audio-utils'
import { cn } from '@/lib/utils'

interface AudioPlayerProps {
  url: string
  className?: string
  autoPlay?: boolean
}

export function AudioPlayer({ url, className, autoPlay = false }: AudioPlayerProps) {
  const { isPlaying, isLoading, currentTime, duration, error, play, pause, seek } = useAudioPlayer(url)
  const [isMuted, setIsMuted] = useState(false)
  const hasAutoPlayed = useRef(false)

  // Auto-play when component mounts (only once)
  useEffect(() => {
    if (autoPlay && !hasAutoPlayed.current && !isLoading && duration > 0) {
      hasAutoPlayed.current = true
      play()
    }
  }, [autoPlay, isLoading, duration, play])

  const handlePlayPause = () => {
    if (isPlaying) {
      pause()
    } else {
      play()
    }
  }

  const handleSeek = (value: number[]) => {
    seek(value[0])
  }

  const toggleMute = () => {
    setIsMuted(!isMuted)
  }

  if (error) {
    return (
      <div className={cn('rounded-lg border border-destructive bg-destructive/10 p-3', className)}>
        <p className="text-sm text-destructive">{error}</p>
      </div>
    )
  }

  return (
    <div className={cn('flex items-center gap-3 rounded-lg border bg-card p-3', className)}>
      <Button
        variant="ghost"
        size="icon"
        onClick={handlePlayPause}
        disabled={isLoading}
        className="shrink-0"
      >
        {isPlaying ? (
          <Pause className="h-4 w-4" />
        ) : (
          <Play className="h-4 w-4" />
        )}
      </Button>

      <div className="flex-1 space-y-1">
        <Slider
          value={[currentTime]}
          max={duration || 100}
          step={0.1}
          onValueChange={handleSeek}
          disabled={isLoading || !duration}
          className="cursor-pointer"
        />
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{formatDuration(currentTime)}</span>
          <span>{duration ? formatDuration(duration) : '--:--'}</span>
        </div>
      </div>

      <Button
        variant="ghost"
        size="icon"
        onClick={toggleMute}
        className="shrink-0"
      >
        {isMuted ? (
          <VolumeX className="h-4 w-4" />
        ) : (
          <Volume2 className="h-4 w-4" />
        )}
      </Button>
    </div>
  )
}
