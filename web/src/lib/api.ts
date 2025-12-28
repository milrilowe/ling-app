const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function callAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`

  try {
    const response = await fetch(url, {
      ...options,
      credentials: 'include', // Send cookies with requests
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ApiError(
        errorData?.error ||
          errorData?.message ||
          `API error: ${response.status}`,
        response.status,
        errorData,
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

export interface WordTiming {
  word: string
  start: number
  end: number
}

export interface WordTimings {
  words: WordTiming[]
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
  wordTimings?: WordTimings
  wordTimingsStatus?: 'none' | 'pending' | 'complete' | 'failed'
  wordTimingsError?: string
}

export interface Thread {
  id: string
  name?: string | null
  archivedAt?: string | null
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

export async function createThread(
  request?: CreateThreadRequest,
): Promise<Thread> {
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
            console.error(
              'Failed to fetch audio URL for message:',
              message.id,
              error,
            )
            // Keep the original key if fetching fails
          }
        }
      }),
    )
  }

  return thread
}

export async function sendMessage(
  threadId: string,
  content: string,
): Promise<Message> {
  return callAPI<Message>(`/api/threads/${threadId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}

export async function getThreads(): Promise<Thread[]> {
  return callAPI<Thread[]>('/api/threads')
}

export async function getArchivedThreads(): Promise<Thread[]> {
  return callAPI<Thread[]>('/api/threads/archived')
}

export async function updateThread(
  threadId: string,
  data: { name?: string | null },
): Promise<Thread> {
  return callAPI<Thread>(`/api/threads/${threadId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export async function deleteThread(threadId: string): Promise<void> {
  await callAPI<{ message: string }>(`/api/threads/${threadId}`, {
    method: 'DELETE',
  })
}

export async function archiveThread(threadId: string): Promise<Thread> {
  return callAPI<Thread>(`/api/threads/${threadId}/archive`, {
    method: 'POST',
  })
}

export async function unarchiveThread(threadId: string): Promise<Thread> {
  return callAPI<Thread>(`/api/threads/${threadId}/unarchive`, {
    method: 'POST',
  })
}

export interface SendAudioMessageResponse {
  userMessage: Message
  assistantMessage: Message
}

export async function sendAudioMessage(
  threadId: string,
  audioBlob: Blob,
): Promise<SendAudioMessageResponse> {
  const formData = new FormData()
  formData.append('audio', audioBlob, 'recording.webm')

  const url = `${API_BASE_URL}/api/threads/${threadId}/messages/audio`

  try {
    const response = await fetch(url, {
      method: 'POST',
      body: formData,
      credentials: 'include', // Send cookies with requests
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ApiError(
        errorData?.error || `API error: ${response.status}`,
        response.status,
        errorData,
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
        data.assistantMessage.audioUrl = await getAudioUrl(
          data.assistantMessage.audioUrl,
        )
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

// ============================================
// Auth API
// ============================================

export interface User {
  id: string
  email: string
  name: string
  avatarUrl?: string
  emailVerified: boolean
}

interface RegisterRequest {
  email: string
  password: string
  name: string
}

interface LoginRequest {
  email: string
  password: string
}

export async function register(data: RegisterRequest): Promise<User> {
  return callAPI<User>('/api/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function login(data: LoginRequest): Promise<User> {
  return callAPI<User>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function logout(): Promise<void> {
  await callAPI<{ message: string }>('/api/auth/logout', {
    method: 'POST',
  })
}

export async function getCurrentUser(): Promise<User | null> {
  try {
    return await callAPI<User>('/api/auth/me')
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null // Not authenticated
    }
    throw error
  }
}

// ============================================
// Subscription & Credits API
// ============================================

export type SubscriptionTier = 'free' | 'basic' | 'pro'

export interface Subscription {
  id: string
  tier: SubscriptionTier
  status: 'active' | 'canceled' | 'past_due'
  currentPeriodStart?: string
  currentPeriodEnd?: string
  cancelAtPeriodEnd: boolean
}

export interface Credits {
  id: string
  balance: number
  monthlyAllowance: number
  usedThisPeriod: number
  lastRefreshedAt: string
}

export interface SubscriptionWithCredits {
  subscription: Subscription
  credits: Credits
}

export interface CreditTransaction {
  id: string
  type: 'debit' | 'credit' | 'refresh'
  amount: number
  balanceAfter: number
  reference?: string
  description: string
  createdAt: string
}

export interface CheckoutResponse {
  url: string
}

export interface PortalResponse {
  url: string
}

export async function getSubscriptionStatus(): Promise<SubscriptionWithCredits> {
  return callAPI<SubscriptionWithCredits>('/api/subscription')
}

export async function getCredits(): Promise<Credits> {
  return callAPI<Credits>('/api/credits')
}

export async function getCreditHistory(): Promise<{
  transactions: CreditTransaction[]
}> {
  return callAPI<{ transactions: CreditTransaction[] }>('/api/credits/history')
}

export async function createCheckoutSession(
  tier: 'basic' | 'pro',
): Promise<CheckoutResponse> {
  return callAPI<CheckoutResponse>('/api/subscription/checkout', {
    method: 'POST',
    body: JSON.stringify({ tier }),
  })
}

export async function createPortalSession(): Promise<PortalResponse> {
  return callAPI<PortalResponse>('/api/subscription/portal', {
    method: 'POST',
  })
}

// Credit costs (should match backend)
export const CREDIT_COSTS = {
  textMessage: 1,
  audioMessage: 5,
} as const

// Tier info for display
export const TIER_INFO = {
  free: { name: 'Free', price: 0, credits: 20 },
  basic: { name: 'Basic', price: 20, credits: 400 },
  pro: { name: 'Pro', price: 50, credits: 1200 },
} as const

// Helper to check if an error is an insufficient credits error
export function isInsufficientCreditsError(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    error.status === 402 &&
    (error.data as { code?: string })?.code === 'INSUFFICIENT_CREDITS'
  )
}

// ============================================
// Pronunciation Stats API
// ============================================

export interface PhonemeAccuracy {
  phoneme: string
  totalAttempts: number
  correctCount: number
  deletionCount: number
  accuracy: number
}

export interface SubstitutionPattern {
  expectedPhoneme: string
  actualPhoneme: string
  count: number
}

export interface PhonemeStatsResponse {
  totalPhonemes: number
  overallAccuracy: number
  phonemeStats: PhonemeAccuracy[]
  commonSubstitutions: SubstitutionPattern[]
}

export async function getPhonemeStats(): Promise<PhonemeStatsResponse> {
  return callAPI<PhonemeStatsResponse>('/api/pronunciation/stats')
}

export { ApiError }
