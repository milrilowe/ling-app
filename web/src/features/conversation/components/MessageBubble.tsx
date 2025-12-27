import { useState } from 'react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { formatMessageTime } from '@/lib/formatting'
import { cn } from '@/lib/utils'
import { Volume2, ChevronDown } from 'lucide-react'
import { useAudioPlayerContext } from '@/contexts/AudioPlayerContext'
import { PronunciationDisplay } from './PronunciationDisplay'
import type { PronunciationAnalysis } from '@/lib/api'

interface MessageBubbleProps {
  role: 'user' | 'assistant'
  content: string
  timestamp: string | Date
  audioUrl?: string
  hasAudio?: boolean
  pronunciationStatus?: 'none' | 'pending' | 'complete' | 'failed'
  pronunciationAnalysis?: PronunciationAnalysis
  pronunciationError?: string
}

export function MessageBubble({
  role,
  content,
  timestamp,
  audioUrl,
  hasAudio,
  pronunciationStatus,
  pronunciationAnalysis,
  pronunciationError,
}: MessageBubbleProps) {
  const isUser = role === 'user'
  const audioPlayer = useAudioPlayerContext()
  const [isPronunciationExpanded, setIsPronunciationExpanded] = useState(false)

  // Check if this specific audio is currently loaded
  const isThisAudioLoaded = audioPlayer.currentUrl === audioUrl
  const isThisAudioPlaying = isThisAudioLoaded && audioPlayer.isPlaying

  // Show chevron when pronunciation is complete
  const showPronunciationToggle = isUser && hasAudio && pronunciationStatus === 'complete'

  const handlePlayAudio = () => {
    if (!audioUrl) return

    // If this specific audio is currently playing, pause it
    if (isThisAudioPlaying) {
      audioPlayer.pause()
    } else {
      // Load and play this audio (will stop any other audio playing)
      audioPlayer.load(audioUrl)
      audioPlayer.play()
    }
  }

  return (
    <div
      className={cn(
        'group mb-4 flex animate-in fade-in slide-in-from-bottom-2 duration-300',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      <div
        className={cn(
          'flex max-w-[80%] gap-3',
          isUser ? 'flex-row-reverse' : 'flex-row'
        )}
      >
        {/* Avatar */}
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback
            className={cn(
              isUser
                ? 'bg-bubble-user text-bubble-user-foreground'
                : 'bg-bubble-ai text-bubble-ai-foreground'
            )}
          >
            {isUser ? 'Y' : 'AI'}
          </AvatarFallback>
        </Avatar>

        {/* Message bubble */}
        <div className="flex flex-col w-fit">
          <div className="relative">
            <div
              className={cn(
                'px-4 py-2 rounded-2xl',
                isUser
                  ? 'bg-bubble-user text-bubble-user-foreground'
                  : 'bg-bubble-ai text-bubble-ai-foreground',
                showPronunciationToggle && 'cursor-pointer'
              )}
              onClick={showPronunciationToggle ? () => setIsPronunciationExpanded(!isPronunciationExpanded) : undefined}
            >
              <div className="flex items-center gap-2">
                <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
                  {content}
                </p>
                {showPronunciationToggle && (
                  <ChevronDown
                    className={cn(
                      'h-3 w-3 shrink-0 opacity-60 transition-transform',
                      isPronunciationExpanded && 'rotate-180'
                    )}
                  />
                )}
              </div>
            </div>

            {/* Audio controls - show on hover like emoji reactions */}
            {hasAudio && audioUrl && (
              <div
                className={cn(
                  'absolute -top-3 opacity-0 transition-opacity group-hover:opacity-100',
                  isUser ? 'right-0' : 'left-0'
                )}
              >
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="secondary"
                        size="icon"
                        className="h-7 w-7 rounded-full shadow-md"
                        onClick={handlePlayAudio}
                      >
                        <Volume2 className={cn('h-4 w-4', isThisAudioPlaying && 'text-primary')} />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{isThisAudioPlaying ? 'Stop audio' : 'Play audio'}</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            )}
          </div>

          {/* Pronunciation analysis - only for user audio messages */}
          {isUser && hasAudio && pronunciationStatus && pronunciationStatus !== 'none' && (
            <PronunciationDisplay
              status={pronunciationStatus}
              analysis={pronunciationAnalysis}
              error={pronunciationError}
              expectedText={content}
              isExpanded={isPronunciationExpanded}
              onToggleExpand={() => setIsPronunciationExpanded(!isPronunciationExpanded)}
            />
          )}

          {/* Timestamp */}
          <span
            className={cn(
              'mt-1 text-xs text-muted-foreground',
              isUser ? 'text-right' : 'text-left'
            )}
          >
            {formatMessageTime(timestamp)}
          </span>
        </div>
      </div>
    </div>
  )
}
