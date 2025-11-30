import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createThread,
  getThread,
  getThreads,
  sendMessage,
  getRandomPrompt,
} from '@/lib/api'

const threadKeys = {
  all: ['threads'] as const,
  detail: (threadId: string) => ['threads', threadId] as const,
}

const promptKeys = {
  random: ['prompts', 'random'] as const,
}

export function useRandomPrompt() {
  return useQuery({
    queryKey: promptKeys.random,
    queryFn: getRandomPrompt,
  })
}

export function useThreads() {
  return useQuery({
    queryKey: threadKeys.all,
    queryFn: getThreads,
  })
}

export function useCreateThread() {
  return useMutation({
    mutationFn: createThread,
  })
}

export function useThread(threadId: string) {
  return useQuery({
    queryKey: threadKeys.detail(threadId),
    queryFn: () => getThread(threadId),
  })
}

export function useSendMessage(threadId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (content: string) => sendMessage(threadId, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: threadKeys.detail(threadId) })
    },
  })
}
