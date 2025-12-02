import type { SendAudioMessageResponse, Thread } from '@/lib/api'
import { sendAudioMessage } from '@/lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'

export function useSendAudioMessage(threadId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (audioBlob: Blob) => sendAudioMessage(threadId, audioBlob),
    onSuccess: (data: SendAudioMessageResponse) => {
      // Update the thread cache with both new messages
      queryClient.setQueryData(
        ['thread', threadId],
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
      queryClient.invalidateQueries({ queryKey: ['thread', threadId] })
    },
  })
}
