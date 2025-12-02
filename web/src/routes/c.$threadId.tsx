import { ConversationPage } from '@/features/conversation/ConversationPage'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/c/$threadId')({
  component: ConversationPage,
})
