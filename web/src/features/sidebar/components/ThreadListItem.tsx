import type { Thread } from '@/lib/api'
import { formatRelativeTime } from '@/lib/formatting'
import { cn } from '@/lib/utils'
import { useNavigate, useParams } from '@tanstack/react-router'
import { MessageSquare } from 'lucide-react'

interface ThreadListItemProps {
  thread: Thread
}

export function ThreadListItem({ thread }: ThreadListItemProps) {
  const navigate = useNavigate()
  const params = useParams({ strict: false })
  const isActive = params.threadId === thread.id

  // Get the first user or AI message as title
  const title =
    thread.messages.find((m) => m.role === 'user')?.content ||
    thread.messages[0]?.content ||
    'New conversation'

  const handleClick = () => {
    navigate({ to: '/c/$threadId', params: { threadId: thread.id } })
  }

  return (
    <button
      onClick={handleClick}
      className={cn(
        'group flex w-full items-start gap-3 rounded-lg px-3 py-2.5 text-left transition-colors',
        'hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
        isActive && 'bg-sidebar-accent text-sidebar-accent-foreground',
      )}
    >
      <MessageSquare className="mt-0.5 h-4 w-4 shrink-0 opacity-70" />
      <div className="flex-1 overflow-hidden">
        <p className="truncate text-sm font-medium leading-tight">{title}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {formatRelativeTime(thread.createdAt)}
        </p>
      </div>
    </button>
  )
}
