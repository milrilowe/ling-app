import { toast } from 'sonner'

export function handleError(error: unknown, context?: string) {
  console.error(context ? `${context}:` : 'Error:', error)

  let message = 'An unexpected error occurred'

  if (error instanceof Error) {
    message = error.message
  } else if (typeof error === 'string') {
    message = error
  }

  toast.error(message)
}
