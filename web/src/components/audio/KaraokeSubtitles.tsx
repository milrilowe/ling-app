import { useMemo } from 'react'
import { cn } from '@/lib/utils'
import type { WordTiming } from '@/lib/api'

interface KaraokeSubtitlesProps {
  text: string
  wordTimings: WordTiming[]
  currentTime: number
  isPlaying: boolean
  className?: string
}

export function KaraokeSubtitles({
  text,
  wordTimings,
  currentTime,
  isPlaying,
  className,
}: KaraokeSubtitlesProps) {
  // Find the current word based on playback time
  const currentWordIndex = useMemo(() => {
    if (!isPlaying || wordTimings.length === 0) return -1

    for (let i = 0; i < wordTimings.length; i++) {
      const timing = wordTimings[i]
      if (currentTime >= timing.start && currentTime < timing.end) {
        return i
      }
    }

    // Check if we're past the last word
    const lastTiming = wordTimings[wordTimings.length - 1]
    if (currentTime >= lastTiming.end) {
      return -1 // Past all words
    }

    return -1
  }, [currentTime, isPlaying, wordTimings])

  // If no word timings, just show static text
  if (wordTimings.length === 0) {
    return (
      <span className={cn('whitespace-pre-wrap break-words', className)}>
        {text}
      </span>
    )
  }

  return (
    <span className={cn('whitespace-pre-wrap break-words', className)}>
      {wordTimings.map((timing, index) => {
        const isPast = isPlaying && currentTime >= timing.end
        const isCurrent = index === currentWordIndex
        const isUpcoming = isPlaying && currentTime < timing.start

        return (
          <span
            key={`${timing.word}-${index}`}
            className={cn(
              'transition-colors duration-150',
              isCurrent && 'text-primary font-medium',
              isPast && 'opacity-70',
              isUpcoming && 'opacity-50'
            )}
          >
            {timing.word}
            {index < wordTimings.length - 1 && ' '}
          </span>
        )
      })}
    </span>
  )
}
