import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getCurrentUser,
  login,
  logout,
  register,
  type User,
} from '@/lib/api'

// Query keys for auth - exported for use in route guards
export const authKeys = {
  user: ['auth', 'user'] as const,
}

/**
 * Hook to get the current authenticated user.
 * Returns null if not authenticated.
 *
 * Key behaviors:
 * - retry: false - Don't retry 401s, just return null
 * - staleTime: 5 min - User data doesn't change often
 * - refetchOnWindowFocus: true - Check auth when user returns to tab
 */
export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.user,
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook for user registration.
 * On success, sets the user in cache (since backend returns user + sets cookie).
 */
export function useRegister() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: register,
    onSuccess: (user: User) => {
      // Set user in cache - the cookie is set by the backend
      queryClient.setQueryData(authKeys.user, user)
    },
  })
}

/**
 * Hook for user login.
 * On success, sets the user in cache.
 */
export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: login,
    onSuccess: (user: User) => {
      // Set user in cache - the cookie is set by the backend
      queryClient.setQueryData(authKeys.user, user)
    },
  })
}

/**
 * Hook for logout.
 * On success, clears the user from cache and invalidates all queries.
 */
export function useLogout() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: logout,
    onSuccess: () => {
      // Clear user from cache
      queryClient.setQueryData(authKeys.user, null)
      // Clear all cached data (threads, etc.) since they're user-specific
      queryClient.clear()
    },
  })
}
