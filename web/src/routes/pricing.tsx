import { createFileRoute, Link } from '@tanstack/react-router'
import { Check, Loader2, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { useSubscription, useCreateCheckout } from '@/hooks/use-subscription'
import { TIER_INFO, CREDIT_COSTS, type SubscriptionTier } from '@/lib/api'
import { cn } from '@/lib/utils'
import { useState } from 'react'

export const Route = createFileRoute('/pricing')({
  component: PricingPage,
})

const features = {
  free: [
    `${TIER_INFO.free.credits} credits per month`,
    `${CREDIT_COSTS.textMessage} credit per text message`,
    `${CREDIT_COSTS.audioMessage} credits per audio message`,
    'Basic conversation practice',
  ],
  basic: [
    `${TIER_INFO.basic.credits} credits per month`,
    `${CREDIT_COSTS.textMessage} credit per text message`,
    `${CREDIT_COSTS.audioMessage} credits per audio message`,
    'Pronunciation analysis',
    'Email support',
  ],
  pro: [
    `${TIER_INFO.pro.credits} credits per month`,
    `${CREDIT_COSTS.textMessage} credit per text message`,
    `${CREDIT_COSTS.audioMessage} credits per audio message`,
    'Pronunciation analysis',
    'Priority support',
    'Early access to new features',
  ],
}

function PricingPage() {
  const { data: subscription, isLoading } = useSubscription()
  const createCheckout = useCreateCheckout()
  const [selectedTier, setSelectedTier] = useState<'basic' | 'pro' | null>(null)

  const currentTier = subscription?.subscription?.tier ?? 'free'

  const handleUpgrade = (tier: 'basic' | 'pro') => {
    setSelectedTier(tier)
    createCheckout.mutate(tier)
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="mx-auto max-w-5xl">
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
            <ArrowLeft className="h-4 w-4" />
            Back to app
          </Link>
        </div>

        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-4">Simple, transparent pricing</h1>
          <p className="text-xl text-muted-foreground">
            Choose the plan that works best for your learning goals
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          {/* Free Tier */}
          <PricingCard
            tier="free"
            name={TIER_INFO.free.name}
            price={TIER_INFO.free.price}
            credits={TIER_INFO.free.credits}
            features={features.free}
            isCurrentPlan={currentTier === 'free'}
            isLoading={isLoading}
          />

          {/* Basic Tier */}
          <PricingCard
            tier="basic"
            name={TIER_INFO.basic.name}
            price={TIER_INFO.basic.price}
            credits={TIER_INFO.basic.credits}
            features={features.basic}
            isCurrentPlan={currentTier === 'basic'}
            isLoading={isLoading}
            onUpgrade={() => handleUpgrade('basic')}
            isUpgrading={createCheckout.isPending && selectedTier === 'basic'}
          />

          {/* Pro Tier */}
          <PricingCard
            tier="pro"
            name={TIER_INFO.pro.name}
            price={TIER_INFO.pro.price}
            credits={TIER_INFO.pro.credits}
            features={features.pro}
            isCurrentPlan={currentTier === 'pro'}
            isLoading={isLoading}
            onUpgrade={() => handleUpgrade('pro')}
            isUpgrading={createCheckout.isPending && selectedTier === 'pro'}
            recommended
          />
        </div>

        <div className="mt-12 text-center text-sm text-muted-foreground">
          <p>All plans include access to AI conversation practice.</p>
          <p className="mt-2">
            Need more credits? Upgrade anytime. Cancel anytime.
          </p>
        </div>
      </div>
    </div>
  )
}

interface PricingCardProps {
  tier: SubscriptionTier
  name: string
  price: number
  credits: number
  features: string[]
  isCurrentPlan: boolean
  isLoading: boolean
  onUpgrade?: () => void
  isUpgrading?: boolean
  recommended?: boolean
}

function PricingCard({
  tier,
  name,
  price,
  features,
  isCurrentPlan,
  isLoading,
  onUpgrade,
  isUpgrading,
  recommended,
}: PricingCardProps) {
  return (
    <Card className={cn(
      'relative flex flex-col',
      recommended && 'border-primary shadow-lg',
      isCurrentPlan && 'ring-2 ring-primary'
    )}>
      {recommended && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <span className="bg-primary text-primary-foreground text-xs font-medium px-3 py-1 rounded-full">
            Recommended
          </span>
        </div>
      )}

      <CardHeader>
        <CardTitle>{name}</CardTitle>
        <CardDescription>
          <span className="text-3xl font-bold text-foreground">
            ${price}
          </span>
          {price > 0 && <span className="text-muted-foreground">/month</span>}
        </CardDescription>
      </CardHeader>

      <CardContent className="flex-1">
        <ul className="space-y-3">
          {features.map((feature, index) => (
            <li key={index} className="flex items-start gap-2">
              <Check className="h-4 w-4 text-primary mt-0.5 shrink-0" />
              <span className="text-sm">{feature}</span>
            </li>
          ))}
        </ul>
      </CardContent>

      <CardFooter className="mt-auto">
        {isLoading ? (
          <Button className="w-full" disabled>
            <Loader2 className="h-4 w-4 animate-spin" />
          </Button>
        ) : isCurrentPlan ? (
          <Button className="w-full" variant="outline" disabled>
            Current Plan
          </Button>
        ) : tier === 'free' ? (
          <Button className="w-full" variant="outline" disabled>
            Free
          </Button>
        ) : (
          <Button
            className="w-full"
            variant={recommended ? 'default' : 'outline'}
            onClick={onUpgrade}
            disabled={isUpgrading}
          >
            {isUpgrading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              `Upgrade to ${name}`
            )}
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}
