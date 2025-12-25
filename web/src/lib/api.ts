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

export interface PhonemeDetail {
  expected: string
  actual: string
  type: 'match' | 'substitute' | 'delete' | 'insert'
  position: number
}

export interface PronunciationAnalysis {
  audio_ipa: string
  expected_ipa: string
  phoneme_count: number
  match_count: number
  substitution_count: number
  deletion_count: number
  insertion_count: number
  phoneme_details: PhonemeDetail[]
  processing_time_ms: number
}

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
  audioUrl?: string
  audioDurationSeconds?: number
  hasAudio?: boolean
  pronunciationStatus?: 'none' | 'pending' | 'complete' | 'failed'
  pronunciationAnalysis?: PronunciationAnalysis
  pronunciationError?: string
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
  const thread = await callAPI<Thread>(`/api/threads/${threadId}`)

  // Convert audio storage keys to presigned URLs
  if (thread.messages) {
    await Promise.all(
      thread.messages.map(async (message) => {
        if (message.hasAudio && message.audioUrl) {
          try {
            // The audioUrl from backend is actually a storage key
            // We need to fetch the presigned URL from MinIO
            message.audioUrl = await getAudioUrl(message.audioUrl)
          } catch (error) {
            console.error('Failed to fetch audio URL for message:', message.id, error)
            // Keep the original key if fetching fails
          }
        }
      })
    )
  }

  return thread
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

export interface SendAudioMessageResponse {
  userMessage: Message
  assistantMessage: Message
}

export async function sendAudioMessage(
  threadId: string,
  audioBlob: Blob
): Promise<SendAudioMessageResponse> {
  const formData = new FormData()
  formData.append('audio', audioBlob, 'recording.webm')

  const url = `${API_BASE_URL}/api/threads/${threadId}/messages/audio`

  try {
    const response = await fetch(url, {
      method: 'POST',
      body: formData,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ApiError(
        errorData?.error || `API error: ${response.status}`,
        response.status,
        errorData
      )
    }

    const data: SendAudioMessageResponse = await response.json()

    // Convert audio storage keys to presigned URLs
    if (data.userMessage.hasAudio && data.userMessage.audioUrl) {
      try {
        data.userMessage.audioUrl = await getAudioUrl(data.userMessage.audioUrl)
      } catch (error) {
        console.error('Failed to fetch user audio URL:', error)
      }
    }

    if (data.assistantMessage.hasAudio && data.assistantMessage.audioUrl) {
      try {
        data.assistantMessage.audioUrl = await getAudioUrl(data.assistantMessage.audioUrl)
      } catch (error) {
        console.error('Failed to fetch assistant audio URL:', error)
      }
    }

    return data
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }
    throw new ApiError('Network error', 0, error)
  }
}

export async function getAudioUrl(audioKey: string): Promise<string> {
  const response = await callAPI<{ url: string }>(`/api/audio/${audioKey}`)
  return response.url
}

export { ApiError }
