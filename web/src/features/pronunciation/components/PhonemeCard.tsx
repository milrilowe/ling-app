import { cn } from '@/lib/utils'

interface PhonemeCardProps {
  ipa: string
  example: string
  accuracy: number | null // null = never practiced
  onClick: () => void
}

export function PhonemeCard({
  ipa,
  example,
  accuracy,
  onClick,
}: PhonemeCardProps) {
  const getColorClasses = () => {
    if (accuracy === null) {
      return 'bg-muted/50 border-muted-foreground/20 hover:border-muted-foreground/40'
    }
    if (accuracy >= 80) {
      return 'bg-green-500/10 border-green-500/50 hover:border-green-500'
    }
    if (accuracy >= 60) {
      return 'bg-yellow-500/10 border-yellow-500/50 hover:border-yellow-500'
    }
    if (accuracy >= 40) {
      return 'bg-orange-500/10 border-orange-500/50 hover:border-orange-500'
    }
    return 'bg-red-500/10 border-red-500/50 hover:border-red-500'
  }

  const getAccuracyColor = () => {
    if (accuracy === null) return 'text-muted-foreground'
    if (accuracy >= 80) return 'text-green-600 dark:text-green-400'
    if (accuracy >= 60) return 'text-yellow-600 dark:text-yellow-400'
    if (accuracy >= 40) return 'text-orange-600 dark:text-orange-400'
    return 'text-red-600 dark:text-red-400'
  }

  return (
    <button
      onClick={onClick}
      className={cn(
        'flex flex-col items-center justify-center p-3 rounded-lg border-2 transition-all',
        'hover:scale-105 active:scale-95',
        'min-w-[70px] min-h-[80px]',
        getColorClasses()
      )}
    >
      <div className="font-mono text-xl font-semibold">/{ipa}/</div>
      <div className="text-[10px] text-muted-foreground truncate max-w-full">
        {example}
      </div>
      {accuracy !== null ? (
        <div className={cn('text-sm font-bold mt-1', getAccuracyColor())}>
          {accuracy.toFixed(0)}%
        </div>
      ) : (
        <div className="text-xs text-muted-foreground mt-1">--</div>
      )}
    </button>
  )
}
