import type { SendAudioMessageResponse, Thread } from '@/lib/api'
import { sendAudioMessage } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'

// Must match the key pattern in use-thread.ts
const threadKeys = {
  detail: (threadId: string) => ['threads', threadId] as const,
}

export function useSendAudioMessage(threadId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (audioBlob: Blob) => sendAudioMessage(threadId, audioBlob),
    onSuccess: (data: SendAudioMessageResponse) => {
      // Update the thread cache with both new messages
      queryClient.setQueryData(
        threadKeys.detail(threadId),
        (oldData: Thread | undefined) => {
          if (!oldData) return oldData

          return {
            ...oldData,
            messages: [
              ...oldData.messages,
              data.userMessage,
              data.assistantMessage,
            ],
          }
        },
      )

      // Invalidate to ensure fresh data
      queryClient.invalidateQueries({ queryKey: threadKeys.detail(threadId) })
    },
  })
}
