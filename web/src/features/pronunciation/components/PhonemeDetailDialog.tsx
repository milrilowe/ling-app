import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { type CanonicalPhoneme, renderExample } from '@/data/phonemes'
import type { PhonemeAccuracy, SubstitutionPattern } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Volume2 } from 'lucide-react'

interface PhonemeDetailDialogProps {
  phoneme: CanonicalPhoneme | null
  stats: PhonemeAccuracy | null
  substitutions: SubstitutionPattern[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface MistakeItem {
  label: string
  count: number
  percentage: number
}

function MistakeCard({ item }: { item: MistakeItem }) {
  return (
    <div className="rounded-lg border bg-card p-3 space-y-2">
      <div className="flex items-center justify-between">
        <span className="font-mono text-sm font-medium">{item.label}</span>
        <span className="text-xs text-muted-foreground tabular-nums">
          {item.count}x
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-muted/50 overflow-hidden">
        <div
          className="h-full rounded-full bg-red-400"
          style={{ width: `${item.percentage}%` }}
        />
      </div>
    </div>
  )
}

export function PhonemeDetailDialog({
  phoneme,
  stats,
  substitutions,
  open,
  onOpenChange,
}: PhonemeDetailDialogProps) {
  if (!phoneme) return null

  const playExampleAudio = () => {
    const audio = new Audio(`/audio/phonemes/${phoneme.example}.mp3`)
    audio.play().catch((err) => {
      console.error('Failed to play audio:', err)
    })
  }

  const exampleParts = renderExample(phoneme.example, phoneme.highlight)

  const getAccuracyColor = (accuracy: number) => {
    if (accuracy >= 80) return 'text-emerald-500'
    if (accuracy >= 60) return 'text-yellow-500'
    if (accuracy >= 40) return 'text-orange-500'
    return 'text-red-500'
  }

  const getAccuracyBarColor = (accuracy: number) => {
    if (accuracy >= 80) return 'bg-emerald-500'
    if (accuracy >= 60) return 'bg-yellow-500'
    if (accuracy >= 40) return 'bg-orange-500'
    return 'bg-red-500'
  }

  // Build mistakes list (deletions + substitutions)
  const buildMistakes = (): MistakeItem[] => {
    if (!stats) return []

    const items: MistakeItem[] = []

    // Add deletions if any
    if (stats.deletionCount > 0) {
      items.push({
        label: 'Skipped',
        count: stats.deletionCount,
        percentage: (stats.deletionCount / stats.totalAttempts) * 100,
      })
    }

    // Add each substitution
    for (const sub of substitutions) {
      items.push({
        label: `→ /${sub.actualPhoneme}/`,
        count: sub.count,
        percentage: (sub.count / stats.totalAttempts) * 100,
      })
    }

    // Sort by count descending
    items.sort((a, b) => b.count - a.count)

    return items
  }

  const mistakes = buildMistakes()
  const hasMistakes = mistakes.length > 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm p-0 gap-0 overflow-hidden flex flex-col h-[420px]">
        {/* Phoneme header - centered */}
        <div className="px-6 pt-6 pb-4 text-center shrink-0">
          <div className="inline-flex items-center gap-2">
            <span className="font-mono text-4xl text-center">
              /{phoneme.ipa}/
            </span>
            <Button
              variant="ghost"
              size="icon"
              onClick={playExampleAudio}
              className="h-10 w-10"
              title={`Hear "${phoneme.example}"`}
            >
              <Volume2 className="h-5 w-5" />
            </Button>
          </div>
          <p className="text-sm text-muted-foreground mt-2 lowercase">
            as in "{exampleParts.before}<span className="font-bold text-foreground">{exampleParts.highlighted}</span>{exampleParts.after}"
          </p>
        </div>

        {/* Accuracy section */}
        {stats && (
          <div className="px-6 pb-4 border-b border-border/50 shrink-0">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-muted-foreground">Accuracy</span>
              <span className={cn('text-xl font-semibold tabular-nums', getAccuracyColor(stats.accuracy))}>
                {stats.accuracy.toFixed(0)}%
              </span>
            </div>
            <div className="h-2 w-full rounded-full bg-muted/50 overflow-hidden">
              <div
                className={cn('h-full rounded-full transition-all', getAccuracyBarColor(stats.accuracy))}
                style={{ width: `${stats.accuracy}%` }}
              />
            </div>
            <p className="text-xs text-muted-foreground mt-1.5 text-right">
              {stats.correctCount}/{stats.totalAttempts} correct
            </p>
          </div>
        )}

        {/* Mistakes list - scrollable */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {stats ? (
            hasMistakes ? (
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground uppercase tracking-wider mb-3">
                  Mistakes
                </p>
                {mistakes.map((item, idx) => (
                  <MistakeCard key={idx} item={item} />
                ))}
              </div>
            ) : (
              <div className="h-full flex items-center justify-center">
                <p className="text-sm text-muted-foreground">
                  No mistakes recorded
                </p>
              </div>
            )
          ) : (
            <div className="h-full flex items-center justify-center">
              <div className="text-center">
                <div className="w-12 h-12 rounded-full bg-muted/50 flex items-center justify-center mx-auto mb-3">
                  <Volume2 className="h-5 w-5 text-muted-foreground" />
                </div>
                <p className="text-sm text-muted-foreground">
                  No practice data yet
                </p>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
