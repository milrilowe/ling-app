import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Loader2, ChevronLeft, ChevronRight, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { usePhonemeStats } from '@/hooks/use-phoneme-stats'
import { PhonemeGrid } from './components/PhonemeGrid'
import { CATEGORY_ORDER, CATEGORY_LABELS } from '@/data/phonemes'

export function PronunciationDashboard() {
  const { data: stats, isLoading, error } = usePhonemeStats()
  const [activeIndex, setActiveIndex] = useState(0)

  const activeCategory = CATEGORY_ORDER[activeIndex]

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground">
        Failed to load pronunciation stats
      </div>
    )
  }

  const goToPrev = () => {
    setActiveIndex((prev) => (prev - 1 + CATEGORY_ORDER.length) % CATEGORY_ORDER.length)
  }

  const goToNext = () => {
    setActiveIndex((prev) => (prev + 1) % CATEGORY_ORDER.length)
  }

  return (
    <>
      {/* Mobile: Carousel layout */}
      <div className="h-full flex items-center justify-center p-6 md:hidden">
        <Button
          variant="ghost"
          size="icon"
          onClick={goToPrev}
          className="shrink-0 h-16 w-16"
        >
          <ChevronLeft className="h-10 w-10" />
        </Button>

        <div className="flex-1 flex flex-col items-center h-full">
          <h2 className="text-3xl font-semibold shrink-0 pt-2">{CATEGORY_LABELS[activeCategory]}</h2>
          <div className="flex-1 flex items-center justify-center">
            <PhonemeGrid activeCategory={activeCategory} />
          </div>
          <div className="flex justify-center gap-2 shrink-0 pb-2">
            {CATEGORY_ORDER.map((category, index) => (
              <button
                key={category}
                onClick={() => setActiveIndex(index)}
                className={cn(
                  'w-2.5 h-2.5 rounded-full transition-all',
                  activeIndex === index
                    ? 'bg-foreground scale-110'
                    : 'bg-muted-foreground/30 hover:bg-muted-foreground/50'
                )}
                title={CATEGORY_LABELS[category]}
              />
            ))}
          </div>
        </div>

        <Button
          variant="ghost"
          size="icon"
          onClick={goToNext}
          className="shrink-0 h-16 w-16"
        >
          <ChevronRight className="h-10 w-10" />
        </Button>
      </div>

      {/* Desktop: Scrollable grid with all categories */}
      <div className="hidden md:block h-full overflow-y-auto">
        {/* Header with title and overall stats */}
        <div className="sticky top-0 bg-background py-4 px-4 border-b mb-6 z-10">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Link to="/">
                <Button variant="ghost" size="icon" className="h-10 w-10">
                  <X className="h-5 w-5" />
                </Button>
              </Link>
              <div className="flex items-center gap-3">
                <h1 className="text-2xl font-semibold">Pronunciation</h1>
                <span className="text-muted-foreground">·</span>
                <button className="text-lg font-medium text-muted-foreground hover:text-foreground transition-colors">
                  English (US)
                </button>
              </div>
            </div>
            {stats && stats.totalPhonemes > 0 && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-muted-foreground">Overall Accuracy</span>
                <div className="w-32 h-2 rounded-full bg-muted overflow-hidden">
                  <div
                    className={cn(
                      'h-full rounded-full',
                      stats.overallAccuracy >= 80 ? 'bg-emerald-500' :
                      stats.overallAccuracy >= 60 ? 'bg-yellow-500' :
                      stats.overallAccuracy >= 40 ? 'bg-orange-500' :
                      'bg-red-500'
                    )}
                    style={{ width: `${stats.overallAccuracy}%` }}
                  />
                </div>
                <span className={cn(
                  'text-sm font-semibold',
                  stats.overallAccuracy >= 80 ? 'text-emerald-500' :
                  stats.overallAccuracy >= 60 ? 'text-yellow-500' :
                  stats.overallAccuracy >= 40 ? 'text-orange-500' :
                  'text-red-500'
                )}>
                  {stats.overallAccuracy.toFixed(0)}%
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="divide-y divide-border px-8 pb-8">
          {CATEGORY_ORDER.map((category) => (
            <section key={category} className="py-6 first:pt-0">
              <h2 className="text-sm font-medium mb-4 text-muted-foreground uppercase tracking-wider">
                {CATEGORY_LABELS[category]}
              </h2>
              <PhonemeGrid activeCategory={category} compact />
            </section>
          ))}
        </div>
      </div>
    </>
  )
}
