import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePhonemeStats } from '@/hooks/use-phoneme-stats'
import { PhonemeGrid } from './components/PhonemeGrid'

function AccuracyBar({ accuracy }: { accuracy: number }) {
  const getColorClass = (acc: number) => {
    if (acc >= 80) return 'bg-green-500'
    if (acc >= 60) return 'bg-yellow-500'
    if (acc >= 40) return 'bg-orange-500'
    return 'bg-red-500'
  }

  return (
    <div className="h-2 w-full rounded-full bg-muted">
      <div
        className={cn('h-full rounded-full transition-all', getColorClass(accuracy))}
        style={{ width: `${Math.min(100, Math.max(0, accuracy))}%` }}
      />
    </div>
  )
}

export function PronunciationDashboard() {
  const { data: stats, isLoading, error } = usePhonemeStats()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-center text-muted-foreground">
        Failed to load pronunciation stats
      </div>
    )
  }

  const hasData = stats && stats.totalPhonemes > 0

  return (
    <div className="space-y-6">
      {/* Overall accuracy card */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Overall Accuracy</CardTitle>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <div className="flex items-baseline gap-2">
                <span className="text-3xl font-bold">
                  {stats.overallAccuracy.toFixed(1)}%
                </span>
                <span className="text-sm text-muted-foreground">
                  ({stats.totalPhonemes.toLocaleString()} phonemes analyzed)
                </span>
              </div>
              <AccuracyBar accuracy={stats.overallAccuracy} />
            </>
          ) : (
            <div className="text-muted-foreground">
              <p>No pronunciation data yet.</p>
              <p className="text-sm mt-1">Start practicing to see your stats!</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Phoneme grid */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">All Phonemes</CardTitle>
        </CardHeader>
        <CardContent>
          <PhonemeGrid />
        </CardContent>
      </Card>

      {/* Color legend */}
      <div className="flex flex-wrap gap-4 text-xs text-muted-foreground justify-center">
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-green-500/30 border border-green-500/50" />
          <span>80%+ Mastered</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-yellow-500/30 border border-yellow-500/50" />
          <span>60-79% Needs practice</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-orange-500/30 border border-orange-500/50" />
          <span>40-59% Struggling</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-red-500/30 border border-red-500/50" />
          <span>&lt;40% Focus here</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-muted border border-muted-foreground/20" />
          <span>Not practiced</span>
        </div>
      </div>
    </div>
  )
}
