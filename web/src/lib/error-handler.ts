import { toast } from 'sonner'

// Gin validation errors look like: "Key: 'Field' Error:Field validation..."
const isGinValidationError = (msg: string) => msg.startsWith('Key:')

// Duck type check for ApiError (has status and data properties)
const isApiError = (
  error: unknown,
): error is Error & { status: number; data?: unknown } =>
  error instanceof Error &&
  'status' in error &&
  typeof (error as { status: unknown }).status === 'number'

export function handleError(error: unknown, context?: string) {
  console.error(context ? `${context}:` : 'Error:', error)

  let message = 'An unexpected error occurred'

  if (isApiError(error)) {
    const data = error.data as { error?: string; message?: string } | undefined
    const rawMessage = data?.error || data?.message || error.message

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
