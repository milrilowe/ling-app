import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getSubscriptionStatus,
  getCredits,
  getCreditHistory,
  createCheckoutSession,
  createPortalSession,
} from '@/lib/api'

export const subscriptionKeys = {
  all: ['subscription'] as const,
  status: () => [...subscriptionKeys.all, 'status'] as const,
  credits: () => [...subscriptionKeys.all, 'credits'] as const,
  creditHistory: () => [...subscriptionKeys.all, 'history'] as const,
}

export function useSubscription() {
  return useQuery({
    queryKey: subscriptionKeys.status(),
    queryFn: getSubscriptionStatus,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useCredits() {
  return useQuery({
    queryKey: subscriptionKeys.credits(),
    queryFn: getCredits,
    staleTime: 30 * 1000, // 30 seconds - refresh more frequently
  })
}

export function useCreditHistory() {
  return useQuery({
    queryKey: subscriptionKeys.creditHistory(),
    queryFn: getCreditHistory,
  })
}

export function useCreateCheckout() {
  return useMutation({
    mutationFn: (tier: 'basic' | 'pro') => createCheckoutSession(tier),
    onSuccess: (data) => {
      // Redirect to Stripe checkout
      window.location.href = data.url
    },
    onError: (error) => {
      console.error('Failed to create checkout session:', error)
    },
  })
}

export function useCreatePortal() {
  return useMutation({
    mutationFn: createPortalSession,
    onSuccess: (data) => {
      // Redirect to Stripe customer portal
      window.location.href = data.url
    },
    onError: (error) => {
      console.error('Failed to create portal session:', error)
    },
  })
}

// Hook to invalidate credits after sending a message
export function useInvalidateCredits() {
  const queryClient = useQueryClient()

  return () => {
    queryClient.invalidateQueries({ queryKey: subscriptionKeys.credits() })
    queryClient.invalidateQueries({ queryKey: subscriptionKeys.status() })
  }
}

// Helper hook to check if user has enough credits
export function useHasCredits(amount: number): boolean {
  const { data } = useCredits()
  return (data?.balance ?? 0) >= amount
}
