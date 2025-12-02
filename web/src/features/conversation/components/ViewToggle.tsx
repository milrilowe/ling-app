import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { MessageSquare, Radio } from 'lucide-react'
import type { ConversationViewMode } from '@/hooks/use-conversation-view'

interface ViewToggleProps {
  currentView: ConversationViewMode
  onToggle: () => void
}

export function ViewToggle({ currentView, onToggle }: ViewToggleProps) {
  const isMessagesView = currentView === 'messages'

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggle}
            className="absolute top-4 right-4 z-10"
          >
            {isMessagesView ? (
              <Radio className="h-5 w-5" />
            ) : (
              <MessageSquare className="h-5 w-5" />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>{isMessagesView ? 'Switch to audio view' : 'Switch to messages view'}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
