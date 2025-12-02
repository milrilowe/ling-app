export type ConversationState =
  | 'idle'
  | 'recording'
  | 'ai-thinking'
  | 'ai-speaking'
  | 'archived'

export interface ConversationMetadata {
  threadId: string | null
  state: ConversationState
  lastInteractionAt: Date | null
  isRecording: boolean
}

export type ConversationMode = 'active' | 'archived'
