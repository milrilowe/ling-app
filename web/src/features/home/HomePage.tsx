import { ConversationUI } from '@/features/conversation/ConversationUI'

export function Home() {
  // No thread creation on mount - let ConversationUI handle it
  return <ConversationUI />
}
