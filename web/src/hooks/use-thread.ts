import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createThread,
  getThread,
  getThreads,
  getArchivedThreads,
  sendMessage,
  getRandomPrompt,
  updateThread,
  deleteThread,
  archiveThread,
  unarchiveThread,
} from '@/lib/api'

const threadKeys = {
  all: ['threads'] as const,
  archived: ['threads', 'archived'] as const,
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
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createThread,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: threadKeys.all })
    },
  })
}

export function useThread(threadId: string) {
  const query = useQuery({
    queryKey: threadKeys.detail(threadId),
    queryFn: () => getThread(threadId),
    // Poll every 5 seconds if there's a pending pronunciation analysis
    refetchInterval: (query) => {
      const hasPendingAnalysis = query.state.data?.messages?.some(
        (m) => m.pronunciationStatus === 'pending'
      )
      return hasPendingAnalysis ? 5000 : false
    },
  })

  return query
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

export function useArchivedThreads() {
  return useQuery({
    queryKey: threadKeys.archived,
    queryFn: getArchivedThreads,
  })
}

export function useUpdateThread() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      threadId,
      data,
    }: {
      threadId: string
      data: { name?: string | null }
    }) => updateThread(threadId, data),
    onSuccess: (_, { threadId }) => {
      queryClient.invalidateQueries({ queryKey: threadKeys.all })
      queryClient.invalidateQueries({ queryKey: threadKeys.archived })
      queryClient.invalidateQueries({ queryKey: threadKeys.detail(threadId) })
    },
  })
}

export function useDeleteThread() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteThread,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: threadKeys.all })
      queryClient.invalidateQueries({ queryKey: threadKeys.archived })
    },
  })
}

export function useArchiveThread() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: archiveThread,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: threadKeys.all })
      queryClient.invalidateQueries({ queryKey: threadKeys.archived })
    },
  })
}

export function useUnarchiveThread() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: unarchiveThread,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: threadKeys.all })
      queryClient.invalidateQueries({ queryKey: threadKeys.archived })
    },
  })
}
