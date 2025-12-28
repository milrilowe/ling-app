import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Mic, MessageSquare } from 'lucide-react'
import type { ConversationViewMode } from '@/features/conversation/ConversationPage'

interface ViewToggleProps {
  value: ConversationViewMode
  onChange: (value: ConversationViewMode) => void
}

export function ViewToggle({ value, onChange }: ViewToggleProps) {
  return (
    <ToggleGroup
      type="single"
      value={value}
      onValueChange={(val) => {
        if (val) onChange(val as ConversationViewMode)
      }}
      variant="outline"
      size="sm"
    >
      <ToggleGroupItem value="audio" className="gap-1.5">
        <Mic className="h-4 w-4" />
        <span>Voice</span>
      </ToggleGroupItem>
      <ToggleGroupItem value="messages" className="gap-1.5">
        <MessageSquare className="h-4 w-4" />
        <span>Chat</span>
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
