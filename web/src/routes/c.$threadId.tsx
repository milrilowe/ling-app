import { ChatThread } from '@/features/chat/ThreadPage'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/c/$threadId')({
  component: ChatThread,
})
