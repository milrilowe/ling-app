import { useEffect } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/contexts/AuthContext'

export const Route = createFileRoute('/auth/callback')({
  component: AuthCallbackPage,
})

/**
 * OAuth callback page.
 * After successful OAuth login, the backend redirects here.
 * The session cookie is already set, so we just need to refresh the user
 * and redirect to the home page.
 */
function AuthCallbackPage() {
  const navigate = useNavigate()
  const { refetchUser } = useAuth()

  useEffect(() => {
    // Refetch user data (cookie should be set by OAuth callback)
    refetchUser()
    // Redirect to home
    navigate({ to: '/' })
  }, [refetchUser, navigate])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <div className="mb-4 h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto" />
        <p className="text-muted-foreground">Completing sign in...</p>
      </div>
    </div>
  )
}
