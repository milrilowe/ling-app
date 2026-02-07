import { toast } from 'sonner'
import { ApiError } from './api'

// Gin validation errors look like: "Key: 'Field' Error:Field validation..."
const isGinValidationError = (msg: string) => msg.startsWith('Key:')

export function handleError(error: unknown, context?: string) {
  console.error(context ? `${context}:` : 'Error:', error)

  let message = 'An unexpected error occurred'

  if (error instanceof ApiError) {
    const data = error.data as { error?: string; message?: string } | undefined
    const rawMessage = data?.error || data?.message || error.message

    // Filter out technical Gin validation errors
    if (!isGinValidationError(rawMessage)) {
      message = rawMessage
    }
  } else if (error instanceof Error) {
    message = error.message
  } else if (typeof error === 'string') {
    message = error
  }

  toast.error(message)
}
