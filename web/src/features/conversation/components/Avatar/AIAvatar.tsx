import { Bot } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AvatarVisualizer } from './AvatarVisualizer'
import { ThinkingDots } from './ThinkingDots'

interface AIAvatarProps {
  isThinking: boolean
  isSpeaking: boolean
  audioLevel?: number
}

export function AIAvatar({ isThinking, isSpeaking, audioLevel = 0.5 }: AIAvatarProps) {
  return (
    <div className="relative flex flex-col items-center">
      {/* Main avatar circle - wrapped in fixed-size container to prevent layout shift */}
      <div className="relative w-48 h-48 flex items-center justify-center">
        <div
          className={cn(
            'relative h-48 w-48 rounded-full bg-gradient-to-br from-info via-success to-primary',
            'transition-all duration-300 flex items-center justify-center',
            isSpeaking && 'scale-105',
            isThinking && 'animate-pulse',
          )}
        >
          {/* Avatar icon */}
          <Bot className="h-24 w-24 text-white" />

          {/* Audio visualizer ring */}
          {isSpeaking && <AvatarVisualizer audioLevel={audioLevel} />}
        </div>

        {/* Thinking indicator - positioned absolutely to not affect layout */}
        {isThinking && (
          <div className="absolute -bottom-8 left-1/2 -translate-x-1/2 whitespace-nowrap">
            <ThinkingDots />
          </div>
        )}
      </div>
    </div>
  )
}
