import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { type CanonicalPhoneme, renderExample, ENGLISH_PHONEMES } from '@/data/phonemes'
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

function AccuracyBar({ accuracy }: { accuracy: number }) {
  const getColorClass = (acc: number) => {
    if (acc >= 80) return 'bg-green-500'
    if (acc >= 60) return 'bg-yellow-500'
    if (acc >= 40) return 'bg-orange-500'
    return 'bg-red-500'
  }

  return (
    <div className="h-3 w-full rounded-full bg-muted">
      <div
        className={cn('h-full rounded-full transition-all', getColorClass(accuracy))}
        style={{ width: `${Math.min(100, Math.max(0, accuracy))}%` }}
      />
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

  const playPhonemeAudio = (ipa: string) => {
    const phonemeData = ENGLISH_PHONEMES.find((p) => p.ipa === ipa)
    if (phonemeData) {
      const audio = new Audio(`/audio/phonemes/${phonemeData.example}.mp3`)
      audio.play().catch((err) => {
        console.error('Failed to play audio:', err)
      })
    }
  }

  const exampleParts = renderExample(phoneme.example, phoneme.highlight)

  const getAccuracyColor = (accuracy: number) => {
    if (accuracy >= 80) return 'text-green-600 dark:text-green-400'
    if (accuracy >= 60) return 'text-yellow-600 dark:text-yellow-400'
    if (accuracy >= 40) return 'text-orange-600 dark:text-orange-400'
    return 'text-red-600 dark:text-red-400'
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
                  {stats.accuracy.toFixed(1)}%
                </div>
                <div className="text-sm text-muted-foreground mt-1">
                  {stats.correctCount} / {stats.totalAttempts} correct
                </div>
              </div>

              {/* Accuracy bar */}
              <AccuracyBar accuracy={stats.accuracy} />

              {/* Common substitutions for this phoneme */}
              {substitutions.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-medium text-muted-foreground">
                    Common mistakes
                  </h4>
                  <ul className="space-y-2">
                    {substitutions.map((sub, idx) => (
                      <li
                        key={idx}
                        className="flex items-center gap-2 text-sm bg-muted/50 rounded-md px-3 py-2"
                      >
                        <span className="text-muted-foreground">You said</span>
                        <span className="font-mono text-base font-semibold">
                          /{sub.actualPhoneme}/
                        </span>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => playPhonemeAudio(sub.actualPhoneme)}
                          className="h-6 w-6"
                          title="Hear this sound"
                        >
                          <Volume2 className="h-3 w-3" />
                        </Button>
                        <span className="text-muted-foreground">instead</span>
                        <span className="ml-auto text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
                          {sub.count}x
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
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
