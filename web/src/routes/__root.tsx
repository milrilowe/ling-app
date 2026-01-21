import { Outlet, createRootRouteWithContext, useLocation, useNavigate } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { Toaster } from 'sonner'
import { PenSquare } from 'lucide-react'

import TanStackQueryDevtools from '@/integrations/tanstack-query/devtools'
import { AuthProvider } from '@/contexts/AuthContext'
import { AudioPlayerProvider } from '@/contexts/AudioPlayerContext'
import { AppSidebar } from '@/components/Sidebar'
import { SidebarProvider, SidebarInset, SidebarTrigger } from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'

import type { QueryClient } from '@tanstack/react-query'

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: RootComponent,
})

function RootComponent() {
  const location = useLocation()
  const navigate = useNavigate()
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register'

  const handleNewChat = () => {
    navigate({ to: '/' })
  }

  return (
    <AuthProvider>
      <AudioPlayerProvider>
        {isAuthPage ? (
          <main className="flex-1 overflow-auto">
            <Outlet />
          </main>
        ) : (
          <SidebarProvider defaultOpen={false}>
            <AppSidebar />
            <SidebarInset>
              {/* Mobile header - ChatGPT style */}
              <header className="sticky top-0 z-40 flex h-12 items-center justify-between border-b bg-background px-3 md:hidden">
                <SidebarTrigger />
                <span className="font-semibold text-sm">UtterLabs</span>
                <Button variant="ghost" size="icon" onClick={handleNewChat}>
                  <PenSquare className="h-5 w-5" />
                </Button>
              </header>
              <main className="flex-1 overflow-auto">
                <Outlet />
              </main>
            </SidebarInset>
          </SidebarProvider>
        )}
        <Toaster position="top-center" />
        <TanStackDevtools
          config={{
            position: 'bottom-right',
          }}
          plugins={[
            {
              name: 'Tanstack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
            TanStackQueryDevtools,
          ]}
        />
      </AudioPlayerProvider>
    </AuthProvider>
  )
}
