import { useState, useMemo } from 'react'
import {
  ENGLISH_PHONEMES,
  CATEGORY_LABELS,
  CATEGORY_ORDER,
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
}

export function PhonemeGrid() {
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
      }
    })
  }, [stats?.phonemeStats])

  // Group phonemes by category
  const groupedPhonemes = useMemo(() => {
    const grouped = new Map<PhonemeCategory, PhonemeWithStats[]>()
    for (const category of CATEGORY_ORDER) {
      grouped.set(category, [])
    }
    for (const phoneme of phonemesWithStats) {
      const list = grouped.get(phoneme.category)
      if (list) {
        list.push(phoneme)
      }
    }
    return grouped
  }, [phonemesWithStats])

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
      accuracy: selectedPhoneme.accuracy,
    }
  }, [selectedPhoneme])

  return (
    <>
      <div className="space-y-6">
        {CATEGORY_ORDER.map((category) => {
          const phonemes = groupedPhonemes.get(category)
          if (!phonemes || phonemes.length === 0) return null

          return (
            <div key={category}>
              <h3 className="text-sm font-medium text-muted-foreground mb-2">
                {CATEGORY_LABELS[category]}
              </h3>
              <div className="grid grid-cols-5 gap-2">
                {phonemes.map((phoneme) => (
                  <PhonemeCard
                    key={phoneme.ipa}
                    ipa={phoneme.ipa}
                    example={phoneme.example}
                    accuracy={phoneme.accuracy}
                    onClick={() => setSelectedPhoneme(phoneme)}
                  />
                ))}
              </div>
            </div>
          )
        })}
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
