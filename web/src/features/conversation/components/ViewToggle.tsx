import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { MessageSquare, Radio } from 'lucide-react'
import type { ConversationViewMode } from '@/features/conversation/ConversationPage'

interface ViewToggleProps {
  currentView: ConversationViewMode
  onToggle: () => void
}

export function ViewToggle({ currentView, onToggle }: ViewToggleProps) {
  const isMessagesView = currentView === 'messages'

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          onClick={onToggle}
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
  )
}
