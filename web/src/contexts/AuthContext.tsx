import { createContext, useContext, type ReactNode } from 'react'
import { useCurrentUser, useLogin, useLogout, useRegister } from '@/hooks/use-auth'
import type { User } from '@/lib/api'

interface AuthContextType {
  // Current user state
  user: User | null | undefined
  isLoading: boolean
  isAuthenticated: boolean
  error: Error | null

  // Auth actions
  login: ReturnType<typeof useLogin>
  register: ReturnType<typeof useRegister>
  logout: ReturnType<typeof useLogout>

  // Refetch user (useful after OAuth callback)
  refetchUser: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
}

/**
 * AuthProvider wraps the app and provides auth state + actions.
 *
 * Usage:
 * - user: The current user or null if not authenticated
 * - isLoading: True while checking initial auth state
 * - isAuthenticated: True if user is logged in
 * - login.mutate({ email, password }): Log in
 * - register.mutate({ email, password, name }): Sign up
 * - logout.mutate(): Log out
 */
export function AuthProvider({ children }: AuthProviderProps) {
  const userQuery = useCurrentUser()
  const loginMutation = useLogin()
  const registerMutation = useRegister()
  const logoutMutation = useLogout()

  const value: AuthContextType = {
    user: userQuery.data,
    isLoading: userQuery.isLoading,
    isAuthenticated: !!userQuery.data,
    error: userQuery.error,

    login: loginMutation,
    register: registerMutation,
    logout: logoutMutation,

    refetchUser: () => userQuery.refetch(),
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

/**
 * Hook to access auth context.
 * Must be used within an AuthProvider.
 */
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
