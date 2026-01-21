import { useState, useMemo } from 'react'
import {
  ENGLISH_PHONEMES,
  type CanonicalPhoneme,
  type PhonemeCategory,
} from '@/data/phonemes'
import { usePhonemeStats } from '@/hooks/use-phoneme-stats'
import type { PhonemeAccuracy } from '@/lib/api'
import { PhonemeCard } from './PhonemeCard'
import { PhonemeDetailDialog } from './PhonemeDetailDialog'

interface PhonemeWithStats extends CanonicalPhoneme {
  accuracy: number | null
  totalAttempts: number
  correctCount: number
  deletionCount: number
}

interface PhonemeGridProps {
  activeCategory: PhonemeCategory
  compact?: boolean
}

export function PhonemeGrid({ activeCategory, compact = false }: PhonemeGridProps) {
  const { data: stats } = usePhonemeStats()
  const [selectedPhoneme, setSelectedPhoneme] = useState<PhonemeWithStats | null>(null)

  // Merge canonical phoneme list with user stats
  const phonemesWithStats = useMemo(() => {
    return ENGLISH_PHONEMES.map((phoneme) => {
      const userStat = stats?.phonemeStats?.find((s) => s.phoneme === phoneme.ipa)
      return {
        ...phoneme,
        accuracy: userStat?.accuracy ?? null,
        totalAttempts: userStat?.totalAttempts ?? 0,
        correctCount: userStat?.correctCount ?? 0,
        deletionCount: userStat?.deletionCount ?? 0,
      }
    })
  }, [stats?.phonemeStats])

  // Filter by active category
  const filteredPhonemes = useMemo(() => {
    return phonemesWithStats.filter((p) => p.category === activeCategory)
  }, [phonemesWithStats, activeCategory])

  // Get substitutions for selected phoneme
  const selectedSubstitutions = useMemo(() => {
    if (!selectedPhoneme || !stats?.commonSubstitutions) return []
    return stats.commonSubstitutions.filter(
      (sub) => sub.expectedPhoneme === selectedPhoneme.ipa
    )
  }, [selectedPhoneme, stats?.commonSubstitutions])

  // Get stats for selected phoneme
  const selectedStats = useMemo((): PhonemeAccuracy | null => {
    if (!selectedPhoneme || selectedPhoneme.accuracy === null) return null
    return {
      phoneme: selectedPhoneme.ipa,
      totalAttempts: selectedPhoneme.totalAttempts,
      correctCount: selectedPhoneme.correctCount,
      deletionCount: selectedPhoneme.deletionCount,
      accuracy: selectedPhoneme.accuracy,
    }
  }, [selectedPhoneme])

  return (
    <>
      <div>
        <div className={`flex flex-wrap ${compact ? 'gap-2 justify-start' : 'gap-3 justify-center'}`}>
          {filteredPhonemes.map((phoneme) => (
            <PhonemeCard
              key={phoneme.ipa}
              ipa={phoneme.ipa}
              example={phoneme.example}
              accuracy={phoneme.accuracy}
              onClick={() => setSelectedPhoneme(phoneme)}
              compact={compact}
            />
          ))}
        </div>
      </div>

      <PhonemeDetailDialog
        phoneme={selectedPhoneme}
        stats={selectedStats}
        substitutions={selectedSubstitutions}
        open={!!selectedPhoneme}
        onOpenChange={(open) => {
          if (!open) setSelectedPhoneme(null)
        }}
      />
    </>
  )
}
