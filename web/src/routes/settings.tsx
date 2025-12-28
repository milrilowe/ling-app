import { createFileRoute, Link } from '@tanstack/react-router'
import { X, LogOut, Loader2, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { useAuth } from '@/contexts/AuthContext'
import { useSubscription, useCredits, useCreatePortal } from '@/hooks/use-subscription'
import { TIER_INFO } from '@/lib/api'

export const Route = createFileRoute('/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const { user, logout } = useAuth()
  const { data: subscription, isLoading: subscriptionLoading } = useSubscription()
  const { data: credits } = useCredits()
  const createPortal = useCreatePortal()

  const currentTier = subscription?.subscription?.tier ?? 'free'
  const tierInfo = TIER_INFO[currentTier]

  const handleLogout = () => {
    logout.mutate()
  }

  const getInitials = (name?: string, email?: string) => {
    if (name) {
      return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    }
    if (email) {
      return email[0].toUpperCase()
    }
    return '?'
  }

  return (
    <div className="min-h-screen bg-background overflow-y-auto">
      {/* Sticky header */}
      <div className="sticky top-0 bg-background py-4 px-4 border-b mb-6 z-10">
        <div className="flex items-center gap-2">
          <Link to="/">
            <Button variant="ghost" size="icon" className="h-10 w-10">
              <X className="h-5 w-5" />
            </Button>
          </Link>
          <h1 className="text-2xl font-semibold">Settings</h1>
        </div>
      </div>

      <div className="mx-auto max-w-2xl px-6 pb-6">

        <div className="space-y-6">
          {/* Account Section */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Account</CardTitle>
              <CardDescription>Your account information</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <Avatar className="h-12 w-12">
                    <AvatarFallback className="text-lg">
                      {getInitials(user?.name, user?.email)}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    {user?.name && (
                      <p className="font-medium">{user.name}</p>
                    )}
                    <p className="text-sm text-muted-foreground">{user?.email}</p>
                  </div>
                </div>
                <Button
                  variant="outline"
                  onClick={handleLogout}
                  disabled={logout.isPending}
                >
                  {logout.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <>
                      <LogOut className="h-4 w-4 mr-2" />
                      Log out
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Subscription Section */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Subscription</CardTitle>
              <CardDescription>Your plan and credits</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {subscriptionLoading ? (
                <div className="flex items-center justify-center py-4">
                  <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                </div>
              ) : (
                <>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">{tierInfo.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {tierInfo.price === 0 ? 'Free plan' : `$${tierInfo.price}/month`}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-medium">{credits?.balance ?? 0} credits</p>
                      <p className="text-sm text-muted-foreground">
                        {tierInfo.credits} included monthly
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-2 pt-2">
                    <Link to="/pricing" className="flex-1">
                      <Button variant="outline" className="w-full">
                        {currentTier === 'free' ? 'Upgrade plan' : 'Change plan'}
                      </Button>
                    </Link>
                    {currentTier !== 'free' && (
                      <Button
                        variant="outline"
                        onClick={() => createPortal.mutate()}
                        disabled={createPortal.isPending}
                      >
                        {createPortal.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <>
                            Manage billing
                            <ExternalLink className="h-4 w-4 ml-2" />
                          </>
                        )}
                      </Button>
                    )}
                  </div>
                </>
              )}
            </CardContent>
          </Card>

        </div>
      </div>
    </div>
  )
}
