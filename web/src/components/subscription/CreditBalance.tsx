import { Coins } from 'lucide-react'
import { useCredits } from '@/hooks/use-subscription'
import { cn } from '@/lib/utils'

interface CreditBalanceProps {
  className?: string
  showLabel?: boolean
}

export function CreditBalance({ className, showLabel = true }: CreditBalanceProps) {
  const { data: credits, isLoading } = useCredits()

  if (isLoading || !credits) {
    return (
      <div className={cn('flex items-center gap-2 text-sm text-muted-foreground', className)}>
        <Coins className="h-4 w-4" />
        <span>...</span>
      </div>
    )
  }

  const percentage = (credits.balance / credits.monthlyAllowance) * 100
  const isLow = percentage < 20
  const isEmpty = credits.balance === 0

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Coins className={cn(
        'h-4 w-4',
        isEmpty ? 'text-destructive' : isLow ? 'text-yellow-500' : 'text-muted-foreground'
      )} />
      <div className="flex flex-col">
        <span className={cn(
          'text-sm font-medium',
          isEmpty ? 'text-destructive' : isLow ? 'text-yellow-500' : ''
        )}>
          {credits.balance} {showLabel && <span className="text-muted-foreground font-normal">/ {credits.monthlyAllowance}</span>}
        </span>
        {showLabel && (
          <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full transition-all',
                isEmpty ? 'bg-destructive' : isLow ? 'bg-yellow-500' : 'bg-primary'
              )}
              style={{ width: `${Math.min(percentage, 100)}%` }}
            />
          </div>
        )}
      </div>
    </div>
  )
}
