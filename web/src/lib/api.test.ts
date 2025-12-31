import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import {
  ApiError,
  isInsufficientCreditsError,
  TIER_INFO,
  CREDIT_COSTS,
} from './api'

describe('ApiError', () => {
  it('creates error with message and status', () => {
    const error = new ApiError('Not found', 404)
    expect(error.message).toBe('Not found')
    expect(error.status).toBe(404)
    expect(error.name).toBe('ApiError')
  })

  it('includes optional data', () => {
    const data = { code: 'NOT_FOUND', details: 'Resource not found' }
    const error = new ApiError('Not found', 404, data)
    expect(error.data).toEqual(data)
  })
})

describe('isInsufficientCreditsError', () => {
  it('returns true for 402 with INSUFFICIENT_CREDITS code', () => {
    const error = new ApiError('Insufficient credits', 402, {
      code: 'INSUFFICIENT_CREDITS',
    })
    expect(isInsufficientCreditsError(error)).toBe(true)
  })

  it('returns false for other 402 errors', () => {
    const error = new ApiError('Payment required', 402, { code: 'OTHER' })
    expect(isInsufficientCreditsError(error)).toBe(false)
  })

  it('returns false for non-ApiError', () => {
    expect(isInsufficientCreditsError(new Error('test'))).toBe(false)
    expect(isInsufficientCreditsError(null)).toBe(false)
    expect(isInsufficientCreditsError(undefined)).toBe(false)
  })
})

describe('Constants', () => {
  it('has correct tier info', () => {
    expect(TIER_INFO.free.credits).toBe(20)
    expect(TIER_INFO.basic.credits).toBe(400)
    expect(TIER_INFO.pro.credits).toBe(1200)
  })

  it('has credit costs defined', () => {
    expect(CREDIT_COSTS.textMessage).toBe(1)
    expect(CREDIT_COSTS.audioMessage).toBe(5)
  })
})

describe('API functions', () => {
  const mockFetch = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    mockFetch.mockReset()
  })

  it('handles successful responses', async () => {
    const mockUser = { id: '1', email: 'test@example.com', name: 'Test' }
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockUser),
    })

    // Import dynamically to use mocked fetch
    const { getCurrentUser } = await import('./api')
    const result = await getCurrentUser()
    expect(result).toEqual(mockUser)
  })

  it('returns null for 401 on getCurrentUser', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'Unauthorized' }),
    })

    const { getCurrentUser } = await import('./api')
    const result = await getCurrentUser()
    expect(result).toBeNull()
  })

  it('throws ApiError on non-401 errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: 'Server error' }),
    })

    const { getThreads } = await import('./api')
    await expect(getThreads()).rejects.toThrow(ApiError)
  })
})
