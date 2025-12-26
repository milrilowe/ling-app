import { redirect } from '@tanstack/react-router'
import { getCurrentUser } from '@/lib/api'
import { authKeys } from '@/hooks/use-auth'
import type { QueryClient } from '@tanstack/react-query'

interface BeforeLoadContext {
  context: {
    queryClient: QueryClient
  }
  location: {
    pathname: string
  }
}

/**
 * Route guard that requires authentication.
 * Use this in a route's beforeLoad to protect it.
 *
 * Usage:
 * ```ts
 * export const Route = createFileRoute('/protected')({
 *   beforeLoad: requireAuth,
 *   component: ProtectedPage,
 * })
 * ```
 */
export async function requireAuth({ context, location }: BeforeLoadContext) {
  // Try to get cached user first, otherwise fetch
  const user = await context.queryClient.ensureQueryData({
    queryKey: authKeys.user,
    queryFn: getCurrentUser,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  if (!user) {
    // Redirect to login with the intended destination
    throw redirect({
      to: '/login',
      search: {
        redirect: location.pathname,
      },
    })
  }

  return { user }
}

/**
 * Route guard that redirects authenticated users away.
 * Useful for login/register pages - redirect to home if already logged in.
 *
 * Usage:
 * ```ts
 * export const Route = createFileRoute('/login')({
 *   beforeLoad: redirectIfAuthenticated,
 *   component: LoginPage,
 * })
 * ```
 */
export async function redirectIfAuthenticated({ context }: BeforeLoadContext) {
  const user = await context.queryClient.ensureQueryData({
    queryKey: authKeys.user,
    queryFn: getCurrentUser,
    staleTime: 5 * 60 * 1000,
  })

  if (user) {
    throw redirect({
      to: '/',
    })
  }
}
