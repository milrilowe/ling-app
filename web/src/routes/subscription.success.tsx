import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect } from 'react'
import { CheckCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useSubscription, useCredits, subscriptionKeys } from '@/hooks/use-subscription'
import { TIER_INFO } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'

export const Route = createFileRoute('/subscription/success')({
  component: SubscriptionSuccessPage,
})

function SubscriptionSuccessPage() {
  const queryClient = useQueryClient()
  const { data: subscription } = useSubscription()
  const { data: credits } = useCredits()

  // Refresh subscription data on mount
  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: subscriptionKeys.all })
  }, [queryClient])

  const tier = subscription?.subscription?.tier ?? 'free'
  const tierInfo = TIER_INFO[tier]

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 h-12 w-12 rounded-full bg-green-100 dark:bg-green-900 flex items-center justify-center">
            <CheckCircle2 className="h-6 w-6 text-green-600 dark:text-green-400" />
          </div>
          <CardTitle className="text-2xl">Welcome to {tierInfo.name}!</CardTitle>
          <CardDescription>
            Your subscription has been activated successfully.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          <div className="bg-muted rounded-lg p-4 text-center">
            <p className="text-sm text-muted-foreground mb-1">Your new credit balance</p>
            <p className="text-3xl font-bold">{credits?.balance ?? tierInfo.credits}</p>
            <p className="text-sm text-muted-foreground">credits per month</p>
          </div>

          <div className="space-y-2 text-sm">
            <p className="text-muted-foreground">With {tierInfo.name}, you can:</p>
            <ul className="space-y-1">
              <li>• Send up to {Math.floor(tierInfo.credits / 1)} text messages/month</li>
              <li>• Send up to {Math.floor(tierInfo.credits / 5)} audio messages/month</li>
              <li>• Get pronunciation feedback on every audio message</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Button asChild className="w-full">
              <Link to="/">Start Practicing</Link>
            </Button>
            <Button asChild variant="outline" className="w-full">
              <Link to="/pricing">View Plans</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
