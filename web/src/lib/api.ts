const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function callAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ApiError(
        errorData?.message || `API error: ${response.status}`,
        response.status,
        errorData
      )
    }

    return response.json()
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }
    throw new ApiError('Network error', 0, error)
  }
}

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export interface Thread {
  id: string
  messages: Message[]
  createdAt: string
}

interface CreateThreadRequest {
  initialPrompt?: string
  firstUserMessage?: string
}

export async function getRandomPrompt(): Promise<string> {
  const response = await callAPI<{ prompt: string }>('/api/prompts/random')
  return response.prompt
}

export async function createThread(request?: CreateThreadRequest): Promise<Thread> {
  return callAPI<Thread>('/api/threads', {
    method: 'POST',
    body: request ? JSON.stringify(request) : undefined,
  })
}

export async function getThread(threadId: string): Promise<Thread> {
  return callAPI<Thread>(`/api/threads/${threadId}`)
}

export async function sendMessage(
  threadId: string,
  content: string
): Promise<Message> {
  return callAPI<Message>(`/api/threads/${threadId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}

export async function getThreads(): Promise<Thread[]> {
  return callAPI<Thread[]>('/api/threads')
}

export { ApiError }
