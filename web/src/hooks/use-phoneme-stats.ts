import { useQuery } from '@tanstack/react-query'
import { getPhonemeStats } from '@/lib/api'

export const phonemeStatsKeys = {
  all: ['phonemeStats'] as const,
  stats: () => [...phonemeStatsKeys.all, 'stats'] as const,
}

export function usePhonemeStats() {
  return useQuery({
    queryKey: phonemeStatsKeys.stats(),
    queryFn: getPhonemeStats,
    staleTime: 60 * 1000, // 1 minute
  })
}
