import { useState, useRef, useEffect } from 'react'
import type { Thread } from '@/lib/api'
import { useNavigate, useParams } from '@tanstack/react-router'
import {
  MessageSquare,
  MoreHorizontal,
  Pencil,
  Trash2,
  Archive,
  ArchiveRestore,
} from 'lucide-react'
import { SidebarMenuItem, SidebarMenuButton } from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DeleteThreadDialog } from './DeleteThreadDialog'
import {
  useDeleteThread,
  useUpdateThread,
  useArchiveThread,
  useUnarchiveThread,
} from '@/hooks/use-thread'

interface ThreadListItemProps {
  thread: Thread
}

export function ThreadListItem({ thread }: ThreadListItemProps) {
  const navigate = useNavigate()
  const params = useParams({ strict: false })
  const isActive = params.threadId === thread.id
  const isArchived = !!thread.archivedAt

  const [isEditing, setIsEditing] = useState(false)
  const [editValue, setEditValue] = useState('')
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const deleteThread = useDeleteThread()
  const updateThread = useUpdateThread()
  const archiveThread = useArchiveThread()
  const unarchiveThread = useUnarchiveThread()

  // Get the display name: explicit name, or first user message, or fallback
  const title =
    thread.name ||
    thread.messages.find((m) => m.role === 'user')?.content ||
    thread.messages[0]?.content ||
    'New conversation'

  // Focus and select all when entering edit mode
  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus()
      inputRef.current.select()
    }
  }, [isEditing])

  const handleClick = () => {
    if (!isEditing) {
      navigate({ to: '/c/$threadId', params: { threadId: thread.id } })
    }
  }

  const startEditing = () => {
    setEditValue(thread.name || title)
    setIsEditing(true)
  }

  const handleRename = () => {
    const trimmed = editValue.trim()
    if (trimmed && trimmed !== title) {
      updateThread.mutate(
        { threadId: thread.id, data: { name: trimmed } },
        {
          onSuccess: () => setIsEditing(false),
          onError: () => setIsEditing(false),
        }
      )
    } else {
      setIsEditing(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleRename()
    } else if (e.key === 'Escape') {
      setIsEditing(false)
    }
  }

  const handleDelete = () => {
    deleteThread.mutate(thread.id, {
      onSuccess: () => {
        setShowDeleteDialog(false)
        if (isActive) {
          navigate({ to: '/' })
        }
      },
    })
  }

  const handleArchive = () => {
    if (isArchived) {
      unarchiveThread.mutate(thread.id)
    } else {
      archiveThread.mutate(thread.id, {
        onSuccess: () => {
          if (isActive) {
            navigate({ to: '/' })
          }
        },
      })
    }
  }

  return (
    <>
      <SidebarMenuItem className="group/item relative">
        <SidebarMenuButton
          onClick={handleClick}
          isActive={isActive}
          tooltip={title}
          className="flex items-start gap-3 pr-8"
        >
          <MessageSquare className="mt-0.5 h-4 w-4 shrink-0" />
          <div className="flex-1 overflow-hidden">
            {isEditing ? (
              <Input
                ref={inputRef}
                value={editValue}
                onChange={(e) => setEditValue(e.target.value)}
                onBlur={handleRename}
                onKeyDown={handleKeyDown}
                onClick={(e) => e.stopPropagation()}
                className="h-6 px-1 py-0 text-sm font-medium"
                disabled={updateThread.isPending}
              />
            ) : (
              <Tooltip delayDuration={700}>
                <TooltipTrigger asChild>
                  <p className="truncate text-sm font-medium leading-tight">
                    {title}
                  </p>
                </TooltipTrigger>
                <TooltipContent side="right" align="center">
                  <p className="max-w-xs">{title}</p>
                </TooltipContent>
              </Tooltip>
            )}
          </div>
        </SidebarMenuButton>
        {/* Hover actions button */}
        {!isEditing && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-1 top-1/2 h-6 w-6 -translate-y-1/2 opacity-0 transition-opacity group-hover/item:opacity-100"
                onClick={(e) => e.stopPropagation()}
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={startEditing}>
                <Pencil className="mr-2 h-4 w-4" />
                Rename
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleArchive}>
                {isArchived ? (
                  <>
                    <ArchiveRestore className="mr-2 h-4 w-4" />
                    Unarchive
                  </>
                ) : (
                  <>
                    <Archive className="mr-2 h-4 w-4" />
                    Archive
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => setShowDeleteDialog(true)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </SidebarMenuItem>

      <DeleteThreadDialog
        open={showDeleteDialog}
        onOpenChange={setShowDeleteDialog}
        onConfirm={handleDelete}
        isLoading={deleteThread.isPending}
      />
    </>
  )
}
