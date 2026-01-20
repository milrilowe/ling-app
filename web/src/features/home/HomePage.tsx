import { Button } from '@/components/ui/button'
import { ConversationUI } from '@/features/conversation/ConversationUI'
import { Languages } from 'lucide-react'

export function Home() {
  return (
    <div className="flex h-full flex-col">
      {/* Utility Panel - always present for consistent spacing */}
      <div className="flex items-center justify-between gap-2 border-b bg-background/95 px-4 py-2 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <Button variant="ghost" size="sm" className="gap-2">
          <Languages className="h-4 w-4" />
          <span>English (US)</span>
        </Button>
        {/* Invisible placeholder to match the height of the thread page buttons */}
        <Button variant="ghost" size="icon" className="invisible">
          <span className="h-5 w-5" />
        </Button>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-hidden">
        <ConversationUI />
      </div>
    </div>
  )
}
