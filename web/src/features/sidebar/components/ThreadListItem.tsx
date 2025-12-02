import type { Thread } from '@/lib/api'
import { formatRelativeTime } from '@/lib/formatting'
import { useNavigate, useParams } from '@tanstack/react-router'
import { MessageSquare } from 'lucide-react'
import { SidebarMenuItem, SidebarMenuButton } from '@/components/ui/sidebar'

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
    <SidebarMenuItem>
      <SidebarMenuButton
        onClick={handleClick}
        isActive={isActive}
        tooltip={title}
        className="flex items-start gap-3"
      >
        <MessageSquare className="mt-0.5 h-4 w-4 shrink-0" />
        <div className="flex-1 overflow-hidden">
          <p className="truncate text-sm font-medium leading-tight">{title}</p>
          <p className="mt-0.5 text-xs opacity-70 group-data-[collapsible=icon]:hidden">
            {formatRelativeTime(thread.createdAt)}
          </p>
        </div>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}
