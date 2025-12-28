import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
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

interface BreakdownItem {
  label: string
  count: number
  color: string
}

function BreakdownBar({ items, total }: { items: BreakdownItem[]; total: number }) {
  if (total === 0) return null

  return (
    <div className="space-y-2">
      {items.map((item, idx) => {
        const percentage = (item.count / total) * 100
        if (item.count === 0) return null

        return (
          <div key={idx} className="space-y-1">
            <div className="flex items-center justify-between text-sm">
              <span className="font-mono">{item.label}</span>
              <span className="text-muted-foreground">
                {item.count} ({percentage.toFixed(0)}%)
              </span>
            </div>
            <div className="h-2 w-full rounded-full bg-muted">
              <div
                className={cn('h-full rounded-full transition-all', item.color)}
                style={{ width: `${percentage}%` }}
              />
            </div>
          </div>
        )
      })}
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
    if (accuracy >= 80) return 'text-green-600 dark:text-green-400'
    if (accuracy >= 60) return 'text-yellow-600 dark:text-yellow-400'
    if (accuracy >= 40) return 'text-orange-600 dark:text-orange-400'
    return 'text-red-600 dark:text-red-400'
  }

  // Build breakdown items
  const buildBreakdown = (): BreakdownItem[] => {
    if (!stats) return []

    const items: BreakdownItem[] = [
      {
        label: 'Correct',
        count: stats.correctCount,
        color: 'bg-green-500',
      },
    ]

    // Add deletions if any
    if (stats.deletionCount > 0) {
      items.push({
        label: 'Skipped',
        count: stats.deletionCount,
        color: 'bg-orange-500',
      })
    }

    // Add each substitution
    for (const sub of substitutions) {
      items.push({
        label: `→ /${sub.actualPhoneme}/`,
        count: sub.count,
        color: 'bg-red-400',
      })
    }

    return items
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="font-mono text-4xl text-center flex items-center justify-center gap-2">
            /{phoneme.ipa}/
            <Button
              variant="ghost"
              size="icon"
              onClick={playExampleAudio}
              className="h-10 w-10"
              title={`Hear "${phoneme.example}"`}
            >
              <Volume2 className="h-5 w-5" />
            </Button>
          </DialogTitle>
          <DialogDescription className="text-center text-base">
            as in "{exampleParts.before}
            <span className="font-bold uppercase">{exampleParts.highlighted}</span>
            {exampleParts.after}"
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {stats ? (
            <>
              {/* Accuracy display */}
              <div className="text-center">
                <div className={cn('text-5xl font-bold', getAccuracyColor(stats.accuracy))}>
                  {stats.accuracy.toFixed(0)}%
                </div>
                <div className="text-sm text-muted-foreground mt-1">
                  {stats.totalAttempts} attempts
                </div>
              </div>

              {/* Breakdown bars */}
              <BreakdownBar items={buildBreakdown()} total={stats.totalAttempts} />
            </>
          ) : (
            <div className="text-center py-8">
              <div className="text-4xl mb-3">🎯</div>
              <p className="text-muted-foreground">
                You haven't practiced this sound yet!
              </p>
              <p className="text-sm text-muted-foreground mt-1">
                Keep practicing to see your stats.
              </p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
