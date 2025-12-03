import { useParams } from '@tanstack/react-router'
import { ConversationUI } from '@/features/conversation/ConversationUI'
import { ConversationMessagesView } from '@/features/conversation/ConversationMessagesView'
import { ViewToggle } from '@/features/conversation/components/ViewToggle'
import { useState } from 'react'
import { useAudioPlayerContext } from '@/contexts/AudioPlayerContext'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { RotateCcw, Languages } from 'lucide-react'

export type ConversationViewMode = 'audio' | 'messages'

export function ConversationPage() {
  const { threadId } = useParams({ from: '/c/$threadId' })
  const [viewMode, setViewMode] = useState<ConversationViewMode>('audio')
  const audioPlayer = useAudioPlayerContext()

  const toggleView = () => {
    setViewMode((prev) => (prev === 'audio' ? 'messages' : 'audio'))
  }

  const handleRepeat = () => {
    if (audioPlayer.currentUrl) {
      audioPlayer.seek(0)
      audioPlayer.play()
    }
  }

  return (
    <div className="flex h-full flex-col">
      {/* Utility Panel - always present */}
      <div className="flex items-center justify-between gap-2 border-b bg-background/95 px-4 py-2 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <Button variant="ghost" size="sm" className="gap-2">
          <Languages className="h-4 w-4" />
          <span>Language</span>
        </Button>

        {threadId && (
          <TooltipProvider>
            <div className="flex items-center gap-2">
              {viewMode === 'audio' && audioPlayer.currentUrl && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleRepeat}
                      disabled={!audioPlayer.currentUrl}
                    >
                      <RotateCcw className="h-5 w-5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Repeat last message</p>
                  </TooltipContent>
                </Tooltip>
              )}
              <ViewToggle currentView={viewMode} onToggle={toggleView} />
            </div>
          </TooltipProvider>
        )}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-hidden">
        {viewMode === 'audio' ? (
          <ConversationUI threadId={threadId} />
        ) : (
          threadId ? (
            <ConversationMessagesView threadId={threadId} />
          ) : (
            <div className="flex h-full items-center justify-center">
              <p className="text-muted-foreground">Start a conversation to view messages</p>
            </div>
          )
        )}
      </div>
    </div>
  )
}
