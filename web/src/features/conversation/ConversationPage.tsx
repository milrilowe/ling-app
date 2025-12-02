import { useParams } from '@tanstack/react-router'
import { ConversationUI } from '@/features/conversation/ConversationUI'

export function ConversationPage() {
  const { threadId } = useParams({ from: '/c/$threadId' })

  return <ConversationUI threadId={threadId} />
}
