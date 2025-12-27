import { useState } from 'react'
import { Zap, Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useCreateCheckout, useCredits } from '@/hooks/use-subscription'
import { TIER_INFO } from '@/lib/api'

interface UpgradePromptProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  creditsNeeded?: number
}

export function UpgradePrompt({ open, onOpenChange, creditsNeeded }: UpgradePromptProps) {
  const { data: credits } = useCredits()
  const createCheckout = useCreateCheckout()
  const [selectedTier, setSelectedTier] = useState<'basic' | 'pro' | null>(null)

  const handleUpgrade = (tier: 'basic' | 'pro') => {
    setSelectedTier(tier)
    createCheckout.mutate(tier)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-yellow-500" />
            Out of Credits
          </DialogTitle>
          <DialogDescription>
            {creditsNeeded ? (
              <>You need {creditsNeeded} credits but only have {credits?.balance ?? 0}.</>
            ) : (
              <>You've used all your credits for this period.</>
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <p className="text-sm text-muted-foreground">
            Upgrade your plan to get more credits and continue learning.
          </p>

          <div className="grid gap-3">
            {/* Basic Plan */}
            <div className="flex items-center justify-between p-4 rounded-lg border bg-card">
              <div>
                <h4 className="font-semibold">{TIER_INFO.basic.name}</h4>
                <p className="text-sm text-muted-foreground">
                  {TIER_INFO.basic.credits} credits/month
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className="font-semibold">${TIER_INFO.basic.price}/mo</span>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleUpgrade('basic')}
                  disabled={createCheckout.isPending}
                >
                  {createCheckout.isPending && selectedTier === 'basic' ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    'Upgrade'
                  )}
                </Button>
              </div>
            </div>

            {/* Pro Plan */}
            <div className="flex items-center justify-between p-4 rounded-lg border bg-card border-primary">
              <div>
                <h4 className="font-semibold flex items-center gap-2">
                  {TIER_INFO.pro.name}
                  <span className="text-xs bg-primary text-primary-foreground px-2 py-0.5 rounded-full">
                    Best Value
                  </span>
                </h4>
                <p className="text-sm text-muted-foreground">
                  {TIER_INFO.pro.credits} credits/month
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className="font-semibold">${TIER_INFO.pro.price}/mo</span>
                <Button
                  size="sm"
                  onClick={() => handleUpgrade('pro')}
                  disabled={createCheckout.isPending}
                >
                  {createCheckout.isPending && selectedTier === 'pro' ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    'Upgrade'
                  )}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
